// Copyright (c) 2023 Red Hat, Inc.

package actions

// This file contains the action for moving AddonDeploymentConfigs between spokes.

import (
	"context"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// migrateAddonDeploymentConfigs is used for moving all AddonDeploymentConfig resources found from the OLD spoke namespace
// to the NEW one, and delete the old ones.
func migrateAddonDeploymentConfigs(ctx context.Context, options Options) {
	logger := log.FromContext(ctx)
	logger.Info("migrating AddOnDeploymentConfig resources", "old-spoke", options.OldSpoke, "new-spoke", options.NewSpoke)

	// fetch AddOnDeploymentConfigs from previous OLD cluster and copy them to the NEW one
	oldConfigs := &addonv1alpha1.AddOnDeploymentConfigList{}
	if err := options.Client.List(ctx, oldConfigs, &client.ListOptions{Namespace: options.OldSpoke}); err != nil {
		logger.Info("no AddOnDeploymentConfigs found", "old-spoke", options.OldSpoke)
	} else {
		// iterate over all found configs, create new ones for the NEW spoke, and delete the OLD ones
		for _, oldConfig := range oldConfigs.Items {
			newConfig := oldConfig.DeepCopy()
			newConfig.SetName(oldConfig.Name)
			newConfig.SetNamespace(options.NewSpoke)

			if err = options.Client.Create(ctx, newConfig); err != nil {
				logger.Error(err, "failed creating new AddOnDeploymentConfig", "new-spoke", options.NewSpoke, "config-name", newConfig.Name)
			}
			if err = options.Client.Delete(ctx, &oldConfig); err != nil {
				logger.Error(err, "failed deleting AddOnDeploymentConfig", "old-spoke", options.OldSpoke, "config-name", oldConfig.Name)
			}
		}
	}
}

// init is registering migrateAddonDeploymentConfigs for running.
func init() {
	actionFuncs = append(actionFuncs, migrateAddonDeploymentConfigs)
}
