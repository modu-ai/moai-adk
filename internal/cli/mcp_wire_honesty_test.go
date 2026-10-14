// mcp_wire_honesty_test.go — SPEC-MCP-NEUTRALITY-001 M2 (card t1632,
// REQ-MN-008 / AC-MN-010's M2 half): a structured-result tool's TEXT block
// carries the SAME JSON as its structured content, and a collection field
// with nothing to return serializes as [] — never null.
package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// TestFactoryMsgList_EmptyIsArrayAndTextMirrors pins AC-MN-010's M2 half:
// factory_msg_list with zero eligible targets answers "messages": [] (a nil
// slice would marshal null), and the text content block carries the same JSON
// the structured content carries.
func TestFactoryMsgList_EmptyIsArrayAndTextMirrors(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root, run := t.TempDir(), "list-empty-run"
	initGitRepo(t, root)
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	activateManagedRun(t, root, run)
	t.Setenv(config.EnvClaudeProjectDir, root)

	// The attributed endpoint: a registered leader owned by this process, so
	// attribution resolves and the claim finds zero eligible targets.
	store, err := factorymsg.Open(root, run)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	start := homestate.CurrentProcessFingerprint()
	if start == "" {
		t.Fatal("test process identity unavailable")
	}
	leader := factorymsg.Peer{ProjectKey: homestate.ProjectKey(root), RunID: run, Backend: "claude",
		Role: "leader", Slot: "leader", SessionUUID: "list-empty-leader", Generation: 1,
		PID: os.Getpid(), ProcessStart: start}
	if _, err := store.RegisterPeer(context.Background(), leader); err != nil {
		t.Fatal(err)
	}
	t.Setenv(config.EnvMoaiSessionPID, strconv.Itoa(os.Getpid()))
	t.Setenv(config.EnvClaudeCodeSessionID, "")

	res := mcpCallRaw(t, handleFactoryMsgList, map[string]any{
		"run_id": run, "project_root": root, "actor": "list-empty-leader",
	})
	if res.IsError {
		t.Fatalf("empty list refused: %s", mcpTextOf(res))
	}
	structured, mErr := json.Marshal(res.StructuredContent)
	if mErr != nil {
		t.Fatalf("marshal structured content: %v", mErr)
	}
	if strings.Contains(string(structured), `"messages":null`) {
		t.Errorf("empty claim set serialized as null: %s", structured)
	}
	if !strings.Contains(string(structured), `"messages":[]`) {
		t.Errorf("empty claim set is not an empty array: %s", structured)
	}
	if mcpTextOf(res) != string(structured) {
		t.Errorf("text block %q does not mirror the structured content %q", mcpTextOf(res), string(structured))
	}
}
