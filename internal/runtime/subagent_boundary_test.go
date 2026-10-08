package runtime

import (
	"os"
	"strings"
	"testing"
)

// TestNoInteractivePrompt is AC-ACE-016's static guard (C-HRA-008 family,
// serving REQ-ACE-007's no-interactive-prompt clause): the runtime package
// hosts the ceiling engine the CLI-side admission seams call, and none of
// its surface may carry an interactive prompt — the CLI runs in subagent
// context, and any operator decision arrives through the orchestrator, not
// through the CLI (SPEC-AUDIT-CEILING-001 C3). Comment-only mentions are
// excluded, matching the canonical subagent-boundary grep surface.
func TestNoInteractivePrompt(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	scanned := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		scanned++
		for i, line := range strings.Split(string(src), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") {
				continue
			}
			if strings.Contains(line, "AskUserQuestion") || strings.Contains(line, "mcp__askuser__") {
				t.Errorf("%s:%d: interactive prompt surface in the CLI-side runtime package", name, i+1)
			}
		}
	}
	if scanned == 0 {
		t.Fatal("no source files scanned — the guard swept nothing")
	}
}
