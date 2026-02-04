package quality

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk-go/internal/hook"
)

// Linter handles automatic code linting for PostToolUse hooks.
// It lints files after Write/Edit operations using language-specific tools.
type Linter struct {
	registry *ToolRegistry
}

// NewLinter creates a new Linter with default registry.
func NewLinter() *Linter {
	return &Linter{
		registry: NewToolRegistry(),
	}
}

// NewLinterWithRegistry creates a Linter using the provided registry.
func NewLinterWithRegistry(registry *ToolRegistry) *Linter {
	return &Linter{
		registry: registry,
	}
}

// LintFile lints a single file using the appropriate linter.
// Returns the tool result or nil if no linter is available.
func (l *Linter) LintFile(ctx context.Context, filePath string, cwd string) (*ToolResult, error) {
	// Check if we should skip this file
	if !l.ShouldLint(filePath) {
		return nil, nil
	}

	// Get linters for this file
	linters := l.registry.GetToolsForFile(filePath, ToolTypeLinter)
	if len(linters) == 0 {
		return nil, nil
	}

	// Try linters in priority order
	for _, linterConfig := range linters {
		// Check if tool is available
		if !l.registry.IsToolAvailable(linterConfig.Name) {
			continue
		}

		// Run the linter
		result := l.registry.RunTool(ctx, linterConfig, filePath, cwd)
		result.ToolName = linterConfig.Name

		// Parse issues from output
		issues := l.parseIssues(result.Output, result.Error)
		result.IssuesFound = len(issues)

		return &result, nil
	}

	// No linter available
	return nil, nil
}

// AutoFix runs a linter with auto-fix enabled on a file.
func (l *Linter) AutoFix(ctx context.Context, filePath string, cwd string) (*ToolResult, error) {
	// Check if we should skip
	if !l.ShouldLint(filePath) {
		return nil, nil
	}

	// Get linters with fix support
	linters := l.registry.GetToolsForFile(filePath, ToolTypeLinter)
	if len(linters) == 0 {
		return nil, nil
	}

	// Try linters that have fix args
	for _, linterConfig := range linters {
		if !l.registry.IsToolAvailable(linterConfig.Name) {
			continue
		}

		// Run the linter (fix args are built into the tool config)
		result := l.registry.RunTool(ctx, linterConfig, filePath, cwd)
		result.ToolName = linterConfig.Name

		return &result, nil
	}

	return nil, nil
}

// ShouldLint checks if a file should be linted based on skip patterns.
func (l *Linter) ShouldLint(filePath string) bool {
	// Get file extension
	ext := strings.ToLower(filepath.Ext(filePath))

	// Check skip extensions
	if SkipExtensions[ext] {
		return false
	}

	// Check for minified files
	if strings.Contains(filepath.Base(filePath), ".min.") {
		return false
	}

	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		return false
	}

	// Check skip directories
	for _, dir := range l.getParentDirectories(filePath) {
		if SkipDirectories[dir] {
			return false
		}
	}

	return true
}

// getParentDirectories returns all parent directory names for a file path.
func (l *Linter) getParentDirectories(filePath string) []string {
	var dirs []string
	path := filePath

	for {
		dir := filepath.Base(path)
		if dir == "." || dir == "/" || dir == "" {
			break
		}
		dirs = append(dirs, dir)

		parent := filepath.Dir(path)
		if parent == path || parent == "." || parent == "/" {
			break
		}
		path = parent
	}

	return dirs
}

// parseIssues parses linter output to extract individual issues.
func (l *Linter) parseIssues(stdout, stderr string) []Issue {
	var issues []Issue

	combined := stdout + "\n" + stderr
	if strings.TrimSpace(combined) == "" {
		return issues
	}

	// Common patterns for issues
	// Format: file:line:col: message (many linters)
	reLocation := regexp.MustCompile(`^([^:]+):(\d+):(\d+):\s*(.+)$`)
	// Format: error: message (ruff, mypy)
	reError := regexp.MustCompile(`error:\s*(.+)`)
	// Format: warning: message
	reWarning := regexp.MustCompile(`warning:\s*(.+)`)

	lines := strings.Split(combined, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip non-issue lines
		if l.isSkipLine(line) {
			continue
		}

		// Try to parse as location-based issue
		if matches := reLocation.FindStringSubmatch(line); len(matches) >= 5 {
			issues = append(issues, Issue{
				File:    matches[1],
				Line:    matches[2],
				Column:  matches[3],
				Message: matches[4],
				Raw:     line,
			})
			continue
		}

		// Try to parse as error
		if matches := reError.FindStringSubmatch(line); len(matches) >= 2 {
			issues = append(issues, Issue{
				Severity: "error",
				Message:  matches[1],
				Raw:      line,
			})
			continue
		}

		// Try to parse as warning
		if matches := reWarning.FindStringSubmatch(line); len(matches) >= 2 {
			issues = append(issues, Issue{
				Severity: "warning",
				Message:  matches[1],
				Raw:      line,
			})
			continue
		}

		// Generic issue - include raw line
		if len(issues) < MaxIssuesToReport {
			issues = append(issues, Issue{
				Message: l.truncateMessage(line, 200),
				Raw:     line,
			})
		}

		// Limit issues
		if len(issues) >= MaxIssuesToReport {
			break
		}
	}

	return issues
}

