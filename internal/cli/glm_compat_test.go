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
// when they contain explicit non-empty values (offered-set values pass
// through per REQ-MMU-004).
func TestLoadGLMConfig_NewFormat(t *testing.T) {
	mgr := setupGLMTestConfig(t, `
llm:
  glm:
    base_url: "https://api.z.ai/api/anthropic"
    models:
      high: "glm-5.3"
      medium: "glm-5.3-flash"
      low: "glm-5.3-flash"
`)
	origDeps := deps
	deps = &Dependencies{Config: mgr}
	defer func() { deps = origDeps }()

	cfg, err := loadGLMConfig("/unused")
	if err != nil {
		t.Fatalf("loadGLMConfig should not error: %v", err)
	}
	if cfg.Models.High != "glm-5.3" {
		t.Errorf("Models.High = %q, want %q", cfg.Models.High, "glm-5.3")
	}
	if cfg.Models.Medium != "glm-5.3-flash" {
		t.Errorf("Models.Medium = %q, want %q", cfg.Models.Medium, "glm-5.3-flash")
	}
	if cfg.Models.Low != "glm-5.3-flash" {
		t.Errorf("Models.Low = %q, want %q", cfg.Models.Low, "glm-5.3-flash")
	}
}

// TestLoadGLMConfig_LegacyFields pins the post-DR-2 behavior (REQ-MMU-004):
// the legacy opus/sonnet/haiku alias FIELDS are deleted, so an llm.yaml
// carrying them loads without error (non-strict loader silently ignores the
// keys — the accepted silent half of DR-2) and the empty tier slots resolve
// to the tier defaults, never to the ignored alias values.
func TestLoadGLMConfig_LegacyFields(t *testing.T) {
	mgr := setupGLMTestConfig(t, `
llm:
  glm:
    base_url: "https://api.z.ai/api/anthropic"
    models:
      high: ""
      medium: ""
      low: ""
      opus: "glm-4.7"
      sonnet: "glm-5.1"
      haiku: "glm-4.6"
`)
	origDeps := deps
	deps = &Dependencies{Config: mgr}
	defer func() { deps = origDeps }()

	cfg, err := loadGLMConfig("/unused")
	if err != nil {
		t.Fatalf("loadGLMConfig should not error (alias keys are ignored, not rejected): %v", err)
	}
	sysDefaults := config.NewDefaultLLMConfig()
	if cfg.Models.High != sysDefaults.GLM.Models.High {
		t.Errorf("Models.High = %q, want the tier default %q (alias keys are ignored — DR-2)", cfg.Models.High, sysDefaults.GLM.Models.High)
	}
	if cfg.Models.Medium != sysDefaults.GLM.Models.Medium {
		t.Errorf("Models.Medium = %q, want the tier default %q (alias keys are ignored — DR-2)", cfg.Models.Medium, sysDefaults.GLM.Models.Medium)
	}
	if cfg.Models.Low != sysDefaults.GLM.Models.Low {
		t.Errorf("Models.Low = %q, want the tier default %q (alias keys are ignored — DR-2)", cfg.Models.Low, sysDefaults.GLM.Models.Low)
	}
}

// TestLoadGLMConfig_MixedFormat verifies that the real tier fields resolve
// their own values while the legacy alias keys (present in the same file)
// contribute nothing — REQ-MMU-004 deleted their fields.
func TestLoadGLMConfig_MixedFormat(t *testing.T) {
	mgr := setupGLMTestConfig(t, `
llm:
  glm:
    base_url: "https://api.z.ai/api/anthropic"
    models:
      high: "glm-5.3"
      medium: ""
      low: "glm-5.3-flash"
      opus: "glm-4.7"
      sonnet: "glm-5.1"
      haiku: "glm-4.6"
`)
	origDeps := deps
	deps = &Dependencies{Config: mgr}
	defer func() { deps = origDeps }()

	cfg, err := loadGLMConfig("/unused")
	if err != nil {
		t.Fatalf("loadGLMConfig should not error: %v", err)
	}
	// high is set directly to an offered id, should use it
	if cfg.Models.High != "glm-5.3" {
		t.Errorf("Models.High = %q, want %q (tier field takes precedence)", cfg.Models.High, "glm-5.3")
	}
	// medium is empty, and the sonnet alias key is ignored — tier default
	sysDefaults := config.NewDefaultLLMConfig()
	if cfg.Models.Medium != sysDefaults.GLM.Models.Medium {
		t.Errorf("Models.Medium = %q, want the tier default %q (alias keys contribute nothing — DR-2)", cfg.Models.Medium, sysDefaults.GLM.Models.Medium)
	}
	// low is set directly, should use it
	if cfg.Models.Low != "glm-5.3-flash" {
		t.Errorf("Models.Low = %q, want %q (tier field takes precedence)", cfg.Models.Low, "glm-5.3-flash")
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
