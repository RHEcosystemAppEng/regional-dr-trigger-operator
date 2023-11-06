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

Note a new resource representing the _Spoke_'s status (initial status update might take a minute):

```shell
$ oc get ResilientCluster -n <managed-cluster-name-goes-here>

NAME                   AVAILABLE
managed-cluster-name   True
```

## MCRA Addon Controller

The _ResilientCluster_ status is determined based on the corresponding [ManagedClusterAddon][acm-clusters], which is
getting updated by the [Registration Component][registration-controller] based on the lease obtained by the
_Addon Agent_.

This is accomplished by a controller named [MCRA Addon Controller](../pkg/controllers/reconcilers/addon.go), watching
_ManagedClusterAddon_ resources for our _Addon_, and creating/updating the corresponding  _ResilientCluster_ resources.

> Note, _ResilientCluster_ resources creation, modification, and deletion, are enforced by a
> _Validation Admission Webhook_, enforcing no outside interference in the API process. 

## MCRA Cluster Controller

Once a _ResilientCluster_ status is updated, a controller named
[MCRA Cluster Controller](../pkg/controllers/reconcilers/cluster.go) is triggered by watching _ResilientCluster_
resources. This controller is in charge of deciding whether the _ResilientCluster_ status requires a new cluster. For
instance if the status was changed from _True_ to _False_.

If a new cluster is required, the controller, based on the pre-configured [Hive Pool][hive-pool]
(see [Configure](configure.md)), will create a [ClusterClaim][hive-claim] marked with a target annotation specifying the
previous spoke name, named `multicluster-resiliency-addon/previous-spoke` (later removed by the claim controller).

## MCRA Claim Controller

The [MCRA Claim Controller](../pkg/controllers/reconcilers/claim.go) watches _Hive_'s _ClusterClaim_ resources annotated
with the aforementioned target annotation, `multicluster-resiliency-addon/previous-spoke` (the annotation is removed
when the controller's work is done). The controller verifies the status of the claim by examining its conditions,
_ClusterRunning_ and _Pending_. If the claim is determined to be completed, meaning, a new cluster is ready, the
controller will proceed to invoke the actions described in [Actions](actions.md) in order to get the new cluster ready
for its workload.

[Go Back](../README.md)

<!--LINKS-->
[acm-clusters]: https://access.redhat.com/documentation/en-us/red_hat_advanced_cluster_management_for_kubernetes/2.8/html-single/clusters/index
[hive-pool]: https://github.com/openshift/hive/blob/master/docs/clusterpools.md
[hive-claim]: https://github.com/openshift/hive/blob/master/docs/clusterpools.md#sample-cluster-claim
[registration-controller]: https://github.com/stolostron/registration
