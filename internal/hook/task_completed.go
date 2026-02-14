package hook

import (
	"context"
	"log/slog"
)

// taskCompletedHandler processes TaskCompleted events.
// It logs task completion events in Agent Teams mode.
type taskCompletedHandler struct{}

// NewTaskCompletedHandler creates a new TaskCompleted event handler.
func NewTaskCompletedHandler() Handler {
	return &taskCompletedHandler{}
}

// EventType returns EventTaskCompleted.
func (h *taskCompletedHandler) EventType() EventType {
	return EventTaskCompleted
}

// Handle processes a TaskCompleted event. It logs the task completion.
// Returns empty output to accept completion (exit code 0).
// To reject completion, return NewTaskRejectedOutput() (exit code 2).
func (h *taskCompletedHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("task completed",
		"session_id", input.SessionID,
	)
	return &HookOutput{}, nil
}
