# IBM Bob Shell Integration Guide

This document explains how the Intelligent Test Orchestrator uses IBM Bob Shell (powered by Anthropic's Claude) for AI-enhanced test generation.

## 📖 Official Setup Guide

**For complete Bob Shell installation and setup:**
[https://internal.bob.ibm.com/docs/shell/install-and-setup](https://internal.bob.ibm.com/docs/shell/install-and-setup)

## Installation

### macOS and Linux

```bash
curl -s https://s3.us-south.cloud-object-storage.appdomain.cloud/bobshell/install-bobshell.sh | bash
```

### With Specific Package Manager

```bash
# Using pnpm
curl -s https://s3.us-south.cloud-object-storage.appdomain.cloud/bobshell/install-bobshell.sh | bash -s -- --pm pnpm

# Using npm
curl -s https://s3.us-south.cloud-object-storage.appdomain.cloud/bobshell/install-bobshell.sh | bash -s -- --pm npm
```

## Overview

The orchestrator leverages IBM Bob Shell (powered by Claude) to generate comprehensive, production-ready test cases that go beyond simple template-based tests. Bob Shell analyzes your code context and creates intelligent test scenarios covering edge cases, error handling, and boundary conditions.

## Why IBM Bob Shell?

- **Superior Code Understanding**: Claude excels at understanding code context and intent
- **Comprehensive Test Coverage**: Generates tests that cover edge cases humans might miss
- **Production-Ready Quality**: Creates well-structured, maintainable test code
- **Multi-Language Support**: Works across Go, Python, JavaScript/TypeScript, Java, and more
- **IBM Integration**: Seamlessly integrates with IBM Cloud and OnePipeline
- **File References**: Use `@` symbol to reference files in prompts
- **Safety Features**: Non-destructive tools by default in non-interactive mode

## Setup

### 1. Get an API Key

#### Recommended: IBM Bob Shell API Key

**📖 Follow the official installation guide:** [https://internal.bob.ibm.com/docs/shell/install-and-setup](https://internal.bob.ibm.com/docs/shell/install-and-setup)

**Getting Your API Key:**
1. Check Bob IDE for your API key
2. Look for Slack welcome message from Ask Bob that includes your API key
3. Note: This is the same API key used for Bob IDE

**Supported Environment Variables** (checked in order):
- `BOBSHELL_API_KEY` - IBM Bob Shell API key (recommended)
- `BOB_API_KEY` - IBM Bob API key
- `ANTHROPIC_API_KEY` - Standard Anthropic API key
- `CLAUDE_API_KEY` - Alternative Claude API key

#### Alternative: Using Anthropic API Directly
If you prefer to use Anthropic's API directly:
1. Sign up at [Anthropic Console](https://console.anthropic.com/)
2. Navigate to API Keys section
3. Create a new API key
4. Copy the key (starts with `sk-ant-`)

### 2. Configure the Orchestrator

#### Option A: Environment Variable (Recommended)

```bash
# Recommended: IBM Bob API key (see https://internal.bob.ibm.com/docs/ide)
export BOB_API_KEY="your-ibm-bob-api-key"

# Alternative: Standard Anthropic API
export ANTHROPIC_API_KEY="sk-ant-your-api-key-here"

# Alternative: Claude API key
export CLAUDE_API_KEY="your-claude-api-key"
```

#### For IBM Cloud OnePipeline

Set the `BOB_API_KEY` environment variable in your pipeline configuration:

1. Navigate to your pipeline settings in IBM Cloud
2. Add environment variable: `BOB_API_KEY`
3. Set the value to your IBM Bob API key (obtained from the [installation guide](https://internal.bob.ibm.com/docs/shell/install-and-setup))
4. Save and trigger a new build

#### Option B: Configuration File

Edit `config/test-orchestrator-config.yaml`:

```yaml
ai_api_key: "sk-ant-your-api-key-here"
```

### 3. Choose Your Model

Available Claude models:

- **claude-3-5-sonnet-20241022** (Recommended) - Best balance of speed and quality
- **claude-3-opus-20240229** - Highest quality, slower
- **claude-3-sonnet-20240229** - Fast, good quality

Configure in `config/test-orchestrator-config.yaml`:

```yaml
ai_model: "claude-3-5-sonnet-20241022"
```

## How It Works

### 1. Template-Based Generation

First, the orchestrator generates basic test templates:

```go
func TestCalculate_HappyPath(t *testing.T) {
    // Arrange
    // TODO: Set up test data
    
    // Act
    result := Calculate()
    
    // Assert
    if result == nil {
        t.Error("Expected non-nil result")
    }
}
```

### 2. Bob Enhancement

Then, Bob analyzes the source code and enhances the tests:

```go
func TestCalculate_BobEnhanced(t *testing.T) {
    // Test with various input combinations
    tests := []struct {
        name     string
        input    int
        expected int
        wantErr  bool
    }{
        {"positive number", 5, 25, false},
        {"zero", 0, 0, false},
        {"negative number", -5, 25, false},
        {"large number", 1000000, 1000000000000, false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Calculate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Calculate() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if result != tt.expected {
                t.Errorf("Calculate() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### 3. Context-Aware Generation

Bob receives:
- Full source code of the function/class
- Function signature and parameters
- Language and framework information
- Existing code patterns

Bob generates:
- Comprehensive test scenarios
- Edge case handling
- Error condition tests
- Boundary value tests
- Integration test suggestions

## Features

### 1. Smart Test Scenarios

Bob generates tests for:
- ✅ Happy path (normal operation)
- ✅ Edge cases (empty inputs, null values, boundaries)
- ✅ Error handling (invalid inputs, exceptions)
- ✅ Boundary conditions (min/max values, limits)
- ✅ Integration scenarios (component interactions)

### 2. Language-Specific Best Practices

Bob adapts to each language's idioms:

**Go:**
- Table-driven tests
- Subtests with `t.Run()`
- Proper error handling
- Benchmark tests

**Python:**
- Pytest fixtures
- Parametrized tests
- Context managers
- Mock objects

**JavaScript/TypeScript:**
- Jest matchers
- Async/await handling
- Mock functions
- Snapshot testing

### 3. Framework Integration

Bob generates tests compatible with:
- Go: `testing` package
- Python: `pytest`, `unittest`
- JavaScript: `jest`, `mocha`
- TypeScript: `jest`, `vitest`
- Java: `junit`

## Usage Examples

### Basic Usage

```bash
# With Bob enhancement
export ANTHROPIC_API_KEY="sk-ant-your-key"
./bin/orchestrator --base main --head feature-branch
```

### Without Bob Enhancement

```bash
# Template-based tests only (no API key needed)
./bin/orchestrator --base main --head feature-branch
```

### Custom Model

```bash
# Use Claude Opus for highest quality
export ANTHROPIC_API_KEY="sk-ant-your-key"
# Edit config to set ai_model: "claude-3-opus-20240229"
./bin/orchestrator --base main --head feature-branch
```

## Cost Considerations

### Token Usage

Bob is called once per:
- Modified function
- Modified class/struct
- High-impact semantic change

Typical costs per PR:
- Small PR (1-3 functions): ~$0.01-0.05
- Medium PR (5-10 functions): ~$0.05-0.15
- Large PR (20+ functions): ~$0.15-0.50

### Cost Optimization

1. **Caching**: Enable response caching in config
2. **Selective Enhancement**: Only enhance complex functions
3. **Model Selection**: Use Sonnet for most cases, Opus for critical code

```yaml
advanced:
  cache_ai_responses: true
  cache_duration: 24  # hours
```

## Best Practices

### 1. Review Generated Tests

Always review Bob-generated tests:
- Verify test logic is correct
- Ensure assertions match requirements
- Add domain-specific test cases
- Update test data as needed

### 2. Iterative Improvement

Bob learns from context:
- Keep existing tests well-structured
- Use clear function names
- Add comments for complex logic
- Maintain consistent patterns

### 3. Security

- Never commit API keys to git
- Use environment variables
- Rotate keys regularly
- Monitor API usage

## Troubleshooting

### Bob Not Generating Tests

**Check:**
1. API key is set: `echo $ANTHROPIC_API_KEY`
2. API key is valid (starts with `sk-ant-`)
3. Network connectivity to Anthropic API
4. Source code is readable

**Solution:**
```bash
# Test API key
curl https://api.anthropic.com/v1/messages \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{"model":"claude-3-5-sonnet-20241022","max_tokens":1024,"messages":[{"role":"user","content":"Hello"}]}'
```

### Rate Limiting

If you hit rate limits:
1. Reduce concurrent test generation
2. Enable caching
3. Upgrade Anthropic plan
4. Use template-only mode temporarily

### Quality Issues

If generated tests aren't good enough:
1. Switch to Claude Opus model
2. Provide more context in source code
3. Add comments explaining complex logic
4. Review and refine generated tests

## Advanced Configuration

### Custom System Prompt

Modify the system prompt in `pkg/generator/test_generator.go`:

```go
System: anthropic.F([]anthropic.TextBlockParam{
    anthropic.NewTextBlock("You are Bob, an expert software testing engineer specializing in [your domain]. Generate comprehensive, production-ready test cases following [your standards]."),
}),
```

### Response Caching

Enable in config:

```yaml
advanced:
  cache_ai_responses: true
  cache_duration: 24
  max_workers: 4
```

### Parallel Generation

Configure workers:

```yaml
advanced:
  parallel_generation: true
  max_workers: 4  # Adjust based on API limits
```

## Comparison: Bob vs Template-Only

| Feature | Template-Only | With Bob |
|---------|--------------|----------|
| Speed | Fast | Moderate |
| Cost | Free | ~$0.01-0.50/PR |
| Coverage | Basic | Comprehensive |
| Edge Cases | Manual | Automatic |
| Quality | Good | Excellent |
| Maintenance | Higher | Lower |

## Support

For issues with Bob integration:
1. Check [Anthropic Documentation](https://docs.anthropic.com/)
2. Review [API Status](https://status.anthropic.com/)
3. Open an issue on GitHub
4. Contact support@anthropic.com

## Future Enhancements

Planned improvements:
- [ ] Streaming responses for faster generation
- [ ] Multi-turn conversations for test refinement
- [ ] Custom test templates per project
- [ ] Learning from test execution results
- [ ] Integration with test coverage tools

---

**Made with ❤️ using Bob (Claude)**