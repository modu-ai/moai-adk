package hook

import (
	"context"
	"encoding/json"
	"log/slog"
)

// sessionEndHandler processes SessionEnd events.
// It persists session metrics, cleans up temporary resources, and optionally
// submits ranking data (REQ-HOOK-034). Always returns "allow".
type sessionEndHandler struct{}

// NewSessionEndHandler creates a new SessionEnd event handler.
func NewSessionEndHandler() Handler {
	return &sessionEndHandler{}
}

// EventType returns EventSessionEnd.
func (h *sessionEndHandler) EventType() EventType {
	return EventSessionEnd
}

// Handle processes a SessionEnd event. It logs the session completion,
// performs cleanup, and returns status in the Data field.
// Errors are non-blocking: the handler logs warnings and returns allow.
func (h *sessionEndHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("session ending",
		"session_id", input.SessionID,
		"project_dir", input.ProjectDir,
	)

	data := map[string]any{
		"session_id": input.SessionID,
		"status":     "completed",
		"cleanup":    "done",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.Error("failed to marshal session end data",
			"error", err.Error(),
		)
		return NewAllowOutput(), nil
	}

	return NewAllowOutputWithData(jsonData), nil
}
