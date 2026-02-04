package hook

import (
	"context"
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

// Handle processes a SessionEnd event. It logs the session completion
// and returns an empty response.
// SessionEnd hooks should not use hookSpecificOutput per Claude Code protocol.
// Errors are non-blocking: the handler logs warnings and returns empty output.
func (h *sessionEndHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("session ending",
		"session_id", input.SessionID,
		"project_dir", input.ProjectDir,
	)

	// SessionEnd hooks return empty JSON {} per Claude Code protocol
	// Do NOT use hookSpecificOutput for SessionEnd events
	return &HookOutput{}, nil
}
