// mcp_boundary_test.go: SPEC-TREND-MCP-001 M2 AC-TMC-010 — the C-HRA-008 /
// REQ-TMC-010 subagent-boundary static guard for the new `moai mcp` CLI source.
//
// The CLI runs in a subagent context; AskUserQuestion is orchestrator-only.
// This test mirrors internal/cli/web_test.go TestWeb_NoAskUserQuestion (the
// canonical static-guard pattern).
package cli

import (
	"os"
	"strings"
	"testing"
)

// TestMCP_NoAskUserQuestion asserts internal/cli/mcp.go does NOT reference
// AskUserQuestion or mcp__askuser.
func TestMCP_NoAskUserQuestion(t *testing.T) {
	t.Parallel()
	src, err := os.ReadFile("mcp.go")
	if err != nil {
		t.Fatalf("read mcp.go: %v", err)
	}
	if strings.Contains(string(src), "AskUserQuestion") {
		t.Error("internal/cli/mcp.go must NOT reference AskUserQuestion (orchestrator-only HARD)")
	}
	if strings.Contains(string(src), "mcp__askuser") {
		t.Error("internal/cli/mcp.go must NOT reference mcp__askuser (orchestrator-only HARD)")
	}
}

// TestMCP_NoInlineGetenv asserts AC-TMC-015: every env-var reference (if any)
// goes through constants. The new mcp.go MUST NOT carry inline os.Getenv("X")
// strings.
func TestMCP_NoInlineGetenv(t *testing.T) {
	t.Parallel()
	src, err := os.ReadFile("mcp.go")
	if err != nil {
		t.Fatalf("read mcp.go: %v", err)
	}
	if strings.Contains(string(src), `os.Getenv("`) {
		t.Errorf("internal/cli/mcp.go must NOT carry inline os.Getenv(\"...\") — use envkeys.go constants (AC-TMC-015):\n%s", string(src))
	}
}
