// Resolution: RETIRE-OBS-ONLY — removed from settings.json registration.
// Retained as observability tap via system.yaml hook.observability_events opt-in.
// SPEC-V3R2-RT-006 REQ-040, REQ-041: Pattern A silent return when opt-in not set.
package hook

import (
	"context"
	"log/slog"
)

// taskCreatedHandler processes TaskCreated events.
// It logs task creation details for session tracking.
// Available since Claude Code v2.1.84+.
type taskCreatedHandler struct {
	cfg ConfigProvider
}

// NewTaskCreatedHandler creates a new TaskCreated event handler without config.
// Returns silently (no opt-in by default).
func NewTaskCreatedHandler() Handler {
	return &taskCreatedHandler{}
}

// NewTaskCreatedHandlerWithConfig creates a TaskCreated handler that reads
// observability opt-in state from the provided config.
func NewTaskCreatedHandlerWithConfig(cfg ConfigProvider) Handler {
	return &taskCreatedHandler{cfg: cfg}
}

// EventType returns EventTaskCreated.
func (h *taskCreatedHandler) EventType() EventType {
	return EventTaskCreated
}

// Handle processes a TaskCreated event. Returns silently when not opted in.
// When observability_events includes "taskCreated", logs task creation details.
//
// SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001: HOI master toggle (hook.opt_in.enabled)
// is checked FIRST as defense-in-depth. When disabled (default), the handler
// short-circuits regardless of the RT-006 per-event whitelist.
func (h *taskCreatedHandler) Handle(_ context.Context, input *HookInput) (*HookOutput, error) {
	if !hookOptInEnabled(h.cfg) {
		// HOI master toggle off — Pattern A silent return.
		return &HookOutput{}, nil
	}
	if !observabilityOptIn(h.cfg, "taskCreated") {
		// Pattern A: silent return.
		return &HookOutput{}, nil
	}
	slog.Info("task created",
		"session_id", input.SessionID,
		"task_id", input.TaskID,
		"task_subject", input.TaskSubject,
	)
	return &HookOutput{}, nil
}
