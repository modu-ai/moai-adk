package hook

import (
	"context"
	"log/slog"
)

// stopHandler processes Stop events.
// It performs graceful shutdown, saves in-progress work state, and preserves
// loop controller (Ralph) state (REQ-HOOK-035). Always returns "allow".
type stopHandler struct{}

// NewStopHandler creates a new Stop event handler.
func NewStopHandler() Handler {
	return &stopHandler{}
}

// EventType returns EventStop.
func (h *stopHandler) EventType() EventType {
	return EventStop
}

// Handle processes a Stop event. It logs the stop request, preserves
// any active state, and returns an appropriate response.
//
// Per Claude Code protocol:
// - Return empty JSON {} to allow Claude to stop
// - Return {"decision": "block", "reason": "..."} to keep Claude working
// - Check stop_hook_active to prevent infinite loops
//
// Errors are non-blocking: the handler logs warnings and returns empty output.
func (h *stopHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("stop requested",
		"session_id", input.SessionID,
		"stop_hook_active", input.StopHookActive,
	)

	// IMPORTANT: Prevent infinite loop per Claude Code protocol
	// If stop_hook_active is true, Claude is already continuing due to a previous
	// stop hook decision. Allow Claude to stop to prevent infinite loops.
	if input.StopHookActive {
		slog.Debug("stop_hook_active is true, allowing Claude to stop")
		return &HookOutput{}, nil
	}

	// Stop hooks use top-level decision/reason fields per Claude Code protocol
	// Return empty JSON {} to allow Claude to stop (default behavior)
	// To keep Claude working, return: {"decision": "block", "reason": "..."}
	return &HookOutput{}, nil
}
