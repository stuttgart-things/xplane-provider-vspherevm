# xplane-provider-vspherevm

Crossplane Provider for VMware vSphere Virtual Machines, built with [Upjet](https://github.com/crossplane/upjet) from [terraform-provider-vsphere](https://github.com/vmware/terraform-provider-vsphere) (v2.15.0).

## Managed Resources

| Kind | API Group | Terraform Resource | Scope |
|------|-----------|-------------------|-------|
| `VirtualMachine` | `virtualmachine.vspherevm.stuttgart-things.com` | `vsphere_virtual_machine` | Cluster + Namespaced |
| `MachineSnapshot` | `virtual.vspherevm.stuttgart-things.com` | `vsphere_virtual_machine_snapshot` | Cluster + Namespaced |
| `MachineClass` | `virtual.vspherevm.stuttgart-things.com` | `vsphere_virtual_machine_class` | Cluster + Namespaced |

Namespaced variants use `*.vspherevm.m.stuttgart-things.com` API groups.

## Install

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: xplane-provider-vspherevm
spec:
  package: ghcr.io/stuttgart-things/xplane-provider-vspherevm-xpkg:v0.2.1
```

## Configure Credentials

Create a Secret with vSphere credentials:

```bash
kubectl create secret generic vsphere-creds -n crossplane-system \
  --from-literal=credentials='{"user":"administrator@vsphere.local","password":"...","server":"vcenter.example.com","allow_unverified_ssl":"true"}'
```

Apply a ProviderConfig:

```yaml
apiVersion: vspherevm.stuttgart-things.com/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      name: vsphere-creds
      namespace: crossplane-system
      key: credentials
```

## Example: Create a VM

```yaml
apiVersion: virtualmachine.vspherevm.stuttgart-things.com/v1alpha1
kind: VirtualMachine
metadata:
  name: my-vm
spec:
  forProvider:
    name: my-crossplane-vm
    resourcePoolId: "${RESOURCE_POOL_ID}"
    datastoreId: "${DATASTORE_ID}"
    numCpus: 2
    memory: 4096
    guestId: ubuntu64Guest
    networkInterface:
      - networkId: "${NETWORK_ID}"
    disk:
      - size: 40
        thinProvisioned: true
  providerConfigRef:
    name: default
```

## Documentation

- [Overview](docs/index.md) -- features, architecture, credential mapping
- [Development](docs/dev.md) -- code generation, prerequisites, gotchas
- [Deployment](docs/deployment.md) -- container image, xpkg, local dev
- [CI/CD](docs/cicd.md) -- GitHub Actions workflows, semantic-release
- [Testing](docs/testing.md) -- Kind cluster, e2e tests

## Development

Run code generation:
```console
go run cmd/generator/main.go "$PWD"
```

Build:
```console
go build ./...
```

## Report a Bug

Open an [issue](https://github.com/stuttgart-things/xplane-provider-vspherevm/issues).
