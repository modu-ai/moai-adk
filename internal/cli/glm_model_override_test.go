package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestGLMCmd_AddsModelOverrides verifies that 'moai glm' injects GLM model overrides
// into the current process env (via setGLMEnv). After the #676 fix, these are no longer
// written to settings.local.json — the process env is what syscall.Exec inherits into
// `claude`. settings.local.json must remain clean to prevent GLM mode from leaking
// into subsequent sessions.
// NOTE: does not call t.Parallel() because it modifies process-level env via setGLMEnv.
func TestGLMCmd_AddsModelOverrides(t *testing.T) {
	// Set GLM_API_KEY env var
	t.Setenv("GLM_API_KEY", "test-api-key-for-model-override-test")
	// Baseline: clear the vars we will check so the test is deterministic.
	t.Setenv("ANTHROPIC_DEFAULT_OPUS_MODEL", "")
	t.Setenv("ANTHROPIC_DEFAULT_SONNET_MODEL", "")
	t.Setenv("ANTHROPIC_DEFAULT_HAIKU_MODEL", "")

	// Create temp project
	tmpDir := t.TempDir()
	moaiDir := filepath.Join(tmpDir, ".moai")
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return tmpDir, nil }
	defer func() { findProjectRootFn = origFn }()

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Override launchClaude to skip actual exec
	origLaunch := launchClaudeFunc
	defer func() { launchClaudeFunc = origLaunch }()
	launchClaudeFunc = func(profile string, args []string) error {
		return nil
	}

	// Run 'moai glm'
	buf := new(bytes.Buffer)
	glmCmd.SetOut(buf)
	glmCmd.SetErr(buf)

	err := glmCmd.RunE(glmCmd, []string{})
	if err != nil {
		t.Fatalf("moai glm should not error, got: %v", err)
	}

	// GLM model overrides must be in the PROCESS ENV (inherited by syscall.Exec),
	// NOT in settings.local.json (which would cause persistence pollution, #676).
	defaults := config.NewDefaultLLMConfig()
	if got := os.Getenv("ANTHROPIC_DEFAULT_OPUS_MODEL"); got == "" {
		t.Error("ANTHROPIC_DEFAULT_OPUS_MODEL must be set in process env after moai glm")
	} else if got != defaults.GLM.Models.High {
		t.Errorf("ANTHROPIC_DEFAULT_OPUS_MODEL = %q, want %q", got, defaults.GLM.Models.High)
	}
	if got := os.Getenv("ANTHROPIC_DEFAULT_SONNET_MODEL"); got == "" {
		t.Error("ANTHROPIC_DEFAULT_SONNET_MODEL must be set in process env after moai glm")
	}
	if got := os.Getenv("ANTHROPIC_DEFAULT_HAIKU_MODEL"); got == "" {
		t.Error("ANTHROPIC_DEFAULT_HAIKU_MODEL must be set in process env after moai glm")
	}
	if got := os.Getenv("ANTHROPIC_AUTH_TOKEN"); got == "" {
		t.Error("ANTHROPIC_AUTH_TOKEN must be set in process env after moai glm")
	}
	if got := os.Getenv("ANTHROPIC_BASE_URL"); got == "" {
		t.Error("ANTHROPIC_BASE_URL must be set in process env after moai glm")
	}

	// settings.local.json must NOT contain GLM env vars (#676 regression guard).
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	if data, err := os.ReadFile(settingsPath); err == nil {
		content := string(data)
		if strings.Contains(content, "ANTHROPIC_BASE_URL") {
			t.Error("settings.local.json must NOT contain ANTHROPIC_BASE_URL (regression: #676)")
		}
		if strings.Contains(content, "ANTHROPIC_DEFAULT_OPUS_MODEL") {
			t.Error("settings.local.json must NOT contain ANTHROPIC_DEFAULT_OPUS_MODEL (regression: #676)")
		}
	}
}

