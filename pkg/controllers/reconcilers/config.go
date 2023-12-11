// Copyright (c) 2023 Red Hat, Inc.

package reconcilers

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// This file contains utility functions for loading the configuration for use with the various controllers.

type Config struct {
	HivePoolName string
}

// loadConfiguration will first attempt to load the configmap from the cluster-namespace, if failed, will load the one
// from the manager namespace.
func loadConfiguration(ctx context.Context, c client.Client, configName, clusterNamespace, managerNamespace string) (Config, error) {
	logger := log.FromContext(ctx)

	subject := types.NamespacedName{
		Namespace: clusterNamespace,
		Name:      configName,
	}

	cmap := &corev1.ConfigMap{}
	// return configmap from cluster namespace if available
	if err := c.Get(ctx, subject, cmap); err == nil {
		logger.Info("using config from cluster namespace")
		return configMapToConfig(cmap), nil
	}

	logger.Info("using config from manager namespace")
	subject.Namespace = managerNamespace
	// load configmap from manager namespace
	if err := c.Get(ctx, subject, cmap); err != nil {
		return Config{}, err
	}

	return configMapToConfig(cmap), nil
}

// configMapToConfig is used for extracting known keys from a ConfigMap and build a new Config from the extracted values.
func configMapToConfig(configMap *corev1.ConfigMap) Config {
	config := Config{}
	if poolName, found := configMap.Data["hive_pool_name"]; found {
		config.HivePoolName = poolName
	}

	return config
}
