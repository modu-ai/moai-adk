// SPEC-V3R3-HARNESS-001 / T-M4-02
// archiveLegacySkills 함수 통합 테스트.
// 16개 레거시 스킬 + 사용자 my-harness 스킬 + moai-meta-harness가 있는
// 프로젝트에서 archiveLegacySkills를 호출했을 때:
//   - 16개 레거시 스킬이 아카이브됨
//   - my-harness-* 스킬은 건드리지 않음
//   - 출력 형식 검증

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestArchiveLegacySkills_Integration은 레거시 스킬 아카이브 흐름을
// 통합적으로 검증한다.
func TestArchiveLegacySkills_Integration(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// 16개 레거시 스킬 생성
	for _, id := range legacySkillIDs {
		makeSkillDir(t, root, id, "# "+id+" content")
	}

	// 사용자 커스터마이징 스킬 생성 (보존되어야 함)
	makeSkillDir(t, root, "my-harness-test", "# user custom skill")

	// moai-meta-harness (코어 스킬, 건드리지 않아야 함)
	makeSkillDir(t, root, "moai-meta-harness", "# meta harness")

	var out bytes.Buffer
	archived, err := archiveLegacySkills(root, &out)
	if err != nil {
		t.Fatalf("archiveLegacySkills: %v", err)
	}

	// 16개 모두 아카이브되었는지 검증
	if archived != len(legacySkillIDs) {
		t.Errorf("archived count = %d, want %d", archived, len(legacySkillIDs))
	}

	// 각 레거시 스킬에 대해 아카이브 디렉토리 존재 확인
	for _, id := range legacySkillIDs {
		archiveDir := filepath.Join(root, ".moai", "archive", "skills", archiveVersion, id)
		if _, statErr := os.Stat(archiveDir); statErr != nil {
			t.Errorf("archive not created for %s: %v", id, statErr)
		}
	}

	// my-harness-test 스킬은 아카이브되지 않았는지 확인
	userArchive := filepath.Join(root, ".moai", "archive", "skills", archiveVersion, "my-harness-test")
	if _, statErr := os.Stat(userArchive); statErr == nil {
		t.Error("my-harness-test should NOT be archived (user customization)")
	}

	// moai-meta-harness도 아카이브되지 않음
	metaArchive := filepath.Join(root, ".moai", "archive", "skills", archiveVersion, "moai-meta-harness")
	if _, statErr := os.Stat(metaArchive); statErr == nil {
		t.Error("moai-meta-harness should NOT be archived (not a legacy skill)")
	}

	// 출력 형식 검증: archive: <id> → ... 라인들
	output := out.String()
	for _, id := range legacySkillIDs {
		expected := "archive: " + id
		if !strings.Contains(output, expected) {
			t.Errorf("output missing archive line for %s, output:\n%s", id, output)
		}
	}

	// summary 라인: "total: N skills archived"
	if !strings.Contains(output, "total:") {
		t.Errorf("output missing summary line, got:\n%s", output)
	}
}

// TestArchiveLegacySkills_PartialPresent는 일부 레거시 스킬만 있는 경우에도
// 정상 동작하는지 검증한다.
func TestArchiveLegacySkills_PartialPresent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// 처음 5개만 생성
	presentSkills := legacySkillIDs[:5]
	for _, id := range presentSkills {
		makeSkillDir(t, root, id, "# "+id)
	}

	var out bytes.Buffer
	archived, err := archiveLegacySkills(root, &out)
	if err != nil {
		t.Fatalf("archiveLegacySkills partial: %v", err)
	}

	if archived != len(presentSkills) {
		t.Errorf("archived = %d, want %d", archived, len(presentSkills))
	}
}
