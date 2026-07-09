package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/harness"
)

// readUsageLog reads the harness usage-log.jsonl under root and returns the
// parsed events. Used by M1 tests (AC-HRR-001/002, EC-1) to assert that failure
// events entered the learning loop.
func readUsageLog(t *testing.T, root string) []harness.Event {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".moai", "harness", "usage-log.jsonl"))
	if err != nil {
		t.Fatalf("read usage-log.jsonl: %v", err)
	}
	var events []harness.Event
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var evt harness.Event
		if err := json.Unmarshal([]byte(line), &evt); err == nil {
			events = append(events, evt)
		}
	}
	return events
}

// findEvent returns a pointer to the first event of the given type, or nil.
func findEvent(events []harness.Event, et harness.EventType) *harness.Event {
	for i := range events {
		if events[i].EventType == et {
			return &events[i]
		}
	}
	return nil
}

// eventKey reconstructs the "event_type:subject:context_hash" pattern key that
// AggregatePatterns derives from a usage-log line.
func eventKey(e *harness.Event) string {
	return string(e.EventType) + ":" + e.Subject + ":" + e.ContextHash
}

// TestPostToolFailure_RecordsToolFailureEvent covers AC-HRR-001: the
// PostToolUseFailure handler emits a tool_failure:<tool>:<signature> event to
// usage-log.jsonl through the harness observer. The signature is the
// low-cardinality ErrorCategory token (D1 resolution).
func TestPostToolFailure_RecordsToolFailureEvent(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	t.Setenv(config.EnvClaudeProjectDir, root)

	input := &HookInput{
		SessionID:     "sess-fail-001",
		ToolName:      "Bash",
		ToolUseID:     "tool-123",
		Error:         "exit status 1",
		HookEventName: "PostToolUseFailure",
	}

	h := NewPostToolUseFailureHandler()
	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle error: %v", err)
	}
	if out == nil {
		t.Fatal("nil output")
	}

	events := readUsageLog(t, root)
	found := findEvent(events, harness.EventTypeToolFailure)
	if found == nil {
		t.Fatalf("no tool_failure event recorded; events=%v", events)
	}
	// AC-HRR-001: key form tool_failure:Bash:ExitError (signature = error-class token).
	wantKey := "tool_failure:Bash:ExitError"
	if got := eventKey(found); got != wantKey {
		t.Errorf("event key = %q, want %q", got, wantKey)
	}
}

// TestPostToolFailure_EmptyToolName_NonDegenerateKey covers EC-1: an empty tool
// name MUST NOT produce a degenerate tool_failure::<sig> key (which AC-HRR-004
// would later exclude). The handler substitutes a non-empty placeholder so the
// subject slot is never empty while the signature stays non-empty.
func TestPostToolFailure_EmptyToolName_NonDegenerateKey(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	t.Setenv(config.EnvClaudeProjectDir, root)

	input := &HookInput{
		SessionID:     "sess-ec1",
		ToolName:      "", // empty — EC-1
		Error:         "exit status 2",
		HookEventName: "PostToolUseFailure",
	}

	h := NewPostToolUseFailureHandler()
	if _, err := h.Handle(context.Background(), input); err != nil {
		t.Fatalf("Handle error: %v", err)
	}

	events := readUsageLog(t, root)
	found := findEvent(events, harness.EventTypeToolFailure)
	if found == nil {
		t.Fatalf("no tool_failure event recorded; events=%v", events)
	}
	if found.Subject == "" {
		t.Error("Subject (tool) is empty — degenerate key tool_failure::<sig>")
	}
	if found.ContextHash == "" {
		t.Error("ContextHash (signature) is empty")
	}
}

// TestEvidenceWriter_RecordsTestFailEvent covers AC-HRR-002: when the evidence
// writer detects a test failure (isFail), it emits a
// test_fail:<package>: event to usage-log.jsonl. The package is derived from
// the test command (go test ./internal/hook/... → internal/hook).
func TestEvidenceWriter_RecordsTestFailEvent(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	t.Setenv(config.EnvClaudeProjectDir, root)

	toolInput, _ := json.Marshal(map[string]string{"command": "go test ./internal/hook/..."})
	toolResponse, _ := json.Marshal(map[string]any{"exit_code": 1})

	input := &HookInput{
		SessionID:    "sess-test-001",
		ToolName:     "Bash",
		ToolInput:    toolInput,
		ToolResponse: toolResponse,
	}

	logEvidence(input)

	events := readUsageLog(t, root)
	found := findEvent(events, harness.EventTypeTestFail)
	if found == nil {
		t.Fatalf("no test_fail event recorded; events=%v", events)
	}
	// AC-HRR-002: key form test_fail:internal/hook: (empty context hash is by design).
	wantKey := "test_fail:internal/hook:"
	if got := eventKey(found); got != wantKey {
		t.Errorf("event key = %q, want %q", got, wantKey)
	}
}

// TestExtractTestPackage covers the package-extraction helper used by the
// test_fail recording path. Low-cardinality package identifiers keep the pattern
// keys aggregatable.
func TestExtractTestPackage(t *testing.T) {
	cases := []struct {
		command string
		want    string
	}{
		{"go test ./internal/hook/...", "internal/hook"},
		{"go test ./internal/hook", "internal/hook"},
		{"go test ./...", ""},
		{"go test", ""},
		{"ls -la", ""},
	}
	for _, c := range cases {
		got := extractTestPackage(c.command)
		if got != c.want {
			t.Errorf("extractTestPackage(%q) = %q, want %q", c.command, got, c.want)
		}
	}
}
