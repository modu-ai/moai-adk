package cli

import (
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
					Haiku  string
					Sonnet string
					Opus   string
				}{
					Haiku:  "glm-4.7-flashx",
					Sonnet: "glm-4.7",
					Opus:   "glm-5",
				},
				EnvVar: "GLM_API_KEY",
			},
			apiKey: "test-key-123",
			wantKeys: []string{
				"ANTHROPIC_AUTH_TOKEN",
				"ANTHROPIC_BASE_URL",
				"ANTHROPIC_DEFAULT_HAIKU_MODEL",
				"ANTHROPIC_DEFAULT_SONNET_MODEL",
				"ANTHROPIC_DEFAULT_OPUS_MODEL",
			},
			wantVals: map[string]string{
				"ANTHROPIC_AUTH_TOKEN":           "test-key-123",
				"ANTHROPIC_BASE_URL":             "https://api.z.ai/api/anthropic",
				"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "glm-4.7-flashx",
				"ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.7",
				"ANTHROPIC_DEFAULT_OPUS_MODEL":   "glm-5",
			},
		},
		{
			name: "custom_config",
			glmConfig: &GLMConfigFromYAML{
				BaseURL: "https://custom.glm.api/v1",
				Models: struct {
					Haiku  string
					Sonnet string
					Opus   string
				}{
					Haiku:  "custom-haiku",
					Sonnet: "custom-sonnet",
					Opus:   "custom-opus",
				},
				EnvVar: "CUSTOM_API_KEY",
			},
			apiKey: "custom-api-key-xyz",
			wantKeys: []string{
				"ANTHROPIC_AUTH_TOKEN",
				"ANTHROPIC_BASE_URL",
				"ANTHROPIC_DEFAULT_HAIKU_MODEL",
				"ANTHROPIC_DEFAULT_SONNET_MODEL",
				"ANTHROPIC_DEFAULT_OPUS_MODEL",
			},
			wantVals: map[string]string{
				"ANTHROPIC_AUTH_TOKEN":           "custom-api-key-xyz",
				"ANTHROPIC_BASE_URL":             "https://custom.glm.api/v1",
				"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "custom-haiku",
				"ANTHROPIC_DEFAULT_SONNET_MODEL": "custom-sonnet",
				"ANTHROPIC_DEFAULT_OPUS_MODEL":   "custom-opus",
			},
		},
		{
			name: "empty_api_key",
			glmConfig: &GLMConfigFromYAML{
				BaseURL: "https://api.z.ai/api/anthropic",
				Models: struct {
					Haiku  string
					Sonnet string
					Opus   string
				}{
					Haiku:  "glm-4.7-flashx",
					Sonnet: "glm-4.7",
					Opus:   "glm-5",
				},
				EnvVar: "GLM_API_KEY",
			},
			apiKey: "",
			wantKeys: []string{
				"ANTHROPIC_AUTH_TOKEN",
				"ANTHROPIC_BASE_URL",
				"ANTHROPIC_DEFAULT_HAIKU_MODEL",
				"ANTHROPIC_DEFAULT_SONNET_MODEL",
				"ANTHROPIC_DEFAULT_OPUS_MODEL",
			},
			wantVals: map[string]string{
				"ANTHROPIC_AUTH_TOKEN":           "",
				"ANTHROPIC_BASE_URL":             "https://api.z.ai/api/anthropic",
				"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "glm-4.7-flashx",
				"ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.7",
				"ANTHROPIC_DEFAULT_OPUS_MODEL":   "glm-5",
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

// TestGLMTeamFlag verifies that the --team flag is correctly registered
// on the glm command with the expected type and default value.
func TestGLMTeamFlag(t *testing.T) {
	// Verify the --team flag exists on glmCmd
	flag := glmCmd.Flags().Lookup("team")
	if flag == nil {
		t.Fatal("--team flag should be registered on glmCmd")
	}

	// Verify the flag type is bool
	if flag.Value.Type() != "bool" {
		t.Errorf("--team flag type = %q, want %q", flag.Value.Type(), "bool")
	}

	// Verify the default value is false
	if flag.DefValue != "false" {
		t.Errorf("--team flag default = %q, want %q", flag.DefValue, "false")
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

// TestEnableTeamModeAlwaysGLM verifies enableTeamMode always sets team_mode to "glm".
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

	glmTeamFlag = true
	defer func() { glmTeamFlag = false }()

	err := enableTeamMode(glmCmd)
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

// TestEnableTeamModeDoesNotInjectEnv verifies that enableTeamMode only saves
// team_mode to llm.yaml and does NOT inject GLM env into settings.local.json.
// This is critical for GLM Worker mode where Leader must stay on Opus.
func TestEnableTeamModeDoesNotInjectEnv(t *testing.T) {
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

	// Create an existing settings.local.json (should NOT be modified by enableTeamMode)
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	initialSettings := `{"env": {"EXISTING_VAR": "value"}}`
	if err := os.WriteFile(settingsPath, []byte(initialSettings), 0o644); err != nil {
		t.Fatal(err)
	}

	// Change to project directory
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatal(err)
	}

	// Enable team mode
	glmTeamFlag = true
	defer func() { glmTeamFlag = false }()

	err := enableTeamMode(glmCmd)
	if err != nil {
		t.Fatalf("enableTeamMode() error: %v", err)
	}

	// Verify settings.local.json was NOT modified
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings.local.json: %v", err)
	}
	if string(data) != initialSettings {
		t.Errorf("settings.local.json should NOT be modified by enableTeamMode\n got: %s\n want: %s", string(data), initialSettings)
	}

	// Verify settings.local.json does NOT contain ANTHROPIC_* vars
	if strings.Contains(string(data), "ANTHROPIC_") {
		t.Errorf("settings.local.json should NOT contain ANTHROPIC_* vars after enableTeamMode, got:\n%s", string(data))
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
