# MultiCluster Resiliency Addon - Metrics

The following _Prometheus_ metrics are reported by the _MultiCluster Resiliency Addon_. The metrics code can be found in
[pkg/metrics](../pkg/metrics).

| Name                                | Description                                                        | Type    | Labels                                |
|-------------------------------------|--------------------------------------------------------------------|---------|---------------------------------------|
| resilient_spoke_not_available_count | Count times the Resilient Spoke cluster was reported not available | Counter | spoke_name                            |
| resilient_spoke_available_count     | Count times the Resilient Spoke cluster was reported available     | Counter | spoke_name                            |
| new_cluster_claim_created           | Count the times we created a new ClusterClaim for Hive             | Counter | pool_name, claim_name, old_spoke_name |
| new_spoke_ready                     | Count the time we got a new ready cluster                          | Counter | old_spoke_name, new_spoke_name        |

[Go Back](../README.md#documentation)
