package mx

import (
	"os"
	"path/filepath"
	"testing"
)

// TestParseValidationConfig_Defaults verifies AC-CONFIG-002:
// when no validation section exists, defaults are applied.
func TestParseValidationConfig_Defaults(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Write mx.yaml without validation section
	mxYAML := `mx:
    thresholds:
        fan_in_anchor: 3
        complexity_warn: 15
`
	if err := os.WriteFile(filepath.Join(dir, "mx.yaml"), []byte(mxYAML), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := ParseValidationConfig(filepath.Join(dir, "mx.yaml"))
	if err != nil {
		t.Fatalf("ParseValidationConfig() error = %v", err)
	}

	// AC-CONFIG-002: verify defaults
	if !cfg.Enabled {
		t.Error("Enabled default should be true")
	}
	if cfg.PostToolUse.TimeoutMs != 500 {
		t.Errorf("PostToolUse.TimeoutMs = %d, want 500", cfg.PostToolUse.TimeoutMs)
	}
	if cfg.SessionEnd.TimeoutMs != 4000 {
		t.Errorf("SessionEnd.TimeoutMs = %d, want 4000", cfg.SessionEnd.TimeoutMs)
	}
	if cfg.Sync.Enforcement != "strict" {
		t.Errorf("Sync.Enforcement = %q, want %q", cfg.Sync.Enforcement, "strict")
	}
	if cfg.EnforcementLevels.P1Anchor != "blocking" {
		t.Errorf("EnforcementLevels.P1Anchor = %q, want %q", cfg.EnforcementLevels.P1Anchor, "blocking")
	}
}

// TestParseValidationConfig_Custom verifies AC-CONFIG-001:
// explicit validation section overrides defaults.
func TestParseValidationConfig_Custom(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mxYAML := `mx:
    thresholds:
        fan_in_anchor: 3
    validation:
        enabled: true
        post_tool_use:
            enabled: false
            timeout_ms: 300
        session_end:
            enabled: true
            timeout_ms: 3000
        sync:
            enforcement: advisory
        enforcement_levels:
            p1_anchor: blocking
            p2_warn: advisory
            p3_note: advisory
            p4_todo: advisory
`
	if err := os.WriteFile(filepath.Join(dir, "mx.yaml"), []byte(mxYAML), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := ParseValidationConfig(filepath.Join(dir, "mx.yaml"))
	if err != nil {
		t.Fatalf("ParseValidationConfig() error = %v", err)
	}

	// AC-CONFIG-001: verify custom values
	if !cfg.Enabled {
		t.Error("Enabled should be true")
	}
	if cfg.PostToolUse.Enabled {
		t.Error("PostToolUse.Enabled should be false")
	}
	if cfg.PostToolUse.TimeoutMs != 300 {
		t.Errorf("PostToolUse.TimeoutMs = %d, want 300", cfg.PostToolUse.TimeoutMs)
	}
	if cfg.SessionEnd.TimeoutMs != 3000 {
		t.Errorf("SessionEnd.TimeoutMs = %d, want 3000", cfg.SessionEnd.TimeoutMs)
	}
	if cfg.Sync.Enforcement != "advisory" {
		t.Errorf("Sync.Enforcement = %q, want %q", cfg.Sync.Enforcement, "advisory")
	}
	if cfg.EnforcementLevels.P2Warn != "advisory" {
		t.Errorf("EnforcementLevels.P2Warn = %q, want %q", cfg.EnforcementLevels.P2Warn, "advisory")
	}
}

// TestParseValidationConfig_MissingFile verifies graceful handling when
// mx.yaml does not exist (returns defaults, no error).
func TestParseValidationConfig_MissingFile(t *testing.T) {
	t.Parallel()

	cfg, err := ParseValidationConfig("/nonexistent/path/mx.yaml")
	if err != nil {
		t.Fatalf("ParseValidationConfig() should not error for missing file, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("ParseValidationConfig() returned nil config for missing file")
	}
	// Should return defaults
	if !cfg.Enabled {
		t.Error("Enabled default should be true for missing file")
	}
}

// TestParseValidationConfig_PostToolUseDefaults verifies PostToolUse
// defaults when section exists but fields are missing.
func TestParseValidationConfig_PostToolUseDefaults(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mxYAML := `mx:
    validation:
        enabled: true
        post_tool_use:
            enabled: true
`
	if err := os.WriteFile(filepath.Join(dir, "mx.yaml"), []byte(mxYAML), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := ParseValidationConfig(filepath.Join(dir, "mx.yaml"))
	if err != nil {
		t.Fatalf("ParseValidationConfig() error = %v", err)
	}

	// timeout_ms not specified → should use default 500
	if cfg.PostToolUse.TimeoutMs != 500 {
		t.Errorf("PostToolUse.TimeoutMs = %d, want 500 (default)", cfg.PostToolUse.TimeoutMs)
	}
}

// TestValidationConfig_DefaultConfig verifies DefaultConfig returns valid defaults.
func TestValidationConfig_DefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultValidationConfig()
	if cfg == nil {
		t.Fatal("DefaultValidationConfig() returned nil")
	}
	if !cfg.Enabled {
		t.Error("default Enabled should be true")
	}
	if cfg.PostToolUse.TimeoutMs <= 0 {
		t.Errorf("default PostToolUse.TimeoutMs = %d, want > 0", cfg.PostToolUse.TimeoutMs)
	}
	if cfg.SessionEnd.TimeoutMs <= 0 {
		t.Errorf("default SessionEnd.TimeoutMs = %d, want > 0", cfg.SessionEnd.TimeoutMs)
	}
}
