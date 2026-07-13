package verify

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSubagentBoundary is the CI guard for the subagent-boundary discipline:
// hook-domain code must never invoke the orchestrator-only AskUserQuestion
// channel. Scans every non-test Go file in this package, excluding comment
// lines (comments legitimately DOCUMENT the boundary, matching the canonical
// grep's comment filter).
func TestSubagentBoundary(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Clean(name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for i, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "//") {
				continue
			}
			for _, forbidden := range []string{"AskUserQuestion", "mcp__askuser"} {
				if strings.Contains(line, forbidden) {
					t.Errorf("%s:%d references %q — subagent-boundary violation", name, i+1, forbidden)
				}
			}
		}
	}
}
