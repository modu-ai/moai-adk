package template_test

// TestNoAskUserQuestionInSubagents verifies that .claude/agents/**/*.md files
// do not contain the literal AskUserQuestion( in their body.
//
// SPEC-V3R2-RT-004 AC-11: subagents MUST NOT call AskUserQuestion directly.
// Subagents must report to the orchestrator via BlockerReport.
// On violation, sentinel: ASKUSERQUESTION_IN_SUBAGENT
//
// Rationale: .claude/rules/moai/core/agent-common-protocol.md §User Interaction Boundary.
// Subagents run in isolated stateless contexts where AskUserQuestion does not
// function. Only the orchestrator may use AskUserQuestion.

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoAskUserQuestionInSubagents walks .claude/agents/**/*.md (NOT skills)
// and asserts body contains no literal AskUserQuestion(.
func TestNoAskUserQuestionInSubagents(t *testing.T) {
	t.Parallel()

	// Locate the project root (based on go.mod)
	projectRoot := findProjectRoot(t)
	agentsDir := filepath.Join(projectRoot, ".claude", "agents")

	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		t.Skipf(".claude/agents/ not found at %s; skipping audit", agentsDir)
	}

	// Walk .claude/agents/**/*.md
	err := filepath.WalkDir(agentsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Read file + inspect body (starts after frontmatter)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		body := extractBody(data)
		scanner := bufio.NewScanner(bytes.NewReader(body))
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			if strings.Contains(line, "AskUserQuestion(") {
				// Emit sentinel + fail the test
				rel, _ := filepath.Rel(projectRoot, path)
				t.Errorf(
					"ASKUSERQUESTION_IN_SUBAGENT: %s body line %d contains AskUserQuestion(; "+
						"subagents must use BlockerReport, not AskUserQuestion. "+
						"See agent-common-protocol.md §User Interaction Boundary.",
					rel, lineNum,
				)
			}
		}
		return scanner.Err()
	})

	if err != nil {
		t.Fatalf("WalkDir error: %v", err)
	}
}

// extractBody returns the body following a YAML frontmatter (---...--- block).
// If no frontmatter is present, returns the entire content.
func extractBody(data []byte) []byte {
	content := string(data)
	if !strings.HasPrefix(content, "---") {
		return data
	}

	// Find the second ---
	rest := content[3:] // after the first ---
	idx := strings.Index(rest, "---")
	if idx < 0 {
		return data
	}

	// Content after the second ---
	body := rest[idx+3:]
	return []byte(body)
}

// findProjectRoot walks up to find the directory containing go.mod.
func findProjectRoot(t *testing.T) string {
	t.Helper()
	// Walk upward starting from the current working directory
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found; cannot determine project root")
		}
		dir = parent
	}
}
