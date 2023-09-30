// Copyright (c) 2023 Red Hat, Inc.

package manager

import (
	"context"
	"embed"
	"fmt"
	"k8s.io/client-go/rest"
	"open-cluster-management.io/api/addon/v1alpha1"
	v1 "open-cluster-management.io/api/cluster/v1"

	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/klog/v2"
	"open-cluster-management.io/addon-framework/pkg/addonfactory"
	"open-cluster-management.io/addon-framework/pkg/addonmanager"
	"open-cluster-management.io/addon-framework/pkg/agent"
	"open-cluster-management.io/addon-framework/pkg/utils"
)

//go:embed agenttemplates
var FS embed.FS

const (
	addonName      = "multicluster-resiliency-addon"
	agentNamespace = "open-cluster-management-agent-addon"
)

type Values struct {
	KubeConfigSecret string
	SpokeName        string
	AddonName        string
	AgentNamespace   string
}

func Run(ctx context.Context, kubeConfig *rest.Config) error {
	klog.Info("running manager")

	addonMgr, err := addonmanager.New(kubeConfig)
	if err != nil {
		return err
	}

	agentName := rand.String(5)
	regOpts := &agent.RegistrationOption{
		CSRConfigurations: agent.KubeClientSignerConfigurations(addonName, agentName),
		CSRApproveCheck:   utils.DefaultCSRApprover(agentName),
	}

	agentAddon, err := addonfactory.
		NewAgentAddonFactory(addonName, FS, "agenttemplates").
		WithGetValuesFuncs(generateTemplateValues).
		WithAgentRegistrationOption(regOpts).
		BuildTemplateAgentAddon()
	if err != nil {
		return err
	}

	if err = addonMgr.AddAgent(agentAddon); err != nil {
		return err
	}

	go func() {
		if err := addonMgr.Start(ctx); err != nil {
			klog.Fatalf("failed to add start addon: %v", err)
		}
	}()

	<-ctx.Done()

	return nil
}

func generateTemplateValues(cluster *v1.ManagedCluster, addon *v1alpha1.ManagedClusterAddOn) (addonfactory.Values, error) {
	values := Values{
		KubeConfigSecret: fmt.Sprintf("%s-hub-kubeconfig", addon.Name),
		SpokeName:        cluster.Name,
		AddonName:        addonName,
		AgentNamespace:   agentNamespace,
	}
	return addonfactory.StructToValues(values), nil
}
