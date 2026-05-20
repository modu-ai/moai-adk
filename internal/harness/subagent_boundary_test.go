// Package harness_test — M6 subagent boundary test (C-HRA-008).
// Verifies that NO source file under internal/harness/ or internal/hook/ contains
// a direct reference to "AskUserQuestion" or "mcp__askuser".
// This is the binary CI guard for Constraint C-HRA-008 (plan.md §6 Constraints table).
//
// Failure = a future developer accidentally wired AskUserQuestion into the harness pipeline.
package harness_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSubagentBoundary_NoAskUserQuestion is the C-HRA-008 binary CI guard.
// It walks internal/harness/ and internal/hook/ and fails if any .go source file
// contains a non-comment line with the literal strings "AskUserQuestion" or "mcp__askuser".
//
// Comment lines (lines starting with optional whitespace + "//") are excluded because
// all harness/hook packages document their non-invocation in godoc comments (e.g.,
// "// [HARD] No AskUserQuestion calls. Blocker reports are emitted..."). Those comment
// lines are precisely the intended documentation pattern. What C-HRA-008 forbids is
// actual Go call-site invocations in executable code — not explanatory comments.
//
// Test files (*_test.go) are also excluded — they may reference sentinel names
// in comment blocks for documentation or guard purposes.
func TestSubagentBoundary_NoAskUserQuestion(t *testing.T) {
	t.Parallel()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}

	// Walk up to project root (directory containing go.mod).
	projectRoot := findProjectRoot(t, cwd)

	dirsToScan := []string{
		filepath.Join(projectRoot, "internal", "harness"),
		filepath.Join(projectRoot, "internal", "hook"),
	}

	forbidden := []string{
		"AskUserQuestion",
		"mcp__askuser",
	}

	var violations []string

	for _, dir := range dirsToScan {
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".go") {
				return nil
			}
			// Exclude test files.
			if strings.HasSuffix(path, "_test.go") {
				return nil
			}

			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}

			// Scan line by line; skip comment lines.
			for lineNum, line := range strings.Split(string(data), "\n") {
				trimmed := strings.TrimSpace(line)
				// Skip pure comment lines — those document non-invocation intent.
				if strings.HasPrefix(trimmed, "//") {
					continue
				}
				for _, needle := range forbidden {
					if strings.Contains(line, needle) {
						rel, _ := filepath.Rel(projectRoot, path)
						violations = append(violations, fmt.Sprintf("%s:%d: non-comment line contains %q", rel, lineNum+1, needle))
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("WalkDir %s: %v", dir, err)
		}
	}

	if len(violations) > 0 {
		t.Errorf("C-HRA-008 VIOLATED — AskUserQuestion/mcp__askuser found in harness/hook source (executable code, not comments):\n%s",
			strings.Join(violations, "\n"))
	}
}

// findProjectRoot walks upward from startDir until it finds a directory containing go.mod.
func findProjectRoot(t *testing.T, startDir string) string {
	t.Helper()
	dir := startDir
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found walking up from %s", startDir)
		}
		dir = parent
	}
}
