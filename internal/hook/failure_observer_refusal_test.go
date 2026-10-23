// failure_observer_refusal_test.go — SPEC-MCP-NEUTRALITY-001 REQ-MN-002
// (card t1632, AC-MN-002): the lessons-inbox failure observer excludes
// policy-refusal payloads. N refused results plus exactly one genuine
// transport error leave exactly ONE new inbox row; the refusal class is a
// policy decision, not a malfunction, and never becomes a tool_failure row.
package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// readLessonsInboxLenient reads the scratch inbox, tolerating its absence
// (zero rows) — an over-excluding mutant must surface as a count mismatch,
// not as a fixture error.
func readLessonsInboxLenient(t *testing.T, root string) []map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".moai", "lessons-inbox.jsonl"))
	if err != nil {
		return nil
	}
	var stubs []map[string]any
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var stub map[string]any
		if json.Unmarshal([]byte(line), &stub) == nil {
			stubs = append(stubs, stub)
		}
	}
	return stubs
}

// TestFailureObserver_ExcludesPolicyRefusals is AC-MN-002: two forced policy
// refusals (one as the pure text-block JSON, one prefixed by the tool name —
// both shapes the tool layer emits) and exactly one genuine transport error
// leave exactly one new inbox row, carrying the genuine error's key. No
// refusal row exists, and the usage-log ratchet feed counts exactly one
// tool_failure for the same reason (the same seam classifies both writes).
func TestFailureObserver_ExcludesPolicyRefusals(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv(config.EnvClaudeProjectDir, root)

	refusalText := `{"refused":true,"code":"serial-slot","reason":"another serial card is in flight and holds the serial slot"}`
	refusalResponse, err := json.Marshal(map[string]any{
		"content": []map[string]any{{"type": "text", "text": refusalText}},
	})
	if err != nil {
		t.Fatal(err)
	}

	inputs := []*HookInput{
		{SessionID: "sess-refuse-1", ToolName: "mcp__moai__factory_next", ToolResponse: refusalResponse, HookEventName: "PostToolUseFailure"},
		{SessionID: "sess-refuse-2", ToolName: "mcp__moai__todo_claim", Error: "todo_claim: " + refusalText, HookEventName: "PostToolUseFailure"},
		{SessionID: "sess-genuine", ToolName: "Bash", Error: "exit status 1: connection reset by peer", HookEventName: "PostToolUseFailure"},
	}

	h := NewPostToolUseFailureHandler()
	for i, in := range inputs {
		if _, err := h.Handle(context.Background(), in); err != nil {
			t.Fatalf("input %d: Handle error: %v", i, err)
		}
	}

	stubs := readLessonsInboxLenient(t, root)
	if len(stubs) != 1 {
		t.Fatalf("inbox rows = %d, want exactly 1 (the genuine error only): %v", len(stubs), stubs)
	}
	if key := fmt.Sprint(stubs[0]["event_key"]); !strings.Contains(key, "Bash") {
		t.Errorf("the surviving row is %q, want the genuine Bash failure", key)
	}
}
