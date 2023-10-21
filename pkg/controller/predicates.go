// Copyright (c) 2023 Red Hat, Inc.

package controller

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// verifyOwnerPredicate takes a function and returns a predicate that verifies event object for all events against
// the function.
func verifyOwnerPredicate(fn func(obj client.Object) bool) predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(createEvent event.CreateEvent) bool {
			return fn(createEvent.Object)
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			return fn(deleteEvent.Object)
		},
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			return fn(updateEvent.ObjectOld) && fn(updateEvent.ObjectNew)
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			return fn(genericEvent.Object)
		},
	}
}

// verifyObjectOwnerIsOurAddon is a utility function returning true if one of the owners for a client.Object is this
// Addon. Use it with verifyOwnerPredicate.
func verifyObjectOwnerIsOurAddon(obj client.Object) bool {
	for _, owner := range obj.GetOwnerReferences() {
		if owner.Kind == "ClusterManagementAddOn" && owner.Name == "multicluster-resiliency-addon" {
			return true
		}
	}
	return false
}

// objectOwnerIsAResilientCluster is a utility function returning true if one of the owners kind for a client.Object is
// a ResilientCluster. i.e. it was created by us. Use it with verifyOwnerPredicate.
func objectOwnerIsAResilientCluster(obj client.Object) bool {
	for _, owner := range obj.GetOwnerReferences() {
		if owner.Kind == "ResilientCluster" {
			return true
		}
	}
	return false
}
