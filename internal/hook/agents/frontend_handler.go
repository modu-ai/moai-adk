package agents

import (
	"context"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// frontendHandler handles frontend development hooks.
type frontendHandler struct {
	baseHandler
}

// NewFrontendHandler creates a new frontend handler for the given action.
// Actions: validation, verification
func NewFrontendHandler(action string) hook.Handler {
	event := hook.EventPreToolUse
	if action == "verification" {
		event = hook.EventPostToolUse
	}

	return &frontendHandler{
		baseHandler: baseHandler{
			action: action,
			event:  event,
			agent:  "frontend",
		},
	}
}

// Handle processes frontend hooks.
func (h *frontendHandler) Handle(ctx context.Context, input *hook.HookInput) (*hook.HookOutput, error) {
	// TODO: Implement frontend-specific logic
	// - validation: Check component design, accessibility before changes
	// - verification: Verify UI/UX quality after changes

	return hook.NewAllowOutput(), nil
}

func (h *frontendHandler) EventType() hook.EventType {
	return h.event
}
