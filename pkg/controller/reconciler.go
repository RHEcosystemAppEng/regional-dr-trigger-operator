// Copyright (c) 2023 Red Hat, Inc.

package controller

// This file hosts the AddonReconciler implementation registering for the framework's ManagedClusterAddOn CRs.

import (
	"context"
	"fmt"
	ramenv1alpha1 "github.com/ramendr/ramen/api/v1alpha1"
	mcra "github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/metrics"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	clusterv1 "open-cluster-management.io/api/cluster/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// okToFailoverStates is a fixed array listing the state a DRPlacementControl needs to be in for us to initiate a failover.
var okToFailoverStates = [...]ramenv1alpha1.DRState{ramenv1alpha1.Deploying, ramenv1alpha1.Deployed, ramenv1alpha1.Relocated}

// AddonReconciler is a receiver representing the MultiCluster-Resiliency-Addon operator reconciler for
// ManagedClusterAddOn CRs.
type AddonReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// SetupWithManager is used for setting up the controller named 'mcra-addon-controller' with the manager. Using
// Predicates for filtering, only accepting ManagedClusterAddon with the target annotation key
// 'multicluster-resiliency-addon/dr-controller', with a value pointing to a DRPlacementCluster for failing over.
func (r *AddonReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("mcra-addon-controller").
		For(&clusterv1.ManagedCluster{}).
		WithEventFilter(verifyManagedCluster(addonInstalled(mcra.AddonName))).
		WithEventFilter(verifyManagedCluster(acceptedByHub())).
		WithEventFilter(verifyManagedCluster(joinedHub())).
		WithEventFilter(verifyManagedCluster(notAvailable())).
		Complete(r)
}

// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;delete;patch
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=authorization.k8s.io,resources=subjectaccessreviews,verbs=get;create
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests;certificatesigningrequests/approval,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=signers,verbs=approve
// +kubebuilder:rbac:groups=cluster.open-cluster-management.io,resources=managedclusters,verbs=get;list;watch;update;delete;patch
// +kubebuilder:rbac:groups=work.open-cluster-management.io,resources=manifestworks,verbs=create;update;get;list;watch;delete;deletecollection;patch
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=clustermanagementaddons,verbs=get;list;watch
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=clustermanagementaddons/finalizers,verbs=update
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=clustermanagementaddons/status,verbs=update;patch
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=managedclusteraddons,verbs=get;list;watch
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=managedclusteraddons,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=managedclusteraddons/finalizers,verbs=update
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=managedclusteraddons/status,verbs=update;patch
// +kubebuilder:rbac:groups=authentication.k8s.io,resources=tokenreviews,verbs=create
// +kubebuilder:rbac:groups=ramendr.openshift.io,resources=drplacementcontrols,verbs=*
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=addondeploymentconfigs,verbs=*

// Reconcile is watching ManagedClusters and will trigger a DRPlacementControl failover. Note, not eligible
// ManagedClusters for failover. i.e., the cluster is not accepted by the hub, hasn't joined the hub, or is available.
func (r *AddonReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// the name in the request is the managed cluster name and cluster-namespace name.
	mcName := req.Name

	metrics.DRClusterNotAvailable.WithLabelValues(mcName).Inc()

	drControls := &ramenv1alpha1.DRPlacementControlList{}
	if err := r.Client.List(ctx, drControls); err != nil {
		logger.Error(err, "failed to fetch DRPlacementControlsList")
		return ctrl.Result{}, err
	}

	failuresFound := false
	for _, drControl := range drControls.Items {
		// check if the current dr-control belongs to the managed cluster reported not available
		if drControl.Status.PreferredDecision.ClusterName == mcName {
			// check if the current dr-control phase is ok for failing over
			if isPhaseOkForFailover(drControl) {
				drControlSubject := types.NamespacedName{Namespace: drControl.Namespace, Name: drControl.Name}

				// check if the peer is ready
				if meta.IsStatusConditionFalse(drControl.Status.Conditions, ramenv1alpha1.ConditionPeerReady) {
					logger.Error(
						fmt.Errorf("attempting to failover %s, peer not ready", drControlSubject.String()),
						"peer not ready")
					return ctrl.Result{Requeue: true}, nil
				}

				drControlObj := &ramenv1alpha1.DRPlacementControl{}
				if err := r.Client.Get(ctx, drControlSubject, drControlObj); err != nil {
					logger.Error(err, fmt.Sprintf("failed fetching DRPlacementControl %s", drControlSubject.String()))
					return ctrl.Result{}, err
				}

				failoverPatch := &ramenv1alpha1.DRPlacementControl{
					Spec: ramenv1alpha1.DRPlacementControlSpec{
						Action: ramenv1alpha1.ActionFailover,
					},
				}

				if err := r.Client.Patch(ctx, drControlObj, client.StrategicMergeFrom(failoverPatch)); err != nil {
					// mark error found but don't break - we might have more dr-controls to patch
					logger.Error(err, fmt.Sprintf("failed to failover %s DRPlacementControl for application %s", drControl.Name, drControl.Namespace))
					failuresFound = true
				}

				metrics.DRApplicationFailover.WithLabelValues(mcName, drControl.Name, drControl.Namespace)
			}
		}
	}

	if failuresFound {
		return ctrl.Result{}, fmt.Errorf("failed to failover one or more DR placement controls")
	}

	return ctrl.Result{}, nil
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
