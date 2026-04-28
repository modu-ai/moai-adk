// Package permission implements an 8-tier permission stack with bubble mode support.
// This is the core of MoAI's safety architecture, providing typed permission resolution
// with provenance tracking for every decision.
package permission

import (
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// PermissionMode defines how agent permissions are resolved.
// These values map directly to Claude Code's permission modes.
type PermissionMode string

const (
	// ModeDefault is the standard permission mode - tool invocations are
	// resolved through the 8-tier stack, prompting for unknown operations.
	ModeDefault PermissionMode = "default"

	// ModeAcceptEdits allows all Read/Write/Edit operations without prompting,
	// but still requires permission for destructive operations (e.g., Bash with write flags).
	ModeAcceptEdits PermissionMode = "acceptEdits"

	// ModeBypassPermissions allows all operations without prompting.
	// This mode should be rejected when security.yaml permission.strict_mode is true.
	ModeBypassPermissions PermissionMode = "bypassPermissions"

	// ModePlan denies all write operations regardless of allowlist.
	// Used for planning and analysis phases where code modification must be blocked.
	ModePlan PermissionMode = "plan"

	// ModeBubble routes permission prompts to the parent session's AskUserQuestion
	// channel instead of the fork agent's own channel. This is the default for
	// fork agents that inherit parent context.
	ModeBubble PermissionMode = "bubble"
)

// String returns the string representation of the permission mode.
func (m PermissionMode) String() string {
	return string(m)
}

// ParsePermissionMode converts a string to a PermissionMode.
// Returns an error if the string is not a valid mode.
func ParsePermissionMode(s string) (PermissionMode, error) {
	mode := PermissionMode(s)
	switch mode {
	case ModeDefault, ModeAcceptEdits, ModeBypassPermissions, ModePlan, ModeBubble:
		return mode, nil
	default:
		return ModeDefault, fmt.Errorf("invalid permission mode: %s", s)
	}
}

// IsValid returns true if the PermissionMode is one of the valid values.
func (m PermissionMode) IsValid() bool {
	switch m {
	case ModeDefault, ModeAcceptEdits, ModeBypassPermissions, ModePlan, ModeBubble:
		return true
	default:
		return false
	}
}

// PermissionRule represents a single permission rule in the stack.
// Each rule carries provenance information (Origin) identifying which file
// contributed the rule, enabling audit trails and debugging.
type PermissionRule struct {
	// Pattern is the glob-style pattern to match against tool invocations.
	// Examples: "Bash(go test:*)", "Read(*)", "Write(/tmp/*)"
	Pattern string

	// Action is the decision to apply when the pattern matches.
	Action Decision

	// Source is the configuration tier that contributed this rule.
	// Higher priority sources override lower priority sources.
	Source config.Source

	// Origin is the file path that contributed this rule.
	// Examples: ".claude/settings.json", "/etc/moai/settings.json"
	// This field enables provenance tracking and debugging.
	Origin string
}

// Matches returns true if the rule pattern matches the given tool invocation.
// Pattern format: "ToolName(pattern)" or "ToolName:*" for wildcard arguments.
// Input format: "ToolName actual_arguments" or just "ToolName".
//
// Pattern examples:
//   - "Bash(go test:*)" matches "Bash" tool with input starting with "go test"
//   - "Read(*)" matches "Read" tool with any input
//   - "Write(/tmp/*)" matches "Write" tool with input starting with "/tmp/"
//   - "*" matches all tools and inputs
func (r *PermissionRule) Matches(tool, input string) bool {
	if r.Pattern == "*" {
		return true
	}

	// Split pattern into tool and argument parts
	patternParts := strings.SplitN(r.Pattern, "(", 2)
	if len(patternParts) < 2 {
		// Simple tool name pattern (no args)
		return r.Pattern == tool
	}

	patternTool := patternParts[0]
	if patternTool != tool && patternTool != "*" {
		return false
	}

	// Extract argument pattern (before closing paren)
	argPattern := strings.TrimSuffix(patternParts[1], ")")

	// Wildcard argument pattern matches any input
	if argPattern == "*" {
		return true
	}

	// Check if input matches the argument pattern
	// For patterns like "go test:*", check if input starts with "go test:"
	if prefix, ok := strings.CutSuffix(argPattern, ":*"); ok {
		return strings.HasPrefix(input, prefix)
	}

	// For patterns like "/tmp/*", check if input starts with "/tmp/"
	if prefix, ok := strings.CutSuffix(argPattern, "/*"); ok {
		return strings.HasPrefix(input, prefix)
	}

	// For patterns like "*.go", check if input ends with ".go"
	if suffix, ok := strings.CutPrefix(argPattern, "*."); ok {
		return strings.HasSuffix(input, suffix)
	}

	// Exact match
	return input == argPattern
}

// String returns a string representation of the rule for debugging.
func (r *PermissionRule) String() string {
	return fmt.Sprintf("%s|%s|%s|%s", r.Source, r.Action, r.Origin, r.Pattern)
}

// Decision represents a permission decision outcome.
type Decision string

const (
	// DecisionAllow permits the operation to proceed.
	DecisionAllow Decision = "allow"

	// DecisionAsk prompts the user for confirmation.
	// In bubble mode, the prompt is routed to the parent session.
	DecisionAsk Decision = "ask"

	// DecisionDeny blocks the operation.
	DecisionDeny Decision = "deny"
)

// String returns the string representation of the decision.
func (d Decision) String() string {
	return string(d)
}

// PreAllowlist returns the built-in pre-allowlist for common development operations.
// These rules are at SrcBuiltin tier and cover ~80% of read/verify operations to
// reduce bubble fatigue. The pre-allowlist is always active and cannot be overridden
// by lower-tier rules.
//
// The pre-allowlist includes:
//   - Read operations: Read(*), Glob(*), Grep(*)
//   - Common test commands: go test, golangci-lint run, ruff check, npm test, pytest
//
// These patterns are considered safe because they:
//   - Do not modify files
//   - Are read-only verification operations
//   - Are standard development workflows
//
// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-006
func PreAllowlist() []PermissionRule {
	return []PermissionRule{
		{
			Pattern: "Read(*)",
			Action:  DecisionAllow,
			Source:  config.SrcBuiltin,
			Origin:  "pre-allowlist",
		},
		{
			Pattern: "Glob(*)",
			Action:  DecisionAllow,
			Source:  config.SrcBuiltin,
			Origin:  "pre-allowlist",
		},
		{
			Pattern: "Grep(*)",
			Action:  DecisionAllow,
			Source:  config.SrcBuiltin,
			Origin:  "pre-allowlist",
		},
		{
			Pattern: "Bash(go test:*)",
			Action:  DecisionAllow,
			Source:  config.SrcBuiltin,
			Origin:  "pre-allowlist",
		},
		{
			Pattern: "Bash(golangci-lint run:*)",
			Action:  DecisionAllow,
			Source:  config.SrcBuiltin,
			Origin:  "pre-allowlist",
		},
		{
			Pattern: "Bash(ruff check:*)",
			Action:  DecisionAllow,
			Source:  config.SrcBuiltin,
			Origin:  "pre-allowlist",
		},
		{
			Pattern: "Bash(npm test:*)",
			Action:  DecisionAllow,
			Source:  config.SrcBuiltin,
			Origin:  "pre-allowlist",
		},
		{
			Pattern: "Bash(pytest:*)",
			Action:  DecisionAllow,
			Source:  config.SrcBuiltin,
			Origin:  "pre-allowlist",
		},
	}
}

// IsWriteOperation returns true if the tool invocation is a write operation.
// Write operations include:
//   - Write tool
//   - Edit tool
//   - Bash with known write patterns (rm, mv, cp, git push, etc.)
//
// This is used to enforce ModePlan, which denies all write operations.
//
// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-020
func IsWriteOperation(tool, input string) bool {
	switch tool {
	case "Write", "Edit":
		return true
	case "Bash":
		// Check for known write patterns in bash commands
		writePatterns := []string{
			"rm ", "rmdir ",
			"mv ", "mv ", // move/rename
			"cp ", "copy ",
			"mkdir ", "mktemp ",
			"touch ",
			"echo ", "printf ", "cat >", "tee ",
			"git commit ", "git push ", "git pull ",
			"git merge ", "git rebase ",
			"npm install ", "pip install ", "go get ",
			"docker build", "docker push ",
		}
		inputLower := strings.ToLower(input)
		for _, pattern := range writePatterns {
			if strings.Contains(inputLower, pattern) {
				return true
			}
		}
	}
	return false
}
