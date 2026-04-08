package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tanya-shanker/test-genix/pkg/bobshell"
	"github.com/tanya-shanker/test-genix/pkg/types"
)

// TestGenerator generates tests using AI/LLM capabilities
type TestGenerator struct {
	config      *types.Config
	aiClient    *bobshell.Client
	stats       *types.GenerationStats
	projectRoot string
}

// NewTestGenerator creates a new test generator
func NewTestGenerator(config *types.Config, projectRoot string) *TestGenerator {
	var client *bobshell.Client
	if config.AIAPIKey != "" {
		client = bobshell.NewClient(config.AIAPIKey)
	}

	return &TestGenerator{
		config:      config,
		aiClient:    client,
		projectRoot: projectRoot,
		stats: &types.GenerationStats{
			UnitTestsGenerated:       0,
			FunctionalTestsGenerated: 0,
			TotalTestCases:           0,
			ExecutionTimeSeconds:     0,
			FilesProcessed:           0,
		},
	}
}

// GenerateTests generates tests based on detected changes
func (tg *TestGenerator) GenerateTests(changes *types.ChangeInfo, unitOutputDir, functionalOutputDir string) (*types.GenerationStats, error) {
	startTime := time.Now()

	fmt.Println("🤖 Generating AI-powered tests with Bob (Claude)...")

	// Create output directories
	if err := os.MkdirAll(unitOutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create unit output dir: %w", err)
	}
	if err := os.MkdirAll(functionalOutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create functional output dir: %w", err)
	}

	// Generate unit tests for modified functions
	for _, fn := range changes.ModifiedFunctions {
		if err := tg.generateUnitTestsForFunction(fn, unitOutputDir); err != nil {
			fmt.Printf("⚠️  Failed to generate tests for function %s: %v\n", fn.Name, err)
		}
	}

	// Generate unit tests for modified classes/structs
	for _, cls := range changes.ModifiedClasses {
		if err := tg.generateUnitTestsForClass(cls, unitOutputDir); err != nil {
			fmt.Printf("⚠️  Failed to generate tests for class %s: %v\n", cls.Name, err)
		}
	}

	// Generate functional tests for high-impact semantic changes
	for _, change := range changes.SemanticChanges {
		if change.Impact == "high" || change.Impact == "critical" {
			if err := tg.generateFunctionalTest(change, functionalOutputDir); err != nil {
				fmt.Printf("⚠️  Failed to generate functional test for %s: %v\n", change.Name, err)
			}
		}
	}

	// Generate integration tests for affected modules
	for _, module := range changes.AffectedModules {
		if err := tg.generateIntegrationTests(module, unitOutputDir); err != nil {
			fmt.Printf("⚠️  Failed to generate integration tests for module %s: %v\n", module, err)
		}
	}

	tg.stats.ExecutionTimeSeconds = time.Since(startTime).Seconds()
	tg.stats.FilesProcessed = len(changes.ChangedFiles)

	fmt.Printf("✅ Generated %d unit tests\n", tg.stats.UnitTestsGenerated)
	fmt.Printf("✅ Generated %d functional tests\n", tg.stats.FunctionalTestsGenerated)

	return tg.stats, nil
}

// generateUnitTestsForFunction generates unit tests for a specific function
func (tg *TestGenerator) generateUnitTestsForFunction(fn types.FunctionChange, outputDir string) error {
	framework := tg.config.TestFrameworks[fn.Language]
	if framework == "" {
		framework = "testing" // Default for Go
	}

	// Generate test file name
	testFilename := tg.generateTestFilename(fn.File, fn.Language)
	testFilepath := filepath.Join(outputDir, testFilename)

	// Read source code for context
	sourceCode, err := tg.readSourceFile(fn.File)
	if err != nil {
		sourceCode = "" // Continue without source context
	}

	// Generate test cases
	testCases, err := tg.generateTestCasesForFunction(fn, sourceCode, framework)
	if err != nil {
		return err
	}

	// Write test file with source file path for package detection
	if err := tg.writeTestFileWithPackage(testFilepath, testCases, fn.Language, framework, fn.File); err != nil {
		return err
	}

	tg.stats.UnitTestsGenerated++
	tg.stats.TotalTestCases += len(testCases)

	return nil
}

