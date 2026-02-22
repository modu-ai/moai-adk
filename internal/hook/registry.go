package hook

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// @MX:ANCHOR: [AUTO] Hook Registry는 모든 Claude Code 이벤트 핸들러의 중앙 등록 및 디스패치 시스템입니다. 순차 실행, 타임아웃, block short-circuit을 지원합니다.
// @MX:REASON: fan_in=20+, 모든 훅 이벤트의 진입점이며 시스템의 핵심 인프라입니다
// registry is the default implementation of the Registry interface.
// It manages handler registration and sequential event dispatch with
// block short-circuit and timeout support.
type registry struct {
	cfg      ConfigProvider
	handlers map[EventType][]Handler
	timeout  time.Duration
}

// @MX:NOTE: [AUTO] 기본 타임아웃은 30초(DefaultHookTimeout)입니다. 타임아웃 내에 핸들러가 완료되지 않으면 ErrHookTimeout이 반환됩니다.
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

	// Apply timeout from registry configuration
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	for i, h := range handlers {
		slog.Debug("dispatching handler",
			"event", string(event),
			"handler_index", i,
			"handler_total", len(handlers),
		)

		output, err := h.Handle(ctx, input)

		// Check for context deadline exceeded (timeout)
		if ctx.Err() != nil {
			slog.Error("hook execution timed out",
				"event", string(event),
				"handler_index", i,
				"timeout", r.timeout.String(),
			)
			return nil, fmt.Errorf("%w: %v", ErrHookTimeout, ctx.Err())
		}

		// Handler returned an error: stop dispatch chain
		if err != nil {
			slog.Error("handler returned error",
				"event", string(event),
				"handler_index", i,
				"error", err.Error(),
			)
			return nil, fmt.Errorf("handler %d for event %s: %w", i, event, err)
		}

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
	}

	return r.defaultOutputForEvent(event), nil
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
		// Per Claude Code protocol, hookEventName must be "PreToolUse" in output.
		return &HookOutput{
			HookSpecificOutput: &HookSpecificOutput{
				HookEventName:      "PreToolUse",
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
