package agents

import (
	"context"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// backendHandler handles backend development hooks.
type backendHandler struct {
	baseHandler
}

// NewBackendHandler creates a new backend handler for the given action.
// Actions: validation, verification
func NewBackendHandler(action string) hook.Handler {
	event := hook.EventPreToolUse
	if action == "verification" {
		event = hook.EventPostToolUse
	}

	return &backendHandler{
		baseHandler: baseHandler{
			action: action,
			event:  event,
			agent:  "backend",
		},
	}
}

// Handle processes backend hooks.
func (h *backendHandler) Handle(ctx context.Context, input *hook.HookInput) (*hook.HookOutput, error) {
	// TODO: Implement backend-specific logic
	// - validation: Check API design, database schema before changes
	// - verification: Verify backend code quality after changes

	return hook.NewAllowOutput(), nil
}

func (h *backendHandler) EventType() hook.EventType {
	return h.event
}
