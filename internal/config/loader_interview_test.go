package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestLoadInterviewConfig_ValidParse verifies that a valid interview.yaml is parsed
// with clarity_threshold, plan.max_rounds, and skip_conditions populated.
// Maps to: REQ-MIG003-001/002/003/007/011, AC-MIG003-02/07/10
func TestLoadInterviewConfig_ValidParse(t *testing.T) {
	path := filepath.Join("testdata", "interview-valid", "interview.yaml")
	cfg, err := LoadInterviewConfig(path)
	if err != nil {
		t.Fatalf("LoadInterviewConfig(%q): unexpected error: %v", path, err)
	}
	if cfg == nil {
		t.Fatal("LoadInterviewConfig returned nil config")
	}

	// REQ-MIG003-011: clarity_threshold = 4
	if cfg.ClarityThreshold != 4 {
		t.Errorf("ClarityThreshold: want 4, got %d", cfg.ClarityThreshold)
	}

	// REQ-MIG003-011: plan.max_rounds = 5
	if cfg.Plan.MaxRounds != 5 {
		t.Errorf("Plan.MaxRounds: want 5, got %d", cfg.Plan.MaxRounds)
	}

	// REQ-MIG003-011: skip_conditions length = 3
	if len(cfg.SkipConditions) != 3 {
		t.Errorf("SkipConditions: want len=3, got %d", len(cfg.SkipConditions))
	}
}

// TestLoadInterviewConfig_MissingFile_ReturnsDefaults verifies that an absent
// interview.yaml returns defaults with no error.
// Maps to: REQ-MIG003-004, AC-MIG003-03
func TestLoadInterviewConfig_MissingFile_ReturnsDefaults(t *testing.T) {
	path := filepath.Join("testdata", "interview-missing", "interview.yaml")
	cfg, err := LoadInterviewConfig(path)
	if err != nil {
		t.Fatalf("LoadInterviewConfig with absent file: expected no error, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadInterviewConfig returned nil config on missing file")
	}
}

// TestLoadInterviewConfig_Malformed_ReturnsErrInvalidYAML verifies that a
// missing/malformed interview.yaml from a malformed fixture returns ErrInvalidYAML.
// Maps to: REQ-MIG003-003/007, AC-MIG003-04/07
func TestLoadInterviewConfig_Malformed_ReturnsErrInvalidYAML(t *testing.T) {
	// Write a temporary malformed file for this test
	tmpDir := t.TempDir()
	malformedPath := filepath.Join(tmpDir, "interview.yaml")
	if err := os.WriteFile(malformedPath, []byte("interview:\n  clarity_threshold: [\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, err := LoadInterviewConfig(malformedPath)
	if err == nil {
		t.Fatal("expected error for malformed YAML, got nil")
	}
	if !errors.Is(err, ErrInvalidYAML) {
		t.Errorf("expected errors.Is(err, ErrInvalidYAML), got: %v", err)
	}
}
