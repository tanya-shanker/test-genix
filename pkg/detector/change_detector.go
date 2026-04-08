package detector

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tanya-shanker/test-genix/pkg/types"
)

// ChangeDetector analyzes code changes in pull requests
type ChangeDetector struct {
	baseBranch  string
	headBranch  string
	projectRoot string
}

// NewChangeDetector creates a new change detector
func NewChangeDetector(baseBranch, headBranch, projectRoot string) *ChangeDetector {
	return &ChangeDetector{
		baseBranch:  baseBranch,
		headBranch:  headBranch,
		projectRoot: projectRoot,
	}
}

// DetectChanges analyzes all code changes
func (cd *ChangeDetector) DetectChanges() (*types.ChangeInfo, error) {
	fmt.Println("🔍 Analyzing code changes...")

	changes := &types.ChangeInfo{
		ChangedFiles:      []types.FileChange{},
		AddedFiles:        []types.FileChange{},
		DeletedFiles:      []types.FileChange{},
		ModifiedFunctions: []types.FunctionChange{},
		ModifiedClasses:   []types.ClassChange{},
		AffectedModules:   []string{},
		SemanticChanges:   []types.SemanticChange{},
	}

	// Get git diff
	diffOutput, err := cd.getGitDiff()
	if err != nil {
		return nil, fmt.Errorf("failed to get git diff: %w", err)
	}

	// Parse diff to get changed files
	if err := cd.parseDiff(diffOutput, changes); err != nil {
		return nil, fmt.Errorf("failed to parse diff: %w", err)
	}

	// Analyze semantic changes for each file
	for _, fileInfo := range changes.ChangedFiles {
		if cd.isAnalyzableFile(fileInfo.Path) {
			cd.analyzeSemanticChanges(fileInfo, changes)
		}
	}

	// Identify affected modules
	cd.identifyAffectedModules(changes)

	fmt.Printf("✅ Detected %d changed files\n", len(changes.ChangedFiles))
	fmt.Printf("✅ Identified %d modified functions\n", len(changes.ModifiedFunctions))
	fmt.Printf("✅ Identified %d modified classes\n", len(changes.ModifiedClasses))

	return changes, nil
}

// getGitDiff retrieves the git diff between branches
func (cd *ChangeDetector) getGitDiff() (string, error) {
	// Try to get diff with base branch
	cmd := exec.Command("git", "diff", fmt.Sprintf("origin/%s...%s", cd.baseBranch, cd.headBranch), "--name-status")
	cmd.Dir = cd.projectRoot
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Fallback to comparing with HEAD
		fmt.Println("⚠️  Could not compare with base branch, using HEAD")
		cmd = exec.Command("git", "diff", "HEAD~1", "--name-status")
		cmd.Dir = cd.projectRoot
		output, err = cmd.CombinedOutput()
		if err != nil {
			return "", err
		}
	}

	return string(output), nil
}

// parseDiff parses git diff output to identify changed files
func (cd *ChangeDetector) parseDiff(diffOutput string, changes *types.ChangeInfo) error {
	scanner := bufio.NewScanner(strings.NewReader(diffOutput))

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		status := parts[0]
		filepath := parts[1]

		fileInfo := types.FileChange{
			Path:         filepath,
			Status:       status,
			Language:     cd.detectLanguage(filepath),
			LinesChanged: 0,
		}

		// Get detailed diff for this file
		detailedDiff, linesChanged := cd.getDetailedDiff(filepath)
		fileInfo.LinesChanged = linesChanged
		fileInfo.Diff = detailedDiff

		// Categorize by status
		switch {
		case strings.HasPrefix(status, "M"):
			changes.ChangedFiles = append(changes.ChangedFiles, fileInfo)
		case strings.HasPrefix(status, "A"):
			changes.AddedFiles = append(changes.AddedFiles, fileInfo)
		case strings.HasPrefix(status, "D"):
			changes.DeletedFiles = append(changes.DeletedFiles, fileInfo)
		}
	}

	return scanner.Err()
}

