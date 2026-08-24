package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/constitution"
)

// validConstitutionRegistryForDoctor is a registry content valid for doctor tests.
// It references CLAUDE.md, so the test tempDir must contain CLAUDE.md to avoid an orphan.
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

// duplicateIDRegistry returns a registry containing duplicate IDs.
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

// emptyFrozenRegistryForDoctor returns a registry without any Frozen entries.
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

// claudeMDWithClauses carries both clauses referenced by
// validConstitutionRegistryForDoctor, so the registry validates drift-free.
// A CLAUDE.md missing them is the drift fixture (see
// TestCheckConstitution_DriftFailsLikeValidate).
const claudeMDWithClauses = "# Test\n\nSPEC+EARS format is the requirement form.\n\nTRUST 5 is the quality framework.\n"

// TestCheckConstitution_ValidRegistry verifies that a valid registry returns OK.
// Relates to AC-CON-001-001.
func TestCheckConstitution_ValidRegistry(t *testing.T) {
	dir := t.TempDir()

	// Create CLAUDE.md (used for file-presence verification)
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(claudeMDWithClauses), 0o600); err != nil {
		t.Fatalf("CLAUDE.md 생성 오류: %v", err)
	}

	registryPath := filepath.Join(dir, "zone-registry.md")
	if err := os.WriteFile(registryPath, []byte(validConstitutionRegistryForDoctor), 0o600); err != nil {
		t.Fatalf("registry 파일 생성 오류: %v", err)
	}

	check := checkConstitution(dir, registryPath, false, false)

	if check.Status != uikit.CheckOK {
		t.Errorf("유효한 registry에서 Status = %q, want %q\n메시지: %s", check.Status, uikit.CheckOK, check.Message)
	}
}

// TestCheckConstitution_DriftFailsLikeValidate verifies #1594: doctor must not
// report a registry as OK on entry count alone while `moai constitution
// validate` reports errors against the same checkout. The registry here loads
// cleanly (structure is sound) but its clauses are absent from the source file,
// which is exactly what validate flags as DRIFT.
func TestCheckConstitution_DriftFailsLikeValidate(t *testing.T) {
	dir := t.TempDir()

	// CLAUDE.md exists (so the entries are not orphans) but carries neither
	// clause — the drift condition.
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Test\n"), 0o600); err != nil {
		t.Fatalf("CLAUDE.md 생성 오류: %v", err)
	}

	registryPath := filepath.Join(dir, "zone-registry.md")
	if err := os.WriteFile(registryPath, []byte(validConstitutionRegistryForDoctor), 0o600); err != nil {
		t.Fatalf("registry 파일 생성 오류: %v", err)
	}

	// Cross-check: the dedicated validate command sees the same drift.
	result, err := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: registryPath,
		ProjectDir:   dir,
	})
	if err != nil {
		t.Fatalf("constitution.Validate 오류: %v", err)
	}
	if result.Status != constitution.ValidateStatusDrift {
		t.Fatalf("fixture 전제 실패: validate Status = %q, want %q", result.Status, constitution.ValidateStatusDrift)
	}

	check := checkConstitution(dir, registryPath, true, false)

	if check.Status != uikit.CheckFail {
		t.Errorf("drift registry Status = %q, want %q\n메시지: %s", check.Status, uikit.CheckFail, check.Message)
	}
	if !strings.Contains(check.Message, "validate") {
		t.Errorf("메시지가 validate를 가리켜야 한다: %s", check.Message)
	}
	if check.Detail == "" {
		t.Error("verbose 모드에서 drift 항목 detail이 비어 있으면 안 된다")
	}
}

// TestCheckConstitution_SkipValidateBypass verifies that
// MOAI_CONSTITUTION_SKIP_VALIDATE=1 leaves doctor on its structural verdict
// rather than turning the bypass into a Fail.
func TestCheckConstitution_SkipValidateBypass(t *testing.T) {
	t.Setenv("MOAI_CONSTITUTION_SKIP_VALIDATE", "1")

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Test\n"), 0o600); err != nil {
		t.Fatalf("CLAUDE.md 생성 오류: %v", err)
	}
	registryPath := filepath.Join(dir, "zone-registry.md")
	if err := os.WriteFile(registryPath, []byte(validConstitutionRegistryForDoctor), 0o600); err != nil {
		t.Fatalf("registry 파일 생성 오류: %v", err)
	}

	check := checkConstitution(dir, registryPath, false, false)

	if check.Status != uikit.CheckOK {
		t.Errorf("skip-validate bypass Status = %q, want %q\n메시지: %s", check.Status, uikit.CheckOK, check.Message)
	}
}

