# Workspace Configuration Fix Summary

## Date: 2026-04-08 (Updated)

## Problem
The Tekton pipeline was experiencing `TaskRunValidationFailed` errors preventing pipeline execution.

## Root Causes Identified

### 1. Incorrect volumeMounts Usage (Previously Fixed)
- Tasks were using `volumeMounts` in step definitions
- Tekton automatically mounts workspaces - manual volumeMounts cause validation errors
- **Status**: ✅ Fixed in previous commits

### 2. Inconsistent Workspace Declarations
- `static-scan-task` and `deploy-task` declared `tools-ws` workspace but never used it
- Pipeline didn't provide `tools-ws` to these tasks (lines 118-140 in pipeline.yaml)
- This created confusion and potential for future errors
- **Status**: ✅ Fixed in this commit

## Changes Made (Part 1 - Initial Fix)

### 1. Cleaned Up static-scan-task.yaml
**Before:**
```yaml
workspaces:
  - name: output
    description: Workspace for pipeline artifacts
  - name: tools-ws
    description: Workspace for shared tools and dependencies
```

**After:**
```yaml
workspaces:
  - name: output
    description: Workspace for pipeline artifacts
```

**Reason**: Static scan task only needs access to source code, not shared tools.

### 2. Cleaned Up deploy-task.yaml
**Before:**
```yaml
workspaces:
  - name: output
    description: Workspace for pipeline artifacts
  - name: tools-ws
    description: Workspace for shared tools and dependencies
```

**After:**
```yaml
workspaces:
  - name: output
    description: Workspace for pipeline artifacts
```

**Reason**: Deploy task only needs access to artifacts, not shared tools.

### 3. Added Documentation
Created two comprehensive guides:

#### TROUBLESHOOTING.md (135 lines)
- Common Tekton issues and solutions
- TaskRunValidationFailed error explanation
- Workspace management best practices
- Debugging tips
- Fixed issues log

#### WORKSPACE_GUIDE.md (267 lines)
- Complete workspace architecture explanation
- Workspace usage by each task
- Workspace mapping in pipeline
- Best practices and patterns
- Troubleshooting guide
- Storage requirements table

## Changes Made (Part 2 - Secrets Workspace Removal)

### Issue: Pod Initialization Failure
**Error Message:**
```
pod status "Initialized":"False"
message: "containers with incomplete status: [prepare place-scripts working-dir-initializer]"
```

**Root Cause:**
The `ai-test-generation-task` was using BOTH methods for secrets:
1. Secrets workspace declaration (line 48-49)
2. secretKeyRef in environment variables (lines 70-81)

This created a conflict during pod initialization because:
- The task declared a `secrets` workspace
- The pipeline provided the workspace
- But the task was actually using `secretKeyRef` to access secrets
- Tekton's init containers got confused trying to mount both

**Solution:** Remove secrets workspace entirely and use only `secretKeyRef` (Tekton best practice)

### 3. Removed Secrets Workspace from ai-test-generation-task.yaml
**Before:**
```yaml
workspaces:
  - name: output
  - name: tools-ws
  - name: secrets    # ❌ Conflicted with secretKeyRef
```

**After:**
```yaml
workspaces:
  - name: output
  - name: tools-ws
```

Secrets are now accessed only via `secretKeyRef`:
```yaml
env:
  - name: BOBSHELL_API_KEY
    valueFrom:
      secretKeyRef:
        name: bobshell-api-key
        key: api-key
        optional: true
```

### 4. Updated pipeline.yaml
Removed `secrets` workspace from:
- Pipeline-level workspace declarations (lines 45-51)
- ai-test-generation task workspace mapping (lines 110-116)

### 5. Updated pr-trigger.yaml
Removed secrets workspace provisioning (lines 117-119)

### 6. Updated Documentation
Updated WORKSPACE_GUIDE.md to reflect:
- Only 2 workspaces (pipeline-ws, tools-ws)
- Secrets managed via secretKeyRef
- Best practices for Tekton secret management

## Final Workspace Configuration (Updated)

### Pipeline Level (2 workspaces)
```yaml
workspaces:
  - name: pipeline-ws    # 1Gi - Source code and artifacts
  - name: tools-ws       # 500Mi - Shared tools (Go)
```

