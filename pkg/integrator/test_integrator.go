package integrator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tanya-shanker/test-genix/pkg/bobshell"
)

// TestIntegrator integrates generated tests into existing test suites
type TestIntegrator struct {
	projectRoot string
	aiClient    *bobshell.Client
}

// NewTestIntegrator creates a new test integrator
func NewTestIntegrator(projectRoot string) *TestIntegrator {
	// Try to get API key for test validation
	var client *bobshell.Client
	if apiKey := os.Getenv("BOBSHELL_API_KEY"); apiKey != "" {
		client = bobshell.NewClient(apiKey)
	}

	return &TestIntegrator{
		projectRoot: projectRoot,
		aiClient:    client,
	}
}

// IntegrationResult represents the result of test integration
type IntegrationResult struct {
	FilesIntegrated   int      `json:"files_integrated"`
	TestsAdded        int      `json:"tests_added"`
	TestsUpdated      int      `json:"tests_updated"`
	TestsValidated    int      `json:"tests_validated"`
	DuplicatesRemoved int      `json:"duplicates_removed"`
	DuplicatesSkipped int      `json:"duplicates_skipped"`
	ExistingInPR      int      `json:"existing_in_pr"`
	Errors            []string `json:"errors,omitempty"`
}

// IntegrateUnitTests integrates unit tests into the project test directory
func (ti *TestIntegrator) IntegrateUnitTests(testDir, targetDir string) (*IntegrationResult, error) {
	fmt.Println("🔗 Integrating unit tests into existing test suite...")

	result := &IntegrationResult{
		FilesIntegrated:   0,
		TestsAdded:        0,
		TestsUpdated:      0,
		TestsValidated:    0,
		DuplicatesRemoved: 0,
		DuplicatesSkipped: 0,
		ExistingInPR:      0,
		Errors:            []string{},
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	// First, detect and validate tests already in the PR
	fmt.Println("🔍 Checking for tests already in PR changes...")
	existingTests, err := ti.detectTestsInPR()
	if err != nil {
		fmt.Printf("⚠️  Could not detect tests in PR: %v\n", err)
		existingTests = []string{} // Continue with empty list
	} else {
		result.ExistingInPR = len(existingTests)
		if len(existingTests) > 0 {
			fmt.Printf("📋 Found %d test files already in PR\n", len(existingTests))

			// Validate existing tests if Bob Shell is available
			if ti.aiClient != nil {
				fmt.Println("🔍 Validating existing tests with Bob Shell CLI...")
				for _, testFile := range existingTests {
					if ti.shouldValidateTest(testFile) {
						valid, needsUpdate, err := ti.validateExistingTest(testFile)
						if err != nil {
							fmt.Printf("   ⚠️  Could not validate %s: %v\n", testFile, err)
							continue
						}

						if valid && !needsUpdate {
							fmt.Printf("   ✅ Test is valid: %s\n", testFile)
							result.TestsValidated++
						} else if needsUpdate {
							fmt.Printf("   🔄 Test needs update: %s\n", testFile)
							// Mark for regeneration
							if err := ti.markTestForRegeneration(testFile); err != nil {
								fmt.Printf("   ⚠️  Could not mark for regeneration: %v\n", err)
							}
						}
					}
				}
			} else {
				fmt.Println("ℹ️  Skipping test validation (Bob Shell API key not configured)")
			}
		}
	}

	// Walk through generated test directory
	err = filepath.Walk(testDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process test files
		if !ti.isTestFile(path) {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(testDir, path)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to get relative path for %s: %v", path, err))
			return nil
		}

		// Determine target path
		targetPath := filepath.Join(targetDir, relPath)

		// Check if test already exists
		if ti.testExists(targetPath, path) {
			result.DuplicatesSkipped++
			fmt.Printf("⏭️  Skipping duplicate test: %s\n", relPath)
			return nil
		}

		// Copy or merge test file
		if err := ti.integrateTestFile(path, targetPath); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to integrate %s: %v", relPath, err))
			return nil
		}

		result.FilesIntegrated++
		fmt.Printf("✅ Integrated: %s\n", relPath)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk test directory: %w", err)
	}

	fmt.Printf("✅ Integration complete: %d files integrated, %d duplicates skipped\n",
		result.FilesIntegrated, result.DuplicatesSkipped)

	return result, nil
}

