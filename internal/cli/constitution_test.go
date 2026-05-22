package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// sampleRegistryContent is the minimal registry markdown used for tests.
const sampleRegistryContent = `# Test Registry

## Entries

` + "```yaml" + `
- id: CONST-V3R2-001
  zone: Frozen
  file: CLAUDE.md
  anchor: "#1-core-identity"
  clause: "SPEC+EARS format"
  canary_gate: true

- id: CONST-V3R2-002
  zone: Frozen
  file: CLAUDE.md
  anchor: "#quality-gates"
  clause: "TRUST 5"
  canary_gate: true

- id: CONST-V3R2-003
  zone: Evolvable
  file: CLAUDE.md
  anchor: "#time-estimation"
  clause: "Never use time predictions in plans or reports."
  canary_gate: false
` + "```" + `
`

// TestConstitutionListAllEntries verifies rendering of all registry entries.
// Relates to AC-CON-001-001.
func TestConstitutionListAllEntries(t *testing.T) {
	dir := t.TempDir()
	registryPath := filepath.Join(dir, "zone-registry.md")
	if err := os.WriteFile(registryPath, []byte(sampleRegistryContent), 0o600); err != nil {
		t.Fatalf("registry 파일 생성 오류: %v", err)
	}

	var buf bytes.Buffer
	err := runConstitutionList(&buf, io.Discard, dir, registryPath, nil, "", "table")
	if err != nil {
		t.Fatalf("runConstitutionList 오류: %v", err)
	}

	output := buf.String()
	// All three entries must appear in the output
	for _, id := range []string{"CONST-V3R2-001", "CONST-V3R2-002", "CONST-V3R2-003"} {
		if !strings.Contains(output, id) {
			t.Errorf("출력에 %q가 포함되지 않았다\n출력: %s", id, output)
		}
	}
}

// TestConstitutionListFilterFrozen verifies the --zone frozen filter.
// Maps directly to AC-CON-001-002.
func TestConstitutionListFilterFrozen(t *testing.T) {
	dir := t.TempDir()
	registryPath := filepath.Join(dir, "zone-registry.md")
	if err := os.WriteFile(registryPath, []byte(sampleRegistryContent), 0o600); err != nil {
		t.Fatalf("registry 파일 생성 오류: %v", err)
	}

	frozenZone := constitution.ZoneFrozen
	var buf bytes.Buffer
	err := runConstitutionList(&buf, io.Discard, dir, registryPath, &frozenZone, "", "table")
	if err != nil {
		t.Fatalf("runConstitutionList --zone frozen 오류: %v", err)
	}

	output := buf.String()

	// Only Frozen entries must appear
	if !strings.Contains(output, "CONST-V3R2-001") {
		t.Errorf("출력에 CONST-V3R2-001이 포함되지 않았다")
	}
	if !strings.Contains(output, "CONST-V3R2-002") {
		t.Errorf("출력에 CONST-V3R2-002가 포함되지 않았다")
	}

	// Evolvable entries must be excluded
	if strings.Contains(output, "CONST-V3R2-003") {
		t.Errorf("출력에 Evolvable 엔트리 CONST-V3R2-003이 포함되어서는 안 된다")
	}
}

// TestConstitutionListFilterByFile verifies the --file filter.
func TestConstitutionListFilterByFile(t *testing.T) {
	dir := t.TempDir()
	registryPath := filepath.Join(dir, "zone-registry.md")

	content := `# Test Registry

## Entries

` + "```yaml" + `
- id: CONST-V3R2-001
  zone: Frozen
  file: CLAUDE.md
  anchor: "#1-core-identity"
  clause: "SPEC+EARS format"
  canary_gate: true

- id: CONST-V3R2-002
  zone: Frozen
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#quality-gates"
  clause: "TRUST 5"
  canary_gate: true
` + "```" + `
`
	if err := os.WriteFile(registryPath, []byte(content), 0o600); err != nil {
		t.Fatalf("registry 파일 생성 오류: %v", err)
	}

	var buf bytes.Buffer
	err := runConstitutionList(&buf, io.Discard, dir, registryPath, nil, "CLAUDE.md", "table")
	if err != nil {
		t.Fatalf("runConstitutionList --file 오류: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "CONST-V3R2-001") {
		t.Errorf("출력에 CONST-V3R2-001이 포함되지 않았다")
	}
	if strings.Contains(output, "CONST-V3R2-002") {
		t.Errorf("출력에 moai-constitution.md 엔트리가 포함되어서는 안 된다")
	}
}

// TestConstitutionListJSON verifies JSON-format output.
func TestConstitutionListJSON(t *testing.T) {
	dir := t.TempDir()
	registryPath := filepath.Join(dir, "zone-registry.md")
	if err := os.WriteFile(registryPath, []byte(sampleRegistryContent), 0o600); err != nil {
		t.Fatalf("registry 파일 생성 오류: %v", err)
	}

	var buf bytes.Buffer
	err := runConstitutionList(&buf, io.Discard, dir, registryPath, nil, "", "json")
	if err != nil {
		t.Fatalf("runConstitutionList --format json 오류: %v", err)
	}

	output := buf.String()

	// Output must be valid JSON
	var result struct {
		Entries []map[string]any `json:"entries"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("JSON 파싱 오류: %v\n출력: %s", err, output)
	}

	if len(result.Entries) != 3 {
		t.Errorf("JSON entries 수 = %d, want 3", len(result.Entries))
	}
}

// TestConstitutionListRegistryMissing_FileNotFound verifies that a missing registry file returns an error.
// Maps directly to AC-CON-001-013 (file-missing subtest).
func TestConstitutionListRegistryMissing_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	nonExistentPath := filepath.Join(dir, "nonexistent-registry.md")

	var buf bytes.Buffer
	err := runConstitutionList(&buf, io.Discard, dir, nonExistentPath, nil, "", "table")
	if err == nil {
		t.Fatal("존재하지 않는 registry 경로에서 오류를 반환해야 한다")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, nonExistentPath) {
		t.Errorf("오류 메시지에 경로 %q가 포함되어야 한다: %v", nonExistentPath, err)
	}
}

// TestConstitutionListRegistryMissing_PermissionDenied verifies that a registry file without permissions returns an error.
// Maps directly to AC-CON-001-013 (permission-denied subtest).
func TestConstitutionListRegistryMissing_PermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows에서는 os.Chmod가 Unix 권한을 시뮬레이션하지 않아 건너뜁니다")
	}
	if os.Getuid() == 0 {
		t.Skip("root에서는 권한 거부 테스트를 건너뜁니다")
	}

	dir := t.TempDir()
	registryPath := filepath.Join(dir, "zone-registry.md")
	if err := os.WriteFile(registryPath, []byte(sampleRegistryContent), 0o600); err != nil {
		t.Fatalf("registry 파일 생성 오류: %v", err)
	}

	// Strip permissions
	if err := os.Chmod(registryPath, 0o000); err != nil {
		t.Fatalf("chmod 오류: %v", err)
	}
	defer func() {
		_ = os.Chmod(registryPath, 0o600)
	}()

	var buf bytes.Buffer
	err := runConstitutionList(&buf, io.Discard, dir, registryPath, nil, "", "table")
	if err == nil {
		t.Fatal("권한 없는 registry 파일에서 오류를 반환해야 한다")
	}
}
