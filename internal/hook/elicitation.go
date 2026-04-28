// Resolution: RETIRE-OBS-ONLY — removed from settings.json registration.
// Retained as observability tap via system.yaml hook.observability_events opt-in.
package hook

import (
	"context"
	"log/slog"
)

// elicitationHandler processes Elicitation events.
// Fired when an MCP server requests user input during a session.
// Available since Claude Code v2.1.76+.
type elicitationHandler struct{}

// NewElicitationHandler creates a new Elicitation event handler.
func NewElicitationHandler() Handler {
	return &elicitationHandler{}
}

// EventType returns EventElicitation.
func (h *elicitationHandler) EventType() EventType {
	return EventElicitation
}

// Handle processes an Elicitation event. It logs the MCP server and tool
// that triggered the elicitation request.
func (h *elicitationHandler) Handle(_ context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("mcp elicitation requested",
		"session_id", input.SessionID,
		"server_name", input.ElicitationServerName,
		"mcp_tool", input.MCPToolName,
	)
	return &HookOutput{}, nil
}

// elicitationResultHandler processes ElicitationResult events.
// Fired after user responds to an MCP elicitation request.
// Available since Claude Code v2.1.76+.
type elicitationResultHandler struct{}

// NewElicitationResultHandler creates a new ElicitationResult event handler.
func NewElicitationResultHandler() Handler {
	return &elicitationResultHandler{}
}

// EventType returns EventElicitationResult.
func (h *elicitationResultHandler) EventType() EventType {
	return EventElicitationResult
}

// Handle processes an ElicitationResult event. It logs the completion
// of the MCP elicitation interaction.
func (h *elicitationResultHandler) Handle(_ context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("mcp elicitation completed",
		"session_id", input.SessionID,
		"server_name", input.ElicitationServerName,
	)
	return &HookOutput{}, nil
}