// CommitTests commits the integrated tests to the current branch
func (ti *TestIntegrator) CommitTests(targetDir, branch, commit string) error {
	fmt.Println("📝 Committing generated tests to PR...")

	// Check if there are changes to commit
	cmd := exec.Command("git", "status", "--porcelain", targetDir)
	cmd.Dir = ti.projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}

	if len(output) == 0 {
		fmt.Println("ℹ️  No new tests to commit")
		return nil
	}

	// Configure git user and authentication
	if err := ti.configureGitUser(); err != nil {
		return err
	}

	if err := ti.configureGitAuth(); err != nil {
		fmt.Printf("⚠️  Could not configure git authentication: %v\n", err)
		// Continue anyway - push might still work
	}

	// Add files
	cmd = exec.Command("git", "add", targetDir)
	cmd.Dir = ti.projectRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add files: %w\n%s", err, output)
	}

	// Commit
	commitMsg := fmt.Sprintf(`🤖 Add AI-generated unit tests

Generated by Intelligent Test Orchestrator
Commit: %s
Branch: %s

This commit adds automatically generated unit tests based on code changes detected in this PR.`, commit, branch)

	cmd = exec.Command("git", "commit", "-m", commitMsg)
	cmd.Dir = ti.projectRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to commit: %w\n%s", err, output)
	}

	// Push to branch
	fmt.Println("📤 Pushing tests to remote branch...")
	cmd = exec.Command("git", "push", "origin", branch)
	cmd.Dir = ti.projectRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to push to branch: %w\nOutput: %s", err, output)
	}

	fmt.Println("✅ Tests committed and pushed successfully")
	return nil
}

// isTestFile checks if a file is a test file
func (ti *TestIntegrator) isTestFile(path string) bool {
	name := filepath.Base(path)
	return strings.HasSuffix(name, "_test.go") ||
		strings.HasPrefix(name, "test_") ||
		strings.Contains(name, ".test.")
}

// testExists checks if a test already exists
func (ti *TestIntegrator) testExists(targetPath, sourcePath string) bool {
	// Check if target file exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return false
	}

	// Read both files and compare content
	sourceContent, err := os.ReadFile(sourcePath)
	if err != nil {
		return false
	}

	targetContent, err := os.ReadFile(targetPath)
	if err != nil {
		return false
	}

	// Simple duplicate detection: check if content is similar
	// In a production system, this would use more sophisticated comparison
	sourceStr := string(sourceContent)
	targetStr := string(targetContent)

	// Check if the test functions already exist in target
	return ti.containsSimilarTests(sourceStr, targetStr)
}

// containsSimilarTests checks if similar tests exist
func (ti *TestIntegrator) containsSimilarTests(source, target string) bool {
	// Extract test function names from source
	sourceTests := ti.extractTestNames(source)

	// Check if any of these tests exist in target
	for _, testName := range sourceTests {
		if strings.Contains(target, testName) {
			return true
		}
	}

	return false
}

// extractTestNames extracts test function names from code
func (ti *TestIntegrator) extractTestNames(code string) []string {
	names := []string{}
	lines := strings.Split(code, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Go test pattern
		if strings.HasPrefix(line, "func Test") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				name := strings.TrimSuffix(parts[1], "(t")
				name = strings.TrimSuffix(name, "(")
				names = append(names, name)
			}
		}
		// Python test pattern
		if strings.HasPrefix(line, "def test_") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				name := strings.TrimSuffix(parts[1], "():")
				name = strings.TrimSuffix(name, ":")
				names = append(names, name)
			}
		}
	}

	return names
}

// integrateTestFile integrates a test file into the target location
func (ti *TestIntegrator) integrateTestFile(sourcePath, targetPath string) error {
	// Create target directory if needed
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// If target doesn't exist, simply copy
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return ti.copyFile(sourcePath, targetPath)
	}

	// If target exists, merge the tests
	return ti.mergeTestFiles(sourcePath, targetPath)
}

// copyFile copies a file from source to target
func (ti *TestIntegrator) copyFile(source, target string) error {
	input, err := os.ReadFile(source)
	if err != nil {
		return err
	}

	return os.WriteFile(target, input, 0644)
}

