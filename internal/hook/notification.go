// Resolution: RETIRE-OBS-ONLY — removed from settings.json registration.
// Retained as observability tap via system.yaml hook.observability_events opt-in.
// SPEC-V3R2-RT-006 REQ-040, REQ-041: Pattern A silent return when opt-in not set.
// SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001: HOI master toggle (hook.opt_in.enabled) is
// checked FIRST as defense-in-depth.
// SPEC-V3R6-HOOK-ASYNC-EXPAND-001 M4 (REQ-HAE-004): dual-gate conditional async —
// when BOTH hook.opt_in.enabled (Gate 1, HOI master) AND observability.enabled
// (Gate 2, REQ-OBS-005 master) are true, the payload-logging side-effect
// executes in a background goroutine bounded by asyncDeadline (5s). When either
// gate is false, the handler returns the empty payload immediately with NO
// goroutine spawned (zero-overhead path).
package hook

import (
	"context"
	"log/slog"
	"sync"
)

// notificationHandler processes Notification events.
// It logs notifications sent by Claude Code when observability is opted in.
type notificationHandler struct {
	cfg ConfigProvider
	// wg tracks in-flight async side-effect goroutines.
	wg sync.WaitGroup
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

// waitGroup returns the handler's internal *sync.WaitGroup for testutil.WaitForAsync.
func (h *notificationHandler) waitGroup() *sync.WaitGroup {
	return &h.wg
}

// EventType returns EventNotification.
func (h *notificationHandler) EventType() EventType {
	return EventNotification
}

// Handle processes a Notification event with dual-gate conditional async per
// REQ-HAE-004:
//
//   - Gate 1: hookOptInEnabled (HOI master from system.yaml hook.opt_in.enabled)
//   - Gate 2: IsObservabilityEnabled (REQ-OBS-005 master from observability.yaml)
//
// When BOTH gates pass, the payload-logging side-effect runs in a goroutine
// bounded by asyncDeadline. When EITHER gate is false, returns &HookOutput{}
// immediately with no goroutine spawned (zero-overhead path).
//
// Additionally, when observability_events whitelist (RT-006) does NOT include
// "notification", the handler also short-circuits per Pattern A — preserving
// the existing SPEC-V3R2-RT-006 REQ-040 semantics independently.
func (h *notificationHandler) Handle(_ context.Context, input *HookInput) (*HookOutput, error) {
	// Gate 1 (HOI master toggle from system.yaml hook.opt_in.enabled).
	if !hookOptInEnabled(h.cfg) {
		return &HookOutput{}, nil
	}
	// Per-event RT-006 whitelist (independent gate, preserved from prior behavior).
	if !observabilityOptIn(h.cfg, "notification") {
		return &HookOutput{}, nil
	}
	// Gate 2 (REQ-OBS-005 master from observability.yaml).
	if !IsObservabilityEnabled() {
		return &HookOutput{}, nil
	}

	// All gates pass → async logging side-effect.
	asyncCtx, cancel := context.WithTimeout(context.Background(), asyncDeadline)
	h.wg.Add(1)
	go func() {
		defer cancel()
		defer h.wg.Done()
		h.logNotification(asyncCtx, input)
	}()

	return &HookOutput{}, nil
}

// logNotification emits the slog.Info entry asynchronously. The slog handler
// is responsible for any payload processing; this handler only invokes slog
// and honors the deadline ctx defensively.
func (h *notificationHandler) logNotification(ctx context.Context, input *HookInput) {
	select {
	case <-ctx.Done():
		slog.Warn("notification async: cancelled before log emission",
			"session_id", input.SessionID,
			"error", ctx.Err(),
		)
		return
	default:
	}
	slog.Info("notification received",
		"session_id", input.SessionID,
		"title", input.Title,
		"message", input.Message,
		"type", input.NotificationType,
	)
}
