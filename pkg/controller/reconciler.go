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
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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
		WithEventFilter(verifyObject(hasAddon(ctx, mgr.GetClient(), mcra.AddonName))).
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

// Reconcile is watching ManagedClusters and will trigger a failover when required.
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

	if !mc.DeletionTimestamp.IsZero() {
		// no further progress is required while deleting objects
		return ctrl.Result{}, nil
	}

	// determine whether a ManagedCluster is available.
	if meta.IsStatusConditionFalse(mc.Status.Conditions, clusterv1.ManagedClusterConditionAvailable) {
		metrics.DRClusterNotAvailable.WithLabelValues(mc.Name).Inc()

		drControls := &ramenv1alpha1.DRPlacementControlList{}
		if err := r.Client.List(ctx, drControls); err != nil {
			// TODO handle no controllers or other error
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
	}

	return ctrl.Result{}, nil
}

// verifyObject takes a function and returns a predicate that verifies the event object for all events against
// the function.
func verifyObject(fn func(obj client.Object) bool) predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(createEvent event.CreateEvent) bool {
			return fn(createEvent.Object)
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			return fn(deleteEvent.Object)
		},
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			return fn(updateEvent.ObjectOld)
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			return fn(genericEvent.Object)
		},
	}
}

// hasAddon is a utility function that takes an addon name and returns a function that takes a client.Object and
// returns true if we have the required ManagedClusterAddon in a namespace of the Object (expected to be ManagedCluster).
func hasAddon(ctx context.Context, clnt client.Client, addon string) func(obj client.Object) bool {
	return func(obj client.Object) bool {
		mcaSubject := types.NamespacedName{Name: addon, Namespace: obj.GetName()}
		mca := &addonv1alpha1.ManagedClusterAddOn{}
		// return whether the managedclusteraddon exists on the managedcluster cluster-namespace
		if err := clnt.Get(ctx, mcaSubject, mca); err != nil {
			if !errors.IsNotFound(err) {
				logger := log.FromContext(ctx)
				logger.Error(err, fmt.Sprintf("failed fetching %s ManagedClusterAddon for %s ManagedCluster", addon, obj.GetName()))
			}
			// the addon is NOT installed for the managed cluster
			return false
		}
		// the addon is installed for the managed cluster
		return true
	}
}
