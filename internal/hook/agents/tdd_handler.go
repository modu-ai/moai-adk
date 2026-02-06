package agents

import (
	"context"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// tddHandler handles TDD (Test-Driven Development) workflow hooks.
type tddHandler struct {
	baseHandler
}

// NewTDDHandler creates a new TDD handler for the given action.
// Actions: pre-implementation, post-implementation, completion
func NewTDDHandler(action string) hook.Handler {
	event := hook.EventPreToolUse
	switch action {
	case "post-implementation":
		event = hook.EventPostToolUse
	case "completion":
		event = hook.EventSubagentStop
	}

	return &tddHandler{
		baseHandler: baseHandler{
			action: action,
			event:  event,
			agent:  "tdd",
		},
	}
}

// Handle processes TDD workflow hooks.
func (h *tddHandler) Handle(ctx context.Context, input *hook.HookInput) (*hook.HookOutput, error) {
	// TODO: Implement TDD workflow logic
	// - pre-implementation: RED phase - ensure test exists before implementation
	// - post-implementation: GREEN/REFACTOR phase - verify tests pass
	// - completion: Report TDD workflow completion

	return hook.NewAllowOutput(), nil
}

func (h *tddHandler) EventType() hook.EventType {
	return h.event
}
