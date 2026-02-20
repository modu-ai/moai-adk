package quality

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Package-level compiled regexps to avoid repeated compilation.
var (
	reIssuesSummary = regexp.MustCompile(`(\d+)\s+issues?\s+found`)
	reFileLineCol   = regexp.MustCompile(`^.+:\d+:\d+:\s+\w+`)
	reFileLine      = regexp.MustCompile(`^.+:\d+:\s+\w+`)
	reFixedSummary  = regexp.MustCompile(`(\d+)\s+(?:issues?\s+)?(?:fixed|corrected|resolved)`)
)

// Linter handles automatic code linting per REQ-HOOK-080.
type Linter struct {
	registry *toolRegistry
}

// NewLinter creates a new Linter with default registry.
func NewLinter(registry *toolRegistry) *Linter {
	if registry == nil {
		registry = NewToolRegistry()
	}
	return &Linter{
		registry: registry,
	}
}

// LintFile runs linter on a file per REQ-HOOK-080.
func (l *Linter) LintFile(ctx context.Context, filePath string) (*ToolResult, error) {
	// Check if file exists before processing
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	// Get linters for this file
	tools := l.registry.GetToolsForFile(filePath, ToolTypeLinter)
	if len(tools) == 0 {
		return &ToolResult{Success: true, IssuesFound: 0}, nil
	}

	// Get directory for command execution
	cwd := filepath.Dir(filePath)
	if cwd == "" {
		cwd = "."
	}

	// Try each linter in priority order
	for _, tool := range tools {
		// Check if tool is available
		if !l.registry.IsToolAvailable(tool.Command) {
			continue
		}

		// Run the linter
		result := l.registry.RunTool(tool, filePath, cwd)

		// Check for file-not-found type errors
		if result.Error != "" && (strings.Contains(result.Error, "no such file") ||
			strings.Contains(result.Error, "cannot find") ||
			strings.Contains(result.Error, "does not exist")) {
			return nil, errors.New(result.Error)
		}

		// Parse issues from output
		if result.Success || result.ExitCode != 0 {
			issues := l.parseIssuesFromOutput(result.Output, filePath)
			result.IssuesFound = issues
		}

		return &result, nil
	}

	return &ToolResult{Success: true, IssuesFound: 0}, nil
}

// AutoFix attempts to auto-fix linting issues per REQ-HOOK-082.
func (l *Linter) AutoFix(ctx context.Context, filePath string) (*ToolResult, error) {
	// Get linters for this file
	tools := l.registry.GetToolsForFile(filePath, ToolTypeLinter)
	if len(tools) == 0 {
		return nil, nil
	}

	// Get directory for command execution
	cwd := filepath.Dir(filePath)
	if cwd == "" {
		cwd = "."
	}

	// Try each linter that supports auto-fix
	for _, tool := range tools {
		// Check if tool is available and supports auto-fix
		if !l.registry.IsToolAvailable(tool.Command) {
			continue
		}
		if len(tool.FixArgs) == 0 {
			continue
		}

		// Create fix tool args
		fixTool := tool
		fixTool.Args = append(fixTool.Args, fixTool.FixArgs...)

		// Run the linter with fix args
		result := l.registry.RunTool(fixTool, filePath, cwd)

		if result.Success {
			// Count issues that were fixed
			result.IssuesFixed = l.countFixedIssues(result.Output)
		}

		return &result, nil
	}

	return nil, nil
}

// ParseIssues parses linter output and returns a ToolResult with issue count.
func (l *Linter) ParseIssues(output, filePath string) *ToolResult {
	result := &ToolResult{
		Output: output,
	}
	result.IssuesFound = l.parseIssuesFromOutput(output, filePath)
	return result
}

// GenerateSummary generates a summary message for linter results per REQ-HOOK-081.
func (l *Linter) GenerateSummary(result *ToolResult) string {
	if result == nil {
		return "No linter results available"
	}

	if result.IssuesFound == 0 {
		return "No issues found"
	}

	if result.IssuesFound <= 5 {
		return result.Output
	}

	// Truncate to first 5 issues
	lines := strings.Split(result.Output, "\n")
	issueCount := 0
	var summaryLines []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if l.isIssueLine(line) {
			issueCount++
			if issueCount <= 5 {
				summaryLines = append(summaryLines, line)
			}
		}
		if issueCount > 5 {
			summaryLines = append(summaryLines, "")
			summaryLines = append(summaryLines, "... and more")
			break
		}
	}

	return strings.Join(summaryLines, "\n")
}

// parseIssuesFromOutput counts issues in linter output.
func (l *Linter) parseIssuesFromOutput(output, filePath string) int {
	if output == "" {
		return 0
	}

	// Common patterns for linter issues:
	// - file:line:column: message
	// - file:line: message
	// - "error:" or "warning:" prefixes
	baseName := filepath.Base(filePath)

	// Count issue lines
	count := 0
	lines := strings.SplitSeq(output, "\n")
	for line := range lines {
		if l.isIssueLine(line) {
			// Check if issue is for our file
			if strings.Contains(line, baseName) || !strings.Contains(line, ":") {
				count++
			}
		}
	}

	// Also check for summary patterns like "Found 3 issues"
	matches := reIssuesSummary.FindStringSubmatch(output)
	if len(matches) > 1 {
		// Use the count from summary if available
		return count
	}

	return count
}

// isIssueLine checks if a line looks like a linter issue.
func (l *Linter) isIssueLine(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}

	// Check for common issue patterns
	issuePatterns := []string{
		": error:",
		": warning:",
		": note:",
		" ERR ",
		" WARN ",
		"error:",
		"warning:",
	}

	for _, pattern := range issuePatterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}

	// Check for file:line:col pattern
	if reFileLineCol.MatchString(line) {
		return true
	}

	return reFileLine.MatchString(line)
}

// countFixedIssues estimates how many issues were fixed by auto-fix.
func (l *Linter) countFixedIssues(output string) int {
	// Look for patterns like "fixed 3 issues", "3 problems fixed", etc.
	matches := reFixedSummary.FindStringSubmatch(output)
	if len(matches) > 1 {
		count := 0
		for _, c := range matches[1] {
			count = count*10 + int(c-'0')
		}
		return count
	}

	// Fallback: count lines with "fixed" or similar
	count := 0
	lines := strings.SplitSeq(output, "\n")
	for line := range lines {
		if strings.Contains(strings.ToLower(line), "fixed") {
			count++
		}
	}
	return count
}
