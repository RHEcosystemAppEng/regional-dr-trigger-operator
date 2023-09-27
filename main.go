// Copyright (c) 2023 Red Hat, Inc.

package main

import (
	"context"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/cmd"
	"k8s.io/klog/v2"
	"os"
)

func main() {
	if err := cmd.Run(context.Background()); err != nil {
		klog.Errorf("failed to run mcra: %v", err)
		os.Exit(1)
	}
}
