# Tekton Pipeline Troubleshooting Guide

## Common Issues and Solutions

### TaskRunValidationFailed Error

**Symptom**: Pipeline fails to start with `TaskRunValidationFailed` error

**Root Cause**: Incorrect use of `volumeMounts` in task step definitions

**Solution**: Remove `volumeMounts` from task steps. Tekton automatically mounts workspaces based on workspace declarations.

#### ❌ Incorrect Pattern (causes validation error):
```yaml
steps:
  - name: step-name
    volumeMounts:  # Don't do this!
      - name: tools-ws
        mountPath: /tools
    script: |
      # script content
```

#### ✅ Correct Pattern:
```yaml
workspaces:
  - name: tools-ws
    description: Workspace for shared tools

steps:
  - name: step-name
    workingDir: $(workspaces.output.path)
    script: |
      # Access tools workspace using Tekton variable
      GO_DIR="$(workspaces.tools-ws.path)/go"
      export PATH="$GO_DIR/bin:$PATH"
```

### Workspace Not Found Error

**Symptom**: Task fails with "workspace not found" or "path does not exist"

**Solution**: Ensure workspace is:
1. Declared in the task's `workspaces` section
2. Provided in the pipeline's task reference
3. Provisioned in the trigger with appropriate PVC

### Go Command Not Found

**Symptom**: Tasks fail with "go: command not found"

**Solution**: Use shared tools workspace:
1. Install Go once in setup task to `$(workspaces.tools-ws.path)/go`
2. All subsequent tasks access Go from tools-ws
3. Set PATH: `export PATH="$(workspaces.tools-ws.path)/go/bin:$PATH"`

### Workspace Already Exists Error

**Symptom**: Setup task fails with "destination path already exists"

**Solution**: Clean workspace before cloning:
```bash
if [ -d "$(workspaces.output.path)/.git" ]; then
  echo "Cleaning existing workspace..."
  rm -rf $(workspaces.output.path)/*
  rm -rf $(workspaces.output.path)/.[!.]*
fi
```

## Best Practices

### 1. Workspace Management
- Use `pipeline-ws` (1Gi) for source code and artifacts
- Use `tools-ws` (500Mi) for shared tools and dependencies
- Use `secrets` workspace for sensitive data

### 2. Tool Installation
- Install shared tools (Go, Node.js, etc.) once in setup task
- Store in tools-ws workspace for reuse across tasks
- Set PATH in each task to access shared tools

### 3. Error Handling
- Always use `set -eo pipefail` in bash scripts
- Add debug mode support with `PIPELINE_DEBUG` parameter
- Include clear error messages and logging

### 4. Resource Limits
- Set appropriate workspace sizes based on needs
- Monitor PVC usage in IBM Cloud
- Clean up old PVCs regularly

## Debugging Tips

### Enable Debug Mode
Set `pipeline-debug: "1"` in pipeline parameters to enable verbose logging:
```yaml
params:
  - name: pipeline-debug
    value: "1"
```

### Check Workspace Contents
Add debug steps to list workspace contents:
```bash
echo "Workspace contents:"
ls -la $(workspaces.output.path)
echo "Tools workspace:"
ls -la $(workspaces.tools-ws.path)
```

### Verify Environment
Check available commands and environment variables:
```bash
echo "PATH: $PATH"
echo "Go version:"
go version || echo "Go not found"
echo "Available commands:"
which git gh go || true
```

## Fixed Issues

### 2026-04-08: Removed volumeMounts from Task Definitions
- **Files Modified**: 
  - `.tekton/setup-task.yaml`
  - `.tekton/test-task.yaml`
  - `.tekton/ai-test-generation-task.yaml`
- **Issue**: TaskRunValidationFailed error preventing pipeline execution
- **Fix**: Removed incorrect `volumeMounts` declarations from all task steps
- **Result**: Tekton now properly handles workspace mounting automatically

## Additional Resources

- [Tekton Workspaces Documentation](https://tekton.dev/docs/pipelines/workspaces/)
- [IBM Cloud Tekton Pipelines](https://cloud.ibm.com/docs/ContinuousDelivery?topic=ContinuousDelivery-tekton-pipelines)
- [Tekton Task Best Practices](https://tekton.dev/docs/pipelines/tasks/)