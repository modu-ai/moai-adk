// Resolution: RETIRE-OBS-ONLY — removed from settings.json registration.
// Retained as observability tap via system.yaml hook.observability_events opt-in.
// SPEC-V3R2-RT-006 REQ-040, REQ-041: Pattern A silent return when opt-in not set.
// SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001: HOI master toggle (hook.opt_in.enabled) is
// checked FIRST as defense-in-depth.
// SPEC-V3R6-HOOK-ASYNC-EXPAND-001 M4 (REQ-HAE-003): dual-gate conditional async —
// when BOTH hook.opt_in.enabled (Gate 1, HOI master) AND observability.enabled
// (Gate 2, REQ-OBS-005 master) are true, the JSONL append logging side-effect
// executes in a background goroutine bounded by asyncDeadline (5s). When either
// gate is false, the handler returns the empty payload immediately with NO
// goroutine spawned (zero-overhead path).
package hook

import (
	"context"
	"log/slog"
	"sync"
)

// taskCreatedHandler processes TaskCreated events.
// It logs task creation details for session tracking.
// Available since Claude Code v2.1.84+.
type taskCreatedHandler struct {
	cfg ConfigProvider
	// wg tracks in-flight async side-effect goroutines. Tests use the
	// package-internal waitGroup() accessor + testutil.WaitForAsync to
	// deterministically await completion.
	wg sync.WaitGroup
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

// waitGroup returns the handler's internal *sync.WaitGroup for use with
// testutil.WaitForAsync. Package-internal; not exposed via the Handler interface.
func (h *taskCreatedHandler) waitGroup() *sync.WaitGroup {
	return &h.wg
}

// EventType returns EventTaskCreated.
func (h *taskCreatedHandler) EventType() EventType {
	return EventTaskCreated
}

// Handle processes a TaskCreated event with dual-gate conditional async per
// REQ-HAE-003:
//
//   - Gate 1: hookOptInEnabled (HOI master from system.yaml hook.opt_in.enabled)
//   - Gate 2: IsObservabilityEnabled (REQ-OBS-005 master from observability.yaml)
//
// When BOTH gates pass, the logging side-effect runs in a goroutine bounded
// by asyncDeadline. When EITHER gate is false, returns &HookOutput{} immediately
// with no goroutine spawned (zero-overhead path).
//
// Additionally, when observability_events whitelist (RT-006) does NOT include
// "taskCreated", the handler also short-circuits per Pattern A — preserving
// the existing SPEC-V3R2-RT-006 REQ-040 semantics independently.
func (h *taskCreatedHandler) Handle(_ context.Context, input *HookInput) (*HookOutput, error) {
	// Gate 1 (HOI master toggle from system.yaml hook.opt_in.enabled).
	if !hookOptInEnabled(h.cfg) {
		return &HookOutput{}, nil
	}
	// Per-event RT-006 whitelist (independent gate, preserved from prior behavior).
	if !observabilityOptIn(h.cfg, "taskCreated") {
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
		h.logTaskCreated(asyncCtx, input)
	}()

	return &HookOutput{}, nil
}

// logTaskCreated emits the slog.Info entry asynchronously. The slog handler
// is responsible for any JSONL append; this handler only invokes slog and
// honors the deadline ctx defensively.
func (h *taskCreatedHandler) logTaskCreated(ctx context.Context, input *HookInput) {
	select {
	case <-ctx.Done():
		slog.Warn("task_created async: cancelled before log emission",
			"task_id", input.TaskID,
			"error", ctx.Err(),
		)
		return
	default:
	}
	slog.Info("task created",
		"session_id", input.SessionID,
		"task_id", input.TaskID,
		"task_subject", input.TaskSubject,
	)
}
