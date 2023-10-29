// Copyright (c) 2023 Red Hat, Inc.

package actions

// This file contains the action for moving the ConfigMap between spokes.

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// migrateConfigMap is used for moving our ConfigMap from the OLD spoke to the NEW one.
func migrateConfigMap(ctx context.Context, options Options) {
	logger := log.FromContext(ctx)
	logger.Info("migrating ConfigMap resource", "old-spoke", options.OldSpoke, "new-spoke", options.NewSpoke, "config-name", options.ConfigMapName)

	// the ConfigMap resides in the cluster-namespace
	oldConfigSubject := types.NamespacedName{
		Namespace: options.OldSpoke,
		Name:      options.ConfigMapName,
	}

	// fetch ConfigMap from OLD cluster, create a copy in the NEW cluster and delete the OLD one
	oldConfig := &corev1.ConfigMap{}
	if err := options.Client.Get(ctx, oldConfigSubject, oldConfig); err != nil {
		logger.Info("no ConfigMap found", "old-spoke", options.OldSpoke, "config-name", options.ConfigMapName)
	} else {
		// create new config for NEW spoke and delete the OLD one
		newConfig := oldConfig.DeepCopy()

		newConfig.SetName(options.ConfigMapName)
		newConfig.SetNamespace(options.NewSpoke)
		newConfig.SetAnnotations(oldConfig.GetAnnotations())
		newConfig.SetFinalizers(oldConfig.GetFinalizers())
		newConfig.SetOwnerReferences(oldConfig.GetOwnerReferences())
		newConfig.SetLabels(oldConfig.GetLabels())

		if err = options.Client.Create(ctx, newConfig); err != nil {
			logger.Error(err, "failed creating new ConfigMap", "new-spoke", options.NewSpoke, "config-name", options.ConfigMapName)
		}

		if err = options.Client.Delete(ctx, oldConfig); err != nil {
			logger.Error(err, "failed deleting ConfigMap from", "old-spoke", options.OldSpoke, "config-name", options.ConfigMapName)
		}
	}
}

// init is registering migrateConfigMap for running.
func init() {
	actionFuncs = append(actionFuncs, migrateConfigMap)
}
