package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// sampleRegistryContent는 테스트용 minimal registry 마크다운이다.
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

// TestConstitutionListAllEntries는 registry 전체 엔트리 렌더링을 검증한다.
// AC-CON-001-001 관련.
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
	// 모든 3개 엔트리가 출력에 포함되어야 한다
	for _, id := range []string{"CONST-V3R2-001", "CONST-V3R2-002", "CONST-V3R2-003"} {
		if !strings.Contains(output, id) {
			t.Errorf("출력에 %q가 포함되지 않았다\n출력: %s", id, output)
		}
	}
}

// TestConstitutionListFilterFrozen은 --zone frozen 필터를 검증한다.
// AC-CON-001-002 직접 매핑.
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

	// Frozen 엔트리만 포함되어야 한다
	if !strings.Contains(output, "CONST-V3R2-001") {
		t.Errorf("출력에 CONST-V3R2-001이 포함되지 않았다")
	}
	if !strings.Contains(output, "CONST-V3R2-002") {
		t.Errorf("출력에 CONST-V3R2-002가 포함되지 않았다")
	}

	// Evolvable 엔트리는 제외되어야 한다
	if strings.Contains(output, "CONST-V3R2-003") {
		t.Errorf("출력에 Evolvable 엔트리 CONST-V3R2-003이 포함되어서는 안 된다")
	}
}

// TestConstitutionListFilterByFile은 --file 필터를 검증한다.
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

// TestConstitutionListJSON은 JSON 형식 출력을 검증한다.
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

	// 유효한 JSON이어야 한다
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

// TestConstitutionListRegistryMissing_FileNotFound는 registry 파일 없음 시 에러를 검증한다.
// AC-CON-001-013 직접 매핑 (file-missing subtest).
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

// TestConstitutionListRegistryMissing_PermissionDenied는 권한 없는 registry 파일 시 에러를 검증한다.
// AC-CON-001-013 직접 매핑 (permission-denied subtest).
func TestConstitutionListRegistryMissing_PermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root에서는 권한 거부 테스트를 건너뜁니다")
	}

	dir := t.TempDir()
	registryPath := filepath.Join(dir, "zone-registry.md")
	if err := os.WriteFile(registryPath, []byte(sampleRegistryContent), 0o600); err != nil {
		t.Fatalf("registry 파일 생성 오류: %v", err)
	}

	// 권한 제거
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
