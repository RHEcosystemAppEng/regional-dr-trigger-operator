// Copyright (c) 2023 Red Hat, Inc.

package controller

// This file hosts our ClusterReconciler implementation registering for our ResilientCluster CRs.

import (
	"context"
	"fmt"
	apiv1 "github.com/rhecosystemappeng/multicluster-resiliency-addon/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const mcraFinalizerName = "multicluster-resiliency-addon/finalizer"

// ClusterReconciler is a receiver representing the MultiCluster-Resiliency-Addon operator reconciler for
// ResilientCluster CRs.
type ClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=appeng.ecosystem.redhat.com,resources=resilientclusters,verbs=*
// +kubebuilder:rbac:groups=appeng.ecosystem.redhat.com,resources=resilientclusters/finalizer,verbs=*
// +kubebuilder:rbac:groups=appeng.ecosystem.redhat.com,resources=resilientclusters/status,verbs=*
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=addondeploymentconfigs,verbs=get;list;watch

// Reconcile is watching ResilientCluster CRs, determining whether a new Spoke cluster is required, and handling
// the cluster provisioning using OpenShift Hive API.
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	subject := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}

	// fetch the ResilientCluster cr, end loop if not found
	rc := &apiv1.ResilientCluster{}
	if err := r.Client.Get(ctx, subject, rc); err != nil {
		if errors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("%s not found", subject.String()))
			return ctrl.Result{}, nil
		}

		logger.Error(err, fmt.Sprintf("%s fetch failed", subject.String()))
		return ctrl.Result{}, err
	}

	// deletion cleanup
	if !rc.DeletionTimestamp.IsZero() {
		// ResilientCluster is in delete process
		if controllerutil.ContainsFinalizer(rc, mcraFinalizerName) {
			// TODO add cleanup code here

			// when cleanup done, remove the finalizer
			controllerutil.RemoveFinalizer(rc, mcraFinalizerName)
			if err := r.Client.Update(ctx, rc); err != nil {
				logger.Error(err, fmt.Sprintf("%s failed removing finalizer", subject.String()))
				return ctrl.Result{}, err
			}
		}

		// no further progress is required while deleting objects
		return ctrl.Result{}, nil
	}
	// TODO add business logic here

	return ctrl.Result{}, nil
}

// SetupWithManager is used for setting up the controller named 'mcra-managed-cluster-cluster-controller' with the
// manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("mcra-managed-cluster-cluster-controller").
		For(&apiv1.ResilientCluster{}).
		Complete(r)
}
