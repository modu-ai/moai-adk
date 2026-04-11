package hook

import (
	"context"
	"log/slog"
)

// setupHandler processes Setup events.
// Fired via --init, --init-only, or --maintenance CLI flags.
// Available since Claude Code v2.1.10+.
type setupHandler struct{}

// NewSetupHandler creates a new Setup event handler.
func NewSetupHandler() Handler {
	return &setupHandler{}
}

// EventType returns EventSetup.
func (h *setupHandler) EventType() EventType {
	return EventSetup
}

// Handle processes a Setup event. It logs the setup invocation for observability.
// Future enhancements may perform environment validation or initial configuration.
func (h *setupHandler) Handle(_ context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("setup event received",
		"session_id", input.SessionID,
	)
	return &HookOutput{}, nil
}
