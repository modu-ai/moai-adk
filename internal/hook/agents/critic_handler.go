package agents

import (
	"context"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// criticHandler handles critic workflow hooks.
type criticHandler struct {
	baseHandler
}

// NewCriticHandler creates a new critic handler for the given action.
// Actions: completion, question-generated, report-created
func NewCriticHandler(action string) hook.Handler {
	return &criticHandler{
		baseHandler: baseHandler{
			action: action,
			event:  hook.EventSubagentStop,
			agent:  "critic",
		},
	}
}

// Handle processes critic hooks.
func (h *criticHandler) Handle(ctx context.Context, input *hook.HookInput) (*hook.HookOutput, error) {
	// TODO: Implement critic-specific logic
	// - completion: Report critic workflow completion
	// - question-generated: Track question generation metrics
	// - report-created: Log critic report creation

	return hook.NewAllowOutput(), nil
}

func (h *criticHandler) EventType() hook.EventType {
	return h.event
}
