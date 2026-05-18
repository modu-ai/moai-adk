package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestLoadDesignConfig_ValidParse verifies that a valid design.yaml is parsed
// with gan_loop fields and sprint_contract populated.
// Maps to: REQ-MIG003-001/002/003/007/014, AC-MIG003-02/07/11
func TestLoadDesignConfig_ValidParse(t *testing.T) {
	path := filepath.Join("testdata", "design-valid", "design.yaml")
	cfg, err := LoadDesignConfig(path)
	if err != nil {
		t.Fatalf("LoadDesignConfig(%q): unexpected error: %v", path, err)
	}
	if cfg == nil {
		t.Fatal("LoadDesignConfig returned nil config")
	}

	// REQ-MIG003-014: gan_loop.pass_threshold = 0.75
	if cfg.GanLoop.PassThreshold != 0.75 {
		t.Errorf("GanLoop.PassThreshold: want 0.75, got %f", cfg.GanLoop.PassThreshold)
	}

	// REQ-MIG003-014: gan_loop.max_iterations = 5
	if cfg.GanLoop.MaxIterations != 5 {
		t.Errorf("GanLoop.MaxIterations: want 5, got %d", cfg.GanLoop.MaxIterations)
	}

	// REQ-MIG003-014: sprint_contract.enabled = true
	if !cfg.GanLoop.SprintContract.Enabled {
		t.Error("GanLoop.SprintContract.Enabled: want true, got false")
	}
}

// TestLoadDesignConfig_MissingFile_ReturnsDefaults verifies that an absent
// design.yaml returns defaults with no error.
// Maps to: REQ-MIG003-004, AC-MIG003-03
func TestLoadDesignConfig_MissingFile_ReturnsDefaults(t *testing.T) {
	path := filepath.Join("testdata", "nonexistent-dir", "design.yaml")
	cfg, err := LoadDesignConfig(path)
	if err != nil {
		t.Fatalf("LoadDesignConfig with absent file: expected no error, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadDesignConfig returned nil config on missing file")
	}
}

// TestLoadDesignConfig_Malformed_ReturnsErrInvalidYAML verifies that malformed
// design.yaml returns an error wrapping ErrInvalidYAML.
// Maps to: REQ-MIG003-003/007, AC-MIG003-04/07
func TestLoadDesignConfig_Malformed_ReturnsErrInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	malformedPath := filepath.Join(tmpDir, "design.yaml")
	if err := os.WriteFile(malformedPath, []byte("design:\n  gan_loop: [\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, err := LoadDesignConfig(malformedPath)
	if err == nil {
		t.Fatal("expected error for malformed YAML, got nil")
	}
	if !errors.Is(err, ErrInvalidYAML) {
		t.Errorf("expected errors.Is(err, ErrInvalidYAML), got: %v", err)
	}
}

// TestLoadDesignConfig_PassThresholdFloorViolation verifies that a pass_threshold
// below the FROZEN floor 0.60 returns ErrPassThresholdFloor.
// Maps to: REQ-MIG003-014 (OQ2 decision: YES), AC-MIG003-11
func TestLoadDesignConfig_PassThresholdFloorViolation(t *testing.T) {
	path := filepath.Join("testdata", "design-pass-threshold-violation", "design.yaml")
	_, err := LoadDesignConfig(path)
	if err == nil {
		t.Fatal("expected error for pass_threshold < 0.60, got nil")
	}
	if !errors.Is(err, ErrPassThresholdFloor) {
		t.Errorf("expected errors.Is(err, ErrPassThresholdFloor), got: %v", err)
	}
}
