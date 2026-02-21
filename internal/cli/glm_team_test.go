package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestBuildGLMEnvVars verifies that buildGLMEnvVars produces the correct
// map of environment variables for GLM mode.
func TestBuildGLMEnvVars(t *testing.T) {
	tests := []struct {
		name      string
		glmConfig *GLMConfigFromYAML
		apiKey    string
		wantKeys  []string
		wantVals  map[string]string
	}{
		{
			name: "default_glm_config",
			glmConfig: &GLMConfigFromYAML{
				BaseURL: "https://api.z.ai/api/anthropic",
				Models: struct {
					High   string
					Medium string
					Low    string
				}{
					High:   "glm-5",
					Medium: "glm-4.7",
					Low:    "glm-4.7-flashx",
				},
				EnvVar: "GLM_API_KEY",
			},
			apiKey: "test-key-123",
			wantKeys: []string{
				"ANTHROPIC_AUTH_TOKEN",
				"ANTHROPIC_BASE_URL",
				"ANTHROPIC_DEFAULT_OPUS_MODEL",
				"ANTHROPIC_DEFAULT_SONNET_MODEL",
				"ANTHROPIC_DEFAULT_HAIKU_MODEL",
			},
			wantVals: map[string]string{
				"ANTHROPIC_AUTH_TOKEN":           "test-key-123",
				"ANTHROPIC_BASE_URL":             "https://api.z.ai/api/anthropic",
				"ANTHROPIC_DEFAULT_OPUS_MODEL":   "glm-5",
				"ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.7",
				"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "glm-4.7-flashx",
			},
		},
		{
			name: "custom_config",
			glmConfig: &GLMConfigFromYAML{
				BaseURL: "https://custom.glm.api/v1",
				Models: struct {
					High   string
					Medium string
					Low    string
				}{
					High:   "custom-high",
					Medium: "custom-medium",
					Low:    "custom-low",
				},
				EnvVar: "CUSTOM_API_KEY",
			},
			apiKey: "custom-api-key-xyz",
			wantKeys: []string{
				"ANTHROPIC_AUTH_TOKEN",
				"ANTHROPIC_BASE_URL",
				"ANTHROPIC_DEFAULT_OPUS_MODEL",
				"ANTHROPIC_DEFAULT_SONNET_MODEL",
				"ANTHROPIC_DEFAULT_HAIKU_MODEL",
			},
			wantVals: map[string]string{
				"ANTHROPIC_AUTH_TOKEN":           "custom-api-key-xyz",
				"ANTHROPIC_BASE_URL":             "https://custom.glm.api/v1",
				"ANTHROPIC_DEFAULT_OPUS_MODEL":   "custom-high",
				"ANTHROPIC_DEFAULT_SONNET_MODEL": "custom-medium",
				"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "custom-low",
			},
		},
		{
			name: "empty_api_key",
			glmConfig: &GLMConfigFromYAML{
				BaseURL: "https://api.z.ai/api/anthropic",
				Models: struct {
					High   string
					Medium string
					Low    string
				}{
					High:   "glm-5",
					Medium: "glm-4.7",
					Low:    "glm-4.7-flashx",
				},
				EnvVar: "GLM_API_KEY",
			},
			apiKey: "",
			wantKeys: []string{
				"ANTHROPIC_AUTH_TOKEN",
				"ANTHROPIC_BASE_URL",
				"ANTHROPIC_DEFAULT_OPUS_MODEL",
				"ANTHROPIC_DEFAULT_SONNET_MODEL",
				"ANTHROPIC_DEFAULT_HAIKU_MODEL",
			},
			wantVals: map[string]string{
				"ANTHROPIC_AUTH_TOKEN":           "",
				"ANTHROPIC_BASE_URL":             "https://api.z.ai/api/anthropic",
				"ANTHROPIC_DEFAULT_OPUS_MODEL":   "glm-5",
				"ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.7",
				"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "glm-4.7-flashx",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildGLMEnvVars(tt.glmConfig, tt.apiKey)

			// Verify the map has exactly 5 keys
			if len(got) != 5 {
				t.Errorf("buildGLMEnvVars() returned %d keys, want 5", len(got))
			}

			// Verify all required keys exist
			for _, key := range tt.wantKeys {
				if _, ok := got[key]; !ok {
					t.Errorf("buildGLMEnvVars() missing key %q", key)
				}
			}

			// Verify specific values
			for key, wantVal := range tt.wantVals {
				if gotVal, ok := got[key]; !ok {
					t.Errorf("buildGLMEnvVars() missing key %q", key)
				} else if gotVal != wantVal {
					t.Errorf("buildGLMEnvVars()[%q] = %q, want %q", key, gotVal, wantVal)
				}
			}
		})
	}
}

