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
// Generates CLAUDE.md and zone-registry.md under projectDir.
func setupGuardRegistry(t *testing.T, content string) (projectDir, registryPath string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Test"), 0o600); err != nil {
		t.Fatalf("CLAUDE.md 생성 오류: %v", err)
	}
	regPath := filepath.Join(dir, "zone-registry.md")
	if err := os.WriteFile(regPath, []byte(content), 0o600); err != nil {
		t.Fatalf("registry 파일 생성 오류: %v", err)
	}
	return dir, regPath
}

// TestConstitutionGuard_NoViolations verifies that the guard returns OK when no Frozen zone changes occur.
// Relates to AC-CON-001-003.
func TestConstitutionGuard_NoViolations(t *testing.T) {
	dir, regPath := setupGuardRegistry(t, validConstitutionRegistryForDoctor)

	var buf bytes.Buffer
	result := runConstitutionGuard(&buf, io.Discard, dir, regPath, []string{})
	if result != nil {
		t.Errorf("위반 없음 시 nil 반환 기대, got: %v", result)
	}

	output := buf.String()
	if !strings.Contains(output, "OK") && !strings.Contains(output, "no violations") && !strings.Contains(output, "위반") {
		t.Logf("출력: %s", output)
	}
}

// TestConstitutionGuard_RegistryMissing verifies that a missing registry returns an error.
func TestConstitutionGuard_RegistryMissing(t *testing.T) {
	dir := t.TempDir()
	nonExistentPath := filepath.Join(dir, "nonexistent.md")

	var buf bytes.Buffer
	result := runConstitutionGuard(&buf, io.Discard, dir, nonExistentPath, []string{})
	if result == nil {
		t.Fatal("registry 없음 시 에러를 반환해야 한다")
	}
}

// TestConstitutionGuard_DetectsFrozenViolation verifies detection of Frozen zone changes.
// Maps directly to AC-CON-001-003.
func TestConstitutionGuard_DetectsFrozenViolation(t *testing.T) {
	dir, regPath := setupGuardRegistry(t, validConstitutionRegistryForDoctor)

	// Violation list: includes a Frozen entry ID
	violations := []string{"CONST-V3R2-001"}
	var buf bytes.Buffer
	result := runConstitutionGuard(&buf, io.Discard, dir, regPath, violations)

	if result == nil {
		t.Fatal("Frozen 위반 탐지 시 에러를 반환해야 한다")
	}
	if !strings.Contains(result.Error(), "CONST-V3R2-001") {
		t.Errorf("에러 메시지에 위반 ID가 포함되어야 한다: %v", result)
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
		t.Errorf("Evolvable zone 변경은 에러가 아니어야 한다: %v", result)
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
		t.Error("constitution guard 서브커맨드가 등록되어 있어야 한다")
	}
}
