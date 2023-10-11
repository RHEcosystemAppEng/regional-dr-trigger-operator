// Copyright (c) 2023 Red Hat, Inc.

package v1

// This file hosts the API types for K8s.

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// +groupName=appeng.ecosystem.redhat.com
// +k8s:deepcopy-gen=package

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "appeng.ecosystem.redhat.com", Version: "v1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// Install adds the types in this group-version to the given scheme.
	Install = SchemeBuilder.AddToScheme
)

type (
	// ResilientClusterSpec defines the specification for creating a ResilientCluster
	ResilientClusterSpec struct {
		// StatusHistory represents the number of previous ClusterStatus CRs we should take into consideration for
		// determining the Spoke cluster status
		// +kubebuilder:validation:Minimum=1
		// +kubebuilder:validation:Maximum=5
		// +kubebuilder:default=1
		StatusHistory int `json:"statusHistory,omitempty"`
	}

	// ClustarAvailability is a bool representing whether not the Spoke cluster is available
	ClustarAvailability bool

	// ClusterStatus represents a status of the Spoke cluster at a specific time
	ClusterStatus struct {
		Availability ClustarAvailability `json:"availability"`
		Time         metav1.Time         `json:"time"`
	}

	// ResilientClusterStatus defines the current status of the Spoke cluster.
	ResilientClusterStatus struct {
		InitialStatus *ClusterStatus `json:"initialStatus"`

		CurrentStatus *ClusterStatus `json:"currentStatus"`

		// +kubebuilder:validation:MinItems=1
		// +kubebuilder:validation:MaxItems=5
		// +kubebuilder:validation:UniqueItems=true
		PreviousStatuses []*ClusterStatus `json:"previousStatuses"`
	}

	// ResilientCluster is used by the MultiCluster-Resiliency-Addon for maintain the status and state of each cluster
	// running the Addon Agent. CRs should live in the relevant cluster-namespaces. One per Spoke, named after the
	// cluster it represents.
	// +kubebuilder:object:root=true
	// +kubebuilder:resource:scope=Namespaced,shortName=rstc
	ResilientCluster struct {
		metav1.TypeMeta   `json:",inline"`
		metav1.ObjectMeta `json:"metadata,omitempty"`

		Spec   ResilientClusterSpec   `json:"spec,omitempty"`
		Status ResilientClusterStatus `json:"status"`
	}

	// ResilientClusterList is a List resource for ResilientCluster resources.
	// +kubebuilder:object:root=true
	ResilientClusterList struct {
		metav1.TypeMeta `json:",inline"`
		metav1.ListMeta `json:"metadata,omitempty"`
		Items           []ResilientCluster `json:"items"`
	}
)

const (
	ClusterAvailable    ClustarAvailability = true
	ClusterNotAvailable ClustarAvailability = false
)

func init() {
	SchemeBuilder.Register(&ResilientCluster{}, &ResilientClusterList{})
}
