# Tekton Workspace Configuration Guide

## Overview

This document explains the workspace configuration for the Intelligent Test Orchestrator pipeline.

## Workspace Architecture

### 1. Pipeline-Level Workspaces (pipeline.yaml)

The pipeline defines three workspaces that are shared across tasks:

```yaml
workspaces:
  - name: pipeline-ws      # Main workspace for source code and artifacts
  - name: tools-ws         # Shared tools and dependencies
  - name: secrets          # API keys and tokens
```

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

  - name: secrets
    secret:
      secretName: pipeline-secrets  # API keys from Kubernetes secret
```

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

### ✅ ai-test-generation-task (Lines 43-49 in ai-test-generation-task.yaml)
```yaml
workspaces:
  - name: output      # Maps to pipeline-ws
  - name: tools-ws    # Access shared Go installation
  - name: secrets     # Access API keys
```

**Purpose:**
- `output`: Access source code, generate tests, commit changes
- `tools-ws`: Use Go installation from setup task
- `secrets`: Read Bob Shell API key and GitHub token

**Key Operations:**
- Uses Go from tools-ws to build test orchestrator
- Reads API keys from secrets workspace
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
      - name: secrets
        workspace: secrets        # Task's 'secrets' → Pipeline's 'secrets'
```

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
API keys and tokens are stored in Kubernetes secrets and mounted as workspace:

```yaml
# Create secret
kubectl create secret generic pipeline-secrets \
  --from-literal=api-key=your-bob-api-key \
  --from-literal=token=your-github-token

# Access in task
env:
  - name: BOBSHELL_API_KEY
    valueFrom:
      secretKeyRef:
        name: bobshell-api-key
        key: api-key
```

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
| secrets | N/A | API keys (from Kubernetes secret) | Persistent |

## Summary

The workspace configuration follows these principles:

1. **Separation of Concerns**: Different workspaces for different purposes
2. **Resource Efficiency**: Shared tools workspace avoids redundant installations
3. **Security**: Secrets isolated in dedicated workspace
4. **Clarity**: Clear naming and mapping between task and pipeline levels
5. **Minimal Coupling**: Tasks only declare workspaces they actually use

This design ensures efficient resource usage, clear data flow, and maintainable pipeline configuration.