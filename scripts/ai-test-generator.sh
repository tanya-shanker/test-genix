#!/usr/bin/env bash
###############################################################################
# AI Test Generator - Portable OnePipeline Step
#
# This script can be added to any microservice's .one-pipeline.yaml to enable
# AI-powered test generation using IBM Bob CLI.
#
# Prerequisites:
#   - Bob CLI installed (see: https://internal.bob.ibm.com/docs/shell/install-and-setup)
#   - BOB_API_KEY environment variable set
#   - Git repository with PR context
#
# Setup Guide:
#   For Bob CLI installation and configuration in pipeline environments:
#   https://internal.bob.ibm.com/docs/shell/install-and-setup
#
# Usage:
#   ./ai-test-generator.sh [options]
#
# Options:
#   --base-branch BRANCH    Base branch for comparison (default: main)
#   --functional-repo REPO  Functional test repository (optional)
#   --coverage-target PCT   Coverage target percentage (default: 80)
#   --skip-functional       Skip functional test generation
#   --dry-run              Show what would be done without making changes
###############################################################################

set -eo pipefail

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
BASE_BRANCH="${BASE_BRANCH:-main}"
COVERAGE_TARGET="${COVERAGE_TARGET:-80}"
FUNCTIONAL_TEST_REPO="${FUNCTIONAL_TEST_REPO:-}"
SKIP_FUNCTIONAL="${SKIP_FUNCTIONAL:-false}"
DRY_RUN="${DRY_RUN:-false}"
WORK_DIR="$(pwd)"
GENERATED_TESTS_DIR="${WORK_DIR}/.ai-generated-tests"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --base-branch)
      BASE_BRANCH="$2"
      shift 2
      ;;
    --functional-repo)
      FUNCTIONAL_TEST_REPO="$2"
      shift 2
      ;;
    --coverage-target)
      COVERAGE_TARGET="$2"
      shift 2
      ;;
    --skip-functional)
      SKIP_FUNCTIONAL=true
      shift
      ;;
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

###############################################################################
# Helper Functions
###############################################################################

log_info() {
  echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
  echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
  echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
  echo -e "${RED}❌ $1${NC}"
}

print_header() {
  echo ""
  echo "=================================================="
  echo "$1"
  echo "=================================================="
  echo ""
}

###############################################################################
# Installation Functions
###############################################################################

install_bob_cli() {
  log_info "Checking for Bob Shell..."
  
  if command -v bob &> /dev/null; then
    log_success "Bob Shell already installed: $(bob --version 2>&1 || echo 'version unknown')"
    return 0
  fi
  
  log_info "Bob Shell not found. Installing..."
  log_info "Reference: https://internal.bob.ibm.com/docs/shell/install-and-setup"
  
  # Install Bob Shell using the official installation script
  # For macOS and Linux
  if [[ "$OSTYPE" == "darwin"* ]] || [[ "$OSTYPE" == "linux-gnu"* ]]; then
    log_info "Installing Bob Shell for macOS/Linux..."
    
    # Check if we can specify package manager
    if command -v pnpm &> /dev/null; then
      curl -s https://s3.us-south.cloud-object-storage.appdomain.cloud/bobshell/install-bobshell.sh | bash -s -- --pm pnpm 2>/dev/null || \
      curl -s https://s3.us-south.cloud-object-storage.appdomain.cloud/bobshell/install-bobshell.sh | bash 2>/dev/null
    elif command -v npm &> /dev/null; then
      curl -s https://s3.us-south.cloud-object-storage.appdomain.cloud/bobshell/install-bobshell.sh | bash -s -- --pm npm 2>/dev/null || \
      curl -s https://s3.us-south.cloud-object-storage.appdomain.cloud/bobshell/install-bobshell.sh | bash 2>/dev/null
    else
      curl -s https://s3.us-south.cloud-object-storage.appdomain.cloud/bobshell/install-bobshell.sh | bash 2>/dev/null
    fi
  else
    log_warning "Unsupported OS: $OSTYPE"
    log_warning "Please install Bob Shell manually following:"
    log_warning "  https://internal.bob.ibm.com/docs/shell/install-and-setup"
  fi
  
  # Check if installation succeeded
  if command -v bob &> /dev/null; then
    log_success "Bob Shell installed successfully"
    return 0
  fi
  
  log_warning "Bob Shell installation failed or not found in PATH."
  log_warning "Please install manually following:"
  log_warning "  https://internal.bob.ibm.com/docs/shell/install-and-setup"
  log_warning ""
  log_info "Continuing with direct API fallback..."
  
  return 1
}

