package hook

import (
	"context"
	"log/slog"

	"github.com/modu-ai/moai-adk/internal/hook/security"
)

// postToolGuardianHandler runs the regex security guardian's buffer scan inside
// the post-tool handler's own process for Write / Edit / MultiEdit events,
// replacing the separate `moai hook security-scan` subprocess the settings
// PostToolUse matcher used to spawn. The subcommand itself is retained for user
// projects whose settings still name it.
//
// The handler is advisory-only: it emits the guardian banner and findings on
// hookSpecificOutput.additionalContext and never a decision, never a non-zero
// exit. Carrying the advisory on additionalContext — rather than folding it into
// systemMessage, which the post-tool handler already owns — is what lets the
// registry's mergeHandlerOutput keep both advisories on a single event.
//
// @MX:ANCHOR: [AUTO] postToolGuardianHandler is the in-process PostToolUse guardian, registered alongside the post-tool handler in internal/cli/deps.go.
// @MX:REASON: [AUTO] fan_in >= 3 — deps.go registration, the merge regression test, and the registry always-run tail all depend on its advisory-only contract.
type postToolGuardianHandler struct{}

// NewPostToolGuardianHandler creates the in-process PostToolUse security
// guardian handler.
func NewPostToolGuardianHandler() Handler {
	return &postToolGuardianHandler{}
}

// EventType returns EventPostToolUse.
func (h *postToolGuardianHandler) EventType() EventType {
	return EventPostToolUse
}

// AlwaysRun marks the guardian as reachable even when a preceding handler
// short-circuits Dispatch. A future PostToolUse handler that blocks must not be
// able to silently mute the security scan.
func (h *postToolGuardianHandler) AlwaysRun() bool {
	return true
}

// Handle scans the written content and returns the guardian advisory, or nil
// when there is nothing to say. Every path is fail-open: a payload the scanner
// cannot read yields a silent nil, never an error.
func (h *postToolGuardianHandler) Handle(_ context.Context, input *HookInput) (*HookOutput, error) {
	if input == nil {
		return nil, nil
	}
	switch input.ToolName {
	case "Write", "Edit", "MultiEdit":
	default:
		return nil, nil
	}

	content := security.ExtractToolInputContent(input.ToolInput)
	if content == "" {
		return nil, nil
	}
	advisory := security.ScanBufferAdvisory(content)
	if advisory == "" {
		return nil, nil
	}

	slog.Debug("security guardian finding surfaced",
		"tool_name", input.ToolName,
		"session_id", input.SessionID,
	)

	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:     "PostToolUse",
			AdditionalContext: advisory,
		},
	}, nil
}
