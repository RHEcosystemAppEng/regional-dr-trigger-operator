// Copyright (c) 2023 Red Hat, Inc.

package controller

// This file hosts tests cases for unit testing the reconciliation loop.

import (
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ramenv1alpha1 "github.com/ramendr/ramen/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Context("Testing reconciliation", func() {
	// Testing Plan:
	// - Two testing Namespaces. Each will have one DRPlacementControl declared with the Relocate.
	// - Two Spoke clusters. One of them will be used as a "failed" cluster.
	//
	// Note: events for unavailable clusters are cherry-picked using Predicates,
	// this means that every reconciliation loop invocation assumes unavailability, no patching required.
	//
	// Note: the events source is a ManagedCluster, but we only need its name, the spoke name, which we get from the
	// request. No patching of the ManagedCluster required.
	//
	// An eligible for failover DR application, is a DRPlacementControl matching the following conditions:
	// - The PreferredCluster is the event cluster
	// - The Phase in one of Deploying, Deployed, or Relocated
	// - Has Condition PeerAvailable set to True
	//
	// The test scenarios in this file, will run various cases of eligible and not eligible clusters.
	//
	// See reconciler_suite_test for testing data, utilities, and environment setup and teardown.

	targetSpoke1 := "spoke1"
	targetSpoke2 := "spoke2"

	drControlNs1BP := ramenv1alpha1.DRPlacementControl{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ns1-test-control",
			Namespace: testNamespace1.Name,
		},
		Spec: ramenv1alpha1.DRPlacementControlSpec{
			Action: ramenv1alpha1.ActionRelocate,
		},
	}
	var drControlNs1 *ramenv1alpha1.DRPlacementControl

	drControlNs2BP := &ramenv1alpha1.DRPlacementControl{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "n2-test-control",
			Namespace: testNamespace2.Name,
		},
		Spec: ramenv1alpha1.DRPlacementControlSpec{
			Action: ramenv1alpha1.ActionRelocate,
		},
	}
	var drControlNs2 *ramenv1alpha1.DRPlacementControl

	BeforeEach(func(ctx SpecContext) {
		// create testing objects before each test
		drControlNs1 = drControlNs1BP.DeepCopy()
		drControlNs2 = drControlNs2BP.DeepCopy()

		Expect(testClient.Create(ctx, drControlNs1)).To(Succeed())
		Expect(testClient.Create(ctx, drControlNs2)).To(Succeed())
	})

	AfterEach(func(ctx SpecContext) {
		// delete testing objects after each test
		Expect(testClient.Delete(ctx, drControlNs1)).To(Succeed())
		Expect(testClient.Delete(ctx, drControlNs2)).To(Succeed())
	})

	Specify("Only patch applications preferring the cluster triggering the event", func(ctx SpecContext) {
		// patch application on the first namespace to be READY for failover and prefer SPOKE1
		drControlNs1Patch, err := json.Marshal(statusForPatching(targetSpoke1, ramenv1alpha1.Deployed, metav1.ConditionTrue))
		Expect(err).NotTo(HaveOccurred())
		Expect(testClient.Status().Patch(ctx, drControlNs1, client.RawPatch(types.MergePatchType, drControlNs1Patch))).To(Succeed())

		// patch application on the second namespace to be READY for failover and prefer SPOKE2
		drControlNs2Patch, err := json.Marshal(statusForPatching(targetSpoke2, ramenv1alpha1.Deployed, metav1.ConditionTrue))
		Expect(err).NotTo(HaveOccurred())
		Expect(testClient.Status().Patch(ctx, drControlNs2, client.RawPatch(types.MergePatchType, drControlNs2Patch))).To(Succeed())

		// reconcile event for SPOKE1 and expect success
		result, err := sut.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: targetSpoke1}})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeNil())

		// verify the application from the first namespace is now set to fail-over
		drControlNs1Sut := &ramenv1alpha1.DRPlacementControl{}
		Expect(testClient.Get(ctx, types.NamespacedName{Name: drControlNs1.Name, Namespace: drControlNs1.Namespace}, drControlNs1Sut)).To(Succeed())
		Expect(drControlNs1Sut.Spec.Action).To(Equal(ramenv1alpha1.ActionFailover))

		// verify the application from the second namespace is still relocated
		drControlNs2Sut := &ramenv1alpha1.DRPlacementControl{}
		Expect(testClient.Get(ctx, types.NamespacedName{Name: drControlNs2.Name, Namespace: drControlNs2.Namespace}, drControlNs2Sut)).To(Succeed())
		Expect(drControlNs2Sut.Spec.Action).To(Equal(ramenv1alpha1.ActionRelocate))
	})

	Specify("Only patch applications in suitable phase", func(ctx SpecContext) {
		// patch application on the first namespace to be READY for failover
		drControlNs1Patch, err := json.Marshal(statusForPatching(targetSpoke2, ramenv1alpha1.Deploying, metav1.ConditionTrue))
		Expect(err).NotTo(HaveOccurred())
		Expect(testClient.Status().Patch(ctx, drControlNs1, client.RawPatch(types.MergePatchType, drControlNs1Patch))).To(Succeed())

		// patch application on the second namespace to report a phase NOT SUITABLE for a failover
		drControlNs2Patch, err := json.Marshal(statusForPatching(targetSpoke2, ramenv1alpha1.Initiating, metav1.ConditionTrue))
		Expect(err).NotTo(HaveOccurred())
		Expect(testClient.Status().Patch(ctx, drControlNs2, client.RawPatch(types.MergePatchType, drControlNs2Patch))).To(Succeed())

		// reconcile event and expect success
		result, err := sut.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: targetSpoke2}})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeNil())

		// verify the application from the first namespace is now set to fail-over
		drControlNs1Sut := &ramenv1alpha1.DRPlacementControl{}
		Expect(testClient.Get(ctx, types.NamespacedName{Name: drControlNs1.Name, Namespace: drControlNs1.Namespace}, drControlNs1Sut)).To(Succeed())
		Expect(drControlNs1Sut.Spec.Action).To(Equal(ramenv1alpha1.ActionFailover))

		// verify the application from the second namespace is still relocated
		drControlNs2Sut := &ramenv1alpha1.DRPlacementControl{}
		Expect(testClient.Get(ctx, types.NamespacedName{Name: drControlNs2.Name, Namespace: drControlNs2.Namespace}, drControlNs2Sut)).To(Succeed())
		Expect(drControlNs2Sut.Spec.Action).To(Equal(ramenv1alpha1.ActionRelocate))
	})

	Specify("Only patch applications with an available peer", func(ctx SpecContext) {
		// patch application on the first namespace to be READY for failover
		drControlNs1Patch, err := json.Marshal(statusForPatching(targetSpoke1, ramenv1alpha1.Deployed, metav1.ConditionTrue))
		Expect(err).NotTo(HaveOccurred())
		Expect(testClient.Status().Patch(ctx, drControlNs1, client.RawPatch(types.MergePatchType, drControlNs1Patch))).To(Succeed())

		// patch application on the second namespace to report an UNAVAILABLE PEER
		drControlNs2Patch, err := json.Marshal(statusForPatching(targetSpoke1, ramenv1alpha1.Deployed, metav1.ConditionFalse))
		Expect(err).NotTo(HaveOccurred())
		Expect(testClient.Status().Patch(ctx, drControlNs2, client.RawPatch(types.MergePatchType, drControlNs2Patch))).To(Succeed())

		// reconcile event and expect failure
		result, err := sut.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: targetSpoke1}})
		Expect(err).To(MatchError("failed to failover one or more DR placement controls"))
		Expect(result).NotTo(BeNil())

		// verify the application from the first namespace is now set to fail-over
		drControlNs1Sut := &ramenv1alpha1.DRPlacementControl{}
		Expect(testClient.Get(ctx, types.NamespacedName{Name: drControlNs1.Name, Namespace: drControlNs1.Namespace}, drControlNs1Sut)).To(Succeed())
		Expect(drControlNs1Sut.Spec.Action).To(Equal(ramenv1alpha1.ActionFailover))

		// verify the application from the second namespace is still relocated
		drControlNs2Sut := &ramenv1alpha1.DRPlacementControl{}
		Expect(testClient.Get(ctx, types.NamespacedName{Name: drControlNs2.Name, Namespace: drControlNs2.Namespace}, drControlNs2Sut)).To(Succeed())
		Expect(drControlNs2Sut.Spec.Action).To(Equal(ramenv1alpha1.ActionRelocate))
	})
})
