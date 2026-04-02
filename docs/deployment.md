# Deployment

## Container Image

The provider image is built as a multi-stage Docker image (Go builder + distroless runtime) and pushed to GHCR:

```
ghcr.io/stuttgart-things/xplane-provider-vspherevm:<version>
ghcr.io/stuttgart-things/xplane-provider-vspherevm:latest
```

Each [GitHub release](https://github.com/stuttgart-things/xplane-provider-vspherevm/releases) publishes a semver-tagged image.

Dockerfile location: `cluster/images/xplane-provider-vspherevm/Dockerfile`

## Crossplane xpkg

The Crossplane package (xpkg) embeds the runtime image, CRDs, and package metadata:

```
ghcr.io/stuttgart-things/xplane-provider-vspherevm-xpkg:<version>
ghcr.io/stuttgart-things/xplane-provider-vspherevm-xpkg:latest
```

Build the xpkg locally:

```bash
crossplane xpkg build \
  --package-root=package \
  --embed-runtime-image=ghcr.io/stuttgart-things/xplane-provider-vspherevm:<version> \
  -o xplane-provider-vspherevm.xpkg
```

### Install via Crossplane

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: xplane-provider-vspherevm
spec:
  package: ghcr.io/stuttgart-things/xplane-provider-vspherevm-xpkg:v0.1.0  # or :latest
```

### Verify

```bash
# Check the provider is installed and healthy
kubectl get providers xplane-provider-vspherevm

# Check all 11 CRDs were installed
kubectl get crds | grep vspherevm

# Check VirtualMachine resources
kubectl get virtualmachine
```

## ProviderConfig Setup

Create the vSphere credentials Secret and ProviderConfig:

```bash
kubectl apply -f - <<'EOF'
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
---
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
EOF
```

The credentials are extracted as JSON and mapped to Terraform provider arguments in `internal/clients/vsphere.go`:

```go
ps.Configuration = map[string]any{
    "user":                 creds["user"],
    "password":             creds["password"],
    "vsphere_server":       creds["server"],
    "allow_unverified_ssl": creds["allow_unverified_ssl"],
}
```

## Local Development

```bash
# Init build submodule (first time)
make submodules

# Install CRDs and start the provider locally
kubectl apply -R -f package/crds
go run cmd/provider/main.go \
  --debug \
  --terraform-version=1.5.7 \
  --terraform-provider-source=vmware/vsphere \
  --terraform-provider-version=2.15.0
```

## Provider Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--debug` / `-d` | `false` | Enable debug logging |
| `--leader-election` / `-l` | `false` | Enable leader election for HA |
| `--poll` | `10m` | How often to check each resource for drift |
| `--max-reconcile-rate` | `10` | Max reconciliations per second |
| `--terraform-version` | (required) | Terraform CLI version |
| `--terraform-provider-source` | (required) | TF provider source (`vmware/vsphere`) |
| `--terraform-provider-version` | (required) | TF provider version (`2.15.0`) |
