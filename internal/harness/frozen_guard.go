// Package harness — frozen_guard.go implements the FROZEN guard that
// prevents writes to moai-managed directories during meta-harness invocation
// (SPEC-V3R3-PROJECT-HARNESS-001 REQ-PH-011, T-P2-02).
//
// Write flow: EnsureAllowed(path) → returns nil (allowed) or error (blocked).
package harness

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// allowedPrefixes is the exhaustive list of path prefixes that
// meta-harness output is permitted to write into. Any path starting
// with one of these prefixes (after slash-normalisation) is ALLOWED.
var allowedPrefixes = []string{
	".claude/agents/my-harness/",
	".claude/skills/my-harness-",
	".moai/harness/",
}

// frozenPrefixes is the exhaustive list of moai-managed path prefixes
// that are permanently FROZEN. Any write attempt to a path starting with
// one of these prefixes is rejected with a FrozenViolationError.
var frozenPrefixes = []string{
	".claude/agents/moai/",
	".claude/skills/moai-",
	".claude/skills/moai/",
	".claude/rules/moai/",
}

// FrozenViolationError is returned when IsAllowedPath or EnsureAllowed
// detects a write attempt against a moai-managed (FROZEN) path prefix.
type FrozenViolationError struct {
	// Path is the path that triggered the violation.
	Path string
	// Reason describes why the path is forbidden.
	Reason string
}

// Error implements the error interface.
// Format: "FROZEN_VIOLATION: <path>: <reason>"
func (e *FrozenViolationError) Error() string {
	return fmt.Sprintf("FROZEN_VIOLATION: %s: %s", e.Path, e.Reason)
}

// IsAllowedPath checks whether path is an explicitly allowed write target.
//
// Rules (evaluated in order):
//  1. Empty paths are rejected with an error.
//  2. Paths containing ".." path segments are rejected (traversal defence).
//  3. If path starts with a forbidden prefix → *FrozenViolationError.
//  4. If path starts with an allowed prefix → (true, nil).
//  5. Otherwise → (false, nil) — neutral, caller decides.
//
// All comparisons use forward-slash-normalised paths so the guard works
// consistently across operating systems.
func IsAllowedPath(path string) (bool, error) {
	if path == "" {
		return false, errors.New("harness: IsAllowedPath: path must not be empty")
	}

	// Normalise to forward slashes for consistent prefix matching.
	norm := filepath.ToSlash(path)

	// Reject path-traversal attempts.
	if strings.Contains(norm, "..") {
		return false, fmt.Errorf("harness: IsAllowedPath: path traversal not allowed: %q", path)
	}

	// Forbidden check first (FROZEN guard takes priority).
	for _, prefix := range frozenPrefixes {
		if strings.HasPrefix(norm, prefix) {
			return false, &FrozenViolationError{
				Path:   path,
				Reason: fmt.Sprintf("path is in moai-managed FROZEN area (prefix %q)", prefix),
			}
		}
	}

	// Allowed check.
	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(norm, prefix) {
			return true, nil
		}
	}

	// Neutral — neither allowed nor forbidden.
	return false, nil
}

// @MX:ANCHOR: [AUTO] EnsureAllowed — FROZEN gate that must be called before meta-harness writes
// @MX:REASON: First checkpoint of all Phase 6 write paths (REQ-PH-011). Expected fan_in >= 3.
// @MX:SPEC: SPEC-V3R3-PROJECT-HARNESS-001 T-P2-02

// EnsureAllowed returns nil only if path is an explicitly allowed write
// target. It returns an error in all other cases:
//   - empty or traversal path → validation error
//   - forbidden prefix → *FrozenViolationError
//   - neutral (not in allowed list) → generic error (guard blocks by default)
func EnsureAllowed(path string) error {
	ok, err := IsAllowedPath(path)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("harness: EnsureAllowed: path %q is not in the allowed write-prefix list", path)
	}
	return nil
}
