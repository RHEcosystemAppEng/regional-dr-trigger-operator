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

	agentCmd := factory.
		NewControllerCommandConfig("multicluster-resiliency-addon-agent", version.Get(), agt.Run).
		NewCommand()
	agentCmd.Use = "agent"
	agentCmd.Short = "Multicluster Resiliency Addon Agent"

	agentCmd.Flags().StringVar(&agtOpts.HubKubeConfigFile, "hub-kubeconfig", "blabla", "TODO")
	agentCmd.Flags().StringVar(&agtOpts.SpokeName, "spoke-name", "blabla", "TODO")
	agentCmd.Flags().StringVar(&agtOpts.AddonName, "addon-name", "blabla", "TODO")
	agentCmd.Flags().StringVar(&agtOpts.AgentNamespace, "agent-namespace", "blabla", "TODO")

	mcraCmd.AddCommand(agentCmd)
}