// generateUnitTestsForClass generates unit tests for a class/struct
func (tg *TestGenerator) generateUnitTestsForClass(cls types.ClassChange, outputDir string) error {
	framework := tg.config.TestFrameworks[cls.Language]
	if framework == "" {
		framework = "testing"
	}

	testFilename := tg.generateTestFilename(cls.File, cls.Language)
	testFilepath := filepath.Join(outputDir, testFilename)

	// Read source code
	sourceCode, _ := tg.readSourceFile(cls.File)

	// Generate test cases for each method
	var allTestCases []types.TestCase
	for _, method := range cls.Methods {
		testCases := tg.generateTestCasesForMethod(cls.Name, method, cls.Language, framework, sourceCode)
		allTestCases = append(allTestCases, testCases...)
	}

	// Write test file with source file path for package detection
	if err := tg.writeTestFileWithPackage(testFilepath, allTestCases, cls.Language, framework, cls.File); err != nil {
		return err
	}

	tg.stats.UnitTestsGenerated++
	tg.stats.TotalTestCases += len(allTestCases)

	return nil
}

// generateFunctionalTest generates functional/E2E test
func (tg *TestGenerator) generateFunctionalTest(change types.SemanticChange, outputDir string) error {
	language := tg.detectLanguageFromFile(change.File)

	testFilename := fmt.Sprintf("test_%s_functional%s", change.Name, tg.getExtension(language))
	testFilepath := filepath.Join(outputDir, testFilename)

	// Generate functional test scenarios
	scenarios := tg.generateFunctionalScenarios(change, language)

	// Write functional test file
	if err := tg.writeFunctionalTestFile(testFilepath, scenarios, language); err != nil {
		return err
	}

	tg.stats.FunctionalTestsGenerated++
	tg.stats.TotalTestCases += len(scenarios)

	return nil
}

// generateIntegrationTests generates integration tests for module interactions
func (tg *TestGenerator) generateIntegrationTests(module string, outputDir string) error {
	testFilename := fmt.Sprintf("test_%s_integration_test.go", module)
	testFilepath := filepath.Join(outputDir, testFilename)

	testCases := tg.generateIntegrationTestCases(module)

	if len(testCases) > 0 {
		if err := tg.writeTestFile(testFilepath, testCases, "go", "testing"); err != nil {
			return err
		}
		tg.stats.UnitTestsGenerated++
		tg.stats.TotalTestCases += len(testCases)
	}

	return nil
}

// generateTestCasesForFunction generates test cases for a function using AI
func (tg *TestGenerator) generateTestCasesForFunction(fn types.FunctionChange, sourceCode, framework string) ([]types.TestCase, error) {
	testCases := []types.TestCase{}

	// Generate happy path test
	testCases = append(testCases, types.TestCase{
		Name:        fmt.Sprintf("Test%s_HappyPath", tg.capitalize(fn.Name)),
		Description: fmt.Sprintf("Test %s with valid inputs", fn.Name),
		Type:        "happy_path",
		Code:        tg.generateHappyPathTest(fn, framework),
	})

	// Generate edge case tests if enabled
	if tg.config.GenerateEdgeCases {
		testCases = append(testCases, types.TestCase{
			Name:        fmt.Sprintf("Test%s_EdgeCases", tg.capitalize(fn.Name)),
			Description: fmt.Sprintf("Test %s with edge case inputs", fn.Name),
			Type:        "edge_case",
			Code:        tg.generateEdgeCaseTest(fn, framework),
		})
	}

	// Generate error handling tests
	testCases = append(testCases, types.TestCase{
		Name:        fmt.Sprintf("Test%s_ErrorHandling", tg.capitalize(fn.Name)),
		Description: fmt.Sprintf("Test %s error handling", fn.Name),
		Type:        "error_handling",
		Code:        tg.generateErrorTest(fn, framework),
	})

	// Use Bob/Claude to enhance tests if API key is available
	if tg.aiClient != nil && sourceCode != "" {
		enhancedTests, err := tg.enhanceTestsWithBob(fn, sourceCode, testCases)
		if err == nil {
			testCases = enhancedTests
		}
	}

	return testCases, nil
}

