// Copyright (c) 2023 Red Hat, Inc.

package controller

// This file hosts our Cluster Reconciler implementation registering for our ResilientCluster CRs.

import (
	"context"
	apiv1 "github.com/rhecosystemappeng/multicluster-resiliency-addon/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const finalizerName = "multicluster-resiliency-addon/finalizer"

// ClusterReconciler is a receiver representing the MultiCluster-Resiliency-Addon operator reconciler for
// ResilientCluster CRs.
type ClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=appeng.ecosystem.redhat.com,resources=resilientclusters,verbs=*
// +kubebuilder:rbac:groups=appeng.ecosystem.redhat.com,resources=resilientclusters/finalizer,verbs=*
// +kubebuilder:rbac:groups=appeng.ecosystem.redhat.com,resources=resilientclusters/status,verbs=*

func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("TODO-ADD-CLUSTER-RECONCILE-LOGIC")
	// TODO

	return ctrl.Result{Requeue: false}, nil
}

// SetupWithManager is used for setting up the controller with the manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("mcra-managed-cluster-cluster-controller").
		For(&apiv1.ResilientCluster{}).
		Complete(r)
}
