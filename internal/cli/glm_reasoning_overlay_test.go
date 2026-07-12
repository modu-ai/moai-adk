package cli

// glm_reasoning_overlay_test.go — SPEC-MODEL-TIER-PLANTYPE-001 M5 wiring tests
// (REQ-MTP-030, AC-MTP-032a/032b). Verifies the GLM effort overlay is WIRED into
// the GLM launch path (setGLMEnv / injectGLMEnvForTeam via the shared
// glmReasoningEnvVars helper) and that the Branch-B reasoning key has inject↔clear
// parity in buildTmuxClearVars.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestGLMReasoningEnvVars_SessionMax asserts the shared overlay wire point injects
// the session-global reasoning-control value (reasoning-max) derived from the
// effort overlay (REQ-MTP-030 Branch-B).
func TestGLMReasoningEnvVars_SessionMax(t *testing.T) {
	got := glmReasoningEnvVars()
	val, ok := got[config.EnvAnthropicReasoningEffort]
	if !ok {
		t.Fatalf("glmReasoningEnvVars() missing %q; got %v", config.EnvAnthropicReasoningEffort, got)
	}
	if val != "max" {
		t.Errorf("glmReasoningEnvVars()[%q] = %q, want %q (coding-max session default)",
			config.EnvAnthropicReasoningEffort, val, "max")
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

// TestInjectGLMEnvForTeam_WritesReasoningEnv asserts injectGLMEnvForTeam wires the
// reasoning-control env into settings.local.json (the team GLM launch path,
// AC-MTP-032a).
func TestInjectGLMEnvForTeam_WritesReasoningEnv(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(claudeDir, "settings.local.json")

	glmCfg := &GLMConfigFromYAML{BaseURL: "https://api.z.ai/api/anthropic"}
	glmCfg.Models.High = "glm-5.2"
	glmCfg.Models.Medium = "glm-4.7"
	glmCfg.Models.Low = "glm-4.5-air"
	glmCfg.Models.Fable = "glm-4.7"

	if err := injectGLMEnvForTeam(settingsPath, glmCfg, "test-key"); err != nil {
		t.Fatalf("injectGLMEnvForTeam: %v", err)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings.local.json: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, config.EnvAnthropicReasoningEffort) {
		t.Errorf("settings.local.json must contain %q (overlay wired at the team launch path); got:\n%s",
			config.EnvAnthropicReasoningEffort, content)
	}
}
