// Copyright (c) 2023 Red Hat, Inc.

package reconcilers

// This file hosts the ClaimReconciler implementation registering for Hive's ClusterClaim CRs.

import (
	"context"
	"fmt"
	hivev1 "github.com/openshift/hive/apis/hive/v1"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/controllers/actions"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/mcra"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/metrics"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// ClaimReconciler is a receiver representing the MultiCluster-Resiliency-Addon operator reconciler for
// ClusterClaim CRs.
type ClaimReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Options
}

// setupWithManager is used for setting up the controller named 'mcra-managed-cluster-claim-controller' with the manager.
// It uses predicates as event filters for verifying only handling ClusterClaim CRs created by us.
func (r *ClaimReconciler) setupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("mcra-claim-controller").
		For(&hivev1.ClusterClaim{}).
		WithEventFilter(verifyObject(hasAnnotation(mcra.AnnotationPreviousSpoke))).
		Complete(r)
}

// +kubebuilder:rbac:groups=hive.openshift.io,resources=clusterpools,verbs=get;list;watch
// +kubebuilder:rbac:groups=hive.openshift.io,resources=clusterclaims,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups=hive.openshift.io,resources=clusterclaims/finalizers,verbs=update
// +kubebuilder:rbac:groups=hive.openshift.io,resources=clusterdeployments,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups=cluster.open-cluster-management.io,resources=managedclusters,verbs=get;list;watch;create;update;delete

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

	// decide the status of the ClusterClaim, we require both conditions to exist and have the appropriate values
	running := false
	pending := true
	for _, condition := range claim.Status.Conditions {
		switch condition.Type {
		case hivev1.ClusterRunningCondition:
			if condition.Status == corev1.ConditionTrue {
				running = true // cluster is running
			}
		case hivev1.ClusterClaimPendingCondition:
			if condition.Status == corev1.ConditionFalse {
				pending = false // cluster is not pending
			}
		}
	}

	// verify decided status, requeue if not done
	if pending || !running {
		logger.Info("claim is not done yet done")
		return ctrl.Result{Requeue: true}, nil
	}

	// the OLD spoke name was set as an annotation when we created the ClusterClaim in ClusterReconciler
	oldSpokeName := claim.GetAnnotations()[mcra.AnnotationPreviousSpoke]
	// the NEW spoke name is the target namespace in which the ClusterDeployment was created
	newSpokeName := claim.Spec.Namespace

	// perform all actions required for replacing a cluster
	actions.PerformReplace(ctx, actions.Options{Client: r.Client, OldSpoke: oldSpokeName, NewSpoke: newSpokeName, ConfigMapName: r.Options.ConfigMapName})

	// when done, remove the annotation
	annotations := claim.GetAnnotations()
	delete(annotations, mcra.AnnotationPreviousSpoke)
	claim.SetAnnotations(annotations)

	if err := r.Client.Update(ctx, claim); err != nil {
		logger.Error(err, fmt.Sprintf("%s failed removing finalizer and annotation from claim", claimSubject.String()))
		return ctrl.Result{}, err
	}

	metrics.NewSpokeReady.WithLabelValues(oldSpokeName, newSpokeName).Inc()

	return ctrl.Result{}, nil
}

// init is registering the ClaimReconciler setup function for execution.
func init() {
	reconcilerFuncs = append(reconcilerFuncs, func(mgr manager.Manager, options Options) error {
		return (&ClaimReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme(), Options: options}).setupWithManager(mgr)
	})
}
