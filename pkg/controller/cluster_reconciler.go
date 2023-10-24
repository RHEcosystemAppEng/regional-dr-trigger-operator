// Copyright (c) 2023 Red Hat, Inc.

package controller

// This file hosts our ClusterReconciler implementation registering for our ResilientCluster CRs.

import (
	"context"
	"fmt"
	hivev1 "github.com/openshift/hive/apis/hive/v1"
	apiv1 "github.com/rhecosystemappeng/multicluster-resiliency-addon/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ClusterReconciler is a receiver representing the MultiCluster-Resiliency-Addon operator reconciler for
// ResilientCluster CRs.
type ClusterReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	ConfigMapName string
}

type Config struct {
	HivePoolName string
}

// SetupWithManager is used for setting up the controller named 'mcra-managed-cluster-cluster-controller' with the
// manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
		if controllerutil.ContainsFinalizer(rc, finalizerUsedByMcra) {
			// TODO add cleanup code here

			// when cleanup done, remove the finalizer
			controllerutil.RemoveFinalizer(rc, finalizerUsedByMcra)
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
	config, err := r.loadConfiguration(ctx, rc.Namespace)
	if err != nil {
		logger.Error(err, "unable to load configuration")
		return ctrl.Result{}, err
	}

	pool, err := r.loadClusterPool(ctx, config.HivePoolName)
	if err != nil {
		logger.Error(err, "unable to load hive pool")
		return ctrl.Result{}, err
	}

	if err = verifyPool(pool); err != nil {
		logger.Error(err, "verify hive pool failed")
		return ctrl.Result{}, err
	}

	claimName := fmt.Sprintf("mcra-claim-%s", rand.String(4))
	newClaim := &hivev1.ClusterClaim{}
	newClaim.SetName(claimName)
	newClaim.SetNamespace(config.HivePoolName)
	newClaim.SetAnnotations(map[string]string{annotationPreviousSpoke: rc.Name})

	if err = r.Client.Create(ctx, newClaim); err != nil {
		logger.Error(err, "failed creating ClusterClaim")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// loadConfiguration will first attempt to load the configmap from the cluster-namespace, if failed, will load the one
// from the manager namespace.
func (r *ClusterReconciler) loadConfiguration(ctx context.Context, clusterNamespace string) (Config, error) {
	logger := log.FromContext(ctx)

	subject := types.NamespacedName{
		Namespace: clusterNamespace,
		Name:      r.ConfigMapName,
	}

	cmap := &corev1.ConfigMap{}
	// return configmap from cluster namespace if available
	if err := r.Client.Get(ctx, subject, cmap); err == nil {
		logger.Info("using config from cluster namespace")
		return configMapToConfig(cmap), nil
	}

	logger.Info("using config from manager namespace")
	managerNamespace, exist := os.LookupEnv("POD_NAMESPACE")
	if !exist {
		return Config{}, fmt.Errorf("unable to load manager namespace from POD_NAMESPACE")
	}

	subject.Namespace = managerNamespace
	// load configmap from manager namespace
	if err := r.Client.Get(ctx, subject, cmap); err != nil {
		return Config{}, err
	}

	return configMapToConfig(cmap), nil
}

// loadClusterPool is used to load a ClusterPool, the assumption is that the ClusterPool name and namespace are
// identical.
func (r *ClusterReconciler) loadClusterPool(ctx context.Context, poolName string) (*hivev1.ClusterPool, error) {
	subject := types.NamespacedName{
		Namespace: poolName,
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
		rc.Status.CurrentStatus != rc.Status.PreviousStatus &&
		rc.Status.CurrentStatus != rc.Status.InitialStatus
}

// configMapToConfig is used to extract known keys from a ConfigMap and build a new Config from the extracted values.
// currently we're only working with `hive_pool_name`, but this is where we can add more configuration values.
func configMapToConfig(configMap *corev1.ConfigMap) Config {
	config := Config{}
	if poolName, found := configMap.Data["hive_pool_name"]; found {
		config.HivePoolName = poolName
	}

	return config
}

// verifyPool is used to verify a hivev1.ClusterPool is ok and a ClusterClaim can be made. Initial implementation is
// based on pool's ready status. Further verifications, i.e. checking condition statuses, can be added here.
func verifyPool(pool *hivev1.ClusterPool) error {
	if pool.Status.Ready > 0 {
		return nil
	}
	return fmt.Errorf("cluster pool is not ready for claims")
}
