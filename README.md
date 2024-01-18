# Regional DR Trigger Operator

Connecting [Red Hat Advanced Cluster Management][acm] with [Red Hat OpenShift Data Foundation][odf]'s
[Disaster Recovery][dr] scenarios. The _Regional DR Trigger Operator_ will trigger a [Regional DR][regional] failover
for all applications running on an unavailable _Managed Cluster_.

## Metrics

| Name                           | Description                                                                        | Labels                                                |
|--------------------------------|------------------------------------------------------------------------------------|-------------------------------------------------------|
| dr_application_failover_count  | Counter for DR Applications failover initiated by the Regional DR Trigger Operator | dr_cluster_name, dr_control_name, dr_application_name |

## Contributing Guidelines

See [Contributing Guidelines](.github/CONTRIBUTING.md) for further information.

<!--LINKS-->
[acm]: https://www.redhat.com/en/technologies/management/advanced-cluster-management
[odf]: https://access.redhat.com/documentation/en-us/red_hat_openshift_data_foundation/4.14
[dr]: https://access.redhat.com/documentation/en-us/red_hat_openshift_data_foundation/4.14/html/configuring_openshift_data_foundation_disaster_recovery_for_openshift_workloads/index
[regional]: https://access.redhat.com/documentation/en-us/red_hat_openshift_data_foundation/4.14/html/configuring_openshift_data_foundation_disaster_recovery_for_openshift_workloads/rdr-solution
