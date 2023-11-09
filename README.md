# MultiCluster Resiliency Addon

The _MultiCluster Resiliency Addon_ is a [Red Hat Advanced Cluster Management][acm] _Addon_. It replaces unavailable
_Managed Clusters_ by leveraging [OpenShift Hive][hive].

Helm chart for deploying the Addon can be found [here][chart].

## Usage

Create a _ManagedClusterAddon_ resource in a _Managed Cluster Namespace_ to make it _Resilient_. 

```yaml
apiVersion: addon.open-cluster-management.io/v1alpha1
kind: ManagedClusterAddOn
metadata:
  name: multicluster-resiliency-addon
  namespace: "<managed-cluster-name-goes-here>"
spec:
  installNamespace: open-cluster-management-agent-addon
```

Create a _ConfigMap_ named _multicluster-resiliency-addon-config_, setting the target [Hive ClusterPool][pool] for
claiming new clusters from.

> The _Addon_ takes its configuration from either the _Managed Cluster Namespace_ or the _open-cluster-management_ one.
> The former will take precedence.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: multicluster-resiliency-addon-config
  namespace: "<open-cluster-management | managed-cluster-name>"
data:
    hive_pool_name: "<pool-name-goes-here>"
```

## Verify

After installing the _Addon_ in a _Managed Cluster Namespace_, a _ResilientCluster_ reflecting the cluster status is
created:

```shell
$ oc get ResilientCluster -n <managed-cluster-name-goes-here>

NAME                   AVAILABLE
managed-cluster-name   True
```

Once the _Cluster_ availability is set to _True_, when no longer available, the  _MultiCluster Resiliency Addon_ will
replace it.

## Documentation

* [Configure](docs/configure.md) - Learn how to configure the Addon
* [Hacking](docs/hacking.md) - A walk through the Addon process
* [Actions](docs/actions.md) - A list of actions preformed by the Addon for replacing a cluster
* [Coding](docs/coding.md) - A guide for developers
* [Metrics](docs/metrics.md) - A list of _Prometheus_ metrics reported by the Addon
* [Contributing](.github/CONTRIBUTING.md) - Contributing guidelines

<!--LINKS-->
[acm]: https://www.redhat.com/en/technologies/management/advanced-cluster-management
[hive]: https://github.com/openshift/hive
[pool]: https://github.com/openshift/hive/blob/master/docs/clusterpools.md
[chart]: https://github.com/RHEcosystemAppEng/multicluster-resiliency-addon-chart
