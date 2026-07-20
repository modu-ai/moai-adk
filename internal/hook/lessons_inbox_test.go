package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// readLessonsInbox reads .moai/lessons-inbox.jsonl under root and returns the
// raw parsed stub maps. Used by M3 tests (AC-HRR-008) to assert that failure
// events append structured lesson stubs to the repo-local inbox.
func readLessonsInbox(t *testing.T, root string) []map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".moai", "lessons-inbox.jsonl"))
	if err != nil {
		t.Fatalf("read lessons-inbox.jsonl: %v", err)
	}
	var stubs []map[string]any
	for _, line := range splitLines(string(data)) {
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

// splitLines is a minimal strings.Split wrapper kept local to avoid an extra
// import in the helper.
func splitLines(s string) []string {
	var out []string
	cur := ""
	for _, r := range s {
		if r == '\n' {
			out = append(out, cur)
			cur = ""
			continue
		}
		cur += string(r)
	}
	out = append(out, cur)
	return out
}

// TestRecordToolFailureEvent_AppendsLessonsInboxStub covers AC-HRR-008: when a
// tool_failure event is recorded, a structured lesson stub is appended to
// .moai/lessons-inbox.jsonl carrying at minimum timestamp, event_key, failure
// summary, and source identifier. The line parses as valid JSON.
func TestRecordToolFailureEvent_AppendsLessonsInboxStub(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	t.Setenv(config.EnvClaudeProjectDir, root)

	input := &HookInput{
		SessionID:     "sess-inbox-001",
		ToolName:      "Bash",
		ToolUseID:     "tool-xyz",
		Error:         "exit status 1",
		HookEventName: "PostToolUseFailure",
	}

	h := NewPostToolUseFailureHandler()
	if _, err := h.Handle(context.Background(), input); err != nil {
		t.Fatalf("Handle error: %v", err)
	}

	stubs := readLessonsInbox(t, root)
	if len(stubs) != 1 {
		t.Fatalf("expected 1 lessons-inbox stub, got %d", len(stubs))
	}
	stub := stubs[0]
	// Schema minimum fields (REQ-HRR-006 / D3).
	for _, field := range []string{"timestamp", "event_key", "summary", "source"} {
		if v, ok := stub[field].(string); !ok || v == "" {
			t.Errorf("stub missing/empty field %q: got %v", field, stub[field])
		}
	}
	// event_key should carry the tool_failure prefix.
	if ek, _ := stub["event_key"].(string); ek == "" || ek != "tool_failure:Bash:ExitError" {
		t.Errorf("event_key = %q, want tool_failure:Bash:ExitError", ek)
	}
}

// TestRecordTestFailEvent_AppendsLessonsInboxStub covers AC-HRR-008 for the
// test_fail path: a test-failure event also appends a lesson stub.
func TestRecordTestFailEvent_AppendsLessonsInboxStub(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	t.Setenv(config.EnvClaudeProjectDir, root)

	toolInput, _ := json.Marshal(map[string]string{"command": "go test ./internal/hook/..."})
	toolResponse, _ := json.Marshal(map[string]any{"exit_code": 1})
	input := &HookInput{
		SessionID:    "sess-inbox-002",
		ToolName:     "Bash",
		ToolInput:    toolInput,
		ToolResponse: toolResponse,
	}

	logEvidence(input)

	stubs := readLessonsInbox(t, root)
	if len(stubs) != 1 {
		t.Fatalf("expected 1 lessons-inbox stub, got %d", len(stubs))
	}
	stub := stubs[0]
	for _, field := range []string{"timestamp", "event_key", "summary", "source"} {
		if v, ok := stub[field].(string); !ok || v == "" {
			t.Errorf("stub missing/empty field %q: got %v", field, stub[field])
		}
	}
	if ek, _ := stub["event_key"].(string); ek == "" || ek != "test_fail:internal/hook:" {
		t.Errorf("event_key = %q, want test_fail:internal/hook:", ek)
	}
}

// TestLessonsInbox_AppendOnly verifies EC-3/EC-4: repeated failures append
// multiple stubs (append-only JSONL — no read-modify-write of the whole file).
func TestLessonsInbox_AppendOnly(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	t.Setenv(config.EnvClaudeProjectDir, root)

	for i := 0; i < 3; i++ {
		input := &HookInput{
			SessionID:     "sess-append",
			ToolName:      "Bash",
			Error:         "exit status 1",
			HookEventName: "PostToolUseFailure",
		}
		h := NewPostToolUseFailureHandler()
		if _, err := h.Handle(context.Background(), input); err != nil {
			t.Fatalf("Handle[%d] error: %v", i, err)
		}
	}

	stubs := readLessonsInbox(t, root)
	if len(stubs) != 3 {
		t.Errorf("expected 3 appended stubs (append-only), got %d", len(stubs))
	}
}

// TestTruncateSummary covers the failure-summary helper's edge cases: empty
// error text falls back to the category token, and long text is truncated on a
// rune boundary with an ellipsis.
func TestTruncateSummary(t *testing.T) {
	cases := []struct {
		name      string
		errorText string
		fallback  string
		want      string
	}{
		{"empty falls back to category", "  ", "ExitError", "ExitError"},
		{"short text passes through", "exit status 1", "ExitError", "exit status 1"},
		{"long text truncated", string(make([]rune, 300)), "X", string(make([]rune, 200)) + "…"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := truncateSummary(c.errorText, c.fallback)
			if got != c.want {
				t.Errorf("truncateSummary() = %q, want %q", got, c.want)
			}
		})
	}
}
