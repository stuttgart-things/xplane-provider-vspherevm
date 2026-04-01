# xplane-provider-vspherevm

Crossplane Provider for VMware vSphere Virtual Machines, built with [Upjet](https://github.com/crossplane/upjet) from the [terraform-provider-vsphere](https://github.com/vmware/terraform-provider-vsphere).

## Managed Resources

| Kind | Terraform Resource | Description |
|------|-------------------|-------------|
| `VirtualMachine` | `vsphere_virtual_machine` | Full VM lifecycle (create, clone, customize) |
| `VirtualMachineSnapshot` | `vsphere_virtual_machine_snapshot` | Point-in-time VM snapshots |
| `VirtualMachineClass` | `vsphere_virtual_machine_class` | VM classes for Supervisor clusters |

## Getting Started

See [docs/](docs/) for full documentation.

## Developing

Run code-generation pipeline:
```console
go run cmd/generator/main.go "$PWD"
```

Run against a Kubernetes cluster:

```console
make run
```

Build, push, and install:

```console
make all
```

Build binary:

```console
make build
```

## Report a Bug

For filing bugs, suggesting improvements, or requesting new features, please
open an [issue](https://github.com/stuttgart-things/xplane-provider-vspherevm/issues).
