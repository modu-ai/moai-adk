package agents

import (
	"context"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// challengeHandler handles challenge workflow hooks.
type challengeHandler struct {
	baseHandler
}

// NewChallengeHandler creates a new challenge handler for the given action.
// Actions: completion, question-generated, report-created
func NewChallengeHandler(action string) hook.Handler {
	return &challengeHandler{
		baseHandler: baseHandler{
			action: action,
			event:  hook.EventSubagentStop,
			agent:  "challenge",
		},
	}
}

// Handle processes challenge hooks.
func (h *challengeHandler) Handle(ctx context.Context, input *hook.HookInput) (*hook.HookOutput, error) {
	// TODO: Implement challenge-specific logic
	// - completion: Report challenge workflow completion
	// - question-generated: Track question generation metrics
	// - report-created: Log challenge report creation

	return hook.NewAllowOutput(), nil
}

func (h *challengeHandler) EventType() hook.EventType {
	return h.event
}
