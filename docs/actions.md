# MultiCluster Resiliency Addon - Actions

When a _Hive_'s [ClusterClaim][hive-claim] created by the _MultiCluster Resiliency Addon_ is completed, the following
actions are preformed in order to switch the workload to the new cluster.

> Note, the actions described here are preformed **only** for created by the _Addon_, identified with a target
> annotation (see [Hacking](hacking.md#mcra-claim-controller)).

> Note, the creation of a _ManagedCluster_ resource representing the new cluster in _ACM_, is handled by
> [ACM's ClusterClaim Controller][cluster-claim-controller] and not by the _Addon_.

| Action                                               | Description                                                                                                                                                         |
|------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [Compare _ManagedCluster_ Resources][compare-mc]     | Copies labels and annotations from the OLD _MC_ to the NEW one, overriding only the _clusterset_ label. Deletes the OLD _MC_ when done.                             |
| [Delete Old _ClusterDeployment_][delete-cd]          | Deletes the _ClusterDeployment_ from the OLD Spoke.                                                                                                                 |
| [Migrate known _AddonDeploymentConfig_][migrate-adc] | Moves any _AddonDeploymentConfig_ resources associated with the _Addon_'s _ManagedClusterAddon_ from the OLD Spoke to the NEW one.                                  |
| [Migrate known _ConfigMap_][migrate-cm]              | Moves the _Addon_'s _ConfigMap_ if found, from the OLD Spoke to the NEW one.                                                                                        |
| [Migrate addon's _ManagedClusterAddon_][migrate-mca] | Moves the _Addon_'s _ManagedClusterAddon_ if found, from the OLD Spoke to the NEW one. If the NEW one was already created by the Addon, the content will be merged. |

[Go Back](../README.md#documentation)

<!--LINKS-->
[hive-claim]: https://github.com/openshift/hive/blob/master/docs/clusterpools.md#sample-cluster-claim
[cluster-claim-controller]: https://github.com/stolostron/clusterclaims-controller

<!--ACTIONS-->
[compare-mc]: ../pkg/controllers/actions/compare_managed_cluster_and_delete_old.go
[delete-cd]: ../pkg/controllers/actions/delete_old_cluster_deployment.go
[migrate-adc]: ../pkg/controllers/actions/migrate_addon_deployment_configs.go
[migrate-cm]: ../pkg/controllers/actions/migrate_config_map.go
[migrate-mca]: ../pkg/controllers/actions/migrate_managed_cluster_addon.go
