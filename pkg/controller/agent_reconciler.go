// Copyright (c) 2023 Red Hat, Inc.

package controller

// This file hosts our Agent Reconciler implementation registering for framework's ManagedClusterAddOn CRs.

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// AgentReconciler is a receiver representing the MultiCluster-Resiliency-Addon operator reconciler for
// ManagedClusterAddOn CRs.
type AgentReconciler struct {
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

func (r *AgentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// fetch the ManagedClusterAddon resource
	mcAddonNsn := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}
	mcAddon := &addonv1alpha1.ManagedClusterAddOn{}
	if err := r.Client.Get(ctx, mcAddonNsn, mcAddon); err != nil {
		if errors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("%s named not found in namepsace.", mcAddonNsn.String()))
			return ctrl.Result{}, nil
		}

		logger.Error(err, "failed to fetch ManagedClusterAddon")
		return ctrl.Result{}, err
	}

	// is the ManagedClusterAddon for our addon?
	if len(mcAddon.OwnerReferences) == 0 ||
		mcAddon.OwnerReferences[0].Kind != "ClusterManagementAddOn" ||
		mcAddon.OwnerReferences[0].Name != "multicluster-resiliency-addon" {
		return ctrl.Result{}, nil
	}

	// is the ManagedClusterAddon being deleted?
	if !mcAddon.ObjectMeta.DeletionTimestamp.IsZero() {
		logger.Info(fmt.Sprintf(
			"ManagedClusterAddon for %s is being deleted, delting corresponding ResilientCluster",
			mcAddonNsn.Namespace))

		// TODO delete ResilientCluster

		return ctrl.Result{}, nil
	}

	// TODO check if ResilientCluster exists
	// TODO if exists, update current addon state
	// TODO if not create new and set owner and initial state

	return ctrl.Result{}, nil
}

// SetupWithManager is used for setting up the controller with the manager.
func (r *AgentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("mcra-managed-cluster-agent-controller").
		For(&addonv1alpha1.ManagedClusterAddOn{}).
		Complete(r)
}
