// Copyright (c) 2023 Red Hat, Inc.

package controller

import (
	"context"
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/hashicorp/go-multierror"
	ramenv1alpha1 "github.com/ramendr/ramen/api/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// okToFailoverStates is a fixed array listing the state a DRPlacementControl needs to be in for us to initiate a failover.
var okToFailoverStates = [...]ramenv1alpha1.DRState{ramenv1alpha1.Deploying, ramenv1alpha1.Deployed, ramenv1alpha1.Relocated}

var drApplicationFailoverMetric = *prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "dr_application_failover_count",
	Help: "Counter for DR Applications failover initiated by the Regional DR Trigger Operator",
}, []string{"dr_cluster_name", "dr_control_name", "dr_application_name"})

// DRTriggerController is a receiver representing the DRTriggerOperator controller for ManagedCluster CRs
type DRTriggerController struct {
	Client client.Client
	Scheme *runtime.Scheme
}

// SetupWithManager is used for setting up the controller. Using Predicates for filtering, only accepting
// ManagedCluster eligible for failing over
func (r *DRTriggerController) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("regional-dr-trigger-controller").
		For(&clusterv1.ManagedCluster{}).
		WithEventFilter(predicate.Funcs{
			DeleteFunc: func(e event.DeleteEvent) bool {
				return false
			},
		}).
		Complete(r)
}

// +kubebuilder:rbac:groups="",resources=events,verbs=create
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;create;update
// +kubebuilder:rbac:groups=cluster.open-cluster-management.io,resources=managedclusters,verbs=get;watch;list
// +kubebuilder:rbac:groups=ramendr.openshift.io,resources=drplacementcontrols,verbs=get;watch;list;patch
// +kubebuilder:rbac:groups=authorization.k8s.io,resources=subjectaccessreviews,verbs=get;create
// +kubebuilder:rbac:groups=authentication.k8s.io,resources=tokenreviews,verbs=create

// Reconcile is watching ManagedClusters and will trigger a DRPlacementControl failover. Note, not eligible
// events for failover. i.e., the cluster is not accepted by the hub, hasn't joined the hub, or is available. // Are
// filtered out by event filtering Predicates.
func (r *DRTriggerController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithName("mc-controller")
	ctx = log.IntoContext(ctx, logger)

	mc := &clusterv1.ManagedCluster{}
	if err := r.Client.Get(ctx, req.NamespacedName, mc); err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Info("managed cluster deleted")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	logger.Info("got request for managed cluster")

	if !meta.IsStatusConditionTrue(mc.Status.Conditions, clusterv1.ManagedClusterConditionJoined) {
		logger.Info("managed cluster not joined")
		return ctrl.Result{}, nil
	}

	if !mc.Spec.HubAcceptsClient {
		logger.Info("managed cluster not accepted by hub")
		return ctrl.Result{}, nil
	}

	if meta.IsStatusConditionTrue(mc.Status.Conditions, clusterv1.ManagedClusterConditionAvailable) {
		logger.Info("managed cluster is available, no failing over required")
		return ctrl.Result{}, nil
	}

	drControls := &ramenv1alpha1.DRPlacementControlList{}
	if err := r.Client.List(ctx, drControls); err != nil {
		return ctrl.Result{}, err
	}

	var errs *multierror.Error
	for _, drControl := range drControls.Items {

		// dr controls using current managed cluster
		if drControl.Status.PreferredDecision.ClusterName == mc.Name {
			logger.Info("found dr control for managed cluster",
				"drpc_name", drControl.Name, "drpc_ns", drControl.Namespace)
			// dr controls not already failed-over
			if drControl.Spec.Action != ramenv1alpha1.ActionFailover {
				// dr control in phase suitable for a failover
				if isPhaseOkForFailover(drControl) {
					// dr control peer is ready
					if meta.IsStatusConditionTrue(drControl.Status.Conditions, ramenv1alpha1.ConditionPeerReady) {
						// patch do control and initiate a failover process
						if err := r.patchDRPlacementControl(ctx, drControl, ramenv1alpha1.ActionFailover); err != nil {
							errs = multierror.Append(err, errs)
						} else {
							logger.Info("successfully patched dr control for a failover",
								"drpc_name", drControl.Name, "drpc_ns", drControl.Namespace)
							drApplicationFailoverMetric.WithLabelValues(mc.Name, drControl.Name, drControl.Namespace).Inc()
						}
					} else {
						logger.Info("dr control peer not available for a failover",
							"drpc_name", drControl.Name, "drpc_ns", drControl.Namespace)
					}
				} else {
					logger.Info("dr control not in suitable phase for a failover", "drpc_name",
						drControl.Name, "drpc_ns", drControl.Namespace, "dr_phase", drControl.Status.Phase)
				}
			} else {
				logger.Info("dr control failover already initiated", "drpc_name",
					drControl.Name, "drpc_ns", drControl.Namespace)
			}
		}
	}

	return ctrl.Result{}, errs.ErrorOrNil()
}

// patchDRPlacementControl is used to patch a DRPlacementControl for triggering a failover process
func (r *DRTriggerController) patchDRPlacementControl(ctx context.Context, control ramenv1alpha1.DRPlacementControl, action ramenv1alpha1.DRAction) error {
	drControlObj := &ramenv1alpha1.DRPlacementControl{}
	drControlSubject := types.NamespacedName{Namespace: control.Namespace, Name: control.Name}
	if err := r.Client.Get(ctx, drControlSubject, drControlObj); err != nil {
		return err
	}

	failoverPatch := &ramenv1alpha1.DRPlacementControl{
		Spec: ramenv1alpha1.DRPlacementControlSpec{
			Action: action,
		},
	}

	rawPatch, err := json.Marshal(failoverPatch)
	if err != nil {
		return err
	}

	if err = r.Client.Patch(ctx, drControlObj, client.RawPatch(types.MergePatchType, rawPatch)); err != nil {
		return err
	}

	return nil
}

// isPhaseOkForFailover is a utility function that returns true if the DRPlacementControl.Status.Spec is in a state
// allowed for failing over. i.e., Deployed.
func isPhaseOkForFailover(control ramenv1alpha1.DRPlacementControl) bool {
	for _, state := range okToFailoverStates {
		if state == control.Status.Phase {
			return true
		}
	}
	return false
}

func init() {
	metrics.Registry.MustRegister(drApplicationFailoverMetric)
}
