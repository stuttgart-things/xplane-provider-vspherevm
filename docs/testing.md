# Testing

## Kind Cluster Setup

### 1. Create a Kind cluster

```bash
cat <<'EOF' > /tmp/crossplane-vsphere-test.yaml
kind: Cluster
name: crossplane-vsphere-test
apiVersion: kind.x-k8s.io/v1alpha4
featureGates:
  ImageVolume: True
networking:
  apiServerAddress: '10.100.136.192'
  disableDefaultCNI: True
  kubeProxyMode: none
nodes:
  - role: control-plane
    image: kindest/node:v1.35.0
    extraPortMappings:
      - containerPort: 6443
        hostPort: 34361
        protocol: TCP
  - role: worker
    image: kindest/node:v1.35.0
EOF

kind create cluster --config /tmp/crossplane-vsphere-test.yaml
kind get kubeconfig --name crossplane-vsphere-test > ~/.kube/vsphere-test
yq -i '.clusters[0].cluster.server |= sub("0\.0\.0\.0", "10.100.136.192")' ~/.kube/vsphere-test
```

### 2. Install Cilium

```bash
kubectl apply -k https://github.com/stuttgart-things/helm/infra/crds/cilium

dagger call -m github.com/stuttgart-things/dagger/helm@v0.57.0 \
  helmfile-operation \
  --helmfile-ref "git::https://github.com/stuttgart-things/helm.git@infra/cilium.yaml.gotmpl" \
  --operation apply \
  --state-values "config=kind,clusterName=vsphere-test,configureLB=false" \
  --kube-config file://$HOME/.kube/vsphere-test \
  --progress plain -vv
```

### 3. Install Crossplane

```bash
dagger call -m github.com/stuttgart-things/dagger/helm@v0.57.0 \
  helmfile-operation \
  --helmfile-ref "git::https://github.com/stuttgart-things/helm.git@cicd/crossplane.yaml.gotmpl" \
  --operation apply \
  --state-values "version=2.2.0" \
  --kube-config file://$HOME/.kube/vsphere-test \
  --progress plain -vv
```

### 4. Install CRDs and run the provider

```bash
kubectl apply -R -f package/crds

go run cmd/provider/main.go \
  --debug \
  --terraform-version=1.5.7 \
  --terraform-provider-source=vmware/vsphere \
  --terraform-provider-version=2.15.0
```

## End-to-End Test with Released Provider Package

After the cluster and Crossplane are ready, install the released provider xpkg and test.

### 1. Install the provider

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: xplane-provider-vspherevm
spec:
  package: ghcr.io/stuttgart-things/xplane-provider-vspherevm-xpkg:v0.1.0
```

```bash
kubectl get providers xplane-provider-vspherevm
# Wait until INSTALLED=True and HEALTHY=True
```

### 2. Create ProviderConfig with vSphere credentials

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

### 3. Create a test VirtualMachine

```bash
kubectl apply -f - <<'EOF'
apiVersion: virtualmachine.vspherevm.stuttgart-things.com/v1alpha1
kind: VirtualMachine
metadata:
  name: test-vm
spec:
  forProvider:
    name: test-vm
    resourcePoolId: ${RESOURCE_POOL_ID}
    datastoreId: ${DATASTORE_ID}
    numCpus: 2
    memory: 2048
    guestId: ubuntu64Guest
    networkInterface:
      - networkId: ${NETWORK_ID}
    disk:
      - size: 20
        thinProvisioned: true
  providerConfigRef:
    name: default
EOF
```

### 4. Verify

```bash
# VirtualMachine should be Ready + Synced
kubectl get virtualmachine test-vm

# Check events for details
kubectl describe virtualmachine test-vm

# Check provider logs
kubectl logs -n crossplane-system -l pkg.crossplane.io/revision -c package-runtime --tail=50
```

## Unit Tests

```bash
go test ./... -v -count=1
```

## Lint

```bash
golangci-lint run ./...
```

## Code Generation Verification

After any code gen changes, verify:

```bash
# All 11 CRDs generated
ls package/crds/ | wc -l
# expected: 11

# Build passes
go build ./...

# CRDs are valid
kubectl apply --dry-run=server -R -f package/crds/
```

## Testdata

| File | Description |
|------|-------------|
| `examples/cluster/providerconfig/providerconfig.yaml` | ProviderConfig with credentials |
| `examples/install.yaml` | Provider installation manifest |
