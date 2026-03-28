package hook

import (
	"context"
	"encoding/json"
	"log/slog"
)

// genericHandler is a passthrough handler for events that only need logging.
type genericHandler struct {
	eventType EventType
}

// NewGenericHandler creates a handler that logs the event and returns empty output.
func NewGenericHandler(eventType EventType) Handler {
	return &genericHandler{eventType: eventType}
}

// EventType returns the configured event type.
func (h *genericHandler) EventType() EventType {
	return h.eventType
}

// Handle logs the event and returns a passthrough response.
func (h *genericHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("hook event received",
		"event", string(h.eventType),
		"session_id", input.SessionID,
		"cwd", input.CWD,
	)

	data := map[string]any{
		"event":      string(h.eventType),
		"session_id": input.SessionID,
		"status":     "processed",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.Error("failed to marshal event data", "error", err.Error())
		return &HookOutput{}, nil
	}

	return &HookOutput{Data: jsonData}, nil
}
