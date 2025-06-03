// Copyright (c) 2023 Red Hat, Inc.

package controller

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ramenv1alpha1 "github.com/ramendr/ramen/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Context("DR Trigger Controller", func() {
	It("should not reconcile if the managed cluster have not joined the hub", func(ctx SpecContext) {
		testName := "mc-not-joined"

		By("Create a ManagedCluster")
		mc := &clusterv1.ManagedCluster{ObjectMeta: metav1.ObjectMeta{Name: testName}}
		Expect(testClient.Create(ctx, mc)).To(Succeed())

		By("Reconcile for the MC")
		res, err := drtController.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(mc)})
		Expect(res.Requeue).To(BeFalse())
		Expect(err).NotTo(HaveOccurred())

		By("Cleanups")
		Expect(testClient.Delete(ctx, mc)).To(Succeed())
	})

	It("should not reconcile if the managed cluster is not accepted by the hub", func(ctx SpecContext) {
		testName := "mc-not-accepted"

		By("Create a ManagedCluster")
		mc := &clusterv1.ManagedCluster{ObjectMeta: metav1.ObjectMeta{Name: testName}}
		Expect(testClient.Create(ctx, mc)).To(Succeed())

		By("Update the MC status")
		mc.Status = clusterv1.ManagedClusterStatus{Conditions: []metav1.Condition{{
			Type:               clusterv1.ManagedClusterConditionJoined,
			Status:             metav1.ConditionTrue,
			Reason:             "MC_Joined",
			LastTransitionTime: metav1.Now(),
		}}}
		Expect(testClient.Status().Update(ctx, mc)).To(Succeed())

		By("Reconcile for the MC")
		res, err := drtController.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(mc)})
		Expect(res.Requeue).To(BeFalse())
		Expect(err).NotTo(HaveOccurred())

		By("Cleanups")
		Expect(testClient.Delete(ctx, mc)).To(Succeed())
	})

	It("should not reconcile if the managed cluster is not available", func(ctx SpecContext) {
		testName := "mc-not-available"

		By("Create a ManagedCluster")
		mc := &clusterv1.ManagedCluster{
			ObjectMeta: metav1.ObjectMeta{Name: testName},
			Spec:       clusterv1.ManagedClusterSpec{HubAcceptsClient: true},
		}
		Expect(testClient.Create(ctx, mc)).To(Succeed())

		By("Update the MC status")
		mc.Status = clusterv1.ManagedClusterStatus{Conditions: []metav1.Condition{{
			Type:               clusterv1.ManagedClusterConditionJoined,
			Status:             metav1.ConditionTrue,
			Reason:             "MC_Joined",
			LastTransitionTime: metav1.Now(),
		}}}
		Expect(testClient.Status().Update(ctx, mc)).To(Succeed())

		By("Reconcile for the MC")
		res, err := drtController.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(mc)})
		Expect(res.Requeue).To(BeFalse())
		Expect(err).NotTo(HaveOccurred())

		By("Cleanups")
		Expect(testClient.Delete(ctx, mc)).To(Succeed())
	})

	It("should only failover dr controls preferring the current cluster", func(ctx SpecContext) {
		testName := "only-failover-selected"

		By("Create a ManagedCluster")
		mc := &clusterv1.ManagedCluster{
			ObjectMeta: metav1.ObjectMeta{Name: testName},
			Spec:       clusterv1.ManagedClusterSpec{HubAcceptsClient: true},
		}
		Expect(testClient.Create(ctx, mc)).To(Succeed())

		By("Update the MC status")
		mc.Status = clusterv1.ManagedClusterStatus{Conditions: []metav1.Condition{
			{
				Type:               clusterv1.ManagedClusterConditionJoined,
				Status:             metav1.ConditionTrue,
				Reason:             "MC_Joined",
				LastTransitionTime: metav1.Now(),
			},
			{
				Type:               clusterv1.ManagedClusterConditionAvailable,
				Status:             metav1.ConditionTrue,
				Reason:             "MC_Available",
				LastTransitionTime: metav1.Now(),
			},
		}}
		Expect(testClient.Status().Update(ctx, mc)).To(Succeed())

		By("Create a Namespace for the right DRPolicyControl")
		rightNs := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testName + "right-ns"}}
		Expect(testClient.Create(ctx, rightNs)).To(Succeed())

		By("Create the right DRPolicyControl")
		rightDr := &ramenv1alpha1.DRPlacementControl{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testName + "right-dr",
				Namespace: rightNs.Name,
			},
			Spec: ramenv1alpha1.DRPlacementControlSpec{
				Action: ramenv1alpha1.ActionRelocate,
			},
		}
		Expect(testClient.Create(ctx, rightDr))

		By("Update the right DRPC status")
		rightDr.Status = ramenv1alpha1.DRPlacementControlStatus{
			PreferredDecision: ramenv1alpha1.PlacementDecision{
				ClusterName: mc.Name,
			},
			Phase: ramenv1alpha1.Deployed,
			Conditions: []metav1.Condition{
				{
					Type:               ramenv1alpha1.ConditionPeerReady,
					Status:             metav1.ConditionTrue,
					Reason:             "DR_Peer_Ready",
					LastTransitionTime: metav1.Now(),
				},
			},
		}
		Expect(testClient.Status().Update(ctx, rightDr)).To(Succeed())

		By("Create a Namespace for the wrong DRPolicyControl")
		wrongNs := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testName + "wrong-ns"}}
		Expect(testClient.Create(ctx, wrongNs)).To(Succeed())

		By("Create the right DRPolicyControl")
		wrongDr := &ramenv1alpha1.DRPlacementControl{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testName + "wrong-dr",
				Namespace: wrongNs.Name,
			},
			Spec: ramenv1alpha1.DRPlacementControlSpec{
				Action: ramenv1alpha1.ActionRelocate,
			},
		}
		Expect(testClient.Create(ctx, wrongDr))

		By("Update the wrong DRPC status")
		wrongDr.Status = ramenv1alpha1.DRPlacementControlStatus{
			PreferredDecision: ramenv1alpha1.PlacementDecision{
				ClusterName: "not_the_correct_managed_cluster", // NOTE DRPC not preferring the cluster we're testing
			},
			Phase: ramenv1alpha1.Deployed,
			Conditions: []metav1.Condition{
				{
					Type:               ramenv1alpha1.ConditionPeerReady,
					Status:             metav1.ConditionTrue,
					Reason:             "DR_Peer_Ready",
					LastTransitionTime: metav1.Now(),
				},
			},
		}
		Expect(testClient.Status().Update(ctx, wrongDr)).To(Succeed())

		By("Reconcile for the MC")
		res, err := drtController.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(mc)})
		Expect(res.Requeue).To(BeFalse())
		Expect(err).NotTo(HaveOccurred())

		By("Verify the right DRPC was failed-over")
		Eventually(func() error {
			rightDrUpdate := &ramenv1alpha1.DRPlacementControl{}
			if err := testClient.Get(ctx, client.ObjectKeyFromObject(rightDr), rightDrUpdate); err != nil {
				return err
			}
			if rightDrUpdate.Spec.Action != ramenv1alpha1.ActionFailover {
				return fmt.Errorf("not failed over")
			}
			return nil
		}).Should(Succeed())

		By("Verify the wrong DRPC was not failed-over")
		Eventually(func() error {
			wrongDrUpdate := &ramenv1alpha1.DRPlacementControl{}
			if err := testClient.Get(ctx, client.ObjectKeyFromObject(wrongDr), wrongDrUpdate); err != nil {
				return err
			}
			if wrongDrUpdate.Spec.Action != ramenv1alpha1.ActionRelocate {
				return fmt.Errorf("not failed over")
			}
			return nil
		}).Should(Succeed())

		By("Cleanups")
		Expect(testClient.Delete(ctx, wrongDr)).To(Succeed())
		Expect(testClient.Delete(ctx, wrongNs)).To(Succeed())
		Expect(testClient.Delete(ctx, rightDr)).To(Succeed())
		Expect(testClient.Delete(ctx, rightNs)).To(Succeed())
		Expect(testClient.Delete(ctx, mc)).To(Succeed())
	})

	It("should only failover dr controls in a suitable phase", func(ctx SpecContext) {
		testName := "only-failover-matching-phase"

		By("Create a ManagedCluster")
		mc := &clusterv1.ManagedCluster{
			ObjectMeta: metav1.ObjectMeta{Name: testName},
			Spec:       clusterv1.ManagedClusterSpec{HubAcceptsClient: true},
		}
		Expect(testClient.Create(ctx, mc)).To(Succeed())

		By("Update the MC status")
		mc.Status = clusterv1.ManagedClusterStatus{Conditions: []metav1.Condition{
			{
				Type:               clusterv1.ManagedClusterConditionJoined,
				Status:             metav1.ConditionTrue,
				Reason:             "MC_Joined",
				LastTransitionTime: metav1.Now(),
			},
			{
				Type:               clusterv1.ManagedClusterConditionAvailable,
				Status:             metav1.ConditionTrue,
				Reason:             "MC_Available",
				LastTransitionTime: metav1.Now(),
			},
		}}
		Expect(testClient.Status().Update(ctx, mc)).To(Succeed())

		By("Create a Namespace for the right DRPolicyControl")
		rightNs := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testName + "right-ns"}}
		Expect(testClient.Create(ctx, rightNs)).To(Succeed())

		By("Create the right DRPolicyControl")
		rightDr := &ramenv1alpha1.DRPlacementControl{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testName + "right-dr",
				Namespace: rightNs.Name,
			},
			Spec: ramenv1alpha1.DRPlacementControlSpec{
				Action: ramenv1alpha1.ActionRelocate,
			},
		}
		Expect(testClient.Create(ctx, rightDr))

		By("Update the right DRPC status")
		rightDr.Status = ramenv1alpha1.DRPlacementControlStatus{
			PreferredDecision: ramenv1alpha1.PlacementDecision{
				ClusterName: mc.Name,
			},
			Phase: ramenv1alpha1.Deploying,
			Conditions: []metav1.Condition{
				{
					Type:               ramenv1alpha1.ConditionPeerReady,
					Status:             metav1.ConditionTrue,
					Reason:             "DR_Peer_Ready",
					LastTransitionTime: metav1.Now(),
				},
			},
		}
		Expect(testClient.Status().Update(ctx, rightDr)).To(Succeed())

		By("Create a Namespace for the wrong DRPolicyControl")
		wrongNs := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testName + "wrong-ns"}}
		Expect(testClient.Create(ctx, wrongNs)).To(Succeed())

		By("Create the right DRPolicyControl")
		wrongDr := &ramenv1alpha1.DRPlacementControl{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testName + "wrong-dr",
				Namespace: wrongNs.Name,
			},
			Spec: ramenv1alpha1.DRPlacementControlSpec{
				Action: ramenv1alpha1.ActionRelocate,
			},
		}
		Expect(testClient.Create(ctx, wrongDr))

		By("Update the wrong DRPC status")
		wrongDr.Status = ramenv1alpha1.DRPlacementControlStatus{
			PreferredDecision: ramenv1alpha1.PlacementDecision{
				ClusterName: mc.Name, // NOTE both DRPCs preferring the cluster we're testing
			},
			Phase: ramenv1alpha1.Initiating, // NOTE DRPC not in a phase suitable for failing over
			Conditions: []metav1.Condition{
				{
					Type:               ramenv1alpha1.ConditionPeerReady,
					Status:             metav1.ConditionTrue,
					Reason:             "DR_Peer_Ready",
					LastTransitionTime: metav1.Now(),
				},
			},
		}
		Expect(testClient.Status().Update(ctx, wrongDr)).To(Succeed())

		By("Reconcile for the MC")
		res, err := drtController.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(mc)})
		Expect(res.Requeue).To(BeFalse())
		Expect(err).NotTo(HaveOccurred())

		By("Verify the right DRPC was failed-over")
		Eventually(func() error {
			rightDrUpdate := &ramenv1alpha1.DRPlacementControl{}
			if err := testClient.Get(ctx, client.ObjectKeyFromObject(rightDr), rightDrUpdate); err != nil {
				return err
			}
			if rightDrUpdate.Spec.Action != ramenv1alpha1.ActionFailover {
				return fmt.Errorf("not failed over")
			}
			return nil
		}).Should(Succeed())

		By("Verify the wrong DRPC was not failed-over")
		Eventually(func() error {
			wrongDrUpdate := &ramenv1alpha1.DRPlacementControl{}
			if err := testClient.Get(ctx, client.ObjectKeyFromObject(wrongDr), wrongDrUpdate); err != nil {
				return err
			}
			if wrongDrUpdate.Spec.Action != ramenv1alpha1.ActionRelocate {
				return fmt.Errorf("not failed over")
			}
			return nil
		}).Should(Succeed())

		By("Cleanups")
		Expect(testClient.Delete(ctx, wrongDr)).To(Succeed())
		Expect(testClient.Delete(ctx, wrongNs)).To(Succeed())
		Expect(testClient.Delete(ctx, rightDr)).To(Succeed())
		Expect(testClient.Delete(ctx, rightNs)).To(Succeed())
		Expect(testClient.Delete(ctx, mc)).To(Succeed())
	})

	It("should only failover dr controls with peer in a ready state", func(ctx SpecContext) {
		testName := "only-failover-ready-peers"

		By("Create a ManagedCluster")
		mc := &clusterv1.ManagedCluster{
			ObjectMeta: metav1.ObjectMeta{Name: testName},
			Spec:       clusterv1.ManagedClusterSpec{HubAcceptsClient: true},
		}
		Expect(testClient.Create(ctx, mc)).To(Succeed())

		By("Update the MC status")
		mc.Status = clusterv1.ManagedClusterStatus{Conditions: []metav1.Condition{
			{
				Type:               clusterv1.ManagedClusterConditionJoined,
				Status:             metav1.ConditionTrue,
				Reason:             "MC_Joined",
				LastTransitionTime: metav1.Now(),
			},
			{
				Type:               clusterv1.ManagedClusterConditionAvailable,
				Status:             metav1.ConditionTrue,
				Reason:             "MC_Available",
				LastTransitionTime: metav1.Now(),
			},
		}}
		Expect(testClient.Status().Update(ctx, mc)).To(Succeed())

		By("Create a Namespace for the right DRPolicyControl")
		rightNs := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testName + "right-ns"}}
		Expect(testClient.Create(ctx, rightNs)).To(Succeed())

		By("Create the right DRPolicyControl")
		rightDr := &ramenv1alpha1.DRPlacementControl{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testName + "right-dr",
				Namespace: rightNs.Name,
			},
			Spec: ramenv1alpha1.DRPlacementControlSpec{
				Action: ramenv1alpha1.ActionRelocate,
			},
		}
		Expect(testClient.Create(ctx, rightDr))

		By("Update the right DRPC status")
		rightDr.Status = ramenv1alpha1.DRPlacementControlStatus{
			PreferredDecision: ramenv1alpha1.PlacementDecision{
				ClusterName: mc.Name,
			},
			Phase: ramenv1alpha1.Deployed,
			Conditions: []metav1.Condition{
				{
					Type:               ramenv1alpha1.ConditionPeerReady,
					Status:             metav1.ConditionTrue,
					Reason:             "DR_Peer_Ready",
					LastTransitionTime: metav1.Now(),
				},
			},
		}
		Expect(testClient.Status().Update(ctx, rightDr)).To(Succeed())

		By("Create a Namespace for the wrong DRPolicyControl")
		wrongNs := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testName + "wrong-ns"}}
		Expect(testClient.Create(ctx, wrongNs)).To(Succeed())

		By("Create the right DRPolicyControl")
		wrongDr := &ramenv1alpha1.DRPlacementControl{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testName + "wrong-dr",
				Namespace: wrongNs.Name,
			},
			Spec: ramenv1alpha1.DRPlacementControlSpec{
				Action: ramenv1alpha1.ActionRelocate,
			},
		}
		Expect(testClient.Create(ctx, wrongDr))

		By("Update the wrong DRPC status")
		wrongDr.Status = ramenv1alpha1.DRPlacementControlStatus{
			PreferredDecision: ramenv1alpha1.PlacementDecision{
				ClusterName: mc.Name, // NOTE both DRPCs preferring the cluster we're testing
			},
			Phase: ramenv1alpha1.Deployed, // NOTE DRPC in a phase suitable for failing over
			Conditions: []metav1.Condition{
				{
					Type:               ramenv1alpha1.ConditionPeerReady,
					Status:             metav1.ConditionFalse, // NOTE DRPC peer not in a ready state
					Reason:             "DR_Peer_Not_Ready",
					LastTransitionTime: metav1.Now(),
				},
			},
		}
		Expect(testClient.Status().Update(ctx, wrongDr)).To(Succeed())

		By("Reconcile for the MC")
		res, err := drtController.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(mc)})
		Expect(res.Requeue).To(BeFalse())
		Expect(err).NotTo(HaveOccurred())

		By("Verify the right DRPC was failed-over")
		Eventually(func() error {
			rightDrUpdate := &ramenv1alpha1.DRPlacementControl{}
			if err := testClient.Get(ctx, client.ObjectKeyFromObject(rightDr), rightDrUpdate); err != nil {
				return err
			}
			if rightDrUpdate.Spec.Action != ramenv1alpha1.ActionFailover {
				return fmt.Errorf("not failed over")
			}
			return nil
		}).Should(Succeed())

		By("Verify the wrong DRPC was not failed-over")
		Eventually(func() error {
			wrongDrUpdate := &ramenv1alpha1.DRPlacementControl{}
			if err := testClient.Get(ctx, client.ObjectKeyFromObject(wrongDr), wrongDrUpdate); err != nil {
				return err
			}
			if wrongDrUpdate.Spec.Action != ramenv1alpha1.ActionRelocate {
				return fmt.Errorf("not failed over")
			}
			return nil
		}).Should(Succeed())

		By("Cleanups")
		Expect(testClient.Delete(ctx, wrongDr)).To(Succeed())
		Expect(testClient.Delete(ctx, wrongNs)).To(Succeed())
		Expect(testClient.Delete(ctx, rightDr)).To(Succeed())
		Expect(testClient.Delete(ctx, rightNs)).To(Succeed())
		Expect(testClient.Delete(ctx, mc)).To(Succeed())
	})
})
