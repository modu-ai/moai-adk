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
// contains an actual call-site invocation of AskUserQuestion or mcp__askuser.
//
// Detection strategy (post REQ-TST-011 / SPEC-V3R6-TEST-REFACTOR-001 M4 refinement):
// The guard targets ACTUAL call-site signatures — `AskUserQuestion(` or
// `mcp__askuser(` with an open paren attached. This avoids false-positives when
// the harness emits SPEC template content that legitimately mentions the
// "AskUserQuestion gate" in human-readable prose (e.g., proposalgen scaffolder
// writing a §5.2 Out-of-Scope section describing the orchestrator's
// AskUserQuestion gate semantics — that is documentation OUTPUT, not a call).
//
// What C-HRA-008 forbids is executable Go invocations of the prohibited tools
// from harness/hook source files — not the literal occurrence of the word in
// string literals or comments.
//
// Excluded contexts:
//   - Comment lines (lines starting with optional whitespace + "//")
//   - Test files (*_test.go) — they may reference sentinel names in comments
//     for documentation or guard purposes.
//
// SPEC anchor: SPEC-V3R6-AGENT-TEAM-REBUILD-001 §B.5 (no harness AskUserQuestion
// callsite). M4 of SPEC-V3R6-TEST-REFACTOR-001 refines detection to call-site
// patterns to eliminate false-positive on documentation prose in code.
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

	// Detect ACTUAL call-site patterns (open-paren attached). This matches:
	//   AskUserQuestion(...)
	//   mcp__askuser(...)
	//   client.AskUserQuestion(...)
	//   askuser.AskUserQuestion(...)
	// while NOT matching:
	//   - documentation prose: "the AskUserQuestion gate"
	//   - string literals: "Use AskUserQuestion to ..."
	//   - bare type/identifier references without invocation
	forbiddenCallsite := []string{
		"AskUserQuestion(",
		"mcp__askuser(",
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
				for _, needle := range forbiddenCallsite {
					if strings.Contains(line, needle) {
						rel, _ := filepath.Rel(projectRoot, path)
						violations = append(violations, fmt.Sprintf("%s:%d: non-comment line contains call-site %q", rel, lineNum+1, needle))
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
		t.Errorf("C-HRA-008 VIOLATED — AskUserQuestion/mcp__askuser call-site found in harness/hook source (executable code, not comments or string literals):\n%s",
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
