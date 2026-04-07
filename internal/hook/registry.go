package hook

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/hook/trace"
)

// @MX:ANCHOR: [AUTO] Hook Registry is the central registration and dispatch system for all Claude Code event handlers. Supports sequential execution, timeout, and block short-circuit.
// @MX:REASON: fan_in=20+, entry point for all hook events, core system infrastructure
// registry is the default implementation of the Registry interface.
// It manages handler registration and sequential event dispatch with
// block short-circuit and timeout support.
type registry struct {
	cfg         ConfigProvider
	handlers    map[EventType][]Handler
	timeout     time.Duration
	traceWriter *trace.TraceWriter
	// logDir is set when observability is enabled; TraceWriter is created lazily on first Dispatch.
	logDir string
	mu     sync.Mutex
}

// @MX:NOTE: [AUTO] Default timeout is 30 seconds (DefaultHookTimeout). If handler does not complete within timeout, ErrHookTimeout is returned.
// NewRegistry creates a new Registry with the default timeout (30 seconds).
func NewRegistry(cfg ConfigProvider) *registry {
	return &registry{
		cfg:      cfg,
		handlers: make(map[EventType][]Handler),
		timeout:  DefaultHookTimeout,
	}
}

// NewRegistryWithTimeout creates a new Registry with a custom timeout duration.
func NewRegistryWithTimeout(cfg ConfigProvider, timeout time.Duration) *registry {
	return &registry{
		cfg:      cfg,
		handlers: make(map[EventType][]Handler),
		timeout:  timeout,
	}
}

// Register adds a handler to the registry for its declared event type.
func (r *registry) Register(handler Handler) {
	event := handler.EventType()
	r.handlers[event] = append(r.handlers[event], handler)
	slog.Debug("handler registered",
		"event", string(event),
		"handler_count", len(r.handlers[event]),
	)
}

// Dispatch sends an event to all registered handlers for the given event type.
// Handlers are executed sequentially within a timeout context. If any handler
// returns Decision "block", remaining handlers are skipped and the block result
// is returned immediately (REQ-HOOK-003). If all handlers succeed, Decision
// "allow" is returned (REQ-HOOK-004).
//
// Note: Stop and SessionEnd events should NOT include hookSpecificOutput per
// Claude Code protocol. These events return empty JSON {} instead.
func (r *registry) Dispatch(ctx context.Context, event EventType, input *HookInput) (*HookOutput, error) {
	handlers := r.handlers[event]
	if len(handlers) == 0 {
		slog.Debug("no handlers registered for event", "event", string(event))
		return r.defaultOutputForEvent(event), nil
	}

	// Lazily initialize TraceWriter on first Dispatch with a known SessionID (REQ-OBS-001).
	if input != nil {
		r.ensureTraceWriter(input.SessionID)
	}

	// Apply timeout from registry configuration
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// Start with the default output for this event type and accumulate
	// non-blocking fields (e.g. SystemMessage) from each handler.
	merged := r.defaultOutputForEvent(event)

	for i, h := range handlers {
		slog.Debug("dispatching handler",
			"event", string(event),
			"handler_index", i,
			"handler_total", len(handlers),
		)

		start := time.Now()
		output, err := h.Handle(ctx, input)
		elapsed := time.Since(start)

		// Check for context deadline exceeded (timeout)
		if ctx.Err() != nil {
			slog.Error("hook execution timed out",
				"event", string(event),
				"handler_index", i,
				"timeout", r.timeout.String(),
			)
			r.writeTrace(input, event, h, elapsed, nil, ctx.Err())
			return nil, fmt.Errorf("%w: %v", ErrHookTimeout, ctx.Err())
		}

		// Handler returned an error: stop dispatch chain
		if err != nil {
			slog.Error("handler returned error",
				"event", string(event),
				"handler_index", i,
				"error", err.Error(),
			)
			r.writeTrace(input, event, h, elapsed, nil, err)
			return nil, fmt.Errorf("handler %d for event %s: %w", i, event, err)
		}

		// Write trace entry for successful handler execution.
		r.writeTrace(input, event, h, elapsed, output, nil)

		// Handler returned block: short-circuit remaining handlers
		// Check both top-level decision (Stop, PostToolUse) and
		// hookSpecificOutput.permissionDecision (PreToolUse)
		if output != nil && isBlockDecision(output) {
			reason := getBlockReason(output)
			slog.Info("handler blocked action",
				"event", string(event),
				"handler_index", i,
				"reason", reason,
			)
			return output, nil
		}

		// Handler signalled exit code 2 (TeammateIdle keep-working, TaskCompleted reject).
		// Short-circuit so the caller (CLI) can exit with code 2.
		if output != nil && output.ExitCode == 2 {
			slog.Info("handler requested exit code 2",
				"event", string(event),
				"handler_index", i,
			)
			return output, nil
		}

		// Accumulate SystemMessage from non-blocking handlers so that
		// messages from earlier handlers (e.g. auto-update notifications,
		// memory injection) are not silently discarded.
		if output != nil && output.SystemMessage != "" {
			if merged.SystemMessage != "" {
				merged.SystemMessage += "\n" + output.SystemMessage
			} else {
				merged.SystemMessage = output.SystemMessage
			}
		}
	}

	return merged, nil
}

