package hook

import (
	"context"
	"log/slog"
)

// cwdChangedHandler processes CwdChanged events.
// Fired when the working directory changes during a session.
// Available since Claude Code v2.1.83+.
type cwdChangedHandler struct{}

// NewCwdChangedHandler creates a new CwdChanged event handler.
func NewCwdChangedHandler() Handler {
	return &cwdChangedHandler{}
}

// EventType returns EventCwdChanged.
func (h *cwdChangedHandler) EventType() EventType {
	return EventCwdChanged
}

// Handle processes a CwdChanged event. It logs the new working directory.
func (h *cwdChangedHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("working directory changed",
		"session_id", input.SessionID,
		"cwd", input.CWD,
	)
	return &HookOutput{}, nil
}
