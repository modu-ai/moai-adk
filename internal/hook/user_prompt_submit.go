package hook

import (
	"context"
	"log/slog"
	"strings"
)

// userPromptSubmitHandler processes UserPromptSubmit events.
// It logs user prompt submissions for auditing and injects workflow context
// when workflow keywords are detected in the prompt.
type userPromptSubmitHandler struct{}

// NewUserPromptSubmitHandler creates a new UserPromptSubmit event handler.
func NewUserPromptSubmitHandler() Handler {
	return &userPromptSubmitHandler{}
}

// EventType returns EventUserPromptSubmit.
func (h *userPromptSubmitHandler) EventType() EventType {
	return EventUserPromptSubmit
}

// workflowKeywords are prompt keywords that indicate an active MoAI workflow context.
var workflowKeywords = []string{"loop", "run", "plan"}

// detectWorkflowContext checks whether the prompt contains any workflow keywords
// and returns a non-empty additionalContext string if a match is found.
func detectWorkflowContext(prompt string) string {
	lower := strings.ToLower(prompt)
	for _, kw := range workflowKeywords {
		if strings.Contains(lower, kw) {
			return "workflow keyword '" + kw + "' detected — MoAI workflow context may be active"
		}
	}
	return ""
}

// Handle processes a UserPromptSubmit event. It logs the prompt submission.
// The prompt is truncated to 100 characters for privacy.
// When workflow keywords are found the output carries an additionalContext hint.
func (h *userPromptSubmitHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	prompt := input.Prompt
	preview := prompt
	if len(preview) > 100 {
		preview = preview[:100] + "..."
	}
	slog.Info("user prompt submitted",
		"session_id", input.SessionID,
		"prompt_preview", preview,
	)

	additionalCtx := detectWorkflowContext(prompt)
	if additionalCtx == "" {
		return &HookOutput{}, nil
	}

	slog.Info("workflow context detected in prompt",
		"session_id", input.SessionID,
		"additional_context", additionalCtx,
	)

	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:     "UserPromptSubmit",
			AdditionalContext: additionalCtx,
		},
	}, nil
}
