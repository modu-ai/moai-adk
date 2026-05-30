package config

import (
	"path/filepath"
	"testing"
)

// SPEC-V3R6-PROMPT-CACHE-001 M2 — cache.yaml config schema.
//
// These tests define the contract for LoadCacheConfig: schema validation of the
// cacheStrategy section (enabled / session_ttl / spec_ttl / min_cacheable_tokens)
// with safe-default fallback on missing or invalid input (REQ-PC-005, AC-PC-002).

// TestLoadCacheConfig_ValidParse verifies that a valid cache.yaml is parsed with
// all four cacheStrategy keys populated.
func TestLoadCacheConfig_ValidParse(t *testing.T) {
	path := filepath.Join("testdata", "cache-valid", "cache.yaml")
	cfg, err := LoadCacheConfig(path)
	if err != nil {
		t.Fatalf("LoadCacheConfig(%q): unexpected error: %v", path, err)
	}
	if cfg == nil {
		t.Fatal("LoadCacheConfig returned nil config")
	}
	if !cfg.Enabled {
		t.Errorf("Enabled: want true, got false")
	}
	if cfg.SessionTTL != "1h" {
		t.Errorf("SessionTTL: want %q, got %q", "1h", cfg.SessionTTL)
	}
	if cfg.SpecTTL != "5m" {
		t.Errorf("SpecTTL: want %q, got %q", "5m", cfg.SpecTTL)
	}
	if cfg.MinCacheableTokens != 2048 {
		t.Errorf("MinCacheableTokens: want 2048, got %d", cfg.MinCacheableTokens)
	}
}

// TestLoadCacheConfig_MissingFile_ReturnsSafeDefault verifies that an absent
// cache.yaml returns the safe default (enabled: false) with no error.
func TestLoadCacheConfig_MissingFile_ReturnsSafeDefault(t *testing.T) {
	path := filepath.Join("testdata", "nonexistent-dir", "cache.yaml")
	cfg, err := LoadCacheConfig(path)
	if err != nil {
		t.Fatalf("LoadCacheConfig with absent file: expected no error, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadCacheConfig returned nil config on missing file")
	}
	if cfg.Enabled {
		t.Errorf("missing file safe default: Enabled want false, got true")
	}
}

// TestLoadCacheConfig_Malformed_ReturnsSafeDefault verifies that malformed YAML
// falls back to the safe default (enabled: false) rather than erroring — per
// plan.md M2: "Missing/invalid → log warning, fall back to enabled: false".
func TestLoadCacheConfig_Malformed_ReturnsSafeDefault(t *testing.T) {
	path := filepath.Join("testdata", "cache-malformed", "cache.yaml")
	cfg, err := LoadCacheConfig(path)
	if err != nil {
		t.Fatalf("malformed cache.yaml: expected safe-default fallback (no error), got: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadCacheConfig returned nil on malformed input")
	}
	if cfg.Enabled {
		t.Errorf("malformed safe default: Enabled want false, got true")
	}
}

// TestLoadCacheConfig_InvalidSessionTTL_FallsBackToDisabled verifies that an
// out-of-enum session_ttl value triggers the safe default (enabled: false).
func TestLoadCacheConfig_InvalidSessionTTL_FallsBackToDisabled(t *testing.T) {
	path := filepath.Join("testdata", "cache-invalid-ttl", "cache.yaml")
	cfg, err := LoadCacheConfig(path)
	if err != nil {
		t.Fatalf("invalid session_ttl: expected safe-default fallback, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("nil config on invalid session_ttl")
	}
	if cfg.Enabled {
		t.Errorf("invalid session_ttl: Enabled want false (safe default), got true")
	}
}

// TestDefaultCacheConfig verifies the built-in defaults.
func TestDefaultCacheConfig(t *testing.T) {
	cfg := DefaultCacheConfig()
	if cfg == nil {
		t.Fatal("DefaultCacheConfig returned nil")
	}
	// Default ships disabled (safe default) until the user opts in.
	if cfg.Enabled {
		t.Errorf("DefaultCacheConfig Enabled: want false, got true")
	}
	if cfg.SessionTTL != "1h" {
		t.Errorf("DefaultCacheConfig SessionTTL: want %q, got %q", "1h", cfg.SessionTTL)
	}
	if cfg.SpecTTL != "5m" {
		t.Errorf("DefaultCacheConfig SpecTTL: want %q, got %q", "5m", cfg.SpecTTL)
	}
	if cfg.MinCacheableTokens != 2048 {
		t.Errorf("DefaultCacheConfig MinCacheableTokens: want 2048, got %d", cfg.MinCacheableTokens)
	}
}

// TestCacheConfig_Validate covers the Validate() invariants for each field.
func TestCacheConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     CacheConfig
		wantErr bool
	}{
		{
			name:    "valid 1h/5m",
			cfg:     CacheConfig{Enabled: true, SessionTTL: "1h", SpecTTL: "5m", MinCacheableTokens: 2048},
			wantErr: false,
		},
		{
			name:    "valid session 5m",
			cfg:     CacheConfig{Enabled: true, SessionTTL: "5m", SpecTTL: "5m", MinCacheableTokens: 1024},
			wantErr: false,
		},
		{
			name:    "valid session off",
			cfg:     CacheConfig{Enabled: true, SessionTTL: "off", SpecTTL: "off", MinCacheableTokens: 2048},
			wantErr: false,
		},
		{
			name:    "invalid session_ttl value",
			cfg:     CacheConfig{Enabled: true, SessionTTL: "2h", SpecTTL: "5m", MinCacheableTokens: 2048},
			wantErr: true,
		},
		{
			name:    "invalid spec_ttl value",
			cfg:     CacheConfig{Enabled: true, SessionTTL: "1h", SpecTTL: "1h", MinCacheableTokens: 2048},
			wantErr: true,
		},
		{
			name:    "negative min_cacheable_tokens",
			cfg:     CacheConfig{Enabled: true, SessionTTL: "1h", SpecTTL: "5m", MinCacheableTokens: -1},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr && err == nil {
				t.Errorf("Validate(): expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate(): unexpected error: %v", err)
			}
		})
	}
}
