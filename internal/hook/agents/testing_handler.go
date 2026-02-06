package agents

import (
	"context"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// testingHandler handles testing strategy hooks.
type testingHandler struct {
	baseHandler
}

// NewTestingHandler creates a new testing handler for the given action.
// Actions: verification, completion
func NewTestingHandler(action string) hook.Handler {
	event := hook.EventPostToolUse
	if action == "completion" {
		event = hook.EventSubagentStop
	}

	return &testingHandler{
		baseHandler: baseHandler{
			action: action,
			event:  event,
			agent:  "testing",
		},
	}
}

// Handle processes testing hooks.
func (h *testingHandler) Handle(ctx context.Context, input *hook.HookInput) (*hook.HookOutput, error) {
	// TODO: Implement testing-specific logic
	// - verification: Verify test quality, coverage after test changes
	// - completion: Report testing strategy completion

	return hook.NewAllowOutput(), nil
}

func (h *testingHandler) EventType() hook.EventType {
	return h.event
}
