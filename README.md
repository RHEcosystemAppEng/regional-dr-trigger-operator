# MultiCluster Resiliency Addon

The _MultiCluster Resiliency Addon_ is a [Red Hat Advanced Cluster Management][acm] _Addon_ built for
[Red Hat OpenShift Data Foundation][odf] and used for [Disaster Recovery][dr] scenarios. It will trigger a [Regional DR][regional]
failover for all applications running on an unavailable cluster.

The Helm chart for deploying the Addon can be found [here][chart].

## Usage

Create a _ManagedClusterAddon_ in a cluster-namespace to make it "_Resilient_". If the corresponded _ManagedCluster_
will report unavailability, this addon will trigger a failover for all the _DRPlacementControls_ declaring the reporting
cluster as their primary one.

```yaml
apiVersion: addon.open-cluster-management.io/v1alpha1
kind: ManagedClusterAddOn
metadata:
  name: multicluster-resiliency-addon
  namespace: "<managed-cluster-name-goes-here>"
spec: {}
```

## Documentation

* [Configure](docs/configure.md) - Learn how to configure the Addon
* [Hacking](docs/hacking.md) - A walk through the Addon process
* [Coding](docs/coding.md) - A guide for developers
* [Metrics](docs/metrics.md) - A list of _Prometheus_ metrics reported by the Addon
* [Contributing](.github/CONTRIBUTING.md) - Contributing guidelines

<!--LINKS-->
[acm]: https://www.redhat.com/en/technologies/management/advanced-cluster-management
[odf]: https://access.redhat.com/documentation/en-us/red_hat_openshift_data_foundation/4.14
[dr]: https://access.redhat.com/documentation/en-us/red_hat_openshift_data_foundation/4.14/html/configuring_openshift_data_foundation_disaster_recovery_for_openshift_workloads/index
[regional]: https://access.redhat.com/documentation/en-us/red_hat_openshift_data_foundation/4.14/html/configuring_openshift_data_foundation_disaster_recovery_for_openshift_workloads/rdr-solution
[chart]: https://github.com/RHEcosystemAppEng/multicluster-resiliency-addon-chart
