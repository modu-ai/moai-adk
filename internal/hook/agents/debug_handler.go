package agents

import (
	"context"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// debugHandler handles debugging hooks.
type debugHandler struct {
	baseHandler
}

// NewDebugHandler creates a new debug handler for the given action.
// Actions: verification, completion
func NewDebugHandler(action string) hook.Handler {
	event := hook.EventPostToolUse
	if action == "completion" {
		event = hook.EventSubagentStop
	}

	return &debugHandler{
		baseHandler: baseHandler{
			action: action,
			event:  event,
			agent:  "debug",
		},
	}
}

// Handle processes debug hooks.
func (h *debugHandler) Handle(ctx context.Context, input *hook.HookInput) (*hook.HookOutput, error) {
	// TODO: Implement debug-specific logic
	// - verification: Verify diagnostic analysis accuracy
	// - completion: Report debugging session completion

	return hook.NewAllowOutput(), nil
}

func (h *debugHandler) EventType() hook.EventType {
	return h.event
}
