// SPEC-V3R3-HARNESS-001 / T-M4-04
// --dry-run 플래그 테스트: 파일시스템 변경 없이 계획된 작업을 출력해야 한다.

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestDryRunArchive는 dryRunArchiveLegacySkills가 파일시스템을 변경하지 않고
// [dry-run] 접두사가 붙은 출력을 생성하는지 검증한다.
func TestDryRunArchive(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// 16개 중 3개 스킬만 생성
	presentSkills := []string{
		"moai-domain-backend",
		"moai-domain-frontend",
		"moai-domain-database",
	}
	for _, id := range presentSkills {
		makeSkillDir(t, root, id, "# "+id)
	}

	// mtime 스냅샷 캡처
	skillsDir := filepath.Join(root, ".claude", "skills")
	preMtimes := snapshotMtimes(t, skillsDir)

	// dry-run 실행
	var out bytes.Buffer
	err := dryRunArchiveLegacySkills(root, &out)
	if err != nil {
		t.Fatalf("dryRunArchiveLegacySkills returned error: %v", err)
	}

	// 파일시스템 변경 없음 검증 (mtime 비교)
	postMtimes := snapshotMtimes(t, skillsDir)
	for path, preMtime := range preMtimes {
		postMtime, ok := postMtimes[path]
		if !ok {
			t.Errorf("file disappeared during dry-run: %s", path)
			continue
		}
		if !preMtime.Equal(postMtime) {
			t.Errorf("mtime changed during dry-run for %s: pre=%v post=%v",
				path, preMtime, postMtime)
		}
	}

	// 아카이브 디렉토리가 생성되지 않았는지 확인
	archiveBase := filepath.Join(root, ".moai", "archive")
	if _, err := os.Stat(archiveBase); err == nil {
		t.Error("archive directory was created during dry-run (should not mutate filesystem)")
	}

	// 출력 검증: [dry-run] 접두사 + 3개 스킬 + summary 라인
	output := out.String()
	if !strings.Contains(output, "[dry-run]") {
		t.Errorf("output missing [dry-run] prefix, got:\n%s", output)
	}
	for _, id := range presentSkills {
		if !strings.Contains(output, id) {
			t.Errorf("output missing skill ID %s, got:\n%s", id, output)
		}
	}

	// summary 라인 검증
	if !strings.Contains(output, "total:") {
		t.Errorf("output missing summary line containing 'total:', got:\n%s", output)
	}
}

// TestDryRunArchive_NoSkills는 레거시 스킬이 없을 때 dry-run이
// 빈 계획을 출력하고 에러 없이 종료하는지 검증한다.
func TestDryRunArchive_NoSkills(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	var out bytes.Buffer
	err := dryRunArchiveLegacySkills(root, &out)
	if err != nil {
		t.Fatalf("dryRunArchiveLegacySkills on empty project: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "total:") {
		t.Errorf("output missing summary line, got:\n%s", output)
	}
	// 0 skills 아카이브
	if !strings.Contains(output, "0 skills archived") {
		t.Errorf("expected '0 skills archived' in output, got:\n%s", output)
	}
}

// snapshotMtimes는 dir 하위 모든 파일의 경로→mtime 맵을 반환한다.
func snapshotMtimes(t *testing.T, dir string) map[string]time.Time {
	t.Helper()
	result := make(map[string]time.Time)
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return nil
		}
		result[path] = info.ModTime()
		return nil
	})
	return result
}
