# IBM Cloud Tekton OnePipeline Configuration

This directory contains the Tekton pipeline configuration for the Intelligent Test Orchestrator.

## 📁 Directory Structure

```
.tekton/
├── README.md                           # This file
├── pipeline.yaml                       # Main pipeline definition
├── triggers/
│   └── pr-trigger.yaml                # PR trigger configuration
└── tasks/
    ├── setup-task.yaml                # Setup and clone repository
    ├── test-task.yaml                 # Run existing tests
    ├── ai-test-generation-task.yaml   # AI-powered test generation
    ├── static-scan-task.yaml          # Static code analysis
    └── deploy-task.yaml               # Deployment task
```

## 🚀 Quick Setup

### 1. Prerequisites

- IBM Cloud account with OnePipeline enabled
- GitHub or GitHub Enterprise repository
- Bob Shell API key (from https://internal.bob.ibm.com)
- GitHub personal access token (for PR operations)

### 2. Create Secrets in IBM Cloud

#### Option A: Using IBM Cloud Secrets Manager

```bash
# Create Bob Shell API key secret
ibmcloud secrets-manager secret-create \
  --secret-type arbitrary \
  --name bobshell-api-key \
  --secret-data '{"api-key":"your-bob-api-key-here"}'

# Create GitHub token secret
ibmcloud secrets-manager secret-create \
  --secret-type arbitrary \
  --name github-token \
  --secret-data '{"token":"your-github-token-here"}'
```

#### Option B: Using Kubernetes Secrets

```bash
# Create Bob Shell API key secret
kubectl create secret generic bobshell-api-key \
  --from-literal=api-key='your-bob-api-key-here' \
  -n tekton-pipelines

# Create GitHub token secret
kubectl create secret generic github-token \
  --from-literal=token='your-github-token-here' \
  -n tekton-pipelines
```

### 3. Configure Pipeline in IBM Cloud UI

1. Navigate to your IBM Cloud Toolchain
2. Add a new "Delivery Pipeline" tool
3. Select "Tekton" as the pipeline type
4. Point to this repository
5. Set the pipeline definition path: `.tekton/pipeline.yaml`

### 4. Configure Environment Variables

In the IBM Cloud OnePipeline UI, configure these environment variables:

#### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `BOBSHELL_API_KEY` | Bob Shell API key | `sk-ant-...` |
| `GIT_BRANCH` | Current branch | Auto-set by pipeline |
| `GIT_COMMIT` | Current commit SHA | Auto-set by pipeline |

#### Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `BASE_BRANCH` | Base branch for comparison | `main` |
| `COVERAGE_TARGET` | Target coverage percentage | `80` |
| `FUNCTIONAL_TEST_REPO` | Functional test repository | `` |
| `GITHUB_TOKEN` | GitHub personal access token | `` |
| `GITHUB_ENTERPRISE_URL` | GitHub Enterprise URL | `` |
| `PIPELINE_DEBUG` | Enable debug mode | `0` |

### 5. Configure Triggers

The pipeline can be triggered by:

- **Pull Request Events**: Automatically runs on PR creation/update
- **Push Events**: Runs on push to specific branches
- **Manual Trigger**: Can be triggered manually from UI

See [`.tekton/triggers/pr-trigger.yaml`](.tekton/triggers/pr-trigger.yaml) for trigger configuration.

## 📊 Pipeline Stages

### Stage 1: Setup
- Clones the repository
- Installs dependencies (Go, jq, curl, etc.)
- Downloads Go modules

### Stage 2: Test
- Runs existing test suite
- Measures baseline code coverage
- Supports Go, Node.js, Python, Java

### Stage 3: AI Test Generation ⭐
- Detects code changes in PR
- Generates unit tests using Bob Shell
- Generates functional tests
- Posts coverage summary to PR
- Creates functional test PR (if applicable)
- **Non-blocking**: Pipeline continues even if this fails

### Stage 4: Static Scan
- Runs static code analysis
- Supports golangci-lint, ESLint, Pylint

### Stage 5: Deploy
- Placeholder for deployment logic
- Customize based on your deployment target

## 🔧 Customization

### Modify Pipeline Stages

Edit [`.tekton/pipeline.yaml`](pipeline.yaml) to:
- Add new stages
- Modify stage order
- Change stage dependencies
- Add custom parameters

### Modify Tasks

Edit task files in [`.tekton/tasks/`](tasks/) to:
- Change container images
- Modify scripts
- Add new steps
- Configure resource limits

### Add Triggers

Create new trigger files in [`.tekton/triggers/`](triggers/) for:
- Different event types
- Branch-specific triggers
- Scheduled runs

## 📝 Example: Adding a New Stage

```yaml
# Add to .tekton/pipeline.yaml
tasks:
  - name: security-scan
    runAfter:
      - static-scan
    taskRef:
      name: security-scan-task
    workspaces:
      - name: output
        workspace: pipeline-ws
```

Then create `.tekton/tasks/security-scan-task.yaml`:

```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: security-scan-task
spec:
  workspaces:
    - name: output
  steps:
    - name: scan
      image: aquasec/trivy:latest
      script: |
        #!/bin/sh
        trivy fs --severity HIGH,CRITICAL .
```

## 🐛 Troubleshooting

### Pipeline Fails at Setup Stage

**Problem**: Cannot clone repository

**Solution**: Check repository URL and access permissions

```bash
# Verify repository access
git clone <repository-url>
```

### AI Test Generation Fails

**Problem**: Bob Shell API key not found

**Solution**: Verify secret is created and accessible

```bash
# Check secret exists
kubectl get secret bobshell-api-key -n tekton-pipelines

# Verify secret content
kubectl get secret bobshell-api-key -n tekton-pipelines -o yaml
```

### Tests Not Running

**Problem**: Test framework not detected

**Solution**: Ensure your project has proper test configuration files:
- Go: `go.mod`
- Node.js: `package.json`
- Python: `requirements.txt` or `setup.py`
- Java: `pom.xml`

### Enable Debug Mode

Set `PIPELINE_DEBUG=1` in environment variables to see detailed logs:

```bash
# In IBM Cloud UI, add environment variable:
PIPELINE_DEBUG=1
```

## 📚 Additional Resources

- [IBM Cloud OnePipeline Documentation](https://cloud.ibm.com/docs/devsecops)
- [Tekton Documentation](https://tekton.dev/docs/)
- [Bob Shell Documentation](https://internal.bob.ibm.com/docs)
- [Project README](../README.md)
- [OnePipeline Integration Guide](../ONEPIPELINE_INTEGRATION.md)

## 🆘 Support

For issues or questions:
1. Check the [troubleshooting section](#-troubleshooting)
2. Review pipeline logs in IBM Cloud UI
3. Check [END_TO_END_FLOW.md](../END_TO_END_FLOW.md) for workflow details
4. Open an issue in the repository

## 📄 License

See [LICENSE](../LICENSE) file in the root directory.