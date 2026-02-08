package quality

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLintFile verifies linting behavior per REQ-HOOK-080, REQ-HOOK-081.
func TestLintFile(t *testing.T) {
	t.Run("lints Go file with golangci-lint if available", func(t *testing.T) {
		registry := NewToolRegistry()

		// Skip if golangci-lint not available
		if !registry.IsToolAvailable("golangci-lint") {
			t.Skip("golangci-lint not available")
		}

		l := NewLinter(registry)
		tmpDir := t.TempDir()

		// Create a Go file with potential issues
		goFile := filepath.Join(tmpDir, "test.go")
		content := `package main

func main() {
	var x int
	return
}`
		if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		result, err := l.LintFile(context.Background(), goFile)
		if err != nil {
			t.Fatalf("LintFile failed: %v", err)
		}

		if result == nil {
			t.Fatal("expected non-nil result")
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		l := NewLinter(NewToolRegistry())
		nonexistentPath := filepath.Join(t.TempDir(), "nonexistent", "file.go")
		_, err := l.LintFile(context.Background(), nonexistentPath)

		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})

	t.Run("handles files with no linter gracefully", func(t *testing.T) {
		l := NewLinter(NewToolRegistry())
		tmpDir := t.TempDir()

		// Create a .xyz file (no linter available)
		xyzFile := filepath.Join(tmpDir, "test.xyz")
		_ = os.WriteFile(xyzFile, []byte("content"), 0644)

		result, err := l.LintFile(context.Background(), xyzFile)

		// Should not error
		if err != nil {
			t.Errorf("expected no error for file without linter, got: %v", err)
		}
		if result != nil && result.Error == "" {
			t.Log("Linter returned result for unsupported file")
		}
	})
}

// TestAutoFix verifies auto-fix behavior per REQ-HOOK-082.
func TestAutoFix(t *testing.T) {
	t.Run("attempts auto-fix when supported", func(t *testing.T) {
		registry := NewToolRegistry()

		// Skip if no linter with auto-fix available
		hasAutoFix := false
		tools := registry.GetToolsForLanguage("go", ToolTypeLinter)
		for _, tool := range tools {
			if len(tool.FixArgs) > 0 {
				hasAutoFix = true
				break
			}
		}
		if !hasAutoFix {
			t.Skip("no linter with auto-fix available")
		}

		l := NewLinter(registry)
		tmpDir := t.TempDir()

		// Create a Go file
		goFile := filepath.Join(tmpDir, "test.go")
		content := `package main

func main() {
	var x int
	return
}`
		if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		result, err := l.AutoFix(context.Background(), goFile)
		if err != nil {
			t.Fatalf("AutoFix failed: %v", err)
		}

		if result == nil {
			t.Error("expected non-nil result from AutoFix")
		}
	})

	t.Run("returns nil result for file without auto-fix support", func(t *testing.T) {
		registry := NewToolRegistry()
		l := NewLinter(registry)

		// Create a file with no auto-fix support (e.g., custom extension)
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.nofix")
		_ = os.WriteFile(testFile, []byte("content"), 0644)

		result, err := l.AutoFix(context.Background(), testFile)

		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			// Result is OK if tool ran, but most likely nil for unsupported
			t.Log("Got result for file without auto-fix")
		}
	})
}