// isBlockDecision checks if the output represents a blocking decision.
// Per Claude Code protocol:
// - Stop/PostToolUse use top-level decision = "block"
// - PreToolUse uses hookSpecificOutput.permissionDecision = "deny"
func isBlockDecision(output *HookOutput) bool {
	// Check top-level decision (Stop, PostToolUse)
	if output.Decision == DecisionBlock {
		return true
	}
	// Check hookSpecificOutput.permissionDecision (PreToolUse)
	if output.HookSpecificOutput != nil && output.HookSpecificOutput.PermissionDecision == DecisionDeny {
		return true
	}
	return false
}

// getBlockReason extracts the reason from a blocking output.
func getBlockReason(output *HookOutput) string {
	// Check top-level reason first (Stop, PostToolUse)
	if output.Reason != "" {
		return output.Reason
	}
	// Check hookSpecificOutput.permissionDecisionReason (PreToolUse)
	if output.HookSpecificOutput != nil && output.HookSpecificOutput.PermissionDecisionReason != "" {
		return output.HookSpecificOutput.PermissionDecisionReason
	}
	return ""
}

// defaultOutputForEvent returns the appropriate default output based on event type.
// Stop, SessionEnd, SessionStart, and PreCompact events return empty HookOutput per Claude Code protocol.
// PreToolUse and PostToolUse events return HookOutput with hookSpecificOutput.
func (r *registry) defaultOutputForEvent(event EventType) *HookOutput {
	switch event {
	case EventStop, EventSessionEnd, EventSessionStart, EventPreCompact,
		EventSubagentStop, EventPostToolUseFailure, EventNotification,
		EventSubagentStart, EventUserPromptSubmit, EventTeammateIdle,
		EventTaskCompleted, EventWorktreeCreate, EventWorktreeRemove:
		// These events do NOT use hookSpecificOutput per Claude Code protocol
		// Return empty JSON {}
		return &HookOutput{}
	case EventPermissionRequest:
		// PermissionRequest defaults to "ask" (defer to user).
		// Per Claude Code protocol (v2.1.59+), hookEventName must be "PermissionRequest".
		return &HookOutput{
			HookSpecificOutput: &HookSpecificOutput{
				HookEventName:      "PermissionRequest",
				PermissionDecision: DecisionAsk,
			},
		}
	case EventPreToolUse:
		return NewAllowOutput()
	case EventPostToolUse:
		return NewPostToolOutput("")
	default:
		return &HookOutput{}
	}
}

// Handlers returns all handlers registered for the given event type.
func (r *registry) Handlers(event EventType) []Handler {
	return r.handlers[event]
}

// SetTraceWriter attaches an async TraceWriter to the registry.
// When set, every handler execution will produce a trace entry (REQ-OBS-001).
// Passing nil disables tracing.
func (r *registry) SetTraceWriter(tw *trace.TraceWriter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.traceWriter = tw
}

// EnableObservability configures the registry to create a per-session TraceWriter
// lazily on the first Dispatch call that has a non-empty SessionID (REQ-OBS-001).
// logDir is the directory where JSONL trace files will be written.
// Calling this after SetTraceWriter is a no-op for any session already started.
func (r *registry) EnableObservability(logDir string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logDir = logDir
}

// ensureTraceWriter lazily creates the TraceWriter for the given sessionID if
// observability is enabled and no writer has been set yet. Thread-safe.
func (r *registry) ensureTraceWriter(sessionID string) {
	if sessionID == "" || r.logDir == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.traceWriter == nil {
		r.traceWriter = trace.NewTraceWriter(r.logDir, sessionID)
	}
}

// writeTrace builds and enqueues a TraceEntry from a completed handler execution.
// It is a no-op when traceWriter is nil (observability disabled).
// The write is non-blocking (REQ-OBS-003): entries are sent to the async writer.
func (r *registry) writeTrace(
	input *HookInput,
	event EventType,
	h Handler,
	elapsed time.Duration,
	output *HookOutput,
	handlerErr error,
) {
	r.mu.Lock()
	tw := r.traceWriter
	r.mu.Unlock()
	if tw == nil {
		return
	}

	entry := trace.TraceEntry{
		Timestamp:  time.Now().Add(-elapsed), // approximate invocation time
		Event:      string(event),
		Handler:    fmt.Sprintf("%T", h),
		DurationMs: elapsed.Milliseconds(),
		SessionID:  input.SessionID,
	}

	if input != nil {
		entry.Tool = input.ToolName
	}

	if output != nil {
		// Extract the effective decision string.
		entry.Decision = effectiveDecision(output)
		entry.Reason = effectiveReason(output)
	}

	if handlerErr != nil {
		entry.Error = handlerErr.Error()
	}

	tw.Write(entry)
}

// effectiveDecision extracts the decision string from a HookOutput.
// PreToolUse uses hookSpecificOutput.permissionDecision; others use top-level decision.
func effectiveDecision(output *HookOutput) string {
	if output.HookSpecificOutput != nil && output.HookSpecificOutput.PermissionDecision != "" {
		return output.HookSpecificOutput.PermissionDecision
	}
	return output.Decision
}

// effectiveReason extracts the reason string from a HookOutput (REQ-OBS-008).
func effectiveReason(output *HookOutput) string {
	if output.HookSpecificOutput != nil && output.HookSpecificOutput.PermissionDecisionReason != "" {
		return output.HookSpecificOutput.PermissionDecisionReason
	}
	return output.Reason
}
