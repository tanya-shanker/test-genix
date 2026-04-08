# 🔄 End-to-End Flow: AI Test Generation with Bob Shell

This document explains the complete workflow of the Intelligent Test Orchestrator with a real-world example.

## 🎯 Key Features

- ✅ **Automatic Unit Test Generation** - Tests committed to source PR
- ✅ **Automatic Functional Test Generation** - Separate PR created (**REQUIRED**)
- ✅ **Coverage Summary Comments** - Posted to **BOTH** PRs automatically
- ✅ **Bob Shell Integration** - AI-powered comprehensive test scenarios
- ✅ **PR Linking** - Source PR and Functional Test PR are cross-referenced

## 📋 Table of Contents
1. [Overview](#overview)
2. [Architecture Diagram](#architecture-diagram)
3. [Detailed Flow](#detailed-flow)
4. [Real-World Example](#real-world-example)
5. [Sequence Diagram](#sequence-diagram)
6. [Component Interactions](#component-interactions)

---

## Overview

The Intelligent Test Orchestrator automatically generates tests for every pull request using IBM Bob Shell. It analyzes code changes, generates comprehensive tests, and integrates them seamlessly into your codebase.

### Key Components

```
┌─────────────────────────────────────────────────────────────┐
│                    Developer Workflow                        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  IBM Cloud OnePipeline                       │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  1. Setup Stage                                       │  │
│  │     - Install dependencies                            │  │
│  │     - Install Bob Shell                               │  │
│  └──────────────────────────────────────────────────────┘  │
│                              │                               │
│                              ▼                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  2. Test Stage                                        │  │
│  │     - Run existing tests                              │  │
│  │     - Measure baseline coverage                       │  │
│  └──────────────────────────────────────────────────────┘  │
│                              │                               │
│                              ▼                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  3. AI Test Generation Stage ⭐                       │  │
│  │     - Detect code changes                             │  │
│  │     - Generate tests with Bob Shell                   │  │
│  │     - Integrate tests into codebase                   │  │
│  │     - Measure new coverage                            │  │
│  │     - Post coverage comment to source PR              │  │
│  │     - Create functional test PR (REQUIRED)            │  │
│  └──────────────────────────────────────────────────────┘  │
│                              │                               │
│                              ▼                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  4. Build & Deploy Stages                             │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Pull Request Created                         │
│                    (Developer pushes code changes)                   │
└────────────────────────────────┬────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    OnePipeline Triggered                             │
│                                                                       │
│  ┌────────────────────────────────────────────────────────────┐    │
│  │  ai-test-generator.sh                                       │    │
│  │                                                              │    │
│  │  ┌──────────────────────────────────────────────────────┐  │    │
│  │  │  1. Change Detection                                  │  │    │
│  │  │     • git diff origin/main...HEAD                     │  │    │
│  │  │     • Identify modified files                         │  │    │
│  │  │     • Filter out test files, docs, configs            │  │    │
│  │  └──────────────────────────────────────────────────────┘  │    │
│  │                         │                                   │    │
│  │                         ▼                                   │    │
│  │  ┌──────────────────────────────────────────────────────┐  │    │
│  │  │  2. Bob Shell Installation Check                      │  │    │
│  │  │     • Check if 'bob' command exists                   │  │    │
│  │  │     • If not: Install via curl script                 │  │    │
│  │  │     • Verify BOBSHELL_API_KEY is set                  │  │    │
│  │  └──────────────────────────────────────────────────────┘  │    │
│  │                         │                                   │    │
│  │                         ▼                                   │    │
│  │  ┌──────────────────────────────────────────────────────┐  │    │
│  │  │  3. Test Generation Loop                              │  │    │
│  │  │     For each changed file:                            │  │    │
│  │  │     ┌────────────────────────────────────────────┐    │  │    │
│  │  │     │  a. Read file content                       │    │  │    │
│  │  │     │  b. Detect language (Go/Python/JS/Java)     │    │  │    │
│  │  │     │  c. Create prompt for Bob Shell             │    │  │    │
│  │  │     │  d. Call: bob -p "Generate tests..."        │    │  │    │
│  │  │     │  e. Extract test code from response         │    │  │    │
│  │  │     │  f. Save to appropriate test file           │    │  │    │
│  │  │     └────────────────────────────────────────────┘    │  │    │
│  │  └──────────────────────────────────────────────────────┘  │    │
│  │                         │                                   │    │
│  │                         ▼                                   │    │
│  │  ┌──────────────────────────────────────────────────────┐  │    │
│  │  │  4. Coverage Analysis                                 │  │    │
│  │  │     • Run tests with coverage                         │  │    │
│  │  │     • Compare before/after coverage                   │  │    │
│  │  │     • Generate coverage report JSON                   │  │    │
│  │  └──────────────────────────────────────────────────────┘  │    │
│  │                         │                                   │    │
│  │                         ▼                                   │    │
│  │  ┌──────────────────────────────────────────────────────┐  │    │
│  │  │  5. Commit Generated Tests                            │  │    │
│  │  │     • git add .                                       │  │    │
│  │  │     • git commit -m "Add AI-generated tests"          │  │    │
│  │  │     • git push origin HEAD                            │  │    │
│  │  └──────────────────────────────────────────────────────┘  │    │
│  │                         │                                   │    │
│  │                         ▼                                   │    │
│  │  ┌──────────────────────────────────────────────────────┐  │    │
│  │  │  6. Post Coverage Comment to Source PR                │  │    │
│  │  │     • Format coverage summary                         │  │    │
│  │  │     • Post comment via GitHub API                     │  │    │
│  │  │     • Include link to functional test PR              │  │    │
│  │  └──────────────────────────────────────────────────────┘  │    │
│  │                         │                                   │    │
│  │                         ▼                                   │    │
│  │  ┌──────────────────────────────────────────────────────┐  │    │
│  │  │  7. Functional Test PR (REQUIRED)                     │  │    │
│  │  │     • Clone functional test repo                      │  │    │
│  │  │     • Copy functional tests                           │  │    │
│  │  │     • Create new branch                               │  │    │
│  │  │     • Push and create PR via GitHub API               │  │    │
│  │  │     • Include coverage in PR description              │  │    │
│  │  │     • Link back to source PR                          │  │    │
│  │  │     ⚠️  Pipeline FAILS if this step fails             │  │    │
│  │  └──────────────────────────────────────────────────────┘  │    │
│  └────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    Pipeline Logs Display                             │
│                                                                       │
│  📊 Test Coverage Summary:                                           │
│     Before:  65%                                                     │
│     After:   82%                                                     │
│     Delta:   +17%                                                    │
│                                                                       │
│  📝 Generated Tests:                                                 │
│     Unit tests: 3 files                                              │
│     Functional tests: 1 file                                         │
│                                                                       │
│  ✅ Tests committed to source PR #123                                │
│  ✅ Coverage comment posted to PR #123                               │
│  ✅ Functional test PR created: #456                                 │
│  🔗 PRs linked: #123 ↔ #456                                          │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Detailed Flow

### Phase 1: Setup & Prerequisites

```
Developer                    GitHub                    OnePipeline
    │                           │                           │
    │  1. Create feature branch │                           │
    │──────────────────────────>│                           │
    │                           │                           │
    │  2. Make code changes     │                           │
    │  (e.g., add new API)      │                           │
    │                           │                           │
    │  3. Push to GitHub        │                           │
    │──────────────────────────>│                           │
    │                           │                           │
    │  4. Create Pull Request   │                           │
    │──────────────────────────>│                           │
    │                           │                           │
    │                           │  5. Trigger Pipeline      │
    │                           │──────────────────────────>│
    │                           │                           │
    │                           │                           │  6. Setup Stage
    │                           │                           │     - Install deps
    │                           │                           │     - Install Bob Shell
    │                           │                           │
```

### Phase 2: Change Detection

```
OnePipeline                                          Git Repository
    │                                                       │
    │  1. Fetch base branch (main)                         │
    │──────────────────────────────────────────────────────>│
    │                                                       │
    │  2. Get diff: git diff origin/main...HEAD            │
    │──────────────────────────────────────────────────────>│
    │                                                       │
    │  3. Parse changed files                              │
    │<──────────────────────────────────────────────────────│
    │     • src/api/user_service.go (modified)             │
    │     • src/api/auth_handler.go (new)                  │
    │     • README.md (modified) ← Skip                    │
    │     • config.yaml (modified) ← Skip                  │
    │                                                       │
    │  4. Filter relevant files                            │
    │     ✓ src/api/user_service.go                        │
    │     ✓ src/api/auth_handler.go                        │
    │                                                       │
```

### Phase 3: Test Generation with Bob Shell

```
ai-test-generator.sh              Bob Shell API              File System
        │                              │                          │
        │  1. Read file content        │                          │
        │─────────────────────────────────────────────────────────>│
        │  (src/api/user_service.go)   │                          │
        │                              │                          │
        │  2. Create prompt            │                          │
        │     "Generate unit tests     │                          │
        │      for this Go code..."    │                          │
        │                              │                          │
        │  3. Call Bob Shell           │                          │
        │─────────────────────────────>│                          │
        │     bob -p "prompt"          │                          │
        │                              │                          │
        │                              │  4. AI Processing        │
        │                              │     - Analyze code       │
        │                              │     - Generate tests     │
        │                              │     - Include edge cases │
        │                              │                          │
        │  5. Receive test code        │                          │
        │<─────────────────────────────│                          │
        │     ```go                    │                          │
        │     func TestUserService...  │                          │
        │     ```                      │                          │
        │                              │                          │
        │  6. Extract & save test      │                          │
        │─────────────────────────────────────────────────────────>│
        │     src/api/user_service_test.go                        │
        │                              │                          │
```

### Phase 4: Coverage Analysis & Commit

```
ai-test-generator.sh         Test Runner              Git Repository
        │                         │                          │
        │  1. Run tests            │                          │
        │─────────────────────────>│                          │
        │     go test -cover       │                          │
        │                          │                          │
        │  2. Get coverage         │                          │
        │<─────────────────────────│                          │
        │     82% (was 65%)        │                          │
        │                          │                          │
        │  3. Generate report      │                          │
        │     coverage-report.json │                          │
        │                          │                          │
        │  4. Stage changes        │                          │
        │─────────────────────────────────────────────────────>│
        │     git add .            │                          │
        │                          │                          │
        │  5. Commit tests         │                          │
        │─────────────────────────────────────────────────────>│
        │     "🤖 Add AI-generated │                          │
        │      unit tests [skip ci]"                          │
        │                          │                          │
        │  6. Push to PR branch    │                          │
        │─────────────────────────────────────────────────────>│
        │                          │                          │
```

### Phase 5: Post Coverage Comment to Source PR

```
ai-test-generator.sh              GitHub API              Source PR #123
        │                              │                          │
        │  1. Format coverage summary  │                          │
        │     - Before: 65%            │                          │
        │     - After: 78%             │                          │
        │     - Delta: +13%            │                          │
        │     - Target: 80%            │                          │
        │                              │                          │
        │  2. Post PR comment          │                          │
        │─────────────────────────────>│                          │
        │     POST /repos/.../issues/  │                          │
        │          123/comments        │                          │
        │                              │                          │
        │                              │  3. Comment appears      │
        │                              │─────────────────────────>│
        │                              │  "🤖 AI Test Generation  │
        │                              │   Summary..."            │
        │                              │                          │
        │  4. Confirmation             │                          │
        │<─────────────────────────────│                          │
        │     Comment ID: 789          │                          │
        │                              │                          │
```

### Phase 6: Create Functional Test PR (REQUIRED)

```
ai-test-generator.sh         Functional Repo          GitHub API
        │                         │                          │
        │  1. Clone func repo      │                          │
        │─────────────────────────>│                          │
        │                          │                          │
        │  2. Create branch        │                          │
        │─────────────────────────>│                          │
        │     feature/auth-tests   │                          │
        │                          │                          │
        │  3. Copy func tests      │                          │
        │─────────────────────────>│                          │
        │                          │                          │
        │  4. Commit & push        │                          │
        │─────────────────────────>│                          │
        │                          │                          │
        │  5. Create PR            │                          │
        │─────────────────────────────────────────────────────>│
        │     POST /repos/.../pulls│                          │
        │                          │                          │
        │  6. PR created           │                          │
        │<─────────────────────────────────────────────────────│
        │     PR #456              │                          │
        │                          │                          │
        │  7. Update source PR     │                          │
        │─────────────────────────────────────────────────────>│
        │     Add comment with     │                          │
        │     link to PR #456      │                          │
        │                          │                          │
        │  ⚠️  If this fails,      │                          │
        │     pipeline FAILS       │                          │
        │                          │                          │
```

---

## Real-World Example

### Scenario: Adding a New User Authentication API

#### Step 1: Developer Creates Feature

**File: `src/api/auth_handler.go`** (New File)

```go
package api

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

type AuthHandler struct {
    userService *UserService
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    user, err := h.userService.Authenticate(req.Email, req.Password)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }
    
    token, err := generateToken(user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}
```

#### Step 2: Developer Pushes Code

```bash
git checkout -b feature/user-authentication
git add src/api/auth_handler.go
git commit -m "Add user authentication endpoint"
git push origin feature/user-authentication
```

#### Step 3: Create Pull Request

```
PR #123: Add user authentication endpoint
Branch: feature/user-authentication → main
Files changed: 1 file (+45 lines)
```

#### Step 4: OnePipeline Triggers

**Pipeline Log Output:**

```
==================================================
🤖 AI Test Generation with Bob Shell
==================================================

📊 PR Information:
   Repository: myorg/user-service
   Branch: feature/user-authentication
   Commit: abc123def
   Base Branch: main

✅ API key found - proceeding with test generation
🚀 Starting AI-powered test generation...

📝 Detected changes in 1 file(s):
   • src/api/auth_handler.go

🔍 Analyzing: src/api/auth_handler.go
   Language: Go
   Type: API Handler

🤖 Generating unit tests with Bob Shell...
   Prompt: "Generate comprehensive unit tests for this Go authentication handler..."
   
✅ Generated unit tests: src/api/auth_handler_test.go

==================================================
📈 Test Coverage Summary
==================================================

📊 Coverage Metrics:
  Before:  65%
  After:   78%
  Delta:   +13%

🎯 Target Analysis:
  Target:  80%
  Met:     false (2% short)

==================================================
📋 Test Generation Details
==================================================

📝 Generated Test Files: 1
   • src/api/auth_handler_test.go (8 test cases)

Test Cases Generated:
   ✓ TestAuthHandler_Login_Success
   ✓ TestAuthHandler_Login_InvalidJSON
   ✓ TestAuthHandler_Login_InvalidCredentials
   ✓ TestAuthHandler_Login_UserNotFound
   ✓ TestAuthHandler_Login_TokenGenerationFailed
   ✓ TestAuthHandler_Login_EmptyEmail
   ✓ TestAuthHandler_Login_EmptyPassword
   ✓ TestAuthHandler_Login_SQLInjectionAttempt

==================================================
💬 Posting Coverage Summary to PR
==================================================

📝 Posting comment to PR #123...
✅ Coverage summary comment posted

==================================================
🔗 Creating Functional Test PR
==================================================

📝 Generating functional tests...
✅ Functional tests generated: 1 file(s)

🔄 Creating PR in functional-tests repository...
✅ Functional test PR created: #456
🔗 Link added to source PR #123

==================================================
✅ AI Test Generation Complete
==================================================

✅ Tests committed to source PR #123
💬 Coverage comment posted to PR #123
✅ Functional test PR created: #456
🔗 PRs linked: #123 ↔ #456
📦 Generated tests archived: generated-tests.tar.gz
```

#### Step 5: Generated Test File

**File: `src/api/auth_handler_test.go`** (Auto-generated by Bob Shell)

```go
package api

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of UserService
type MockUserService struct {
    mock.Mock
}

func (m *MockUserService) Authenticate(email, password string) (*User, error) {
    args := m.Called(email, password)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

func TestAuthHandler_Login_Success(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    mockService := new(MockUserService)
    handler := &AuthHandler{userService: mockService}
    
    expectedUser := &User{
        ID:    1,
        Email: "test@example.com",
        Name:  "Test User",
    }
    
    mockService.On("Authenticate", "test@example.com", "password123").
        Return(expectedUser, nil)
    
    // Create request
    reqBody := LoginRequest{
        Email:    "test@example.com",
        Password: "password123",
    }
    jsonBody, _ := json.Marshal(reqBody)
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
    c.Request.Header.Set("Content-Type", "application/json")
    
    // Execute
    handler.Login(c)
    
    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    
    assert.NotEmpty(t, response["token"])
    assert.NotNil(t, response["user"])
    
    mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    handler := &AuthHandler{userService: new(MockUserService)}
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("POST", "/login", bytes.NewBufferString("invalid json"))
    c.Request.Header.Set("Content-Type", "application/json")
    
    // Execute
    handler.Login(c)
    
    // Assert
    assert.Equal(t, http.StatusBadRequest, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Contains(t, response["error"], "invalid")
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    mockService := new(MockUserService)
    handler := &AuthHandler{userService: mockService}
    
    mockService.On("Authenticate", "test@example.com", "wrongpassword").
        Return(nil, errors.New("invalid credentials"))
    
    reqBody := LoginRequest{
        Email:    "test@example.com",
        Password: "wrongpassword",
    }
    jsonBody, _ := json.Marshal(reqBody)
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
    c.Request.Header.Set("Content-Type", "application/json")
    
    // Execute
    handler.Login(c)
    
    // Assert
    assert.Equal(t, http.StatusUnauthorized, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "Invalid credentials", response["error"])
    
    mockService.AssertExpectations(t)
}

// ... 5 more test cases ...
```

#### Step 6: Developer Reviews PR

**PR #123 Updates:**

```
Files changed: 2 files
  • src/api/auth_handler.go (+45 lines)
  • src/api/auth_handler_test.go (+156 lines) ← Auto-generated

Commits:
  1. Add user authentication endpoint (by developer)
  2. 🤖 Add AI-generated unit tests [skip ci] (by AI Test Orchestrator)

Coverage: 65% → 78% (+13%)

Checks:
  ✅ Build passing
  ✅ Tests passing (8/8)
  ⚠️  Coverage target not met (78% < 80%)
```

**Automated PR Comment (Posted by AI Test Orchestrator):**

```markdown
## 🤖 AI Test Generation Summary

### 📊 Unit Test Coverage

```
Coverage Before:  65%
Coverage After:   78%
Coverage Delta:   +13%
Target:           80%
Target Met:       ⚠️ false
```

### 📝 Generated Tests

- **Unit Tests:** 1 file(s)
- **Functional Tests:** 1 file(s)

### 🔗 Related PRs

- Functional Tests PR: myorg/functional-tests#456

---
*Generated by Intelligent Test Orchestrator with Bob Shell (Claude 3.5 Sonnet)*
```

#### Step 7: Functional Test PR Created (REQUIRED)

**PR #456 in `myorg/functional-tests` repository:**

```markdown
## 🤖 AI-Generated Functional Tests

This PR contains functional tests automatically generated by the Intelligent Test Orchestrator using **Bob Shell (Claude 3.5 Sonnet)**.

### 📊 Test Coverage Summary

```
Functional Test Coverage: 85%
Test Files Generated: 1
```

### 📝 Source Information

- **Source Repository:** myorg/user-service
- **Source Branch:** feature/user-authentication
- **Source Commit:** abc123def
- **Generated:** 2024-01-15T10:30:00Z

### 🧪 Generated Tests

This PR includes **1** functional test file(s) covering:
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

- Source PR: myorg/user-service#123

---
*Generated by Intelligent Test Orchestrator with Bob Shell*
```

**Files in Functional Test PR:**

```
tests/
└── user_service/
    └── auth_handler_functional_test.go
```

**Example Functional Test Content:**

```go
package functional_tests

import (
    "testing"
    "net/http"
    "bytes"
    "encoding/json"
)

func TestUserAuthentication_EndToEnd(t *testing.T) {
    // Setup test server
    baseURL := "http://localhost:8080"
    
    // Test 1: Successful login flow
    t.Run("Successful Login Flow", func(t *testing.T) {
        loginReq := map[string]string{
            "email": "test@example.com",
            "password": "password123",
        }
        
        body, _ := json.Marshal(loginReq)
        resp, err := http.Post(baseURL+"/api/login", "application/json", bytes.NewBuffer(body))
        
        if err != nil {
            t.Fatalf("Login request failed: %v", err)
        }
        defer resp.Body.Close()
        
        if resp.StatusCode != http.StatusOK {
            t.Errorf("Expected status 200, got %d", resp.StatusCode)
        }
        
        var result map[string]interface{}
        json.NewDecoder(resp.Body).Decode(&result)
        
        if result["token"] == nil {
            t.Error("Expected token in response")
        }
    })
    
    // Test 2: Invalid credentials
    t.Run("Invalid Credentials", func(t *testing.T) {
        loginReq := map[string]string{
            "email": "test@example.com",
            "password": "wrongpassword",
        }
        
        body, _ := json.Marshal(loginReq)
        resp, err := http.Post(baseURL+"/api/login", "application/json", bytes.NewBuffer(body))
        
        if err != nil {
            t.Fatalf("Login request failed: %v", err)
        }
        defer resp.Body.Close()
        
        if resp.StatusCode != http.StatusUnauthorized {
            t.Errorf("Expected status 401, got %d", resp.StatusCode)
        }
    })
    
    // ... more functional test scenarios ...
}
```

---

## Sequence Diagram

```
Developer    GitHub    OnePipeline    Bob Shell    Test Files    Coverage
    │           │            │             │            │            │
    │  Push     │            │             │            │            │
    │──────────>│            │             │            │            │
    │           │            │             │            │            │
    │  Create PR│            │             │            │            │
    │──────────>│            │             │            │            │
    │           │            │             │            │            │
    │           │  Trigger   │             │            │            │
    │           │───────────>│             │            │            │
    │           │            │             │            │            │
    │           │            │  Setup      │            │            │
    │           │            │  Install    │            │            │
    │           │            │  Bob Shell  │            │            │
    │           │            │             │            │            │
    │           │            │  Get Diff   │            │            │
    │           │<───────────│             │            │            │
    │           │            │             │            │            │
    │           │            │  Read Files │            │            │
    │           │            │─────────────────────────>│            │
    │           │            │             │            │            │
    │           │            │  Generate   │            │            │
    │           │            │  Tests      │            │            │
    │           │            │────────────>│            │            │
    │           │            │             │            │            │
    │           │            │             │  AI Magic  │            │
    │           │            │             │  ⚡⚡⚡     │            │
    │           │            │             │            │            │
    │           │            │  Test Code  │            │            │
    │           │            │<────────────│            │            │
    │           │            │             │            │            │
    │           │            │  Write Tests│            │            │
    │           │            │─────────────────────────>│            │
    │           │            │             │            │            │
    │           │            │  Run Tests  │            │            │
    │           │            │─────────────────────────────────────>│
    │           │            │             │            │            │
    │           │            │  Coverage   │            │            │
    │           │            │<────────────────────────────────────│
    │           │            │  78%        │            │            │
    │           │            │             │            │            │
    │           │            │  Commit     │            │            │
    │           │<───────────│             │            │            │
    │           │            │             │            │            │
    │           │            │  Post PR    │            │            │
    │           │            │  Comment    │            │            │
    │           │<───────────│             │            │            │
    │           │  Coverage  │             │            │            │
    │           │  Summary   │             │            │            │
    │           │            │             │            │            │
    │           │            │  Create     │            │            │
    │           │            │  Functional │            │            │
    │           │            │  Test PR    │            │            │
    │           │<───────────│             │            │            │
    │           │  PR #456   │             │            │            │
    │           │            │             │            │            │
    │  Notify   │            │             │            │            │
    │<──────────│            │             │            │            │
    │  "Tests   │            │             │            │            │
    │  Added +  │            │             │            │            │
    │  Func PR" │            │             │            │            │
    │           │            │             │            │            │
```

---

## Component Interactions

### 1. Change Detection Component

```
Input:  Git diff between base and head branches
Output: List of modified source files

┌─────────────────────────────────────────┐
│  Change Detector                         │
│                                          │
│  1. Fetch git diff                       │
│  2. Parse changed files                  │
│  3. Filter by file type:                 │
│     ✓ .go, .py, .js, .ts, .java         │
│     ✗ _test.*, test_*, *.test.*         │
│     ✗ .md, .txt, .yaml, .json           │
│  4. Return source files only             │
└─────────────────────────────────────────┘
```

### 2. Bob Shell Integration Component

```
Input:  Source file content + language
Output: Generated test code

┌─────────────────────────────────────────┐
│  Bob Shell Wrapper                       │
│                                          │
│  1. Construct prompt:                    │
│     "Generate {type} tests for this      │
│      {language} code. Include:           │
│      - Happy path scenarios              │
│      - Edge cases                        │
│      - Error handling                    │
│      - Boundary conditions"              │
│                                          │
│  2. Call Bob Shell:                      │
│     bob -p "prompt"                      │
│                                          │
│  3. Parse response:                      │
│     Extract code from markdown blocks    │
│                                          │
│  4. Validate test code:                  │
│     Check syntax, imports, structure     │
└─────────────────────────────────────────┘
```

### 3. Test Integration Component

```
Input:  Generated test code + target file
Output: Test file written to disk

┌─────────────────────────────────────────┐
│  Test Integrator                         │
│                                          │
│  1. Determine test file path:            │
│     Go:     {file}_test.go               │
│     Python: test_{file}.py               │
│     JS/TS:  {file}.test.{ext}            │
│     Java:   {File}Test.java              │
│                                          │
│  2. Check for existing tests:            │
│     If exists: Append new tests          │
│     If not: Create new file              │
│                                          │
│  3. Maintain consistency:                │
│     - Match import style                 │
│     - Follow naming conventions          │
│     - Use same test framework            │
│                                          │
│  4. Write test file                      │
└─────────────────────────────────────────┘
```

### 4. Coverage Analysis Component

```
Input:  Test suite
Output: Coverage metrics + report

┌─────────────────────────────────────────┐
│  Coverage Analyzer                       │
│                                          │
│  1. Run tests with coverage:             │
│     Go:     go test -cover               │
│     Python: pytest --cov                 │
│     JS:     npm test -- --coverage       │
│     Java:   mvn test jacoco:report       │
│                                          │
│  2. Parse coverage output                │
│                                          │
│  3. Calculate metrics:                   │
│     - Total coverage %                   │
│     - Coverage delta                     │
│     - Lines covered/total                │
│     - Target comparison                  │
│                                          │
│  4. Generate JSON report                 │
└─────────────────────────────────────────┘
```

---

## Summary

### What Happens Automatically

1. ✅ **Detects** all code changes in your PR
2. ✅ **Generates** comprehensive tests using Bob Shell
3. ✅ **Integrates** unit tests into your codebase
4. ✅ **Measures** coverage improvement
5. ✅ **Commits** unit tests back to your PR
6. ✅ **Posts** coverage summary comment to your PR
7. ✅ **Creates** functional test PR (REQUIRED)
8. ✅ **Includes** coverage summary in functional test PR
9. ✅ **Reports** results in pipeline logs

### What You Need to Do

1. 📝 Write your code
2. 🔀 Create a pull request
3. 👀 Review the generated unit tests in your PR
4. 👀 Review the functional test PR (separate)
5. ✅ Merge both PRs when satisfied

### Time Saved

- **Manual test writing**: 2-4 hours per feature
- **With AI Test Orchestrator**: 5-10 minutes (automated)
- **Time saved**: ~95% reduction in test writing time

---

**Made with Bob Shell** 🤖