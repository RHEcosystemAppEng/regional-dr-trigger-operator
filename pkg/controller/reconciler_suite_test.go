// Copyright (c) 2023 Red Hat, Inc.

package controller

// This file sets up the testing suite for testing the reconciliation loop.

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ramenv1alpha1 "github.com/ramendr/ramen/api/v1alpha1"
	plv1 "github.com/stolostron/multicloud-operators-placementrule/pkg/apis/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"testing"
)

var testClient client.Client
var testEnv *envtest.Environment

var sut *DRTriggerReconciler

// TestReconciler is used for plugging in Ginkgo and Gomega
func TestReconciler(t *testing.T) {
	RegisterFailHandler(Fail)       // Set Gomega to report failure to Ginkgo
	RunSpecs(t, "Reconciler Suite") // run Ginkgo with testing
}

var _ = BeforeSuite(func(ctx SpecContext) {
	// create testing environment (including required external crds)
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("testdata", "external_crds")},
	}

	// install the scheme
	scheme := runtime.NewScheme()
	Expect(installTypes(scheme)).To(Succeed())
	Expect(corev1.AddToScheme(scheme)) // i.e. Namespace

	// start testing environment and get config for the client
	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	// create and save the test client
	testClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(testClient).NotTo(BeNil())

	// set the reconciler as the subject under test
	sut = &DRTriggerReconciler{Scheme: testClient.Scheme(), Client: testClient}

	// create testing namespaces
	Expect(testClient.Create(ctx, testNamespace1)).To(Succeed())
	Expect(testClient.Create(ctx, testNamespace2)).To(Succeed())
})

var _ = AfterSuite(func() {
	// tear down testing environment
	Expect(testEnv.Stop()).To(Succeed())
})

// ###############################################
// #### TESTING OBJECTS AND UTILITY FUNCTIONS ####
// ###############################################
// testNamespace1 is the first Namespace used for testing
var testNamespace1 = &corev1.Namespace{
	ObjectMeta: metav1.ObjectMeta{
		Name: "testing-patched-1",
	},
}

// testNamespace2 is the second Namespace used for testing
var testNamespace2 = &corev1.Namespace{
	ObjectMeta: metav1.ObjectMeta{
		Name: "testing-patched-2",
	},
}

// statusForPatching is a utility functions used for creating DRPlacementControl for patching via the Status endpoint
func statusForPatching(preferredCluster string, controlPhase ramenv1alpha1.DRState, peerReadyStatus metav1.ConditionStatus) *ramenv1alpha1.DRPlacementControl {
	return &ramenv1alpha1.DRPlacementControl{
		Status: ramenv1alpha1.DRPlacementControlStatus{
			PreferredDecision: plv1.PlacementDecision{
				ClusterName: preferredCluster,
			},
			Phase: controlPhase,
			Conditions: []metav1.Condition{
				{
					Type:               ramenv1alpha1.ConditionPeerReady,
					Status:             peerReadyStatus,
					Reason:             "rr",
					LastTransitionTime: metav1.Now(),
				},
			},
		},
	}
}
