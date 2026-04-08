# Architecture Documentation - Intelligent Test Orchestrator

## Overview

The Intelligent Test Orchestrator is a Go-based system that automatically generates, integrates, and manages test cases for code changes in pull requests. It uses AI/LLM capabilities, AST analysis, and intelligent heuristics to create comprehensive test suites.

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    IBM Cloud OnePipeline                     │
│                                                               │
│  ┌──────────┐  ┌──────────┐  ┌─────────────────────────┐   │
│  │  Setup   │→ │   Test   │→ │  AI Test Generation     │   │
│  └──────────┘  └──────────┘  └─────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│              Intelligent Test Orchestrator                   │
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Change     │→ │     Test     │→ │     Test     │      │
│  │   Detector   │  │  Generator   │  │  Integrator  │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│         ↓                  ↓                  ↓              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Coverage   │  │      PR      │  │   Reports    │      │
│  │   Analyzer   │  │   Creator    │  │  Generator   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                    External Services                         │
│                                                               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ OpenAI   │  │  GitHub  │  │   Git    │  │  Test    │   │
│  │   API    │  │   API    │  │  Repo    │  │Framework │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Change Detector (`pkg/detector`)

**Purpose**: Analyzes git diffs to identify semantic code changes.

**Key Features**:
- Git diff parsing
- AST-based semantic analysis
- Multi-language support (Go, Python, JS/TS, Java)
- Function and class extraction
- Module dependency identification

**Flow**:
```
Git Diff → Parse Files → AST Analysis → Extract Changes → Identify Impact
```

**Output**: `ChangeInfo` structure containing:
- Changed/added/deleted files
- Modified functions and classes
- Affected modules
- Semantic change metadata

### 2. Test Generator (`pkg/generator`)

**Purpose**: Generates comprehensive test cases using AI and templates.

**Key Features**:
- Template-based test generation
- AI-enhanced test scenarios (via OpenAI)
- Multi-layer test creation (unit, integration, functional)
- Language-specific test frameworks
- Edge case and error handling tests

**Test Types**:
1. **Unit Tests**: Individual function/method tests
2. **Integration Tests**: Module interaction tests
3. **Functional Tests**: End-to-end feature tests
4. **Bob-Enhanced Tests**: Claude-generated comprehensive scenarios

**Flow**:
```
Changes → Analyze Context → Generate Templates → AI Enhancement → Write Tests
```

### 3. Test Integrator (`pkg/integrator`)

**Purpose**: Integrates generated tests into existing test suites.

**Key Features**:
- Duplicate detection
- Test file merging
- Naming convention preservation
- Git commit automation
- Conflict resolution

**Flow**:
```
Generated Tests → Check Duplicates → Merge/Copy → Commit → Push
```

### 4. PR Creator (`pkg/prcreator`)

**Purpose**: Creates pull requests for functional tests.

**Key Features**:
- Functional test repository management
- Branch creation and management
- GitHub API integration
- PR template customization
- Metadata inclusion

**Flow**:
```
Functional Tests → Clone Repo → Create Branch → Copy Tests → Commit → Create PR
```

### 5. Coverage Analyzer (`pkg/coverage`)

**Purpose**: Analyzes test coverage before and after changes.

**Key Features**:
- Multi-language coverage support
- Before/after comparison
- Target validation
- Detailed reporting
- Baseline caching

**Supported Tools**:
- Go: `go test -cover`
- Python: `pytest --cov`
- JavaScript: `jest --coverage`

**Flow**:
```
Run Tests → Parse Coverage → Compare → Generate Report → Validate Target
```

## Data Flow

### Complete Workflow

```
1. PR Created/Updated
   ↓
2. OnePipeline Triggered
   ↓
3. Change Detector Analyzes Diff
   ↓
4. Test Generator Creates Tests
   ├─→ Unit Tests (to current PR)
   └─→ Functional Tests (to separate PR)
   ↓
5. Test Integrator Merges Tests
   ↓
6. Coverage Analyzer Runs
   ↓
7. Reports Generated
   ↓
8. Artifacts Archived
```

### Data Structures

#### ChangeInfo
```go
type ChangeInfo struct {
    ChangedFiles      []FileChange
    AddedFiles        []FileChange
    DeletedFiles      []FileChange
    ModifiedFunctions []FunctionChange
    ModifiedClasses   []ClassChange
    AffectedModules   []string
    SemanticChanges   []SemanticChange
}
```