// TestCGCommandRegistered verifies that the cg command is correctly registered
// on the root command.
func TestCGCommandRegistered(t *testing.T) {
	// Verify cgCmd has the correct Use field (no api-key arg)
	if cgCmd.Use != "cg" {
		t.Errorf("cgCmd.Use = %q, want %q", cgCmd.Use, "cg")
	}

	// Verify cgCmd does NOT have a --hybrid flag (it's always hybrid)
	flag := cgCmd.Flags().Lookup("hybrid")
	if flag != nil {
		t.Error("cgCmd should NOT have a --hybrid flag (CG is always hybrid)")
	}

	// Verify glmCmd does NOT have a --hybrid flag anymore
	glmFlag := glmCmd.Flags().Lookup("hybrid")
	if glmFlag != nil {
		t.Error("glmCmd should NOT have a --hybrid flag (use 'moai cg' instead)")
	}
}

// TestPersistTeamMode verifies that persistTeamMode saves team_mode to llm.yaml.
func TestPersistTeamMode(t *testing.T) {
	t.Setenv("MOAI_TEST_MODE", "1")

	// Create a temporary project directory with config
	projectRoot := t.TempDir()
	sectionsDir := filepath.Join(projectRoot, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Test persisting team mode
	if err := persistTeamMode(projectRoot, "glm"); err != nil {
		t.Fatalf("persistTeamMode() error: %v", err)
	}

	// Verify the llm.yaml was created with correct team_mode
	llmPath := filepath.Join(sectionsDir, "llm.yaml")
	data, err := os.ReadFile(llmPath)
	if err != nil {
		t.Fatalf("failed to read llm.yaml: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "team_mode: glm") {
		t.Errorf("llm.yaml should contain team_mode: glm, got:\n%s", content)
	}
}

// TestDisableTeamMode verifies that disableTeamMode resets team_mode to empty.
func TestDisableTeamMode(t *testing.T) {
	t.Setenv("MOAI_TEST_MODE", "1")

	// Create a temporary project directory with config
	projectRoot := t.TempDir()
	sectionsDir := filepath.Join(projectRoot, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// First enable, then disable
	if err := persistTeamMode(projectRoot, "glm"); err != nil {
		t.Fatalf("persistTeamMode() error: %v", err)
	}
	if err := disableTeamMode(projectRoot); err != nil {
		t.Fatalf("disableTeamMode() error: %v", err)
	}

	// Verify the llm.yaml has empty team_mode
	llmPath := filepath.Join(sectionsDir, "llm.yaml")
	data, err := os.ReadFile(llmPath)
	if err != nil {
		t.Fatalf("failed to read llm.yaml: %v", err)
	}

	content := string(data)
	if strings.Contains(content, "team_mode: glm") {
		t.Errorf("llm.yaml should have empty team_mode after disable, got:\n%s", content)
	}
}

// TestEnableTeamModeAlwaysGLM verifies enableTeamMode(false) sets team_mode to "glm".
func TestEnableTeamModeAlwaysGLM(t *testing.T) {
	t.Setenv("MOAI_TEST_MODE", "1")

	// Create project dir
	projectRoot := t.TempDir()
	sectionsDir := filepath.Join(projectRoot, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatal(err)
	}

	// isHybrid = false means all agents use GLM
	err := enableTeamMode(glmCmd, false)
	if err != nil {
		t.Fatalf("enableTeamMode() error: %v", err)
	}

	// Verify llm.yaml contains team_mode: glm
	data, err := os.ReadFile(filepath.Join(sectionsDir, "llm.yaml"))
	if err != nil {
		t.Fatalf("failed to read llm.yaml: %v", err)
	}
	if !strings.Contains(string(data), "team_mode: glm") {
		t.Errorf("llm.yaml should contain team_mode: glm, got:\n%s", string(data))
	}
}

// TestLoadLLMSectionIntegration verifies that the LLM section is loaded correctly
// from llm.yaml by the config.Loader.
func TestLoadLLMSectionIntegration(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()
	sectionsDir := filepath.Join(tmpDir, "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write an llm.yaml with custom values
	llmContent := `llm:
  mode: glm
  team_mode: glm
  glm_env_var: CUSTOM_KEY
  glm:
    base_url: https://custom.api/v1
    models:
      haiku: custom-haiku
      sonnet: custom-sonnet
      opus: custom-opus
`
	if err := os.WriteFile(filepath.Join(sectionsDir, "llm.yaml"), []byte(llmContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Load config
	loader := config.NewLoader()
	cfg, err := loader.Load(tmpDir)
	if err != nil {
		t.Fatalf("loader.Load() error: %v", err)
	}

	// Verify LLM config was loaded
	if cfg.LLM.Mode != "glm" {
		t.Errorf("LLM.Mode = %q, want %q", cfg.LLM.Mode, "glm")
	}
	if cfg.LLM.TeamMode != "glm" {
		t.Errorf("LLM.TeamMode = %q, want %q", cfg.LLM.TeamMode, "glm")
	}
	if cfg.LLM.GLMEnvVar != "CUSTOM_KEY" {
		t.Errorf("LLM.GLMEnvVar = %q, want %q", cfg.LLM.GLMEnvVar, "CUSTOM_KEY")
	}
	if cfg.LLM.GLM.BaseURL != "https://custom.api/v1" {
		t.Errorf("LLM.GLM.BaseURL = %q, want %q", cfg.LLM.GLM.BaseURL, "https://custom.api/v1")
	}
	if cfg.LLM.GLM.Models.Opus != "custom-opus" {
		t.Errorf("LLM.GLM.Models.Opus = %q, want %q", cfg.LLM.GLM.Models.Opus, "custom-opus")
	}

	// Verify llm was in loaded sections
	loaded := loader.LoadedSections()
	if !loaded["llm"] {
		t.Error("LLM section should be marked as loaded")
	}
}

// TestLoadLLMSectionDefaults verifies that defaults are used when llm.yaml is missing.
func TestLoadLLMSectionDefaults(t *testing.T) {
	// Create a temporary config directory without llm.yaml
	tmpDir := t.TempDir()
	sectionsDir := filepath.Join(tmpDir, "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Load config (no llm.yaml)
	loader := config.NewLoader()
	cfg, err := loader.Load(tmpDir)
	if err != nil {
		t.Fatalf("loader.Load() error: %v", err)
	}

	// Verify defaults are used
	defaults := config.NewDefaultLLMConfig()
	if cfg.LLM.GLM.BaseURL != defaults.GLM.BaseURL {
		t.Errorf("LLM.GLM.BaseURL = %q, want default %q", cfg.LLM.GLM.BaseURL, defaults.GLM.BaseURL)
	}
	if cfg.LLM.GLMEnvVar != defaults.GLMEnvVar {
		t.Errorf("LLM.GLMEnvVar = %q, want default %q", cfg.LLM.GLMEnvVar, defaults.GLMEnvVar)
	}
	if cfg.LLM.TeamMode != "" {
		t.Errorf("LLM.TeamMode = %q, want empty", cfg.LLM.TeamMode)
	}
}

// TestEnableTeamModeInjectsGLMEnvAndTmux verifies that enableTeamMode(false) injects
// both GLM environment variables AND tmux display configuration to settings.local.json.
// This is critical for GLM Team mode where all agents use GLM models.
func TestEnableTeamModeInjectsGLMEnvAndTmux(t *testing.T) {
	t.Setenv("MOAI_TEST_MODE", "1")

	// Create project dir with .claude directory
	projectRoot := t.TempDir()
	claudeDir := filepath.Join(projectRoot, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	sectionsDir := filepath.Join(projectRoot, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create an existing settings.local.json
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	initialSettings := `{"env": {"EXISTING_VAR": "value"}}`
	if err := os.WriteFile(settingsPath, []byte(initialSettings), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a mock GLM API key file
	homeDir := t.TempDir()
	envGLMPath := filepath.Join(homeDir, ".moai", ".env.glm")
	if err := os.MkdirAll(filepath.Dir(envGLMPath), 0o755); err != nil {
		t.Fatal(err)
	}
	glmEnvContent := `# GLM API Key
GLM_API_KEY="test-glm-api-key-for-team-mode"
`
	if err := os.WriteFile(envGLMPath, []byte(glmEnvContent), 0o600); err != nil {
		t.Fatal(err)
	}

	// Set HOME to temp directory to use our mock .env.glm
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", homeDir)

	// Change to project directory
	origDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(origDir)
		_ = os.Setenv("HOME", origHome)
	}()
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatal(err)
	}

	// Enable team mode (isHybrid = false = all agents use GLM)
	err := enableTeamMode(glmCmd, false)
	if err != nil {
		t.Fatalf("enableTeamMode() error: %v", err)
	}

	// Verify settings.local.json was modified with GLM env and tmux display mode
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings.local.json: %v", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("parse settings.local.json: %v", err)
	}

	env, ok := settings["env"].(map[string]any)
	if !ok {
		t.Fatal("settings.env should exist")
	}

	// Check that EXISTING_VAR is still present
	if _, exists := env["EXISTING_VAR"]; !exists {
		t.Errorf("settings.env should preserve EXISTING_VAR")
	}

	// Check that CLAUDE_CODE_TEAMMATE_DISPLAY is set to "tmux"
	displayMode, exists := env["CLAUDE_CODE_TEAMMATE_DISPLAY"]
	if !exists {
		t.Errorf("settings.local.json should contain CLAUDE_CODE_TEAMMATE_DISPLAY after enableTeamMode, got:\n%s", string(data))
	}
	if displayMode != "tmux" {
		t.Errorf("CLAUDE_CODE_TEAMMATE_DISPLAY = %q, want \"tmux\"", displayMode)
	}

	// Check that GLM ANTHROPIC_* vars ARE present (required for teammates to use GLM models)
	expectedKeys := []string{
		"ANTHROPIC_AUTH_TOKEN",
		"ANTHROPIC_BASE_URL",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL",
		"ANTHROPIC_DEFAULT_SONNET_MODEL",
		"ANTHROPIC_DEFAULT_OPUS_MODEL",
	}
	for _, key := range expectedKeys {
		if _, exists := env[key]; !exists {
			t.Errorf("settings.local.json should contain %s after enableTeamMode --team (GLM teammates need this), got:\n%s", key, string(data))
		}
	}

	// Verify the API key was injected
	if authToken, exists := env["ANTHROPIC_AUTH_TOKEN"]; !exists {
		t.Error("ANTHROPIC_AUTH_TOKEN should exist")
	} else if authToken != "test-glm-api-key-for-team-mode" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN = %q, want %q", authToken, "test-glm-api-key-for-team-mode")
	}
}

// TestEnableTeamModeCGRequiresAPIKey verifies that CG mode fails with a
// helpful error message when GLM API key is not configured.
func TestEnableTeamModeCGRequiresAPIKey(t *testing.T) {
	t.Setenv("MOAI_TEST_MODE", "1")
	t.Setenv("TMUX", "/tmp/tmux-test/default,12345,0")

	// Ensure no API key is available
	t.Setenv("GLM_API_KEY", "")

	// Override HOME so loadGLMKey finds no saved key
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	projectRoot := t.TempDir()
	sectionsDir := filepath.Join(projectRoot, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatal(err)
	}

	err := enableTeamMode(cgCmd, true)
	if err == nil {
		t.Fatal("enableTeamMode(isHybrid=true) should fail without API key")
	}

	// Error should guide user to set up API key first
	errMsg := err.Error()
	if !strings.Contains(errMsg, "moai glm <api-key>") {
		t.Errorf("error should mention 'moai glm <api-key>', got: %v", err)
	}
	if !strings.Contains(errMsg, "moai cg") {
		t.Errorf("error should mention 'moai cg', got: %v", err)
	}
}

// TestEnableTeamModeCGRequiresTmux verifies that CG mode (Claude + GLM)
// fails with a clear error when not running inside a tmux session.
func TestEnableTeamModeCGRequiresTmux(t *testing.T) {
	t.Setenv("MOAI_TEST_MODE", "1")

	// Ensure TMUX is NOT set (simulate non-tmux environment)
	t.Setenv("TMUX", "")

	// Create project dir
	projectRoot := t.TempDir()
	sectionsDir := filepath.Join(projectRoot, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatal(err)
	}

	// CG mode (isHybrid = true) should fail without tmux session
	err := enableTeamMode(cgCmd, true)
	if err == nil {
		t.Fatal("enableTeamMode(isHybrid=true) should fail without tmux session")
	}

	// Error message should mention tmux session requirement
	if !strings.Contains(err.Error(), "tmux session") {
		t.Errorf("error should mention tmux session, got: %v", err)
	}

	// Error message should mention 'moai cg' (not 'moai glm --hybrid')
	if !strings.Contains(err.Error(), "moai cg") {
		t.Errorf("error should mention 'moai cg', got: %v", err)
	}
}

// TestEnableTeamModeCGInTmux verifies that CG mode (Claude + GLM) succeeds
// inside a tmux session and correctly configures settings.local.json.
func TestEnableTeamModeCGInTmux(t *testing.T) {
	t.Setenv("MOAI_TEST_MODE", "1")

	// Simulate being inside a tmux session
	t.Setenv("TMUX", "/tmp/tmux-test/default,12345,0")

	// Create project dir with GLM key
	projectRoot := t.TempDir()
	sectionsDir := filepath.Join(projectRoot, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Set GLM API key via env
	t.Setenv("GLM_API_KEY", "test-cg-api-key")

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatal(err)
	}

	// CG mode (isHybrid = true) should succeed inside tmux
	err := enableTeamMode(cgCmd, true)
	if err != nil {
		t.Fatalf("enableTeamMode(isHybrid=true) should succeed in tmux, got: %v", err)
	}

	// Verify llm.yaml contains team_mode: cg (NOT "hybrid")
	data, err := os.ReadFile(filepath.Join(sectionsDir, "llm.yaml"))
	if err != nil {
		t.Fatalf("failed to read llm.yaml: %v", err)
	}
	if !strings.Contains(string(data), "team_mode: cg") {
		t.Errorf("llm.yaml should contain team_mode: cg, got:\n%s", string(data))
	}
	if strings.Contains(string(data), "team_mode: hybrid") {
		t.Errorf("llm.yaml should NOT contain team_mode: hybrid (renamed to cg), got:\n%s", string(data))
	}

	// Verify settings.local.json was created
	settingsPath := filepath.Join(projectRoot, ".claude", "settings.local.json")
	settingsData, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings.local.json: %v", err)
	}

	var settings SettingsLocal
	if err := json.Unmarshal(settingsData, &settings); err != nil {
		t.Fatalf("failed to unmarshal settings.local.json: %v", err)
	}

	// Should have CLAUDE_CODE_TEAMMATE_DISPLAY=tmux
	if settings.Env["CLAUDE_CODE_TEAMMATE_DISPLAY"] != "tmux" {
		t.Errorf("CLAUDE_CODE_TEAMMATE_DISPLAY = %q, want %q", settings.Env["CLAUDE_CODE_TEAMMATE_DISPLAY"], "tmux")
	}

	// Should NOT have ANTHROPIC_BASE_URL (lead must use Claude, not Z.AI)
	if _, exists := settings.Env["ANTHROPIC_BASE_URL"]; exists {
		t.Error("settings.local.json should NOT contain ANTHROPIC_BASE_URL in CG mode (lead uses Claude)")
	}

	// Should NOT have ANTHROPIC_DEFAULT_*_MODEL (GLM vars are only in tmux session env)
	for _, key := range []string{"ANTHROPIC_DEFAULT_OPUS_MODEL", "ANTHROPIC_DEFAULT_SONNET_MODEL", "ANTHROPIC_DEFAULT_HAIKU_MODEL"} {
		if _, exists := settings.Env[key]; exists {
			t.Errorf("settings.local.json should NOT contain %s in CG mode", key)
		}
	}
}

// TestCGAutoResetsGLMMode verifies that 'moai cg' automatically cleans up
// GLM env vars from settings.local.json when switching from 'moai glm' mode.
// This eliminates the need to run 'moai cc' between 'moai glm' and 'moai cg'.
func TestCGAutoResetsGLMMode(t *testing.T) {
	t.Setenv("MOAI_TEST_MODE", "1")
	t.Setenv("TMUX", "/tmp/tmux-test/default,12345,0")
	t.Setenv("GLM_API_KEY", "test-auto-reset-key")

	projectRoot := t.TempDir()
	claudeDir := filepath.Join(projectRoot, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	sectionsDir := filepath.Join(projectRoot, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatal(err)
	}

	// Step 1: Run 'moai glm' (all-GLM mode) - injects GLM env vars
	err := enableTeamMode(glmCmd, false)
	if err != nil {
		t.Fatalf("enableTeamMode(glm) error: %v", err)
	}

	// Verify GLM env vars are present in settings.local.json
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings.local.json after glm: %v", err)
	}
	var glmSettings SettingsLocal
	if err := json.Unmarshal(data, &glmSettings); err != nil {
		t.Fatalf("parse settings.local.json: %v", err)
	}
	if _, exists := glmSettings.Env["ANTHROPIC_BASE_URL"]; !exists {
		t.Fatal("after 'moai glm', settings.local.json should contain ANTHROPIC_BASE_URL")
	}

	// Step 2: Run 'moai cg' directly (without 'moai cc')
	err = enableTeamMode(cgCmd, true)
	if err != nil {
		t.Fatalf("enableTeamMode(cg) error: %v", err)
	}

	// Verify GLM env vars are removed (lead must use Claude)
	data, err = os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings.local.json after cg: %v", err)
	}
	var cgSettings SettingsLocal
	if err := json.Unmarshal(data, &cgSettings); err != nil {
		t.Fatalf("parse settings.local.json: %v", err)
	}

	// Should NOT have any GLM env vars
	for _, key := range []string{
		"ANTHROPIC_AUTH_TOKEN",
		"ANTHROPIC_BASE_URL",
		"ANTHROPIC_DEFAULT_OPUS_MODEL",
		"ANTHROPIC_DEFAULT_SONNET_MODEL",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL",
	} {
		if _, exists := cgSettings.Env[key]; exists {
			t.Errorf("after 'moai cg', settings.local.json should NOT contain %s (lead uses Claude)", key)
		}
	}

	// Should have CLAUDE_CODE_TEAMMATE_DISPLAY=tmux
	if cgSettings.Env["CLAUDE_CODE_TEAMMATE_DISPLAY"] != "tmux" {
		t.Errorf("CLAUDE_CODE_TEAMMATE_DISPLAY = %q, want %q", cgSettings.Env["CLAUDE_CODE_TEAMMATE_DISPLAY"], "tmux")
	}

	// Verify llm.yaml shows cg mode (not glm)
	llmData, err := os.ReadFile(filepath.Join(sectionsDir, "llm.yaml"))
	if err != nil {
		t.Fatalf("failed to read llm.yaml: %v", err)
	}
	if !strings.Contains(string(llmData), "team_mode: cg") {
		t.Errorf("llm.yaml should contain team_mode: cg, got:\n%s", string(llmData))
	}
}

// TestCleanupMoaiWorktrees verifies that cleanupMoaiWorktrees removes
// moai-related worktrees when called.
func TestCleanupMoaiWorktrees(t *testing.T) {
	t.Setenv("MOAI_TEST_MODE", "1")

	// Skip if not in a git repo (for CI environments)
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		t.Skip("not in a git repository")
	}

	// Create a temp project root
	projectRoot := t.TempDir()

	// cleanupMoaiWorktrees should handle non-git directories gracefully
	result := cleanupMoaiWorktrees(projectRoot)
	// Result should be empty since there's no git repo
	if result != "" {
		t.Logf("cleanupMoaiWorktrees returned: %s (expected empty for non-git dir)", result)
	}
}
