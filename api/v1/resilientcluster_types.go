// Copyright (c) 2023 Red Hat, Inc.

// +k8s:deepcopy-gen=package
package v1

// This file hosts the API types for K8s.

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// +groupName=appeng.ecosystem.redhat.com

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "appeng.ecosystem.redhat.com", Version: "v1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// Install adds the types in this group-version to the given scheme.
	Install = SchemeBuilder.AddToScheme
)

type (
	ResilientClusterSpec struct {
	}

	ResilientClusterStatus struct {
		// Conditions describes the status of the monitored Spoke Cluster
		Conditions []metav1.Condition `json:"conditions,omitempty"`
	}

	// +kubebuilder:object:root=true
	// +kubebuilder:subresource:status
	// +kubebuilder:resource:scope=Namespaced,shortName=rstc

	// ResilientCluster is used by the MultiCluster-Resiliency-Addon for maintain the status and state of each cluster
	// running the Addon Agent. CRs should live in the relevant cluster-namespaces. One per Spoke, named after the
	// cluster it represents.
	ResilientCluster struct {
		metav1.TypeMeta   `json:",inline"`
		metav1.ObjectMeta `json:"metadata,omitempty"`

		Spec   ResilientClusterSpec   `json:"spec,omitempty"`
		Status ResilientClusterStatus `json:"status,omitempty"`
	}

	//+kubebuilder:object:root=true

	ResilientClusterList struct {
		metav1.TypeMeta `json:",inline"`
		metav1.ListMeta `json:"metadata,omitempty"`
		Items           []ResilientCluster `json:"items"`
	}
)

func init() {
	SchemeBuilder.Register(&ResilientCluster{}, &ResilientClusterList{})
}
