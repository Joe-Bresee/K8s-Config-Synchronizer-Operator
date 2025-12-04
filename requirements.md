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

## CRD: ConfigSync
- TODO: I want a more loosely defined CR - right now I need to define a CR target with type deployment if my gitsource has a deployment type - which seems a little tight and unneeded. I am going to spend some time designing the configsync CR a bit more - also nede to think about multiple resource deployments, maybe multi-repo/tenancy/env.

### spec
- `source`
  - `git`:
    - `repo` (string) — HTTPS/SSH URL
    - `path` (string) — path to file in repo
    - `revision` (string, optional, default `main`)
-- `targets` (list)
- `namespace`
- `type` `deployment`, etc
- `name`
- `key` (optional, for Secret)
- `refreshInterval` (string, optional)

### status
- `lastSyncedTime`
- `sourceRevision` (e.g., Git SHA)
- `appliedTargets` (int)
- `conditions` (list of conditions: Synced, Failed, InvalidSource)

---

## Operator Behavior

### Source Fetching
- Fetches Git source defined in ConfigSync CR and if remote has a more recent hash, re-syncs repository code with local temp repo.
- Caches difference for efficiency / in case there's no updates.

### Templating / Transformation
- TODO: `render_template(data: dict, target: dict) -> dict`
- Apply simple variable interpolation
- Support Jinja2 (Python) or Go templates
- Optional: allow per-target overrides

### Target Application
- Applies manifest changes to target defined in the ConfigSync CR, returns Errors if any occur - reports success

### Reconciliation Triggers
- TODO: On CR create/update/delete
- On refresh interval

### Error Handling
- Log errors with structured logging
- Update `.status.conditions` for Synced, Failed, InvalidSource
- TODO: Emit Kubernetes events
- TODO: Retry with exponential backoff

### TODO: Rollbacks
- implement rollback functionality to most previous/specific commit SHA.

### TODO: pruning & garbage collection

### TODO: determine apply semantics
- is this a server-side apply only, or client side?

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