package hook

import (
	"bytes"
	"context"
	"log/slog"
)

// updatedInputMarker is the sentinel key injected into tool inputs that were
// modified by a PreToolUse hook. Its presence signals a re-verification need.
const updatedInputMarker = `"__updated_input_marker__"`

// permissionRequestHandler processes PermissionRequest events.
// It logs permission requests and defers to the default decision.
type permissionRequestHandler struct{}

// NewPermissionRequestHandler creates a new PermissionRequest event handler.
func NewPermissionRequestHandler() Handler {
	return &permissionRequestHandler{}
}

// EventType returns EventPermissionRequest.
func (h *permissionRequestHandler) EventType() EventType {
	return EventPermissionRequest
}

// Handle processes a PermissionRequest event. It logs the permission request
// and returns "ask" (defer to user/default settings).
//
// When the ToolInput contains the __updated_input_marker__ sentinel, the
// handler denies the permission to prevent prompt injection via modified
// tool inputs (updatedInput re-verification, SPEC-OPUS47-COMPAT-001 T-015).
func (h *permissionRequestHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("permission requested",
		"session_id", input.SessionID,
		"tool_name", input.ToolName,
	)

	// Re-verify: deny if the tool input was modified by a previous hook.
	if len(input.ToolInput) > 0 && bytes.Contains(input.ToolInput, []byte(updatedInputMarker)) {
		slog.Warn("denying permission: tool input contains updated_input_marker",
			"session_id", input.SessionID,
			"tool_name", input.ToolName,
		)
		return NewPermissionRequestOutput("deny", "tool input was modified by a previous hook (updatedInput re-verification)"), nil
	}

	// Return nil output to defer to user's permission mode (bypassPermissions, acceptEdits, etc.).
	// Returning permissionDecision:"ask" would override bypass mode and always show the dialog.
	// By returning nil, Claude Code applies its normal permission system.
	return nil, nil
}
