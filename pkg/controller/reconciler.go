// Copyright (c) 2023 Red Hat, Inc.

package controller

// This file hosts the DRTriggerReconciler implementation registering for ManagedCluster CRs and triggering a failover.

import (
	"context"
	"fmt"
	ramenv1alpha1 "github.com/ramendr/ramen/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"regional-dr-trigger-operator/pkg/metrics"

	clusterv1 "open-cluster-management.io/api/cluster/v1"
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

// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch
// +kubebuilder:rbac:groups=cluster.open-cluster-management.io,resources=managedclusters,verbs=get;watch
// +kubebuilder:rbac:groups=ramendr.openshift.io,resources=drplacementcontrols,verbs=*

// Reconcile is watching ManagedClusters and will trigger a DRPlacementControl failover. Note, not eligible
// events for failover. i.e., the cluster is not accepted by the hub, hasn't joined the hub, or is available. // Are
// filtered out by event filtering Predicates.
func (r *DRTriggerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// the name in the request is the managed cluster name and cluster-namespace name.
	mcName := req.Name

	metrics.DRClusterNotAvailable.WithLabelValues(mcName).Inc()

	// fetch all DRPlacementControl in the Hub Cluster
	drControls := &ramenv1alpha1.DRPlacementControlList{}
	if err := r.Client.List(ctx, drControls); err != nil {
		logger.Error(err, "failed to fetch DRPlacementControlsList")
		return ctrl.Result{}, err
	}

	failuresFound := false
	// iterate over DRPlacementControls and cherry-pick ones that belongs to the current Managed Cluster
	for _, drControl := range drControls.Items {
		// check if the current DRPlacementControl belongs to the ManagedCluster reported not available
		if drControl.Status.PreferredDecision.ClusterName == mcName {
			// check if the current DRPlacementControl phase is ok for failing over
			if isPhaseOkForFailover(drControl) {
				drControlSubject := types.NamespacedName{Namespace: drControl.Namespace, Name: drControl.Name}

				// check if the peer is ready and requeue if it's not
				if !meta.IsStatusConditionTrue(drControl.Status.Conditions, ramenv1alpha1.ConditionPeerReady) {
					logger.Error(
						fmt.Errorf("attempting to failover %s, peer not ready", drControlSubject.String()),
						"peer not ready")
					return ctrl.Result{Requeue: true}, nil
				}

				// fetch the current DRPlacementControl
				drControlObj := &ramenv1alpha1.DRPlacementControl{}
				if err := r.Client.Get(ctx, drControlSubject, drControlObj); err != nil {
					logger.Error(err, fmt.Sprintf("failed fetching DRPlacementControl %s", drControlSubject.String()))
					return ctrl.Result{}, err
				}

				// create the failover action patch
				failoverPatch := &ramenv1alpha1.DRPlacementControl{
					Spec: ramenv1alpha1.DRPlacementControlSpec{
						Action: ramenv1alpha1.ActionFailover,
					},
				}

				// patch the DRPlacementControl for the failover action
				if err := r.Client.Patch(ctx, drControlObj, client.StrategicMergeFrom(failoverPatch)); err != nil {
					// mark error found but don't break - we might have more DRPlacementControl to patch
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
