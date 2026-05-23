// Resolution: RETIRE-OBS-ONLY — removed from settings.json registration.
// Retained as observability tap via system.yaml hook.observability_events opt-in.
// SPEC-V3R2-RT-006 REQ-040, REQ-041: Pattern A silent return when opt-in not set.
package hook

import (
	"context"
	"log/slog"
)

// notificationHandler processes Notification events.
// It logs notifications sent by Claude Code when observability is opted in.
type notificationHandler struct {
	cfg ConfigProvider
}

// NewNotificationHandler creates a new Notification event handler without
// a config provider. The handler returns HookOutput{} silently (no opt-in).
func NewNotificationHandler() Handler {
	return &notificationHandler{}
}

// NewNotificationHandlerWithConfig creates a new Notification event handler
// that reads observability opt-in state from the provided config.
func NewNotificationHandlerWithConfig(cfg ConfigProvider) Handler {
	return &notificationHandler{cfg: cfg}
}

// EventType returns EventNotification.
func (h *notificationHandler) EventType() EventType {
	return EventNotification
}

// Handle processes a Notification event. Returns silently when not opted in.
// When observability_events includes "notification", logs the event details.
//
// SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001: HOI master toggle (hook.opt_in.enabled)
// is checked FIRST as defense-in-depth. When disabled (default), the handler
// short-circuits regardless of the RT-006 per-event whitelist.
func (h *notificationHandler) Handle(_ context.Context, input *HookInput) (*HookOutput, error) {
	if !hookOptInEnabled(h.cfg) {
		// HOI master toggle off — Pattern A silent return.
		return &HookOutput{}, nil
	}
	if !observabilityOptIn(h.cfg, "notification") {
		// Pattern A: silent return, no logging, no user-facing message.
		return &HookOutput{}, nil
	}
	slog.Info("notification received",
		"session_id", input.SessionID,
		"title", input.Title,
		"message", input.Message,
		"type", input.NotificationType,
	)
	return &HookOutput{}, nil
}
