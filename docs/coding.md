# MultiCluster Resiliency Addon - Coding

**We don't have tests yet, neither unit nor integration. So BE CAREFUL!**

It's recommended to start with the [hacking document](hacking.md) before continuing further with this one.

## Commands

This operator offers two commands, `mcra agent` ([cmd/agent.go](../cmd/agent.go)) and `mcra manager`
([cmd/manager.go](../cmd/manager.go)). The root command, logging configuration, and main execution function are in
[cmd/mcra.go](../cmd/mcra.go).

### Agent

The `mcra agent` command, implemented in [pkg/agent/agent.go](../pkg/agent/agent.go), is used for the _Addon Agent_
deployment, running on _Spokes_. The agent implementation is pretty straight forward, used only for obtaining and
updating a lease against the _Hub_.

### Manager

The `mcra manager` command, implemented in [pkg/manager/manager.go](../pkg/manager/manager.go), is used for the
_Addon Manager_ deployment, running on the _Hub_. The manager implementation is used for creating the _Addon Manager_,
describing the _Addon Agent_, and running the [controller](#controller).

The functions for describing the _Addon Agent_ deployment can be found in
[pkg/manager/agent.go](../pkg/manager/agent.go).<br/>
The _Deployment_ templates for applying on _Spokes_ can be found in 
[pkg/manager/templates/agent](../pkg/manager/templates/agent).

The functions for registering _Addon Agents_ can be found in
[pkg/manager/registration.go](../pkg/manager/registration.go).<br/>
The _RBAC_ templates for agents permissions on the _Hub_
can be found in [pkg/manager/templates/rbac](../pkg/manager/templates/rbac).

The utility function for loading templates, can be found in [pkg/manager/template.go](../pkg/manager/template.go). 

## Controller

Watches _ManagedClusters_ for availability status. If not available, a failover for all the _DRPlacementControls_
declaring the reporting cluster as their primary one, will be triggered.

The code can be found in [pkg/controller](../pkg/controller).

Predicates for filtering controller events can be found in [pkg/controller/predicates.go](../pkg/controller/predicates.go).

[Go Back](../README.md#documentation)
