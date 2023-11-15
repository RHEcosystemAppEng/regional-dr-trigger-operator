// Copyright (c) 2023 Red Hat, Inc.

package actions

// This file contains the action for moving the ManagedClusterAddon between spokes.

import (
	"context"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/mcra"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// migrateManagedClusterAddon is used for moving the ManagedClusterAddon from the OLD spoke to the NEW one.
func migrateManagedClusterAddon(ctx context.Context, options Options) {
	logger := log.FromContext(ctx)
	logger.Info("migrating ManagedClusterAddon resource", "old-spoke", options.OldSpoke, "new-spoke", options.NewSpoke)

	// the ManagedClusterAddon resides in the cluster-namespace
	oldMcaSubject := types.NamespacedName{
		Namespace: options.OldSpoke,
		Name:      mcra.AddonName,
	}

	// fetch ManagedClusterAddOn from OLD cluster, create a copy in the NEW cluster and delete the OLD one
	oldMca := &addonv1alpha1.ManagedClusterAddOn{}
	if err := options.Client.Get(ctx, oldMcaSubject, oldMca); err != nil {
		logger.Error(err, "failed fetching ManagedClusterAddon", "old-spoke", options.OldSpoke)
		return
	}

	// the ManagedClusterAddon resides in the cluster-namespace
	newMcaSubject := types.NamespacedName{
		Namespace: options.NewSpoke,
		Name:      mcra.AddonName,
	}

	// attempt to fetch ManagedClusterAddOn from NEW cluster, if exists - compare spec, if not create new
	newMca := &addonv1alpha1.ManagedClusterAddOn{}
	if err := options.Client.Get(ctx, newMcaSubject, newMca); err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "failed fetching ManagedClusterAddon", "new-spoke", options.NewSpoke)
			return
		}

		// create new ManagedClusterAddon for the NEW spoke and
		if err = createNewMca(ctx, oldMca, options); err != nil {
			logger.Error(err, "failed updating new ManagedClusterAddon", "new-spoke", options.NewSpoke)
		}

	} else {
		// compare new ManagedClusterAddon with previous one (new one created by addon install strategy)
		if err = updateNewMca(ctx, newMca, oldMca, options); err != nil {
			logger.Error(err, "failed creating new ManagedClusterAddon", "new-spoke", options.NewSpoke)
		}
	}

	// delete the OLD one Spoke
	if err := options.Client.Delete(ctx, oldMca); err != nil {
		logger.Error(err, "failed deleting ManagedClusterAddon", "old-spoke", options.OldSpoke)
	}
}

// createNewMca is used for creating a new ManagedClusterAddon from a previous one
func createNewMca(ctx context.Context, oldMca *addonv1alpha1.ManagedClusterAddOn, options Options) error {
	newMca := oldMca.DeepCopy()
	newMca.ObjectMeta = metav1.ObjectMeta{}

	newMca.SetName(mcra.AddonName)
	newMca.SetNamespace(options.NewSpoke)
	newMca.SetLabels(oldMca.GetLabels())

	newMca.SetOwnerReferences(oldMca.GetOwnerReferences())
	newMca.SetFinalizers(oldMca.GetFinalizers())

	annotations := oldMca.GetAnnotations()
	annotations[mcra.AnnotationCreatedBy] = mcra.AddonName
	annotations[mcra.AnnotationFromAnnotation] = options.OldSpoke
	newMca.SetAnnotations(annotations)

	return options.Client.Create(ctx, newMca)
}

// updateNewMca is used for updating a pre-existing ManagedClusterAddon from a previous one. Presumably the new once was
// created by the addon when configured to install-all-strategy.
func updateNewMca(ctx context.Context, newMca, oldMca *addonv1alpha1.ManagedClusterAddOn, options Options) error {
	labels := newMca.GetLabels()
	maps.Copy(labels, oldMca.GetLabels())
	newMca.SetLabels(labels)

	owners := newMca.GetOwnerReferences()
	for _, oldOwner := range oldMca.GetOwnerReferences() {
		if !slices.Contains(owners, oldOwner) {
			owners = append(owners, oldOwner)
		}
	}
	newMca.SetOwnerReferences(owners)

	finalizers := newMca.GetFinalizers()
	for _, oldFinalizer := range oldMca.GetFinalizers() {
		if !slices.Contains(finalizers, oldFinalizer) {
			finalizers = append(finalizers, oldFinalizer)
		}
	}

	annotations := newMca.GetAnnotations()
	maps.Copy(annotations, oldMca.GetAnnotations())
	annotations[mcra.AnnotationCreatedBy] = mcra.AddonName
	annotations[mcra.AnnotationFromAnnotation] = options.OldSpoke
	newMca.SetAnnotations(annotations)

	return options.Client.Update(ctx, newMca)
}

// init is registering migrateManagedClusterAddon for running.
func init() {
	actionFuncs = append(actionFuncs, migrateManagedClusterAddon)
}
