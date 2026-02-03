package hook

import (
	"context"
	"encoding/json"
	"testing"
)

func TestStopHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewStopHandler()

	if got := h.EventType(); got != EventStop {
		t.Errorf("EventType() = %q, want %q", got, EventStop)
	}
}

func TestStopHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        *HookInput
		wantDecision string
	}{
		{
			name: "normal stop",
			input: &HookInput{
				SessionID:     "sess-stop-1",
				CWD:           t.TempDir(),
				HookEventName: "Stop",
				ProjectDir:    t.TempDir(),
			},
			wantDecision: DecisionAllow,
		},
		{
			name: "stop without project dir",
			input: &HookInput{
				SessionID:     "sess-stop-2",
				CWD:           "/tmp",
				HookEventName: "Stop",
			},
			wantDecision: DecisionAllow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewStopHandler()
			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			if got.Decision != tt.wantDecision {
				t.Errorf("Decision = %q, want %q", got.Decision, tt.wantDecision)
			}
			if got.Data != nil && !json.Valid(got.Data) {
				t.Errorf("Data is not valid JSON: %s", got.Data)
			}
		})
	}
}