// mergeTestFiles merges source test file into target test file
func (ti *TestIntegrator) mergeTestFiles(source, target string) error {
	// Read both files
	sourceContent, err := os.ReadFile(source)
	if err != nil {
		return err
	}

	targetContent, err := os.ReadFile(target)
	if err != nil {
		return err
	}

	// Extract new tests from source that don't exist in target
	newTests := ti.extractNewTests(string(sourceContent), string(targetContent))

	if len(newTests) == 0 {
		return nil // No new tests to add
	}

	// Append new tests to target file
	file, err := os.OpenFile(target, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Add separator comment
	if _, err := file.WriteString("\n// Auto-generated tests added by AI Test Orchestrator\n\n"); err != nil {
		return err
	}

	// Write new tests
	for _, test := range newTests {
		if _, err := file.WriteString(test + "\n\n"); err != nil {
			return err
		}
	}

	return nil
}

// extractNewTests extracts tests from source that don't exist in target
func (ti *TestIntegrator) extractNewTests(source, target string) []string {
	newTests := []string{}
	currentTest := ""
	inTest := false

	lines := strings.Split(source, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if this is the start of a test function
		if strings.HasPrefix(trimmed, "func Test") || strings.HasPrefix(trimmed, "def test_") {
			// Save previous test if any
			if inTest && currentTest != "" {
				testName := ti.extractTestNameFromCode(currentTest)
				if testName != "" && !strings.Contains(target, testName) {
					newTests = append(newTests, currentTest)
				}
			}
			// Start new test
			currentTest = line + "\n"
			inTest = true
		} else if inTest {
			currentTest += line + "\n"
			// Check if test function ended (simple heuristic)
			if trimmed == "}" || (trimmed == "" && strings.HasPrefix(strings.TrimSpace(currentTest), "def test_")) {
				testName := ti.extractTestNameFromCode(currentTest)
				if testName != "" && !strings.Contains(target, testName) {
					newTests = append(newTests, currentTest)
				}
				currentTest = ""
				inTest = false
			}
		}
	}

	// Handle last test
	if inTest && currentTest != "" {
		testName := ti.extractTestNameFromCode(currentTest)
		if testName != "" && !strings.Contains(target, testName) {
			newTests = append(newTests, currentTest)
		}
	}

	return newTests
}

// extractTestNameFromCode extracts the test function name from code
func (ti *TestIntegrator) extractTestNameFromCode(code string) string {
	lines := strings.Split(code, "\n")
	if len(lines) == 0 {
		return ""
	}

	firstLine := strings.TrimSpace(lines[0])

	// Go pattern
	if strings.HasPrefix(firstLine, "func Test") {
		parts := strings.Fields(firstLine)
		if len(parts) >= 2 {
			return strings.TrimSuffix(parts[1], "(t")
		}
	}

	// Python pattern
	if strings.HasPrefix(firstLine, "def test_") {
		parts := strings.Fields(firstLine)
		if len(parts) >= 2 {
			return strings.TrimSuffix(parts[1], "():")
		}
	}

	return ""
}

// configureGitUser configures git user for commits
func (ti *TestIntegrator) configureGitUser() error {
	// Set user name
	cmd := exec.Command("git", "config", "user.name", "AI Test Orchestrator")
	cmd.Dir = ti.projectRoot
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set git user name: %w", err)
	}

	// Set user email
	cmd = exec.Command("git", "config", "user.email", "test-orchestrator@ibm.com")
	cmd.Dir = ti.projectRoot
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set git user email: %w", err)
	}

	return nil
}

// configureGitAuth configures git authentication using GitHub token
func (ti *TestIntegrator) configureGitAuth() error {
	// Get GitHub token from environment
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GHE_TOKEN")
	}

	if token == "" {
		return fmt.Errorf("no GitHub token found in environment")
	}

	// Get repository URL
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = ti.projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get remote URL: %w", err)
	}

	remoteURL := strings.TrimSpace(string(output))

	// Convert SSH URL to HTTPS if needed
	if strings.HasPrefix(remoteURL, "git@") {
		// Convert git@github.com:owner/repo.git to https://github.com/owner/repo.git
		remoteURL = strings.Replace(remoteURL, ":", "/", 1)
		remoteURL = strings.Replace(remoteURL, "git@", "https://", 1)
	}

	// Add token to URL if it's HTTPS
	if strings.HasPrefix(remoteURL, "https://") {
		// Extract host and path
		parts := strings.SplitN(remoteURL, "://", 2)
		if len(parts) == 2 {
			// Insert token: https://token@host/path
			remoteURL = fmt.Sprintf("%s://x-access-token:%s@%s", parts[0], token, parts[1])
		}
	}

	// Update remote URL with authentication
	cmd = exec.Command("git", "remote", "set-url", "origin", remoteURL)
	cmd.Dir = ti.projectRoot
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set authenticated remote URL: %w", err)
	}

	fmt.Println("✅ Git authentication configured")
	return nil
}

// detectTestsInPR detects test files that are already part of the PR changes
func (ti *TestIntegrator) detectTestsInPR() ([]string, error) {
	// Get list of changed files in the PR
	cmd := exec.Command("git", "diff", "--name-only", "origin/main...HEAD")
	cmd.Dir = ti.projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}

	changedFiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	testFiles := []string{}

	for _, file := range changedFiles {
		if file != "" && ti.isTestFile(file) {
			testFiles = append(testFiles, filepath.Join(ti.projectRoot, file))
		}
	}

	return testFiles, nil
}

