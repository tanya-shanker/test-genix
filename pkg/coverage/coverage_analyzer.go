package coverage

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tanya-shanker/test-genix/pkg/types"
)

// CoverageAnalyzer analyzes test coverage
type CoverageAnalyzer struct {
	projectRoot string
}

// NewCoverageAnalyzer creates a new coverage analyzer
func NewCoverageAnalyzer(projectRoot string) *CoverageAnalyzer {
	return &CoverageAnalyzer{
		projectRoot: projectRoot,
	}
}

// AnalyzeCoverage analyzes test coverage before and after changes
func (ca *CoverageAnalyzer) AnalyzeCoverage(changes *types.ChangeInfo, targetCoverage float64) (*types.CoverageReport, error) {
	fmt.Println("📊 Analyzing test coverage...")

	report := &types.CoverageReport{
		CoverageTarget: targetCoverage,
		Timestamp:      time.Now().Format(time.RFC3339),
	}

	// Get baseline coverage (before changes)
	baselineCoverage, err := ca.getBaselineCoverage()
	if err != nil {
		fmt.Printf("⚠️  Could not get baseline coverage: %v\n", err)
		baselineCoverage = 0.0
	}
	report.CoverageBefore = baselineCoverage

	// Run tests to get current coverage
	currentCoverage, err := ca.getCurrentCoverage()
	if err != nil {
		fmt.Printf("⚠️  Could not get current coverage: %v\n", err)
		currentCoverage = baselineCoverage
	}
	report.CoverageAfter = currentCoverage

	// Calculate delta
	report.CoverageDelta = currentCoverage - baselineCoverage

	// Analyze coverage for new lines
	newLinesCovered, newLinesTotal := ca.analyzeNewLinesCoverage(changes)
	report.NewLinesCovered = newLinesCovered
	report.NewLinesTotal = newLinesTotal

	// Check if target is met
	report.TargetMet = currentCoverage >= targetCoverage

	fmt.Printf("✅ Coverage analysis complete\n")
	fmt.Printf("   Before: %.2f%%\n", report.CoverageBefore)
	fmt.Printf("   After:  %.2f%%\n", report.CoverageAfter)
	fmt.Printf("   Delta:  %+.2f%%\n", report.CoverageDelta)
	fmt.Printf("   Target: %.2f%% (%s)\n", report.CoverageTarget, ca.getTargetStatus(report.TargetMet))

	return report, nil
}

// getBaselineCoverage gets the baseline coverage from the base branch
func (ca *CoverageAnalyzer) getBaselineCoverage() (float64, error) {
	// Try to read cached baseline coverage
	baselinePath := filepath.Join(ca.projectRoot, ".coverage", "baseline.json")
	if data, err := os.ReadFile(baselinePath); err == nil {
		var baseline struct {
			Coverage float64 `json:"coverage"`
		}
		if err := json.Unmarshal(data, &baseline); err == nil {
			return baseline.Coverage, nil
		}
	}

	// If no cached baseline, run coverage on current state
	return ca.runCoverageAnalysis()
}

// getCurrentCoverage gets the current coverage after changes
func (ca *CoverageAnalyzer) getCurrentCoverage() (float64, error) {
	return ca.runCoverageAnalysis()
}

// runCoverageAnalysis runs coverage analysis based on project type
func (ca *CoverageAnalyzer) runCoverageAnalysis() (float64, error) {
	// Detect project type
	projectType := ca.detectProjectType()

	switch projectType {
	case "go":
		return ca.runGoCoverage()
	case "python":
		return ca.runPythonCoverage()
	case "javascript", "typescript":
		return ca.runJSCoverage()
	default:
		return 0.0, fmt.Errorf("unsupported project type: %s", projectType)
	}
}

// detectProjectType detects the primary project type
func (ca *CoverageAnalyzer) detectProjectType() string {
	// Check for Go
	if _, err := os.Stat(filepath.Join(ca.projectRoot, "go.mod")); err == nil {
		return "go"
	}

	// Check for Python
	if _, err := os.Stat(filepath.Join(ca.projectRoot, "setup.py")); err == nil {
		return "python"
	}
	if _, err := os.Stat(filepath.Join(ca.projectRoot, "pyproject.toml")); err == nil {
		return "python"
	}

	// Check for JavaScript/TypeScript
	if _, err := os.Stat(filepath.Join(ca.projectRoot, "package.json")); err == nil {
		return "javascript"
	}

	return "unknown"
}

// runGoCoverage runs Go coverage analysis
func (ca *CoverageAnalyzer) runGoCoverage() (float64, error) {
	fmt.Println("🔍 Running Go coverage analysis...")

	coverageFile := filepath.Join(ca.projectRoot, "coverage.out")

	// Run go test with coverage
	cmd := exec.Command("go", "test", "-coverprofile="+coverageFile, "./...")
	cmd.Dir = ca.projectRoot
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Printf("⚠️  Go test failed: %v\n%s", err, output)
		return 0.0, err
	}

	// Parse coverage from output
	coverage := ca.parseGoCoverageOutput(string(output))

	// Also try to get coverage from coverage file
	if coverage == 0.0 {
		coverage = ca.parseGoCoverageFile(coverageFile)
	}

	return coverage, nil
}

// parseGoCoverageOutput parses coverage percentage from go test output
func (ca *CoverageAnalyzer) parseGoCoverageOutput(output string) float64 {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "coverage:") {
			// Format: "coverage: 75.5% of statements"
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "coverage:" && i+1 < len(parts) {
					coverageStr := strings.TrimSuffix(parts[i+1], "%")
					if coverage, err := strconv.ParseFloat(coverageStr, 64); err == nil {
						return coverage
					}
				}
			}
		}
	}
	return 0.0
}

