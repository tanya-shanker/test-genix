package prcreator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"
	"github.com/tanya-shanker/test-genix/pkg/types"
	"golang.org/x/oauth2"
)

// PRCreator creates pull requests for functional tests
type PRCreator struct {
	config         *types.Config
	githubClient   *github.Client
	projectRoot    string
	functionalRepo string
}

// NewPRCreator creates a new PR creator
func NewPRCreator(config *types.Config, projectRoot string) *PRCreator {
	var client *github.Client

	if config.GitHubToken != "" {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: config.GitHubToken},
		)
		tc := oauth2.NewClient(ctx, ts)

		if config.GitHubEnterpriseURL != "" {
			client, _ = github.NewEnterpriseClient(config.GitHubEnterpriseURL, config.GitHubEnterpriseURL, tc)
		} else {
			client = github.NewClient(tc)
		}
	}

	return &PRCreator{
		config:         config,
		githubClient:   client,
		projectRoot:    projectRoot,
		functionalRepo: config.FunctionalTestRepo,
	}
}

// CreateFunctionalTestPR creates a PR for functional tests
func (pc *PRCreator) CreateFunctionalTestPR(testDir, sourceBranch, sourceCommit string) (*types.PRCreationResult, error) {
	fmt.Println("🔀 Creating PR for functional tests...")

	// Check if functional test repository is configured
	if pc.functionalRepo == "" {
		fmt.Println("ℹ️  Functional test repository not configured - skipping PR creation")
		fmt.Println("   Set FUNCTIONAL_TEST_REPO environment variable to:")
		fmt.Println("     - 'owner/repo' format (e.g., 'your-org/functional-tests')")
		fmt.Println("     - Full URL (e.g., 'https://github.com/your-org/functional-tests.git')")
		return nil, nil
	}

	// Check if functional tests were generated
	if !pc.hasFunctionalTests(testDir) {
		fmt.Println("ℹ️  No functional tests generated - skipping PR creation")
		return nil, nil
	}

	// Clone or update functional test repository
	repoPath, err := pc.prepareFunctionalTestRepo()
	if err != nil {
		return nil, fmt.Errorf("failed to prepare functional test repo: %w", err)
	}

	// Create a new branch for the tests
	branchName := fmt.Sprintf("auto-tests/%s-%d", sourceBranch, time.Now().Unix())
	if err := pc.createBranch(repoPath, branchName); err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}

	// Copy functional tests to the repository
	if err := pc.copyFunctionalTests(testDir, repoPath); err != nil {
		return nil, fmt.Errorf("failed to copy functional tests: %w", err)
	}

	// Commit and push changes
	if err := pc.commitAndPush(repoPath, branchName, sourceBranch, sourceCommit); err != nil {
		return nil, fmt.Errorf("failed to commit and push: %w", err)
	}

	// Create pull request using GitHub API
	result, err := pc.createGitHubPR(branchName, sourceBranch, sourceCommit)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub PR: %w", err)
	}

	fmt.Printf("✅ Functional test PR created: %s\n", result.PRURL)

	return result, nil
}

// hasFunctionalTests checks if functional tests were generated
func (pc *PRCreator) hasFunctionalTests(testDir string) bool {
	entries, err := os.ReadDir(testDir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.Contains(entry.Name(), "functional") {
			return true
		}
	}

	return false
}

// prepareFunctionalTestRepo clones or updates the functional test repository
func (pc *PRCreator) prepareFunctionalTestRepo() (string, error) {
	repoPath := filepath.Join(pc.projectRoot, ".tmp", "functional-tests")

	// Check if repo already exists
	if _, err := os.Stat(repoPath); err == nil {
		// Repository exists, update it
		fmt.Println("📥 Updating functional test repository...")
		cmd := exec.Command("git", "pull", "origin", "main")
		cmd.Dir = repoPath
		if output, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("⚠️  Failed to update repo: %v\n%s", err, output)
		}
		return repoPath, nil
	}

	// Clone the repository
	fmt.Println("📥 Cloning functional test repository...")

	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(repoPath), 0755); err != nil {
		return "", err
	}

	// Construct repository URL
	repoURL := pc.getFunctionalRepoURL()

	cmd := exec.Command("git", "clone", repoURL, repoPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to clone repo: %w\n%s", err, output)
	}

	return repoPath, nil
}

