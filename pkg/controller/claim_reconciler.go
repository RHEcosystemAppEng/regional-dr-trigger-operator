// Copyright (c) 2023 Red Hat, Inc.

package controller

import (
	"context"
	"fmt"
	hivev1 "github.com/openshift/hive/apis/hive/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
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
// +kubebuilder:rbac:groups=hive.openshift.io,resources=clusterdeployments,verbs=get;list;create;delete

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

			// the OLD spoke name was set as an annotation when we created the ClusterClaim in ClusterReconciler
			oldSpokeName := claim.GetAnnotations()[annotationPreviousSpoke]
			// the NEW spoke name is the target namespace in which the ClusterDeployment was created
			newSpokeName := claim.Spec.Namespace

			// the ClusterDeployment resides in the cluster-namespace with a matching name
			oldDeploymentSubject := types.NamespacedName{
				Namespace: oldSpokeName,
				Name:      oldSpokeName,
			}

			// fetch ClusterDeployment from previous OLD cluster and delete it if exists
			oldDeployment := &hivev1.ClusterDeployment{}
			if err := r.Client.Get(ctx, oldDeploymentSubject, oldDeployment); err != nil {
				logger.Info(fmt.Sprintf("no ClusterDeployments found on %s", oldSpokeName))
			} else {
				if err = r.Client.Delete(ctx, oldDeployment); err != nil {
					logger.Error(err, fmt.Sprintf("failed deleting ClusterDepolyment %v", oldDeploymentSubject))
				}
			}

			// the ManagedClusterAddon resides in the cluster-namespace
			oldMcaSubject := types.NamespacedName{
				Namespace: oldSpokeName,
				Name:      "multicluster-resiliency-addon",
			}

			// fetch ManagedClusterAddOn from OLD cluster, create a copy in the NEW cluster and delete the OLD one
			oldMca := &addonv1alpha1.ManagedClusterAddOn{}
			if err := r.Client.Get(ctx, oldMcaSubject, oldMca); err != nil {
				logger.Error(err, fmt.Sprintf("failed fetching ManagedClusterAddon %s", oldMcaSubject))
			} else {
				newMca := oldMca.DeepCopy()

				newMca.SetName("multicluster-resiliency-addon")
				newMca.SetNamespace(newSpokeName)

				newMca.SetLabels(oldMca.GetLabels())
				newMca.SetOwnerReferences(oldMca.GetOwnerReferences())
				newMca.SetFinalizers(oldMca.GetFinalizers())
				newMca.SetManagedFields(oldMca.GetManagedFields())

				annotations := oldMca.GetAnnotations()
				annotations[annotationFromAnnotation] = oldSpokeName
				newMca.SetAnnotations(annotations)

				if err = r.Client.Create(ctx, newMca); err != nil {
					logger.Error(err, fmt.Sprintf("failed creating new ManagedClusterAddon in %s", newSpokeName))
				}
			}

			// fetch AddOnDeploymentConfigs from previous OLD cluster and copy them to the NEW one
			oldConfigs := &addonv1alpha1.AddOnDeploymentConfigList{}
			if err := r.Client.List(ctx, oldConfigs, &client.ListOptions{Namespace: oldSpokeName}); err != nil {
				logger.Info(fmt.Sprintf("no AddOnDeploymentConfigs found on %s", oldSpokeName))
			} else {
				for _, oldConfig := range oldConfigs.Items {
					newConfig := oldConfig.DeepCopy()
					newConfig.SetName(oldConfig.Name)
					newConfig.SetNamespace(newSpokeName)
					if err = r.Client.Create(ctx, newConfig); err != nil {
						logger.Error(err, fmt.Sprintf("failed creating AddOnDeploymentConfig %s in %s", newConfig.Name, newSpokeName))
					}
				}
			}

			// when done, remove the finalizer and annotation
			annotations := claim.GetAnnotations()
			delete(annotations, annotationPreviousSpoke)
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
