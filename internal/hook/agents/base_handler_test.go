package agents

import (
	"context"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// TestBaseHandler_DirectAccess tests baseHandler methods directly.
// Since baseHandler is an unexported struct in this package, these tests
// exercise the Handle and EventType methods defined on baseHandler itself.
func TestBaseHandler_DirectHandle(t *testing.T) {
	t.Parallel()

	h := &baseHandler{
		action: "test-action",
		event:  hook.EventPreToolUse,
		agent:  "test-agent",
	}

	output, err := h.Handle(context.Background(), &hook.HookInput{})
	if err != nil {
		t.Fatalf("baseHandler.Handle() error: %v", err)
	}
	if output == nil {
		t.Fatal("baseHandler.Handle() returned nil output")
	}
}

func TestBaseHandler_DirectHandle_NilInput(t *testing.T) {
	t.Parallel()

	h := &baseHandler{
		action: "action",
		event:  hook.EventPostToolUse,
		agent:  "agent",
	}

	output, err := h.Handle(context.Background(), nil)
	if err != nil {
		t.Fatalf("baseHandler.Handle(nil) error: %v", err)
	}
	if output == nil {
		t.Fatal("baseHandler.Handle(nil) returned nil output")
	}
}

func TestBaseHandler_DirectEventType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		event hook.EventType
	}{
		{hook.EventPreToolUse},
		{hook.EventPostToolUse},
		{hook.EventSubagentStop},
	}

	for _, tt := range tests {
		t.Run(string(tt.event), func(t *testing.T) {
			t.Parallel()

			h := &baseHandler{
				event: tt.event,
			}
			got := h.EventType()
			if got != tt.event {
				t.Errorf("EventType() = %q, want %q", got, tt.event)
			}
		})
	}
}
