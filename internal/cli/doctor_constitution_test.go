package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// validConstitutionRegistryForDoctor는 doctor 테스트용 유효한 registry 내용이다.
// CLAUDE.md를 참조하므로 테스트 tempDir에 CLAUDE.md를 생성해야 orphan이 발생하지 않는다.
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

// duplicateIDRegistry는 중복 ID가 있는 registry를 반환한다.
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

// emptyFrozenRegistryForDoctor는 Frozen 엔트리가 없는 registry를 반환한다.
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

// TestCheckConstitution_ValidRegistry는 유효한 registry에서 OK를 반환함을 검증한다.
// AC-CON-001-001 관련.
func TestCheckConstitution_ValidRegistry(t *testing.T) {
	dir := t.TempDir()

	// CLAUDE.md 생성 (파일 존재 확인용)
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Test"), 0o600); err != nil {
		t.Fatalf("CLAUDE.md 생성 오류: %v", err)
	}

	registryPath := filepath.Join(dir, "zone-registry.md")
	if err := os.WriteFile(registryPath, []byte(validConstitutionRegistryForDoctor), 0o600); err != nil {
		t.Fatalf("registry 파일 생성 오류: %v", err)
	}

	check := checkConstitution(dir, registryPath, false, false)

	if check.Status != CheckOK {
		t.Errorf("유효한 registry에서 Status = %q, want %q\n메시지: %s", check.Status, CheckOK, check.Message)
	}
}

// TestCheckConstitution_RegistryMissing는 registry 파일 없음 시 Warn을 반환함을 검증한다.
func TestCheckConstitution_RegistryMissing(t *testing.T) {
	dir := t.TempDir()
	nonExistentPath := filepath.Join(dir, "nonexistent-registry.md")

	check := checkConstitution(dir, nonExistentPath, false, false)

	if check.Status != CheckWarn {
		t.Errorf("registry 없음 Status = %q, want %q", check.Status, CheckWarn)
	}
}

// TestCheckConstitution_DuplicateIDs는 중복 ID registry 시 Fail을 반환함을 검증한다.
func TestCheckConstitution_DuplicateIDs(t *testing.T) {
	dir := t.TempDir()

	// CLAUDE.md 생성
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Test"), 0o600); err != nil {
		t.Fatalf("CLAUDE.md 생성 오류: %v", err)
	}

	registryPath := filepath.Join(dir, "zone-registry.md")
	if err := os.WriteFile(registryPath, []byte(duplicateIDRegistryForDoctor), 0o600); err != nil {
		t.Fatalf("registry 파일 생성 오류: %v", err)
	}

	check := checkConstitution(dir, registryPath, false, false)

	if check.Status != CheckFail {
		t.Errorf("중복 ID registry Status = %q, want %q\n메시지: %s", check.Status, CheckFail, check.Message)
	}
}

// TestCheckConstitution_EmptyFrozen는 Frozen 엔트리 없음 시 Warn을 반환함을 검증한다.
// AC-CON-001-006 관련.
func TestCheckConstitution_EmptyFrozen(t *testing.T) {
	dir := t.TempDir()

	// CLAUDE.md 생성
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Test"), 0o600); err != nil {
		t.Fatalf("CLAUDE.md 생성 오류: %v", err)
	}

	registryPath := filepath.Join(dir, "zone-registry.md")
	if err := os.WriteFile(registryPath, []byte(emptyFrozenRegistryForDoctor), 0o600); err != nil {
		t.Fatalf("registry 파일 생성 오류: %v", err)
	}

	check := checkConstitution(dir, registryPath, false, false)

	if check.Status != CheckWarn {
		t.Errorf("empty Frozen Status = %q, want %q\n메시지: %s", check.Status, CheckWarn, check.Message)
	}

	if !strings.Contains(check.Message, "Frozen") {
		t.Errorf("경고 메시지에 'Frozen'이 포함되어야 한다: %s", check.Message)
	}
}

// TestCheckConstitution_StrictMode는 MOAI_CONSTITUTION_STRICT=1 시 orphan이 Fail로 처리됨을 검증한다.
// AC-CON-001-006 관련.
func TestCheckConstitution_StrictMode(t *testing.T) {
	dir := t.TempDir()
	// CLAUDE.md를 생성하지 않아 orphan 발생 (파일 미존재)

	registryPath := filepath.Join(dir, "zone-registry.md")
	// sampleRegistryContent uses CLAUDE.md which won't exist in dir
	if err := os.WriteFile(registryPath, []byte(sampleRegistryContent), 0o600); err != nil {
		t.Fatalf("registry 파일 생성 오류: %v", err)
	}

	// strictMode=true, orphan 발생 → Fail
	check := checkConstitution(dir, registryPath, false, true)

	if check.Status != CheckFail {
		t.Errorf("strict mode + orphan Status = %q, want %q\n메시지: %s", check.Status, CheckFail, check.Message)
	}
}
