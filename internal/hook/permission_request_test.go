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
			// Handler returns nil to defer to user's permission mode.
			// Returning non-nil with "ask" would override bypassPermissions.
			if got != nil {
				t.Errorf("expected nil output (defer to system), got %+v", got)
			}
		})
	}
}
