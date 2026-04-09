package hook

import (
	"context"
	"log/slog"
)

// readOnlyTools is the set of tool names that are safe to retry after a PermissionDenied event.
// These tools only read data and never modify the filesystem or external state.
var readOnlyTools = map[string]bool{
	"Read":                 true,
	"Grep":                 true,
	"Glob":                 true,
	"WebFetch":             true,
	"WebSearch":            true,
	"Skill":                true,
	"ListMcpResourcesTool": true,
	"ReadMcpResourceTool":  true,
}

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

// Handle processes a PermissionDenied event.
// Read-only tools (Read, Grep, Glob, WebFetch, WebSearch, Skill, ListMcpResourcesTool,
// ReadMcpResourceTool) are safe to retry because they do not modify any state.
// Write operations are not retried to prevent unintended side effects.
func (h *permissionDeniedHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	if readOnlyTools[input.ToolName] {
		slog.Info("permission denied for read-only tool, signaling retry",
			"session_id", input.SessionID,
			"tool_name", input.ToolName,
		)
		return NewPermissionDeniedRetryOutput(), nil
	}

	slog.Info("permission denied for tool, no retry",
		"session_id", input.SessionID,
		"tool_name", input.ToolName,
	)
	return &HookOutput{}, nil
}
