# 🔧 OnePipeline Integration Guide

This guide shows how to integrate the **AI Test Generator** into any microservice's existing `.one-pipeline.yaml` file using **IBM Bob Shell**.

## 📚 Prerequisites

Before integrating the AI Test Generator, you need to set up Bob Shell in your pipeline environment:

**📖 Bob Shell Installation Guide:** [https://internal.bob.ibm.com/docs/shell/install-and-setup](https://internal.bob.ibm.com/docs/shell/install-and-setup)

### Installation

Bob Shell can be installed using:

```bash
# For macOS and Linux
curl -s https://s3.us-south.cloud-object-storage.appdomain.cloud/bobshell/install-bobshell.sh | bash

# With specific package manager (e.g., pnpm)
curl -s https://s3.us-south.cloud-object-storage.appdomain.cloud/bobshell/install-bobshell.sh | bash -s -- --pm pnpm
```

### Authentication

Get your API key from:
- Bob IDE
- Slack welcome message from Ask Bob

Set the environment variable:
```bash
export BOBSHELL_API_KEY=your-api-key
```

## 📋 Quick Integration

### Step 1: Add the Script to Your Repository

Copy the `scripts/ai-test-generator.sh` script to your microservice repository:

```bash
# Create scripts directory if it doesn't exist
mkdir -p scripts

# Copy the AI test generator script
curl -o scripts/ai-test-generator.sh \
  https://raw.githubusercontent.com/your-org/test-genix/main/scripts/ai-test-generator.sh

# Make it executable
chmod +x scripts/ai-test-generator.sh
```

### Step 2: Add the Pipeline Step

Add this step to your microservice's `.one-pipeline.yaml`:

```yaml
ai-test-generation:
  abort_on_failure: false  # Non-blocking step
  dind: true
  image: icr.io/continuous-delivery/pipeline/pipeline-base-image:2.71
  script: |
    #!/usr/bin/env bash
    set -eo pipefail
    
    echo "🤖 Starting AI Test Generation..."
    
    # Run the AI test generator
    ./scripts/ai-test-generator.sh \
      --base-branch "${BASE_BRANCH:-main}" \
      --coverage-target "${COVERAGE_TARGET:-80}" \
      --functional-repo "${FUNCTIONAL_TEST_REPO:-}"
    
    echo "✅ AI Test Generation Complete"
```

### Step 3: Configure Environment Variables

Set these environment variables in your IBM Cloud OnePipeline configuration:

**Required:**
- `BOBSHELL_API_KEY` - Your IBM Bob Shell API key (recommended)
  - Get from Bob IDE or Slack welcome message from Ask Bob
  - Alternative: `BOB_API_KEY`, `ANTHROPIC_API_KEY`, or `CLAUDE_API_KEY`
- `GIT_BRANCH` - Current branch name (usually auto-set by pipeline)
- `GIT_COMMIT` - Current commit SHA (usually auto-set by pipeline)

**📖 For Bob Shell setup:** Follow the instructions at [https://internal.bob.ibm.com/docs/shell/install-and-setup](https://internal.bob.ibm.com/docs/shell/install-and-setup)

**Optional:**
- `BASE_BRANCH` - Base branch for comparison (default: `main`)
- `COVERAGE_TARGET` - Target coverage percentage (default: `80`)
- `FUNCTIONAL_TEST_REPO` - Repository for functional tests (e.g., `org/functional-tests`)
- `GITHUB_TOKEN` - GitHub token for creating PRs (required if using functional test repo)

## 🎯 Complete Example

Here's a complete example of a microservice's `.one-pipeline.yaml` with AI test generation:

```yaml
version: '1'

setup:
  image: icr.io/continuous-delivery/pipeline/pipeline-base-image:2.71
  script: |
    #!/usr/bin/env bash
    set -eo pipefail
    
    echo "Setting up environment..."
    
    # Install dependencies
    apt-get update && apt-get install -y jq curl git
    
    # Install language-specific tools
    if [ -f "go.mod" ]; then
      # Go setup
      go mod download
    elif [ -f "package.json" ]; then
      # Node.js setup
      npm install
    elif [ -f "requirements.txt" ]; then
      # Python setup
      pip install -r requirements.txt
    fi
    
    echo "✅ Setup complete"

test:
  abort_on_failure: false
  image: icr.io/continuous-delivery/pipeline/pipeline-base-image:2.71
  script: |
    #!/usr/bin/env bash
    set -eo pipefail
    
    echo "Running existing test suite..."
    
    # Run tests based on project type
    if [ -f "go.mod" ]; then
      go test ./... -v -coverprofile=coverage.out
      go tool cover -func=coverage.out
    elif [ -f "package.json" ]; then
      npm test -- --coverage
    elif [ -f "requirements.txt" ]; then
      pytest --cov --cov-report=term
    fi

ai-test-generation:
  abort_on_failure: false  # Non-blocking - won't fail the pipeline
  dind: true
  image: icr.io/continuous-delivery/pipeline/pipeline-base-image:2.71
  script: |
    #!/usr/bin/env bash
    set -eo pipefail
    
    echo "=================================================="
    echo "🤖 AI Test Generation with Bob"
    echo "=================================================="
    echo ""
    
    # Check if this is a PR build
    if [ -z "$GIT_BRANCH" ]; then
      echo "⚠️  Not a PR build - skipping AI test generation"
      exit 0
    fi
    
    echo "📊 PR Information:"
    echo "   Branch: $GIT_BRANCH"
    echo "   Commit: $GIT_COMMIT"
    echo "   Base: ${BASE_BRANCH:-main}"
    echo ""
    
    # Run AI test generator
    ./scripts/ai-test-generator.sh \
      --base-branch "${BASE_BRANCH:-main}" \
      --coverage-target "${COVERAGE_TARGET:-80}" \
      --functional-repo "${FUNCTIONAL_TEST_REPO:-}"
    
    # Display summary
    echo ""
    echo "=================================================="
    echo "📈 Test Generation Summary"
    echo "=================================================="
    
    if [ -f ".ai-generated-tests/coverage-report.json" ]; then
      cat .ai-generated-tests/coverage-report.json | jq -r '
        "Coverage Before: \(.coverage_before)%",
        "Coverage After:  \(.coverage_after)%",
        "Coverage Delta:  \(.coverage_delta)%",
        "",
        "Target: \(.coverage_target)%",
        "Target Met: \(.target_met)"
      '
    fi
    
    echo ""
    echo "✅ AI Test Generation Complete"

build:
  image: icr.io/continuous-delivery/pipeline/pipeline-base-image:2.71
  script: |
    #!/usr/bin/env bash
    echo "Building application..."
    # Your build commands here

deploy:
  image: icr.io/continuous-delivery/pipeline/pipeline-base-image:2.71
  script: |
    #!/usr/bin/env bash
    echo "Deploying application..."
    # Your deployment commands here
```

## 🎨 Customization Options

### Option 1: Skip Functional Tests

If you only want unit tests in the same PR:

```yaml
ai-test-generation:
  script: |
    ./scripts/ai-test-generator.sh \
      --base-branch main \
      --skip-functional
```

### Option 2: Dry Run Mode

Test the script without making changes:

```yaml
ai-test-generation:
  script: |
    ./scripts/ai-test-generator.sh \
      --base-branch main \
      --dry-run
```

### Option 3: Custom Coverage Target

Set a specific coverage target:

```yaml
ai-test-generation:
  script: |
    ./scripts/ai-test-generator.sh \
      --base-branch main \
      --coverage-target 90
```

### Option 4: Functional Test Repository

Generate functional tests in a separate repository:

```yaml
ai-test-generation:
  script: |
    ./scripts/ai-test-generator.sh \
      --base-branch main \
      --functional-repo "your-org/functional-tests"
```

## 🔑 API Key Configuration

### Setting Up Bob Shell API Key

**📖 Official Setup Guide:** [https://internal.bob.ibm.com/docs/shell/install-and-setup](https://internal.bob.ibm.com/docs/shell/install-and-setup)

The script supports multiple API key environment variables (checked in order):

1. **`BOBSHELL_API_KEY`** - IBM Bob Shell API key (recommended)
   - Get from Bob IDE or Slack welcome message from Ask Bob
2. **`BOB_API_KEY`** - IBM Bob API key
3. **`ANTHROPIC_API_KEY`** - Standard Anthropic API key
4. **`CLAUDE_API_KEY`** - Alternative Claude API key

**To configure in IBM Cloud OnePipeline:**
1. Navigate to your pipeline settings
2. Add environment variable: `BOBSHELL_API_KEY`
3. Set the value to your IBM Bob Shell API key
4. Save and trigger a new build

**Getting Your API Key:**
- Check Bob IDE for your API key
- Look for Slack welcome message from Ask Bob
- Follow the [official documentation](https://internal.bob.ibm.com/docs/shell/install-and-setup) for detailed instructions

## 📊 What Gets Generated

### Unit Tests
- Generated in the same directory as source files
- Follow language-specific naming conventions:
  - Go: `*_test.go`
  - Python: `test_*.py`
  - JavaScript/TypeScript: `*.test.js` or `*.test.ts`
  - Java: `*Test.java`

### Functional Tests
- Generated in `.ai-generated-tests/functional/`
- Automatically committed to a separate PR in the functional test repository
- Include end-to-end test scenarios

### Coverage Reports
- Saved to `.ai-generated-tests/coverage-report.json`
- Displayed in pipeline logs
- Include before/after comparison

## 🚀 Features

### ✅ Smart Change Detection
- Analyzes git diff to identify changed files
- Skips test files, config files, and documentation
- Focuses on source code files

### ✅ AI-Powered Test Generation
- Uses Bob CLI (Claude 3.5 Sonnet) for intelligent test creation
- Generates comprehensive test scenarios including:
  - Happy path cases
  - Edge cases and boundary conditions
  - Error handling
  - Integration points

### ✅ Multi-Language Support
- **Go**: Full support with `go test` integration
- **Python**: Support with pytest and coverage.py
- **JavaScript/TypeScript**: Support with Jest/npm test
- **Java**: Basic support with JUnit conventions

### ✅ Automatic Integration
- Tests are automatically committed to the current PR
- Functional tests create separate PRs in the functional test repo
- Maintains existing test structure and conventions

### ✅ Coverage Analysis
- Measures coverage before and after test generation
- Compares against configurable targets
- Provides detailed reports in pipeline logs

## 🔍 Troubleshooting

### Issue: Bob CLI Not Found

**Solution:** The script will automatically try to install Bob CLI. If that fails, it will fall back to direct API calls.

### Issue: No API Key

**Error:** `No API key found. Please set one of: BOB_API_KEY, ANTHROPIC_API_KEY, CLAUDE_API_KEY`

**Solution:** Set one of the API key environment variables in your pipeline configuration.

### Issue: No Changes Detected

**Warning:** `No changes detected`

**Solution:** This is normal if there are no code changes in the PR. The script will exit gracefully.

### Issue: Failed to Generate Tests

**Warning:** `Could not extract test code from Bob's response`

**Solution:** This can happen if Bob's response doesn't contain code blocks. The script will continue with other files.

## 📝 Best Practices

1. **Run as Non-Blocking**: Always set `abort_on_failure: false` so test generation failures don't block your pipeline

2. **Review Generated Tests**: AI-generated tests should be reviewed before merging, just like any other code

3. **Adjust Coverage Targets**: Start with a reasonable target (e.g., 80%) and adjust based on your project needs

4. **Use Functional Test Repo**: For large projects, separate functional tests into their own repository for better organization

5. **Monitor API Usage**: Bob API calls may have rate limits. Monitor your usage and follow IBM's guidelines

6. **Follow Bob Setup Guide**: Always refer to [https://internal.bob.ibm.com/docs/shell/install-and-setup](https://internal.bob.ibm.com/docs/shell/install-and-setup) for the latest setup instructions and best practices

## 🎯 Example Workflow

1. **Developer creates PR** with code changes
2. **Pipeline runs** and detects the PR
3. **AI Test Generator** analyzes the changes
4. **Bob generates tests** for modified files
5. **Tests are committed** to the PR automatically
6. **Coverage report** is displayed in pipeline logs
7. **Functional tests** (if any) create a separate PR
8. **Developer reviews** and merges both PRs

## 📚 Additional Resources

- **[IBM Bob CLI Installation](https://internal.bob.ibm.com/docs/shell/install-and-setup)** - Official Bob CLI setup and configuration
- [IBM Cloud OnePipeline Docs](https://cloud.ibm.com/docs/ContinuousDelivery)
- [Test Orchestrator Architecture](./ARCHITECTURE.md)
- [Quick Start Guide](./QUICKSTART.md)
- [Bob Integration Details](./BOB_INTEGRATION.md)

## 🆘 Support

For issues or questions:
1. Check the troubleshooting section above
2. Review the script logs in your pipeline output
3. Open an issue in the test-genix repository
4. Contact your DevOps team

---

**Made with Bob (Claude 3.5 Sonnet)** 🤖