# Agent Deployment Configuration

The agent deployment can be configured using a global _AddonDeploymentConfig_ named
_multicluster-resiliency-addon-deploy-config_ in the _open-cluster-management_ namespace:

> Note the agent installation namespace on the _Spoke_ can be customized using either the _ManagedClusterAddOn_'s _spec.installNamespace_
> or the _AddOnDeploymentConfig_'s _spec.agentInstallNamespace_. The latter takes precedence regardless if it global or per-cluster.

```yaml
apiVersion: addon.open-cluster-management.io/v1alpha1
kind: AddOnDeploymentConfig
metadata:
  name: multicluster-resiliency-addon-deploy-config
  namespace: open-cluster-management
spec:
# agentInstallNamespace is not yet supported by acm 2.8
#  agentInstallNamespace: open-cluster-management-agent-addon
  customizedVariables:
  - name: AgentReplicas
    value: "1"
```

Configuration per-cluster takes precedence. The _ManagedClusterAddon_ resource takes a reference for said configuration:

```yaml
apiVersion: addon.open-cluster-management.io/v1alpha1
kind: ManagedClusterAddOn
metadata:
  name: multicluster-resiliency-addon
  namespace: "<managed-cluster-name-goes-here>"
spec:
  installNamespace: open-cluster-management-agent-addon
  configs:
  - group: addon.open-cluster-management.io
    resource: addondeploymentconfigs
    name: multicluster-resiliency-addon-deploy-config
    namespace: "<managed-cluster-name-goes-here>"
```

[Go Back](../README.md#documentation)
