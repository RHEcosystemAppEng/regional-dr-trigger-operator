// Copyright (c) 2023 Red Hat, Inc.

package actions

// This file contains the action for copying a ManagedCluster Spec between spokes.

import (
	"context"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/mcra"
	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// compareManagedClusterAndDeleteOld is used for reconciling ManagedCluster spec, labels, annotations, finalizers, and
// owner references, from the ManagedCluster representing the OLD spoke into the MangedCluster representing the NEW one.
// When done, It deletes the OLD ManagedCluster.
func compareManagedClusterAndDeleteOld(ctx context.Context, options Options) {
	logger := log.FromContext(ctx)
	logger.Info("comparing ManagedCluster resources", "old-spoke", options.OldSpoke, "new-spoke", options.NewSpoke)

	// fetch the OLD ManagedCluster or break
	oldMc := &clusterv1.ManagedCluster{}
	if err := options.Client.Get(ctx, types.NamespacedName{Name: options.OldSpoke}, oldMc); err != nil {
		logger.Error(err, "failed fetching old ManagedCluster", "old-spoke", options.OldSpoke)
		return
	}

	// fetch the NEW ManagedCluster or break
	newMc := &clusterv1.ManagedCluster{}
	if err := options.Client.Get(ctx, types.NamespacedName{Name: options.NewSpoke}, newMc); err != nil {
		logger.Error(err, "failed fetching new ManagedCluster", "new-spoke", options.NewSpoke)
		return
	}

	// compare the NEW ManagedCluster with the OLD one or break
	labels := newMc.GetLabels()
	maps.Copy(labels, oldMc.GetLabels())
	newMc.SetLabels(labels)

	annotations := oldMc.GetAnnotations()
	annotations[mcra.AnnotationCreatedBy] = mcra.AddonName
	newMc.SetAnnotations(annotations)

	if err := options.Client.Update(ctx, newMc); err != nil {
		logger.Error(err, "failed updating new ManagedCluster", "new-spoke", options.NewSpoke)
		return
	}

	if err := options.Client.Delete(ctx, oldMc); err != nil {
		logger.Error(err, "failed deleting old ManagedCluster", "old-spoke", options.OldSpoke)
	}
}

// init is registering compareManagedClusterAndDeleteOld for running.
func init() {
	actionFuncs = append(actionFuncs, compareManagedClusterAndDeleteOld)
}
