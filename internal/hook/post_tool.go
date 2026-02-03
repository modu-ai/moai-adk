package hook

import (
	"context"
	"encoding/json"
	"log/slog"
)

// postToolHandler processes PostToolUse events.
// It collects tool execution metrics and prepares statusline data
// (REQ-HOOK-033). This handler is observation-only and always returns "allow".
type postToolHandler struct{}

// NewPostToolHandler creates a new PostToolUse event handler.
func NewPostToolHandler() Handler {
	return &postToolHandler{}
}

// EventType returns EventPostToolUse.
func (h *postToolHandler) EventType() EventType {
	return EventPostToolUse
}

// Handle processes a PostToolUse event. It collects metrics about the tool
// execution (tool name, output size) and returns them in the Data field.
// Always returns Decision "allow" (observation only).
func (h *postToolHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Debug("collecting post-tool metrics",
		"tool_name", input.ToolName,
		"session_id", input.SessionID,
	)

	metrics := map[string]any{
		"tool_name":  input.ToolName,
		"session_id": input.SessionID,
	}

	// Collect output size metric
	if len(input.ToolOutput) > 0 {
		metrics["output_size"] = len(input.ToolOutput)
	}

	// Collect input size metric
	if len(input.ToolInput) > 0 {
		metrics["input_size"] = len(input.ToolInput)
	}

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		slog.Error("failed to marshal post-tool metrics",
			"error", err.Error(),
		)
		return NewAllowOutput(), nil
	}

	return NewAllowOutputWithData(jsonData), nil
}
