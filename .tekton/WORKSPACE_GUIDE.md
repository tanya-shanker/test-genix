# Tekton Workspace Configuration Guide

## Overview

This document explains the workspace configuration for the Intelligent Test Orchestrator pipeline.

## Workspace Architecture

### 1. Pipeline-Level Workspaces (pipeline.yaml)

The pipeline defines two workspaces that are shared across tasks:

```yaml
workspaces:
  - name: pipeline-ws      # Main workspace for source code and artifacts
  - name: tools-ws         # Shared tools and dependencies
```

**Note**: API keys and tokens are provided via Kubernetes `secretKeyRef` in environment variables, not via a workspace.

### 2. Workspace Provisioning (pr-trigger.yaml)

Workspaces are provisioned when a pipeline run is triggered:

```yaml
workspaces:
  - name: pipeline-ws
    volumeClaimTemplate:
      spec:
        accessModes: [ReadWriteOnce]
        resources:
          requests:
            storage: 1Gi      # Source code, test results, coverage reports

  - name: tools-ws
    volumeClaimTemplate:
      spec:
        accessModes: [ReadWriteOnce]
        resources:
          requests:
            storage: 500Mi    # Go installation, dependencies

```

**Note**: Secrets are now provided directly via `secretKeyRef` in task environment variables, eliminating the need for a secrets workspace.

## Workspace Usage by Task

### ✅ setup-task (Lines 23-27 in setup-task.yaml)
```yaml
workspaces:
  - name: output      # Maps to pipeline-ws
  - name: tools-ws    # Shared tools workspace
```

**Purpose:**
- `output`: Clone repository, store source code
- `tools-ws`: Install Go (once) for all subsequent tasks

**Key Operations:**
- Clones repository to `$(workspaces.output.path)`
- Installs Go to `$(workspaces.tools-ws.path)/go`
- Sets up GOPATH at `$(workspaces.tools-ws.path)/gopath`

### ✅ test-task (Lines 14-18 in test-task.yaml)
```yaml
workspaces:
  - name: output      # Maps to pipeline-ws
  - name: tools-ws    # Access shared Go installation
```

**Purpose:**
- `output`: Access source code, run tests, store coverage reports
- `tools-ws`: Use Go installation from setup task

**Key Operations:**
- Runs tests using Go from `$(workspaces.tools-ws.path)/go/bin`
- Generates `coverage-before.out` in output workspace
- Stores baseline coverage in `.coverage-before`

### ✅ ai-test-generation-task (Lines 43-46 in ai-test-generation-task.yaml)
```yaml
workspaces:
  - name: output      # Maps to pipeline-ws
  - name: tools-ws    # Access shared Go installation
```

**Purpose:**
- `output`: Access source code, generate tests, commit changes
- `tools-ws`: Use Go installation from setup task

**Secrets Management:**
API keys are provided via `secretKeyRef` in environment variables:
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

**Key Operations:**
- Uses Go from tools-ws to build test orchestrator
- Reads API keys from Kubernetes secrets via environment variables
- Generates tests and commits to output workspace
- Creates coverage comparison report

### ✅ static-scan-task (Lines 14-16 in static-scan-task.yaml)
```yaml
workspaces:
  - name: output      # Maps to pipeline-ws
```

**Purpose:**
- `output`: Access source code for static analysis

**Key Operations:**
- Runs linters (golangci-lint, eslint, pylint) if available
- Analyzes code in output workspace

### ✅ deploy-task (Lines 14-16 in deploy-task.yaml)
```yaml
workspaces:
  - name: output      # Maps to pipeline-ws
```

**Purpose:**
- `output`: Access built artifacts for deployment

**Key Operations:**
- Deploys application from output workspace
- Placeholder for deployment logic (kubectl, helm, cf, docker)

## Workspace Mapping in Pipeline

The pipeline maps logical workspace names to physical workspaces:

```yaml
tasks:
  - name: setup
    workspaces:
      - name: output
        workspace: pipeline-ws    # Task's 'output' → Pipeline's 'pipeline-ws'
      - name: tools-ws
        workspace: tools-ws       # Task's 'tools-ws' → Pipeline's 'tools-ws'

  - name: ai-test-generation
    workspaces:
      - name: output
        workspace: pipeline-ws
      - name: tools-ws
        workspace: tools-ws
```

