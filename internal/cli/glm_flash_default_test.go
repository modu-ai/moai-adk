package cli

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestGLMFlashDefaultEnvInjection is the boot smoke for the flash default
// switch: with NO llm.yaml on disk (the default configuration), a `moai glm`
// launch resolves glm-5.3-flash into every ANTHROPIC_DEFAULT_*_MODEL slot and
// sizes the statusline/auto-compact windows at 1,000,000. The observation is
// env-level ONLY — the injection maps the launcher builds — with NO live z.ai
// API dependency, no interactive prompt, and no API round-trip.
//
// Two legs, mirroring the two injection surfaces:
//
//	leg 1  buildTmuxInjectVars — the mutation-free map (tmux session env);
//	leg 2  setGLMEnv — the process-env path (the tmux-absent fallback).
//
// Env isolation: t.Setenv registers restore for every key setGLMEnv mutates —
// t.TempDir scopes config FILES, not environment variables, so the process-env
// leg needs the explicit registration (audit note R1). Not parallel:
// t.Setenv/t.Chdir are incompatible with t.Parallel.
func TestGLMFlashDefaultEnvInjection(t *testing.T) {
	// Clean cwd + absent llm.yaml root: the loader's default slots are the
	// configuration under test; no project-level override may leak in.
	t.Chdir(t.TempDir())

	glmConfig, err := loadGLMConfig(t.TempDir())
	if err != nil {
		t.Fatalf("loadGLMConfig on an absent llm.yaml should fall to defaults, got: %v", err)
	}

	want := config.DefaultGLM53Flash
	slotEnvs := []string{
		config.EnvAnthropicDefaultOpusModel,
		config.EnvAnthropicDefaultSonnetModel,
		config.EnvAnthropicDefaultHaikuModel,
		config.EnvAnthropicDefaultFableModel,
	}
	slots := map[string]string{
		config.EnvAnthropicDefaultOpusModel:   glmConfig.Models.High,
		config.EnvAnthropicDefaultSonnetModel: glmConfig.Models.Medium,
		config.EnvAnthropicDefaultHaikuModel:  glmConfig.Models.Low,
		config.EnvAnthropicDefaultFableModel:  glmConfig.Models.Fable,
	}
	for _, envKey := range slotEnvs {
		if slots[envKey] != want {
			t.Errorf("default slot feeding %s = %q, want %q", envKey, slots[envKey], want)
		}
	}

	// Leg 1 — the mutation-free tmux injection map.
	injectVars := buildTmuxInjectVars(glmConfig, "some-token")
	for _, envKey := range slotEnvs {
		if got := injectVars[envKey]; got != want {
			t.Errorf("buildTmuxInjectVars[%s] = %q, want %q (tmux session env)", envKey, got, want)
		}
	}
	if got := injectVars[config.EnvStatuslineContextSize]; got != "1000000" {
		t.Errorf("buildTmuxInjectVars[%s] = %q, want %q (1M statusline window)", config.EnvStatuslineContextSize, got, "1000000")
	}
	if got := injectVars[config.EnvClaudeCodeAutoCompactWindow]; got != "1000000" {
		t.Errorf("buildTmuxInjectVars[%s] = %q, want %q (1M auto-compact window)", config.EnvClaudeCodeAutoCompactWindow, got, "1000000")
	}

	// Leg 2 — the process-env path (setGLMEnv). Register restore for every
	// key it mutates so the test leaves the process environment unchanged.
	for _, k := range []string{
		config.EnvAnthropicAuthToken,
		config.EnvAnthropicBaseURL,
		config.EnvAnthropicDefaultOpusModel,
		config.EnvAnthropicDefaultSonnetModel,
		config.EnvAnthropicDefaultHaikuModel,
		config.EnvAnthropicDefaultFableModel,
		config.EnvClaudeCodeAutoCompactWindow,
		config.EnvClaudeCodeMaxContextTokens,
		config.EnvClaudeCodeDisableExperimentalBetas,
		config.EnvAnthropicReasoningEffort,
		"API_TIMEOUT_MS",
		"Z_AI_API_KEY",
	} {
		t.Setenv(k, "")
	}
	setGLMEnv(glmConfig, "some-token")
	for _, envKey := range slotEnvs {
		if got := os.Getenv(envKey); got != want {
			t.Errorf("process env %s = %q after setGLMEnv, want %q", envKey, got, want)
		}
	}
	if got := os.Getenv(config.EnvClaudeCodeAutoCompactWindow); got != "1000000" {
		t.Errorf("process env %s = %q after setGLMEnv, want %q", config.EnvClaudeCodeAutoCompactWindow, got, "1000000")
	}
}
