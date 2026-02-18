package hook

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// autoUpdateTimeout is the maximum time the auto-update handler will wait
// for the update function to complete. It must be shorter than
// DefaultHookTimeout (30s) so that the registry's ctx.Err() check after
// Handle() returns finds the context still valid. It must also be shorter
// than the SessionStart hook timeout configured in settings.json (30s) so
// that the hook process exits before Claude Code forcibly terminates it.
const autoUpdateTimeout = 25 * time.Second

// AutoUpdateResult holds the outcome of an automatic binary update attempt.
type AutoUpdateResult struct {
	// Updated is true if a new binary was installed.
	Updated bool
	// PreviousVersion is the version before the update.
	PreviousVersion string
	// NewVersion is the version after the update.
	NewVersion string
	// Error holds any non-fatal error encountered during the update.
	Error error
}

// AutoUpdateFunc is a callback that performs the binary self-update.
// It is provided by the CLI layer to avoid circular dependencies
// between the hook package and the update package.
type AutoUpdateFunc func(ctx context.Context) (*AutoUpdateResult, error)

// autoUpdateHandler processes SessionStart events to automatically
// check for and install binary updates. Errors are logged but never
// propagated to prevent blocking the session.
type autoUpdateHandler struct {
	updateFn AutoUpdateFunc
}

// NewAutoUpdateHandler creates a SessionStart handler that runs the
// given update function on every session start. The handler is non-blocking:
// errors are logged and a SystemMessage notification is sent on success.
func NewAutoUpdateHandler(fn AutoUpdateFunc) Handler {
	return &autoUpdateHandler{updateFn: fn}
}

// EventType returns EventSessionStart.
func (h *autoUpdateHandler) EventType() EventType {
	return EventSessionStart
}

// Handle executes the auto-update callback and returns a SystemMessage
// if a new version was installed. All errors are logged and swallowed.
//
// The update runs with an independent context bounded by autoUpdateTimeout
// (25s). This ensures the handler always returns before:
//   - The registry's DefaultHookTimeout (30s) expires and ctx.Err() becomes
//     non-nil, which would cause the registry to return ErrHookTimeout.
//   - The Claude Code SessionStart hook timeout (30s in settings.json) elapses
//     and the moai process is forcibly terminated, producing a hook error.
//
// This prevents spurious "SessionStart:startup hook error" messages that
// occur on Windows when the /model command triggers a new session start
// (e.g., switching to Opus 4.6) and the update check takes longer than
// expected due to network latency or binary-in-use conditions.
func (h *autoUpdateHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	if h.updateFn == nil {
		return &HookOutput{}, nil
	}

	// Use an independent context with a timeout shorter than both the registry
	// DefaultHookTimeout and the Claude Code hook process timeout. This
	// guarantees the handler returns in time regardless of network conditions.
	updateCtx, cancel := context.WithTimeout(context.Background(), autoUpdateTimeout)
	defer cancel()

	result, err := h.updateFn(updateCtx)
	if err != nil {
		slog.Debug("auto-update check failed", "error", err)
		return &HookOutput{}, nil
	}

	if result == nil || !result.Updated {
		return &HookOutput{}, nil
	}

	msg := fmt.Sprintf(
		"MoAI-ADK updated from %s to %s. Please restart your terminal for the new version.",
		result.PreviousVersion, result.NewVersion,
	)
	slog.Info("auto-update completed",
		"previous", result.PreviousVersion,
		"new", result.NewVersion,
	)

	return &HookOutput{SystemMessage: msg}, nil
}
