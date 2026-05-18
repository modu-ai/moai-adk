package config

import (
	"errors"
	"path/filepath"
	"testing"
)

// TestLoadContextConfig_ValidParse verifies that a valid context.yaml is parsed
// correctly with token_budget and search fields populated.
// Maps to: REQ-MIG003-001/002/003/007/010, AC-MIG003-02/07/09
func TestLoadContextConfig_ValidParse(t *testing.T) {
	path := filepath.Join("testdata", "context-valid", "context.yaml")
	cfg, err := LoadContextConfig(path)
	if err != nil {
		t.Fatalf("LoadContextConfig(%q): unexpected error: %v", path, err)
	}
	if cfg == nil {
		t.Fatal("LoadContextConfig returned nil config")
	}

	// REQ-MIG003-010: token_budget fields
	if cfg.TokenBudget.MaxInjectionTokens != 5000 {
		t.Errorf("TokenBudget.MaxInjectionTokens: want 5000, got %d", cfg.TokenBudget.MaxInjectionTokens)
	}
	if cfg.TokenBudget.SkipIfUsageAbove != 150000 {
		t.Errorf("TokenBudget.SkipIfUsageAbove: want 150000, got %d", cfg.TokenBudget.SkipIfUsageAbove)
	}

	// REQ-MIG003-010: search.date_range_days (staleness_window_days alias)
	if cfg.Search.DateRangeDays != 30 {
		t.Errorf("Search.DateRangeDays: want 30, got %d", cfg.Search.DateRangeDays)
	}
}

// TestLoadContextConfig_MissingFile_ReturnsDefaults verifies that an absent
// context.yaml returns defaults with no error.
// Maps to: REQ-MIG003-004, AC-MIG003-03
func TestLoadContextConfig_MissingFile_ReturnsDefaults(t *testing.T) {
	path := filepath.Join("testdata", "nonexistent-dir", "context.yaml")
	cfg, err := LoadContextConfig(path)
	if err != nil {
		t.Fatalf("LoadContextConfig with absent file: expected no error, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadContextConfig returned nil config on missing file")
	}
}

// TestLoadContextConfig_Malformed_ReturnsErrInvalidYAML verifies that malformed
// context.yaml returns an error wrapping ErrInvalidYAML.
// Maps to: REQ-MIG003-003/007, AC-MIG003-04/07
func TestLoadContextConfig_Malformed_ReturnsErrInvalidYAML(t *testing.T) {
	path := filepath.Join("testdata", "context-malformed", "context.yaml")
	_, err := LoadContextConfig(path)
	if err == nil {
		t.Fatal("expected error for malformed YAML, got nil")
	}
	if !errors.Is(err, ErrInvalidYAML) {
		t.Errorf("expected errors.Is(err, ErrInvalidYAML), got: %v", err)
	}
}
