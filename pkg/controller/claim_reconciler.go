// Copyright (c) 2023 Red Hat, Inc.

package controller

import (
	"context"
	"fmt"
	hivev1 "github.com/openshift/hive/apis/hive/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"open-cluster-management.io/api/addon/v1alpha1"
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

// SetupWithManager is used for setting up the controller named 'mcra-managed-cluster-claim-controller' with the manager.
// It uses predicates as event filters for verifying only handling ClusterClaim CRs created by us.
func (r *ClaimReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("mcra-claim-controller").
		For(&hivev1.ClusterClaim{}).
		WithEventFilter(verifyObject(hasAnnotation(annotationPreviousSpoke))).
		Complete(r)
}

// +kubebuilder:rbac:groups=hive.openshift.io,resources=clusterpools,verbs=get;list
// +kubebuilder:rbac:groups=hive.openshift.io,resources=clusterclaims,verbs=get;list;create;watch
// +kubebuilder:rbac:groups=hive.openshift.io,resources=clusterdeployments,verbs=*

// Reconcile is watching ClusterClaim CRs updating the appropriate ResilientCluster CRs, and deleting replaced
// ClusterClaim CRs. Note, further permissions are listed in ClusterReconciler.Reconcile and AddonReconciler.Reconcile.
func (r *ClaimReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	claimSubject := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}

	// fetch the ClusterClaim cr, end loop if not found
	claim := &hivev1.ClusterClaim{}
	if err := r.Client.Get(ctx, claimSubject, claim); err != nil {
		if errors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("%s not found", claimSubject.String()))
			return ctrl.Result{}, nil
		}

		logger.Error(err, fmt.Sprintf("%s fetch failed", claimSubject.String()))
		return ctrl.Result{}, err
	}

	// if ClusterClaim is in delete process
	if !claim.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(claim, finalizerUsedByMcra) {
			// TODO verify ClusterClaim status before proceeding

			// the old spoke name is the cluster-namespace name
			oldSpokeName := claim.GetAnnotations()[annotationPreviousSpoke]
			// the new spoke name is the namespace in which the deployment exists
			newSpokeName := claim.Spec.Namespace

			// the ClusterDeployment resource resides in the cluster-namespace and has the same name as its claim
			deploymentSubject := types.NamespacedName{
				Namespace: oldSpokeName,
				Name:      claim.Name,
			}

			// fetch ClusterDeployment from previous cluster
			deployment := &hivev1.ClusterDeployment{}
			if err := r.Client.Get(ctx, deploymentSubject, deployment); err != nil {
				// TODO handle errors - don't return without removing annotation and finalizer
			}

			// the ManagedClusterAddon resides in the cluster-namespace
			mcaSubject := types.NamespacedName{
				Namespace: oldSpokeName,
				Name:      "multicluster-resiliency-addon",
			}

			// fetch ManagedClusterAddOn from previous cluster
			mca := &v1alpha1.ManagedClusterAddOn{}
			if err := r.Client.Get(ctx, mcaSubject, mca); err != nil {
				// TODO handle errors - don't return without removing annotation and finalizer
			}

			// fetch all AddOnDeploymentConfig from the cluster namespace
			configs := &v1alpha1.AddOnDeploymentConfigList{}
			if err := r.Client.List(ctx, configs, &client.ListOptions{Namespace: oldSpokeName}); err != nil {
				// TODO handle errors - don't return without removing annotation and finalizer
			}

			// TODO - delete ClusterDeployment
			// TODO - copy ManagedClusterAddon to new cluster-namespace
			// TODO - copy (if any) all AddOnDeploymentConfig to new cluster-namespace
			// TODO - move replace spoke in ManagedClusterSets (add rbac markers)

			// when done, remove the finalizer and annotation
			annotations := claim.GetAnnotations()
			delete(claim.GetAnnotations(), annotationPreviousSpoke)
			claim.SetAnnotations(annotations)

			controllerutil.RemoveFinalizer(claim, finalizerUsedByMcra)

			if err := r.Client.Update(ctx, claim); err != nil {
				logger.Error(err, fmt.Sprintf("%s failed removing finalizer and annotation from claim", claimSubject.String()))
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}
