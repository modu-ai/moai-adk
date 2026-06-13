// Package permission implements an 8-tier permission stack with bubble mode support.
// This is the core of MoAI's safety architecture, providing typed permission resolution
// with provenance tracking for every decision.
package permission

import (
	"fmt"
	"strings"
	"sync"

	"github.com/modu-ai/moai-adk/internal/config"
)

// @MX:NOTE: [AUTO] PermissionMode 5-enum — CC official modes (default/acceptEdits/bypassPermissions/plan/bubble)
// @MX:NOTE: [AUTO] bubble is a first-class mode that escalates a fork agent's prompts to the parent session's AskUserQuestion channel
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
		if !strings.HasPrefix(input, prefix) {
			return false
		}
		// SECURITY (SPEC-SEC-HARDEN-001 M1): a ":*" prefix rule must not be a
		// command-chain bypass. If the remainder past the matched prefix carries
		// an unquoted shell command separator, the chained command rides in on an
		// allowed prefix — so report no match and let the input fall through to the
		// normal ask/deny path instead of being silently allowed.
		return !hasUnquotedShellSeparator(input[len(prefix):])
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

// hasUnquotedShellSeparator reports whether s contains a shell command separator
// (`;`, `&&`, `||`, `|`, `$(`, backtick, or newline) outside of a single- or
// double-quoted segment. It is a deliberately lightweight lexical scan, NOT a full
// POSIX shell parser (here-docs, escaped quotes, `$'...'`, process substitution are
// out of scope per SPEC-SEC-HARDEN-001 §F.4). The scan errs toward reporting a
// separator when ambiguous, which is the security-safe direction for an allow-rule
// matcher: a borderline input fails to match the prefix rule and falls through to
// the normal ask/deny path rather than being silently allowed.
//
// SECURITY (SPEC-SEC-HARDEN-001 M1 D1): an UNTERMINATED quote — the scan reaching
// end-of-string while still inside a single- or double-quoted segment — is itself
// ambiguous. Without this guard a trailing open quote swallows every following
// separator (e.g. `go test "; rm -rf /` treats the `;` as quoted), re-opening the
// command-chain bypass M1 set out to close. Per the §M1 "err toward denying when
// ambiguous" invariant, treat an unterminated quote as a separator so the ":*"
// guard denies. The §F.4 carve-out covers *escaped* quotes (`\"`), NOT *unterminated*
// ones — those are in-scope for D1.
//
// @MX:NOTE: [AUTO] SPEC-SEC-HARDEN-001 M1 — quote-aware shell-separator scan guards the ":*" prefix branch against command-chain bypass; unterminated quote (D1) is treated as ambiguous → deny.
func hasUnquotedShellSeparator(s string) bool {
	var inSingle, inDouble bool
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case inSingle:
			if c == '\'' {
				inSingle = false
			}
		case inDouble:
			if c == '"' {
				inDouble = false
			}
		case c == '\'':
			inSingle = true
		case c == '"':
			inDouble = true
		case c == ';' || c == '|' || c == '&' || c == '`' || c == '\n' || c == '>' || c == '<':
			// Single-char separators (and the first char of `&&` / `||`):
			// any unquoted occurrence is a command-chain boundary.
			//
			// SECURITY (SPEC-SEC-HARDEN-002 M4): `>` / `<` are shell redirect
			// operators. Without them an allow-listed read/test command becomes
			// an arbitrary-file-write primitive (e.g. `go test > /etc/cron.d/payload`
			// resolves to ALLOW). The digit-prefixed form `2>` is also caught
			// because its `>` is unquoted. `&>` / `>&` / `2>&1` stay conservatively
			// denied via the `&` case above (no special-casing — errs safe).
			return true
		case c == '$' && i+1 < len(s) && s[i+1] == '(':
			// Command substitution `$(...)`.
			return true
		}
	}
	// D1 containment: an unterminated quote at end-of-string is ambiguous —
	// any separator inside it was silently swallowed, so deny by reporting true.
	return inSingle || inDouble
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

// preAllowlistOnce holds the global state for sync.Once-based caching.
// T-RT002-14: PreAllowlist hot-path optimization — builds the slice only on the first call.
var (
	preAllowlistOnce  sync.Once
	preAllowlistCache []PermissionRule
)

// PreAllowlist returns the built-in pre-allowlist for common development operations.
// These rules are at SrcBuiltin tier and cover ~80% of read/verify operations to
// reduce bubble fatigue. The pre-allowlist is always active and cannot be overridden
// by lower-tier rules.
//
// Uses sync.Once caching so the slice is built only on the first call (hot-path optimization).
//
// The pre-allowlist includes:
//   - Read operations: Read(*), Glob(*), Grep(*)
//   - Common test commands: go test, golangci-lint run, ruff check, npm test, pytest
//
// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-006
//
// @MX:ANCHOR: [AUTO] PreAllowlist is the SrcBuiltin-based rule source for the 8-tier resolver
// @MX:REASON: [AUTO] fan_in=3: called from resolver.go::checkTier, stack_test.go, spawn_test.go
func PreAllowlist() []PermissionRule {
	preAllowlistOnce.Do(func() {
		preAllowlistCache = []PermissionRule{
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
	})
	// Return a copy to prevent cache mutation.
	result := make([]PermissionRule, len(preAllowlistCache))
	copy(result, preAllowlistCache)
	return result
}

// IsWriteOperation returns true if the tool invocation is a write operation.
// Write operations include:
//   - Write tool
//   - Edit tool
//   - Bash with known write patterns (rm, mv, cp, git push, etc.)
//
// This is used to enforce ModePlan, which denies all write operations.
//
// T-RT002-20: enhanced write patterns — deduplication, additional patterns, refined matchers.
//
// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-020
func IsWriteOperation(tool, input string) bool {
	switch tool {
	case "Write", "Edit":
		return true
	case "Bash":
		trimmed := strings.TrimSpace(input)
		inputLower := strings.ToLower(trimmed)

		// Refined prefix-based matcher (prevents partial-substring false positives).
		prefixPatterns := []string{
			"rm ", "rmdir ",
			"mv ",        // mv (duplicate entry removed).
			"cp ", "copy ",
			"mkdir ", "mktemp ",
			"touch ",
			"git commit ", "git push ", "git pull ",
			"git merge ", "git rebase ",
			"git reset --hard", "git clean ",
			"npm install ", "pip install ", "go get ",
			"docker build", "docker push ",
			"make install",
		}
		for _, p := range prefixPatterns {
			if strings.HasPrefix(inputLower, p) {
				return true
			}
		}

		// Refined echo: only commands starting with "echo " (prevents false positives like echo variable references).
		if strings.HasPrefix(trimmed, "echo ") {
			return true
		}

		// Refined cat > redirect.
		if strings.HasPrefix(inputLower, "cat >") || strings.HasPrefix(inputLower, "tee ") {
			return true
		}

		// printf redirect pattern.
		if strings.HasPrefix(inputLower, "printf ") && strings.Contains(inputLower, ">") {
			return true
		}

		// dd of= pattern.
		if strings.HasPrefix(inputLower, "dd ") && strings.Contains(inputLower, "of=") {
			return true
		}
	}
	return false
}