check_prerequisites() {
  log_info "Checking prerequisites..."
  
  # Check for required tools
  local missing_tools=()
  
  for tool in git jq curl; do
    if ! command -v "$tool" &> /dev/null; then
      missing_tools+=("$tool")
    fi
  done
  
  if [ ${#missing_tools[@]} -gt 0 ]; then
    log_error "Missing required tools: ${missing_tools[*]}"
    log_info "Installing missing tools..."
    apt-get update && apt-get install -y "${missing_tools[@]}" || {
      log_error "Failed to install required tools"
      return 1
    }
  fi
  
  # Check for Bob API key (BOBSHELL_API_KEY or alternatives)
  if [ -z "$BOBSHELL_API_KEY" ] && [ -z "$BOB_API_KEY" ] && [ -z "$ANTHROPIC_API_KEY" ] && [ -z "$CLAUDE_API_KEY" ]; then
    log_error "No API key found. Please set one of:"
    log_error "  - BOBSHELL_API_KEY (recommended for Bob Shell)"
    log_error "  - BOB_API_KEY"
    log_error "  - ANTHROPIC_API_KEY"
    log_error "  - CLAUDE_API_KEY"
    log_error ""
    log_error "Get your API key from Bob IDE or Slack welcome message from Ask Bob"
    return 1
  fi
  
  # Set unified API keys for both Bob Shell and direct API
  export BOBSHELL_API_KEY="${BOBSHELL_API_KEY:-${BOB_API_KEY:-${ANTHROPIC_API_KEY:-$CLAUDE_API_KEY}}}"
  export ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-${BOB_API_KEY:-${BOBSHELL_API_KEY:-$CLAUDE_API_KEY}}}"
  
  log_success "All prerequisites met"
  return 0
}

###############################################################################
# Git Operations
###############################################################################

get_pr_diff() {
  log_info "Fetching PR diff..."
  
  # Ensure we have the latest changes
  git fetch origin "$BASE_BRANCH" 2>/dev/null || true
  
  # Get the diff
  local diff_output
  diff_output=$(git diff "origin/$BASE_BRANCH"...HEAD 2>/dev/null || git diff HEAD~1 2>/dev/null || echo "")
  
  if [ -z "$diff_output" ]; then
    log_warning "No changes detected"
    return 1
  fi
  
  echo "$diff_output" > "${GENERATED_TESTS_DIR}/pr-diff.txt"
  
  local changed_files
  changed_files=$(git diff --name-only "origin/$BASE_BRANCH"...HEAD 2>/dev/null || git diff --name-only HEAD~1 2>/dev/null || echo "")
  
  local file_count
  file_count=$(echo "$changed_files" | grep -v '^$' | wc -l)
  
  log_success "Detected changes in $file_count file(s)"
  echo "$changed_files" | head -10
  
  if [ "$file_count" -gt 10 ]; then
    log_info "... and $((file_count - 10)) more files"
  fi
  
  return 0
}

###############################################################################
# Test Generation with Bob CLI
###############################################################################

generate_tests_with_bob() {
  local file_path="$1"
  local test_type="$2"  # unit, integration, or functional
  
  log_info "Generating $test_type tests for: $file_path"
  
  # Read the file content
  if [ ! -f "$file_path" ]; then
    log_warning "File not found: $file_path"
    return 1
  fi
  
  local file_content
  file_content=$(cat "$file_path")
  
  local file_ext="${file_path##*.}"
  local language="unknown"
  
  # Detect language
  case "$file_ext" in
    go) language="Go" ;;
    py) language="Python" ;;
    js|ts|jsx|tsx) language="JavaScript/TypeScript" ;;
    java) language="Java" ;;
    *) language="$file_ext" ;;
  esac
  
  # Create prompt for Bob
  local prompt="You are an expert test engineer. Analyze the following $language code and generate comprehensive $test_type tests.

File: $file_path
Language: $language

