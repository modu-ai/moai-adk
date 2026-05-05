// manager_develop_present_test.go: embedded FS에 manager-develop.md 존재 여부 검증.
// REQ-RA-003 매핑.
//
// M1 RED phase: manager-develop.md가 아직 embedded FS에 없으므로 의도적으로 FAIL.
// M2에서 manager-develop.md 추가 후 GREEN이 됨.
package template

import (
	"io/fs"
	"testing"
)

// TestManagerDevelopPresentInEmbeddedFS는 embedded FS에
// manager-develop.md가 존재하고 최소 크기를 충족하는지 검증한다.
//
// REQ-RA-003: make build 이후 embedded FS에 manager-develop.md가 포함되어야 함
// 예상 RED: 파일이 없으므로 FAIL (M2에서 GREEN)
func TestManagerDevelopPresentInEmbeddedFS(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() 오류: %v", err)
	}

	const managerDevelopPath = ".claude/agents/moai/manager-develop.md"

	// 파일 존재 확인
	info, statErr := fs.Stat(fsys, managerDevelopPath)
	if statErr != nil {
		t.Fatalf("RETIREMENT_INCOMPLETE_manager-tdd: manager-develop.md가 embedded FS에 없음. "+
			"SPEC-V3R3-RETIRED-AGENT-001 M2에서 추가 필요 (REQ-RA-003): %v", statErr)
	}

	// 크기 sanity check: 최소 5000 bytes (full agent definition 검증)
	const minSize = 5000
	if info.Size() < minSize {
		t.Errorf("manager-develop.md 크기가 너무 작음: %d bytes (최소 %d bytes 필요). "+
			"완전한 에이전트 정의 파일이 아닐 수 있음 (REQ-RA-003)", info.Size(), minSize)
	}
}

// TestManagerDevelopFrontmatterValid는 manager-develop.md의 frontmatter가
// 필수 필드를 포함하는지 검증한다.
//
// REQ-RA-001: manager-develop.md는 완전한 frontmatter를 가져야 함
// 예상 RED: 파일 자체가 없으므로 FAIL (파일 읽기 실패 → TestManagerDevelopPresentInEmbeddedFS와 연동)
func TestManagerDevelopFrontmatterValid(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() 오류: %v", err)
	}

	const managerDevelopPath = ".claude/agents/moai/manager-develop.md"

	data, readErr := fs.ReadFile(fsys, managerDevelopPath)
	if readErr != nil {
		t.Fatalf("manager-develop.md 읽기 실패 (파일이 없음 — M2에서 추가 필요): %v", readErr)
	}

	fm, _, parseErr := parseFrontmatterAndBody(string(data))
	if parseErr != "" {
		t.Fatalf("manager-develop.md frontmatter 파싱 실패: %s", parseErr)
	}

	// 필수 frontmatter 필드 검증 (REQ-RA-001: full frontmatter)
	requiredFields := []string{
		"name",
		"description",
		"tools",
		"model",
	}
	for _, field := range requiredFields {
		val, ok := fm[field]
		if !ok || val == "" {
			t.Errorf("manager-develop.md frontmatter에 필수 필드 '%s' 없음 또는 비어 있음 (REQ-RA-001)", field)
		}
	}

	// retired: 필드가 없어야 함 (활성 에이전트)
	if _, hasRetired := fm["retired"]; hasRetired {
		t.Errorf("manager-develop.md는 활성 에이전트여야 하므로 'retired:' 필드가 없어야 함 (REQ-RA-001)")
	}

	// name이 manager-develop인지 확인
	if name, ok := fm["name"]; ok && name != "manager-develop" {
		t.Errorf("manager-develop.md의 name 필드가 'manager-develop'이어야 함, 실제: %q", name)
	}
}
