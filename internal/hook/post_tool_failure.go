package hook

import (
	"context"
	"log/slog"
)

// postToolUseFailureHandler processes PostToolUseFailure events.
// It implements gutter detection to track recurring error patterns
// and warns users when context degradation is detected.
type postToolUseFailureHandler struct{}

// NewPostToolUseFailureHandler creates a new PostToolUseFailure event handler.
func NewPostToolUseFailureHandler() Handler {
	return &postToolUseFailureHandler{}
}

// EventType returns EventPostToolUseFailure.
func (h *postToolUseFailureHandler) EventType() EventType {
	return EventPostToolUseFailure
}

// Handle processes a PostToolUseFailure event. It logs the tool failure,
// tracks error patterns for gutter detection, and returns a system message
// when gutter state is detected (same error pattern 3+ times).
func (h *postToolUseFailureHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	// Always log the failure for diagnostics
	slog.Info("tool execution failed",
		"session_id", input.SessionID,
		"tool_name", input.ToolName,
		"tool_use_id", input.ToolUseID,
		"error", input.Error,
		"is_interrupt", input.IsInterrupt,
	)

	// Skip gutter detection for interrupts and empty errors
	if !IsGutterRelevantError(input) {
		return &HookOutput{}, nil
	}

	// Create gutter tracker
	tracker := ResolveGutterTracker(input)
	if tracker == nil {
		// No valid project root, skip gutter detection
		return &HookOutput{}, nil
	}

	// Track the failure and check for gutter detection
	gutterDetected, err := tracker.TrackFailure(input.ToolName, input.Error)
	if err != nil {
		slog.Warn("gutter tracker: failed to track failure",
			"session_id", input.SessionID,
			"tool_name", input.ToolName,
			"error", err,
		)
		return &HookOutput{}, nil
	}

	// If gutter detected, return system message warning the user
	if gutterDetected {
		// Get pattern count for the message
		signature := tracker.patternSignature(input.ToolName, input.Error)
		if pattern, exists := tracker.state.Patterns[signature]; exists {
			return &HookOutput{
				SystemMessage: GetSystemMessageForGutter(input.ToolName, pattern.Count),
			}, nil
		}
	}

	return &HookOutput{}, nil
}
