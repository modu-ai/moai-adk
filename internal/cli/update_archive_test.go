// SPEC-V3R3-HARNESS-001 / T-M4-01
// archiveSkill 함수에 대한 테이블 기반 테스트.
// RED 단계: archiveSkill 미구현 → 컴파일 실패로 RED 확인.

package cli

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// hashDir은 디렉토리 내 모든 파일의 SHA-256 해시를 계산하여
// 경로→해시 맵을 반환한다. 바이트 단위 동등성 검증에 사용.
func hashDir(t *testing.T, dir string) map[string]string {
	t.Helper()
	hashes := make(map[string]string)
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return err
		}
		hashes[rel] = fmt.Sprintf("%x", h.Sum(nil))
		return nil
	})
	if err != nil {
		t.Fatalf("hashDir(%s): %v", dir, err)
	}
	return hashes
}

// makeSkillDir는 테스트용 스킬 디렉토리를 projectRoot/.claude/skills/<id> 에 생성한다.
func makeSkillDir(t *testing.T, projectRoot, skillID, content string) {
	t.Helper()
	dir := filepath.Join(projectRoot, ".claude", "skills", skillID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("makeSkillDir mkdir: %v", err)
	}
	skillFile := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte(content), 0o644); err != nil {
		t.Fatalf("makeSkillDir writeFile: %v", err)
	}
}

// TestArchiveSkill_Present는 스킬 디렉토리가 존재할 때
// 아카이브가 올바르게 생성되는지 검증한다.
func TestArchiveSkill_Present(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	skillID := "moai-domain-backend"
	content := "# moai-domain-backend\nTest content."
	makeSkillDir(t, root, skillID, content)

	err := archiveSkill(root, skillID)
	if err != nil {
		t.Fatalf("archiveSkill returned error: %v", err)
	}

	// 아카이브 경로 검증
	archiveDir := filepath.Join(root, ".moai", "archive", "skills", "v2.16", skillID)
	if _, statErr := os.Stat(archiveDir); statErr != nil {
		t.Fatalf("archive directory not created: %v", statErr)
	}

	// 원본과 아카이브 내용이 바이트 단위로 동일한지 SHA-256으로 확인
	srcDir := filepath.Join(root, ".claude", "skills", skillID)
	srcHashes := hashDir(t, srcDir)
	dstHashes := hashDir(t, archiveDir)

	if len(srcHashes) != len(dstHashes) {
		t.Errorf("file count mismatch: src=%d dst=%d", len(srcHashes), len(dstHashes))
	}
	for rel, srcHash := range srcHashes {
		dstHash, ok := dstHashes[rel]
		if !ok {
			t.Errorf("file missing in archive: %s", rel)
			continue
		}
		if srcHash != dstHash {
			t.Errorf("content mismatch for %s: src=%s dst=%s", rel, srcHash, dstHash)
		}
	}
}

// TestArchiveSkill_Absent는 소스 스킬 디렉토리가 없을 때
// archiveSkill이 nil을 반환(멱등성)하는지 검증한다.
func TestArchiveSkill_Absent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	// 스킬 디렉토리를 생성하지 않음

	err := archiveSkill(root, "moai-domain-frontend")
	if err != nil {
		t.Fatalf("archiveSkill should return nil when source absent, got: %v", err)
	}
}

// TestArchiveSkill_Idempotent는 동일한 내용으로 두 번 호출해도
// 에러 없이 멱등하게 동작하는지 검증한다.
func TestArchiveSkill_Idempotent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	skillID := "moai-domain-database"
	content := "# moai-domain-database"
	makeSkillDir(t, root, skillID, content)

	// 첫 번째 아카이브
	if err := archiveSkill(root, skillID); err != nil {
		t.Fatalf("first archiveSkill: %v", err)
	}

	// 두 번째 아카이브 (동일 내용) → 에러 없어야 함
	if err := archiveSkill(root, skillID); err != nil {
		t.Fatalf("second archiveSkill (idempotent): %v", err)
	}
}

