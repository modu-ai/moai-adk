package hook

import (
	"context"
	"testing"
)

func TestStopFailureHandler_EventType(t *testing.T) {
	h := NewStopFailureHandler()
	if h.EventType() != EventStopFailure {
		t.Errorf("EventType() = %q, want %q", h.EventType(), EventStopFailure)
	}
}

func TestStopFailureHandler_Handle(t *testing.T) {
	t.Parallel()
	h := NewStopFailureHandler()
	input := &HookInput{
		SessionID:     "test-session",
		HookEventName: "StopFailure",
		Error:         "rate_limit_exceeded",
	}
	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
}
