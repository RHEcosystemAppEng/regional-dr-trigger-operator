// Copyright (c) 2023 Red Hat, Inc.

package controller

// This file hosts our Agent Reconciler implementation registering for framework's ManagedClusterAddOn CRs.

import (
	"context"
	"fmt"
	apiv1 "github.com/rhecosystemappeng/multicluster-resiliency-addon/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

// AddonReconciler is a receiver representing the MultiCluster-Resiliency-Addon operator reconciler for
// ManagedClusterAddOn CRs.
type AddonReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=managedclusteraddons,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=configmaps;events,verbs=get;list;watch;create;update;delete;deletecollection;patch
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=authorization.k8s.io,resources=subjectaccessreviews,verbs=get;create
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests;certificatesigningrequests/approval,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=signers,verbs=approve
// +kubebuilder:rbac:groups=cluster.open-cluster-management.io,resources=managedclusters,verbs=get;list;watch
// +kubebuilder:rbac:groups=work.open-cluster-management.io,resources=manifestworks,verbs=create;update;get;list;watch;delete;deletecollection;patch
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=managedclusteraddons/finalizers,verbs=update
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=clustermanagementaddons/finalizers,verbs=update
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=clustermanagementaddons,verbs=get;list;watch
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=managedclusteraddons,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=managedclusteraddons/status,verbs=update;patch

func (r *AddonReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// name and namespace are identical for both the ManagedClusterAddon and ResilientCluster crs
	subject := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}

	// fetch the ManagedClusterAddon cr, end loop if not found
	mca := &addonv1alpha1.ManagedClusterAddOn{}
	if err := r.Client.Get(ctx, subject, mca); err != nil {
		if errors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("%s not found", subject.String()))
			return ctrl.Result{}, nil
		}

		logger.Error(err, "failed to fetch ManagedClusterAddon")
		return ctrl.Result{}, err
	}

	// is the ManagedClusterAddon for our addon? end loop if not
	if len(mca.OwnerReferences) == 0 ||
		mca.OwnerReferences[0].Kind != "ClusterManagementAddOn" ||
		mca.OwnerReferences[0].Name != "multicluster-resiliency-addon" {
		return ctrl.Result{}, nil
	}

	// attempt to fetch the ResilientCluster cr, note if found or not
	rc := &apiv1.ResilientCluster{}
	rcFound := true
	if err := r.Client.Get(ctx, subject, rc); err != nil {
		// only not-found errors are acceptable here
		if !errors.IsNotFound(err) {
			logger.Error(err, fmt.Sprintf("ResilientCluster not found %s", subject.String()))
			return ctrl.Result{}, err
		}
		rcFound = false
	}

	// is the ManagedClusterAddon being deleted? if so, we need to delete the corresponding ResilientCluster
	if !mca.DeletionTimestamp.IsZero() {
		logger.Info(fmt.Sprintf("ManagedClusterAddon %s is being deleted, deleting ResilientCluster", subject.String()))

		// do we have a ResilientCluster? if so, we need to delete it
		if rcFound {
			// if the ResilientCluster NOT already being deleted, we need to delete it
			if rc.DeletionTimestamp.IsZero() {
				if err := r.Client.Delete(ctx, rc); err != nil {
					logger.Error(err, fmt.Sprintf("failed deleting ResilientCluster %s", subject.String()))
					return ctrl.Result{}, err
				}
			}
		}

		return ctrl.Result{}, nil
	}

	// do we have a ResilientCluster? we need to either create or update it
	if rcFound {
		// ResilientCluster exists, we need to update with the current status
		rc.Status.Conditions = mca.Status.Conditions
		if err := r.Client.Update(ctx, rc); err != nil {
			logger.Error(err, fmt.Sprintf("failed updating ResilientCluster %s", subject.String()))
			return ctrl.Result{}, err
		}
	} else {
		// ResilientCluster doesn't exist, we need to create it
		rc.SetName(subject.Name)
		rc.SetNamespace(subject.Namespace)
		rc.SetCreationTimestamp(metav1.NewTime(time.Now()))
		rc.SetFinalizers([]string{finalizerName})
		rc.SetOwnerReferences([]metav1.OwnerReference{
			*metav1.NewControllerRef(mca, mca.GetObjectKind().GroupVersionKind()),
		})

		if err := r.Client.Create(ctx, rc); err != nil {
			logger.Error(err, fmt.Sprintf("failed creating ResilientCluster %s", subject.String()))
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager is used for setting up the controller with the manager.
func (r *AddonReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("mcra-managed-cluster-agent-controller").
		For(&addonv1alpha1.ManagedClusterAddOn{}).
		Complete(r)
}