**Note**: Secrets are provided via `secretKeyRef` in the task's environment variables, not via workspace mapping.

## Best Practices

### 1. Workspace Naming Convention
- **Task Level**: Use generic names (`output`, `tools-ws`, `secrets`)
- **Pipeline Level**: Use descriptive names (`pipeline-ws`, `tools-ws`, `secrets`)
- **Mapping**: Pipeline maps task names to pipeline names

### 2. Workspace Access Patterns
```bash
# In task scripts, access workspaces using Tekton variables:
WORKSPACE_PATH="$(workspaces.output.path)"
TOOLS_PATH="$(workspaces.tools-ws.path)"
SECRETS_PATH="$(workspaces.secrets.path)"

# Example: Access Go installation
GO_DIR="$(workspaces.tools-ws.path)/go"
export PATH="$GO_DIR/bin:$PATH"
```

### 3. Shared Tools Pattern
The `tools-ws` workspace enables efficient tool sharing:

1. **Setup Task**: Install tools once
   ```bash
   # Install Go to shared workspace
   wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
   tar -C "$(workspaces.tools-ws.path)" -xzf go1.21.0.linux-amd64.tar.gz
   ```

2. **Subsequent Tasks**: Reuse installed tools
   ```bash
   # Use Go from shared workspace
   export PATH="$(workspaces.tools-ws.path)/go/bin:$PATH"
   go version
   ```

### 4. Secrets Management
API keys and tokens are stored in Kubernetes secrets and accessed via `secretKeyRef`:

```yaml
# Create secrets
kubectl create secret generic bobshell-api-key \
  --from-literal=api-key=your-bob-api-key

kubectl create secret generic github-token \
  --from-literal=token=your-github-token

# Access in task environment variables
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

This approach is cleaner than using a secrets workspace and follows Tekton best practices.

## Workspace Lifecycle

1. **Pipeline Trigger**: PVCs created for pipeline-ws and tools-ws
2. **Setup Task**: Clones repo to pipeline-ws, installs Go to tools-ws
3. **Test Task**: Accesses code from pipeline-ws, uses Go from tools-ws
4. **AI Generation**: Accesses all three workspaces
5. **Static Scan**: Accesses code from pipeline-ws
6. **Deploy**: Accesses artifacts from pipeline-ws
7. **Pipeline Complete**: PVCs can be cleaned up (manual or automatic)

## Troubleshooting

### Issue: Workspace Not Found
**Symptom**: Task fails with "workspace not found"

**Solution**: Ensure workspace is:
1. Declared in task's `workspaces` section
2. Provided in pipeline's task reference
3. Provisioned in trigger with PVC or secret

### Issue: Tool Not Found
**Symptom**: "go: command not found" in test or AI generation task

**Solution**: Verify:
1. Setup task completed successfully
2. Go installed to `$(workspaces.tools-ws.path)/go`
3. PATH set correctly: `export PATH="$(workspaces.tools-ws.path)/go/bin:$PATH"`

### Issue: Permission Denied
**Symptom**: Cannot write to workspace

**Solution**: Check:
1. PVC access mode is ReadWriteOnce
2. Task runs with appropriate service account
3. Workspace path exists and is writable

## Storage Requirements

| Workspace | Size | Purpose | Cleanup |
|-----------|------|---------|---------|
| pipeline-ws | 1Gi | Source code, test results, coverage reports | After pipeline run |
| tools-ws | 500Mi | Go installation, dependencies | After pipeline run |

**Note**: Secrets are managed via Kubernetes `secretKeyRef`, not via workspaces.

## Summary

The workspace configuration follows these principles:

1. **Separation of Concerns**: Different workspaces for different purposes
2. **Resource Efficiency**: Shared tools workspace avoids redundant installations
3. **Security**: Secrets managed via Kubernetes secretKeyRef (not workspaces)
4. **Clarity**: Clear naming and mapping between task and pipeline levels
5. **Minimal Coupling**: Tasks only declare workspaces they actually use
6. **Best Practices**: Following Tekton recommendations for secret management

This design ensures efficient resource usage, clear data flow, and maintainable pipeline configuration.