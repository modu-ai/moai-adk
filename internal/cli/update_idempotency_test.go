// SPEC-V3R3-HARNESS-001 / T-M4-05
// archiveLegacySkills 멱등성 테스트:
// 두 번 실행해도 아카이브 상태가 변하지 않음을 검증한다.

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestArchiveIdempotency는 archiveLegacySkills를 두 번 호출했을 때
// 아카이브 상태(파일 목록 + 해시)가 변하지 않음을 검증한다.
func TestArchiveIdempotency(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// 16개 레거시 스킬 생성
	for _, id := range legacySkillIDs {
		makeSkillDir(t, root, id, "# "+id+" idempotency test content")
	}

	// 1회차 실행
	var out1 bytes.Buffer
	archived1, err := archiveLegacySkills(root, &out1)
	if err != nil {
		t.Fatalf("first archiveLegacySkills: %v", err)
	}

	// 아카이브 상태 스냅샷
	archiveRoot := filepath.Join(root, ".moai", "archive", "skills", archiveVersion)
	snap1 := takeArchiveSnapshot(t, archiveRoot)

	// 2회차 실행 (멱등)
	var out2 bytes.Buffer
	archived2, err := archiveLegacySkills(root, &out2)
	if err != nil {
		t.Fatalf("second archiveLegacySkills: %v", err)
	}

	// 아카이브 상태 스냅샷 재수집
	snap2 := takeArchiveSnapshot(t, archiveRoot)

	// 파일 수 동일
	if len(snap1) != len(snap2) {
		t.Errorf("archive file count changed: first=%d second=%d",
			len(snap1), len(snap2))
	}

	// 해시 동일 (중복/변조 없음)
	for path, hash1 := range snap1 {
		hash2, ok := snap2[path]
		if !ok {
			t.Errorf("file missing in second snapshot: %s", path)
			continue
		}
		if hash1 != hash2 {
			t.Errorf("archive content changed for %s: first=%s second=%s",
				path, hash1, hash2)
		}
	}

	// archived count 검증: 두 번째 실행에서 새로 아카이브된 파일은 없어야 함
	// (소스가 그대로이면 drift 에러 없이 0 또는 동일 수 반환)
	// archived1 == len(legacySkillIDs), archived2 == 0 (이미 아카이브됨) 또는 동일
	if archived1 != len(legacySkillIDs) {
		t.Errorf("first run archived=%d, want %d", archived1, len(legacySkillIDs))
	}
	// 두 번째 실행: 소스가 그대로이므로 drift 없이 모두 스킵 → 0 또는 동일
	if archived2 != 0 {
		t.Errorf("second run should archive 0 (already done), got %d", archived2)
	}
}

// TestArchiveIdempotency_NoNewFiles는 두 번째 실행 후 아카이브에
// 새 파일이 추가되지 않았는지 검증한다.
func TestArchiveIdempotency_NoNewFiles(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// 일부 스킬만 생성
	for _, id := range legacySkillIDs[:8] {
		makeSkillDir(t, root, id, "content for "+id)
	}

	var out bytes.Buffer
	if _, err := archiveLegacySkills(root, &out); err != nil {
		t.Fatalf("first run: %v", err)
	}

	archiveRoot := filepath.Join(root, ".moai", "archive", "skills", archiveVersion)
	before := countDirFiles(t, archiveRoot)

	if _, err := archiveLegacySkills(root, &out); err != nil {
		t.Fatalf("second run: %v", err)
	}

	after := countDirFiles(t, archiveRoot)
	if before != after {
		t.Errorf("archive file count changed: before=%d after=%d (duplicate files?)",
			before, after)
	}
}

// takeArchiveSnapshot은 archiveRoot 하위 모든 파일의 경로→해시 맵을 반환한다.
func takeArchiveSnapshot(t *testing.T, archiveRoot string) map[string]string {
	t.Helper()
	result := make(map[string]string)
	_ = filepath.WalkDir(archiveRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(archiveRoot, path)
		if err != nil {
			return err
		}
		rawHash, err := hashFile(path)
		if err != nil {
			return err
		}
		result[rel] = string(rawHash)
		return nil
	})
	return result
}

// countDirFiles는 dir 하위의 파일 수를 반환한다.
func countDirFiles(t *testing.T, dir string) int {
	t.Helper()
	var count int
	_ = filepath.WalkDir(dir, func(_ string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		count++
		return nil
	})
	return count
}
