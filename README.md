# Config-Synchronizer-Operator

<img width="1536" height="1024" alt="ChatGPT Image Dec 4, 2025, 11_41_10 AM" src="https://github.com/user-attachments/assets/e5b584dd-c8f3-4b0b-9c84-4b2b51d45f7a" />


A Kubernetes operator that synchronizes configuration from a Git source into target resources on a cluster.

[![Watch a quick demo](https://youtu.be/75jLiJl56Ro/maxresdefault.jpg)](https://youtu.be/75jLiJl56Ro)


## Motivation

During my internship, I worked extensively with Kubernetes: deploying Helm charts, routing Ingress, debugging Pods, and creating cloud-native workflows in Argo Workflows. I found it fascinating and really wanted to explore Kubernetes internals further.

Creating an operator seemed like a perfect way to deepen my understanding while building something practical. Inspired by ArgoCD's GitOps strategy, this project provides a lightweight alternative for smaller clusters, useful for local development on `kind` or testing environments.

This project allows me to:

- Work with Git repositories and practice Go (which I've been learning in my other repo gophercises) (via [Gophercises](https://gophercises.com))  
- Learn Kubernetes concepts like CRDs, controllers, reconcilers, and status conditions  
 
MVP notes
 - This operator currently applies raw manifests from the configured source directly to the cluster.
 - Templating/rendering (Helm/Kustomize/text templates) is intentionally deferred for the MVP. See `todo.md` or `requirements.md` for planned templating work.
 - Runtime validation: the operator performs a server-side dry-run before applying manifests to catch admission/validation errors. This behavior can be disabled for tests.
- Understand reconciliation loops and GitOps workflows  
- learn how to use kubeapi and strengthen my understanding of neceessary rbacs and kubernetes interactions
- learn how to do go authentication via https and ssh
- going to learn how to support ca cert checking
- learn how to think about deployment lifecycles, gitops methodologies, environemtns/cluster/multi-tenancy

## Features

### ✅ **Currently Implemented:**
- **Git Source Integration**: Clone and fetch from Git repositories with SSH/HTTPS authentication
- **Manifest Application**: Parse and apply YAML manifests to Kubernetes resources
- **Status Management**: Track sync status with proper Kubernetes conditions (`Degraded`)
- **Reconciliation Loop**: Configurable refresh intervals with change detection via Git SHA comparison
- **Multi-Target Support**: Apply configuration to multiple Kubernetes resources from a single source
- **RBAC**: Proper role-based access controls for cluster operations

### 🚧 **Planned/In-Progress:**
- **Templating System**: Go template support for dynamic configuration rendering
- **Enhanced Validation**: Comprehensive YAML/manifest validation before application  
- **Testing Suite**: Unit tests and integration tests with envtest
- **Rollback Support**: Revert to previous Git commits
- **Pruning & Garbage Collection**: Clean up orphaned resources
- **Multi-Environment Support**: Branch/environment-specific configurations

 ## Technologies

 - Language: Go (module-based project)
 - Operator framework: Kubebuilder / controller-runtime
 - Git library: `go-git` for cloning and fetching repositories
 - YAML parsing: `sigs.k8s.io/yaml`
 - Build & tooling: `Makefile`, `go` toolchain, `kustomize` for manifests
 - Local testing: `kind` for local Kubernetes clusters, Docker for images


## Project Status

**Current State**: MVP is functional with core GitOps capabilities. The operator can successfully:
- Clone Git repositories and detect changes
- Apply Kubernetes manifests to cluster resources
- Track synchronization status and handle failures

**Development Phase**: Testing and enhancement phase - core functionality works but needs comprehensive testing and additional features.

## Quickstart

### Prerequisites
Ensure you have the following tools installed:
- Go 1.21+
- Docker
- kubectl 
- kind (for local testing)
- kubebuilder (for development)

### Quick Setup

1. **Create a Kind cluster:**
```bash
kind create cluster --name config-sync
```

2. **Install CRDs and deploy the operator:**
```bash
# Install the ConfigSync CRD
make install

# Run the controller locally (recommended for development)
make run
```

**OR** build and deploy as container:

```bash
# Build controller image and load into Kind cluster
make docker-build IMG=controller:latest
kind load docker-image controller:latest --name config-sync

# Deploy the controller
make deploy IMG=controller:latest
```

3. **Test with a sample ConfigSync:**
```bash
kubectl apply -f config/samples/configs_v1alpha1_configsync.yaml
```

### Development Commands
```bash
# Generate CRDs and code after API changes
make manifests generate

# Build the project
make build

# Run tests (requires envtest binaries - see Known Issues)
make test

# View available targets
make help
```

## Known Issues & Limitations

1. **Testing Infrastructure**: Tests require envtest binaries that aren't currently installed. Run `make envtest` to install them.
2. **Templating**: Go template support is planned but not yet implemented.
3. **Rollback**: Rollback to previous Git commits is not yet implemented.
4. **Multi-branch**: Environment-specific branch support is planned.
