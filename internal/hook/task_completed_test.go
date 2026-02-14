package hook

import (
	"context"
	"testing"
)

func TestTaskCompletedHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewTaskCompletedHandler()

	if got := h.EventType(); got != EventTaskCompleted {
		t.Errorf("EventType() = %q, want %q", got, EventTaskCompleted)
	}
}

func TestTaskCompletedHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input *HookInput
	}{
		{
			name: "task completed with session",
			input: &HookInput{
				SessionID:     "sess-tc-1",
				HookEventName: "TaskCompleted",
			},
		},
		{
			name: "minimal task completed",
			input: &HookInput{
				SessionID: "sess-tc-2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewTaskCompletedHandler()
			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			if got.HookSpecificOutput != nil {
				t.Error("TaskCompleted hook should not set hookSpecificOutput")
			}
			if got.ExitCode != 0 {
				t.Errorf("ExitCode = %d, want 0", got.ExitCode)
			}
		})
	}
}
