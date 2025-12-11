# TODO — Config Synchronizer Operator

**Updated**: December 11, 2025
**Status**: MVP Core Complete - Testing & Enhancement Phase

## ✅ COMPLETED ITEMS

1. ✅ **Initialize Kubebuilder Project** - Base project structure is in place
2. ✅ **Define CRD Types** - Full ConfigSync API implemented with proper validation
3. ✅ **Scaffold Controller** - Controller is wired and functional
4. ✅ **Implement Source Fetchers** - Git source fetching with SSH/HTTPS auth complete
5. ✅ **Basic Validation** - YAML parsing and basic validation implemented
6. ✅ **Apply Targets** - Manifest application with server-side apply working
7. ✅ **Reconcile Loop & Watches** - Full reconciliation with change detection via Git SHA
8. ✅ **Basic Status & Conditions** - Status tracking with Degraded condition implemented  
9. ✅ **RBAC & Manifests** - Proper permissions and deployment manifests generated

## 🚧 HIGH PRIORITY - NEXT STEPS

### 11. **Testing Infrastructure** (CRITICAL)
- **Issue**: Tests fail due to missing envtest binaries
- **Actions**:
  - Install envtest binaries: `make envtest` 
  - Fix test setup in `suite_test.go`
  - Add unit tests for Git fetching, manifest parsing, apply logic
  - Add integration tests with real Git repositories
- **Files**: `internal/controller/suite_test.go`, `internal/sources/*_test.go`, `internal/apply/*_test.go`

### 12. **Templating System** (HIGH)
- **Goal**: Support Go templates for dynamic configuration
- **Actions**:
  - Create `internal/template/` package
  - Implement `renderTemplate(data, templateStr) -> renderedData`
  - Add template parsing and validation
  - Update controller to use templating before applying manifests
- **Files**: `internal/template/render.go`, update `configsync_controller.go`

### 13. **Enhanced Error Handling** (MEDIUM)
- **Actions**:
  - Fix `setCondition` to only update `LastTransitionTime` when status changes
  - Add Kubernetes event emission for better observability  
  - Implement retry with exponential backoff
  - Add more specific condition types (Available, Progressing)
- **Files**: `configsync_controller.go`

## 📋 MEDIUM PRIORITY

### 14. **Validation Enhancements**
- JSON schema validation for manifests
- Pre-apply validation checks
- Template syntax validation

### 15. **Rollback Support**
- Add rollback spec field to ConfigSync
- Implement revert to previous/specific Git SHA
- Track deployment history

### 16. **Multi-Environment Support** 
- Branch-specific configurations
- Environment-based templating
- Multi-tenancy considerations
## 📋 LOW PRIORITY / STRETCH GOALS

### 17. **Pruning & Garbage Collection**
- Track applied resources and clean up orphaned ones
- Add finalizers for proper cleanup
- Implement resource ownership tracking

### 18. **Advanced Features**
- Webhook triggers for Git push events
- SOPS/KMS encrypted secret support  
- Multi-cluster sync capabilities
- Helm chart support
- Kustomize integration

### 19. **CI/CD & Documentation**
- GitHub Actions for automated testing
- Comprehensive documentation with examples
- Demo videos and tutorials

---

## 🚨 CURRENT BLOCKERS & ISSUES

1. **Testing Environment**: envtest binaries missing - prevents running `make test`
2. **Condition Logic**: `setCondition` updates `LastTransitionTime` on every call (causes status churn)
3. **Docker Permissions**: Docker group membership requires re-login to take effect

## 🎯 IMMEDIATE NEXT ACTIONS

1. **Fix Testing** (30 min):
   ```bash
   make envtest  # Install test binaries
   go test ./internal/controller/...  # Verify tests pass
   ```

2. **Implement Basic Templating** (2-3 hours):
   - Create `internal/template/render.go` 
   - Add Go text/template support
   - Update controller to use templating

3. **Add Unit Tests** (1-2 hours):
   - Git fetcher tests with fixtures
   - Manifest application tests  
   - Template rendering tests

**Estimated MVP-to-Production**: ~1-2 weeks of focused development


<!-- idea: rollback support -->
<!-- multi-branch / env support -->
<!-- go back and fix kubebuilder validation for branch, revision and add branch to sync -->
<!-- //right now assuming https. Will need to add functionality for ssh later. Will need to make/reade secret for auth
Add full logging + error types + conditions updates

Generate unit tests for Git logic

Add compare-SHA logic in your Reconcile loop

Add server-side apply code for applying manifests -->

<!-- rbac
 -->
KNOWN HOST SUPPORT
gitignore for sensitive stuff when testing
- first probably raw manifest apply support - then include helm support.

richer error/fmt handling in fetch.go

Gaps & Risks (highest impact first)

Apply-loop incomplete: Without a robust apply_target, operator won't create/patch target ConfigMap/Secret as intended. Files: apply.go, reconciler apply loop.
Templating & validation missing: No template rendering or config validation; dangerous to apply raw source directly.
Condition handling bug: setCondition currently updates LastTransitionTime on every call — leads to noisy status churn. Needs logic to set LastTransitionTime only when the condition Status changes.
RBAC verification: Generated RBAC exists but verify it allows get/list/watch for Secrets/ConfigMaps and create/patch for applied resources. Also ensure the controller ServiceAccount is assigned required roles.
Tests coverage: No unit tests for fetchers/apply/templating. envtest/e2e not wired to validate full behavior.
Temp-dir management: Fetchers write to temp dirs; reconciler must defer os.RemoveAll or use an in-memory approach to avoid leaks.
Secrets handling: SSH/HTTPS credentials read from Secrets — ensure permissions and secure filesystem writes (mode 0600) are enforced (current code uses restrictive perms for secret files but confirm across all places).
Manifests generation: make manifests was previously flaky during earlier iterations — re-run to ensure CRD schemas reflect current api types.
Concrete next actions (prioritized)

Implement apply loop and target applier (highest priority)
Files to add/change: apply.go (or extend apply.go), update reconciler loop in configsync_controller.go to call applyTarget(ctx, r.Client, target, renderedData).
Acceptance: applying a sample ConfigSync causes target ConfigMap/Secret to be created/updated in cluster (can test on Kind).
Add templating and validation
Files: internal/template/render.go (use Go text/template or sprig functions), internal/validate/validate.go (YAML/JSON schema check).
Acceptance: renderer applies templates per Spec.Targets with test cases.
Fix setCondition behavior
Change: compute existing condition; only update LastTransitionTime when Status changes.
Files: modify configsync_controller.go setCondition.
RBAC & manifests verification
Run: make manifests, make generate, inspect config/rbac/* and config/crd/bases/*.
Ensure permissions for Secrets/ConfigMaps and server-side apply/patch are present.
Add unit tests for fetchers and apply
Files: internal/sources/*_test.go, internal/apply/*_test.go.
Use table tests and isolated temp dirs; for Git fetcher, use local git repo fixtures or go-git in-memory repos.
CI skeleton
Add GitHub Actions to run go test, go vet, make manifests on PRs.
Commands to run locally to validate current state

Build all packages: go build [repos](http://_vscodecontentref_/29).
Generate manifests (after code changes): make manifests
Run unit tests: go test [repos](http://_vscodecontentref_/30).
Build dev image: make dev-image (faster iteration)