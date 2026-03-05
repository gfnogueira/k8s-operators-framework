# Kubernetes Operators with Go — Tech Talk

## Repository Contents

```
.
├── myapp-operator/         # Complete Go Operator PoC (working code)
│   ├── api/                #   CRD type definitions
│   ├── controllers/        #   Reconciliation logic
│   ├── config/             #   K8s manifests (CRD, RBAC, samples)
│   ├── main.go             #   Entrypoint
│   ├── Dockerfile          #   Container build
│   ├── Makefile            #   Common targets
│   └── README.md           #   Detailed docs + demo script

```

## Quick Start

```bash
# Go to the operator directory
cd myapp-operator

# Install CRD + run locally
kubectl apply -f config/crd/myapp-crd.yaml
go run main.go

# In another terminal — create a MyApp
kubectl apply -f config/samples/myapp-sample-nginx.yaml
kubectl get myapps -w
```

## Prerequisites

- Go 1.22+ installed
- A running Kubernetes cluster (`kind`, `minikube`, `k3d`, or real cluster)
- `kubectl` configured
- Run `go mod tidy` in `myapp-operator/` before first build to generate `go.sum`
