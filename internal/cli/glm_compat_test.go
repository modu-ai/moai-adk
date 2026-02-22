package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// setupGLMTestConfig creates a ConfigManager loaded from a temp directory
// containing an llm.yaml with the given content. The caller is responsible
// for restoring deps after the test.
func setupGLMTestConfig(t *testing.T, llmYAML string) *config.ConfigManager {
	t.Helper()
	tmpDir := t.TempDir()
	sectionsDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sectionsDir, "llm.yaml"), []byte(llmYAML), 0o644); err != nil {
		t.Fatalf("failed to write llm.yaml: %v", err)
	}
	mgr := config.NewConfigManager()
	if _, err := mgr.Load(tmpDir); err != nil {
		t.Fatalf("config load failed: %v", err)
	}
	return mgr
}

// TestLoadGLMConfig_NewFormat verifies that high/medium/low fields are used
// when they contain explicit non-empty values.
func TestLoadGLMConfig_NewFormat(t *testing.T) {
	mgr := setupGLMTestConfig(t, `
llm:
  glm:
    base_url: "https://api.z.ai/api/anthropic"
    models:
      high: "new-high-model"
      medium: "new-medium-model"
      low: "new-low-model"
`)
	origDeps := deps
	deps = &Dependencies{Config: mgr}
	defer func() { deps = origDeps }()

	cfg, err := loadGLMConfig("/unused")
	if err != nil {
		t.Fatalf("loadGLMConfig should not error: %v", err)
	}
	if cfg.Models.High != "new-high-model" {
		t.Errorf("Models.High = %q, want %q", cfg.Models.High, "new-high-model")
	}
	if cfg.Models.Medium != "new-medium-model" {
		t.Errorf("Models.Medium = %q, want %q", cfg.Models.Medium, "new-medium-model")
	}
	if cfg.Models.Low != "new-low-model" {
		t.Errorf("Models.Low = %q, want %q", cfg.Models.Low, "new-low-model")
	}
}

// TestLoadGLMConfig_LegacyFields verifies that opus/sonnet/haiku are used
// as fallbacks when high/medium/low are explicitly empty.
func TestLoadGLMConfig_LegacyFields(t *testing.T) {
	mgr := setupGLMTestConfig(t, `
llm:
  glm:
    base_url: "https://api.z.ai/api/anthropic"
    models:
      high: ""
      medium: ""
      low: ""
      opus: "legacy-high-model"
      sonnet: "legacy-medium-model"
      haiku: "legacy-low-model"
`)
	origDeps := deps
	deps = &Dependencies{Config: mgr}
	defer func() { deps = origDeps }()

	cfg, err := loadGLMConfig("/unused")
	if err != nil {
		t.Fatalf("loadGLMConfig should not error: %v", err)
	}
	if cfg.Models.High != "legacy-high-model" {
		t.Errorf("Models.High = %q, want %q (legacy fallback from opus)", cfg.Models.High, "legacy-high-model")
	}
	if cfg.Models.Medium != "legacy-medium-model" {
		t.Errorf("Models.Medium = %q, want %q (legacy fallback from sonnet)", cfg.Models.Medium, "legacy-medium-model")
	}
	if cfg.Models.Low != "legacy-low-model" {
		t.Errorf("Models.Low = %q, want %q (legacy fallback from haiku)", cfg.Models.Low, "legacy-low-model")
	}
}

// TestLoadGLMConfig_MixedFormat verifies that new fields take precedence
// over legacy fields when both are present, and legacy fills in missing new fields.
func TestLoadGLMConfig_MixedFormat(t *testing.T) {
	mgr := setupGLMTestConfig(t, `
llm:
  glm:
    base_url: "https://api.z.ai/api/anthropic"
    models:
      high: "new-high"
      medium: ""
      low: "new-low"
      opus: "legacy-high"
      sonnet: "legacy-medium"
      haiku: "legacy-low"
`)
	origDeps := deps
	deps = &Dependencies{Config: mgr}
	defer func() { deps = origDeps }()

	cfg, err := loadGLMConfig("/unused")
	if err != nil {
		t.Fatalf("loadGLMConfig should not error: %v", err)
	}
	// high is set directly, should use it
	if cfg.Models.High != "new-high" {
		t.Errorf("Models.High = %q, want %q (new field takes precedence)", cfg.Models.High, "new-high")
	}
	// medium is empty, should fall back to sonnet
	if cfg.Models.Medium != "legacy-medium" {
		t.Errorf("Models.Medium = %q, want %q (legacy fallback from sonnet)", cfg.Models.Medium, "legacy-medium")
	}
	// low is set directly, should use it
	if cfg.Models.Low != "new-low" {
		t.Errorf("Models.Low = %q, want %q (new field takes precedence)", cfg.Models.Low, "new-low")
	}
}

// TestLoadGLMConfig_EmptyFieldsFallToDefaults verifies that when both new and
// legacy model fields are empty, the system defaults are used.
func TestLoadGLMConfig_EmptyFieldsFallToDefaults(t *testing.T) {
	mgr := setupGLMTestConfig(t, `
llm:
  glm:
    base_url: "https://api.z.ai/api/anthropic"
    models:
      high: ""
      medium: ""
      low: ""
      opus: ""
      sonnet: ""
      haiku: ""
`)
	origDeps := deps
	deps = &Dependencies{Config: mgr}
	defer func() { deps = origDeps }()

	cfg, err := loadGLMConfig("/unused")
	if err != nil {
		t.Fatalf("loadGLMConfig should not error: %v", err)
	}
	// When all fields are empty, system defaults must be used.
	sysDefaults := config.NewDefaultLLMConfig()
	if cfg.Models.High != sysDefaults.GLM.Models.High {
		t.Errorf("Models.High = %q, want default %q", cfg.Models.High, sysDefaults.GLM.Models.High)
	}
	if cfg.Models.Medium != sysDefaults.GLM.Models.Medium {
		t.Errorf("Models.Medium = %q, want default %q", cfg.Models.Medium, sysDefaults.GLM.Models.Medium)
	}
	if cfg.Models.Low != sysDefaults.GLM.Models.Low {
		t.Errorf("Models.Low = %q, want default %q", cfg.Models.Low, sysDefaults.GLM.Models.Low)
	}
}