// generateTestCasesForMethod generates test cases for a class method
func (tg *TestGenerator) generateTestCasesForMethod(className, methodName, language, framework, sourceCode string) []types.TestCase {
	testCases := []types.TestCase{}

	testCases = append(testCases, types.TestCase{
		Name:        fmt.Sprintf("Test%s_%s", className, tg.capitalize(methodName)),
		Description: fmt.Sprintf("Test %s.%s", className, methodName),
		Type:        "method_test",
		Code:        tg.generateMethodTest(className, methodName, language, framework),
	})

	return testCases
}

// generateFunctionalScenarios generates functional test scenarios
func (tg *TestGenerator) generateFunctionalScenarios(change types.SemanticChange, language string) []types.TestCase {
	scenarios := []types.TestCase{}

	// Try to generate AI-powered test if Bob client is available
	if tg.aiClient != nil {
		fmt.Printf("🤖 Generating AI-powered E2E test for %s using Bob/Claude...\n", change.Name)
		aiTest, err := tg.generateE2ETestWithBob(change, language)
		if err == nil && aiTest != "" {
			fmt.Printf("✅ Successfully generated AI-powered E2E test for %s (%d chars)\n", change.Name, len(aiTest))
			scenarios = append(scenarios, types.TestCase{
				Name:        fmt.Sprintf("Test%s_E2E", tg.capitalize(change.Name)),
				Description: fmt.Sprintf("End-to-end test for %s", change.Name),
				Type:        "e2e",
				Code:        aiTest,
			})
			return scenarios
		}
		// Fall back to template if AI generation fails
		fmt.Printf("⚠️  AI test generation failed for %s, using template: %v\n", change.Name, err)
	} else {
		fmt.Printf("⚠️  Bob client not available (API key missing), using template for %s\n", change.Name)
	}

	// Fallback to template-based generation
	scenarios = append(scenarios, types.TestCase{
		Name:        fmt.Sprintf("Test%s_E2E", tg.capitalize(change.Name)),
		Description: fmt.Sprintf("End-to-end test for %s", change.Name),
		Type:        "e2e",
		Code:        tg.generateE2ETest(change, language),
	})

	return scenarios
}

// generateIntegrationTestCases generates integration test cases for module
func (tg *TestGenerator) generateIntegrationTestCases(module string) []types.TestCase {
	testCases := []types.TestCase{}

	testCases = append(testCases, types.TestCase{
		Name:        fmt.Sprintf("Test%s_Integration", tg.capitalize(module)),
		Description: fmt.Sprintf("Integration test for %s module", module),
		Type:        "integration",
		Code:        tg.generateIntegrationTestCode(module),
	})

	return testCases
}

// enhanceTestsWithBob uses Bob (Claude) to enhance generated tests
func (tg *TestGenerator) enhanceTestsWithBob(fn types.FunctionChange, sourceCode string, testCases []types.TestCase) ([]types.TestCase, error) {
	_ = context.Background() // For future use

	prompt := fmt.Sprintf(`Given this function:
%s

Generate comprehensive test cases that cover:
1. Happy path scenarios
2. Edge cases
3. Error handling
4. Boundary conditions

Function name: %s
Language: %s

Provide test code in %s format.`, sourceCode, fn.Name, fn.Language, fn.Language)

	message, err := tg.aiClient.CreateMessage(bobshell.MessageRequest{
		Model:     tg.config.AIModel,
		MaxTokens: 2000,
		Messages: []bobshell.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		System: "You are Bob, an expert software testing engineer. Generate comprehensive, production-ready test cases.",
	})

	if err != nil {
		return testCases, err
	}

	bobGeneratedCode := message.ExtractText()

	if bobGeneratedCode != "" {
		// Append Bob-enhanced test
		testCases = append(testCases, types.TestCase{
			Name:        fmt.Sprintf("Test%s_BobEnhanced", tg.capitalize(fn.Name)),
			Description: fmt.Sprintf("Bob-enhanced comprehensive test for %s", fn.Name),
			Type:        "bob_enhanced",
			Code:        bobGeneratedCode,
		})
	}

	return testCases, nil
}

