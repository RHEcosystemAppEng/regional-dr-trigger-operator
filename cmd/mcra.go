// Copyright (c) 2023 Red Hat, Inc.

package cmd

import (
	"context"
	"github.com/spf13/cobra"
)

const (
	addonName      = "multicluster-resiliency-addon"
	agentNamespace = "open-cluster-management-agent-addon"
)

var mcraCmd = cobra.Command{
	Use:          "mcra",
	Short:        "Multicluster Resiliency Addon",
	Long:         "multicluster-resiliency-addon TODO add description",
	SilenceUsage: true,
}

func Run(ctx context.Context) error {
	return mcraCmd.ExecuteContext(ctx)
}
