package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tanya-shanker/test-genix/pkg/coverage"
	"github.com/tanya-shanker/test-genix/pkg/detector"
	"github.com/tanya-shanker/test-genix/pkg/generator"
	"github.com/tanya-shanker/test-genix/pkg/integrator"
	"github.com/tanya-shanker/test-genix/pkg/prcreator"
	"github.com/tanya-shanker/test-genix/pkg/types"
	"gopkg.in/yaml.v3"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
)

func main() {
	// Parse command line flags
	baseBranch := flag.String("base", "main", "Base branch name")
	headBranch := flag.String("head", "", "Head branch name")
	configPath := flag.String("config", "config/test-orchestrator-config.yaml", "Configuration file path")
	outputDir := flag.String("output", "generated-tests", "Output directory for generated tests")
	projectRoot := flag.String("root", ".", "Project root directory")
	flag.Parse()

	// Get head branch from environment if not provided
	if *headBranch == "" {
		*headBranch = os.Getenv("GIT_BRANCH")
		if *headBranch == "" {
			*headBranch = getCurrentBranch(*projectRoot)
		}
	}

	// Print banner
	printBanner()

	// Load configuration
	config, err := loadConfig(*configPath)
	if err != nil {
		logError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	// Override config with environment variables
	overrideConfigFromEnv(config)

	// Create output directories
	unitTestDir := filepath.Join(*outputDir, "unit")
	functionalTestDir := filepath.Join(*outputDir, "functional")
	reportsDir := filepath.Join(*outputDir, "reports")

	if err := os.MkdirAll(unitTestDir, 0755); err != nil {
		logError(fmt.Sprintf("Failed to create output directories: %v", err))
		os.Exit(1)
	}
	if err := os.MkdirAll(functionalTestDir, 0755); err != nil {
		logError(fmt.Sprintf("Failed to create output directories: %v", err))
		os.Exit(1)
	}
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		logError(fmt.Sprintf("Failed to create output directories: %v", err))
		os.Exit(1)
	}

	// Step 1: Detect changes
	logInfo("Step 1: Detecting code changes...")
	changeDetector := detector.NewChangeDetector(*baseBranch, *headBranch, *projectRoot)
	changes, err := changeDetector.DetectChanges()
	if err != nil {
		logError(fmt.Sprintf("Failed to detect changes: %v", err))
		os.Exit(1)
	}

	// Save changes report
	changesPath := filepath.Join(reportsDir, "changes.json")
	if err := saveJSON(changes, changesPath); err != nil {
		logWarning(fmt.Sprintf("Failed to save changes report: %v", err))
	}

	// Step 2: Generate tests
	logInfo("Step 2: Generating AI-powered tests...")
	testGenerator := generator.NewTestGenerator(config, *projectRoot)
	stats, err := testGenerator.GenerateTests(changes, unitTestDir, functionalTestDir)
	if err != nil {
		logError(fmt.Sprintf("Failed to generate tests: %v", err))
		os.Exit(1)
	}

	// Save generation stats
	statsPath := filepath.Join(reportsDir, "generation-report.json")
	if err := saveJSON(stats, statsPath); err != nil {
		logWarning(fmt.Sprintf("Failed to save generation report: %v", err))
	}

	// Step 3: Integrate unit tests
	logInfo("Step 3: Integrating unit tests into test suite...")
	testIntegrator := integrator.NewTestIntegrator(*projectRoot)
	targetTestDir := filepath.Join(*projectRoot, "tests")
	integrationResult, err := testIntegrator.IntegrateUnitTests(unitTestDir, targetTestDir)
	if err != nil {
		logError(fmt.Sprintf("Failed to integrate tests: %v", err))
		os.Exit(1)
	}

	// Save integration report
	integrationPath := filepath.Join(reportsDir, "integration-report.json")
	if err := saveJSON(integrationResult, integrationPath); err != nil {
		logWarning(fmt.Sprintf("Failed to save integration report: %v", err))
	}

	// Commit tests to current PR
	commit := os.Getenv("GIT_COMMIT")
	if commit == "" {
		commit = getCurrentCommit(*projectRoot)
	}

	if err := testIntegrator.CommitTests(targetTestDir, *headBranch, commit); err != nil {
		logWarning(fmt.Sprintf("Failed to commit tests: %v", err))
	}

	// Step 4: Create functional test PR
	logInfo("Step 4: Creating PR for functional tests...")
	prCreator := prcreator.NewPRCreator(config, *projectRoot)
	prResult, err := prCreator.CreateFunctionalTestPR(functionalTestDir, *headBranch, commit)
	if err != nil {
		logWarning(fmt.Sprintf("Failed to create functional test PR: %v", err))
	} else if prResult != nil {
		// Save PR creation report
		prPath := filepath.Join(reportsDir, "pr-creation-report.json")
		if err := saveJSON(prResult, prPath); err != nil {
			logWarning(fmt.Sprintf("Failed to save PR report: %v", err))
		}
		logSuccess(fmt.Sprintf("Functional test PR created: %s", prResult.PRURL))
	}

	// Step 5: Analyze coverage
	logInfo("Step 5: Analyzing test coverage...")
	coverageAnalyzer := coverage.NewCoverageAnalyzer(*projectRoot)
	coverageReport, err := coverageAnalyzer.AnalyzeCoverage(changes, config.CoverageTarget)
	if err != nil {
		logWarning(fmt.Sprintf("Failed to analyze coverage: %v", err))
	} else {
		// Save coverage report
		coveragePath := filepath.Join(reportsDir, "coverage-report.json")
		if err := coverageAnalyzer.SaveCoverageReport(coverageReport, coveragePath); err != nil {
			logWarning(fmt.Sprintf("Failed to save coverage report: %v", err))
		}

		// Copy to project root for pipeline display
		rootCoveragePath := filepath.Join(*projectRoot, "coverage-report.json")
		if err := saveJSON(coverageReport, rootCoveragePath); err != nil {
			logWarning(fmt.Sprintf("Failed to save root coverage report: %v", err))
		}

		// Print coverage summary
		fmt.Println()
		fmt.Println(coverageAnalyzer.GenerateCoverageSummary(coverageReport))
	}

	// Print final summary
	printSummary(stats, integrationResult, prResult, *outputDir)

	logSuccess("AI test generation workflow complete!")
}

