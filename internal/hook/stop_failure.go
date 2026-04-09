package hook

import (
	"context"
	"log/slog"
)

// stopFailureHandler processes StopFailure events.
// It logs API errors (rate limit, auth failure) that caused the turn to end abnormally,
// and returns a user-facing systemMessage for actionable error types.
type stopFailureHandler struct{}

// NewStopFailureHandler creates a new StopFailure event handler.
func NewStopFailureHandler() Handler {
	return &stopFailureHandler{}
}

// EventType returns EventStopFailure.
func (h *stopFailureHandler) EventType() EventType {
	return EventStopFailure
}

// Handle processes a StopFailure event.
// It checks input.ErrorType (v2.1.78+ protocol) and falls back to input.Error
// for older protocol versions. Returns a systemMessage for known error types.
// Always non-blocking — never returns an error.
func (h *stopFailureHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	// Determine the effective error type: prefer ErrorType, fall back to Error.
	errType := input.ErrorType
	if errType == "" {
		errType = input.Error
	}

	slog.Warn("stop failure: turn ended due to API error",
		"session_id", input.SessionID,
		"error_type", errType,
		"error_message", input.ErrorMessage,
	)

	var msg string
	switch errType {
	case "rate_limit":
		msg = "Rate limit reached. Wait a moment before continuing."
	case "authentication_failed":
		msg = "Authentication failed. Check your API key or run 'moai glm setup'."
	case "billing_error":
		msg = "Billing error detected. Check your account status."
	case "max_output_tokens":
		msg = "Output token limit reached. Try breaking the task into smaller steps."
	default:
		// Unknown or empty error type — log only, no user-facing message.
		return &HookOutput{}, nil
	}

	return &HookOutput{SystemMessage: msg}, nil
}
