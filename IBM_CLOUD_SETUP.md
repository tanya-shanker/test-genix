# IBM Cloud Tekton OnePipeline Setup Guide

Complete guide for setting up the Intelligent Test Orchestrator on IBM Cloud OnePipeline.

## 📋 Table of Contents

1. [Prerequisites](#prerequisites)
2. [Step-by-Step Setup](#step-by-step-setup)
3. [Configuration](#configuration)
4. [Testing the Pipeline](#testing-the-pipeline)
5. [Troubleshooting](#troubleshooting)
6. [Advanced Configuration](#advanced-configuration)

---

## Prerequisites

### Required

- ✅ IBM Cloud account with OnePipeline/DevSecOps enabled
- ✅ GitHub or GitHub Enterprise repository
- ✅ Bob Shell API key (obtain from https://internal.bob.ibm.com)
- ✅ GitHub personal access token with `repo` and `write:discussion` scopes

### Optional

- IBM Cloud Secrets Manager or Key Protect instance
- Functional test repository (for functional test PR creation)

---

## Step-by-Step Setup

### Step 1: Create IBM Cloud Toolchain

1. **Navigate to IBM Cloud Console**
   ```
   https://cloud.ibm.com/devops/toolchains
   ```

2. **Create New Toolchain**
   - Click "Create toolchain"
   - Select "Build your own toolchain"
   - Name: `intelligent-test-orchestrator`
   - Region: Select your preferred region
   - Resource group: Select your resource group

3. **Add GitHub Integration**
   - Click "Add tool" → "GitHub"
   - Select "Existing" repository
   - Repository URL: Your repository URL
   - Enable "Track deployment of code changes"

### Step 2: Configure Secrets

#### Option A: Using IBM Cloud Secrets Manager (Recommended)

```bash
# Install IBM Cloud CLI if not already installed
curl -fsSL https://clis.cloud.ibm.com/install/linux | sh

# Login to IBM Cloud
ibmcloud login --sso

# Target your resource group
ibmcloud target -g <your-resource-group>

# Create Secrets Manager instance (if not exists)
ibmcloud resource service-instance-create my-secrets-manager \
  secrets-manager standard us-south

# Create Bob Shell API key secret
ibmcloud secrets-manager secret-create \
  --secret-type arbitrary \
  --name bobshell-api-key \
  --description "Bob Shell API key for AI test generation" \
  --secret-data '{"api-key":"your-bob-api-key-here"}'

# Create GitHub token secret
ibmcloud secrets-manager secret-create \
  --secret-type arbitrary \
  --name github-token \
  --description "GitHub token for PR operations" \
  --secret-data '{"token":"ghp_your_github_token_here"}'
```

#### Option B: Using Environment Variables (Quick Start)

In the IBM Cloud OnePipeline UI:

1. Navigate to your pipeline
2. Go to "Environment properties"
3. Add these secure properties:
   - `BOBSHELL_API_KEY`: Your Bob Shell API key
   - `GITHUB_TOKEN`: Your GitHub personal access token

### Step 3: Add Delivery Pipeline

1. **In your Toolchain, click "Add tool"**
   - Select "Delivery Pipeline"
   - Name: `intelligent-test-orchestrator-pipeline`
   - Pipeline type: **Tekton**

2. **Configure Pipeline**
   - Pipeline definition: **Repository**
   - Repository: Select your GitHub repository
   - Branch: `main`
   - Path: `.one-pipeline.yaml`

3. **Configure Triggers**
   - Add trigger: **Git Repository**
   - Event: `Pull Request`
   - Branch: `.*` (all branches)
   - Enable: ✅ "When a pull request is opened or updated"

### Step 4: Configure Environment Variables

In the Pipeline settings, add these environment variables:

#### Required Variables

| Variable | Value | Type | Description |
|----------|-------|------|-------------|
| `BOBSHELL_API_KEY` | `your-api-key` | Secure | Bob Shell API key |
| `GITHUB_TOKEN` | `your-token` | Secure | GitHub personal access token |

#### Optional Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `BASE_BRANCH` | `main` | Base branch for comparison |
| `COVERAGE_TARGET` | `80` | Target coverage percentage |
| `FUNCTIONAL_TEST_REPO` | `` | Functional test repo (org/repo) |
| `FUNCTIONAL_TEST_REPO_URL` | `` | Full URL to functional test repo |
| `GITHUB_ENTERPRISE_URL` | `` | GitHub Enterprise URL if applicable |
| `PIPELINE_DEBUG` | `0` | Enable debug mode (0 or 1) |

### Step 5: Configure Workspace

1. **In Pipeline settings, configure workspace**
   - Workspace type: **Persistent Volume Claim**
   - Storage class: `ibmc-file-gold` (or your preferred storage class)
   - Size: `1Gi` (minimum)

2. **Configure Service Account**
   - Service account: `default` (or create custom SA)
   - Ensure SA has permissions to:
     - Create PipelineRuns
     - Access secrets
     - Create PVCs

### Step 6: Test the Setup

1. **Create a test PR**
   ```bash
   git checkout -b test/ai-test-generation
   echo "// Test change" >> test_file.go
   git add test_file.go
   git commit -m "test: trigger AI test generation"
   git push origin test/ai-test-generation
   ```

2. **Create Pull Request on GitHub**
   - The pipeline should trigger automatically
   - Monitor progress in IBM Cloud Pipeline UI

3. **Verify Pipeline Execution**
   - Check each stage completes successfully
   - Verify AI test generation stage runs
   - Check for PR comment with coverage summary
   - Verify functional test PR is created (if applicable)

---

## Configuration

### Pipeline Configuration Files

```
.
├── .one-pipeline.yaml              # Main pipeline configuration
├── .tekton/
│   ├── README.md                   # Tekton-specific documentation
│   ├── pipeline.yaml               # Tekton pipeline definition
│   ├── triggers/
│   │   └── pr-trigger.yaml        # PR trigger configuration
│   └── tasks/
│       ├── setup-task.yaml        # Setup task
│       ├── test-task.yaml         # Test task
│       ├── ai-test-generation-task.yaml  # AI test generation
│       ├── static-scan-task.yaml  # Static analysis
│       └── deploy-task.yaml       # Deployment
└── scripts/
    └── ai-test-generator.sh       # AI test generator script
```

### Customizing the Pipeline

#### Modify `.one-pipeline.yaml`

```yaml
# Add custom stage
custom-stage:
  abort_on_failure: false
  image: icr.io/continuous-delivery/pipeline/pipeline-base-image:2.71
  script: |
    #!/usr/bin/env bash
    echo "Running custom stage..."
    # Your custom logic here
```

#### Modify Test Generation Behavior

Edit `scripts/ai-test-generator.sh`:

```bash
# Change coverage target
COVERAGE_TARGET="${COVERAGE_TARGET:-85}"  # Default to 85%

# Change test generation prompt
PROMPT="Generate comprehensive unit tests with 100% coverage..."
```

---

## Testing the Pipeline

### Manual Pipeline Run

1. **Via IBM Cloud UI**
   - Navigate to your pipeline
   - Click "Run Pipeline"
   - Select branch
   - Click "Run"

2. **Via CLI**
   ```bash
   # Install tkn CLI
   brew install tektoncd-cli  # macOS
   # or
   curl -LO https://github.com/tektoncd/cli/releases/download/v0.32.0/tkn_0.32.0_Linux_x86_64.tar.gz
   tar xvzf tkn_0.32.0_Linux_x86_64.tar.gz -C /usr/local/bin/ tkn
   
   # Run pipeline
   tkn pipeline start intelligent-test-orchestrator-pipeline \
     --param repository=https://github.com/your-org/your-repo \
     --param branch=main \
     --workspace name=pipeline-ws,claimName=pipeline-pvc
   ```

### Verify Pipeline Output

Expected output in pipeline logs:

```
==================================================
🤖 AI Test Generation with Bob Shell
==================================================

📊 PR Information:
   Repository: your-org/your-repo
   Branch: feature/new-feature
   Commit: abc123def

✅ API key found - proceeding with test generation
🚀 Starting AI-powered test generation...

📝 Detected changes in 2 file(s):
   • src/api/handler.go
   • src/service/processor.go

🤖 Generating unit tests with Bob Shell...

✅ Generated unit tests: src/api/handler_test.go
✅ Generated unit tests: src/service/processor_test.go

==================================================
📈 Test Coverage Summary
==================================================

Coverage Before:  65%
Coverage After:   82%
Coverage Delta:   +17%

==================================================
💬 Posting Coverage Summary to PR
==================================================

✅ Coverage summary comment posted

==================================================
🔗 Creating Functional Test PR
==================================================

✅ Functional test PR created: #456
🔗 Link added to source PR #123

==================================================
✅ AI Test Generation Complete
==================================================
```

---

## Troubleshooting

### Issue: Pipeline Not Triggering

**Symptoms**: PR created but pipeline doesn't start

**Solutions**:

1. **Check Trigger Configuration**
   ```bash
   # Verify trigger exists
   tkn trigger list
   
   # Check event listener
   kubectl get eventlistener -n tekton-pipelines
   ```

2. **Verify Webhook**
   - Go to GitHub repository settings
   - Navigate to "Webhooks"
   - Check webhook URL matches event listener
   - Verify recent deliveries show success

3. **Check Service Account Permissions**
   ```bash
   kubectl get sa tekton-triggers-sa -n tekton-pipelines
   kubectl describe rolebinding tekton-triggers-rolebinding -n tekton-pipelines
   ```

### Issue: Bob Shell API Key Not Found

**Symptoms**: Error message "Bob Shell API key not found"

**Solutions**:

1. **Verify Secret Exists**
   ```bash
   # For Kubernetes secrets
   kubectl get secret bobshell-api-key -n tekton-pipelines
   
   # For Secrets Manager
   ibmcloud secrets-manager secret bobshell-api-key
   ```

2. **Check Secret Format**
   ```bash
   # Secret should have 'api-key' field
   kubectl get secret bobshell-api-key -n tekton-pipelines -o yaml
   ```

3. **Verify Environment Variable**
   - Check pipeline environment properties
   - Ensure `BOBSHELL_API_KEY` is set as secure property

### Issue: Tests Not Generated

**Symptoms**: Pipeline completes but no tests generated

**Solutions**:

1. **Check File Changes**
   ```bash
   # Verify files were detected
   git diff origin/main...HEAD --name-only
   ```

2. **Enable Debug Mode**
   - Set `PIPELINE_DEBUG=1` in environment variables
   - Re-run pipeline
   - Check detailed logs

3. **Verify Bob Shell Installation**
   ```bash
   # In pipeline logs, look for:
   ✅ Bob Shell installed successfully
   Bob Shell version: X.X.X
   ```

### Issue: Functional Test PR Not Created

**Symptoms**: Unit tests generated but functional test PR fails

**Solutions**:

1. **Check Repository Configuration**
   - Verify `FUNCTIONAL_TEST_REPO` is set correctly
   - Format: `org/repo-name`
   - Ensure repository exists and is accessible

2. **Verify GitHub Token Permissions**
   - Token needs `repo` scope
   - Token needs `write:discussion` scope for PR comments

3. **Check Repository Access**
   ```bash
   # Test GitHub token
   curl -H "Authorization: token $GITHUB_TOKEN" \
     https://api.github.com/repos/org/functional-tests
   ```

### Issue: Coverage Not Calculated

**Symptoms**: Coverage shows as 0% or N/A

**Solutions**:

1. **Verify Test Framework**
   - Go: Ensure `go.mod` exists
   - Node.js: Ensure `package.json` with test script
   - Python: Ensure `pytest` or `coverage` installed

2. **Check Test Execution**
   ```bash
   # Manually run tests
   go test ./... -coverprofile=coverage.out
   go tool cover -func=coverage.out
   ```

3. **Verify Coverage File Generation**
   - Check for `coverage-before.out` in logs
   - Check for `coverage-after.out` in logs

---

## Advanced Configuration

### Custom Docker Images

Use custom images for specific stages:

```yaml
# .one-pipeline.yaml
ai-test-generation:
  image: your-registry/custom-image:tag
  script: |
    #!/usr/bin/env bash
    # Your custom logic
```

### Parallel Test Execution

Modify pipeline to run tests in parallel:

```yaml
# .tekton/pipeline.yaml
tasks:
  - name: unit-tests
    taskRef:
      name: test-task
  - name: integration-tests
    taskRef:
      name: integration-test-task
  # Both run in parallel
```

### Custom Coverage Thresholds

Set different thresholds per project:

```bash
# In environment variables
COVERAGE_TARGET=90  # For critical services
COVERAGE_TARGET=70  # For less critical services
```

### Multi-Repository Support

Configure for multiple functional test repositories:

```bash
# Environment variables
FUNCTIONAL_TEST_REPO_1=org/api-tests
FUNCTIONAL_TEST_REPO_2=org/e2e-tests
```

### Scheduled Runs

Add cron trigger for nightly test generation:

```yaml
# .tekton/triggers/cron-trigger.yaml
apiVersion: triggers.tekton.dev/v1beta1
kind: TriggerTemplate
metadata:
  name: nightly-test-generation
spec:
  params:
    - name: branch
      default: main
  resourcetemplates:
    - apiVersion: tekton.dev/v1beta1
      kind: PipelineRun
      metadata:
        generateName: nightly-test-run-
      spec:
        pipelineRef:
          name: intelligent-test-orchestrator-pipeline
        params:
          - name: branch
            value: $(tt.params.branch)
```

---

## Additional Resources

- [IBM Cloud OnePipeline Documentation](https://cloud.ibm.com/docs/devsecops)
- [Tekton Documentation](https://tekton.dev/docs/)
- [Bob Shell Documentation](https://internal.bob.ibm.com/docs)
- [GitHub Actions to Tekton Migration](https://tekton.dev/docs/how-to-guides/migrating-from-github-actions/)
- [Project README](README.md)
- [OnePipeline Integration Guide](ONEPIPELINE_INTEGRATION.md)
- [End-to-End Flow](END_TO_END_FLOW.md)

---

## Support

For issues or questions:

1. Check the [Troubleshooting](#troubleshooting) section
2. Review pipeline logs in IBM Cloud UI
3. Check `.tekton/README.md` for Tekton-specific help
4. Open an issue in the repository

---

**Last Updated**: 2024-01-15
**Version**: 1.0.0