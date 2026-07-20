// Resolution: UPGRADE — error classification (ExitError/TimeoutError/ContextCancelled/PermissionDenied/SandboxViolation/OOMKilled/UnknownFailure).
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
//
// SPEC-HARNESS-RATCHET-REWIRE-001 REQ-HRR-001: the handler also records a
// tool_failure:<tool>:<category> event to usage-log.jsonl so the failure signal
// enters the harness learning loop. The category (low-cardinality error-class
// token) is the <signature>. Recording is fail-open (failure_observer.go logs
// and swallows errors) — a learning-loop write failure never blocks the user's
// session end.
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

	// REQ-HRR-001: feed the failure signal into the learning loop. The category
	// is the low-cardinality <signature> token (D1 resolution). EC-1: an empty
	// tool name is substituted with a non-empty placeholder inside the recorder
	// so no degenerate tool_failure::<sig> key is emitted.
	recordToolFailureEvent(input, category)

	// REQ-CC2122-HOOK-001-006: PostToolUseFailure from/in/at outcome: includes "failure" field.
	writeHookMetric(input, "handle-post-tool-failure", "failure")

	return &HookOutput{
		SystemMessage: message,
	}, nil
}

// classificationText aggregates every available error-text source into a
// single haystack for the ordered category matchers (REQ-HFC-001/002).
// Real Claude Code PostToolUseFailure payloads nest the failure text under
// tool_response.error / tool_response.stderr with an empty top-level error
// field (issue #1089); the union widens the haystack without changing the
// ordered first-match semantics. tool_response normalization reuses
// decodeToolResponse (evidence_writer.go), which degrades gracefully on
// malformed / bare-string / absent payloads (REQ-HFC-006 fail-open).
func classificationText(input *HookInput) string {
	var b strings.Builder
	b.WriteString(input.Error)
	if len(input.ToolResponse) > 0 {
		if text, _ := decodeToolResponse(input.ToolResponse); text != "" {
			b.WriteByte('\n')
			b.WriteString(text)
		}
	}
	return b.String()
}

// maxErrorExcerptLen bounds the raw-error excerpt appended to the
// UnknownFailure message (REQ-HFC-005).
const maxErrorExcerptLen = 200

// rawErrorExcerpt returns a bounded excerpt of the observed raw error text.
// Precedence (REQ-HFC-002): the top-level error field wins for the
// human-facing excerpt; nested tool_response text is the fallback.
// Returns "" when no raw text is observable.
func rawErrorExcerpt(input *HookInput) string {
	text := strings.TrimSpace(input.Error)
	if text == "" && len(input.ToolResponse) > 0 {
		decoded, _ := decodeToolResponse(input.ToolResponse)
		text = strings.TrimSpace(decoded)
	}
	if text == "" {
		return ""
	}
	runes := []rune(text)
	if len(runes) > maxErrorExcerptLen {
		return string(runes[:maxErrorExcerptLen]) + "..."
	}
	return text
}

// classifyError determines the error category based on error signature.
func (h *postToolUseFailureHandler) classifyError(input *HookInput) ErrorCategory {
	errorText := strings.ToLower(classificationText(input))

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
func (h *postToolUseFailureHandler) formatMessage(category ErrorCategory, input *HookInput) string {
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
		// REQ-HFC-005: never emit a content-free UnknownFailure when raw error
		// text is observable — append a bounded excerpt so the model receives
		// an actionable signal. The fixed string remains only for payloads
		// with no observable text at all.
		if excerpt := rawErrorExcerpt(input); excerpt != "" {
			return "UnknownFailure: Tool execution failed: " + excerpt
		}
		return "UnknownFailure: Tool execution failed. Review error logs for details."
	}
}
