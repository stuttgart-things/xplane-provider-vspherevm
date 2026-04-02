# Testing

## Prerequisites

- Kubernetes cluster (Kind, k3s, or any cluster)
- [Crossplane](https://crossplane.io) installed
- vSphere/vCenter access with VM creation permissions
- `kubectl` configured

## Quick Start: Deploy and Test

### 1. Install Crossplane

```bash
helm repo add crossplane https://charts.crossplane.io/stable
helm install crossplane crossplane/crossplane \
  --namespace crossplane-system --create-namespace --wait
```

### 2. Install the provider

```bash
kubectl apply -f https://raw.githubusercontent.com/stuttgart-things/xplane-provider-vspherevm/main/examples/install.yaml
```

Wait for the provider to become healthy:

```bash
kubectl get providers xplane-provider-vspherevm -w
# NAME                         INSTALLED   HEALTHY   PACKAGE                                                              AGE
# xplane-provider-vspherevm    True        True      ghcr.io/stuttgart-things/xplane-provider-vspherevm-xpkg:latest       30s
```

Verify CRDs are registered (11 total):

```bash
kubectl get crds | grep vspherevm
```

### 3. Configure vSphere credentials

```bash
kubectl apply -f - <<'EOF'
apiVersion: v1
kind: Secret
metadata:
  name: vsphere-creds
  namespace: crossplane-system
type: Opaque
stringData:
  credentials: |
    {
      "user": "administrator@vsphere.local",
      "password": "<VSPHERE_PASSWORD>",
      "server": "<VCENTER_FQDN_OR_IP>",
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
      name: vsphere-creds
      namespace: crossplane-system
      key: credentials
EOF
```

### 4. Get vSphere resource IDs

You need managed object reference IDs for resource pool, datastore, network, and template UUID. Use `govc` or the vSphere UI:

```bash
# Install govc if needed: https://github.com/vmware/govmomi/releases
export GOVC_URL=https://<VCENTER_FQDN_OR_IP>/sdk
export GOVC_USERNAME=administrator@vsphere.local
export GOVC_PASSWORD=<password>
export GOVC_INSECURE=true

# Resource Pool ID
govc pool.info -json /LabUL/host/Cluster-V6.7/Resources | jq -r '.resourcePools[0].self.value'
# → e.g. resgroup-481

# Datastore ID
govc datastore.info -json /LabUL/datastore/UL-ESX-SAS-02 | jq -r '.datastores[0].self.value'
# → e.g. datastore-255

# Network ID (list all networks with their MOIDs)
govc ls -l -I /LabUL/network/
# → e.g. network-263  /LabUL/network/LAB-10.31.103

# Template UUID
govc vm.info -json <template-name> | jq -r '.virtualMachines[0].config.uuid'
# → e.g. 423483d0-5dd4-def9-5c87-94c0f513bab4
```

### 5. Create a test VirtualMachine

```bash
kubectl apply -f - <<'EOF'
apiVersion: virtualmachine.vspherevm.stuttgart-things.com/v1alpha1
kind: VirtualMachine
metadata:
  name: test-vm
spec:
  forProvider:
    name: crossplane-test-vm
    resourcePoolId: "<RESOURCE_POOL_MOID>"
    datastoreId: "<DATASTORE_MOID>"
    numCpus: 2
    memory: 4096
    guestId: ubuntu64Guest
    networkInterface:
      - networkId: "<NETWORK_MOID>"
    disk:
      - size: 40
        thinProvisioned: true
  providerConfigRef:
    name: default
EOF
```

### 6. Verify

```bash
# VirtualMachine should become Ready + Synced
kubectl get virtualmachine test-vm -w

# Check events for details
kubectl describe virtualmachine test-vm

# Check provider logs
kubectl logs -n crossplane-system -l pkg.crossplane.io/revision -c package-runtime --tail=50
```

### 7. Cleanup

```bash
kubectl delete virtualmachine test-vm
# Wait for the VM to be deleted from vSphere, then remove the provider
kubectl delete provider xplane-provider-vspherevm
```

---

## Kind Cluster Setup (detailed)

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

---

## Local Development Run

Run the provider locally against a cluster (useful for debugging):

```bash
# Install CRDs first
kubectl apply -R -f package/crds

# Run the provider
go run cmd/provider/main.go \
  --debug \
  --terraform-version=1.5.7 \
  --terraform-provider-source=vmware/vsphere \
  --terraform-provider-version=2.15.0
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
| `examples/install.yaml` | Provider installation manifest (latest) |
| `examples/cluster/providerconfig/secret.yaml` | Example credentials Secret |
| `examples/cluster/providerconfig/providerconfig.yaml` | ProviderConfig |
| `examples/cluster/virtualmachine/virtualmachine.yaml` | VirtualMachine example |
| `examples/cluster/virtual/machinesnapshot.yaml` | MachineSnapshot example |
| `examples/cluster/virtual/machineclass.yaml` | MachineClass example |
