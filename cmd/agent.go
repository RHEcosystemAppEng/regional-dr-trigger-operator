// Copyright (c) 2023 Red Hat, Inc.

package cmd

import (
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/agent"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/version"
	"open-cluster-management.io/addon-framework/pkg/cmd/factory"
)

func init() {
	agtOpts := &agent.Options{}
	agt := agent.Agent{Options: agtOpts}

	agtCmd := factory.
		NewControllerCommandConfig("multicluster-resiliency-addon-agent", version.Get(), agt.Run).
		NewCommand()
	agtCmd.Use = "agent"
	agtCmd.Short = "Multicluster Resiliency Addon Agent"

	agtCmd.Flags().StringVar(&agtOpts.HubKubeConfigFile, "hub-kubeconfig", "blabla", "TODO")
	agtCmd.Flags().StringVar(&agtOpts.SpokeName, "spoke-name", "blabla", "TODO")
	agtCmd.Flags().StringVar(&agtOpts.AgentNamespace, "agent-namespace", "blabla", "TODO")

	mcraCmd.AddCommand(agtCmd)
}
