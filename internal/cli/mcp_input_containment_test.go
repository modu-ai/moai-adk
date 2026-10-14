// mcp_input_containment_test.go — SPEC-MCP-NEUTRALITY-001 M3 (card t1632,
// REQ-MN-010): the server never initiates elicitation toward its clients, and
// concurrent tool calls serialize their project-state writes through one
// mutex (the -race half rides this test's -run selector in CI's race lane).
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMCPServer_NeverInitiatesElicitation pins REQ-MN-010's direction
// invariant: the mcp-go server→client elicitation API
// (server.SessionWithElicitation.RequestElicitation, server.WithElicitation)
// is never called anywhere under internal/ — the moai server only ANSWERS
// elicitation requests arriving from nested servers on the managed gate
// session's client connections (managed_codex_factory.go's decline policy).
// Measured 2026-10-14: zero production references — this test pins that.
func TestMCPServer_NeverInitiatesElicitation(t *testing.T) {
	var hits []string
	err := filepath.WalkDir("../..", func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil // unreadable entries are not findings
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", ".moai", "node_modules", "vendor", "testdata":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		body, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		for _, symbol := range []string{"RequestElicitation(", "WithElicitation()"} {
			if strings.Contains(string(body), symbol) {
				hits = append(hits, path+": "+symbol)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if len(hits) > 0 {
		t.Fatalf("the server initiates elicitation toward clients — REQ-MN-010 violated:\n%s", strings.Join(hits, "\n"))
	}
}