// isSkipLine checks if a line should be skipped during issue parsing.
func (l *Linter) isSkipLine(line string) bool {
	lowerLine := strings.ToLower(line)

	skipPatterns := []string{
		"running",
		"checking",
		"finished",
		"success",
		"found 0 errors",
		"no issues found",
		"all checks passed",
		"warning:", // Already handled by regex
		"info:",
	}

	for _, pattern := range skipPatterns {
		if strings.Contains(lowerLine, pattern) {
			return true
		}
	}

	// Skip lines that look like progress indicators
	if strings.HasPrefix(line, "✔") || strings.HasPrefix(line, "✓") {
		return true
	}

	return false
}

// truncateMessage truncates a message to a maximum length.
func (l *Linter) truncateMessage(msg string, maxLen int) string {
	if len(msg) <= maxLen {
		return msg
	}
	return msg[:maxLen] + "..."
}

// Issue represents a single lint issue.
type Issue struct {
	File     string
	Line     string
	Column   string
	Severity string
	Message  string
	Raw      string
}

// FormatIssueSummary formats a list of issues into a summary string.
func (l *Linter) FormatIssueSummary(issues []Issue) string {
	if len(issues) == 0 {
		return ""
	}

	var parts []string
	for i, issue := range issues {
		if i >= 3 {
			remaining := len(issues) - i
			if remaining > 0 {
				parts = append(parts, fmt.Sprintf("(+%d more)", remaining))
			}
			break
		}

		part := issue.Message
		if issue.Line != "" {
			part = fmt.Sprintf("Line %s: %s", issue.Line, issue.Message)
		}
		parts = append(parts, part)
	}

	return strings.Join(parts, "; ")
}

// LintForHook lints a file and returns a HookOutput for PostToolUse.
// This is the main integration point with the hook system.
func (l *Linter) LintForHook(ctx context.Context, filePath string, cwd string) *hook.HookOutput {
	result, err := l.LintFile(ctx, filePath, cwd)
	if err != nil {
		return &hook.HookOutput{
			HookSpecificOutput: &hook.HookSpecificOutput{
				HookEventName:     "PostToolUse",
				AdditionalContext: fmt.Sprintf("Linter error: %v", err),
			},
		}
	}

	if result == nil {
		// File was skipped or no linter available
		return hook.NewSuppressOutput()
	}

	// If issues were auto-fixed, report that
	if result.FileModified {
		return &hook.HookOutput{
			HookSpecificOutput: &hook.HookSpecificOutput{
				HookEventName:     "PostToolUse",
				AdditionalContext: fmt.Sprintf("Lint: %d issues auto-fixed with %s", result.IssuesFixed, result.ToolName),
			},
		}
	}

	// If issues found but not fixed, provide summary
	if result.IssuesFound > 0 {
		issues := l.parseIssues(result.Output, result.Error)
		summary := l.FormatIssueSummary(issues)

		if summary != "" {
			return &hook.HookOutput{
				HookSpecificOutput: &hook.HookSpecificOutput{
					HookEventName:     "PostToolUse",
					AdditionalContext: fmt.Sprintf("Lint issues found: %s", summary),
				},
			}
		}
	}

	// No issues found or linter ran successfully
	return hook.NewSuppressOutput()
}

// EventType implements hook.Handler interface.
func (l *Linter) EventType() hook.EventType {
	return hook.EventPostToolUse
}

// Handle implements hook.Handler interface for PostToolUse events.
func (l *Linter) Handle(ctx context.Context, input *hook.HookInput) (*hook.HookOutput, error) {
	// Only process Write and Edit tools
	if input.ToolName != "Write" && input.ToolName != "Edit" {
		return hook.NewAllowOutput(), nil
	}

	// Extract file path from tool input
	filePath, err := extractFilePath(input)
	if err != nil || filePath == "" {
		return hook.NewAllowOutput(), nil
	}

	// Lint the file
	output := l.LintForHook(ctx, filePath, input.CWD)

	return output, nil
}
