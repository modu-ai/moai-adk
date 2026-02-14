package hook

import (
	"context"
	"log/slog"
)

// userPromptSubmitHandler processes UserPromptSubmit events.
// It logs user prompt submissions for auditing.
type userPromptSubmitHandler struct{}

// NewUserPromptSubmitHandler creates a new UserPromptSubmit event handler.
func NewUserPromptSubmitHandler() Handler {
	return &userPromptSubmitHandler{}
}

// EventType returns EventUserPromptSubmit.
func (h *userPromptSubmitHandler) EventType() EventType {
	return EventUserPromptSubmit
}

// Handle processes a UserPromptSubmit event. It logs the prompt submission.
// The prompt is truncated to 100 characters for privacy.
func (h *userPromptSubmitHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	prompt := input.Prompt
	if len(prompt) > 100 {
		prompt = prompt[:100] + "..."
	}
	slog.Info("user prompt submitted",
		"session_id", input.SessionID,
		"prompt_preview", prompt,
	)
	return &HookOutput{}, nil
}
