package config

// defaults_clifix_test.go — SPEC-CLIFIX-HYGIENE-001 REQ-HYG-001-004 (M3)
//
// Asserts the single-source threshold/timeout constants exist with the
// expected baseline values and that the three former tier-threshold sites
// plus the two former dispatcher-timeout sites have a single derivation point
// in this package. These are compile-time presence checks; the behavior-preservation
// of the call-site rewrites is covered by the existing characterization net
// (M1) and the cli-level parity test (internal/cli/glm_env_parity_test.go is
// the GLM-env analogue; tier thresholds are covered by internal/harness
// classifier tests).

import (
	"testing"
	"time"
)

// TestDefaultTierThresholdsCanonical asserts the collapsed SSOT carries the
// pre-M3 baseline vector exactly. The three former inline sites
// (internal/cli/harness.go runHarnessStatus fallback, harness.go
// defaultLearningConfig struct initializer, internal/cli/hook.go
// defaultTierThresholds) all referenced `[]int{1, 3, 5, 10}`; behavior is
// unchanged iff this constant still equals that vector.
func TestDefaultTierThresholdsCanonical(t *testing.T) {
	t.Parallel()

	want := []int{1, 3, 5, 10}
	got := DefaultTierThresholds
	if len(got) != len(want) {
		t.Fatalf("DefaultTierThresholds len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i, g := range got {
		if g != want[i] {
			t.Errorf("DefaultTierThresholds[%d] = %d, want %d", i, g, want[i])
		}
	}
}

// TestDefaultHookDispatcherTimeoutCanonical asserts the dispatcher timeout
// carries the pre-M3 baseline (30s). The two former inline sites
// (internal/cli/hook.go:237 and :361) both used `30 * time.Second`.
func TestDefaultHookDispatcherTimeoutCanonical(t *testing.T) {
	t.Parallel()

	if want := 30 * time.Second; DefaultHookDispatcherTimeout != want {
		t.Errorf("DefaultHookDispatcherTimeout = %v, want %v", DefaultHookDispatcherTimeout, want)
	}
}

// TestGLMEnvConstantsCanonical asserts the three GLM env-var NAME constants
// carry the pre-M3 baseline string values, so the behavior of every inject and
// clear site is preserved.
func TestGLMEnvConstantsCanonical(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		got  string
		want string
	}{
		{"EnvClaudeCodeDisableExperimentalBetas", EnvClaudeCodeDisableExperimentalBetas, "CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS"},
		{"EnvClaudeCodeDisableNonessentialTraffic", EnvClaudeCodeDisableNonessentialTraffic, "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC"},
		{"EnvClaudeCodeTeammateDisplay", EnvClaudeCodeTeammateDisplay, "CLAUDE_CODE_TEAMMATE_DISPLAY"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
}
