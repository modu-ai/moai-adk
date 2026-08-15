package cli

// glm_reasoning_overlay_test.go — SPEC-MODEL-TIER-PLANTYPE-001 M5 wiring tests
// (REQ-MTP-030, AC-MTP-032a/032b). Verifies the GLM effort overlay is WIRED into
// the GLM launch path (setGLMEnv via the shared glmReasoningEnvVars helper) and
// that the Branch-B reasoning key has inject↔clear parity in buildTmuxClearVars.

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestGLMReasoningEnvVars_SessionHigh asserts the shared overlay wire point injects
// the session-global reasoning-control value (reasoning-high) derived from the
// effort overlay (REQ-MTP-030 Branch-B).
func TestGLMReasoningEnvVars_SessionHigh(t *testing.T) {
	got := glmReasoningEnvVars()
	val, ok := got[config.EnvAnthropicReasoningEffort]
	if !ok {
		t.Fatalf("glmReasoningEnvVars() missing %q; got %v", config.EnvAnthropicReasoningEffort, got)
	}
	if val != "high" {
		t.Errorf("glmReasoningEnvVars()[%q] = %q, want %q (session default)",
			config.EnvAnthropicReasoningEffort, val, "high")
	}
}

// TestBuildTmuxClearVars_ReasoningParity asserts the Branch-B reasoning key is in
// the moai cc teardown clear list (inject↔clear parity, REQ-CGH-009 / AC-MTP-032b).
func TestBuildTmuxClearVars_ReasoningParity(t *testing.T) {
	clearVars := buildTmuxClearVars()
	found := false
	for _, v := range clearVars {
		if v == config.EnvAnthropicReasoningEffort {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("buildTmuxClearVars() must include %q for inject↔clear parity; got %v",
			config.EnvAnthropicReasoningEffort, clearVars)
	}
}