// getDetailedDiff gets the detailed diff for a specific file
func (cd *ChangeDetector) getDetailedDiff(filepath string) (string, int) {
	cmd := exec.Command("git", "diff", fmt.Sprintf("origin/%s...%s", cd.baseBranch, cd.headBranch), "--", filepath)
	cmd.Dir = cd.projectRoot
	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", 0
	}

	diff := string(output)
	linesChanged := cd.countChangedLines(diff)

	return diff, linesChanged
}

// countChangedLines counts the number of changed lines in a diff
func (cd *ChangeDetector) countChangedLines(diff string) int {
	count := 0
	scanner := bufio.NewScanner(strings.NewReader(diff))

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			count++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			count++
		}
	}

	return count
}

// detectLanguage detects programming language from file extension
func (cd *ChangeDetector) detectLanguage(filepath string) string {
	ext := strings.ToLower(filepath[strings.LastIndex(filepath, "."):])

	langMap := map[string]string{
		".go":   "go",
		".py":   "python",
		".js":   "javascript",
		".ts":   "typescript",
		".java": "java",
		".rb":   "ruby",
		".php":  "php",
		".cpp":  "cpp",
		".c":    "c",
		".cs":   "csharp",
	}

	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return "unknown"
}

// isAnalyzableFile checks if a file can be analyzed for semantic changes
func (cd *ChangeDetector) isAnalyzableFile(filepath string) bool {
	analyzableExts := []string{".go", ".py", ".js", ".ts", ".java"}
	ext := strings.ToLower(filepath[strings.LastIndex(filepath, "."):])

	for _, ae := range analyzableExts {
		if ext == ae {
			return true
		}
	}
	return false
}

// analyzeSemanticChanges analyzes semantic changes in a file
func (cd *ChangeDetector) analyzeSemanticChanges(fileInfo types.FileChange, changes *types.ChangeInfo) {
	switch fileInfo.Language {
	case "go":
		cd.analyzeGoChanges(fileInfo, changes)
	case "python":
		cd.analyzePythonChanges(fileInfo, changes)
	case "javascript", "typescript":
		cd.analyzeJSChanges(fileInfo, changes)
	}
}

// analyzeGoChanges analyzes Go file changes using AST
func (cd *ChangeDetector) analyzeGoChanges(fileInfo types.FileChange, changes *types.ChangeInfo) {
	fullPath := filepath.Join(cd.projectRoot, fileInfo.Path)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fullPath, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("⚠️  Could not parse %s: %v\n", fullPath, err)
		return
	}

	// Extract functions and types
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			funcInfo := types.FunctionChange{
				Name:     x.Name.Name,
				File:     fileInfo.Path,
				Line:     fset.Position(x.Pos()).Line,
				Type:     "function",
				Language: "go",
				Args:     cd.extractGoFuncArgs(x),
			}
			changes.ModifiedFunctions = append(changes.ModifiedFunctions, funcInfo)

			// Add semantic change
			changes.SemanticChanges = append(changes.SemanticChanges, types.SemanticChange{
				Type:   "function_modified",
				Name:   x.Name.Name,
				File:   fileInfo.Path,
				Impact: "medium",
			})

		case *ast.TypeSpec:
			if structType, ok := x.Type.(*ast.StructType); ok {
				classInfo := types.ClassChange{
					Name:     x.Name.Name,
					File:     fileInfo.Path,
					Line:     fset.Position(x.Pos()).Line,
					Type:     "struct",
					Language: "go",
					Methods:  cd.extractGoStructMethods(node, x.Name.Name),
				}
				changes.ModifiedClasses = append(changes.ModifiedClasses, classInfo)

				// Add semantic change
				changes.SemanticChanges = append(changes.SemanticChanges, types.SemanticChange{
					Type:   "struct_modified",
					Name:   x.Name.Name,
					File:   fileInfo.Path,
					Impact: "high",
				})
				_ = structType // Use the variable
			}
		}
		return true
	})
}

