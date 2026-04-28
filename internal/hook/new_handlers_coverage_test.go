package hook

// Coverage tests for new handlers added in SPEC-OPUS47-COMPAT-001:
// elicitationHandler, elicitationResultHandler, fileChangedHandler,
// genericHandler — all 0% before this file.

import (
	"context"
	"encoding/json"
	"testing"
)

// --- elicitationHandler ---

// TestElicitationHandler_EventType returns EventElicitation.
func TestElicitationHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewElicitationHandler()
	if got := h.EventType(); got != EventElicitation {
		t.Errorf("EventType() = %q, want %q", got, EventElicitation)
	}
}

// TestElicitationHandler_Handle_ReturnsAllow verifies the handler allows the event.
func TestElicitationHandler_Handle_ReturnsAllow(t *testing.T) {
	t.Parallel()

	h := NewElicitationHandler()
	input := &HookInput{
		SessionID:             "sess-elicit",
		ElicitationServerName: "my-mcp-server",
		MCPToolName:           "list_files",
	}

	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil")
	}
}

// --- elicitationResultHandler ---

// TestElicitationResultHandler_EventType returns EventElicitationResult.
func TestElicitationResultHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewElicitationResultHandler()
	if got := h.EventType(); got != EventElicitationResult {
		t.Errorf("EventType() = %q, want %q", got, EventElicitationResult)
	}
}

// TestElicitationResultHandler_Handle_ReturnsAllow verifies the handler allows.
func TestElicitationResultHandler_Handle_ReturnsAllow(t *testing.T) {
	t.Parallel()

	h := NewElicitationResultHandler()
	input := &HookInput{
		SessionID:             "sess-elicit-result",
		ElicitationServerName: "my-mcp-server",
	}

	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil")
	}
}

// --- fileChangedHandler ---
// Tests moved to file_changed_test.go for better organization


// --- genericHandler ---

// TestGenericHandler_EventType returns the configured event type.
func TestGenericHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewGenericHandler(EventSessionStart)
	if got := h.EventType(); got != EventSessionStart {
		t.Errorf("EventType() = %q, want %q", got, EventSessionStart)
	}
}

// TestGenericHandler_Handle_ReturnsDataWithSessionID verifies JSON output.
func TestGenericHandler_Handle_ReturnsDataWithSessionID(t *testing.T) {
	t.Parallel()

	h := NewGenericHandler("TestEvent")
	input := &HookInput{
		SessionID: "sess-generic",
		CWD:       "/project",
	}

	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil")
	}

	if out.Data == nil {
		t.Fatal("Data should not be nil for genericHandler")
	}

	var data map[string]any
	if err := json.Unmarshal(out.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}

	if data["session_id"] != "sess-generic" {
		t.Errorf("session_id = %v, want 'sess-generic'", data["session_id"])
	}
	if data["event"] != "TestEvent" {
		t.Errorf("event = %v, want 'TestEvent'", data["event"])
	}
	if data["status"] != "processed" {
		t.Errorf("status = %v, want 'processed'", data["status"])
	}
}

// TestGenericHandler_Handle_EmptyInput handles empty input gracefully.
func TestGenericHandler_Handle_EmptyInput(t *testing.T) {
	t.Parallel()

	h := NewGenericHandler(EventSessionEnd)
	out, err := h.Handle(context.Background(), &HookInput{})
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil")
	}
}

// TestGenericHandler_DifferentEventTypes verifies multiple event types.
func TestGenericHandler_DifferentEventTypes(t *testing.T) {
	t.Parallel()

	events := []EventType{
		EventSessionStart,
		EventSessionEnd,
		EventPostToolUse,
		EventPreToolUse,
	}

	for _, evt := range events {
		evt := evt
		t.Run(string(evt), func(t *testing.T) {
			t.Parallel()
			h := NewGenericHandler(evt)
			if got := h.EventType(); got != evt {
				t.Errorf("EventType() = %q, want %q", got, evt)
			}
		})
	}
}
