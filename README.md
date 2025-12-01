*This project is in progress*
# Config-Synchronizer-Operator

A Kubernetes operator that synchronizes configuration from a source (Git repository, ConfigMap, or Secret) into target resources on a cluster.

as of this readme update commit, I've been able to get a functional syncing of a resource defined in a git source which is awesome. On initial start, it's able to apply the new manifest, and when I edited the deployment manifest in my test git source, my operator was able to detect the change during its reconciliation loop and apply it to the cluster. There's no pruning at the moment which might be a nice touch, and I think I may need to re-design my CR to be a bit more adaptable to manifest changes.

## Motivation

During my internship, I worked extensively with Kubernetes: deploying Helm charts, routing Ingress, debugging Pods, and creating cloud-native workflows in Argo Workflows. I found it fascinating and really wanted to explore Kubernetes internals further.

Creating an operator seemed like a perfect way to deepen my understanding while building something practical. Inspired by ArgoCD's GitOps strategy, this project provides a lightweight alternative for smaller clusters, useful for local development on `kind` or testing environments.

This project allows me to:

- Work with Git repositories and practice Go (which I've been learning in my other repo gophercises) (via [Gophercises](https://gophercises.com))  
- Learn Kubernetes concepts like CRDs, controllers, reconcilers, and status conditions  
 
MVP notes
 - This operator currently applies raw manifests from the configured source directly to the cluster.
 - Templating/rendering (Helm/Kustomize/text templates) is intentionally deferred for the MVP. See `todo.md` for planned templating work.
 - Runtime validation: the operator performs a server-side dry-run before applying manifests to catch admission/validation errors. This behavior can be disabled for tests.
- Understand reconciliation loops and GitOps workflows  
- learn how to use kubeapi and strengthen my understanding of neceessary rbacs and kubernetes interactions
- learn how to do go authentication via https and ssh
- going to learn how to support ca cert checking

## Features

- Watch a configuration source:
  - Git repository (optionally with SSH or HTTPS authentication)
  - ConfigMap
  - Secret
- Reconcile resources on a configurable refresh interval
- Maintain status conditions:
  - `Available`: configuration successfully applied
  - `Progressing`: reconciliation is ongoing
  - `Degraded`: errors detected (invalid source, apply failures, etc.)
- Apply changes to multiple target ConfigMaps or Secrets

## Installation

```bash
# Build and run the operator locally
make install
make run
```

readme in progress.