// TestCGCmd_DoesNotAddModelOverridesToSettings verifies that 'moai cg' does NOT add
// GLM model overrides to settings.local.json. CG mode uses tmux session-level env vars
// instead. This is the intended behavior.
func TestCGCmd_DoesNotAddModelOverridesToSettings(t *testing.T) {
	// Set GLM_API_KEY env var
	t.Setenv("GLM_API_KEY", "test-api-key-for-cg-test")
	t.Setenv("MOAI_TEST_MODE", "1") // Skip tmux requirement

	// Create temp project
	tmpDir := t.TempDir()
	moaiDir := filepath.Join(tmpDir, ".moai")
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return tmpDir, nil }
	defer func() { findProjectRootFn = origFn }()

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Override launchClaude to skip actual exec
	origLaunch := launchClaudeFunc
	defer func() { launchClaudeFunc = origLaunch }()
	launchClaudeFunc = func(profile string, args []string) error {
		return nil
	}

	// Run 'moai cg'
	buf := new(bytes.Buffer)
	cgCmd.SetOut(buf)
	cgCmd.SetErr(buf)

	err := cgCmd.RunE(cgCmd, []string{})
	if err != nil {
		t.Fatalf("moai cg should not error in test mode, got: %v", err)
	}

	// Verify settings.local.json was created
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("settings.local.json should be created: %v", err)
	}

	content := string(data)

	// CG mode should NOT have GLM model overrides in settings.local.json
	// because they're set at tmux session level instead
	if strings.Contains(content, "ANTHROPIC_DEFAULT_OPUS_MODEL") {
		t.Error("settings.local.json should NOT contain ANTHROPIC_DEFAULT_OPUS_MODEL in CG mode (uses tmux env)")
	}
	if strings.Contains(content, "ANTHROPIC_DEFAULT_SONNET_MODEL") {
		t.Error("settings.local.json should NOT contain ANTHROPIC_DEFAULT_SONNET_MODEL in CG mode (uses tmux env)")
	}
	if strings.Contains(content, "ANTHROPIC_DEFAULT_HAIKU_MODEL") {
		t.Error("settings.local.json should NOT contain ANTHROPIC_DEFAULT_HAIKU_MODEL in CG mode (uses tmux env)")
	}

	// CG mode should have teammateMode set to tmux (native key)
	if !strings.Contains(content, "\"teammateMode\"") {
		t.Error("settings.local.json should contain teammateMode in CG mode")
	}
	if !strings.Contains(content, "\"tmux\"") {
		t.Error("teammateMode should be set to \"tmux\" in CG mode")
	}

	// CG mode should NOT have GLM auth vars in settings.local.json
	if strings.Contains(content, "ANTHROPIC_AUTH_TOKEN") {
		t.Error("settings.local.json should NOT contain ANTHROPIC_AUTH_TOKEN in CG mode (uses tmux env)")
	}
	if strings.Contains(content, "ANTHROPIC_BASE_URL") {
		t.Error("settings.local.json should NOT contain ANTHROPIC_BASE_URL in CG mode (uses tmux env)")
	}
}

