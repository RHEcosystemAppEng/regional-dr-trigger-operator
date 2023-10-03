// Copyright (c) 2023 Red Hat, Inc.

package cmd

import (
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/agent"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/version"
	"open-cluster-management.io/addon-framework/pkg/cmd/factory"
)

// init is used for creating the Agent Commend, incorporate its flags, and binding it to the root MCRA Command.
func init() {
	agt := agent.NewAgent()

	agtCmd := factory.
		NewControllerCommandConfig("multicluster-resiliency-addon-agent", version.Get(), agt.Run).
		NewCommand()
	agtCmd.Use = "agent"
	agtCmd.Short = "Multicluster Resiliency Addon Agent"

	agtCmd.Flags().StringVar(&agt.Options.HubKubeConfigFile, "hub-kubeconfig", "blabla", "TODO")
	agtCmd.Flags().StringVar(&agt.Options.SpokeName, "spoke-name", "blabla", "TODO")
	agtCmd.Flags().StringVar(&agt.Options.AgentNamespace, "agent-namespace", "blabla", "TODO")

	mcraCmd.AddCommand(agtCmd)
}