func loadConfig(path string) (*types.Config, error) {
	// Check if config file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logWarning(fmt.Sprintf("Config file not found at %s, using defaults", path))
		return getDefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config types.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func getDefaultConfig() *types.Config {
	return &types.Config{
		TestFrameworks: map[string]string{
			"go":         "testing",
			"python":     "pytest",
			"javascript": "jest",
			"typescript": "jest",
			"java":       "junit",
		},
		CoverageTarget:      80.0,
		MaxTestsPerFunction: 5,
		GenerateEdgeCases:   true,
		GenerateMocks:       true,
		FunctionalTestRepo:  "", // Must be set via FUNCTIONAL_TEST_REPO environment variable
		AIModel:             "gpt-4",
		TestPatterns: types.TestPatterns{
			Unit:       "test_{function_name}",
			Functional: "test_{feature_name}_e2e",
		},
	}
}

func overrideConfigFromEnv(config *types.Config) {
	// Check for Anthropic API key (standard)
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.AIAPIKey = apiKey
	}
	// Also check for BOB_API_KEY (if Bob has a separate key)
	if apiKey := os.Getenv("BOB_API_KEY"); apiKey != "" {
		config.AIAPIKey = apiKey
	}
	// Check for CLAUDE_API_KEY as alternative
	if apiKey := os.Getenv("CLAUDE_API_KEY"); apiKey != "" {
		config.AIAPIKey = apiKey
	}
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		config.GitHubToken = token
	}
	if token := os.Getenv("GHE_TOKEN"); token != "" {
		config.GitHubToken = token
	}
	if url := os.Getenv("GITHUB_ENTERPRISE_URL"); url != "" {
		config.GitHubEnterpriseURL = url
	}
	if repo := os.Getenv("FUNCTIONAL_TEST_REPO"); repo != "" {
		config.FunctionalTestRepo = repo
	} else if config.FunctionalTestRepo == "" {
		// Warn if functional test repo is not configured
		fmt.Println("⚠️  Warning: FUNCTIONAL_TEST_REPO environment variable not set")
		fmt.Println("   Functional test PR creation will be skipped")
		fmt.Println("   Set FUNCTIONAL_TEST_REPO to 'owner/repo' format (e.g., 'your-org/functional-tests')")
	}
}