// TestSaveLLMSection_PopulatesDefaultGLMModels verifies that saveLLMSection
// populates empty GLM model values with explicit defaults, preventing
// confusion in llm.yaml.
func TestSaveLLMSection_PopulatesDefaultGLMModels(t *testing.T) {
	tmpDir := t.TempDir()
	sectionsDir := filepath.Join(tmpDir, "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create an LLM config with empty model values (as would happen when
	// persistTeamMode creates a new config)
	llmCfg := config.LLMConfig{
		TeamMode: "glm",
		// GLM models are intentionally left empty
		GLM: config.GLMSettings{
			BaseURL: "",
			Models: config.GLMModels{
				High:   "",
				Medium: "",
				Low:    "",
			},
		},
		GLMEnvVar: "",
	}

	// Save the config
	err := saveLLMSection(sectionsDir, llmCfg)
	if err != nil {
		t.Fatalf("saveLLMSection should succeed: %v", err)
	}

	// Read the saved llm.yaml
	llmPath := filepath.Join(sectionsDir, "llm.yaml")
	data, err := os.ReadFile(llmPath)
	if err != nil {
		t.Fatalf("llm.yaml should exist: %v", err)
	}

	content := string(data)

	// Verify that default values were populated, not empty strings
	if !strings.Contains(content, "glm-5.1") {
		t.Errorf("llm.yaml should contain glm-5.1 as default high model, got:\n%s", content)
	}
	if !strings.Contains(content, "glm-4.7") {
		t.Errorf("llm.yaml should contain glm-4.7 as default medium model, got:\n%s", content)
	}
	if !strings.Contains(content, "glm-4.5-air") {
		t.Errorf("llm.yaml should contain glm-4.5-air as default low model, got:\n%s", content)
	}

	// Verify base URL is populated
	if !strings.Contains(content, "https://api.z.ai/api/anthropic") {
		t.Errorf("llm.yaml should contain default GLM base URL, got:\n%s", content)
	}

	// Verify GLM env var is populated
	if !strings.Contains(content, "GLM_API_KEY") {
		t.Errorf("llm.yaml should contain GLM_API_KEY env var, got:\n%s", content)
	}

	// Verify team_mode is preserved
	if !strings.Contains(content, "team_mode: glm") {
		t.Errorf("llm.yaml should contain team_mode: glm, got:\n%s", content)
	}
}

// TestSaveLLMSection_PreservesCustomGLMModels verifies that saveLLMSection
// preserves custom GLM model values when they are already set.
func TestSaveLLMSection_PreservesCustomGLMModels(t *testing.T) {
	tmpDir := t.TempDir()
	sectionsDir := filepath.Join(tmpDir, "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create an LLM config with custom model values
	llmCfg := config.LLMConfig{
		TeamMode: "glm",
		GLM: config.GLMSettings{
			BaseURL: "https://custom-glm.example.com",
			Models: config.GLMModels{
				High:   "custom-glm-opus",
				Medium: "custom-glm-sonnet",
				Low:    "custom-glm-haiku",
			},
		},
		GLMEnvVar: "CUSTOM_GLM_KEY",
	}

	// Save the config
	err := saveLLMSection(sectionsDir, llmCfg)
	if err != nil {
		t.Fatalf("saveLLMSection should succeed: %v", err)
	}

	// Read the saved llm.yaml
	llmPath := filepath.Join(sectionsDir, "llm.yaml")
	data, err := os.ReadFile(llmPath)
	if err != nil {
		t.Fatalf("llm.yaml should exist: %v", err)
	}

	content := string(data)

	// Verify that custom values are preserved, not replaced with defaults
	if !strings.Contains(content, "custom-glm-opus") {
		t.Errorf("llm.yaml should preserve custom high model, got:\n%s", content)
	}
	if !strings.Contains(content, "custom-glm-sonnet") {
		t.Errorf("llm.yaml should preserve custom medium model, got:\n%s", content)
	}
	if !strings.Contains(content, "custom-glm-haiku") {
		t.Errorf("llm.yaml should preserve custom low model, got:\n%s", content)
	}
	if !strings.Contains(content, "https://custom-glm.example.com") {
		t.Errorf("llm.yaml should preserve custom base URL, got:\n%s", content)
	}
	if !strings.Contains(content, "CUSTOM_GLM_KEY") {
		t.Errorf("llm.yaml should preserve custom env var, got:\n%s", content)
	}

	// Verify defaults are NOT present when custom values are used
	// Note: Legacy fields (opus, sonnet, haiku) are separate from primary fields
	// (high, medium, low) and are populated with defaults unless explicitly set
	if strings.Contains(content, "opus: glm-5.1") && strings.Contains(content, "high: custom-glm-opus") {
		// This is acceptable - high field has custom value, opus field has default
		// User can override opus separately if needed
	} else if strings.Contains(content, "high: glm-5.1") {
		t.Errorf("llm.yaml should NOT contain default glm-5.1 in high field when custom model is set, got:\n%s", content)
	}
	if strings.Contains(content, "https://api.z.ai/api/anthropic") {
		t.Errorf("llm.yaml should NOT contain default base URL when custom URL is set, got:\n%s", content)
	}
}
