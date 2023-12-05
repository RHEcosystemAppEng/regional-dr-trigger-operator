// Copyright (c) 2023 Red Hat, Inc.

package actions

// This file contains the action for copying a ManagedCluster Spec between spokes.

import (
	"context"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// compareManagedClusterAndDeleteOld is used for reconciling ManagedCluster labels and annotations from the
// ManagedCluster representing the OLD spoke into the MangedCluster representing the NEW one. When done, It deletes the
// OLD ManagedCluster.
func compareManagedClusterAndDeleteOld(ctx context.Context, options Options) {
	logger := log.FromContext(ctx)
	logger.Info("comparing ManagedCluster resources", "old-spoke", options.OldSpoke, "new-spoke", options.NewSpoke)

	// fetch the OLD ManagedCluster or break
	oldMc := &clusterv1.ManagedCluster{}
	if err := options.Client.Get(ctx, types.NamespacedName{Name: options.OldSpoke}, oldMc); err != nil {
		logger.Error(err, "failed fetching old ManagedCluster", "old-spoke", options.OldSpoke)
		return
	}

	// create patch object
	mcPatch := &clusterv1.ManagedCluster{}
	mcPatch.Annotations = oldMc.GetAnnotations()
	mcPatch.Labels = oldMc.GetLabels()

	// fetch the NEW ManagedCluster or break
	newMc := &clusterv1.ManagedCluster{}
	if err := options.Client.Get(ctx, types.NamespacedName{Name: options.NewSpoke}, newMc); err != nil {
		logger.Error(err, "failed fetching new ManagedCluster", "new-spoke", options.NewSpoke)
		return
	}

	// patch the NEW ManageCluster with object data from the OLD ManagedCluster
	if err := options.Client.Patch(ctx, newMc, client.StrategicMergeFrom(mcPatch)); err != nil {
		logger.Error(err, "failed patching new ManagedCluster", "new-spoke", options.NewSpoke)
		return
	}

	// delete the OLD ManagedCluster
	if err := options.Client.Delete(ctx, oldMc); err != nil {
		logger.Error(err, "failed deleting old ManagedCluster", "old-spoke", options.OldSpoke)
	}
}

// init is registering compareManagedClusterAndDeleteOld for running.
func init() {
	actionFuncs = append(actionFuncs, compareManagedClusterAndDeleteOld)
}