// generateE2ETestWithBob uses Bob (Claude) to generate E2E tests based on semantic changes
func (tg *TestGenerator) generateE2ETestWithBob(change types.SemanticChange, language string) (string, error) {
	// Read the source file to get context
	sourceCode, err := tg.readSourceFile(change.File)
	if err != nil {
		sourceCode = fmt.Sprintf("// File: %s\n// Unable to read source code", change.File)
	}

	prompt := fmt.Sprintf(`You are Bob, an expert software testing engineer. Generate a comprehensive end-to-end (E2E) functional test for the following code change:

**Change Type:** %s
**Component Name:** %s
**File:** %s
**Impact:** %s

**Source Code Context:**
%s

**Requirements:**
1. Generate a complete, runnable E2E test function in %s
2. The test should cover the full workflow from setup to cleanup
3. Include realistic test data and assertions
4. Test should validate the end-to-end behavior of the feature
5. Include proper error handling and edge cases
6. Use the standard testing framework for %s
7. DO NOT include package declaration or imports - only the test function
8. Make the test production-ready with meaningful assertions

**Output Format:**
Provide ONLY the test function code without any markdown formatting, explanations, or package/import statements.`,
		change.Type, change.Name, change.File, change.Impact, sourceCode, language, language)

	message, err := tg.aiClient.CreateMessage(bobshell.MessageRequest{
		Model:     tg.config.AIModel,
		MaxTokens: 3000,
		Messages: []bobshell.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		System: "You are Bob, an expert software testing engineer specializing in end-to-end test generation. Generate production-ready, comprehensive E2E tests based on code changes.",
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate E2E test with Bob: %w", err)
	}

	generatedCode := message.ExtractText()

	// Clean up the generated code (remove markdown code blocks if present)
	generatedCode = strings.TrimSpace(generatedCode)
	generatedCode = strings.TrimPrefix(generatedCode, "```go")
	generatedCode = strings.TrimPrefix(generatedCode, "```")
	generatedCode = strings.TrimSuffix(generatedCode, "```")
	generatedCode = strings.TrimSpace(generatedCode)

	if generatedCode == "" {
		return "", fmt.Errorf("Bob generated empty test code")
	}

	return generatedCode, nil
}

// Template generation methods for different test types

func (tg *TestGenerator) generateHappyPathTest(fn types.FunctionChange, framework string) string {
	switch fn.Language {
	case "go":
		return fmt.Sprintf(`func %s(t *testing.T) {
	// Arrange
	// TODO: Set up test data
	
	// Act
	result := %s()
	
	// Assert
	if result == nil {
		t.Error("Expected non-nil result")
	}
	// TODO: Add specific assertions
}`, fmt.Sprintf("Test%s_HappyPath", tg.capitalize(fn.Name)), fn.Name)

	case "python":
		return fmt.Sprintf(`def test_%s_happy_path():
    """Test %s with valid inputs"""
    # Arrange
    # TODO: Set up test data
    
    # Act
    result = %s()
    
    # Assert
    assert result is not None
    # TODO: Add specific assertions`, fn.Name, fn.Name, fn.Name)

	case "javascript", "typescript":
		return fmt.Sprintf(`test('%s - happy path', () => {
    // Arrange
    // TODO: Set up test data
    
    // Act
    const result = %s();
    
    // Assert
    expect(result).toBeDefined();
    // TODO: Add specific assertions
});`, fn.Name, fn.Name)
	}
	return ""
}

func (tg *TestGenerator) generateEdgeCaseTest(fn types.FunctionChange, framework string) string {
	switch fn.Language {
	case "go":
		return fmt.Sprintf(`func %s(t *testing.T) {
	// Test with empty input
	// Test with nil values
	// Test with boundary values
	t.Skip("TODO: Implement edge case tests")
}`, fmt.Sprintf("Test%s_EdgeCases", tg.capitalize(fn.Name)))

	case "python":
		return fmt.Sprintf(`def test_%s_edge_cases():
    """Test %s with edge case inputs"""
    # Test with empty input
    # Test with None
    # Test with boundary values
    pass  # TODO: Implement edge case tests`, fn.Name, fn.Name)
	}
	return ""
}

func (tg *TestGenerator) generateErrorTest(fn types.FunctionChange, framework string) string {
	switch fn.Language {
	case "go":
		return fmt.Sprintf(`func %s(t *testing.T) {
	// Test invalid input handling
	// TODO: Add error case tests
	t.Skip("TODO: Implement error handling tests")
}`, fmt.Sprintf("Test%s_ErrorHandling", tg.capitalize(fn.Name)))

	case "python":
		return fmt.Sprintf(`def test_%s_error_handling():
    """Test %s error handling"""
    import pytest
    
    # Test invalid input handling
    with pytest.raises(Exception):
        %s(invalid_input)`, fn.Name, fn.Name, fn.Name)
	}
	return ""
}

func (tg *TestGenerator) generateMethodTest(className, methodName, language, framework string) string {
	switch language {
	case "go":
		return fmt.Sprintf(`func Test%s_%s(t *testing.T) {
	// Arrange
	instance := &%s{}
	
	// Act
	result := instance.%s()
	
	// Assert
	if result == nil {
		t.Error("Expected non-nil result")
	}
}`, className, tg.capitalize(methodName), className, tg.capitalize(methodName))

	case "python":
		return fmt.Sprintf(`def test_%s_%s():
    """Test %s.%s"""
    # Arrange
    instance = %s()
    
    # Act
    result = instance.%s()
    
    # Assert
    assert result is not None`, className, methodName, className, methodName, className, methodName)
	}
	return ""
}

func (tg *TestGenerator) generateE2ETest(change types.SemanticChange, language string) string {
	return fmt.Sprintf(`func Test%s_E2E(t *testing.T) {
	// Setup: Initialize system state
	// Execute: Perform complete workflow
	// Verify: Check end-to-end behavior
	// Cleanup: Reset system state
	t.Skip("TODO: Implement E2E test")
}`, tg.capitalize(change.Name))
}

func (tg *TestGenerator) generateIntegrationTestCode(module string) string {
	return fmt.Sprintf(`func Test%s_Integration(t *testing.T) {
	// Test module interactions
	// Verify data flow between components
	// Check API contracts
	t.Skip("TODO: Implement integration test")
}`, tg.capitalize(module))
}

// File operations

func (tg *TestGenerator) writeTestFileWithPackage(filepath string, testCases []types.TestCase, language, framework, sourceFile string) error {
	if err := os.MkdirAll(filepath[:strings.LastIndex(filepath, string(os.PathSeparator))], 0755); err != nil {
		return err
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write file header with correct package
	header := tg.generateTestFileHeaderWithPackage(language, framework, sourceFile)
	if _, err := file.WriteString(header + "\n\n"); err != nil {
		return err
	}

	// Write test cases
	for _, testCase := range testCases {
		if _, err := file.WriteString(fmt.Sprintf("// %s\n", testCase.Description)); err != nil {
			return err
		}
		if _, err := file.WriteString(testCase.Code + "\n\n"); err != nil {
			return err
		}
	}

	return nil
}

func (tg *TestGenerator) writeTestFile(filepath string, testCases []types.TestCase, language, framework string) error {
	if err := os.MkdirAll(filepath[:strings.LastIndex(filepath, string(os.PathSeparator))], 0755); err != nil {
		return err
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write file header
	header := tg.generateTestFileHeader(language, framework)
	if _, err := file.WriteString(header + "\n\n"); err != nil {
		return err
	}

	// Write test cases
	for _, testCase := range testCases {
		if _, err := file.WriteString(fmt.Sprintf("// %s\n", testCase.Description)); err != nil {
			return err
		}
		if _, err := file.WriteString(testCase.Code + "\n\n"); err != nil {
			return err
		}
	}

	return nil
}

func (tg *TestGenerator) writeFunctionalTestFile(filepath string, scenarios []types.TestCase, language string) error {
	if err := os.MkdirAll(filepath[:strings.LastIndex(filepath, string(os.PathSeparator))], 0755); err != nil {
		return err
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write proper file header based on language
	header := tg.generateFunctionalTestFileHeader(language)
	if _, err := file.WriteString(header + "\n\n"); err != nil {
		return err
	}

	// Write test scenarios
	for _, scenario := range scenarios {
		if _, err := file.WriteString(fmt.Sprintf("// %s\n", scenario.Description)); err != nil {
			return err
		}
		if _, err := file.WriteString(scenario.Code + "\n\n"); err != nil {
			return err
		}
	}

	return nil
}

func (tg *TestGenerator) generateTestFileHeader(language, framework string) string {
	switch language {
	case "go":
		return fmt.Sprintf(`package main

import (
	"testing"
)

// Auto-generated tests by AI Test Orchestrator (Bob)
// Framework: %s`, framework)

	case "python":
		return fmt.Sprintf(`"""
Auto-generated tests by AI Test Orchestrator (Bob)
Framework: %s
"""

import pytest`, framework)

	case "javascript", "typescript":
		return fmt.Sprintf(`/**
 * Auto-generated tests by AI Test Orchestrator (Bob)
 * Framework: %s
 */`, framework)
	}
	return ""
}

func (tg *TestGenerator) generateTestFileHeaderWithPackage(language, framework, sourceFile string) string {
	switch language {
	case "go":
		// Extract package name from source file
		packageName := tg.extractPackageName(sourceFile)
		if packageName == "" {
			packageName = "main"
		}

		return fmt.Sprintf(`package %s

import (
	"testing"
)

// Auto-generated tests by AI Test Orchestrator (Bob)
// Framework: %s`, packageName, framework)

	case "python":
		return fmt.Sprintf(`"""
Auto-generated tests by AI Test Orchestrator (Bob)
Framework: %s
"""

import pytest`, framework)

	case "javascript", "typescript":
		return fmt.Sprintf(`/**
 * Auto-generated tests by AI Test Orchestrator (Bob)
 * Framework: %s
 */`, framework)
	}
	return ""
}

func (tg *TestGenerator) generateFunctionalTestFileHeader(language string) string {
	switch language {
	case "go":
		return `package main

import (
	"testing"
)

// Functional Test - Auto-generated by AI Test Orchestrator (Bob)`

	case "python":
		return `"""
Functional Test - Auto-generated by AI Test Orchestrator (Bob)
"""

import pytest`

	case "javascript", "typescript":
		return `/**
 * Functional Test - Auto-generated by AI Test Orchestrator (Bob)
 */`
	}
	return "// Functional Test - Auto-generated by AI Test Orchestrator (Bob)"
}

// extractPackageName extracts the package name from a Go source file
func (tg *TestGenerator) extractPackageName(sourceFile string) string {
	content, err := tg.readSourceFile(sourceFile)
	if err != nil {
		return ""
	}

	// Look for package declaration
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}

	return ""
}

// Utility methods

func (tg *TestGenerator) generateTestFilename(sourceFile, language string) string {
	base := filepath.Base(sourceFile)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	switch language {
	case "go":
		return name + "_test.go"
	case "python":
		return "test_" + name + ".py"
	default:
		return "test_" + name + tg.getExtension(language)
	}
}

func (tg *TestGenerator) readSourceFile(filepath string) (string, error) {
	fullPath := filepath
	if !strings.HasPrefix(filepath, "/") {
		fullPath = filepath
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func (tg *TestGenerator) detectLanguageFromFile(filepath string) string {
	ext := strings.ToLower(filepath[strings.LastIndex(filepath, "."):])
	langMap := map[string]string{
		".go":   "go",
		".py":   "python",
		".js":   "javascript",
		".ts":   "typescript",
		".java": "java",
	}
	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return "go"
}

func (tg *TestGenerator) getExtension(language string) string {
	extMap := map[string]string{
		"go":         ".go",
		"python":     ".py",
		"javascript": ".js",
		"typescript": ".ts",
		"java":       ".java",
	}
	if ext, ok := extMap[language]; ok {
		return ext
	}
	return ".go"
}

func (tg *TestGenerator) capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// Made with Bob
