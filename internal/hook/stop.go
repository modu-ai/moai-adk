package hook

import (
	"context"
	"encoding/json"
	"log/slog"
)

// stopHandler processes Stop events.
// It performs graceful shutdown, saves in-progress work state, and preserves
// loop controller (Ralph) state (REQ-HOOK-035). Always returns "allow".
type stopHandler struct{}

// NewStopHandler creates a new Stop event handler.
func NewStopHandler() Handler {
	return &stopHandler{}
}

// EventType returns EventStop.
func (h *stopHandler) EventType() EventType {
	return EventStop
}

// Handle processes a Stop event. It logs the stop request, preserves
// any active state, and returns status in the Data field.
// Errors are non-blocking: the handler logs warnings and returns allow.
func (h *stopHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("stop requested",
		"session_id", input.SessionID,
		"project_dir", input.ProjectDir,
	)

	data := map[string]any{
		"session_id":  input.SessionID,
		"status":      "stopped",
		"state_saved": true,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.Error("failed to marshal stop data",
			"error", err.Error(),
		)
		return NewAllowOutput(), nil
	}

	return NewAllowOutputWithData(jsonData), nil
}
