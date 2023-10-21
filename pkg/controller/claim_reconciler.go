// Copyright (c) 2023 Red Hat, Inc.

package controller

import (
	"context"
	"fmt"
	hivev1 "github.com/openshift/hive/apis/hive/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ClaimReconciler is a receiver representing the MultiCluster-Resiliency-Addon operator reconciler for
// ClusterClaim CRs.
type ClaimReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile is watching ClusterClaim CRs updating the appropriate ResilientCluster CRs, and deleting replaced
// ClusterClaim CRs. Note, further permissions are listed in ClusterReconciler.Reconcile and AddonReconciler.Reconcile.
//
// +kubebuilder:rbac:groups=hive.openshift.io/v1,resources=clusterpools,verbs=get;list
// +kubebuilder:rbac:groups=hive.openshift.io/v1,resources=clusterclaims,verbs=*
func (r *ClaimReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	subject := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}

	// fetch the ClusterClaim cr, end loop if not found
	claim := &hivev1.ClusterClaim{}
	if err := r.Client.Get(ctx, subject, claim); err != nil {
		if errors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("%s not found", subject.String()))
			return ctrl.Result{}, nil
		}

		logger.Error(err, fmt.Sprintf("%s fetch failed", subject.String()))
		return ctrl.Result{}, err
	}

	// deletion cleanup
	if !claim.DeletionTimestamp.IsZero() {
		// ClusterClaim is in delete process
		if controllerutil.ContainsFinalizer(claim, mcraFinalizerName) {
			// TODO add cleanup code here

			// when cleanup done, remove the finalizer
			controllerutil.RemoveFinalizer(claim, mcraFinalizerName)
			if err := r.Client.Update(ctx, claim); err != nil {
				logger.Error(err, fmt.Sprintf("%s failed removing finalizer", subject.String()))
				return ctrl.Result{}, err
			}
		}

		// no further progress is required while deleting objects
		return ctrl.Result{}, nil
	}

	// TODO if claim is ready and has annotation with previous claim name, remove old claim

	return ctrl.Result{}, nil
}

// SetupWithManager is used for setting up the controller named 'mcra-managed-cluster-claim-controller' with the manager.
// It uses predicates as event filters for verifying only handling ClusterClaim CRs created by us.
func (r *ClaimReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("mcra-managed-cluster-claim-controller").
		For(&hivev1.ClusterClaim{}).
		WithEventFilter(verifyOwnerPredicate(objectOwnerIsAResilientCluster)).
		Complete(r)
}
