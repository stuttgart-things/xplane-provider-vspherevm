# xplane-provider-vspherevm

Crossplane Provider for VMware vSphere Virtual Machines, built with [Upjet](https://github.com/crossplane/upjet) from the [terraform-provider-vsphere](https://github.com/vmware/terraform-provider-vsphere) (v2.15.0). It exposes vSphere VM resources as Crossplane Managed Resources (MRs).

## Features

- **Upjet-generated** -- auto-generated CRDs, controllers, and types from the Terraform vSphere provider schema
- **Virtual Machine lifecycle** -- create, update, delete vSphere VMs declaratively via Kubernetes
- **VM Snapshots** -- manage VM snapshots as Crossplane resources
- **VM Classes** -- define and manage VM classes (Supervisor/vSphere with Tanzu)
- **Drift detection** -- Crossplane reconciliation loop detects and corrects config drift
- **Dual scope** -- all resources available as both cluster-scoped and namespaced

## Managed Resources

| Kind | Terraform Resource | API Group (cluster) | Description |
|------|-------------------|---------------------|-------------|
| `VirtualMachine` | `vsphere_virtual_machine` | `virtualmachine.vspherevm.stuttgart-things.com` | Full VM lifecycle (create, clone, customize) |
| `MachineSnapshot` | `vsphere_virtual_machine_snapshot` | `virtual.vspherevm.stuttgart-things.com` | Point-in-time VM snapshots |
| `MachineClass` | `vsphere_virtual_machine_class` | `virtual.vspherevm.stuttgart-things.com` | VM classes for Supervisor clusters |

Namespaced variants use `*.vspherevm.m.stuttgart-things.com` API groups.

## How It Works

```
┌──────────────────┐     ┌──────────────────┐     ┌──────────────────┐
│  VirtualMachine  │────>│  Upjet Controller│────>│  vSphere API     │
│  CR (K8s)        │     │  (Terraform TF)  │     │  (vCenter)       │
└──────────────────┘     └──────────────────┘     └──────────────────┘
                               │
                         ┌─────┴─────┐
                         │ terraform │
                         │ vmware/   │
                         │ vsphere   │
                         │ v2.15.0   │
                         └───────────┘
```

Upjet wraps the Terraform provider as a Crossplane controller. Each Managed Resource maps 1:1 to a Terraform resource. The Upjet runtime calls Terraform plan/apply under the hood.

## Quick Start

### Step 1: Install the Provider

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: xplane-provider-vspherevm
spec:
  package: ghcr.io/stuttgart-things/xplane-provider-vspherevm-xpkg:v0.1.0
```

### Step 2: Create a ProviderConfig

```yaml
apiVersion: vspherevm.stuttgart-things.com/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      name: vsphere-credentials
      namespace: crossplane-system
      key: credentials
```

The credentials Secret must contain the vSphere connection parameters as JSON:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: vsphere-credentials
  namespace: crossplane-system
stringData:
  credentials: |
    {
      "user": "administrator@vsphere.local",
      "password": "your-password",
      "server": "vcenter.example.com",
      "allow_unverified_ssl": "true"
    }
```

The provider maps these JSON keys to Terraform provider arguments:

| JSON key | Terraform argument | Description |
|----------|-------------------|-------------|
| `user` | `user` | vSphere API username |
| `password` | `password` | vSphere API password |
| `server` | `vsphere_server` | vCenter server address |
| `allow_unverified_ssl` | `allow_unverified_ssl` | Skip TLS verification |

### Step 3: Create a VirtualMachine

```yaml
apiVersion: virtualmachine.vspherevm.stuttgart-things.com/v1alpha1
kind: VirtualMachine
metadata:
  name: my-vm
spec:
  forProvider:
    name: my-vm
    resourcePoolId: ${RESOURCE_POOL_ID}
    datastoreId: ${DATASTORE_ID}
    numCpus: 2
    memory: 4096
    guestId: ubuntu64Guest
    networkInterface:
      - networkId: ${NETWORK_ID}
    disk:
      - size: 40
        thinProvisioned: true
  providerConfigRef:
    name: default
```

### Step 4: Clone from Template

```yaml
apiVersion: virtualmachine.vspherevm.stuttgart-things.com/v1alpha1
kind: VirtualMachine
metadata:
  name: my-cloned-vm
spec:
  forProvider:
    name: my-cloned-vm
    resourcePoolId: ${RESOURCE_POOL_ID}
    datastoreId: ${DATASTORE_ID}
    numCpus: 4
    memory: 8192
    clone:
      - templateUuid: ${TEMPLATE_UUID}
        customize:
          - linuxOptions:
              - hostName: my-cloned-vm
                domain: example.com
            networkInterface:
              - ipv4Address: 10.0.0.10
                ipv4Netmask: 24
    networkInterface:
      - networkId: ${NETWORK_ID}
    disk:
      - size: 80
  providerConfigRef:
    name: default
```

## Verify

```bash
$ kubectl get virtualmachine
NAME           READY   SYNCED   EXTERNAL-NAME   AGE
my-vm          True    True     my-vm           5m
my-cloned-vm   True    True     my-cloned-vm    3m

$ kubectl get machinesnapshot
NAME          READY   SYNCED   AGE
pre-upgrade   True    True     1m

# Check all installed CRDs
$ kubectl get crds | grep vspherevm
machineclasses.virtual.vspherevm.stuttgart-things.com
machinesnapshots.virtual.vspherevm.stuttgart-things.com
virtualmachines.virtualmachine.vspherevm.stuttgart-things.com
...
```
