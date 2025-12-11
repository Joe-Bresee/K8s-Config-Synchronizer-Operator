# Config Synchronizer Operator — Copilot Requirements

## Project Summary
Build a Kubernetes operator that:
- Watches a `ConfigSync` Custom Resource (CR)
- Fetches configuration from a source (Git repo)
- Optionally applies templating or transformations
- Synchronizes it into CR-defined kubernetes resource
- Effective logging and status details, maybe metrics endpoint for observability.
- End Of Project: Good presentation, professional-looking github repo project. Maybe a demo youtube video to go along with it.

---

## CRD: ConfigSync ✅ **IMPLEMENTED**

The current CR design is flexible and supports the core use cases:

### spec ✅ **IMPLEMENTED**
- ✅ `source.git`: Git source configuration
  - ✅ `repoURL`: HTTPS/SSH URL to Git repository  
  - ✅ `path`: Path to configuration files in repository
  - ✅ `branch`: Git branch to use (optional)
  - ✅ `revision`: Git revision/commit SHA (optional)
  - ✅ `authMethod`: Authentication method (ssh/https/none)
  - ✅ `authSecretRef`: Reference to authentication secret
- ✅ `targets`: List of target Kubernetes resources
  - ✅ `namespace`: Target namespace
  - ✅ `name`: Target resource name  
  - ✅ `type`: Resource type (ConfigMap/Secret/Deployment)
- ✅ `refreshInterval`: Reconciliation interval (optional)

### status ✅ **IMPLEMENTED**
- ✅ `lastSyncedTime`: Timestamp of last sync
- ✅ `sourceRevision`: Git SHA that was applied
- ✅ `appliedTargets`: Number of successfully applied targets
- ✅ `sourcePath`: Path within repository that was applied
- ✅ `conditions`: Standard Kubernetes conditions (Available, Degraded, etc.)

---

## Operator Behavior

### Source Fetching ✅ **IMPLEMENTED**
- ✅ Clones Git repositories using go-git library
- ✅ Supports SSH and HTTPS authentication via Kubernetes secrets
- ✅ Detects changes by comparing Git SHAs
- ✅ Caches repositories to avoid unnecessary re-cloning
- ✅ Handles authentication errors and connection failures

### Templating / Transformation 🚧 **PLANNED** 
- TODO: `render_template(data: dict, target: dict) -> dict`
- TODO: Apply simple variable interpolation using Go templates
- TODO: Support environment-specific overrides
- TODO: Add validation for template syntax

### Target Application ✅ **IMPLEMENTED**
- ✅ Parses YAML manifests from Git source
- ✅ Applies changes using server-side apply (with dry-run validation)
- ✅ Handles multi-document YAML files
- ✅ Reports success/failure per target

### Reconciliation Triggers ✅ **IMPLEMENTED**
- ✅ On ConfigSync CR create/update/delete
- ✅ On configurable refresh interval
- ✅ Change detection via Git SHA comparison

### Error Handling ✅ **PARTIALLY IMPLEMENTED**
- ✅ Structured logging throughout reconciliation
- ✅ Status conditions updated (Degraded) on failures
- ✅ Proper error propagation and reporting
- TODO: Emit Kubernetes events for better observability
- TODO: Retry with exponential backoff for transient failures

### TODO: Rollbacks 🚧 **PLANNED**
- TODO: Implement rollback functionality to previous/specific commit SHA
- TODO: Add rollback triggers in ConfigSync spec

### TODO: Pruning & Garbage Collection 🚧 **PLANNED** 
- TODO: Track and clean up resources that are no longer in source
- TODO: Add finalizers for proper cleanup on ConfigSync deletion

### Apply Semantics ✅ **IMPLEMENTED**
- ✅ Uses server-side apply for conflict resolution
- ✅ Dry-run validation before actual application
- ✅ Proper error handling for validation failures

---

## Technical Stack

### Go (Kubebuilder)
- `controller-runtime`
- `go-git`
- `yaml.v3`
- Go templating package

---

## Directory Structure

config-synchronizer-operator/
README.md
requirements.md
outline.md
/api -> CRD types
/controllers -> reconciliation logic
/internal -> source fetch, parser, templates
/deploy -> manifests (CRD, RBAC, operator deployment)
Dockerfile
Makefile


---

## Example ConfigSync CR

apiVersion: configs.example.io/v1alpha1
kind: ConfigSync
metadata:
name: example-sync
spec:
source:
git:
repo: https://github.com/myorg/configs.git


path: config/app.yaml
revision: main
targets:
- namespace: default
type: ConfigMap
name: app-config
refreshInterval: 10m


---

## Stretch Goals
- Webhook triggers for Git push events
- SOPS/KMS encrypted secret support
- Multi-cluster sync
- rollback
- multi env/ tenancy
- Tekton pipeline config integration
- Jsonnet/Kustomize transformations
- pruning
- templating
- helm support