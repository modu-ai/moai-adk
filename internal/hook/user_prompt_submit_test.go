package hook

import (
	"context"
	"strings"
	"testing"
)

func TestUserPromptSubmitHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewUserPromptSubmitHandler()

	if got := h.EventType(); got != EventUserPromptSubmit {
		t.Errorf("EventType() = %q, want %q", got, EventUserPromptSubmit)
	}
}

func TestUserPromptSubmitHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input *HookInput
	}{
		{
			name: "short prompt",
			input: &HookInput{
				SessionID:     "sess-ups-1",
				Prompt:        "hello world",
				HookEventName: "UserPromptSubmit",
			},
		},
		{
			name: "long prompt truncated",
			input: &HookInput{
				SessionID:     "sess-ups-2",
				Prompt:        strings.Repeat("a", 200),
				HookEventName: "UserPromptSubmit",
			},
		},
		{
			name: "empty prompt",
			input: &HookInput{
				SessionID:     "sess-ups-3",
				Prompt:        "",
				HookEventName: "UserPromptSubmit",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewUserPromptSubmitHandler()
			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			if got.HookSpecificOutput != nil {
				t.Error("UserPromptSubmit hook should not set hookSpecificOutput")
			}
		})
	}
}
