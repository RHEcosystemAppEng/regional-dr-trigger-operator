// Copyright (c) 2023 Red Hat, Inc.

// +k8s:deepcopy-gen=package
package v1

// This file hosts the API types for K8s and generation instructions for generating manifest with controller-gen.

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// +groupName=appeng.ecosystem.redhat.com

var (
	// groupVersion is group version used to register the objects in this file.
	groupVersion = schema.GroupVersion{Group: "appeng.ecosystem.redhat.com", Version: "v1"}

	// schemeBuilder is used to add go types to the GroupVersionKind scheme.
	schemeBuilder = &scheme.Builder{GroupVersion: groupVersion}

	// Install adds the types in this group-version to the given scheme.
	Install = schemeBuilder.AddToScheme
)

type (
	// ClustarAvailability is a bool representing whether not the Spoke cluster is available. Use ClusterAvailable and
	// ClusterNotAvailable.
	ClustarAvailability string

	// ClusterStatus represents a status of the Spoke cluster at a specific time.
	ClusterStatus struct {
		// +kubebuilder:validation:Enum=True;False
		Availability ClustarAvailability `json:"availability,omitempty"`
		Time         metav1.Time         `json:"time,omitempty"`
	}

	// ClaimInfo encapsulates the info for the related ClusterClaim
	ClaimInfo struct {
		Name string      `json:"name,omitempty"`
		Time metav1.Time `json:"time,omitempty"`
	}

	// ResilientClusterStatus encapsulated the initial, current, and previous statuses of the ResilientCluster.
	ResilientClusterStatus struct {
		InitialStatus  ClusterStatus `json:"initialStatus"`
		CurrentStatus  ClusterStatus `json:"currentStatus"`
		PreviousStatus ClusterStatus `json:"previousStatus,omitempty"`
		CurrentClaim   ClaimInfo     `json:"currentClaim,omitempty"`
		PreviousClaim  ClaimInfo     `json:"previousClaim,omitempty"`
	}

	// ResilientCluster is used by the MultiCluster-Resiliency-Addon for maintain the status and state of each cluster
	// running the Addon Agent. CRs should live in the relevant cluster-namespaces. One per Spoke, named after the
	// cluster it represents.
	//
	// +kubebuilder:object:root=true
	// +kubebuilder:resource:scope=Namespaced,shortName=rstc
	// +kubebuilder:printcolumn:name=Available,type=string,JSONPath=`.status.currentStatus.availability`
	// +kubebuilder:printcolumn:name=Claim,type=string,JSONPath=`.status.currentClaim.name`
	ResilientCluster struct {
		metav1.TypeMeta   `json:",inline"`
		metav1.ObjectMeta `json:"metadata,omitempty"`
		Status            ResilientClusterStatus `json:"status"`
	}

	// ResilientClusterList is a List resource for ResilientCluster resources.
	//
	// +kubebuilder:object:root=true
	ResilientClusterList struct {
		metav1.TypeMeta `json:",inline"`
		metav1.ListMeta `json:"metadata,omitempty"`
		Items           []ResilientCluster `json:"items"`
	}
)

const (
	ClusterAvailable    ClustarAvailability = "True"
	ClusterNotAvailable ClustarAvailability = "False"
)

// init is used to register the Addon API types with the scheme previously configured with groupVersion.
func init() {
	schemeBuilder.Register(&ResilientCluster{}, &ResilientClusterList{})
}
