# 🤖 Intelligent Test Orchestrator

An AI-powered system that automatically detects feature changes in pull requests, generates contextually relevant test cases, and seamlessly integrates them into existing test suites.

## 🎯 Two Ways to Use

### 1. **Portable OnePipeline Step** (Recommended for Microservices)
Add AI test generation to **any microservice's existing `.one-pipeline.yaml`** with a single step. Uses Bob CLI directly - no additional setup required!

👉 **[Quick Integration Guide](./ONEPIPELINE_INTEGRATION.md)** - Add to your pipeline in 5 minutes

### 2. **Standalone Go Application**
Full-featured orchestrator with advanced capabilities for complex projects.

👉 **[Quick Start Guide](./QUICKSTART.md)** - Build and run locally

## 🌟 Features

### 1. **Smart Change Detection**
- Leverages AST analysis to understand semantic impact of code changes
- Identifies affected components, APIs, and business logic flows
- Supports multiple languages: Go, Python, JavaScript/TypeScript, Java

### 2. **Multi-Layer Test Generation**
- **Unit Tests**: Individual functions and methods
- **Integration Tests**: Component interactions
- **Functional Tests**: End-to-end feature behavior
- **AI-Enhanced Tests**: Uses Bob (Claude) for comprehensive test scenarios

### 3. **Intelligent Test Suite Integration**
- Automatically identifies appropriate test suite locations
- Maintains consistency with existing test structure
- Detects and prevents duplicate test scenarios
- Updates test fixtures and mocks as needed

### 4. **Automated Test Maintenance**
- Identifies obsolete tests when features are removed
- Updates existing tests when API signatures change
- Suggests test improvements based on coverage analysis
- Maintains test documentation and comments

### 5. **Coverage Analysis & Reporting**
- Tracks coverage before and after changes
- Generates detailed coverage reports
- Compares against configurable targets
- Provides actionable insights

## 🏗️ Architecture

```
test-genix/
├── main.go                    # Main CLI application
├── pkg/
│   ├── types/                 # Shared type definitions
│   │   └── types.go
│   ├── detector/              # Change detection module
│   │   └── change_detector.go
│   ├── generator/             # AI test generator
│   │   └── test_generator.go
│   ├── integrator/            # Test suite integrator
│   │   └── test_integrator.go
│   ├── prcreator/             # PR creation module
│   │   └── pr_creator.go
│   └── coverage/              # Coverage analyzer
│       └── coverage_analyzer.go
├── config/
│   └── test-orchestrator-config.yaml
├── .one-pipeline.yaml         # IBM Cloud OnePipeline config
└── go.mod
```

## 🚀 Quick Start

### Option 1: Add to Your Microservice Pipeline (5 minutes)

**Perfect for:** Adding AI test generation to existing microservices

1. **Copy the script to your repository:**
```bash
mkdir -p scripts
curl -o scripts/ai-test-generator.sh \
  https://raw.githubusercontent.com/your-org/test-genix/main/scripts/ai-test-generator.sh
chmod +x scripts/ai-test-generator.sh
```

2. **Add this step to your `.one-pipeline.yaml`:**
```yaml
ai-test-generation:
  abort_on_failure: false  # Non-blocking
  dind: true
  image: icr.io/continuous-delivery/pipeline/pipeline-base-image:2.71
  script: |
    #!/usr/bin/env bash
    ./scripts/ai-test-generator.sh \
      --base-branch "${BASE_BRANCH:-main}" \
      --coverage-target "${COVERAGE_TARGET:-80}"
```

3. **Set your API key in pipeline environment variables:**
```bash
BOBSHELL_API_KEY=your-api-key  # Get from Bob IDE or Slack
# See: https://internal.bob.ibm.com/docs/shell/install-and-setup
```

**That's it!** 🎉 Your pipeline now generates tests automatically for every PR.