#### TestCase
```go
type TestCase struct {
    Name        string
    Description string
    Type        string  // happy_path, edge_case, error_handling
    Code        string
    FilePath    string
}
```

#### CoverageReport
```go
type CoverageReport struct {
    CoverageBefore   float64
    CoverageAfter    float64
    CoverageDelta    float64
    NewLinesCovered  int
    NewLinesTotal    int
    CoverageTarget   float64
    TargetMet        bool
}
```

## Configuration

### Configuration File Structure

```yaml
test_frameworks:      # Framework per language
  go: "testing"
  python: "pytest"

coverage_target: 80.0 # Minimum coverage %

generate_edge_cases: true
generate_mocks: true

functional_test_repo: "org/repo"

ai_model: "gpt-4"

test_patterns:
  unit: "test_{function_name}"
  functional: "test_{feature_name}_e2e"
```

### Environment Variables

- `OPENAI_API_KEY`: OpenAI API key
- `GITHUB_TOKEN`: GitHub access token
- `GIT_BRANCH`: Current branch
- `GIT_COMMIT`: Current commit SHA
- `BASE_BRANCH`: Base branch for comparison

## Integration Points

### 1. IBM Cloud OnePipeline

**Integration**: `.one-pipeline.yaml`

**Stages**:
1. Setup: Install Go and dependencies
2. Test: Run existing tests
3. AI Test Generation: Generate and integrate tests (non-blocking)

**Artifacts**:
- Generated tests archive
- Coverage reports
- Generation statistics

### 2. GitHub API

**Usage**:
- Create pull requests
- Add comments
- Update status checks

**Authentication**: GitHub token via environment variable

### 3. Anthropic API (Bob/Claude)

**Usage**:
- Enhance test generation with Bob
- Generate comprehensive scenarios
- Improve test quality

**Models**: Claude 3.5 Sonnet, Claude 3 Opus, Claude 3 Sonnet

### 4. Git Repository

**Operations**:
- Fetch diffs
- Create branches
- Commit changes
- Push updates

## Scalability Considerations

### Performance

- **Parallel Processing**: Multiple test generation workers
- **Caching**: Bob response caching to reduce API calls
- **Incremental Analysis**: Only analyze changed files

### Reliability

- **Non-blocking**: Pipeline failures don't block deployment
- **Retry Logic**: Automatic retry on transient failures
- **Graceful Degradation**: Works without Bob enhancement

### Extensibility

- **Plugin Architecture**: Easy to add new languages
- **Custom Templates**: Configurable test templates
- **Framework Agnostic**: Supports multiple test frameworks

## Security

### Secrets Management

- Anthropic API keys via environment variables
- No secrets in code or configuration
- Secure token handling

### Access Control

- GitHub token permissions
- Repository access validation
- PR creation authorization

## Monitoring & Observability

### Metrics

- Tests generated per PR
- Coverage improvement
- Generation time
- Success/failure rates

### Logging

- Structured logging
- Error tracking
- Performance metrics
- Audit trail

### Reports

- Coverage reports (JSON)
- Generation statistics
- Integration results
- PR creation status

## Future Enhancements

### Planned Features

1. **Machine Learning**: Learn from existing tests
2. **Test Prioritization**: Identify critical tests
3. **Mutation Testing**: Validate test effectiveness
4. **Visual Regression**: UI test generation
5. **Performance Tests**: Load and stress tests

### Roadmap

- Q1: Enhanced AI models
- Q2: Multi-repository support
- Q3: Test maintenance automation
- Q4: Advanced analytics

## Best Practices

### For Users

1. Review generated tests before merging
2. Customize configuration for your project
3. Set appropriate coverage targets
4. Use Bob enhancement for complex code

### For Developers

1. Follow Go best practices
2. Write comprehensive unit tests
3. Document public APIs
4. Use semantic versioning

## Troubleshooting

### Common Issues

1. **Build Failures**: Check Go version and dependencies
2. **No Tests Generated**: Verify code changes and language support
3. **Coverage Issues**: Ensure test infrastructure is set up
4. **PR Creation Fails**: Validate GitHub token and permissions

### Debug Mode

Enable verbose logging:
```bash
export DEBUG=true
./bin/orchestrator --base main --head feature
```

## References

- [Go Documentation](https://golang.org/doc/)
- [Anthropic API](https://docs.anthropic.com/)
- [GitHub API](https://docs.github.com/en/rest)
- [IBM Cloud OnePipeline](https://cloud.ibm.com/docs/ContinuousDelivery)

---

**Last Updated**: 2026-04-08
**Version**: 1.0.0