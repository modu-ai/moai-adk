package hook

import (
	"context"
	"log/slog"
)

// permissionDeniedHandler processes PermissionDenied events.
// Fired after the auto mode classifier denies a tool call.
// Available since Claude Code v2.1.89+.
type permissionDeniedHandler struct{}

// NewPermissionDeniedHandler creates a new PermissionDenied event handler.
func NewPermissionDeniedHandler() Handler {
	return &permissionDeniedHandler{}
}

// EventType returns EventPermissionDenied.
func (h *permissionDeniedHandler) EventType() EventType {
	return EventPermissionDenied
}

// Handle processes a PermissionDenied event. It logs the denied tool name for auditing.
func (h *permissionDeniedHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("permission denied for tool",
		"session_id", input.SessionID,
		"tool_name", input.ToolName,
	)
	return &HookOutput{}, nil
}
