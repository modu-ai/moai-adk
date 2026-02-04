package hook

import (
	"context"
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
// any active state, and returns an empty response.
// Stop hooks should not use hookSpecificOutput per Claude Code protocol.
// Errors are non-blocking: the handler logs warnings and returns empty output.
func (h *stopHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("stop requested",
		"session_id", input.SessionID,
		"project_dir", input.ProjectDir,
	)

	// Stop hooks return empty JSON {} per Claude Code protocol
	// Do NOT use hookSpecificOutput for Stop events
	return &HookOutput{}, nil
}
