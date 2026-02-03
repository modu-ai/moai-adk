package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

// SecurityPolicy defines tool access control rules for PreToolUse events.
type SecurityPolicy struct {
	// BlockedTools is a list of tool names that are always blocked.
	BlockedTools []string

	// DangerousPatterns is a list of input patterns that trigger blocking
	// when detected in Bash tool commands.
	DangerousPatterns []string
}

// DefaultSecurityPolicy returns a SecurityPolicy with sensible defaults.
func DefaultSecurityPolicy() *SecurityPolicy {
	return &SecurityPolicy{
		BlockedTools: []string{},
		DangerousPatterns: []string{
			"rm -rf /",
			"rm -rf /*",
			":(){ :|:& };:",
			"mkfs.",
			"> /dev/sda",
			"dd if=/dev/zero of=/dev/sda",
		},
	}
}

// preToolHandler processes PreToolUse events.
// It enforces security policies by checking tool names against blocklists
// and scanning tool input for dangerous patterns (REQ-HOOK-031, REQ-HOOK-032).
type preToolHandler struct {
	cfg    ConfigProvider
	policy *SecurityPolicy
}

// NewPreToolHandler creates a new PreToolUse event handler with the given security policy.
func NewPreToolHandler(cfg ConfigProvider, policy *SecurityPolicy) Handler {
	return &preToolHandler{cfg: cfg, policy: policy}
}

// EventType returns EventPreToolUse.
func (h *preToolHandler) EventType() EventType {
	return EventPreToolUse
}

// Handle processes a PreToolUse event. It checks the tool name against the
// blocklist and scans tool input for dangerous patterns. Returns Decision
// "block" with a reason if the tool is denied, or "allow" otherwise.
func (h *preToolHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	// No policy means allow everything
	if h.policy == nil {
		return NewAllowOutput(), nil
	}

	slog.Debug("checking tool security",
		"tool_name", input.ToolName,
		"session_id", input.SessionID,
	)

	// Check if tool is in the blocked list
	for _, blocked := range h.policy.BlockedTools {
		if strings.EqualFold(input.ToolName, blocked) {
			reason := fmt.Sprintf("tool %q is blocked by security policy", input.ToolName)
			slog.Warn("tool blocked",
				"tool_name", input.ToolName,
				"reason", reason,
			)
			return NewBlockOutput(reason), nil
		}
	}

	// Check tool input for dangerous patterns (applies to Bash tool)
	if input.ToolName == "Bash" && len(input.ToolInput) > 0 {
		if reason := h.checkDangerousInput(input.ToolInput); reason != "" {
			slog.Warn("dangerous input detected",
				"tool_name", input.ToolName,
				"reason", reason,
			)
			return NewBlockOutput(reason), nil
		}
	}

	return NewAllowOutput(), nil
}

// checkDangerousInput scans tool input JSON for dangerous command patterns.
// Returns a reason string if a dangerous pattern is found, empty string otherwise.
func (h *preToolHandler) checkDangerousInput(toolInput json.RawMessage) string {
	var parsed map[string]any
	if err := json.Unmarshal(toolInput, &parsed); err != nil {
		// Cannot parse input; allow by default (do not block on parse errors)
		return ""
	}

	command, ok := parsed["command"].(string)
	if !ok {
		return ""
	}

	for _, pattern := range h.policy.DangerousPatterns {
		if strings.Contains(command, pattern) {
			return fmt.Sprintf("dangerous command pattern detected: %q", pattern)
		}
	}

	return ""
}
