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
describing the _Addon Agent_, and running the [controllers](#controllers).

The functions for describing the _Addon Agent_ deployment can be found in
[pkg/manager/agent.go](../pkg/manager/agent.go).<br/>
The _Deployment_ templates for applying on _Spokes_ can be found in 
[pkg/manager/templates/agent](../pkg/manager/templates/agent).

The functions for registering _Addon Agents_ can be found in
[pkg/manager/registration.go](../pkg/manager/registration.go).<br/>
The _RBAC_ templates for agents permissions on the _Hub_
can be found in [pkg/manager/templates/rbac](../pkg/manager/templates/rbac).

The utility function for loading templates, can be found in [pkg/manager/template.go](../pkg/manager/template.go). 

## Controllers

The execution function for initiating the controller manager, run the controllers, and start up the [webhook](#webhook),
is in [pkg/controllers/controllers.go](../pkg/controllers/controllers.go).<br/>
The controllers are registered and executed in [pkg/controllers/reconcilers/reconcilers.go](../pkg/controllers/reconcilers/reconcilers.go).<br/>
The various controller implementations are stored in [pkg/controllers/reconcilers](../pkg/controllers/reconcilers) and
follow this template:

```go
package reconcilers

type AnyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Options
}

func (r *AnyReconciler) setupWithManager(mgr ctrl.Manager) error {
    // code for setting up the controller with a controller manager 
}

func (r *AnyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // code for executing the controller's reconciliation loop
}

func init() {
	reconcilerFuncs = append(reconcilerFuncs, func(mgr manager.Manager, options Options) error {
		return (&AnyReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme(), Options: options}).setupWithManager(mgr)
	})
}
```

Predicates for filtering controller events can be found in
[pkg/controllers/reconcilers/predicates.go](../pkg/controllers/reconcilers/predicates.go).

#### Addon Controller

Triggered for _ACM_'s _ManagedClusterAddon_ resources, filtered by name. In charge of creating and updating the _Spoke_
status for its _ResilientCluster_ resource.<br/>
The code can be found in [pkg/controllers/reconcilers/addon.go](../pkg/controllers/reconcilers/addon.go).

#### Cluster Controller

Triggered for the _Addon_'s _ResilientCluster_ resources. In charge of determining whether a new _Spoke_ is required,
and creating the required _ClusterClaim_ for _Hive_ if so.<br/>
The code can be found in [pkg/controllers/reconcilers/cluster.go](../pkg/controllers/reconcilers/cluster.go).

#### Claim Controller

Triggered for _Hive_'s _ClusterClaim_ resources, filtered by a target annotation,
`multicluster-resiliency-addon/previous-spoke`. In charge of determining the _Claim_ status, if completed, it will
execute the [actions](#actions) required for replacing a cluster.

### Webhook

The only _Admission Webhook_ for the _Addon_ is a _Validating_ one. Used to restrict creation, modification, and
deletion of _ResilientCluster_ resources to the _ServiceAccount_ used for the _Manager_ deployment.<br/>
The code can be found in
[pkg/controllers/webhooks/validate_resilientcluster.go](../pkg/controllers/webhooks/validate_resilientcluster.go).

### Actions

Actions are preformed whenever a new cluster is up and running, and ready to replace its predecessor.

The execution function for preforming the actions is in
[pkg/controllers/actions/actions.go](../pkg/controllers/actions/actions.go).<br/>
The various action implementations are stored in [pkg/controllers/actions](../pkg/controllers/actions) and follow this
template:

```go
package actions

func theActionName(ctx context.Context, options Options) {
    // action code goes here
}

func init() {
    actionFuncs = append(actionFuncs, theActionName)
}
```

Existing action implementations info can be found in the [Action document](actions.md). 

[Go Back](../README.md#documentation)
