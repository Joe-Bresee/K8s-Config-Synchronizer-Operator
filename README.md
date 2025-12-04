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

- Watch a configuration source:
  - Git repository (optionally with SSH or HTTPS authentication)
- Reconcile resources on a configurable refresh interval
- Maintain status conditions:
  - `Available`: configuration successfully applied
  - `Progressing`: reconciliation is ongoing
  - `Degraded`: errors detected (invalid source, apply failures, etc.)
- Apply changes to multiple targets

 ## Technologies

 - Language: Go (module-based project)
 - Operator framework: Kubebuilder / controller-runtime
 - Git library: `go-git` for cloning and fetching repositories
 - YAML parsing: `sigs.k8s.io/yaml`
 - Build & tooling: `Makefile`, `go` toolchain, `kustomize` for manifests
 - Local testing: `kind` for local Kubernetes clusters, Docker for images


## Quickstart

1. Create a Kind cluster (or use any Kubernetes cluster):

```bash
kind create cluster --name config-sync
```

2. Build the controller image and load it into the Kind cluster:

```bash
make docker-build IMG=controller:latest
kind load docker-image controller:latest --name config-sync
```

3. Install CRDs into the cluster
```bash
make install
```

4. Deploy the controller

```bash
make deploy IMG=controller:latest
```

5. Apply the sample `ConfigSync` CR to trigger reconciliation:

```bash
kubectl apply -f config/samples/configs_v1alpha1_configsync.yaml
```

Notes:
- `make install` installs the CRD so Kubernetes recognizes `ConfigSync` resources — run it before applying any `ConfigSync` CRs (and before `make deploy` is safest if you changed CRDs).
- If you prefer running the controller locally (no image build), run `make install` once, then `make run` to start the controller against your kubeconfig.
- If you know what you're doing, feel free to read the makefile and any other code to customize it/run it the way you like. I'll open a Discussions page in case anyone has questions for me about this.
