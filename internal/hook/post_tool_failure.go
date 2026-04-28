package hook

import (
	"context"
	"log/slog"
	"strings"
)

// ErrorCategory represents a classification of tool execution failures.
type ErrorCategory string

const (
	// TimeoutError indicates the tool timed out.
	TimeoutError ErrorCategory = "TimeoutError"

	// PermissionDenied indicates the tool was denied permission.
	PermissionDenied ErrorCategory = "PermissionDenied"

	// ContextCancelled indicates the operation was cancelled.
	ContextCancelled ErrorCategory = "ContextCancelled"

	// SandboxViolation indicates a sandbox policy violation.
	SandboxViolation ErrorCategory = "SandboxViolation"

	// OOMKilled indicates the process was killed for OOM.
	OOMKilled ErrorCategory = "OOMKilled"

	// ExitError indicates a non-zero exit code.
	ExitError ErrorCategory = "ExitError"

	// UnknownFailure indicates an unclassified failure.
	UnknownFailure ErrorCategory = "UnknownFailure"
)

// postToolUseFailureHandler processes PostToolUseFailure events.
// It classifies errors by signature and provides actionable messages.
type postToolUseFailureHandler struct{}

// NewPostToolUseFailureHandler creates a new PostToolUseFailure event handler.
func NewPostToolUseFailureHandler() Handler {
	return &postToolUseFailureHandler{}
}

// EventType returns EventPostToolUseFailure.
func (h *postToolUseFailureHandler) EventType() EventType {
	return EventPostToolUseFailure
}

// Handle processes a PostToolUseFailure event. It classifies the error
// and returns a system message with actionable guidance.
func (h *postToolUseFailureHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	category := h.classifyError(input)
	message := h.formatMessage(category, input)

	slog.Info("tool execution failed",
		"session_id", input.SessionID,
		"tool_name", input.ToolName,
		"tool_use_id", input.ToolUseID,
		"error", input.Error,
		"is_interrupt", input.IsInterrupt,
		"category", category,
	)

	return &HookOutput{
		SystemMessage: message,
	}, nil
}

// classifyError determines the error category based on error signature.
func (h *postToolUseFailureHandler) classifyError(input *HookInput) ErrorCategory {
	errorText := strings.ToLower(input.Error)

	// Check for timeout (exit code 124 or "timeout" in error message)
	if strings.Contains(errorText, "timeout") || strings.Contains(errorText, "context deadline exceeded") {
		return TimeoutError
	}

	// Check for permission denied
	if strings.Contains(errorText, "permission denied") || strings.Contains(errorText, "access denied") {
		return PermissionDenied
	}

	// Check for context cancellation
	if input.IsInterrupt || strings.Contains(errorText, "context canceled") {
		return ContextCancelled
	}

	// Check for sandbox violation
	if strings.Contains(errorText, "sandbox") || strings.Contains(errorText, "seccomp") {
		return SandboxViolation
	}

	// Check for OOM killed (exit code 137 or "oom" in error)
	if strings.Contains(errorText, "oom") || strings.Contains(errorText, "out of memory") || strings.Contains(errorText, "137") {
		return OOMKilled
	}

	// Default to exit error for any non-zero exit
	if strings.Contains(errorText, "exit status") || strings.Contains(errorText, "non-zero") {
		return ExitError
	}

	// Fallback to unknown
	return UnknownFailure
}

// formatMessage generates an actionable message for the error category.
func (h *postToolUseFailureHandler) formatMessage(category ErrorCategory, _ *HookInput) string {
	switch category {
	case TimeoutError:
		return "TimeoutError: Tool execution exceeded time limit. Consider optimizing the operation or increasing timeout settings."

	case PermissionDenied:
		return "PermissionDenied: Tool was denied permission. Check settings.json permissions or run in bypassPermissions mode."

	case ContextCancelled:
		return "ContextCancelled: Operation was cancelled by user or system. No action needed."

	case SandboxViolation:
		return "SandboxViolation: Tool violated sandbox policy. Review the operation and adjust permissions if necessary."

	case OOMKilled:
		return "OOMKilled: Process was terminated for exceeding memory limits. Reduce memory usage or increase system resources."

	case ExitError:
		return "ExitError: Tool exited with non-zero status. Check tool logs for details."

	default:
		return "UnknownFailure: Tool execution failed. Review error logs for details."
	}
}