// getFunctionalRepoURL constructs the functional test repository URL
func (pc *PRCreator) getFunctionalRepoURL() string {
	fmt.Printf("DEBUG: functionalRepo value = '%s'\n", pc.functionalRepo)

	// If functionalRepo is already a full URL, return it as-is
	if strings.HasPrefix(pc.functionalRepo, "http://") || strings.HasPrefix(pc.functionalRepo, "https://") {
		// Remove trailing slash if present
		cleanURL := strings.TrimSuffix(pc.functionalRepo, "/")
		fmt.Printf("DEBUG: Detected full URL, returning: '%s'\n", cleanURL)
		return cleanURL
	}

	// Otherwise, construct the URL from owner/repo format
	var constructedURL string
	if pc.config.GitHubEnterpriseURL != "" {
		// GitHub Enterprise
		constructedURL = fmt.Sprintf("%s/%s.git", pc.config.GitHubEnterpriseURL, pc.functionalRepo)
	} else {
		// GitHub.com
		constructedURL = fmt.Sprintf("https://github.com/%s.git", pc.functionalRepo)
	}
	fmt.Printf("DEBUG: Constructed URL from owner/repo: '%s'\n", constructedURL)
	return constructedURL
}

// createBranch creates a new branch in the repository
func (pc *PRCreator) createBranch(repoPath, branchName string) error {
	fmt.Printf("🌿 Creating branch: %s\n", branchName)

	// Checkout main/master first
	cmd := exec.Command("git", "checkout", "main")
	cmd.Dir = repoPath
	if _, err := cmd.CombinedOutput(); err != nil {
		// Try master if main doesn't exist
		cmd = exec.Command("git", "checkout", "master")
		cmd.Dir = repoPath
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to checkout base branch: %w\n%s", err, output)
		}
	}

	// Pull latest changes
	cmd = exec.Command("git", "pull")
	cmd.Dir = repoPath
	cmd.Run() // Ignore errors

	// Create and checkout new branch
	cmd = exec.Command("git", "checkout", "-b", branchName)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create branch: %w\n%s", err, output)
	}

	return nil
}

// copyFunctionalTests copies functional tests to the repository
func (pc *PRCreator) copyFunctionalTests(testDir, repoPath string) error {
	fmt.Println("📋 Copying functional tests...")

	// Create tests directory in repo if it doesn't exist
	testsDir := filepath.Join(repoPath, "tests", "auto-generated")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		return err
	}

	// Copy all test files
	entries, err := os.ReadDir(testDir)
	if err != nil {
		return err
	}

	copiedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		sourcePath := filepath.Join(testDir, entry.Name())
		targetPath := filepath.Join(testsDir, entry.Name())

		if err := pc.copyFile(sourcePath, targetPath); err != nil {
			fmt.Printf("⚠️  Failed to copy %s: %v\n", entry.Name(), err)
			continue
		}

		copiedCount++
	}

	fmt.Printf("✅ Copied %d functional test files\n", copiedCount)

	return nil
}

// copyFile copies a file from source to target
func (pc *PRCreator) copyFile(source, target string) error {
	input, err := os.ReadFile(source)
	if err != nil {
		return err
	}

	return os.WriteFile(target, input, 0644)
}

// commitAndPush commits and pushes changes
func (pc *PRCreator) commitAndPush(repoPath, branchName, sourceBranch, sourceCommit string) error {
	fmt.Println("💾 Committing and pushing changes...")

	// Configure git user
	cmd := exec.Command("git", "config", "user.name", "AI Test Orchestrator")
	cmd.Dir = repoPath
	cmd.Run()

	cmd = exec.Command("git", "config", "user.email", "test-orchestrator@ibm.com")
	cmd.Dir = repoPath
	cmd.Run()

	// Add all changes
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add files: %w\n%s", err, output)
	}

	// Commit
	commitMsg := fmt.Sprintf(`🤖 Add auto-generated functional tests

Source Branch: %s
Source Commit: %s

These functional tests were automatically generated by the Intelligent Test Orchestrator
based on code changes detected in the source repository.`, sourceBranch, sourceCommit)

	cmd = exec.Command("git", "commit", "-m", commitMsg)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to commit: %w\n%s", err, output)
	}

	// Push to remote
	cmd = exec.Command("git", "push", "-u", "origin", branchName)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to push: %w\n%s", err, output)
	}

	fmt.Println("✅ Changes committed and pushed")

	return nil
}