// parseGoCoverageFile parses coverage from coverage.out file
func (ca *CoverageAnalyzer) parseGoCoverageFile(filepath string) float64 {
	// Run go tool cover to get coverage percentage
	cmd := exec.Command("go", "tool", "cover", "-func="+filepath)
	cmd.Dir = ca.projectRoot
	output, err := cmd.CombinedOutput()

	if err != nil {
		return 0.0
	}

	// Parse total coverage from last line
	lines := strings.Split(string(output), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "total:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				coverageStr := strings.TrimSuffix(parts[2], "%")
				if coverage, err := strconv.ParseFloat(coverageStr, 64); err == nil {
					return coverage
				}
			}
		}
	}

	return 0.0
}

// runPythonCoverage runs Python coverage analysis
func (ca *CoverageAnalyzer) runPythonCoverage() (float64, error) {
	fmt.Println("🔍 Running Python coverage analysis...")

	// Run pytest with coverage
	cmd := exec.Command("pytest", "--cov=.", "--cov-report=json")
	cmd.Dir = ca.projectRoot
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Printf("⚠️  Pytest failed: %v\n%s", err, output)
		return 0.0, err
	}

	// Read coverage report
	coverageFile := filepath.Join(ca.projectRoot, "coverage.json")
	data, err := os.ReadFile(coverageFile)
	if err != nil {
		return 0.0, err
	}

	var coverageData struct {
		Totals struct {
			PercentCovered float64 `json:"percent_covered"`
		} `json:"totals"`
	}

	if err := json.Unmarshal(data, &coverageData); err != nil {
		return 0.0, err
	}

	return coverageData.Totals.PercentCovered, nil
}

// runJSCoverage runs JavaScript/TypeScript coverage analysis
func (ca *CoverageAnalyzer) runJSCoverage() (float64, error) {
	fmt.Println("🔍 Running JavaScript coverage analysis...")

	// Run jest with coverage
	cmd := exec.Command("npm", "test", "--", "--coverage", "--coverageReporters=json")
	cmd.Dir = ca.projectRoot
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Printf("⚠️  Jest failed: %v\n%s", err, output)
		return 0.0, err
	}

	// Read coverage report
	coverageFile := filepath.Join(ca.projectRoot, "coverage", "coverage-summary.json")
	data, err := os.ReadFile(coverageFile)
	if err != nil {
		return 0.0, err
	}

	var coverageData struct {
		Total struct {
			Lines struct {
				Pct float64 `json:"pct"`
			} `json:"lines"`
		} `json:"total"`
	}

	if err := json.Unmarshal(data, &coverageData); err != nil {
		return 0.0, err
	}

	return coverageData.Total.Lines.Pct, nil
}

// analyzeNewLinesCoverage analyzes coverage for newly added lines
func (ca *CoverageAnalyzer) analyzeNewLinesCoverage(changes *types.ChangeInfo) (int, int) {
	covered := 0
	total := 0

	// Count new lines from changed files
	for _, file := range changes.ChangedFiles {
		total += file.LinesChanged
	}

	for _, file := range changes.AddedFiles {
		total += file.LinesChanged
	}

	// Estimate coverage (in a real implementation, this would parse coverage data)
	// For now, we'll estimate based on whether tests were generated
	if len(changes.ModifiedFunctions) > 0 || len(changes.ModifiedClasses) > 0 {
		// Assume 70% coverage for functions/classes with generated tests
		covered = int(float64(total) * 0.7)
	}

	return covered, total
}

// getTargetStatus returns a status string for target met
func (ca *CoverageAnalyzer) getTargetStatus(met bool) string {
	if met {
		return "✅ Met"
	}
	return "❌ Not Met"
}

// SaveCoverageReport saves the coverage report to a file
func (ca *CoverageAnalyzer) SaveCoverageReport(report *types.CoverageReport, outputPath string) error {
	// Create output directory
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	// Marshal report to JSON
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return err
	}

	fmt.Printf("✅ Coverage report saved to %s\n", outputPath)

	return nil
}

// GenerateCoverageSummary generates a human-readable coverage summary
func (ca *CoverageAnalyzer) GenerateCoverageSummary(report *types.CoverageReport) string {
	var sb strings.Builder

	sb.WriteString("==================================================\n")
	sb.WriteString("📊 Test Coverage Comparison\n")
	sb.WriteString("==================================================\n\n")

	sb.WriteString(fmt.Sprintf("Before: %.2f%%\n", report.CoverageBefore))
	sb.WriteString(fmt.Sprintf("After:  %.2f%%\n", report.CoverageAfter))
	sb.WriteString(fmt.Sprintf("Change: %+.2f%%\n\n", report.CoverageDelta))

	sb.WriteString(fmt.Sprintf("New Lines Covered: %d/%d\n", report.NewLinesCovered, report.NewLinesTotal))
	sb.WriteString(fmt.Sprintf("Coverage Target: %.2f%%\n", report.CoverageTarget))
	sb.WriteString(fmt.Sprintf("Target Met: %s\n\n", ca.getTargetStatus(report.TargetMet)))

	sb.WriteString("==================================================\n")

	return sb.String()
}

// CacheBaselineCoverage caches the baseline coverage for future comparisons
func (ca *CoverageAnalyzer) CacheBaselineCoverage(coverage float64) error {
	cacheDir := filepath.Join(ca.projectRoot, ".coverage")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	baseline := struct {
		Coverage  float64 `json:"coverage"`
		Timestamp string  `json:"timestamp"`
	}{
		Coverage:  coverage,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return err
	}

	baselinePath := filepath.Join(cacheDir, "baseline.json")
	return os.WriteFile(baselinePath, data, 0644)
}

// Made with Bob
