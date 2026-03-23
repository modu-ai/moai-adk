package hook

import (
	"context"
	"log/slog"
)

// instructionsLoadedHandler processes InstructionsLoaded events.
// It validates project configuration when CLAUDE.md or rules files are loaded into context.
type instructionsLoadedHandler struct{}

// NewInstructionsLoadedHandler creates a new InstructionsLoaded event handler.
func NewInstructionsLoadedHandler() Handler {
	return &instructionsLoadedHandler{}
}

// EventType returns EventInstructionsLoaded.
func (h *instructionsLoadedHandler) EventType() EventType {
	return EventInstructionsLoaded
}

// Handle processes an InstructionsLoaded event. It logs the loading event
// for debugging context issues. Errors are non-blocking.
func (h *instructionsLoadedHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("instructions loaded",
		"session_id", input.SessionID,
	)

	return &HookOutput{}, nil
}