### Secrets Management
Secrets are provided via Kubernetes `secretKeyRef` in environment variables:
```yaml
env:
  - name: BOBSHELL_API_KEY
    valueFrom:
      secretKeyRef:
        name: bobshell-api-key
        key: api-key
        optional: true
  - name: GITHUB_TOKEN
    valueFrom:
      secretKeyRef:
        name: github-token
        key: token
        optional: true
```

### Task Level Usage

| Task | Workspaces Used | Secrets Method | Purpose |
|------|----------------|----------------|---------|
| setup-task | output, tools-ws | N/A | Clone repo, install Go |
| test-task | output, tools-ws | N/A | Run tests using Go |
| ai-test-generation-task | output, tools-ws | secretKeyRef | Generate tests with Bob API |
| static-scan-task | output | N/A | Run linters on code |
| deploy-task | output | Deploy artifacts |

## Benefits of This Fix

### 1. Clarity
- Each task only declares workspaces it actually uses
- No confusion about which workspaces are needed
- Clear documentation of workspace purpose

### 2. Consistency
- Workspace declarations match actual usage
- Pipeline configuration aligns with task requirements
- No unused workspace declarations

### 3. Maintainability
- Easy to understand workspace flow
- Clear documentation for troubleshooting
- Patterns documented for future tasks

### 4. Efficiency
- Shared tools workspace (tools-ws) avoids redundant installations
- Go installed once in setup, used by test and AI generation tasks
- Optimal storage allocation (1Gi for code, 500Mi for tools)
- Secrets via secretKeyRef avoid unnecessary workspace mounts

### 5. Best Practices
- Following Tekton recommendations for secret management
- Using `secretKeyRef` instead of secrets workspace
- Marking secrets as `optional: true` for graceful degradation

## Verification Steps

To verify the fix works:

1. **Check Task Definitions**
   ```bash
   # Verify no volumeMounts in any task
   grep -r "volumeMounts" .tekton/*.yaml
   # Should return no results
   ```

2. **Verify Workspace Consistency**
   ```bash
   # Check each task declares only needed workspaces
   for file in .tekton/*-task.yaml; do
     echo "=== $file ==="
     grep -A 5 "workspaces:" "$file"
   done
   ```

3. **Test Pipeline Execution**
   - Create a test PR
   - Verify pipeline starts successfully
   - Check all tasks complete without workspace errors

## Related Files

- `.tekton/pipeline.yaml` - Pipeline workspace definitions
- `.tekton/pr-trigger.yaml` - Workspace provisioning
- `.tekton/setup-task.yaml` - Installs Go to tools-ws
- `.tekton/test-task.yaml` - Uses Go from tools-ws
- `.tekton/ai-test-generation-task.yaml` - Uses Go and secrets
- `.tekton/static-scan-task.yaml` - Uses output only (fixed)
- `.tekton/deploy-task.yaml` - Uses output only (fixed)

## Next Steps

1. Test the pipeline with a real PR
2. Monitor for any workspace-related errors
3. Update documentation if new patterns emerge
4. Consider adding workspace usage validation in CI

## Commit Information

### Part 1: Initial Workspace Cleanup
**Commit**: ddbcf54
**Message**: fix: Clean up Tekton workspace configuration and add documentation
**Files Changed**: 4 files, 393 insertions(+), 4 deletions(-)
**New Files**:
- `.tekton/TROUBLESHOOTING.md`
- `.tekton/WORKSPACE_GUIDE.md`

### Part 2: Secrets Workspace Removal
**Commit**: [Pending]
**Message**: fix: Remove secrets workspace and use secretKeyRef for API keys
**Files Changed**: 4 files
**Modified Files**:
- `.tekton/ai-test-generation-task.yaml` - Removed secrets workspace
- `.tekton/pipeline.yaml` - Removed secrets workspace declaration
- `.tekton/pr-trigger.yaml` - Removed secrets workspace provisioning
- `.tekton/WORKSPACE_GUIDE.md` - Updated documentation

## References

- [Tekton Workspaces Documentation](https://tekton.dev/docs/pipelines/workspaces/)
- [IBM Cloud Tekton Pipelines](https://cloud.ibm.com/docs/ContinuousDelivery?topic=ContinuousDelivery-tekton-pipelines)
- Project: [test-genix](https://github.com/tanya-shanker/test-genix)