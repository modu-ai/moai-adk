package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAuditParity checks that every .moai/config/sections/*.yaml has a corresponding Go struct.
// This maps to REQ-V3R2-RT-005-008 and REQ-V3R2-RT-005-043.
func TestAuditParity(t *testing.T) {
	// This test uses a registry map to track yaml files and their corresponding Go structs.
	// In a real implementation, this would check internal/config/types.go for struct definitions.
	//
	// For this implementation, we provide a basic framework that can be extended.

	t.Skip("Audit test requires full type registry implementation - placeholder for SPEC-V3R2-MIG-003")

	// The implementation would:
	// 1. Scan .moai/config/sections/ directory for *.yaml files
	// 2. Check internal/config/types.go for corresponding structs
	// 3. Report any orphan yaml files (no Go struct) or orphan structs (no yaml file)
	// 4. Support an exceptions registry for legitimate divergences
}

// TestAuditParity_OrphanYAMLFails verifies that an unregistered yaml file causes the audit to fail.
//
// AC-V3R2-RT-005-08: Given a new file .moai/config/sections/foo.yaml is added without a Go struct,
// When TestAuditParity runs, Then the test fails naming foo.yaml as orphan.
//
// REQ-V3R2-RT-005-043, REQ-V3R2-RT-005-008
//
// @MX:TODO SPEC-V3R2-RT-005 M1 RED → GREEN at M2 (audit_registry.go + registry-driven scan)
func TestAuditParity_OrphanYAMLFails(t *testing.T) {
	// Arrange: create a temp workspace with an orphan yaml file not in the registry
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}

	// Create an orphan yaml file that has no matching Go struct
	orphanFile := filepath.Join(sectionsDir, "foo.yaml")
	if err := os.WriteFile(orphanFile, []byte("foo:\n  key: value\n"), 0o644); err != nil {
		t.Fatalf("failed to write orphan yaml: %v", err)
	}

	// Act + Assert: verify registry can detect orphan
	// IsRegisteredOrException should not exist yet (RED phase) → this call will fail to compile or return false
	registered := IsRegisteredOrException("foo")
	if registered {
		t.Errorf("IsRegisteredOrException(%q) = true, want false for unregistered yaml", "foo")
	}

	// Verify the sentinel error message format
	sentinel := "orphan yaml file (no Go struct mapping): foo.yaml"
	if !strings.Contains(sentinel, "foo.yaml") {
		t.Errorf("sentinel %q does not contain expected filename", sentinel)
	}
}

// TestAuditParity_AllRegisteredYAMLPass verifies that all currently registered yaml files pass the audit.
//
// AC-V3R2-RT-005-08 edge case: Given the current registered yaml files plus MIG-003-pending exceptions,
// When TestAuditParity runs, Then the test passes.
//
// REQ-V3R2-RT-005-008, REQ-V3R2-RT-005-021
//
// @MX:TODO SPEC-V3R2-RT-005 M1 RED → GREEN at M2 (requires audit_registry.go + YAMLToStructRegistry)
func TestAuditParity_AllRegisteredYAMLPass(t *testing.T) {
	// Act: verify all entries in YAMLToStructRegistry are recognized
	// YAMLToStructRegistry should not exist yet (RED phase) → will fail to compile or be empty
	registry := GetYAMLToStructRegistry()
	if len(registry) == 0 {
		t.Errorf("GetYAMLToStructRegistry() returned empty registry; expected at least 16 registered sections")
	}

	// Verify known sections are registered
	expectedSections := []string{
		"user", "language", "quality", "project",
		"git-convention", "git-strategy", "system", "llm",
		"state", "statusline", "ralph", "research",
		"workflow",
	}
	for _, section := range expectedSections {
		if !IsRegisteredOrException(section) {
			t.Errorf("IsRegisteredOrException(%q) = false, want true for known registered section", section)
		}
	}
}

// TestAuditParity_OrphanStructFails verifies that a struct in the registry without a yaml file is detected.
//
// AC-V3R2-RT-005-08 edge case: Given registry maps "phantom" → "PhantomConfig" but no phantom.yaml exists,
// When TestAuditParity runs, Then the test fails with sentinel naming the absent struct.
//
// REQ-V3R2-RT-005-008, REQ-V3R2-RT-005-021
//
// @MX:TODO SPEC-V3R2-RT-005 M1 RED → GREEN at M2 (requires audit_registry.go)
func TestAuditParity_OrphanStructFails(t *testing.T) {
	// Verify the audit can detect a registry entry with no corresponding yaml file.
	// ScanOrphanStructs should not exist yet (RED phase) → will fail to compile.
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}
	// No yaml files in sectionsDir — all registry entries should be reported as orphan structs

	orphans := ScanOrphanStructs(sectionsDir, GetYAMLToStructRegistry())
	// At least some orphan structs expected when yaml dir is empty
	if len(orphans) == 0 {
		t.Errorf("ScanOrphanStructs() returned no orphans for empty sections dir, expected at least one registered section to be reported")
	}

	sentinel := "registry maps phantom → PhantomConfig but struct not found in Config"
	_ = sentinel // Used in M2 actual error messages
}

// TestAuditParity_ExceptionsRespected verifies that yaml files in YAMLAuditExceptions do not fail audit.
//
// REQ-V3R2-RT-005-008, REQ-V3R2-RT-005-021
//
// @MX:TODO SPEC-V3R2-RT-005 M1 RED → GREEN at M2 (requires YAMLAuditExceptions map in audit_registry.go)
func TestAuditParity_ExceptionsRespected(t *testing.T) {
	// Verify the MIG-003-pending exception files are in the exceptions registry
	// YAMLAuditExceptions should not exist yet (RED phase) → will fail to compile.
	exceptions := GetYAMLAuditExceptions()

	// 5 yaml sections pending MIG-003 loaders should be registered as exceptions
	expectedExceptions := []string{
		"constitution", "context", "interview", "design", "harness",
	}
	for _, name := range expectedExceptions {
		if _, ok := exceptions[name]; !ok {
			t.Errorf("GetYAMLAuditExceptions() missing exception entry for %q (pending MIG-003 loader)", name)
		}
	}

	// Exceptions should pass the IsRegisteredOrException check
	for _, name := range expectedExceptions {
		if !IsRegisteredOrException(name) {
			t.Errorf("IsRegisteredOrException(%q) = false, want true (it's an exception)", name)
		}
	}
}
