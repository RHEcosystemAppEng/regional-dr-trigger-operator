// Copyright (c) 2023 Red Hat, Inc.

package controller

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var _ = Context("Testing predicate functions", func() {
	DescribeTable("Invoking acceptedByHub", func(mc *clusterv1.ManagedCluster, expected bool) {
		Expect(acceptedByHub()(mc)).To(Equal(expected))
	},
		Entry("Should return true if the hub accepted the spoke",
			&clusterv1.ManagedCluster{Spec: clusterv1.ManagedClusterSpec{HubAcceptsClient: true}},
			true),
		Entry("Should return false if the hub did not accept the spoke",
			&clusterv1.ManagedCluster{Spec: clusterv1.ManagedClusterSpec{HubAcceptsClient: false}},
			false),
		Entry("Should return false if the hub did not report accepts status",
			&clusterv1.ManagedCluster{},
			false),
	)

	DescribeTable("Invoking joinedHub", func(mc *clusterv1.ManagedCluster, expected bool) {
		Expect(joinedHub()(mc)).To(Equal(expected))
	},
		Entry("Should return true if the spoke has joined the hub",
			&clusterv1.ManagedCluster{
				Status: clusterv1.ManagedClusterStatus{
					Conditions: []metav1.Condition{
						{Type: clusterv1.ManagedClusterConditionJoined, Status: metav1.ConditionTrue},
					},
				},
			},
			true),
		Entry("Should return false if the spoke had not joined the hub",
			&clusterv1.ManagedCluster{
				Status: clusterv1.ManagedClusterStatus{
					Conditions: []metav1.Condition{
						{Type: clusterv1.ManagedClusterConditionJoined, Status: metav1.ConditionFalse},
					},
				},
			},
			false),
		Entry("Should return false if the spoke did not report join status",
			&clusterv1.ManagedCluster{
				Status: clusterv1.ManagedClusterStatus{
					Conditions: []metav1.Condition{},
				},
			},
			false),
	)

	DescribeTable("Invoking notAvailable", func(mc *clusterv1.ManagedCluster, expected bool) {
		Expect(notAvailable()(mc)).To(Equal(expected))
	},
		Entry("Should return true if the spoke it not available",
			&clusterv1.ManagedCluster{
				Status: clusterv1.ManagedClusterStatus{
					Conditions: []metav1.Condition{
						{Type: clusterv1.ManagedClusterConditionAvailable, Status: metav1.ConditionFalse},
					},
				},
			},
			true),
		Entry("Should return false if the spoke is available",
			&clusterv1.ManagedCluster{
				Status: clusterv1.ManagedClusterStatus{
					Conditions: []metav1.Condition{
						{Type: clusterv1.ManagedClusterConditionAvailable, Status: metav1.ConditionTrue},
					},
				},
			},
			false),
		Entry("Should return true if the spoke did not report availability",
			&clusterv1.ManagedCluster{
				Status: clusterv1.ManagedClusterStatus{
					Conditions: []metav1.Condition{},
				},
			},
			true),
	)
})

var _ = Context("Testing the predicates interface", func() {
	dummyFunc := func(obj client.Object) bool { return true }
	dummyMc := &clusterv1.ManagedCluster{}
	// create Predicates with dummy function
	predicates := verifyManagedCluster(dummyFunc)

	Specify("Create events should be enabled", func() {
		Expect(predicates.Create(event.CreateEvent{Object: dummyMc})).To(BeTrue())
	})

	Specify("Update events should be enabled", func() {
		Expect(predicates.Update(event.UpdateEvent{ObjectOld: dummyMc, ObjectNew: dummyMc})).To(BeTrue())
	})

	Specify("Generic events should be enabled", func() {
		Expect(predicates.Generic(event.GenericEvent{Object: dummyMc})).To(BeTrue())
	})

	Specify("Delete events should be disabled", func() {
		Expect(predicates.Delete(event.DeleteEvent{Object: dummyMc, DeleteStateUnknown: false})).To(BeFalse())
	})
})
