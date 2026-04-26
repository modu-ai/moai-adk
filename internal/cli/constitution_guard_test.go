package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupGuardRegistry creates a temporary registry for guard tests.
// Creates CLAUDE.md and zone-registry.md under projectDir.
func setupGuardRegistry(t *testing.T, content string) (projectDir, registryPath string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Test"), 0o600); err != nil {
		t.Fatalf("failed to create CLAUDE.md: %v", err)
	}
	regPath := filepath.Join(dir, "zone-registry.md")
	if err := os.WriteFile(regPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create registry file: %v", err)
	}
	return dir, regPath
}

// TestConstitutionGuard_NoViolations verifies that OK is returned when there are no Frozen zone changes.
// Related to AC-CON-001-003.
func TestConstitutionGuard_NoViolations(t *testing.T) {
	dir, regPath := setupGuardRegistry(t, validConstitutionRegistryForDoctor)

	var buf bytes.Buffer
	result := runConstitutionGuard(&buf, io.Discard, dir, regPath, []string{})
	if result != nil {
		t.Errorf("expected nil when no violations, got: %v", result)
	}

	output := buf.String()
	if !strings.Contains(output, "OK") && !strings.Contains(output, "no violations") && !strings.Contains(output, "violation") {
		t.Logf("output: %s", output)
	}
}

// TestConstitutionGuard_RegistryMissing verifies that an error is returned when the registry is absent.
func TestConstitutionGuard_RegistryMissing(t *testing.T) {
	dir := t.TempDir()
	nonExistentPath := filepath.Join(dir, "nonexistent.md")

	var buf bytes.Buffer
	result := runConstitutionGuard(&buf, io.Discard, dir, nonExistentPath, []string{})
	if result == nil {
		t.Fatal("must return an error when registry is missing")
	}
}

// TestConstitutionGuard_DetectsFrozenViolation verifies detection of Frozen zone changes.
// Direct mapping to AC-CON-001-003.
func TestConstitutionGuard_DetectsFrozenViolation(t *testing.T) {
	dir, regPath := setupGuardRegistry(t, validConstitutionRegistryForDoctor)

	// Violation list: includes a Frozen entry ID
	violations := []string{"CONST-V3R2-001"}
	var buf bytes.Buffer
	result := runConstitutionGuard(&buf, io.Discard, dir, regPath, violations)

	if result == nil {
		t.Fatal("must return an error when a Frozen violation is detected")
	}
	if !strings.Contains(result.Error(), "CONST-V3R2-001") {
		t.Errorf("error message must contain the violation ID: %v", result)
	}
}

// TestConstitutionGuard_EvolvableViolationNotFatal verifies that Evolvable zone changes are not errors.
func TestConstitutionGuard_EvolvableViolationNotFatal(t *testing.T) {
	dir, regPath := setupGuardRegistry(t, validConstitutionRegistryForDoctor)

	// CONST-V3R2-002 is Evolvable in validConstitutionRegistryForDoctor
	violations := []string{"CONST-V3R2-002"}
	var buf bytes.Buffer
	result := runConstitutionGuard(&buf, io.Discard, dir, regPath, violations)

	if result != nil {
		t.Errorf("Evolvable zone changes must not be errors: %v", result)
	}
}

// TestConstitutionGuard_SubcommandExists verifies that the guard subcommand is registered.
func TestConstitutionGuard_SubcommandExists(t *testing.T) {
	constitutionCmd := newConstitutionCmd()
	var found bool
	for _, sub := range constitutionCmd.Commands() {
		if sub.Use == "guard" {
			found = true
			break
		}
	}
	if !found {
		t.Error("constitution guard subcommand must be registered")
	}
}
