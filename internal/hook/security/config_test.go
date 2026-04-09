package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadExtraSecurityConfig_NotFound(t *testing.T) {
	t.Parallel()
	cfg := LoadExtraSecurityConfig(t.TempDir())
	if cfg != nil {
		t.Error("expected nil when security.yaml does not exist")
	}
}

func TestLoadExtraSecurityConfig_ValidYAML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	yaml := `security:
  extra_dangerous_bash_patterns:
    - 'curl.*\|.*sh'
    - 'chmod\s+777'
  extra_deny_patterns:
    - '\.secret$'
`
	if err := os.WriteFile(filepath.Join(cfgDir, "security.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := LoadExtraSecurityConfig(dir)
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if len(cfg.Security.ExtraDangerousBashPatterns) != 2 {
		t.Errorf("expected 2 bash patterns, got %d", len(cfg.Security.ExtraDangerousBashPatterns))
	}
	if len(cfg.Security.ExtraDenyPatterns) != 1 {
		t.Errorf("expected 1 deny pattern, got %d", len(cfg.Security.ExtraDenyPatterns))
	}
}

func TestLoadExtraSecurityConfig_InvalidYAML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(cfgDir, "security.yaml"), []byte(":::invalid"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := LoadExtraSecurityConfig(dir)
	if cfg != nil {
		t.Error("expected nil for invalid YAML")
	}
}
