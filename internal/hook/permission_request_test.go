package hook

import (
	"context"
	"testing"
)

func TestPermissionRequestHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewPermissionRequestHandler()

	if got := h.EventType(); got != EventPermissionRequest {
		t.Errorf("EventType() = %q, want %q", got, EventPermissionRequest)
	}
}

func TestPermissionRequestHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input *HookInput
	}{
		{
			name: "permission for Bash tool",
			input: &HookInput{
				SessionID:     "sess-perm-1",
				ToolName:      "Bash",
				HookEventName: "PermissionRequest",
			},
		},
		{
			name: "permission for Write tool",
			input: &HookInput{
				SessionID:     "sess-perm-2",
				ToolName:      "Write",
				HookEventName: "PermissionRequest",
			},
		},
		{
			name: "permission without tool name",
			input: &HookInput{
				SessionID: "sess-perm-3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewPermissionRequestHandler()
			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			if got.HookSpecificOutput == nil {
				t.Fatal("PermissionRequest hook should set hookSpecificOutput")
			}
			if got.HookSpecificOutput.PermissionDecision != DecisionAsk {
				t.Errorf("PermissionDecision = %q, want %q", got.HookSpecificOutput.PermissionDecision, DecisionAsk)
			}
			if got.HookSpecificOutput.HookEventName != "PreToolUse" {
				t.Errorf("HookEventName = %q, want %q", got.HookSpecificOutput.HookEventName, "PreToolUse")
			}
		})
	}
}
