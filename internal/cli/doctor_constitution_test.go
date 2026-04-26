package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// validConstitutionRegistryForDoctor is valid registry content for doctor tests.
// Since it references CLAUDE.md, CLAUDE.md must be created in the test tempDir to avoid orphans.
const validConstitutionRegistryForDoctor = `# Test Registry

## Entries

` + "```yaml" + `
- id: CONST-V3R2-001
  zone: Frozen
  file: CLAUDE.md
  anchor: "#1-core-identity"
  clause: "SPEC+EARS format"
  canary_gate: true

- id: CONST-V3R2-002
  zone: Evolvable
  file: CLAUDE.md
  anchor: "#quality-gates"
  clause: "TRUST 5"
  canary_gate: false
` + "```" + `
`

// duplicateIDRegistryForDoctor is a registry with duplicate IDs.
const duplicateIDRegistryForDoctor = `# Test Registry

## Entries

` + "```yaml" + `
- id: CONST-V3R2-001
  zone: Frozen
  file: CLAUDE.md
  anchor: "#1-core-identity"
  clause: "SPEC+EARS format"
  canary_gate: true

- id: CONST-V3R2-001
  zone: Frozen
  file: CLAUDE.md
  anchor: "#trust-5"
  clause: "TRUST 5"
  canary_gate: true
` + "```" + `
`

// emptyFrozenRegistryForDoctor is a registry with no Frozen entries.
const emptyFrozenRegistryForDoctor = `# Test Registry

## Entries

` + "```yaml" + `
- id: CONST-V3R2-001
  zone: Evolvable
  file: CLAUDE.md
  anchor: "#1-core-identity"
  clause: "Some evolvable rule"
  canary_gate: false
` + "```" + `
`

// TestCheckConstitution_ValidRegistry verifies that OK is returned for a valid registry.
// Related to AC-CON-001-001.
func TestCheckConstitution_ValidRegistry(t *testing.T) {
	dir := t.TempDir()

	// Create CLAUDE.md (for file existence check)
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Test"), 0o600); err != nil {
		t.Fatalf("failed to create CLAUDE.md: %v", err)
	}

	registryPath := filepath.Join(dir, "zone-registry.md")
	if err := os.WriteFile(registryPath, []byte(validConstitutionRegistryForDoctor), 0o600); err != nil {
		t.Fatalf("failed to create registry file: %v", err)
	}

	check := checkConstitution(dir, registryPath, false, false)

	if check.Status != CheckOK {
		t.Errorf("valid registry Status = %q, want %q\nmessage: %s", check.Status, CheckOK, check.Message)
	}
}

// TestCheckConstitution_RegistryMissing verifies that Warn is returned when the registry file is absent.
func TestCheckConstitution_RegistryMissing(t *testing.T) {
	dir := t.TempDir()
	nonExistentPath := filepath.Join(dir, "nonexistent-registry.md")

	check := checkConstitution(dir, nonExistentPath, false, false)

	if check.Status != CheckWarn {
		t.Errorf("missing registry Status = %q, want %q", check.Status, CheckWarn)
	}
}

// TestCheckConstitution_DuplicateIDs verifies that Fail is returned for a registry with duplicate IDs.
func TestCheckConstitution_DuplicateIDs(t *testing.T) {
	dir := t.TempDir()

	// Create CLAUDE.md
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Test"), 0o600); err != nil {
		t.Fatalf("failed to create CLAUDE.md: %v", err)
	}

	registryPath := filepath.Join(dir, "zone-registry.md")
	if err := os.WriteFile(registryPath, []byte(duplicateIDRegistryForDoctor), 0o600); err != nil {
		t.Fatalf("failed to create registry file: %v", err)
	}

	check := checkConstitution(dir, registryPath, false, false)

	if check.Status != CheckFail {
		t.Errorf("duplicate ID registry Status = %q, want %q\nmessage: %s", check.Status, CheckFail, check.Message)
	}
}

// TestCheckConstitution_EmptyFrozen verifies that Warn is returned when there are no Frozen entries.
// Related to AC-CON-001-006.
func TestCheckConstitution_EmptyFrozen(t *testing.T) {
	dir := t.TempDir()

	// Create CLAUDE.md
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Test"), 0o600); err != nil {
		t.Fatalf("failed to create CLAUDE.md: %v", err)
	}

	registryPath := filepath.Join(dir, "zone-registry.md")
	if err := os.WriteFile(registryPath, []byte(emptyFrozenRegistryForDoctor), 0o600); err != nil {
		t.Fatalf("failed to create registry file: %v", err)
	}

	check := checkConstitution(dir, registryPath, false, false)

	if check.Status != CheckWarn {
		t.Errorf("empty Frozen Status = %q, want %q\nmessage: %s", check.Status, CheckWarn, check.Message)
	}

	if !strings.Contains(check.Message, "Frozen") {
		t.Errorf("warning message must contain 'Frozen': %s", check.Message)
	}
}

// TestCheckConstitution_StrictMode verifies that orphans are treated as Fail when MOAI_CONSTITUTION_STRICT=1.
// Related to AC-CON-001-006.
func TestCheckConstitution_StrictMode(t *testing.T) {
	dir := t.TempDir()
	// Do not create CLAUDE.md so that an orphan occurs (file not found)

	registryPath := filepath.Join(dir, "zone-registry.md")
	// sampleRegistryContent uses CLAUDE.md which won't exist in dir
	if err := os.WriteFile(registryPath, []byte(sampleRegistryContent), 0o600); err != nil {
		t.Fatalf("failed to create registry file: %v", err)
	}

	// strictMode=true, orphan present → Fail
	check := checkConstitution(dir, registryPath, false, true)

	if check.Status != CheckFail {
		t.Errorf("strict mode + orphan Status = %q, want %q\nmessage: %s", check.Status, CheckFail, check.Message)
	}
}
