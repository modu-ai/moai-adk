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
	// Default to "ask" - defer decision to user/settings.
	// Per Claude Code protocol, hookSpecificOutput.hookEventName must be "PreToolUse"
	// (not "PermissionRequest") because PermissionRequest shares the PreToolUse output schema.
	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:      "PreToolUse",
			PermissionDecision: DecisionAsk,
		},
	}, nil
}
