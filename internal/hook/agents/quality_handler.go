package agents

import (
	"context"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// qualityHandler handles quality verification hooks.
type qualityHandler struct {
	baseHandler
}

// NewQualityHandler creates a new quality handler for the given action.
// Actions: completion
func NewQualityHandler(action string) hook.Handler {
	return &qualityHandler{
		baseHandler: baseHandler{
			action: action,
			event:  hook.EventSubagentStop,
			agent:  "quality",
		},
	}
}

// Handle processes quality hooks.
func (h *qualityHandler) Handle(ctx context.Context, input *hook.HookInput) (*hook.HookOutput, error) {
	// TODO: Implement quality-specific logic
	// - completion: Report quality gate completion

	return hook.NewAllowOutput(), nil
}

func (h *qualityHandler) EventType() hook.EventType {
	return h.event
}
