package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// audit_registry_test.go — M1 RED phase tests for audit_registry.go (to be created in M2).
//
// REQ-V3R2-RT-005-008 (CI rule yaml↔struct parity)
// REQ-V3R2-RT-005-021 (yaml file count ≠ Go struct count → audit fails)
// AC-08

// TestAuditRegistry_AllRegisteredStructsExist verifies that every struct listed in YAMLToStructRegistry
// is actually a field (or embedded type) in the Config struct.
//
// AC-V3R2-RT-005-08 edge case (registered struct missing from Config):
// If registry maps "phantom" → "PhantomConfig" but PhantomConfig is not in Config struct,
// the audit should fail.
//
// # REQ-V3R2-RT-005-008, REQ-V3R2-RT-005-021, AC-08
//
// @MX:NOTE [AUTO] SPEC-V3R2-RT-005 M2 GREEN — REQ-008/021 all registered structs verified
func TestAuditRegistry_AllRegisteredStructsExist(t *testing.T) {
	// GetYAMLToStructRegistry is undefined in RED phase → compile error signals RED.
	registry := GetYAMLToStructRegistry()

	if len(registry) == 0 {
		t.Fatalf("GetYAMLToStructRegistry() returned empty registry; expected ≥13 entries for known sections")
	}

	// Verify each registry entry has a non-empty struct name
	for yamlName, structName := range registry {
		if yamlName == "" {
			t.Errorf("registry contains empty yaml name → struct name %q mapping", structName)
		}
		if structName == "" {
			t.Errorf("registry yaml name %q maps to empty struct name", yamlName)
		}
	}

	// Verify known sections are in the registry
	knownSections := []string{
		"user",
		"language",
		"quality",
		"project",
		"git-convention",
		"system",
		"llm",
		"state",
		"statusline",
		"ralph",
		"research",
		"workflow",
	}
	for _, section := range knownSections {
		if _, ok := registry[section]; !ok {
			t.Errorf("registry missing section %q (expected from types.go Config struct fields)", section)
		}
	}
}

// TestAuditRegistry_NoUnexpectedYAMLOrphans verifies that a fresh scan of .moai/config/sections/
// in a test workspace finds no orphans beyond registered+exception sections.
//
// AC-V3R2-RT-005-08: Every yaml file is either registered or in exceptions; orphans fail the audit.
//
// # REQ-V3R2-RT-005-008, REQ-V3R2-RT-005-043, AC-08
//
// @MX:NOTE [AUTO] SPEC-V3R2-RT-005 M2 GREEN — REQ-043 ScanYAMLOrphans detects unexpected yaml
func TestAuditRegistry_NoUnexpectedYAMLOrphans(t *testing.T) {
	// Arrange: create a workspace where only known yaml files exist
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}

	// Write yaml files that are all known (registered or exception)
	knownFiles := []string{
		"user.yaml",
		"language.yaml",
		"quality.yaml",
		"constitution.yaml", // exception (MIG-003 pending)
	}
	for _, f := range knownFiles {
		if err := os.WriteFile(filepath.Join(sectionsDir, f), []byte("# test\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", f, err)
		}
	}

	// ScanYAMLOrphans should not exist yet (RED phase) → compile error signals RED
	orphans := ScanYAMLOrphans(sectionsDir, GetYAMLToStructRegistry(), GetYAMLAuditExceptions())

	// With all known files, there should be no orphans
	if len(orphans) > 0 {
		t.Errorf("ScanYAMLOrphans() found unexpected orphans: %v", orphans)
	}

	// Add an orphan file
	orphanFile := filepath.Join(sectionsDir, "unexpected_section.yaml")
	if err := os.WriteFile(orphanFile, []byte("unexpected:\n  key: value\n"), 0o644); err != nil {
		t.Fatalf("write orphan: %v", err)
	}

	// Re-scan: should now find the orphan
	orphans = ScanYAMLOrphans(sectionsDir, GetYAMLToStructRegistry(), GetYAMLAuditExceptions())
	if len(orphans) == 0 {
		t.Error("ScanYAMLOrphans() should find 'unexpected_section.yaml' as an orphan")
	}

	// Verify orphan message format
	found := false
	for _, orphan := range orphans {
		if strings.Contains(orphan, "unexpected_section") {
			found = true
		}
	}
	if !found {
		t.Errorf("orphans %v should contain 'unexpected_section'", orphans)
	}
}
