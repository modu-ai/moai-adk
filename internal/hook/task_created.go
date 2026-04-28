// Resolution: RETIRE-OBS-ONLY — removed from settings.json registration.
// Retained as observability tap via system.yaml hook.observability_events opt-in.
package hook

import (
	"context"
	"log/slog"
)

// taskCreatedHandler processes TaskCreated events.
// It logs task creation details for session tracking.
// Available since Claude Code v2.1.84+.
type taskCreatedHandler struct{}

// NewTaskCreatedHandler creates a new TaskCreated event handler.
func NewTaskCreatedHandler() Handler {
	return &taskCreatedHandler{}
}

// EventType returns EventTaskCreated.
func (h *taskCreatedHandler) EventType() EventType {
	return EventTaskCreated
}

// Handle processes a TaskCreated event. It logs task creation details.
func (h *taskCreatedHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("task created",
		"session_id", input.SessionID,
		"task_id", input.TaskID,
		"task_subject", input.TaskSubject,
	)
	return &HookOutput{}, nil
}
