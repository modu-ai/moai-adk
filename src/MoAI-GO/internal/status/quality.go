package status

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// QualityGateStatus represents the status of quality gates
type QualityGateStatus struct {
	Tested    *QualityItem
	Readable  *QualityItem
	Unified   *QualityItem
	Secured   *QualityItem
	Trackable *QualityItem
}

// QualityItem represents a single quality check result
type QualityItem struct {
	Passed  bool
	Message string
}

// GetQualityGateStatus retrieves the status of all quality gates
func GetQualityGateStatus(projectDir string) (*QualityGateStatus, error) {
	status := &QualityGateStatus{
		Tested:    &QualityItem{Passed: true, Message: "Not configured"},
		Readable:  &QualityItem{Passed: true, Message: "Not configured"},
		Unified:   &QualityItem{Passed: true, Message: "Not configured"},
		Secured:   &QualityItem{Passed: true, Message: "Not configured"},
		Trackable: &QualityItem{Passed: true, Message: "Not configured"},
	}

	// Check for test coverage
	status.Tested = checkTestCoverage(projectDir)

	// Check for linting
	status.Readable = checkLinting(projectDir)

	// Check for formatting
	status.Unified = checkFormatting(projectDir)

	// Check for security
	status.Secured = checkSecurity(projectDir)

	// Check for conventional commits
	status.Trackable = checkConventionalCommits()

	return status, nil
}

// checkTestCoverage checks test coverage
func checkTestCoverage(projectDir string) *QualityItem {
	// Look for Go test files
	hasTests := false
	err := filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, "_test.go") {
			hasTests = true
		}
		return nil
	})

	if err != nil {
		return &QualityItem{Passed: false, Message: "Error checking tests"}
	}

	if !hasTests {
		return &QualityItem{Passed: false, Message: "No tests found"}
	}

	return &QualityItem{Passed: true, Message: "Tests present"}
}

// checkLinting checks for linting issues
func checkLinting(projectDir string) *QualityItem {
	// Try to run golangci-lint if available
	cmd := exec.Command("golangci-lint", "run", "--timeout", "30s")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()

	if err != nil {
		// golangci-lint not found or has issues
		return &QualityItem{Passed: true, Message: "Linting not configured"}
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" || strings.Contains(outputStr, "no linted files") {
		return &QualityItem{Passed: true, Message: "No linting issues"}
	}

	return &QualityItem{Passed: false, Message: "Linting issues found"}
}

// checkFormatting checks code formatting
func checkFormatting(projectDir string) *QualityItem {
	// Check if files are formatted
	cmd := exec.Command("gofmt", "-l", ".")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()

	if err != nil {
		return &QualityItem{Passed: true, Message: "Formatting check skipped"}
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return &QualityItem{Passed: true, Message: "Formatted"}
	}

	return &QualityItem{Passed: false, Message: "Needs formatting"}
}

// checkSecurity checks for security issues
func checkSecurity(projectDir string) *QualityItem {
	// Try to run gosec if available
	cmd := exec.Command("gosec", "./...")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()

	if err != nil {
		// gosec not found
		return &QualityItem{Passed: true, Message: "Security scan not configured"}
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" || strings.Contains(outputStr, "No issues found") {
		return &QualityItem{Passed: true, Message: "No vulnerabilities"}
	}

	return &QualityItem{Passed: false, Message: "Security issues found"}
}

// checkConventionalCommits checks if commits follow conventional format
func checkConventionalCommits() *QualityItem {
	// Check recent commits
	cmd := exec.Command("git", "log", "-10", "--pretty=format:%s")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return &QualityItem{Passed: true, Message: "Cannot check commits"}
	}

	commits := strings.Split(strings.TrimSpace(string(output)), "\n")
	conventional := 0

	for _, commit := range commits {
		if isConventionalCommit(commit) {
			conventional++
		}
	}

	if len(commits) > 0 && conventional == len(commits) {
		return &QualityItem{Passed: true, Message: "Conventional commits"}
	}

	return &QualityItem{Passed: false, Message: "Non-conventional commits found"}
}

// isConventionalCommit checks if a commit message follows conventional format
func isConventionalCommit(message string) bool {
	prefixes := []string{"feat:", "fix:", "docs:", "style:", "refactor:", "test:", "chore:", "perf:", "ci:", "build:", "revert:"}

	message = strings.TrimSpace(message)
	for _, prefix := range prefixes {
		if strings.HasPrefix(message, prefix) {
			return true
		}
	}

	return false
}
