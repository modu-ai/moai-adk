package config

import (
	"errors"
	"path/filepath"
	"testing"
)

// TestLoadConstitutionConfig_ValidParse verifies that a valid constitution.yaml
// is correctly parsed into ConstitutionConfig with all sub-fields populated.
// Maps to: REQ-MIG003-001/002/003/007/009, AC-MIG003-02/07/08
func TestLoadConstitutionConfig_ValidParse(t *testing.T) {
	path := filepath.Join("testdata", "constitution-valid", "constitution.yaml")
	cfg, err := LoadConstitutionConfig(path)
	if err != nil {
		t.Fatalf("LoadConstitutionConfig(%q): unexpected error: %v", path, err)
	}
	if cfg == nil {
		t.Fatal("LoadConstitutionConfig returned nil config")
	}

	// ApprovedLanguages
	if len(cfg.ApprovedLanguages) == 0 {
		t.Error("expected non-empty ApprovedLanguages")
	}

	// ForbiddenPatterns (aka ForbiddenLibraries per REQ-MIG003-009)
	if len(cfg.ForbiddenPatterns) == 0 {
		t.Error("expected non-empty ForbiddenPatterns (forbidden_patterns)")
	}

	// Architecture sub-fields
	if len(cfg.Architecture.ForbiddenDependencies) == 0 {
		t.Error("expected non-empty Architecture.ForbiddenDependencies")
	}

	// NamingConventions
	if cfg.NamingConventions.Exported == "" {
		t.Error("expected non-empty NamingConventions.Exported")
	}

	// Security
	if len(cfg.Security.ForbiddenPractices) == 0 {
		t.Error("expected non-empty Security.ForbiddenPractices")
	}
}

// TestLoadConstitutionConfig_MissingFile_ReturnsDefaults verifies that when the
// constitution.yaml file is absent, sensible defaults are returned (no error).
// Maps to: REQ-MIG003-004, AC-MIG003-03
func TestLoadConstitutionConfig_MissingFile_ReturnsDefaults(t *testing.T) {
	path := filepath.Join("testdata", "nonexistent-dir", "constitution.yaml")
	cfg, err := LoadConstitutionConfig(path)
	if err != nil {
		t.Fatalf("LoadConstitutionConfig with absent file: expected no error, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadConstitutionConfig returned nil config on missing file")
	}
}

// TestLoadConstitutionConfig_Malformed_ReturnsErrInvalidYAML verifies that malformed
// YAML returns an error wrapping ErrInvalidYAML.
// Maps to: REQ-MIG003-003/007/008, AC-MIG003-04/07
func TestLoadConstitutionConfig_Malformed_ReturnsErrInvalidYAML(t *testing.T) {
	path := filepath.Join("testdata", "constitution-malformed", "constitution.yaml")
	_, err := LoadConstitutionConfig(path)
	if err == nil {
		t.Fatal("expected error for malformed YAML, got nil")
	}
	if !errors.Is(err, ErrInvalidYAML) {
		t.Errorf("expected errors.Is(err, ErrInvalidYAML), got: %v", err)
	}
}

// TestLoadConstitutionConfig_ForbiddenLibrariesExposed verifies that the
// ForbiddenPatterns (ForbiddenLibraries alias per REQ-MIG003-009) field is
// exposed as a non-empty list when the YAML contains entries.
// Maps to: REQ-MIG003-009, AC-MIG003-08
func TestLoadConstitutionConfig_ForbiddenLibrariesExposed(t *testing.T) {
	path := filepath.Join("testdata", "constitution-valid", "constitution.yaml")
	cfg, err := LoadConstitutionConfig(path)
	if err != nil {
		t.Fatalf("LoadConstitutionConfig: %v", err)
	}
	// ForbiddenPatterns is the field aliased as "ForbiddenLibraries" in godoc.
	if len(cfg.ForbiddenPatterns) == 0 {
		t.Error("REQ-MIG003-009: ForbiddenPatterns (forbidden_libraries alias) is empty; expected non-empty list for policy enforcement")
	}
	// Verify the list contains at least one meaningful entry.
	found := false
	for _, p := range cfg.ForbiddenPatterns {
		if p != "" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ForbiddenPatterns contains only empty strings")
	}
}
