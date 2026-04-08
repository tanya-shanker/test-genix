# Quick Start Guide - Intelligent Test Orchestrator

This guide will help you get started with the Intelligent Test Orchestrator in 5 minutes.

## Prerequisites

- Go 1.21 or higher installed
- Git repository with code changes
- (Optional) Anthropic API key for Bob-enhanced tests

## Step 1: Build the Orchestrator

```bash
# Clone the repository
git clone https://github.com/tanya-shanker/test-genix.git
cd test-genix

# Build using Make
make build

# Or build manually
go build -o bin/orchestrator ./cmd/orchestrator
```

## Step 2: Configure (Optional)

The orchestrator works with default settings, but you can customize:

```bash
# Copy example config
cp config/test-orchestrator-config.yaml config/my-config.yaml

# Edit your config
vim config/my-config.yaml
```

Key settings to customize:
- `functional_test_repo`: Your functional test repository
- `coverage_target`: Desired coverage percentage (default: 80%)
- `ai_model`: AI model to use (default: claude-3-5-sonnet-20241022)

## Step 3: Set Environment Variables (Optional)

For Bob-enhanced test generation (any of these will work):
```bash
export ANTHROPIC_API_KEY="your-anthropic-api-key"  # Standard Anthropic
# OR
export BOB_API_KEY="your-bob-api-key"              # Bob-specific
# OR
export CLAUDE_API_KEY="your-claude-api-key"        # Alternative
```

**Note:** If you're using Bob within Roo Code/Cline, the API key may already be configured and will be automatically detected.

For PR creation:
```bash
export GITHUB_TOKEN="your-github-token"
```

## Step 4: Run the Orchestrator

### Basic Usage

```bash
./bin/orchestrator \
  --base main \
  --head feature-branch \
  --config config/test-orchestrator-config.yaml \
  --output generated-tests \
  --root .
```

### Using Make

```bash
make run-example
```

### In IBM Cloud OnePipeline

The orchestrator runs automatically in your pipeline. Just ensure:
1. `.one-pipeline.yaml` is in your repository
2. Environment variables are set in pipeline configuration
3. Push your changes to trigger the pipeline

## Step 5: Review Generated Tests

After running, check:

```bash
# View generated unit tests
ls -la generated-tests/unit/

# View functional tests
ls -la generated-tests/functional/

# Check coverage report
cat coverage-report.json

# View detailed reports
ls -la generated-tests/reports/
```

## Example Output

```
==================================================
🤖 Intelligent Test Orchestrator
==================================================

🔍 Analyzing code changes...
✅ Detected 3 changed files
✅ Identified 5 modified functions
✅ Identified 2 modified classes

🤖 Generating AI-powered tests...
✅ Generated 8 unit tests
✅ Generated 2 functional tests

🔗 Integrating unit tests into test suite...
✅ Integrated: test_calculator.go
✅ Integration complete: 8 files integrated, 0 duplicates skipped

📊 Analyzing test coverage...
✅ Coverage analysis complete
   Before: 75.50%
   After:  82.30%
   Delta:  +6.80%
   Target: 80.00% (✅ Met)

==================================================
📋 Summary
==================================================

Unit Tests Generated: 8
Functional Tests Generated: 2
Total Test Cases: 15
Execution Time: 12.34s

📁 Generated artifacts:
   - Unit tests: generated-tests/unit
   - Functional tests: generated-tests/functional
   - Reports: generated-tests/reports

==================================================
```

## Testing the Example

Try the orchestrator with the included example:

```bash
# Run tests on the example calculator
go test ./examples/... -v

# Generate tests for the example
./bin/orchestrator \
  --base main \
  --head $(git rev-parse --abbrev-ref HEAD) \
  --config config/test-orchestrator-config.yaml \
  --output generated-tests \
  --root .
```

## Common Commands

```bash
# Build
make build

# Run tests
make test

# Run with coverage
make coverage

# Clean artifacts
make clean

# Format code
make fmt

# Run linters
make lint

# Show help
make help
```

## Troubleshooting

### Build Fails

```bash
# Ensure Go is installed
go version

# Download dependencies
go mod download
go mod tidy

# Try building again
make build
```

### No Tests Generated

Check that:
1. You have code changes in your branch
2. Changed files are in supported languages (Go, Python, JS/TS, Java)
3. Changes include functions or classes (not just comments)

### Coverage Analysis Fails

Ensure you have test infrastructure:
- Go: `go test` works
- Python: `pytest` installed
- JavaScript: `jest` configured

### PR Creation Fails

Verify:
1. `GITHUB_TOKEN` is set
2. `functional_test_repo` is configured correctly
3. You have write access to the functional test repository

## Next Steps

1. **Integrate with CI/CD**: Add to your IBM Cloud OnePipeline
2. **Customize Configuration**: Adjust settings for your project
3. **Enable AI Enhancement**: Add OpenAI API key for better tests
4. **Set Up Functional Tests**: Configure functional test repository
5. **Review Generated Tests**: Ensure quality and relevance

## Getting Help

- Check the [README](README.md) for detailed documentation
- Review [configuration options](config/test-orchestrator-config.yaml)
- See [example tests](examples/) for reference
- Open an issue on GitHub for support

## IBM Cloud OnePipeline Integration

To use in IBM Cloud OnePipeline:

1. Ensure `.one-pipeline.yaml` is in your repository
2. Configure environment variables in pipeline settings:
   - `ANTHROPIC_API_KEY` (optional - for Bob/Claude)
   - `GITHUB_TOKEN` (optional)
   - `FUNCTIONAL_TEST_REPO` (optional)
3. Push changes to trigger pipeline
4. Review generated tests in pipeline artifacts

The orchestrator runs as a non-blocking stage, so pipeline failures won't block your deployment.

---

**Ready to generate tests automatically? Start with `make build && make run-example`!**