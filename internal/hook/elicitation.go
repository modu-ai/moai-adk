// Resolution: RETIRE-OBS-ONLY — removed from settings.json registration.
// Retained as observability tap via system.yaml hook.observability_events opt-in.
// SPEC-V3R2-RT-006 REQ-040, REQ-041: Pattern A silent return when opt-in not set.
package hook

import (
	"context"
	"log/slog"
)

// elicitationHandler processes Elicitation events.
// Fired when an MCP server requests user input during a session.
// Available since Claude Code v2.1.76+.
type elicitationHandler struct {
	cfg ConfigProvider
}

// NewElicitationHandler creates a new Elicitation event handler without config.
// Returns silently (no opt-in by default).
func NewElicitationHandler() Handler {
	return &elicitationHandler{}
}

// NewElicitationHandlerWithConfig creates an Elicitation handler that reads
// observability opt-in state from the provided config.
func NewElicitationHandlerWithConfig(cfg ConfigProvider) Handler {
	return &elicitationHandler{cfg: cfg}
}

// EventType returns EventElicitation.
func (h *elicitationHandler) EventType() EventType {
	return EventElicitation
}

// Handle processes an Elicitation event. Returns silently when not opted in.
// When observability_events includes "elicitation", logs MCP server and tool.
func (h *elicitationHandler) Handle(_ context.Context, input *HookInput) (*HookOutput, error) {
	if !observabilityOptIn(h.cfg, "elicitation") {
		// Pattern A: silent return.
		return &HookOutput{}, nil
	}
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
type elicitationResultHandler struct {
	cfg ConfigProvider
}

// NewElicitationResultHandler creates a new ElicitationResult handler without config.
func NewElicitationResultHandler() Handler {
	return &elicitationResultHandler{}
}

// NewElicitationResultHandlerWithConfig creates an ElicitationResult handler
// that reads observability opt-in state from the provided config.
func NewElicitationResultHandlerWithConfig(cfg ConfigProvider) Handler {
	return &elicitationResultHandler{cfg: cfg}
}

// EventType returns EventElicitationResult.
func (h *elicitationResultHandler) EventType() EventType {
	return EventElicitationResult
}

// Handle processes an ElicitationResult event. Returns silently when not opted in.
// When observability_events includes "elicitationResult", logs the completion.
func (h *elicitationResultHandler) Handle(_ context.Context, input *HookInput) (*HookOutput, error) {
	if !observabilityOptIn(h.cfg, "elicitationResult") {
		// Pattern A: silent return.
		return &HookOutput{}, nil
	}
	slog.Info("mcp elicitation completed",
		"session_id", input.SessionID,
		"server_name", input.ElicitationServerName,
	)
	return &HookOutput{}, nil
}
