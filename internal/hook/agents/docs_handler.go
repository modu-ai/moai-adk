package agents

import (
	"context"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// docsHandler handles documentation hooks.
type docsHandler struct {
	baseHandler
}

// NewDocsHandler creates a new docs handler for the given action.
// Actions: verification, completion
func NewDocsHandler(action string) hook.Handler {
	event := hook.EventPostToolUse
	if action == "completion" {
		event = hook.EventSubagentStop
	}

	return &docsHandler{
		baseHandler: baseHandler{
			action: action,
			event:  event,
			agent:  "docs",
		},
	}
}

// Handle processes documentation hooks.
func (h *docsHandler) Handle(ctx context.Context, input *hook.HookInput) (*hook.HookOutput, error) {
	// TODO: Implement documentation-specific logic
	// - verification: Verify generated documentation quality
	// - completion: Report documentation generation completion

	return hook.NewAllowOutput(), nil
}

func (h *docsHandler) EventType() hook.EventType {
	return h.event
}
