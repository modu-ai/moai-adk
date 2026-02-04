package hook

import (
	"context"
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
		name  string
		input *HookInput
	}{
		{
			name: "normal stop",
			input: &HookInput{
				SessionID:     "sess-stop-1",
				CWD:           t.TempDir(),
				HookEventName: "Stop",
				ProjectDir:    t.TempDir(),
			},
		},
		{
			name: "stop without project dir",
			input: &HookInput{
				SessionID:     "sess-stop-2",
				CWD:           "/tmp",
				HookEventName: "Stop",
			},
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
			// Stop hooks return empty JSON {} per Claude Code protocol
			// They should NOT have hookSpecificOutput set
			if got.HookSpecificOutput != nil {
				t.Error("Stop hook should not set hookSpecificOutput")
			}
		})
	}
}
