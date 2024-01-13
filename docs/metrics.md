# MultiCluster Resiliency Addon - Metrics

The following _Prometheus_ metrics are reported by the _MultiCluster Resiliency Addon_. The metrics code can be found in
[pkg/metrics](../pkg/metrics).

| Name                           | Description                                         | Labels                                                |
|--------------------------------|-----------------------------------------------------|-------------------------------------------------------|
| dr_cluster_not_available_count | Counter for DR clusters identified as not available | dr_cluster_name                                       |
| dr_application_failover_count  | Counter for DR application failover performed       | dr_cluster_name, dr_control_name, dr_application_name |

[Go Back](../README.md#documentation)