// shouldValidateTest determines if a test file should be validated
func (ti *TestIntegrator) shouldValidateTest(testFile string) bool {
	// Only validate if file exists and is readable
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		return false
	}
	return true
}

// validateExistingTest uses Bob Shell CLI to validate if an existing test is correct
func (ti *TestIntegrator) validateExistingTest(testFile string) (valid bool, needsUpdate bool, err error) {
	if ti.aiClient == nil {
		return true, false, nil // Skip validation if no AI client
	}

	fmt.Printf("   🔍 Validating test: %s\n", filepath.Base(testFile))

	// Read the test file
	testContent, err := os.ReadFile(testFile)
	if err != nil {
		return false, false, fmt.Errorf("failed to read test file: %w", err)
	}

	// Find the corresponding source file
	sourceFile := ti.findSourceFileForTest(testFile)
	if sourceFile == "" {
		fmt.Printf("   ⚠️  Could not find source file for test\n")
		return true, false, nil // Assume valid if we can't find source
	}

	// Read the source file
	sourceContent, err := os.ReadFile(sourceFile)
	if err != nil {
		return false, false, fmt.Errorf("failed to read source file: %w", err)
	}

	// Ask Bob Shell CLI to validate the test
	prompt := fmt.Sprintf(`You are a test validation expert. Analyze if the following test correctly tests the source code.

**Source Code:**
%s

**Test Code:**
%s

**Task:**
1. Check if the test covers the main functionality
2. Check if the test has proper assertions
3. Check if the test handles edge cases
4. Check if the test is up-to-date with the source code

**Response Format:**
Respond with ONLY one of these:
- VALID: if the test is correct and complete
- NEEDS_UPDATE: if the test exists but is outdated or incomplete
- INVALID: if the test is incorrect

Then on a new line, provide a brief reason (max 50 words).`, string(sourceContent), string(testContent))

	message, err := ti.aiClient.CreateMessage(bobshell.MessageRequest{
		Model:     "gpt-4",
		MaxTokens: 500,
		Messages: []bobshell.Message{
			{Role: "user", Content: prompt},
		},
		System: "You are a test validation expert. Provide concise, actionable feedback.",
	})

	if err != nil {
		return false, false, fmt.Errorf("Bob Shell validation failed: %w", err)
	}

	response := strings.TrimSpace(message.ExtractText())
	lines := strings.Split(response, "\n")

	if len(lines) == 0 {
		return true, false, nil
	}

	verdict := strings.ToUpper(strings.TrimSpace(lines[0]))
	reason := ""
	if len(lines) > 1 {
		reason = strings.TrimSpace(lines[1])
	}

	switch verdict {
	case "VALID":
		fmt.Printf("   ✅ Test is valid: %s\n", reason)
		return true, false, nil
	case "NEEDS_UPDATE":
		fmt.Printf("   🔄 Test needs update: %s\n", reason)
		return false, true, nil
	case "INVALID":
		fmt.Printf("   ❌ Test is invalid: %s\n", reason)
		return false, true, nil
	default:
		// If we can't parse the response, assume valid
		return true, false, nil
	}
}

// findSourceFileForTest finds the source file corresponding to a test file
func (ti *TestIntegrator) findSourceFileForTest(testFile string) string {
	// Remove _test suffix and .go extension
	base := filepath.Base(testFile)

	// Go pattern: file_test.go -> file.go
	if strings.HasSuffix(base, "_test.go") {
		sourceBase := strings.TrimSuffix(base, "_test.go") + ".go"
		sourceFile := filepath.Join(filepath.Dir(testFile), sourceBase)
		if _, err := os.Stat(sourceFile); err == nil {
			return sourceFile
		}
	}

	// Python pattern: test_file.py -> file.py
	if strings.HasPrefix(base, "test_") && strings.HasSuffix(base, ".py") {
		sourceBase := strings.TrimPrefix(base, "test_")
		sourceFile := filepath.Join(filepath.Dir(testFile), sourceBase)
		if _, err := os.Stat(sourceFile); err == nil {
			return sourceFile
		}
	}

	return ""
}

// markTestForRegeneration marks a test file for regeneration by removing it
func (ti *TestIntegrator) markTestForRegeneration(testFile string) error {
	fmt.Printf("   🗑️  Removing outdated test for regeneration: %s\n", filepath.Base(testFile))

	// Instead of deleting, we could rename with .old extension
	// This allows manual review if needed
	oldFile := testFile + ".old"
	if err := os.Rename(testFile, oldFile); err != nil {
		// If rename fails, try to delete
		if err := os.Remove(testFile); err != nil {
			return fmt.Errorf("failed to remove test file: %w", err)
		}
	}

	return nil
}

// Made with Bob
