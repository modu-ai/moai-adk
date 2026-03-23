package hook

import (
	"context"
	"log/slog"
)

// stopFailureHandler processes StopFailure events.
// It logs API errors (rate limit, auth failure) that caused the turn to end abnormally.
type stopFailureHandler struct{}

// NewStopFailureHandler creates a new StopFailure event handler.
func NewStopFailureHandler() Handler {
	return &stopFailureHandler{}
}

// EventType returns EventStopFailure.
func (h *stopFailureHandler) EventType() EventType {
	return EventStopFailure
}

// Handle processes a StopFailure event. It logs the error details for diagnostics.
// Errors are non-blocking — the handler always returns empty output.
func (h *stopFailureHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Warn("stop failure: turn ended due to API error",
		"session_id", input.SessionID,
		"error", input.Error,
	)

	return &HookOutput{}, nil
}