📖 **[Full Integration Guide](./ONEPIPELINE_INTEGRATION.md)** | 📋 **[Example Pipeline](./examples/microservice-pipeline.yaml)** | 🔑 **[Bob Shell Setup](https://internal.bob.ibm.com/docs/shell/install-and-setup)**

---

### Option 2: Standalone Go Application (Advanced)

**Perfect for:** Complex projects requiring full control and customization

#### Prerequisites

- Go 1.21 or higher
- Git
- Bob/Anthropic API key
- (Optional) GitHub token for PR creation

#### Installation

1. Clone the repository:
```bash
git clone https://github.com/tanya-shanker/test-genix.git
cd test-genix
```

2. Install dependencies:
```bash
go mod download
```

3. Build the orchestrator:
```bash
make build
# or: go build -o bin/orchestrator .
```

#### Configuration

1. Copy the example configuration:
```bash
cp config/test-orchestrator-config.yaml config/my-config.yaml
```

2. Edit the configuration file with your settings:
```yaml
# Set your functional test repository
functional_test_repo: "your-org/functional-tests"

# Configure AI model (optional) - Bob/Claude
ai_model: "claude-3-5-sonnet-20241022"

# Set coverage target
coverage_target: 80.0
```

3. Set environment variables:
```bash
# For IBM Bob Shell (recommended - see https://internal.bob.ibm.com/docs/shell/install-and-setup):
export BOBSHELL_API_KEY="your-api-key"  # Get from Bob IDE or Slack
# OR use alternative keys:
export BOB_API_KEY="your-bob-key"
export ANTHROPIC_API_KEY="your-anthropic-api-key"
export CLAUDE_API_KEY="your-claude-api-key"

# For PR creation (optional):
export GITHUB_TOKEN="your-github-token"
```

**📖 Bob Shell Installation Guide:** [https://internal.bob.ibm.com/docs/shell/install-and-setup](https://internal.bob.ibm.com/docs/shell/install-and-setup)

**Installation:**
```bash
# macOS and Linux
curl -s https://s3.us-south.cloud-object-storage.appdomain.cloud/bobshell/install-bobshell.sh | bash
```

### Usage

#### Basic Usage

```bash
./bin/orchestrator \
  --base main \
  --head feature-branch \
  --config config/test-orchestrator-config.yaml \
  --output generated-tests \
  --root .
```

#### Command Line Options

- `--base`: Base branch for comparison (default: "main")
- `--head`: Head branch with changes (auto-detected from git)
- `--config`: Path to configuration file
- `--output`: Output directory for generated tests
- `--root`: Project root directory (default: ".")

## 🔧 IBM Cloud OnePipeline Integration

The Intelligent Test Orchestrator provides **two integration options** for IBM Cloud OnePipeline:

### 🎯 Recommended: Portable Pipeline Step

**Best for:** Adding AI test generation to existing microservices without modifying their structure.

Simply add the `ai-test-generation` step to any microservice's `.one-pipeline.yaml`:

```yaml
ai-test-generation:
  abort_on_failure: false
  dind: true
  image: icr.io/continuous-delivery/pipeline/pipeline-base-image:2.71
  script: |
    #!/usr/bin/env bash
    ./scripts/ai-test-generator.sh \
      --base-branch "${BASE_BRANCH:-main}" \
      --coverage-target "${COVERAGE_TARGET:-80}"
```

**Features:**
- ✅ Uses Bob CLI directly (no Go compilation needed)
- ✅ Non-blocking step (won't fail your pipeline)
- ✅ Automatic test generation and commit
- ✅ Coverage analysis and reporting
- ✅ Optional functional test PR creation
- ✅ Works with any language (Go, Python, JS/TS, Java)

📖 **[Complete Integration Guide](./ONEPIPELINE_INTEGRATION.md)**

---

### Advanced: Full Go Orchestrator

**Best for:** Projects requiring advanced customization and full control.

See the [`.one-pipeline.yaml`](./.one-pipeline.yaml) in this repository for a complete example using the Go orchestrator.

### Pipeline Configuration (Advanced)

The `.one-pipeline.yaml` file is pre-configured with three stages:

1. **Setup**: Installs Go and dependencies
2. **Test**: Runs existing test suite
3. **AI Test Generation**: Generates and integrates tests (non-blocking)

### Environment Variables

Configure these in your IBM Cloud OnePipeline:

**Required:**
- `GIT_BRANCH`: Current branch name
- `GIT_COMMIT`: Current commit SHA
- `BASE_BRANCH`: Base branch for comparison (default: main)

**Optional:**
- `ANTHROPIC_API_KEY`: Anthropic API key for Bob-enhanced tests
- `GITHUB_TOKEN`: GitHub token for creating PRs
- `FUNCTIONAL_TEST_REPO`: Repository for functional tests

### Pipeline Behavior

- ✅ **Non-blocking**: Test generation failures won't block the pipeline
- 📊 **Coverage Reports**: Automatically generated and displayed
- 🔗 **Artifact Archiving**: Generated tests are archived as pipeline artifacts
- 📝 **PR Creation**: Automatically creates PRs for functional tests

## 📊 Coverage Reports

The orchestrator generates comprehensive coverage reports:

```json
{
  "coverage_before": 75.5,
  "coverage_after": 82.3,
  "coverage_delta": 6.8,
  "new_lines_covered": 145,
  "new_lines_total": 200,
  "coverage_target": 80.0,
  "target_met": true,
  "timestamp": "2026-04-08T05:00:00Z"
}
```

## 🧪 Generated Test Examples

### Go Unit Test

```go
// Test CalculateTotal with valid inputs
func TestCalculateTotal_HappyPath(t *testing.T) {
    // Arrange
    items := []int{1, 2, 3, 4, 5}
    
    // Act
    result := CalculateTotal(items)
    
    // Assert
    if result != 15 {
        t.Errorf("Expected 15, got %d", result)
    }
}
```

### Python Unit Test

```python
def test_calculate_total_happy_path():
    """Test calculate_total with valid inputs"""
    # Arrange
    items = [1, 2, 3, 4, 5]
    
    # Act
    result = calculate_total(items)
    
    # Assert
    assert result == 15
```

## 🎯 Use Cases

### 1. Pull Request Validation
Automatically generate tests for every PR to ensure new code is properly tested.

### 2. Legacy Code Coverage
Gradually improve test coverage for legacy codebases by generating tests for modified code.

### 3. Regression Prevention
Generate tests that capture current behavior before refactoring.

### 4. Documentation
Generated tests serve as executable documentation for code behavior.

## 🔍 How It Works

1. **Change Detection**
   - Analyzes git diff between base and head branches
   - Uses AST parsing to understand semantic changes
   - Identifies modified functions, classes, and modules

2. **Test Generation**
   - Generates unit tests for each modified function/class
   - Creates integration tests for module interactions
   - Produces functional tests for high-impact changes
   - Optionally uses AI to enhance test quality

3. **Test Integration**
   - Identifies appropriate test file locations
   - Checks for duplicate tests
   - Merges new tests with existing test suites
   - Maintains consistent naming and structure

4. **PR Creation**
   - Commits unit tests to current PR branch
   - Creates separate PR for functional tests
   - Includes detailed descriptions and metadata

5. **Coverage Analysis**
   - Runs test suite with coverage enabled
   - Compares before/after coverage
   - Generates detailed reports
   - Validates against coverage targets

## 🛠️ Development

### Running Tests

```bash
go test ./... -v
```

### Running with Coverage

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Building for Production

```bash
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o orchestrator .
```

## 📝 Configuration Reference

See [`config/test-orchestrator-config.yaml`](config/test-orchestrator-config.yaml) for full configuration options.

Key settings:
- `test_frameworks`: Framework per language
- `coverage_target`: Minimum coverage percentage
- `generate_edge_cases`: Enable edge case test generation
- `ai_model`: AI model for enhanced generation
- `functional_test_repo`: Repository for functional tests

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with Go for performance and reliability
- Integrates with OpenAI for AI-enhanced test generation
- Designed for IBM Cloud OnePipeline
- Supports multiple programming languages and test frameworks

## 📞 Support

For issues, questions, or contributions, please open an issue on GitHub.

---

**Made with ❤️ by the Test Automation Team**