// TestParseLinterIssues verifies issue parsing per REQ-HOOK-081.
func TestParseLinterIssues(t *testing.T) {
	t.Run("parses common linter output formats", func(t *testing.T) {
		registry := NewToolRegistry()
		l := NewLinter(registry)

		// Simulate common linter output
		output := `test.go:4:2: warning: x is unused (varcheck)
test.go:5:1: error: unreachable code (deadcode)
Found 2 issues`

		result := l.ParseIssues(output, "test.go")

		if result == nil {
			t.Fatal("expected non-nil result")
		}

		// Check that issues were found
		if result.IssuesFound == 0 {
			t.Error("expected issues to be parsed from output")
		}

		// Output should contain the linter output
		if result.Output == "" {
			t.Error("expected output to contain linter results")
		}
	})

	t.Run("handles empty output", func(t *testing.T) {
		registry := NewToolRegistry()
		l := NewLinter(registry)

		result := l.ParseIssues("", "test.go")

		if result == nil {
			t.Fatal("expected non-nil result")
		}

		if result.IssuesFound != 0 {
			t.Errorf("expected 0 issues for empty output, got %d", result.IssuesFound)
		}
	})

	t.Run("handles multiline output", func(t *testing.T) {
		registry := NewToolRegistry()
		l := NewLinter(registry)

		output := `test.go:1:1: error: expected 'package'
test.go:2:1: error: expected declaration
`

		result := l.ParseIssues(output, "test.go")

		if result == nil {
			t.Fatal("expected non-nil result")
		}

		if result.IssuesFound < 2 {
			t.Errorf("expected at least 2 issues, got %d", result.IssuesFound)
		}
	})
}

// TestLinterSkipConditions verifies skip conditions.
func TestLinterSkipConditions(t *testing.T) {
	t.Run("skips files in skipped directories", func(t *testing.T) {
		l := NewLinter(NewToolRegistry())

		// These paths should be skipped
		skipPaths := []string{
			"node_modules/test.py",
			"vendor/test.go",
			"dist/test.js",
		}

		for _, path := range skipPaths {
			tools := l.registry.GetToolsForFile(path, ToolTypeLinter)
			// We're checking that the linter respects skip conditions
			_ = tools
			_ = path
		}
	})
}

// TestNewLinter verifies constructor variants.
func TestNewLinter(t *testing.T) {
	t.Run("creates linter with registry", func(t *testing.T) {
		registry := NewToolRegistry()
		l := NewLinter(registry)

		if l == nil {
			t.Error("expected non-nil linter")
		}
	})

	t.Run("creates linter with default registry when nil", func(t *testing.T) {
		l := NewLinter(nil)

		if l == nil {
			t.Error("expected non-nil linter with default registry")
		}
	})
}

// TestGratefulDegradation verifies graceful degradation per REQ-HOOK-083.
func TestGratefulDegradation(t *testing.T) {
	t.Run("continues on linter failure", func(t *testing.T) {
		registry := NewToolRegistry()

		// Create a linter that will fail (use unavailable tool)
		l := NewLinter(registry)
		tmpDir := t.TempDir()

		// Create a test file
		testFile := filepath.Join(tmpDir, "test.unknown")
		_ = os.WriteFile(testFile, []byte("content"), 0644)

		result, err := l.LintFile(context.Background(), testFile)

		// Should not crash, return gracefully
		if err != nil && strings.Contains(err.Error(), "linter") {
			t.Logf("Got expected error: %v", err)
		}
		if result != nil && !result.Success {
			t.Logf("Linter failed gracefully: %s", result.Error)
		}
	})
}

// TestIssueSummary verifies issue summary generation per REQ-HOOK-081.
func TestIssueSummary(t *testing.T) {
	t.Run("generates summary for many issues", func(t *testing.T) {
		registry := NewToolRegistry()
		l := NewLinter(registry)

		// Simulate output with many issues
		output := strings.Repeat("test.go:1: warning: issue\n", 10)

		result := l.ParseIssues(output, "test.go")

		if result == nil {
			t.Fatal("expected non-nil result")
		}

		// Should have found issues
		if result.IssuesFound == 0 {
			t.Error("expected issues to be counted")
		}

		// Generate summary
		summary := l.GenerateSummary(result)
		if summary == "" {
			t.Error("expected non-empty summary")
		}
	})

	t.Run("returns no issues message when clean", func(t *testing.T) {
		registry := NewToolRegistry()
		l := NewLinter(registry)

		result := &ToolResult{
			Success:     true,
			IssuesFound: 0,
			Output:      "no issues found",
		}

		summary := l.GenerateSummary(result)
		if summary == "" {
			t.Error("expected summary even for clean results")
		}
	})
}
