package hook

import (
	"context"
	"testing"
)

func TestInstructionsLoadedHandler_EventType(t *testing.T) {
	h := NewInstructionsLoadedHandler()
	if h.EventType() != EventInstructionsLoaded {
		t.Errorf("EventType() = %q, want %q", h.EventType(), EventInstructionsLoaded)
	}
}

func TestInstructionsLoadedHandler_Handle(t *testing.T) {
	t.Parallel()
	h := NewInstructionsLoadedHandler()
	input := &HookInput{
		SessionID:     "test-session",
		HookEventName: "InstructionsLoaded",
	}
	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
}
