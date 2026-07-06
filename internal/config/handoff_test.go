package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewDefaultHandoffConfig verifies the compiled default: Mode == "manual"
// (auto-resume opt-in) and Guide == false.
// AC-AUTORESUME-001 → REQ-001.
func TestNewDefaultHandoffConfig(t *testing.T) {
	t.Parallel()

	hc := NewDefaultHandoffConfig()
	if hc.Mode != "manual" {
		t.Errorf("NewDefaultHandoffConfig().Mode: got %q, want %q", hc.Mode, "manual")
	}
	if hc.Guide != false {
		t.Errorf("NewDefaultHandoffConfig().Guide: got %v, want false", hc.Guide)
	}
	// Struct has exactly the two documented fields (no Consume — YAGNI removed).
	// This literal compiles only because HandoffConfig has no other required
	// field; it is a documentation assertion of the 2-field shape.
	_ = HandoffConfig{Mode: "auto", Guide: true}
}

// TestNewDefaultConfig_IncludesHandoff verifies NewDefaultConfig wires the
// Handoff section so an absent handoff.yaml still resolves the manual default.
func TestNewDefaultConfig_IncludesHandoff(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()
	if cfg.Handoff.Mode != "manual" {
		t.Errorf("NewDefaultConfig().Handoff.Mode: got %q, want %q", cfg.Handoff.Mode, "manual")
	}
}

// TestLoadHandoffSection_PartialOverride verifies REQ-003: a handoff.yaml
// specifying only `mode: auto` (guide omitted) retains the construction-time
// default for guide (false) rather than collapsing it to its zero-value.
// AC-AUTORESUME-003 → REQ-003.
func TestLoadHandoffSection_PartialOverride(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sectionsDir := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}

	// mode only — guide omitted (partial override).
	handoffYAML := []byte("handoff:\n    mode: auto\n")
	if err := os.WriteFile(filepath.Join(sectionsDir, "handoff.yaml"), handoffYAML, 0o644); err != nil {
		t.Fatalf("failed to write handoff.yaml: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(tempDir, ".moai"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Handoff.Mode != "auto" {
		t.Errorf("cfg.Handoff.Mode: got %q, want %q", cfg.Handoff.Mode, "auto")
	}
	if cfg.Handoff.Guide != false {
		t.Errorf("cfg.Handoff.Guide (omitted key): got %v, want default false (zero-value collapse regression)", cfg.Handoff.Guide)
	}
	if !loader.LoadedSections()["handoff"] {
		t.Error("expected handoff section to be marked loaded when handoff.yaml is present")
	}
}

// TestLoadHandoffSection_FullOverride verifies both keys load when both present.
func TestLoadHandoffSection_FullOverride(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sectionsDir := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}

	handoffYAML := []byte("handoff:\n    mode: auto\n    guide: true\n")
	if err := os.WriteFile(filepath.Join(sectionsDir, "handoff.yaml"), handoffYAML, 0o644); err != nil {
		t.Fatalf("failed to write handoff.yaml: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(tempDir, ".moai"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Handoff.Mode != "auto" || cfg.Handoff.Guide != true {
		t.Errorf("cfg.Handoff = %+v, want {Mode:auto Guide:true}", cfg.Handoff)
	}
}

// TestLoadHandoffSection_AbsentUsesDefault verifies an absent handoff.yaml keeps
// the manual default and does NOT mark the section as loaded.
func TestLoadHandoffSection_AbsentUsesDefault(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sectionsDir := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(tempDir, ".moai"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Handoff.Mode != "manual" {
		t.Errorf("cfg.Handoff.Mode with no handoff.yaml: got %q, want default %q", cfg.Handoff.Mode, "manual")
	}
	if loader.LoadedSections()["handoff"] {
		t.Error("expected handoff section to NOT be loaded when file is missing")
	}
}

// TestLoadHandoffSection_InvalidYAML verifies a malformed handoff.yaml is skipped
// gracefully (loader warns, retains default) rather than aborting Load().
func TestLoadHandoffSection_InvalidYAML(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sectionsDir := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}

	badYAML := []byte("handoff:\n    mode: [unterminated\n")
	if err := os.WriteFile(filepath.Join(sectionsDir, "handoff.yaml"), badYAML, 0o644); err != nil {
		t.Fatalf("failed to write handoff.yaml: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(tempDir, ".moai"))
	if err != nil {
		t.Fatalf("Load() should not fail on a single malformed section: %v", err)
	}
	if cfg.Handoff.Mode != "manual" {
		t.Errorf("cfg.Handoff.Mode after malformed handoff.yaml: got %q, want default %q", cfg.Handoff.Mode, "manual")
	}
	if loader.LoadedSections()["handoff"] {
		t.Error("expected handoff section to NOT be marked loaded when the file is malformed")
	}
}

// TestHandoffRegistered verifies the audit registry binds "handoff" so the
// handoff.yaml section is not reported as an orphan (parity gate).
// AC-AUTORESUME-002 → REQ-002.
func TestHandoffRegistered(t *testing.T) {
	t.Parallel()

	if !IsRegisteredOrException("handoff") {
		t.Error("IsRegisteredOrException(\"handoff\") = false, want true (register in audit_registry.go)")
	}
	registry := GetYAMLToStructRegistry()
	if got := registry["handoff"]; got != "HandoffConfig" {
		t.Errorf("yamlToStructRegistry[\"handoff\"] = %q, want %q", got, "HandoffConfig")
	}
}

// TestHandoffStaleTTLDefault documents the auto-only stale TTL single-source
// default consumed by the M3 handler (7 days).
func TestHandoffStaleTTLDefault(t *testing.T) {
	t.Parallel()

	if DefaultHandoffStaleTTL <= 0 {
		t.Fatalf("DefaultHandoffStaleTTL must be positive, got %v", DefaultHandoffStaleTTL)
	}
	wantHours := 7 * 24.0
	if DefaultHandoffStaleTTL.Hours() != wantHours {
		t.Errorf("DefaultHandoffStaleTTL = %v (%.0fh), want %.0fh", DefaultHandoffStaleTTL, DefaultHandoffStaleTTL.Hours(), wantHours)
	}
}