func getCurrentBranch(projectRoot string) string {
	// Try to get from git
	cmd := fmt.Sprintf("cd %s && git rev-parse --abbrev-ref HEAD", projectRoot)
	output, err := execCommand(cmd)
	if err != nil {
		return "main"
	}
	return output
}

func getCurrentCommit(projectRoot string) string {
	cmd := fmt.Sprintf("cd %s && git rev-parse HEAD", projectRoot)
	output, err := execCommand(cmd)
	if err != nil {
		return "unknown"
	}
	return output
}

func execCommand(cmd string) (string, error) {
	// This is a simplified version - in production, use exec.Command properly
	return "", fmt.Errorf("not implemented")
}

func saveJSON(data interface{}, path string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, jsonData, 0644)
}

func printBanner() {
	fmt.Println()
	fmt.Println("==================================================")
	fmt.Println("🤖 Intelligent Test Orchestrator")
	fmt.Println("==================================================")
	fmt.Println()
}

func printSummary(stats *types.GenerationStats, integration *integrator.IntegrationResult, pr *types.PRCreationResult, outputDir string) {
	fmt.Println()
	fmt.Println("==================================================")
	fmt.Println("📋 Summary")
	fmt.Println("==================================================")
	fmt.Println()

	if stats != nil {
		fmt.Printf("Unit Tests Generated: %d\n", stats.UnitTestsGenerated)
		fmt.Printf("Functional Tests Generated: %d\n", stats.FunctionalTestsGenerated)
		fmt.Printf("Total Test Cases: %d\n", stats.TotalTestCases)
		fmt.Printf("Execution Time: %.2fs\n", stats.ExecutionTimeSeconds)
	}

	if integration != nil {
		fmt.Printf("\nIntegration Results:\n")
		fmt.Printf("  Files Integrated: %d\n", integration.FilesIntegrated)
		fmt.Printf("  Duplicates Skipped: %d\n", integration.DuplicatesSkipped)
	}

	if pr != nil {
		fmt.Printf("\nFunctional Test PR: %s\n", pr.PRURL)
	}

	fmt.Println()
	fmt.Println("📁 Generated artifacts:")
	fmt.Printf("   - Unit tests: %s/unit\n", outputDir)
	fmt.Printf("   - Functional tests: %s/functional\n", outputDir)
	fmt.Printf("   - Reports: %s/reports\n", outputDir)
	fmt.Println()
	fmt.Println("==================================================")
}

func logInfo(msg string) {
	fmt.Printf("%s[INFO]%s %s\n", colorBlue, colorReset, msg)
}

func logSuccess(msg string) {
	fmt.Printf("%s[SUCCESS]%s %s\n", colorGreen, colorReset, msg)
}

func logWarning(msg string) {
	fmt.Printf("%s[WARNING]%s %s\n", colorYellow, colorReset, msg)
}

func logError(msg string) {
	fmt.Printf("%s[ERROR]%s %s\n", colorRed, colorReset, msg)
}

// Made with Bob
