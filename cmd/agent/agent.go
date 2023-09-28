package agent

import (
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/agent"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/version"
	"github.com/spf13/cobra"
	"open-cluster-management.io/addon-framework/pkg/cmd/factory"
)

func GetCommand() *cobra.Command {
	agtOpts := &agent.Options{}

	agentCmd := factory.
		NewControllerCommandConfig("multicluster-resiliency-addon-agent", version.Get(), agent.Run(agtOpts)).
		NewCommand()
	agentCmd.Use = "agent"
	agentCmd.Short = "Multicluster Resiliency Addon Agent"

	agentCmd.Flags().StringVar(&agtOpts.HubKubeConfigFile, "hub-kubeconfig", "blabla", "TODO")
	agentCmd.Flags().StringVar(&agtOpts.SpokeName, "spoke-name", "blabla", "TODO")
	agentCmd.Flags().StringVar(&agtOpts.AddonName, "addon-name", "blabla", "TODO")
	agentCmd.Flags().StringVar(&agtOpts.AgentNamespace, "agent-namespace", "blabla", "TODO")

	return agentCmd
}
