package hook

import (
	"context"
	"log/slog"
)

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
func (h *permissionRequestHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("permission requested",
		"session_id", input.SessionID,
		"tool_name", input.ToolName,
	)
	// Return nil output to defer to user's permission mode (bypassPermissions, acceptEdits, etc.).
	// Returning permissionDecision:"ask" would override bypass mode and always show the dialog.
	// By returning nil, Claude Code applies its normal permission system.
	return nil, nil
}
