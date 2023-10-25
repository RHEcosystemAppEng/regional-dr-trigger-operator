// Copyright (c) 2023 Red Hat, Inc.

package actions

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func moveConfigMap(ctx context.Context, options Options) {
	logger := log.FromContext(ctx)

	// the ConfigMap resides in the cluster-namespace
	oldConfigSubject := types.NamespacedName{
		Namespace: options.OldSpoke,
		Name:      options.ConfigName,
	}

	// fetch ConfigMap from OLD cluster, create a copy in the NEW cluster and delete the OLD one
	oldConfig := &corev1.ConfigMap{}
	if err := options.Client.Get(ctx, oldConfigSubject, oldConfig); err != nil {
		logger.Error(err, fmt.Sprintf("failed fetching ManagedClusterAddon %s", oldConfigSubject))
	} else {
		newConfig := oldConfig.DeepCopy()

		newConfig.SetName(options.ConfigName)
		newConfig.SetNamespace(options.NewSpoke)

		newConfig.SetLabels(oldConfig.GetLabels())
		newConfig.SetOwnerReferences(oldConfig.GetOwnerReferences())
		newConfig.SetFinalizers(oldConfig.GetFinalizers())
		newConfig.SetManagedFields(oldConfig.GetManagedFields())
		newConfig.SetAnnotations(oldConfig.GetAnnotations())

		if err = options.Client.Create(ctx, newConfig); err != nil {
			logger.Error(err, fmt.Sprintf("failed creating new ConfigMap in %s", options.NewSpoke))
		}

		if err = options.Client.Delete(ctx, oldConfig); err != nil {
			logger.Error(err, fmt.Sprintf("failed deleting ConfigMap from %s", options.OldSpoke))
		}
	}
}

func init() {
	actionFuncs = append(actionFuncs, moveConfigMap)
}
