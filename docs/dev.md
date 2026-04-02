# Provider Development

Generating the Crossplane provider from the Terraform vSphere provider using Upjet. Follow the steps exactly in order.

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | >= 1.24 | Build toolchain |
| goimports | latest | Required by Upjet code generation |
| Docker | >= 24 | Container builds |
| kubectl | >= 1.28 | Cluster interaction |
| crossplane CLI | >= 2.0 | Package builds |

```bash
# Install goimports (required by Upjet generator)
go install golang.org/x/tools/cmd/goimports@latest
```

## Init

### 1. Clone the repo (scaffolded from upjet-provider-template)

```bash
git clone https://github.com/stuttgart-things/xplane-provider-vspherevm.git
cd xplane-provider-vspherevm
```

The repo was created from the [upjet-provider-template](https://github.com/crossplane/upjet-provider-template) with the following renames applied:

| Item | Template value | Our value |
|------|---------------|-----------|
| Go module | `github.com/crossplane/upjet-provider-template` | `github.com/stuttgart-things/xplane-provider-vspherevm` |
| Project name | `upjet-provider-template` | `xplane-provider-vspherevm` |
| Resource prefix | `template` | `vspherevm` |
| Root API group | `template.crossplane.io` | `vspherevm.stuttgart-things.com` |
| Namespaced API group | `template.m.crossplane.io` | `vspherevm.m.stuttgart-things.com` |
| Registry | `ghcr.io/crossplane-contrib` | `ghcr.io/stuttgart-things` |

### 2. Init the build submodule

```bash
make submodules
```

### 3. Generate the Terraform provider schema

The Makefile fetches the provider binary via `terraform init` and extracts the schema. No Go dependency on the Terraform provider is needed.

**Important:** The provider source is `vmware/vsphere`, not `hashicorp/vsphere` (deprecated).

```bash
make generate.init
```

This produces:

- `config/schema.json` -- 198KB provider schema with all vSphere resources
- `.work/vmware/vsphere/` -- cloned Terraform docs for the scraper

### 4. Configure VM resources

Resource configs live under `config/` with separate directories for cluster and namespaced scopes:

```
config/
├── provider.go                       # Provider-level config (included resources, root group)
├── external_name.go                  # External name configs for all resources
├── cluster/virtualmachine/config.go  # vsphere_virtual_machine (cluster scope)
├── namespaced/virtualmachine/config.go  # vsphere_virtual_machine (namespaced scope)
├── provider-metadata.yaml            # Resource metadata (descriptions, docs)
└── schema.json                       # Generated Terraform provider schema
```

In `config/provider.go`, the provider is configured with root group and resource prefix:

```go
const (
    resourcePrefix = "vspherevm"
    modulePath     = "github.com/stuttgart-things/xplane-provider-vspherevm"
)

// GetProvider returns provider configuration
func GetProvider() *ujconfig.Provider {
    pc := ujconfig.NewProvider([]byte(providerSchema), resourcePrefix, modulePath, []byte(providerMetadata),
        ujconfig.WithRootGroup("vspherevm.stuttgart-things.com"),
        ujconfig.WithIncludeList(ExternalNameConfigured()),
        ...
    )
    ...
}
```

In `config/external_name.go`, all three VM resources are registered:

```go
var ExternalNameConfigs = map[string]config.ExternalName{
    "vsphere_virtual_machine":          config.IdentifierFromProvider,
    "vsphere_virtual_machine_snapshot": config.IdentifierFromProvider,
    "vsphere_virtual_machine_class":    config.IdentifierFromProvider,
}
```

Each resource config sets the short group and kind:

```go
// config/cluster/virtualmachine/config.go
func Configure(p *ujconfig.Provider) {
    p.AddResourceConfigurator("vsphere_virtual_machine", func(r *ujconfig.Resource) {
        r.ShortGroup = "virtualmachine"
        r.Kind = "VirtualMachine"
    })
}
```

### 5. Run code generation

```bash
# Ensure goimports is on PATH
export PATH=$PATH:$(go env GOPATH)/bin

go run cmd/generator/main.go "$PWD"
```

Output:

```
Generated 3 resources with scope Cluster!
Generated 3 resources with scope Namespaced!
```

This generates:

- API types in `apis/cluster/` and `apis/namespaced/` (Go structs with JSON tags)
- Controllers in `internal/controller/`
- `zz_*` generated files (do not edit)

### 6. Generate deepcopy, CRDs, and method sets

The Upjet generator produces types but not deepcopy methods or CRDs. Run these separately:

```bash
# Generate deepcopy methods and CRD manifests
go run sigs.k8s.io/controller-tools/cmd/controller-gen \
  object:headerFile=hack/boilerplate.go.txt \
  paths=./apis/... \
  crd:allowDangerousTypes=true,crdVersions=v1 \
  output:artifacts:config=package/crds

# Generate crossplane-runtime method sets (managed resource interfaces)
go run github.com/crossplane/crossplane-tools/cmd/angryjet \
  generate-methodsets \
  --header-file=hack/boilerplate.go.txt \
  ./apis/...
```

**Note:** The `go generate ./apis/...` command also runs a doc scraper that may fail for some providers (including vsphere) due to doc format differences. Run controller-gen and angryjet directly if the scraper fails.

### 7. Build and verify

```bash
go mod tidy
go build ./...

# Verify CRDs were created (11 total)
ls package/crds/
```

## Generated CRDs

| CRD | Scope | Description |
|-----|-------|-------------|
| `virtualmachines.virtualmachine.vspherevm.stuttgart-things.com` | Cluster | VirtualMachine |
| `machinesnapshots.virtual.vspherevm.stuttgart-things.com` | Cluster | MachineSnapshot |
| `machineclasses.virtual.vspherevm.stuttgart-things.com` | Cluster | MachineClass |
| `virtualmachines.virtualmachine.vspherevm.m.stuttgart-things.com` | Namespaced | VirtualMachine |
| `machinesnapshots.virtual.vspherevm.m.stuttgart-things.com` | Namespaced | MachineSnapshot |
| `machineclasses.virtual.vspherevm.m.stuttgart-things.com` | Namespaced | MachineClass |
| `providerconfigs.vspherevm.stuttgart-things.com` | Cluster | ProviderConfig |
| `providerconfigusages.vspherevm.stuttgart-things.com` | Cluster | ProviderConfigUsage |
| `providerconfigs.vspherevm.m.stuttgart-things.com` | Namespaced | ProviderConfig |
| `providerconfigusages.vspherevm.m.stuttgart-things.com` | Namespaced | ProviderConfigUsage |
| `clusterproviderconfigs.vspherevm.m.stuttgart-things.com` | Cluster | ClusterProviderConfig |

## Updating After Terraform Provider Changes

```bash
# Update provider version in Makefile
# TERRAFORM_PROVIDER_VERSION ?= 2.X.Y

# Re-fetch schema
rm -rf .work/terraform config/schema.json
make generate.init

# Re-run code generation
export PATH=$PATH:$(go env GOPATH)/bin
go run cmd/generator/main.go "$PWD"

# Regenerate deepcopy + CRDs
go run sigs.k8s.io/controller-tools/cmd/controller-gen \
  object:headerFile=hack/boilerplate.go.txt paths=./apis/... \
  crd:allowDangerousTypes=true,crdVersions=v1 output:artifacts:config=package/crds

go run github.com/crossplane/crossplane-tools/cmd/angryjet \
  generate-methodsets --header-file=hack/boilerplate.go.txt ./apis/...

go mod tidy && go build ./...
```

## Project Structure

```
xplane-provider-vspherevm/
├── apis/
│   ├── cluster/
│   │   ├── virtualmachine/v1alpha1/   # VirtualMachine types
│   │   ├── virtual/v1alpha1/          # MachineSnapshot, MachineClass types
│   │   ├── v1alpha1/                  # API group registration
│   │   ├── v1beta1/                   # ProviderConfig types
│   │   └── zz_register.go            # Schema registration
│   ├── namespaced/                    # Same structure (namespaced scope)
│   └── generate.go                   # go:generate directives
├── cmd/
│   ├── generator/main.go            # Upjet code generation entrypoint
│   └── provider/main.go             # Provider binary entrypoint
├── config/
│   ├── provider.go                   # Provider config (root group, resource prefix)
│   ├── external_name.go              # External name configs
│   ├── cluster/virtualmachine/       # Cluster-scope resource config
│   ├── namespaced/virtualmachine/    # Namespaced-scope resource config
│   ├── schema.json                   # Generated TF provider schema (198KB)
│   └── provider-metadata.yaml        # Resource metadata
├── internal/
│   ├── clients/vsphere.go           # Terraform setup + credential extraction
│   ├── controller/                   # Generated controllers
│   ├── features/                     # Feature flags
│   └── version/                      # Version info
├── package/
│   ├── crds/                        # Generated CRD manifests (11 files)
│   └── crossplane.yaml              # Crossplane package metadata
├── cluster/images/xplane-provider-vspherevm/
│   └── Dockerfile                   # Provider container image
├── build/                           # Build submodule (crossplane/build)
└── docs/                            # This documentation
```

## Gotchas

- **Provider source:** Use `vmware/vsphere`, not `hashicorp/vsphere` (deprecated, no v2.x releases)
- **No Go dependency:** The TF provider is NOT imported as a Go module. Upjet uses `terraform init` to fetch the binary and extract the schema via `terraform providers schema -json`
- **goimports required:** The Upjet generator calls `goimports` -- install it before running code gen
- **Doc scraper may fail:** The vsphere provider docs don't match the expected HTML format. Skip the scraper and write `provider-metadata.yaml` manually
- **Old template files:** After initial scaffold, remove all `null_resource` / `template.crossplane.io` files before running code gen to avoid conflicts
