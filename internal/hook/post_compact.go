package hook

import (
	"context"
	"encoding/json"
	"log/slog"
)

// postCompactHandler processes PostCompact events.
// It logs compaction completion and captures post-compaction state for session recovery.
type postCompactHandler struct{}

// NewPostCompactHandler creates a new PostCompact event handler.
func NewPostCompactHandler() Handler {
	return &postCompactHandler{}
}

// EventType returns EventPostCompact.
func (h *postCompactHandler) EventType() EventType {
	return EventPostCompact
}

// Handle processes a PostCompact event. It logs the compaction result
// and returns recovery state in the Data field. Errors are non-blocking.
func (h *postCompactHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("post-compact recovery",
		"session_id", input.SessionID,
	)

	data := map[string]any{
		"session_id": input.SessionID,
		"status":     "recovered",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.Warn("post-compact: failed to marshal data", "error", err)
		return &HookOutput{}, nil
	}

	return &HookOutput{Data: jsonData}, nil
}
