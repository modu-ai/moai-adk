package agents

import (
	"context"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// defaultHandler handles unknown agent actions.
type defaultHandler struct {
	baseHandler
}

// NewDefaultHandler creates a default handler for unknown actions.
func NewDefaultHandler(action string) hook.Handler {
	return &defaultHandler{
		baseHandler: baseHandler{
			action: action,
			event:  hook.EventPreToolUse, // Default to PreToolUse
			agent:  "default",
		},
	}
}

// Handle processes unknown agent actions by allowing them.
func (h *defaultHandler) Handle(ctx context.Context, input *hook.HookInput) (*hook.HookOutput, error) {
	// Allow unknown actions by default
	return hook.NewAllowOutput(), nil
}

func (h *defaultHandler) EventType() hook.EventType {
	return h.event
}
