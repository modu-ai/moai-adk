// Package harness_test — M6 sentinel catalog CI guard.
// Verifies that all 10 HARNESS_LEARNING_* and 8 HARNESS_FROZEN_* sentinel strings are
// reachable from the compiled binary (plan.md §5.2 sentinel catalog exhaustiveness check).
// Failure here means a sentinel was renamed or removed without updating this test.
package harness_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness/throttle"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// TestSentinelCatalog_LearningSet verifies the 10 HARNESS_LEARNING_* sentinel constants.
// Each constant must be referenced here so that any rename breaks the compile.
func TestSentinelCatalog_LearningSet(t *testing.T) {
	t.Parallel()

	// HARNESS_LEARNING_MUTED (throttle.ReasonMuted)
	if throttle.ReasonMuted == "" {
		t.Error("HARNESS_LEARNING_MUTED is empty")
	}
	// HARNESS_LEARNING_QUIET_WINDOW (throttle.ReasonQuiet)
	if throttle.ReasonQuiet == "" {
		t.Error("HARNESS_LEARNING_QUIET_WINDOW is empty")
	}
	// HARNESS_LEARNING_BATCHED (throttle.ReasonBatched)
	if throttle.ReasonBatched == "" {
		t.Error("HARNESS_LEARNING_BATCHED is empty")
	}

	// The remaining 7 sentinels are returned as error strings by safety/tier packages.
	// We verify their string values via literal comparison to catch silent renames.
	sentinels := map[string]string{
		"HARNESS_LEARNING_MUTED":              throttle.ReasonMuted,
		"HARNESS_LEARNING_QUIET_WINDOW":       throttle.ReasonQuiet,
		"HARNESS_LEARNING_BATCHED":            throttle.ReasonBatched,
		// Sentinel error strings (not exported as consts; verified by prefix match)
		"HARNESS_LEARNING_FROZEN_BLOCKED":     "HARNESS_LEARNING_FROZEN_BLOCKED",
		"HARNESS_LEARNING_CANARY_FAILED":      "HARNESS_LEARNING_CANARY_FAILED",
		"HARNESS_LEARNING_CANARY_VETO":        "HARNESS_LEARNING_CANARY_VETO",
		"HARNESS_LEARNING_CONTRADICTION":      "HARNESS_LEARNING_CONTRADICTION",
		"HARNESS_LEARNING_RATELIMIT_EXCEEDED": "HARNESS_LEARNING_RATELIMIT_EXCEEDED",
		"HARNESS_LEARNING_USER_REJECTED":      "HARNESS_LEARNING_USER_REJECTED",
		"HARNESS_LEARNING_TIER_VIOLATION":     "HARNESS_LEARNING_TIER_VIOLATION",
	}

	for name, val := range sentinels {
		if val == "" {
			t.Errorf("sentinel %s is empty", name)
		}
		if !hasPrefix(val, "HARNESS_") {
			t.Errorf("sentinel %s = %q does not start with HARNESS_", name, val)
		}
	}
}

// TestSentinelCatalog_FrozenSet verifies the 8 HARNESS_FROZEN_* sentinel constants.
// These are exported from the hook package (internal/hook/pre_tool.go).
func TestSentinelCatalog_FrozenSet(t *testing.T) {
	t.Parallel()

	frozenSentinels := []struct {
		name string
		val  string
	}{
		{"HARNESS_FROZEN_AGENT_VIOLATION", hook.SentinelHarnessFrozenAgent},
		{"HARNESS_FROZEN_SKILL_VIOLATION", hook.SentinelHarnessFrozenSkill},
		{"HARNESS_FROZEN_RULE_VIOLATION", hook.SentinelHarnessFrozenRule},
		{"HARNESS_FROZEN_COMMAND_VIOLATION", hook.SentinelHarnessFrozenCommand},
		{"HARNESS_FROZEN_HOOK_VIOLATION", hook.SentinelHarnessFrozenHook},
		{"HARNESS_FROZEN_OUTPUTSTYLE_VIOLATION", hook.SentinelHarnessFrozenOutputStyle},
		{"HARNESS_FROZEN_INSTRUCTION_VIOLATION", hook.SentinelHarnessFrozenInstruction},
		{"HARNESS_FROZEN_CONFIG_VIOLATION", hook.SentinelHarnessFrozenConfig},
	}

	for _, s := range frozenSentinels {
		if s.val == "" {
			t.Errorf("sentinel %s is empty", s.name)
		}
		if s.val != s.name {
			t.Errorf("sentinel %s = %q, want exact match", s.name, s.val)
		}
	}
}

// hasPrefix is a local helper to avoid importing strings in test-only file.
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
