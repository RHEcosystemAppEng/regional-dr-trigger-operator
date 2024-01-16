// Copyright (c) 2023 Red Hat, Inc.

package controller

// This file contains predicates related functions for use as filters for controller events.

import (
	"k8s.io/apimachinery/pkg/api/meta"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// verifyManagedCluster takes a function and returns a predicate that verifies the event object for all events against
// the function.
// NOTE: PREVENTING ALL DELETE EVENTS!
func verifyManagedCluster(fn func(obj client.Object) bool) predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(createEvent event.CreateEvent) bool {
			return fn(createEvent.Object)
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			//return fn(deleteEvent.Object)
			return false
		},
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			return fn(updateEvent.ObjectOld)
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			return fn(genericEvent.Object)
		},
	}
}

func acceptedByHub() func(obj client.Object) bool {
	return func(obj client.Object) bool {
		return obj.(*clusterv1.ManagedCluster).Spec.HubAcceptsClient
	}
}

func joinedHub() func(obj client.Object) bool {
	return func(obj client.Object) bool {
		mc := obj.(*clusterv1.ManagedCluster)
		return meta.IsStatusConditionTrue(mc.Status.Conditions, clusterv1.ManagedClusterConditionJoined)
	}
}

func notAvailable() func(obj client.Object) bool {
	return func(obj client.Object) bool {
		mc := obj.(*clusterv1.ManagedCluster)
		return !meta.IsStatusConditionTrue(mc.Status.Conditions, clusterv1.ManagedClusterConditionAvailable)
	}
}
