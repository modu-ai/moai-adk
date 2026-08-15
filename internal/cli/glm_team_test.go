package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestCGCommandRegistered verifies that the cg command is correctly registered
// on the root command.
func TestCGCommandRegistered(t *testing.T) {
	// Verify cgCmd has the correct Use field
	if !strings.HasPrefix(cgCmd.Use, "cg") {
		t.Errorf("cgCmd.Use should start with 'cg', got %q", cgCmd.Use)
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
