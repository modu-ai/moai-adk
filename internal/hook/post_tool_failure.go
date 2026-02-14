package hook

import (
	"context"
	"log/slog"
)

// postToolUseFailureHandler processes PostToolUseFailure events.
// It logs tool execution failures for diagnostics and tracing.
type postToolUseFailureHandler struct{}

// NewPostToolUseFailureHandler creates a new PostToolUseFailure event handler.
func NewPostToolUseFailureHandler() Handler {
	return &postToolUseFailureHandler{}
}

// EventType returns EventPostToolUseFailure.
func (h *postToolUseFailureHandler) EventType() EventType {
	return EventPostToolUseFailure
}

// Handle processes a PostToolUseFailure event. It logs the tool failure
// with error details and returns empty output (allow default processing).
func (h *postToolUseFailureHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("tool execution failed",
		"session_id", input.SessionID,
		"tool_name", input.ToolName,
		"tool_use_id", input.ToolUseID,
		"error", input.Error,
		"is_interrupt", input.IsInterrupt,
	)
	return &HookOutput{}, nil
}