Code:
\`\`\`$file_ext
$file_content
\`\`\`

Requirements:
1. Generate $test_type tests that cover:
   - Happy path scenarios
   - Edge cases and boundary conditions
   - Error handling and validation
   - Integration points (if applicable)

2. Follow these guidelines:
   - Use the project's existing test framework and patterns
   - Include descriptive test names
   - Add comments explaining complex test scenarios
   - Ensure tests are independent and can run in any order
   - Mock external dependencies appropriately

3. Output format:
   - Provide complete, runnable test code
   - Include necessary imports and setup
   - Follow the language's testing conventions

Generate the tests now:"

  # Call Bob Shell (IBM's AI assistant)
  # For Bob Shell usage, see: https://internal.bob.ibm.com/docs/shell/install-and-setup
  local bob_output
  if command -v bob &> /dev/null; then
    # Use Bob Shell with prompt flag
    # Note: Using non-interactive mode, so only non-destructive tools are used by default
    bob_output=$(bob -p "$prompt" 2>/dev/null || echo "")
  else
    # Fallback to direct Anthropic API call if Bob Shell not available
    log_warning "Bob Shell not found, using direct API call"
    log_warning "For better results, install Bob Shell: https://internal.bob.ibm.com/docs/shell/install-and-setup"
    bob_output=$(call_anthropic_api "$prompt")
  fi
  
  if [ -z "$bob_output" ]; then
    log_warning "No output from Bob CLI for $file_path"
    return 1
  fi
  
  # Extract code blocks from Bob's response
  local test_code
  test_code=$(echo "$bob_output" | sed -n '/```/,/```/p' | sed '1d;$d')
  
  if [ -z "$test_code" ]; then
    log_warning "Could not extract test code from Bob's response"
    return 1
  fi
  
  # Determine test file path
  local test_file_path
  test_file_path=$(get_test_file_path "$file_path" "$test_type")
  
  # Save generated test
  mkdir -p "$(dirname "$test_file_path")"
  echo "$test_code" > "$test_file_path"
  
  log_success "Generated $test_type tests: $test_file_path"
  
  return 0
}

call_anthropic_api() {
  local prompt="$1"
  
  local response
  response=$(curl -s -X POST https://api.anthropic.com/v1/messages \
    -H "Content-Type: application/json" \
    -H "x-api-key: $ANTHROPIC_API_KEY" \
    -H "anthropic-version: 2023-06-01" \
    -d "{
      \"model\": \"claude-3-5-sonnet-20241022\",
      \"max_tokens\": 4096,
      \"messages\": [{
        \"role\": \"user\",
        \"content\": $(echo "$prompt" | jq -Rs .)
      }]
    }")
  
  echo "$response" | jq -r '.content[0].text' 2>/dev/null || echo ""
}

get_test_file_path() {
  local source_file="$1"
  local test_type="$2"
  
  local dir
  dir=$(dirname "$source_file")
  local filename
  filename=$(basename "$source_file")
  local name="${filename%.*}"
  local ext="${filename##*.}"
  
  # Determine test file naming convention
  case "$ext" in
    go)
      echo "${dir}/${name}_test.go"
      ;;
    py)
      echo "${dir}/test_${name}.py"
      ;;
    js|ts)
      echo "${dir}/${name}.test.${ext}"
      ;;
    java)
      echo "${dir}/${name}Test.java"
      ;;
    *)
      echo "${dir}/${name}_test.${ext}"
      ;;
  esac
}

###############################################################################
# Coverage Analysis
###############################################################################

analyze_coverage() {
  log_info "Analyzing test coverage..."
  
  local coverage_before=0
  local coverage_after=0
  
  # Run coverage analysis based on project type
  if [ -f "go.mod" ]; then
    coverage_before=$(run_go_coverage "before")
    coverage_after=$(run_go_coverage "after")
  elif [ -f "package.json" ]; then
    coverage_before=$(run_js_coverage "before")
    coverage_after=$(run_js_coverage "after")
  elif [ -f "requirements.txt" ] || [ -f "setup.py" ]; then
    coverage_before=$(run_python_coverage "before")
    coverage_after=$(run_python_coverage "after")
  fi
  
  local coverage_delta
  coverage_delta=$(echo "$coverage_after - $coverage_before" | bc 2>/dev/null || echo "0")
  
  # Create coverage report
  cat > "${GENERATED_TESTS_DIR}/coverage-report.json" <<EOF
{
  "coverage_before": $coverage_before,
  "coverage_after": $coverage_after,
  "coverage_delta": $coverage_delta,
  "coverage_target": $COVERAGE_TARGET,
  "target_met": $([ $(echo "$coverage_after >= $COVERAGE_TARGET" | bc) -eq 1 ] && echo "true" || echo "false"),
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
  
  log_success "Coverage analysis complete"
}

run_go_coverage() {
  go test ./... -coverprofile=coverage.out 2>/dev/null || echo "0"
  go tool cover -func=coverage.out 2>/dev/null | grep total | awk '{print $3}' | sed 's/%//' || echo "0"
}

run_js_coverage() {
  npm test -- --coverage --silent 2>/dev/null | grep "All files" | awk '{print $10}' | sed 's/%//' || echo "0"
}

run_python_coverage() {
  coverage run -m pytest 2>/dev/null && coverage report | grep TOTAL | awk '{print $4}' | sed 's/%//' || echo "0"
}

###############################################################################
# PR Comment Functions
###############################################################################

post_pr_comment() {
  local pr_number="$1"
  local comment_body="$2"
  local repo="${3:-$GITHUB_REPOSITORY}"
  
  if [ -z "$GITHUB_TOKEN" ]; then
    log_warning "No GITHUB_TOKEN set, skipping PR comment"
    return 1
  fi
  
  curl -X POST \
    -H "Authorization: token ${GITHUB_TOKEN}" \
    -H "Accept: application/vnd.github.v3+json" \
    "https://api.github.com/repos/${repo}/issues/${pr_number}/comments" \
    -d "{\"body\": $(echo "$comment_body" | jq -Rs .)}" \
    > /dev/null 2>&1
}

get_current_pr_number() {
  if [ -n "$PULL_REQUEST_NUMBER" ]; then
    echo "$PULL_REQUEST_NUMBER"
    return 0
  fi
  
  # Try to get PR number from branch
  local pr_number
  pr_number=$(curl -s \
    -H "Authorization: token ${GITHUB_TOKEN}" \
    -H "Accept: application/vnd.github.v3+json" \
    "https://api.github.com/repos/${GITHUB_REPOSITORY}/pulls?head=${GITHUB_REPOSITORY%/*}:${GIT_BRANCH}" \
    | jq -r '.[0].number' 2>/dev/null)
  
  if [ "$pr_number" != "null" ] && [ -n "$pr_number" ]; then
    echo "$pr_number"
    return 0
  fi
  
  return 1
}

###############################################################################
# Functional Test PR Creation
###############################################################################

create_functional_test_pr() {
  if [ -z "$FUNCTIONAL_TEST_REPO" ]; then
    log_error "FUNCTIONAL_TEST_REPO not set - functional test PR creation is required"
    return 1
  fi
  
  log_info "Creating PR for functional tests in $FUNCTIONAL_TEST_REPO..."
  
  # Check if we have functional tests generated
  local functional_tests_dir="${GENERATED_TESTS_DIR}/functional"
  if [ ! -d "$functional_tests_dir" ] || [ -z "$(ls -A "$functional_tests_dir")" ]; then
    log_warning "No functional tests generated - skipping functional test PR"
    return 0
  fi
  
  # Clone functional test repo
  local temp_repo="/tmp/functional-tests-$$"
  git clone "https://${GITHUB_TOKEN}@github.com/${FUNCTIONAL_TEST_REPO}.git" "$temp_repo" 2>/dev/null || {
    log_error "Failed to clone functional test repository"
    return 1
  }
  
  cd "$temp_repo"
  
  # Create branch
  local branch_name="ai-tests-$(date +%Y%m%d-%H%M%S)"
  git checkout -b "$branch_name"
  
  # Copy functional tests
  cp -r "$functional_tests_dir"/* .
  
  # Commit and push
  git add .
  git commit -m "Add AI-generated functional tests

Generated by Intelligent Test Orchestrator
Source PR: ${GIT_BRANCH:-unknown}
Timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  
  git push origin "$branch_name"
  
  # Run functional tests and get coverage
  local functional_coverage=0
  if [ -f "go.mod" ]; then
    cd "$temp_repo"
    functional_coverage=$(go test ./... -cover 2>/dev/null | grep "coverage:" | awk '{print $2}' | sed 's/%//' || echo "0")
    cd "$WORK_DIR"
  fi
  
  # Create PR using GitHub API with coverage summary
  local test_count
  test_count=$(find "$functional_tests_dir" -type f -name "*test*" | wc -l)
  
  local pr_body="## 🤖 AI-Generated Functional Tests

This PR contains functional tests automatically generated by the Intelligent Test Orchestrator using **Bob Shell (Claude 3.5 Sonnet)**.

### 📊 Test Coverage Summary

\`\`\`
Functional Test Coverage: ${functional_coverage}%
Test Files Generated: ${test_count}
\`\`\`

### 📝 Source Information

- **Source Repository:** ${GITHUB_REPOSITORY:-unknown}
- **Source Branch:** ${GIT_BRANCH:-unknown}
- **Source Commit:** ${GIT_COMMIT:-unknown}
- **Generated:** $(date -u +%Y-%m-%dT%H:%M:%SZ)

### 🧪 Generated Tests

This PR includes **${test_count}** functional test file(s) covering:
- End-to-end feature behavior
- API integration scenarios
- Service interaction flows
- Error handling and edge cases

### ✅ Review Checklist

- [ ] Review test scenarios for completeness
- [ ] Verify test data and assertions
- [ ] Check for any missing edge cases
- [ ] Ensure tests align with requirements
- [ ] Run tests locally to verify functionality

### 🔗 Related

- Source PR: [Link to source PR if available]

---
*Generated by Intelligent Test Orchestrator with Bob Shell*"

  local pr_response
  pr_response=$(curl -s -X POST \
    -H "Authorization: token ${GITHUB_TOKEN}" \
    -H "Accept: application/vnd.github.v3+json" \
    "https://api.github.com/repos/${FUNCTIONAL_TEST_REPO}/pulls" \
    -d "{
      \"title\": \"🤖 Add AI-generated functional tests from ${GIT_BRANCH}\",
      \"body\": $(echo "$pr_body" | jq -Rs .),
      \"head\": \"$branch_name\",
      \"base\": \"main\"
    }")
  
  local pr_number
  pr_number=$(echo "$pr_response" | jq -r '.number' 2>/dev/null)
  
  cd "$WORK_DIR"
  rm -rf "$temp_repo"
  
  if [ "$pr_number" != "null" ] && [ -n "$pr_number" ]; then
    log_success "Functional test PR #${pr_number} created in $FUNCTIONAL_TEST_REPO"
    echo "$pr_number" > "${GENERATED_TESTS_DIR}/functional-pr-number.txt"
  else
    log_error "Failed to create functional test PR"
    return 1
  fi
}

###############################################################################
# Main Workflow
###############################################################################

main() {
  print_header "🤖 AI Test Generator - Intelligent Test Orchestrator"
  
  # Create working directory
  mkdir -p "$GENERATED_TESTS_DIR"
  
  # Check prerequisites
  check_prerequisites || exit 1
  
  # Install Bob CLI if needed
  install_bob_cli || log_warning "Continuing without Bob CLI (will use API directly)"
  
  # Get PR diff
  get_pr_diff || {
    log_warning "No changes to process"
    exit 0
  }
  
  print_header "📝 Generating Tests"
  
  # Get list of changed files
  local changed_files
  changed_files=$(git diff --name-only "origin/$BASE_BRANCH"...HEAD 2>/dev/null || git diff --name-only HEAD~1 2>/dev/null)
  
  local unit_tests_generated=0
  local functional_tests_generated=0
  
  # Generate tests for each changed file
  while IFS= read -r file; do
    # Skip test files, config files, and documentation
    if [[ "$file" =~ _test\. ]] || [[ "$file" =~ \.test\. ]] || \
       [[ "$file" =~ \.(md|txt|json|yaml|yml)$ ]] || \
       [[ "$file" =~ ^(docs|documentation)/ ]]; then
      continue
    fi
    
    # Generate unit tests
    if generate_tests_with_bob "$file" "unit"; then
      ((unit_tests_generated++))
    fi
    
    # Generate functional tests for API/service files
    if [[ "$file" =~ (api|service|handler|controller) ]]; then
      if generate_tests_with_bob "$file" "functional"; then
        ((functional_tests_generated++))
      fi
    fi
  done <<< "$changed_files"
  
  log_success "Generated $unit_tests_generated unit test(s)"
  log_success "Generated $functional_tests_generated functional test(s)"
  
  # Analyze coverage
  print_header "📊 Coverage Analysis"
  analyze_coverage
  
  # Display coverage report
  if [ -f "${GENERATED_TESTS_DIR}/coverage-report.json" ]; then
    cat "${GENERATED_TESTS_DIR}/coverage-report.json" | jq -r '
      "Before: \(.coverage_before)%",
      "After:  \(.coverage_after)%",
      "Change: \(.coverage_delta)%",
      "",
      "Coverage Target: \(.coverage_target)%",
      "Target Met: \(.target_met)"
    '
  fi
  
  # Commit generated tests to current PR
  if [ "$DRY_RUN" = false ] && [ $unit_tests_generated -gt 0 ]; then
    print_header "💾 Committing Generated Tests"
    
    git add .
    git commit -m "🤖 Add AI-generated unit tests

Generated by Intelligent Test Orchestrator using Bob Shell (Claude 3.5 Sonnet)

📊 Test Generation Summary:
- Unit tests generated: $unit_tests_generated
- Functional tests generated: $functional_tests_generated

[skip ci]" || log_warning "No changes to commit"
    
    git push origin HEAD || log_warning "Failed to push changes"
    
    # Post coverage summary comment to source PR
    print_header "💬 Posting Coverage Summary to PR"
    
    local pr_number
    pr_number=$(get_current_pr_number)
    
    if [ -n "$pr_number" ]; then
      local coverage_before coverage_after coverage_delta target_met
      if [ -f "${GENERATED_TESTS_DIR}/coverage-report.json" ]; then
        coverage_before=$(jq -r '.coverage_before' "${GENERATED_TESTS_DIR}/coverage-report.json")
        coverage_after=$(jq -r '.coverage_after' "${GENERATED_TESTS_DIR}/coverage-report.json")
        coverage_delta=$(jq -r '.coverage_delta' "${GENERATED_TESTS_DIR}/coverage-report.json")
        target_met=$(jq -r '.target_met' "${GENERATED_TESTS_DIR}/coverage-report.json")
      else
        coverage_before="N/A"
        coverage_after="N/A"
        coverage_delta="N/A"
        target_met="false"
      fi
      
      local target_emoji
      if [ "$target_met" = "true" ]; then
        target_emoji="✅"
      else
        target_emoji="⚠️"
      fi
      
      local pr_comment="## 🤖 AI Test Generation Summary

### 📊 Unit Test Coverage

\`\`\`
Coverage Before:  ${coverage_before}%
Coverage After:   ${coverage_after}%
Coverage Delta:   ${coverage_delta}%
Target:           ${COVERAGE_TARGET}%
Target Met:       ${target_emoji} ${target_met}
\`\`\`

### 📝 Generated Tests

- **Unit Tests:** ${unit_tests_generated} file(s)
- **Functional Tests:** ${functional_tests_generated} file(s)

### 🔗 Related PRs

$(if [ $functional_tests_generated -gt 0 ] && [ -f "${GENERATED_TESTS_DIR}/functional-pr-number.txt" ]; then
  local func_pr_num
  func_pr_num=$(cat "${GENERATED_TESTS_DIR}/functional-pr-number.txt")
  echo "- Functional Tests PR: ${FUNCTIONAL_TEST_REPO}#${func_pr_num}"
else
  echo "- No functional tests generated"
fi)

---
*Generated by Intelligent Test Orchestrator with Bob Shell (Claude 3.5 Sonnet)*"

      post_pr_comment "$pr_number" "$pr_comment"
      log_success "Posted coverage summary to PR #${pr_number}"
    else
      log_warning "Could not determine PR number - skipping comment"
    fi
  fi
  
  # Create functional test PR (REQUIRED if functional tests generated)
  if [ $functional_tests_generated -gt 0 ]; then
    print_header "🔀 Creating Functional Test PR (Required)"
    create_functional_test_pr || {
      log_error "Failed to create functional test PR"
      exit 1
    }
  fi
  
  print_header "✅ Test Generation Complete"
  
  log_success "Summary:"
  log_success "  - Unit tests generated: $unit_tests_generated"
  log_success "  - Functional tests generated: $functional_tests_generated"
  log_success "  - Coverage reports: ${GENERATED_TESTS_DIR}/coverage-report.json"
  
  return 0
}

# Run main workflow
main "$@"

# Made with Bob
