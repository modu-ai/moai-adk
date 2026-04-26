// SPEC-V3R3-HARNESS-001 / T-M4-03
// moai migrate restore-skill 서브커맨드 테스트.
// 라운드트립: archiveSkill → restoreSkill 후 바이트 동일 검증.

package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRestoreSkill_RoundTrip은 archive → restore 라운드트립 후
// 원본 파일과 복원된 파일이 바이트 단위로 동일한지 검증한다.
func TestRestoreSkill_RoundTrip(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	skillID := "moai-domain-backend"
	content := "# moai-domain-backend\noriginal content line 1\nline 2"
	makeSkillDir(t, root, skillID, content)

	// 원본 해시 기록
	srcDir := filepath.Join(root, ".claude", "skills", skillID)
	origHashes := hashDir(t, srcDir)

	// 1단계: 아카이브
	if err := archiveSkill(root, skillID); err != nil {
		t.Fatalf("archiveSkill: %v", err)
	}

	// 소스 디렉토리 제거 (restore 전 clean state 시뮬레이션)
	if err := os.RemoveAll(srcDir); err != nil {
		t.Fatalf("RemoveAll source: %v", err)
	}

	// 2단계: 복원
	if err := restoreSkill(root, skillID, false); err != nil {
		t.Fatalf("restoreSkill: %v", err)
	}

	// 복원된 해시와 원본 해시 비교
	restoredHashes := hashDir(t, srcDir)
	if len(origHashes) != len(restoredHashes) {
		t.Errorf("file count: original=%d restored=%d", len(origHashes), len(restoredHashes))
	}
	for rel, origHash := range origHashes {
		restoredHash, ok := restoredHashes[rel]
		if !ok {
			t.Errorf("file missing after restore: %s", rel)
			continue
		}
		if origHash != restoredHash {
			t.Errorf("content mismatch for %s: original=%s restored=%s",
				rel, origHash, restoredHash)
		}
	}
}

// TestRestoreSkill_ArchiveMissing는 아카이브가 없을 때 에러를 반환하는지 검증한다.
func TestRestoreSkill_ArchiveMissing(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	err := restoreSkill(root, "moai-domain-frontend", false)
	if err == nil {
		t.Fatal("expected error when archive is missing, got nil")
	}
}

// TestRestoreSkill_TargetExists는 복원 대상이 이미 존재할 때
// --force 없이 에러를 반환하는지 검증한다.
func TestRestoreSkill_TargetExists(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	skillID := "moai-domain-database"

	makeSkillDir(t, root, skillID, "original")
	if err := archiveSkill(root, skillID); err != nil {
		t.Fatalf("archiveSkill: %v", err)
	}
	// 소스는 그대로 존재

	// --force 없이 복원 시도 → 에러 기대
	err := restoreSkill(root, skillID, false)
	if err == nil {
		t.Fatal("expected error when target exists without --force, got nil")
	}
}

// TestRestoreSkill_TargetExistsWithForce는 --force가 있을 때
// 기존 대상을 덮어쓰는지 검증한다.
func TestRestoreSkill_TargetExistsWithForce(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	skillID := "moai-domain-db-docs"

	makeSkillDir(t, root, skillID, "original content")
	if err := archiveSkill(root, skillID); err != nil {
		t.Fatalf("archiveSkill: %v", err)
	}

	// 소스 내용 변경 (아카이브와 다른 상태)
	skillFile := filepath.Join(root, ".claude", "skills", skillID, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte("modified content"), 0o644); err != nil {
		t.Fatalf("write modified: %v", err)
	}

	// --force=true로 복원 → 성공 기대
	if err := restoreSkill(root, skillID, true); err != nil {
		t.Fatalf("restoreSkill with force: %v", err)
	}

	// 복원 후 내용이 아카이브(원본)와 동일해야 함
	restored, err := os.ReadFile(skillFile)
	if err != nil {
		t.Fatalf("read restored file: %v", err)
	}
	if string(restored) != "original content" {
		t.Errorf("restored content = %q, want %q", string(restored), "original content")
	}
}

// TestRestoreSkill_EmptySkillID는 빈 skillID로 restoreSkill을 호출했을 때
// RESTORE_INVALID_SKILL_ID 에러를 반환하는지 검증한다.
func TestRestoreSkill_EmptySkillID(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	err := restoreSkill(root, "", false)
	if err == nil {
		t.Fatal("expected error for empty skillID, got nil")
	}
	me, ok := err.(*MigrateError)
	if !ok {
		t.Fatalf("expected *MigrateError, got %T: %v", err, err)
	}
	if me.Code != "RESTORE_INVALID_SKILL_ID" {
		t.Errorf("code = %q, want RESTORE_INVALID_SKILL_ID", me.Code)
	}
}

// TestRestoreSkill_InvalidPathTraversal는 skillID에 ".."이 포함된 경우
// RESTORE_INVALID_SKILL_ID 에러를 반환하는지 검증한다.
func TestRestoreSkill_InvalidPathTraversal(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	err := restoreSkill(root, "../evil", false)
	if err == nil {
		t.Fatal("expected error for path traversal skillID, got nil")
	}
}

// TestRestoreSkill_All16RoundTrip는 16개 스킬 모두에 대해
// archive → restore 라운드트립을 검증한다.
func TestRestoreSkill_All16RoundTrip(t *testing.T) {
	t.Parallel()
	for _, skillID := range legacySkillIDs {
		skillID := skillID
		t.Run(skillID, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			content := "# " + skillID + " test content\nsome data"
			makeSkillDir(t, root, skillID, content)

			srcDir := filepath.Join(root, ".claude", "skills", skillID)
			origHashes := hashDir(t, srcDir)

			if err := archiveSkill(root, skillID); err != nil {
				t.Fatalf("archiveSkill: %v", err)
			}
			if err := os.RemoveAll(srcDir); err != nil {
				t.Fatalf("RemoveAll: %v", err)
			}
			if err := restoreSkill(root, skillID, false); err != nil {
				t.Fatalf("restoreSkill: %v", err)
			}

			restoredHashes := hashDir(t, srcDir)
			for rel, origHash := range origHashes {
				if restoredHashes[rel] != origHash {
					t.Errorf("hash mismatch for %s in %s", rel, skillID)
				}
			}
		})
	}
}
