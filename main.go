// Copyright (c) 2023 Red Hat, Inc.

package main

import (
	"fmt"
	"regional-dr-trigger-operator/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(fmt.Sprintf("failed running the regional dr trigger operator: %v", err))
	}
}
