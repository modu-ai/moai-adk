// Package safety — 5-Layer Safety Architecture.
// Layer 1: Frozen Guard (REQ-HL-006).
// Blocks automatic learning updates for FROZEN paths and records violations.
package safety

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// frozenPrefixes is the list of path prefixes where automatic learning updates can never occur.
// REQ-HL-006: This list is hardcoded and cannot be changed by any config/env.
//
// [HARD] This constant cannot be overridden by configuration files or environment variables.
var frozenPrefixes = []string{
	".claude/agents/moai/",
	".claude/skills/moai-",
	".claude/rules/moai/",
	".moai/project/brand/",
}

// IsFrozen returns true if path corresponds to a FROZEN area.
// REQ-HL-006: Normalizes with filepath.Clean, then compares against hardcoded frozenPrefixes.
//
// Characteristics:
//   - Empty path returns false (nothing to block)
//   - Backslashes are converted to slashes for OS-independent handling
//   - Evaluates after removing ".." etc. with filepath.Clean
//   - Absolute paths do not match frozenPrefixes (relative path based)
//
// [HARD] This function's behavior cannot be changed by config/env.
func IsFrozen(path string) bool {
	if path == "" {
		return false
	}

	// Backslash to slash conversion (Windows compatible)
	norm := filepath.ToSlash(path)

	// Normalize "..", "." etc. with filepath.Clean
	norm = filepath.ToSlash(filepath.Clean(norm))

	// Compare against hardcoded prefix list
	for _, prefix := range frozenPrefixes {
		if hasPrefix(norm, prefix) {
			return true
		}
	}

	return false
}

// hasPrefix checks whether s starts with prefix.
// After filepath.Clean, slashes are already unified, so simple comparison is used.
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// violationEntry is the single-line schema for frozen-guard-violations.jsonl.
type violationEntry struct {
	// Timestamp is the violation occurrence time (UTC RFC3339).
	Timestamp time.Time `json:"timestamp"`

	// Path is the target path where the violation was detected.
	Path string `json:"path"`

	// Caller is the identifier of the caller that caused the violation.
	Caller string `json:"caller"`

	// Message is the detailed violation message.
	Message string `json:"message"`
}

// LogViolation records FROZEN path access violations to logPath in JSONL format
// and outputs warning messages to stderr.
// REQ-HL-006: Violations are non-blocking — does not stop process even if recording fails.
//
// @MX:ANCHOR: [AUTO] LogViolation is the single entry point for FROZEN violation recording.
// @MX:REASON: [AUTO] fan_in >= 3: frozen_guard_test.go, pipeline.go, Phase 4 coordinator
func LogViolation(logPath, path, caller string) error {
	entry := violationEntry{
		Timestamp: time.Now().UTC(),
		Path:      path,
		Caller:    caller,
		Message:   fmt.Sprintf("FROZEN_VIOLATION: automatic update attempt blocked for path %s", path),
	}

	// Create parent directory
	dir := filepath.Dir(logPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("safety/frozen_guard: failed to create directory %s: %w", dir, err)
		}
	}

	// JSONL serialization
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("safety/frozen_guard: failed to serialize violation: %w", err)
	}
	data = append(data, '\n')

	// Record in append mode
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("safety/frozen_guard: failed to open violation log file %s: %w", logPath, err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("safety/frozen_guard: failed to write violation log: %w", err)
	}

	// Output stderr warning (non-blocking, for learning system observers)
	fmt.Fprintf(os.Stderr, "[WARN] safety/frozen_guard: FROZEN path access blocked: %s (caller: %s)\n", path, caller)

	return nil
}
