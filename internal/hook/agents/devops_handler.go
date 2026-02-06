package agents

import (
	"context"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// devOpsHandler handles DevOps hooks.
type devOpsHandler struct {
	baseHandler
}

// NewDevOpsHandler creates a new DevOps handler for the given action.
// Actions: verification, completion
func NewDevOpsHandler(action string) hook.Handler {
	event := hook.EventPostToolUse
	if action == "completion" {
		event = hook.EventSubagentStop
	}

	return &devOpsHandler{
		baseHandler: baseHandler{
			action: action,
			event:  event,
			agent:  "devops",
		},
	}
}

// Handle processes DevOps hooks.
func (h *devOpsHandler) Handle(ctx context.Context, input *hook.HookInput) (*hook.HookOutput, error) {
	// TODO: Implement DevOps-specific logic
	// - verification: Verify CI/CD configuration, deployment settings
	// - completion: Report DevOps task completion

	return hook.NewAllowOutput(), nil
}

func (h *devOpsHandler) EventType() hook.EventType {
	return h.event
}
