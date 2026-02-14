package hook

import (
	"context"
	"log/slog"
)

// teammateIdleHandler processes TeammateIdle events.
// It logs teammate idle events in Agent Teams mode.
type teammateIdleHandler struct{}

// NewTeammateIdleHandler creates a new TeammateIdle event handler.
func NewTeammateIdleHandler() Handler {
	return &teammateIdleHandler{}
}

// EventType returns EventTeammateIdle.
func (h *teammateIdleHandler) EventType() EventType {
	return EventTeammateIdle
}

// Handle processes a TeammateIdle event. It logs the teammate idle state.
// Returns empty output to accept idle (exit code 0).
// To keep working, return NewTeammateKeepWorkingOutput() (exit code 2).
func (h *teammateIdleHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("teammate idle",
		"session_id", input.SessionID,
		"agent_id", input.AgentID,
	)
	return &HookOutput{}, nil
}