// createGitHubPR creates a pull request using GitHub API
func (pc *PRCreator) createGitHubPR(branchName, sourceBranch, sourceCommit string) (*types.PRCreationResult, error) {
	if pc.githubClient == nil {
		return nil, fmt.Errorf("GitHub client not initialized - check GitHub token")
	}

	ctx := context.Background()

	// Parse repository owner and name
	var owner, repo string

	// Check if functionalRepo is a full URL
	if strings.HasPrefix(pc.functionalRepo, "http://") || strings.HasPrefix(pc.functionalRepo, "https://") {
		// Extract owner/repo from URL
		// Example: https://github.com/owner/repo.git -> owner/repo
		repoURL := strings.TrimSuffix(pc.functionalRepo, "/")
		repoURL = strings.TrimSuffix(repoURL, ".git")

		// Split by "/" and get last two parts
		parts := strings.Split(repoURL, "/")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid repository URL format: %s", pc.functionalRepo)
		}
		owner = parts[len(parts)-2]
		repo = parts[len(parts)-1]
	} else {
		// Parse as owner/repo format
		parts := strings.Split(pc.functionalRepo, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid repository format: %s (expected 'owner/repo' or full URL)", pc.functionalRepo)
		}
		owner = parts[0]
		repo = parts[1]
	}

	// Create PR
	title := fmt.Sprintf("🤖 Auto-generated functional tests from %s", sourceBranch)
	body := fmt.Sprintf(`## Auto-Generated Functional Tests

This PR contains automatically generated functional tests based on code changes detected in the source repository.

**Source Information:**
- Branch: %s
- Commit: %s

**Generated by:** Intelligent Test Orchestrator
**Timestamp:** %s

### What's Included
- End-to-end test scenarios
- Integration test cases
- Functional validation tests

### Next Steps
1. Review the generated tests
2. Run the test suite to verify functionality
3. Merge if tests are appropriate

---
*This PR was automatically created by the AI Test Orchestrator*`, sourceBranch, sourceCommit, time.Now().Format(time.RFC3339))

	baseBranch := "main"
	newPR := &github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(branchName),
		Base:  github.String(baseBranch),
		Body:  github.String(body),
	}

	pr, _, err := pc.githubClient.PullRequests.Create(ctx, owner, repo, newPR)
	if err != nil {
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	result := &types.PRCreationResult{
		PRURL:      pr.GetHTMLURL(),
		PRNumber:   pr.GetNumber(),
		Branch:     branchName,
		Repository: pc.functionalRepo,
		CreatedAt:  time.Now(),
	}

	return result, nil
}

// CreatePRWithCLI creates a PR using GitHub CLI as fallback
func (pc *PRCreator) CreatePRWithCLI(repoPath, branchName, sourceBranch, sourceCommit string) (*types.PRCreationResult, error) {
	fmt.Println("🔀 Creating PR using GitHub CLI...")

	title := fmt.Sprintf("🤖 Auto-generated functional tests from %s", sourceBranch)
	body := fmt.Sprintf("Auto-generated functional tests from branch %s (commit %s)", sourceBranch, sourceCommit)

	cmd := exec.Command("gh", "pr", "create",
		"--title", title,
		"--body", body,
		"--base", "main",
		"--head", branchName)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to create PR with CLI: %w\n%s", err, output)
	}

	// Parse PR URL from output
	prURL := strings.TrimSpace(string(output))

	result := &types.PRCreationResult{
		PRURL:      prURL,
		Branch:     branchName,
		Repository: pc.functionalRepo,
		CreatedAt:  time.Now(),
	}

	return result, nil
}

// Made with Bob
