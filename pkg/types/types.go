package types

import "time"

// ChangeInfo represents information about code changes
type ChangeInfo struct {
	ChangedFiles      []FileChange     `json:"changed_files"`
	AddedFiles        []FileChange     `json:"added_files"`
	DeletedFiles      []FileChange     `json:"deleted_files"`
	ModifiedFunctions []FunctionChange `json:"modified_functions"`
	ModifiedClasses   []ClassChange    `json:"modified_classes"`
	AffectedModules   []string         `json:"affected_modules"`
	SemanticChanges   []SemanticChange `json:"semantic_changes"`
}

// FileChange represents a changed file
type FileChange struct {
	Path         string `json:"path"`
	Status       string `json:"status"`
	Language     string `json:"language"`
	LinesChanged int    `json:"lines_changed"`
	Diff         string `json:"diff,omitempty"`
}

// FunctionChange represents a modified function
type FunctionChange struct {
	Name     string   `json:"name"`
	File     string   `json:"file"`
	Line     int      `json:"line"`
	Type     string   `json:"type"`
	IsAsync  bool     `json:"is_async,omitempty"`
	Args     []string `json:"args,omitempty"`
	Language string   `json:"language"`
}

// ClassChange represents a modified class
type ClassChange struct {
	Name     string   `json:"name"`
	File     string   `json:"file"`
	Line     int      `json:"line"`
	Type     string   `json:"type"`
	Methods  []string `json:"methods,omitempty"`
	Language string   `json:"language"`
}

// SemanticChange represents a semantic code change
type SemanticChange struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	File   string `json:"file"`
	Impact string `json:"impact"`
}

// TestCase represents a generated test case
type TestCase struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Code        string `json:"code"`
	FilePath    string `json:"file_path"`
}

// GenerationStats tracks test generation statistics
type GenerationStats struct {
	UnitTestsGenerated       int     `json:"unit_tests_generated"`
	FunctionalTestsGenerated int     `json:"functional_tests_generated"`
	TotalTestCases           int     `json:"total_test_cases"`
	ExecutionTimeSeconds     float64 `json:"execution_time_seconds"`
	FilesProcessed           int     `json:"files_processed"`
}

// CoverageReport represents test coverage information
type CoverageReport struct {
	CoverageBefore  float64 `json:"coverage_before"`
	CoverageAfter   float64 `json:"coverage_after"`
	CoverageDelta   float64 `json:"coverage_delta"`
	NewLinesCovered int     `json:"new_lines_covered"`
	NewLinesTotal   int     `json:"new_lines_total"`
	CoverageTarget  float64 `json:"coverage_target"`
	TargetMet       bool    `json:"target_met"`
	Timestamp       string  `json:"timestamp"`
}

// Config represents the orchestrator configuration
type Config struct {
	TestFrameworks      map[string]string `yaml:"test_frameworks"`
	CoverageTarget      float64           `yaml:"coverage_target"`
	MaxTestsPerFunction int               `yaml:"max_tests_per_function"`
	GenerateEdgeCases   bool              `yaml:"generate_edge_cases"`
	GenerateMocks       bool              `yaml:"generate_mocks"`
	FunctionalTestRepo  string            `yaml:"functional_test_repo"`
	AIModel             string            `yaml:"ai_model"`   // e.g., "claude-3-5-sonnet-20241022"
	AIAPIKey            string            `yaml:"ai_api_key"` // Anthropic API key
	GitHubToken         string            `yaml:"github_token"`
	GitHubEnterpriseURL string            `yaml:"github_enterprise_url"`
	TestPatterns        TestPatterns      `yaml:"test_patterns"`
}

// TestPatterns defines naming patterns for tests
type TestPatterns struct {
	Unit       string `yaml:"unit"`
	Functional string `yaml:"functional"`
}

// PRCreationResult represents the result of PR creation
type PRCreationResult struct {
	PRURL      string    `json:"pr_url"`
	PRNumber   int       `json:"pr_number"`
	Branch     string    `json:"branch"`
	Repository string    `json:"repository"`
	CreatedAt  time.Time `json:"created_at"`
}

// Made with Bob
