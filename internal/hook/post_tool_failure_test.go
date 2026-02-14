package hook

import (
	"context"
	"testing"
)

func TestPostToolUseFailureHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewPostToolUseFailureHandler()

	if got := h.EventType(); got != EventPostToolUseFailure {
		t.Errorf("EventType() = %q, want %q", got, EventPostToolUseFailure)
	}
}

func TestPostToolUseFailureHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input *HookInput
	}{
		{
			name: "tool failure with error",
			input: &HookInput{
				SessionID:     "sess-fail-1",
				ToolName:      "Bash",
				ToolUseID:     "tu-1",
				Error:         "command failed",
				IsInterrupt:   false,
				HookEventName: "PostToolUseFailure",
			},
		},
		{
			name: "tool failure with interrupt",
			input: &HookInput{
				SessionID:     "sess-fail-2",
				ToolName:      "Read",
				ToolUseID:     "tu-2",
				Error:         "interrupted",
				IsInterrupt:   true,
				HookEventName: "PostToolUseFailure",
			},
		},
		{
			name: "tool failure without optional fields",
			input: &HookInput{
				SessionID: "sess-fail-3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewPostToolUseFailureHandler()
			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			if got.HookSpecificOutput != nil {
				t.Error("PostToolUseFailure hook should not set hookSpecificOutput")
			}
		})
	}
}
