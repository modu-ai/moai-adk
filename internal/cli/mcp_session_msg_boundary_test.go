// mcp_session_msg_boundary_test.go: SPEC-CODEX-SESSION-MSG-001 M2 AC-CSM-009
// (REQ-CSM-011) — the C-HRA-008 / hardcoding static guards for the
// session_msg_* MCP handler source, mirroring mcp_boundary_test.go
// (TestMCP_NoAskUserQuestion / TestMCP_NoInlineGetenv).
//
// The MCP handler runs in a subagent context; AskUserQuestion is
// orchestrator-only. Env-var access (if any) goes through envkeys.go
// constants (AC-CSM-010).
package cli

import (
	"os"
	"strings"
	"testing"
)

// TestSessionMsg_NoAskUserQuestion asserts internal/cli/mcp_session_msg.go
// does NOT reference AskUserQuestion or mcp__askuser.
func TestSessionMsg_NoAskUserQuestion(t *testing.T) {
	src, err := os.ReadFile("mcp_session_msg.go")
	if err != nil {
		t.Fatalf("read mcp_session_msg.go: %v", err)
	}
	if strings.Contains(string(src), "AskUserQuestion") {
		t.Error("internal/cli/mcp_session_msg.go must NOT reference AskUserQuestion (orchestrator-only HARD, REQ-CSM-011)")
	}
	if strings.Contains(string(src), "mcp__askuser") {
		t.Error("internal/cli/mcp_session_msg.go must NOT reference mcp__askuser (orchestrator-only HARD, REQ-CSM-011)")
	}
}

// TestSessionMsg_NoInlineGetenv asserts AC-CSM-010: mcp_session_msg.go
// carries no inline os.Getenv("X") strings.
func TestSessionMsg_NoInlineGetenv(t *testing.T) {
	src, err := os.ReadFile("mcp_session_msg.go")
	if err != nil {
		t.Fatalf("read mcp_session_msg.go: %v", err)
	}
	if strings.Contains(string(src), `os.Getenv("`) {
		t.Errorf("internal/cli/mcp_session_msg.go must NOT carry inline os.Getenv(\"...\") — use envkeys.go constants (AC-CSM-010):\n%s", string(src))
	}
}
