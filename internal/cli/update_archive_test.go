// SPEC-V3R3-HARNESS-001 / T-M4-01
// Table-driven tests for the archiveSkill function.
// RED phase: archiveSkill is not yet implemented → RED is confirmed via compile failure.

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

// hashDir computes the SHA-256 hash of every file under dir and returns a
// path → hash map. Used to verify byte-level equality.
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
		defer func() { _ = f.Close() }()
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

// makeSkillDir creates a test skill directory at projectRoot/.claude/skills/<id>.
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

// TestArchiveSkill_Present verifies that, when the skill directory exists, the
// archive is created correctly.
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

	// Verify the archive path
	archiveDir := filepath.Join(root, ".moai", "archive", "skills", "v2.16", skillID)
	if _, statErr := os.Stat(archiveDir); statErr != nil {
		t.Fatalf("archive directory not created: %v", statErr)
	}

	// Confirm via SHA-256 that the original and the archived content are byte-identical
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

// TestArchiveSkill_Absent verifies that archiveSkill returns nil when the source
// skill directory does not exist (idempotency).
func TestArchiveSkill_Absent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	// Do not create the skill directory

	err := archiveSkill(root, "moai-domain-frontend")
	if err != nil {
		t.Fatalf("archiveSkill should return nil when source absent, got: %v", err)
	}
}

// TestArchiveSkill_Idempotent verifies that calling archiveSkill twice with the
// same content succeeds idempotently without error.
func TestArchiveSkill_Idempotent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	skillID := "moai-domain-database"
	content := "# moai-domain-database"
	makeSkillDir(t, root, skillID, content)

	// First archive
	if err := archiveSkill(root, skillID); err != nil {
		t.Fatalf("first archiveSkill: %v", err)
	}

	// Second archive (same content) → must succeed without error
	if err := archiveSkill(root, skillID); err != nil {
		t.Fatalf("second archiveSkill (idempotent): %v", err)
	}
}

// TestArchiveSkill_DriftDetected verifies that an error is returned when the
// archive already exists but the source content has changed.
func TestArchiveSkill_DriftDetected(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	skillID := "moai-domain-db-docs"
	makeSkillDir(t, root, skillID, "original content")

	// First archive
	if err := archiveSkill(root, skillID); err != nil {
		t.Fatalf("first archiveSkill: %v", err)
	}

	// Modify the source content (simulates drift)
	skillFile := filepath.Join(root, ".claude", "skills", skillID, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte("modified content"), 0o644); err != nil {
		t.Fatalf("write modified content: %v", err)
	}

	// Second archive → expect an error due to content mismatch
	err := archiveSkill(root, skillID)
	if err == nil {
		t.Fatal("expected error for content drift, got nil")
	}
	if !strings.Contains(err.Error(), "drift") && !strings.Contains(err.Error(), "mismatch") &&
		!strings.Contains(err.Error(), "differ") && !strings.Contains(err.Error(), "ARCHIVE_DRIFT") {
		t.Errorf("error message should mention drift/mismatch, got: %v", err)
	}
}

// TestArchiveSkill_PathTraversal verifies that an error is returned when
// skillID contains ".." or "/" (path-traversal prevention).
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

// TestArchiveSkill_All16Skills iterates over all 16 legacy skills and verifies
// the present / absent cases.
func TestArchiveSkill_All16Skills(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Even-indexed skills are present; odd-indexed skills are absent
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

	// root is not used, so no cleanup is required (t.TempDir handles it automatically)
	_ = root
}

// TestCopyFile_SourceNotExist verifies that copyFile returns an error when the
// source file does not exist.
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

// TestCopyFile_DstDirNotExist verifies that copyFile returns an error when the
// destination directory does not exist.
func TestCopyFile_DstDirNotExist(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	// Create the source file
	src := filepath.Join(root, "src.txt")
	if err := os.WriteFile(src, []byte("data"), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}
	// Do not create the destination directory
	dst := filepath.Join(root, "subdir", "nonexistent", "dst.txt")
	err := copyFile(src, dst)
	if err == nil {
		t.Fatal("expected error when dst directory does not exist, got nil")
	}
}

// TestCopyFile_Success verifies that copyFile copies the file correctly.
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
