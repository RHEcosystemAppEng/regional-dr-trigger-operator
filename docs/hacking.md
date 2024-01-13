# MultiCluster Resiliency Addon - Hacking

The following guide is intended for developers/users looking to get familiar with Addon process.

Install the _Addon Manager_ using _make_, from the root:

```shell
make addon/deploy
```

Install the _Addon Agent_ at the _Managed Cluster_ namespace, by applying the _ManagedClusterAddon_ resource as
described in the [Usage section](../README.md#usage), or:

```shell
SPOKE_NAME=<managed-cluster-name-goes-here> make addon/install
```

> Alternatively, use `addon/uninstall` and `addon/undeploy` for rollback.

[Go Back](../README.md#documentation)