// TestCheckConstitution_RegistryMissing verifies that a missing registry file returns Warn.
func TestCheckConstitution_RegistryMissing(t *testing.T) {
	dir := t.TempDir()
	nonExistentPath := filepath.Join(dir, "nonexistent-registry.md")

	check := checkConstitution(dir, nonExistentPath, false, false)

	if check.Status != uikit.CheckWarn {
		t.Errorf("registry 없음 Status = %q, want %q", check.Status, uikit.CheckWarn)
	}
}

// TestCheckConstitution_DuplicateIDs verifies that a registry with duplicate IDs returns Fail.
func TestCheckConstitution_DuplicateIDs(t *testing.T) {
	dir := t.TempDir()

	// Create CLAUDE.md
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Test"), 0o600); err != nil {
		t.Fatalf("CLAUDE.md 생성 오류: %v", err)
	}

	registryPath := filepath.Join(dir, "zone-registry.md")
	if err := os.WriteFile(registryPath, []byte(duplicateIDRegistryForDoctor), 0o600); err != nil {
		t.Fatalf("registry 파일 생성 오류: %v", err)
	}

	check := checkConstitution(dir, registryPath, false, false)

	if check.Status != uikit.CheckFail {
		t.Errorf("중복 ID registry Status = %q, want %q\n메시지: %s", check.Status, uikit.CheckFail, check.Message)
	}
}

// TestCheckConstitution_EmptyFrozen verifies that the absence of Frozen entries returns Warn.
// Relates to AC-CON-001-006.
func TestCheckConstitution_EmptyFrozen(t *testing.T) {
	dir := t.TempDir()

	// Create CLAUDE.md
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Test"), 0o600); err != nil {
		t.Fatalf("CLAUDE.md 생성 오류: %v", err)
	}

	registryPath := filepath.Join(dir, "zone-registry.md")
	// The clause must be present in the source so this fixture isolates the
	// empty-Frozen condition rather than tripping the drift check.
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Test\n\nSome evolvable rule applies here.\n"), 0o600); err != nil {
		t.Fatalf("CLAUDE.md 갱신 오류: %v", err)
	}
	if err := os.WriteFile(registryPath, []byte(emptyFrozenRegistryForDoctor), 0o600); err != nil {
		t.Fatalf("registry 파일 생성 오류: %v", err)
	}

	check := checkConstitution(dir, registryPath, false, false)

	if check.Status != uikit.CheckWarn {
		t.Errorf("empty Frozen Status = %q, want %q\n메시지: %s", check.Status, uikit.CheckWarn, check.Message)
	}

	if !strings.Contains(check.Message, "Frozen") {
		t.Errorf("경고 메시지에 'Frozen'이 포함되어야 한다: %s", check.Message)
	}
}

// TestCheckConstitution_StrictMode verifies that orphans are treated as Fail when MOAI_CONSTITUTION_STRICT=1.
// Relates to AC-CON-001-006.
func TestCheckConstitution_StrictMode(t *testing.T) {
	dir := t.TempDir()
	// CLAUDE.md is intentionally not created so an orphan is produced (file missing)

	registryPath := filepath.Join(dir, "zone-registry.md")
	// sampleRegistryContent uses CLAUDE.md which won't exist in dir
	if err := os.WriteFile(registryPath, []byte(sampleRegistryContent), 0o600); err != nil {
		t.Fatalf("registry 파일 생성 오류: %v", err)
	}

	// strictMode=true, orphan present → Fail
	check := checkConstitution(dir, registryPath, false, true)

	if check.Status != uikit.CheckFail {
		t.Errorf("strict mode + orphan Status = %q, want %q\n메시지: %s", check.Status, uikit.CheckFail, check.Message)
	}
}