// extractGoFuncArgs extracts function arguments
func (cd *ChangeDetector) extractGoFuncArgs(fn *ast.FuncDecl) []string {
	args := []string{}
	if fn.Type.Params != nil {
		for _, param := range fn.Type.Params.List {
			for _, name := range param.Names {
				args = append(args, name.Name)
			}
		}
	}
	return args
}

// extractGoStructMethods extracts methods for a struct
func (cd *ChangeDetector) extractGoStructMethods(node *ast.File, structName string) []string {
	methods := []string{}

	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Recv != nil && len(fn.Recv.List) > 0 {
				// Check if this is a method of our struct
				if starExpr, ok := fn.Recv.List[0].Type.(*ast.StarExpr); ok {
					if ident, ok := starExpr.X.(*ast.Ident); ok && ident.Name == structName {
						methods = append(methods, fn.Name.Name)
					}
				} else if ident, ok := fn.Recv.List[0].Type.(*ast.Ident); ok && ident.Name == structName {
					methods = append(methods, fn.Name.Name)
				}
			}
		}
		return true
	})

	return methods
}

// analyzePythonChanges analyzes Python file changes using regex
func (cd *ChangeDetector) analyzePythonChanges(fileInfo types.FileChange, changes *types.ChangeInfo) {
	fullPath := filepath.Join(cd.projectRoot, fileInfo.Path)

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return
	}

	// Find function definitions
	funcPattern := regexp.MustCompile(`(?m)^def\s+(\w+)\s*\(`)
	matches := funcPattern.FindAllStringSubmatch(string(content), -1)
	for _, match := range matches {
		if len(match) > 1 {
			changes.ModifiedFunctions = append(changes.ModifiedFunctions, types.FunctionChange{
				Name:     match[1],
				File:     fileInfo.Path,
				Type:     "function",
				Language: "python",
			})
		}
	}

	// Find class definitions
	classPattern := regexp.MustCompile(`(?m)^class\s+(\w+)`)
	matches = classPattern.FindAllStringSubmatch(string(content), -1)
	for _, match := range matches {
		if len(match) > 1 {
			changes.ModifiedClasses = append(changes.ModifiedClasses, types.ClassChange{
				Name:     match[1],
				File:     fileInfo.Path,
				Type:     "class",
				Language: "python",
			})
		}
	}
}

// analyzeJSChanges analyzes JavaScript/TypeScript changes using regex
func (cd *ChangeDetector) analyzeJSChanges(fileInfo types.FileChange, changes *types.ChangeInfo) {
	fullPath := filepath.Join(cd.projectRoot, fileInfo.Path)

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return
	}

	// Find function declarations
	funcPattern := regexp.MustCompile(`(?:function\s+(\w+)|(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?(?:function|\([^)]*\)\s*=>))`)
	matches := funcPattern.FindAllStringSubmatch(string(content), -1)
	for _, match := range matches {
		funcName := match[1]
		if funcName == "" {
			funcName = match[2]
		}
		if funcName != "" {
			changes.ModifiedFunctions = append(changes.ModifiedFunctions, types.FunctionChange{
				Name:     funcName,
				File:     fileInfo.Path,
				Type:     "function",
				Language: fileInfo.Language,
			})
		}
	}

	// Find class declarations
	classPattern := regexp.MustCompile(`class\s+(\w+)`)
	matches = classPattern.FindAllStringSubmatch(string(content), -1)
	for _, match := range matches {
		if len(match) > 1 {
			changes.ModifiedClasses = append(changes.ModifiedClasses, types.ClassChange{
				Name:     match[1],
				File:     fileInfo.Path,
				Type:     "class",
				Language: fileInfo.Language,
			})
		}
	}
}

// identifyAffectedModules identifies modules affected by changes
func (cd *ChangeDetector) identifyAffectedModules(changes *types.ChangeInfo) {
	modules := make(map[string]bool)

	for _, fileInfo := range changes.ChangedFiles {
		parts := strings.Split(fileInfo.Path, string(filepath.Separator))
		if len(parts) > 1 {
			// Assume first directory is module name
			modules[parts[0]] = true
		}
	}

	for module := range modules {
		changes.AffectedModules = append(changes.AffectedModules, module)
	}
}

// Made with Bob
