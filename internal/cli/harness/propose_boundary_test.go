// Package harness CLI boundary guard test.
// SPEC: SPEC-V3R6-HARNESS-PROPOSAL-GEN-001 REQ-PGN-012, REQ-PGN-013.
//
// This test mirrors the structural pattern of TestNew_NoAskUserQuestion in
// internal/cli/worktree/new_test.go. It scans every non-test source file in
// the internal/cli/harness/ subpackage and asserts the absolute absence of
// the AskUserQuestion invocation pattern.
//
// HARD subagent boundary: the CLI MUST NOT prompt the user. AskUserQuestion
// belongs exclusively to the MoAI orchestrator. The orchestrator consumes
// this CLI's JSON output and presents the Approve / Modify / Reject gate
// when auto_delegate is true.
package harness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPropose_NoAskUserQuestion guards against subagent-boundary regression
// for every source file under internal/cli/harness/. The static substring
// scan catches both function calls (`AskUserQuestion(...)`) and the
// canonical parenthesized comment pattern (`// AskUserQuestion(`).
//
// Prose comments that DESCRIBE the orchestrator's AskUserQuestion behavior
// without the parenthesized invocation form remain permitted — they read as
// "the orchestrator presents AskUserQuestion after this CLI returns" and do
// not match `AskUserQuestion(`.
func TestPropose_NoAskUserQuestion(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	var scanned int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(filepath.Join(".", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if strings.Contains(string(src), "AskUserQuestion(") {
			t.Errorf("internal/cli/harness/%s contains forbidden substring `AskUserQuestion(` "+
				"(REQ-PGN-012/013 subagent boundary HARD)", name)
		}
		scanned++
	}

	if scanned == 0 {
		t.Fatal("TestPropose_NoAskUserQuestion scanned 0 source files; CI guard is ineffective. " +
			"At least propose.go must be present in internal/cli/harness/.")
	}
}
