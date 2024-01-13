// Copyright (c) 2023 Red Hat, Inc.

package controller

// This file hosts the AddonReconciler implementation registering for the framework's ManagedClusterAddOn CRs.

import (
	"context"
	"fmt"
	ramenv1alpha1 "github.com/ramendr/ramen/api/v1alpha1"
	mcra "github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/metrics"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// AddonReconciler is a receiver representing the MultiCluster-Resiliency-Addon operator reconciler for
// ManagedClusterAddOn CRs.
type AddonReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// SetupWithManager is used for setting up the controller named 'mcra-addon-controller' with the manager. Using
// Predicates for filtering, only accepting ManagedClusterAddon with the target annotation key
// 'multicluster-resiliency-addon/dr-controller', with a value pointing to a DRPlacementCluster for failing over.
func (r *AddonReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("mcra-addon-controller").
		For(&clusterv1.ManagedCluster{}).
		WithEventFilter(verifyManagedCluster(hasAddon(ctx, mgr.GetClient(), mcra.AddonName))).
		WithEventFilter(verifyManagedCluster(acceptedByHub())).
		WithEventFilter(verifyManagedCluster(joinedHub())).
		WithEventFilter(verifyManagedCluster(notAvailable())).
		Complete(r)
}

// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=authorization.k8s.io,resources=subjectaccessreviews,verbs=get;create
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests;certificatesigningrequests/approval,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=signers,verbs=approve
// +kubebuilder:rbac:groups=cluster.open-cluster-management.io,resources=managedclusters,verbs=get;list;watch
// +kubebuilder:rbac:groups=work.open-cluster-management.io,resources=manifestworks,verbs=create;update;get;list;watch;delete;deletecollection;patch
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=clustermanagementaddons,verbs=get;list;watch
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=clustermanagementaddons/finalizers,verbs=update
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=clustermanagementaddons/status,verbs=update;patch
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=managedclusteraddons,verbs=get;list;watch
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=managedclusteraddons,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=managedclusteraddons/finalizers,verbs=update
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=managedclusteraddons/status,verbs=update;patch
// +kubebuilder:rbac:groups=authentication.k8s.io,resources=tokenreviews,verbs=create
// +kubebuilder:rbac:groups=ramendr.openshift.io,resources=drplacementcontrols,verbs=get;list;update;patch

// Reconcile is watching ManagedClusters and will trigger a DRPlacementControl failover. Note, not eligible
// ManagedClusters for failover. i.e., the cluster is not accepted by the hub, hasn't joined the hub, or is available.
func (r *AddonReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	mc := &clusterv1.ManagedCluster{}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: req.Name}, mc); err != nil {
		if errors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("%s ManagedCluster not found", req.Name))
			return ctrl.Result{}, nil
		}
		logger.Error(err, fmt.Sprintf("%s ManagedCluster failed fetching", req.Name))
		return ctrl.Result{}, err
	}

	metrics.DRClusterNotAvailable.WithLabelValues(mc.Name).Inc()

	drControls := &ramenv1alpha1.DRPlacementControlList{}
	if err := r.Client.List(ctx, drControls); err != nil {
		// TODO handle no controls or other error
		return ctrl.Result{}, err
	}

	failuresFound := false
	for _, drControl := range drControls.Items {
		if drControl.Status.PreferredDecision.ClusterName == mc.Name {
			// TODO take this out of the loop
			failoverPatch := &ramenv1alpha1.DRPlacementControl{
				Spec: ramenv1alpha1.DRPlacementControlSpec{
					Action: ramenv1alpha1.ActionFailover,
				},
			}

			if err := r.Client.Patch(ctx, &drControl, client.StrategicMergeFrom(failoverPatch)); err != nil {
				logger.Error(err, fmt.Sprintf("failed to failover %s DRPlacementControl for application %s", drControl.Name, drControl.Namespace))
				failuresFound = true
			}

			metrics.DRApplicationFailover.WithLabelValues(mc.Name, drControl.Name, drControl.Namespace)
		}
	}

	if failuresFound {
		return ctrl.Result{}, fmt.Errorf("failed to failover one or more DR placement controls")
	}

	return ctrl.Result{}, nil
}
