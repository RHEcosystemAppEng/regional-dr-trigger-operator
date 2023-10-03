// Copyright (c) 2023 Red Hat, Inc.

package main

import (
	"fmt"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(fmt.Sprintf("failed to run mcra: %v", err))
	}
}
