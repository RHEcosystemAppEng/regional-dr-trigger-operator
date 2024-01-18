// Copyright (c) 2023 Red Hat, Inc.

package controller

// This file hosts the DRTriggerReconciler implementation registering for ManagedCluster CRs and triggering a failover.

import (
	"context"
	"encoding/json"
	"fmt"
	ramenv1alpha1 "github.com/ramendr/ramen/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	"regional-dr-trigger-operator/pkg/metrics"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// okToFailoverStates is a fixed array listing the state a DRPlacementControl needs to be in for us to initiate a failover.
var okToFailoverStates = [...]ramenv1alpha1.DRState{ramenv1alpha1.Deploying, ramenv1alpha1.Deployed, ramenv1alpha1.Relocated}

// DRTriggerReconciler is a receiver representing the DRTriggerOperator reconciler for ManagedCluster CRs.
type DRTriggerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// SetupWithManager is used for setting up the controller. Using Predicates for filtering, only accepting
// ManagedCluster eligible for failing over.
func (r *DRTriggerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("regional-dr-trigger-controller").
		For(&clusterv1.ManagedCluster{}).
		WithEventFilter(verifyManagedCluster(acceptedByHub())).
		WithEventFilter(verifyManagedCluster(joinedHub())).
		WithEventFilter(verifyManagedCluster(notAvailable())).
		Complete(r)
}

// +kubebuilder:rbac:groups="*",resources=events,verbs=create
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;update
// +kubebuilder:rbac:groups=cluster.open-cluster-management.io,resources=managedclusters,verbs=get;watch;list
// +kubebuilder:rbac:groups=ramendr.openshift.io,resources=drplacementcontrols,verbs=get;watch;list;patch
// +kubebuilder:rbac:groups=authorization.k8s.io,resources=subjectaccessreviews,verbs=get;create
// +kubebuilder:rbac:groups=authentication.k8s.io,resources=tokenreviews,verbs=create

// Reconcile is watching ManagedClusters and will trigger a DRPlacementControl failover. Note, not eligible
// events for failover. i.e., the cluster is not accepted by the hub, hasn't joined the hub, or is available. // Are
// filtered out by event filtering Predicates.
func (r *DRTriggerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// the name in the request is the managed cluster name and cluster-namespace name.
	mcName := req.Name

	// fetch all DRPlacementControl in the Hub Cluster
	drControls := &ramenv1alpha1.DRPlacementControlList{}
	if err := r.Client.List(ctx, drControls); err != nil {
		logger.Error(err, "failed to fetch DRPlacementControlsList")
		return ctrl.Result{}, err
	}

	failuresFound := false
	// iterate over DRPlacementControls and cherry-pick ones that belongs to the current Managed Cluster
	for _, drControl := range drControls.Items {
		// check if the current DRPlacementControl runs for the triggering ManagedCluster
		if drControl.Status.PreferredDecision.ClusterName == mcName {
			// check if a failover instruction was already recorded
			if drControl.Spec.Action != ramenv1alpha1.ActionFailover {
				// check if the DRPlacementControl is in a phase suitable for a failover, i.e. Deployed (yes) | Initiating (no)
				if isPhaseOkForFailover(drControl) {
					// patch DRPlacementControl and initiate a failover process
					if err := r.patchDRPlacementControl(ctx, drControl, ramenv1alpha1.ActionFailover); err != nil {
						logger.Error(err, fmt.Sprintf("failed to patch failover DRPlacementControl %s in %s", drControl.Name, drControl.Namespace))
						failuresFound = true
					} else {
						logger.Info(fmt.Sprintf("succesfully patched DRPlacementControl %s in %s for a failover", drControl.Name, drControl.Namespace))
						metrics.DRApplicationFailover.WithLabelValues(mcName, drControl.Name, drControl.Namespace).Inc()
					}
				}
			}
		}
	}

	if failuresFound {
		return ctrl.Result{}, fmt.Errorf("failed to failover one or more DR placement controls")
	}

	return ctrl.Result{}, nil
}

// patchDRPlacementControl is used to patch a DRPlacementControl for triggering a failover process
func (r *DRTriggerReconciler) patchDRPlacementControl(ctx context.Context, control ramenv1alpha1.DRPlacementControl, action ramenv1alpha1.DRAction) error {
	drControlSubject := types.NamespacedName{Namespace: control.Namespace, Name: control.Name}

	// check if the peer is ready
	if !meta.IsStatusConditionTrue(control.Status.Conditions, ramenv1alpha1.ConditionPeerReady) {
		return fmt.Errorf("attempting to failover %s, peer not ready", drControlSubject.String())
	}

	// fetch the DRPlacementControl
	drControlObj := &ramenv1alpha1.DRPlacementControl{}
	if err := r.Client.Get(ctx, drControlSubject, drControlObj); err != nil {
		return err
	}

	// create the failover action patch
	failoverPatch := &ramenv1alpha1.DRPlacementControl{
		Spec: ramenv1alpha1.DRPlacementControlSpec{
			Action: action,
		},
	}

	// serialize the patch
	rawPatch, err := json.Marshal(failoverPatch)
	if err != nil {
		return err
	}

	// patch the DRPlacementControl for with the failover patch
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
