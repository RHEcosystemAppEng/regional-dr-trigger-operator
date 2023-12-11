// Copyright (c) 2023 Red Hat, Inc.

package reconcilers

// This file hosts the ClusterReconciler implementation registering for ResilientCluster CRs.

import (
	"context"
	"fmt"
	hivev1 "github.com/openshift/hive/apis/hive/v1"
	apiv1 "github.com/rhecosystemappeng/multicluster-resiliency-addon/api/v1"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/mcra"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/metrics"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// ClusterReconciler is a receiver representing the MultiCluster-Resiliency-Addon operator reconciler for
// ResilientCluster CRs.
type ClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Options
}

// setupWithManager is used for setting up the controller named 'mcra-managed-cluster-cluster-controller' with the
// manager.
func (r *ClusterReconciler) setupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("mcra-cluster-controller").
		For(&apiv1.ResilientCluster{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=appeng.ecosystem.redhat.com,resources=resilientclusters,verbs=*
// +kubebuilder:rbac:groups=appeng.ecosystem.redhat.com,resources=resilientclusters/finalizer,verbs=*
// +kubebuilder:rbac:groups=appeng.ecosystem.redhat.com,resources=resilientclusterclaimbinding,verbs=*
// +kubebuilder:rbac:groups=appeng.ecosystem.redhat.com,resources=resilientclusterclaimbinding/finalizer,verbs=*
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=addondeploymentconfigs,verbs=*

// Reconcile is watching ResilientCluster CRs, determining whether a new Spoke cluster is required, and handling
// the cluster provisioning using OpenShift Hive API. Note, further permissions are listed in AddonReconciler.Reconcile
// and AddonReconciler.Reconcile.
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
		if controllerutil.ContainsFinalizer(rc, mcra.FinalizerResilientClusterCleanup) {
			// TODO add cleanup code here

			// when cleanup done, remove the finalizer
			controllerutil.RemoveFinalizer(rc, mcra.FinalizerResilientClusterCleanup)
			if err := r.Client.Update(ctx, rc); err != nil {
				logger.Error(err, fmt.Sprintf("%s failed removing finalizer", subject.String()))
				return ctrl.Result{}, err
			}
		}

		// no further progress is required while deleting objects
		return ctrl.Result{}, nil
	}

	// decide whether a new ClusterClaim is required based on ResilientCluster status
	if !requiresNewClaim(rc) {
		logger.Info("no claim required")
		return ctrl.Result{}, nil
	}

	logger.Info(fmt.Sprintf("cluster %s requires a new claim", rc.Name))

	managerNamespace, exist := os.LookupEnv("POD_NAMESPACE")
	if !exist {
		return ctrl.Result{}, fmt.Errorf("unable to load manager namespace from POD_NAMESPACE")
	}

	config, err := loadConfiguration(ctx, r.Client, r.ConfigMapName, rc.Namespace, managerNamespace)
	if err != nil {
		logger.Error(err, "unable to load configuration")
		return ctrl.Result{}, err
	}

	pool, err := r.loadClusterPool(ctx, config.HivePoolName, managerNamespace)
	if err != nil {
		logger.Error(err, "unable to load hive pool")
		return ctrl.Result{}, err
	}

	if err = verifyPool(pool); err != nil {
		logger.Error(err, "verify hive pool failed")
		return ctrl.Result{Requeue: true}, err
	}

	claimName := fmt.Sprintf("mcra-claim-%s", rand.String(4))
	newClaim := &hivev1.ClusterClaim{}
	newClaim.SetName(claimName)
	newClaim.SetNamespace(config.HivePoolName)
	newClaim.SetAnnotations(map[string]string{
		mcra.AnnotationCreatedBy:     mcra.AddonName,
		mcra.AnnotationPreviousSpoke: req.Namespace,
	})
	newClaim.Spec = hivev1.ClusterClaimSpec{ClusterPoolName: config.HivePoolName}

	if err = r.Client.Create(ctx, newClaim); err != nil {
		logger.Error(err, "failed creating ClusterClaim")
		return ctrl.Result{}, err
	}

	// the namespace is the name of the old spoke
	metrics.NewClusterClaimCreated.WithLabelValues(config.HivePoolName, claimName, req.Namespace).Inc()

	return ctrl.Result{}, nil
}

// loadClusterPool is used for loading a ClusterPool from the manager's namespace.
func (r *ClusterReconciler) loadClusterPool(ctx context.Context, poolName, managerNamespace string) (*hivev1.ClusterPool, error) {
	subject := types.NamespacedName{
		Namespace: managerNamespace,
		Name:      poolName,
	}

	pool := &hivev1.ClusterPool{}
	return pool, r.Client.Get(ctx, subject, pool)
}

// requiresNewClaim takes an apiv1.ResilientCluster and determines whether a new cluster claim is required. i.e. If the
// cluster is not available, a new claim is required. Currently, the decision is made based on the availability status,
// for future steps we can make this more robust. For instance, check the time of the previous status change and only
// require a new claim if x time has passed, allowing the cluster a change to recuperate.
func requiresNewClaim(rc *apiv1.ResilientCluster) bool {
	return rc.Status.CurrentStatus.Availability != apiv1.ClusterAvailable &&
		rc.Status.PreviousStatus.Availability == apiv1.ClusterAvailable
}

// verifyPool is used for verifying a hivev1.ClusterPool is ok and a ClusterClaim can be made. Initial implementation is
// based on pool's ready status. Further verifications, i.e. checking condition statuses, can be added here.
func verifyPool(pool *hivev1.ClusterPool) error {
	if pool.Status.Size > 0 {
		return nil
	}
	return fmt.Errorf("cluster pool is not ready for claims")
}

// init is registering the ClusterReconciler setup function for execution.
func init() {
	reconcilerFuncs = append(reconcilerFuncs, func(mgr manager.Manager, options Options) error {
		return (&ClusterReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme(), Options: options}).setupWithManager(mgr)
	})
}
