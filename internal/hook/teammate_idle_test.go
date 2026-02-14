package hook

import (
	"context"
	"testing"
)

func TestTeammateIdleHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewTeammateIdleHandler()

	if got := h.EventType(); got != EventTeammateIdle {
		t.Errorf("EventType() = %q, want %q", got, EventTeammateIdle)
	}
}

func TestTeammateIdleHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input *HookInput
	}{
		{
			name: "teammate idle with agent ID",
			input: &HookInput{
				SessionID:     "sess-ti-1",
				AgentID:       "teammate-1",
				HookEventName: "TeammateIdle",
			},
		},
		{
			name: "teammate idle without agent ID",
			input: &HookInput{
				SessionID: "sess-ti-2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewTeammateIdleHandler()
			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			if got.HookSpecificOutput != nil {
				t.Error("TeammateIdle hook should not set hookSpecificOutput")
			}
			if got.ExitCode != 0 {
				t.Errorf("ExitCode = %d, want 0", got.ExitCode)
			}
		})
	}
}
