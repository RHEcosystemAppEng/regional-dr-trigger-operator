// Copyright (c) 2023 Red Hat, Inc.

package controller

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ramenv1alpha1 "github.com/ramendr/ramen/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"testing"
)

var testClient client.Client
var testEnv *envtest.Environment
var drtController *DRTriggerController

// TestController is used for bootstrapping Ginkgo and Gomega
func TestController(t *testing.T) {
	RegisterFailHandler(Fail)            // Set Gomega to report failure to Ginkgo
	RunSpecs(t, "Controller Unit Tests") // run Ginkgo with testing
}

var _ = BeforeSuite(func(ctx SpecContext) {
	By("bootstrapping testing environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("testdata", "external_crds")},
	}

	// install the scheme
	scheme := runtime.NewScheme()
	Expect(clusterv1.Install(scheme)).To(Succeed())
	Expect(ramenv1alpha1.AddToScheme(scheme)).To(Succeed())
	Expect(corev1.AddToScheme(scheme)).To(Succeed()) // i.e. Namespace

	// start testing environment and get config for the client
	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	// save the test client
	testClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(testClient).NotTo(BeNil())
	Expect(err).NotTo(HaveOccurred())

	// create the controller
	drtController = &DRTriggerController{Client: testClient, Scheme: scheme}
})

var _ = AfterSuite(func() {
	By("tearing down testing environment")
	Expect(testEnv.Stop()).To(Succeed())
})
