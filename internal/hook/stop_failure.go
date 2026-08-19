// Resolution: KEEP — error-class systemMessage for rate_limit/billing/auth failures.
package hook

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/modu-ai/moai-adk/internal/goal"
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
	case "oauth_org_not_allowed":
		msg = "Your organization does not allow this OAuth account. Check with your workspace admin."
	case "billing_error":
		msg = "Billing error detected. Check your account status."
	case "max_output_tokens":
		msg = "Output token limit reached. Try breaking the task into smaller steps."
	default:
		// Unknown or empty error type — log only, no user-facing message. The
		// disarm below is still consulted: it classifies the type itself and
		// declines every value that reaches here, so the two switches cannot
		// drift into disagreeing about which types are unrecoverable.
	}

	if note := disarmGoalOnUnrecoverable(input, errType); note != "" {
		if msg == "" {
			msg = note
		} else {
			msg += " " + note
		}
	}
	if msg == "" {
		return &HookOutput{}, nil
	}
	return &HookOutput{SystemMessage: msg}, nil
}

// disarmGoalOnUnrecoverable clears the session's armed goal when the turn died
// on something a retry cannot fix, and returns the sentence saying so — or ""
// when nothing was disarmed.
//
// The disarm exists because `/moai goal` is a reimplementation of the native
// command rather than a wrapper around it, so it inherits none of the native
// self-clear behavior. Its own doctrine states the cost: a goal armed with
// nothing running "spins idle turns until the ceiling, because each turn-end
// finds the condition unmet and no work advancing it". A revoked credential
// ends the work; without this, the goal outlives it.
//
// This is the goal loop's fourth exit and the only one that does not need turns
// to keep completing. The turn ceiling, the runtime block cap, and the
// stagnation guard are all counters over completed turns — they are why an
// armed goal eventually stops, and why it stops late. None of them change here.
//
// Best-effort throughout, matching the handler's non-blocking contract: an
// unresolvable root, an absent goal, or a failed remove skips the disarm and
// says nothing, rather than turning an error report into an error.
func disarmGoalOnUnrecoverable(input *HookInput, errType string) string {
	if !goal.IsUnrecoverableStopFailure(errType) || input == nil || input.SessionID == "" {
		return ""
	}
	root := resolveProjectRootFromInputOrEnv(input, "stop-failure")
	if root == "" {
		return ""
	}
	// Load first, and check the VALUE rather than only the error: LoadGoal
	// reports an absent goal as (nil, nil), not as an error, and ClearGoal is
	// idempotent on absence — so an error-only guard announces a disarm that
	// disarmed nothing. The pre-existing handler tests caught exactly that.
	g, err := goal.LoadGoal(root, input.SessionID)
	if err != nil || g == nil {
		return ""
	}
	if err := goal.ClearGoal(root, input.SessionID); err != nil {
		slog.Warn("stop failure: could not disarm the goal",
			"session_id", input.SessionID, "error", err)
		return ""
	}
	slog.Info("stop failure: goal disarmed",
		"session_id", input.SessionID, "error_type", errType)
	return fmt.Sprintf("The armed goal was disarmed: %s is not something a retry clears, so the goal would have spun idle turns to its ceiling.", errType)
}
