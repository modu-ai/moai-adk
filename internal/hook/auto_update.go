package hook

import (
	"context"
	"fmt"
	"log/slog"
)

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
func (h *autoUpdateHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	if h.updateFn == nil {
		return &HookOutput{}, nil
	}

	result, err := h.updateFn(ctx)
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
