package agents

import (
	"context"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// specHandler handles SPEC documentation hooks.
type specHandler struct {
	baseHandler
}

// NewSpecHandler creates a new SPEC handler for the given action.
// Actions: completion
func NewSpecHandler(action string) hook.Handler {
	return &specHandler{
		baseHandler: baseHandler{
			action: action,
			event:  hook.EventSubagentStop,
			agent:  "spec",
		},
	}
}

// Handle processes SPEC hooks.
func (h *specHandler) Handle(ctx context.Context, input *hook.HookInput) (*hook.HookOutput, error) {
	// TODO: Implement SPEC-specific logic
	// - completion: Report SPEC document generation completion

	return hook.NewAllowOutput(), nil
}

func (h *specHandler) EventType() hook.EventType {
	return h.event
}
