# MyApp Operator — PoC

> A Kubernetes Operator written in Go that manages `MyApp` custom resources.

## What It Does

When you create a `MyApp` CR, the operator automatically:

1. **Creates a Deployment** with the specified image, replicas, and port
2. **Creates a Service** (ClusterIP) to expose the app
3. **Watches for changes** — if you update replicas or image, it reconciles
4. **Reports status** — `ReadyReplicas`, `Phase`, `Conditions`
5. **Handles deletion** — finalizers ensure clean resource removal

## Architecture

```
┌──────────────┐     watches     ┌────────────────┐     manages     ┌──────────────┐
│   MyApp CR   │ ──────────────► │   Controller   │ ──────────────► │  Deployment  │
│  (desired)   │                 │  (reconcile)   │                 │  + Service   │
└──────────────┘                 └────────────────┘                 └──────────────┘
     user writes                    compares &                        actual state
     the YAML                       converges                        in the cluster
```

## Project Structure

```
myapp-operator/
├── main.go                          # Entrypoint — sets up Manager + Controller
├── api/v1alpha1/
│   ├── myapp_types.go               # CRD types (Spec + Status structs)
│   ├── groupversion_info.go         # Schema registration
│   └── zz_generated.deepcopy.go     # DeepCopy implementations
├── controllers/
│   └── myapp_controller.go          # THE RECONCILE LOOP — core logic
├── config/
│   ├── crd/myapp-crd.yaml           # CRD manifest to install in K8s
│   ├── rbac/role.yaml               # RBAC permissions
│   ├── manager/manager.yaml         # Operator Deployment manifest
│   └── samples/                     # MyApp CRs to test with
│       ├── myapp-sample-nginx.yaml
│       ├── myapp-sample-httpbin.yaml
│       └── myapp-sample-scale-test.yaml
├── Dockerfile                       # Multi-stage build (distroless)
├── Makefile                         # Common targets
└── go.mod
```

## Prerequisites

- A running Kubernetes cluster (`kind`, `minikube`, `k3d`, or real cluster)
- `kubectl` configured
- Go 1.22+

**First time setup:**
```bash
go mod tidy   # generates go.sum (required before building)
```

## Option A: Run Locally (recommended for demos)

```bash
# 1. Install the CRD
kubectl apply -f config/crd/myapp-crd.yaml

# 2. Run the operator locally (connects to your current kubecontext)
go run main.go

# 3. In another terminal — create a MyApp
kubectl apply -f config/samples/myapp-sample-nginx.yaml

# 4. Watch it work
kubectl get myapps -w
kubectl get deployments
kubectl get pods
```

## Option B: Run In-Cluster

```bash
# Build and load image
make docker-build IMG=myapp-operator:latest

# For kind:
kind load docker-image myapp-operator:latest

# Deploy everything
make deploy

# Create a sample
make sample-nginx

# Watch
make watch
```


## Key Concepts Demonstrated

| Concept | Where in code |
|---------|---------------|
| **CRD (Custom Resource Definition)** | `api/v1alpha1/myapp_types.go` + `config/crd/myapp-crd.yaml` |
| **Reconciliation Loop** | `controllers/myapp_controller.go` — `Reconcile()` |
| **Owner References** | `ctrl.SetControllerReference()` in the controller |
| **Finalizers** | Deletion handling in `Reconcile()` |
| **Status Subresource** | `r.Status().Update()` at the end of reconcile |
| **RBAC (least privilege)** | `config/rbac/role.yaml` |
| **Watches (primary + secondary)** | `SetupWithManager()` — watches MyApp + owned Deployments |

## Resources

- [Operator SDK Documentation](https://sdk.operatorframework.io)
- [Kubebuilder Book](https://book.kubebuilder.io)
- [controller-runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
- [OperatorHub.io](https://operatorhub.io)
