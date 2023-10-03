// Copyright (c) 2023 Red Hat, Inc.

package main

import (
	"os"

	"github.com/rhecosystemappeng/multicluster-resiliency-addon/cmd"
	"k8s.io/klog/v2"
)

func main() {
	if err := cmd.Execute(); err != nil {
		klog.Fatalf("failed to run mcra: %v", err)
		os.Exit(1)
	}
}
