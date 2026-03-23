package hook

import (
	"context"
	"testing"
)

func TestPostCompactHandler_EventType(t *testing.T) {
	h := NewPostCompactHandler()
	if h.EventType() != EventPostCompact {
		t.Errorf("EventType() = %q, want %q", h.EventType(), EventPostCompact)
	}
}

func TestPostCompactHandler_Handle(t *testing.T) {
	t.Parallel()
	h := NewPostCompactHandler()
	input := &HookInput{
		SessionID:     "test-session",
		HookEventName: "PostCompact",
	}
	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if len(out.Data) == 0 {
		t.Error("expected non-empty Data field")
	}
}
