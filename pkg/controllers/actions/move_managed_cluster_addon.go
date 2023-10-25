// Copyright (c) 2023 Red Hat, Inc.

package actions

import (
	"context"
	"fmt"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/mcra"
	"k8s.io/apimachinery/pkg/types"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func moveManagedClusterAddon(ctx context.Context, options Options) {
	logger := log.FromContext(ctx)

	// the ManagedClusterAddon resides in the cluster-namespace
	oldMcaSubject := types.NamespacedName{
		Namespace: options.OldSpoke,
		Name:      mcra.AddonName,
	}

	// fetch ManagedClusterAddOn from OLD cluster, create a copy in the NEW cluster and delete the OLD one
	oldMca := &addonv1alpha1.ManagedClusterAddOn{}
	if err := options.Client.Get(ctx, oldMcaSubject, oldMca); err != nil {
		logger.Error(err, fmt.Sprintf("failed fetching ManagedClusterAddon %s", oldMcaSubject))
	} else {
		newMca := oldMca.DeepCopy()

		newMca.SetName(mcra.AddonName)
		newMca.SetNamespace(options.NewSpoke)

		newMca.SetLabels(oldMca.GetLabels())
		newMca.SetOwnerReferences(oldMca.GetOwnerReferences())
		newMca.SetFinalizers(oldMca.GetFinalizers())
		newMca.SetManagedFields(oldMca.GetManagedFields())

		annotations := oldMca.GetAnnotations()
		annotations[mcra.AnnotationFromAnnotation] = options.OldSpoke
		newMca.SetAnnotations(annotations)

		if err = options.Client.Create(ctx, newMca); err != nil {
			logger.Error(err, fmt.Sprintf("failed creating new ManagedClusterAddon in %s", options.NewSpoke))
		}

		if err = options.Client.Delete(ctx, oldMca); err != nil {
			logger.Error(err, fmt.Sprintf("failed deleting ManagedClusterAddon from %s", options.OldSpoke))
		}
	}
}

func init() {
	actionFuncs = append(actionFuncs, moveManagedClusterAddon)
}
