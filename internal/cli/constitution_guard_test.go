package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupGuardRegistry는 guard 테스트용 임시 registry를 생성한다.
// projectDir 아래에 CLAUDE.md와 zone-registry.md를 생성한다.
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

// TestConstitutionGuard_NoViolations는 Frozen zone 변경이 없는 경우 OK를 반환함을 검증한다.
// AC-CON-001-003 관련.
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

// TestConstitutionGuard_RegistryMissing은 registry 없음 시 에러를 반환함을 검증한다.
func TestConstitutionGuard_RegistryMissing(t *testing.T) {
	dir := t.TempDir()
	nonExistentPath := filepath.Join(dir, "nonexistent.md")

	var buf bytes.Buffer
	result := runConstitutionGuard(&buf, io.Discard, dir, nonExistentPath, []string{})
	if result == nil {
		t.Fatal("registry 없음 시 에러를 반환해야 한다")
	}
}

// TestConstitutionGuard_DetectsFrozenViolation은 Frozen zone 변경 탐지를 검증한다.
// AC-CON-001-003 직접 매핑.
func TestConstitutionGuard_DetectsFrozenViolation(t *testing.T) {
	dir, regPath := setupGuardRegistry(t, validConstitutionRegistryForDoctor)

	// 위반 목록: Frozen 엔트리 ID를 포함
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

// TestConstitutionGuard_EvolvableViolationNotFatal은 Evolvable zone 변경이 에러가 아님을 검증한다.
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

// TestConstitutionGuard_SubcommandExists는 guard 서브커맨드가 등록되어 있음을 검증한다.
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
