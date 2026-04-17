package hook

import (
	"context"
	"encoding/json"
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

// T-015: updatedInput deny re-verification
//
// When a PermissionRequest arrives with a non-empty ToolInput that contains
// the __updated_input_marker__ sentinel key, it means a previous PreToolUse
// hook injected a modified input. The handler must deny to prevent prompt
// injection via modified tool inputs.

// TestPermissionRequestHandler_UpdatedInputDeny verifies that when a
// PermissionRequest ToolInput contains the __updated_input_marker__, the
// handler denies the permission.
func TestPermissionRequestHandler_UpdatedInputDeny(t *testing.T) {
	t.Parallel()

	// Simulate tool input that was modified by a PreToolUse hook
	updatedRaw, _ := json.Marshal(map[string]interface{}{
		"command":               "echo hello",
		"__updated_input_marker__": true,
	})
	input := &HookInput{
		SessionID:     "sess-updated-1",
		ToolName:      "Bash",
		HookEventName: "PermissionRequest",
		ToolInput:     json.RawMessage(updatedRaw),
	}

	h := NewPermissionRequestHandler()
	ctx := context.Background()
	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// When __updated_input_marker__ is present, handler must deny.
	if got == nil {
		t.Fatal("expected non-nil output (deny) when __updated_input_marker__ is set")
	}
	if got.HookSpecificOutput == nil {
		t.Fatal("expected HookSpecificOutput to be set")
	}
	if got.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want 'deny'", got.HookSpecificOutput.PermissionDecision)
	}
}

// TestPermissionRequestHandler_NormalToolInputNoEffect verifies that a
// normal ToolInput without the marker does not trigger deny.
func TestPermissionRequestHandler_NormalToolInputNoEffect(t *testing.T) {
	t.Parallel()

	normalRaw, _ := json.Marshal(map[string]string{"command": "echo hello"})
	input := &HookInput{
		SessionID:     "sess-normal-tool",
		ToolName:      "Bash",
		HookEventName: "PermissionRequest",
		ToolInput:     json.RawMessage(normalRaw),
	}

	h := NewPermissionRequestHandler()
	ctx := context.Background()
	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil output for normal ToolInput, got %+v", got)
	}
}

// TestPermissionRequestHandler_EmptyToolInputNoEffect verifies that nil/empty
// ToolInput does not trigger deny.
func TestPermissionRequestHandler_EmptyToolInputNoEffect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input json.RawMessage
	}{
		{"nil", nil},
		{"empty", json.RawMessage("")},
		{"null", json.RawMessage("null")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := &HookInput{
				SessionID:     "sess-empty-tool",
				ToolName:      "Bash",
				HookEventName: "PermissionRequest",
				ToolInput:     tt.input,
			}

			h := NewPermissionRequestHandler()
			got, err := h.Handle(context.Background(), input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != nil {
				t.Errorf("name=%q: expected nil output, got %+v", tt.name, got)
			}
		})
	}
}