// TestArchiveSkill_DriftDetected는 아카이브가 이미 존재하는데
// 소스 내용이 변경된 경우 에러를 반환하는지 검증한다.
func TestArchiveSkill_DriftDetected(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	skillID := "moai-domain-db-docs"
	makeSkillDir(t, root, skillID, "original content")

	// 첫 번째 아카이브
	if err := archiveSkill(root, skillID); err != nil {
		t.Fatalf("first archiveSkill: %v", err)
	}

	// 소스 내용 변경 (drift 시뮬레이션)
	skillFile := filepath.Join(root, ".claude", "skills", skillID, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte("modified content"), 0o644); err != nil {
		t.Fatalf("write modified content: %v", err)
	}

	// 두 번째 아카이브 → 내용 불일치로 에러 반환 기대
	err := archiveSkill(root, skillID)
	if err == nil {
		t.Fatal("expected error for content drift, got nil")
	}
	if !strings.Contains(err.Error(), "drift") && !strings.Contains(err.Error(), "mismatch") &&
		!strings.Contains(err.Error(), "differ") && !strings.Contains(err.Error(), "ARCHIVE_DRIFT") {
		t.Errorf("error message should mention drift/mismatch, got: %v", err)
	}
}

// TestArchiveSkill_PathTraversal는 skillID에 ".." 또는 "/"가 포함된
// 경우 에러를 반환하는지 검증한다 (경로 순회 방지).
func TestArchiveSkill_PathTraversal(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	cases := []struct {
		name    string
		skillID string
	}{
		{"dotdot", "../../etc/passwd"},
		{"dotdot_prefix", "../secret"},
		{"slash_in_id", "moai/evil"},
		{"absolute_path", "/etc/passwd"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := archiveSkill(root, tc.skillID)
			if err == nil {
				t.Errorf("expected error for path traversal skillID=%q, got nil", tc.skillID)
			}
		})
	}
}

// TestArchiveSkill_All16Skills는 16개 레거시 스킬을 모두 순회하며
// present / absent 두 가지 케이스를 검증한다.
func TestArchiveSkill_All16Skills(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// 짝수 인덱스는 present, 홀수 인덱스는 absent
	for i, skillID := range legacySkillIDs {
		skillID := skillID
		present := i%2 == 0
		t.Run(skillID, func(t *testing.T) {
			t.Parallel()
			localRoot := t.TempDir()
			if present {
				makeSkillDir(t, localRoot, skillID, fmt.Sprintf("# %s content", skillID))
			}

			err := archiveSkill(localRoot, skillID)
			if err != nil {
				t.Errorf("archiveSkill(%s) unexpected error: %v", skillID, err)
			}

			if present {
				archiveDir := filepath.Join(localRoot, ".moai", "archive", "skills", "v2.16", skillID)
				if _, statErr := os.Stat(archiveDir); statErr != nil {
					t.Errorf("archive not created for %s: %v", skillID, statErr)
				}
			}
		})
	}

	// root는 사용하지 않으므로 정리 불필요 (t.TempDir이 자동 정리)
	_ = root
}

// TestCopyFile_SourceNotExist는 소스 파일이 없을 때 copyFile이
// 에러를 반환하는지 검증한다.
func TestCopyFile_SourceNotExist(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	src := filepath.Join(root, "nonexistent.txt")
	dst := filepath.Join(root, "dst.txt")
	err := copyFile(src, dst)
	if err == nil {
		t.Fatal("expected error when source file does not exist, got nil")
	}
}

// TestCopyFile_DstDirNotExist는 대상 디렉토리가 없을 때 copyFile이
// 에러를 반환하는지 검증한다.
func TestCopyFile_DstDirNotExist(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	// 소스 파일 생성
	src := filepath.Join(root, "src.txt")
	if err := os.WriteFile(src, []byte("data"), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}
	// 대상 디렉토리는 생성하지 않음
	dst := filepath.Join(root, "subdir", "nonexistent", "dst.txt")
	err := copyFile(src, dst)
	if err == nil {
		t.Fatal("expected error when dst directory does not exist, got nil")
	}
}

// TestCopyFile_Success는 copyFile이 정상적으로 파일을 복사하는지 검증한다.
func TestCopyFile_Success(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	src := filepath.Join(root, "src.txt")
	content := []byte("hello archive world")
	if err := os.WriteFile(src, content, 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}
	dst := filepath.Join(root, "dst.txt")
	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("content mismatch: got %q want %q", got, content)
	}
}
