package deploy

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// TestMigrateLegacyMemoryDir_BacksUpBeforeRemoval pins REQ-UDS-008
// (SPEC-UPDATE-DATA-SURVIVAL-001 M2, AC-UDS-006).
//
// When both .moai/memory/ and .moai/state/ exist, the migration removes the
// legacy directory outright. Nothing else in the update subsystem backs that
// directory up — it appears in none of preserveInventoryRoots,
// backup.BackupMoaiConfig (which covers .moai/config only), or
// userOwnedScanRoots — so the removal is unrecoverable user data loss unless a
// copy is taken first.
//
// The assertion is on the sentinel's BYTES, not merely on a path existing:
// an empty file at the right place would satisfy a path-only check while the
// user's content was still gone.
func TestMigrateLegacyMemoryDir_BacksUpBeforeRemoval(t *testing.T) {
	root := t.TempDir()

	legacyDir := filepath.Join(root, defs.MoAIDir, "memory")
	stateDir := filepath.Join(root, defs.MoAIDir, defs.StateSubdir)

	// Nested so the copy is exercised on a tree, not a single flat file.
	sentinelDir := filepath.Join(legacyDir, "notes")
	if err := os.MkdirAll(sentinelDir, defs.DirPerm); err != nil {
		t.Fatalf("create legacy memory dir: %v", err)
	}
	if err := os.MkdirAll(stateDir, defs.DirPerm); err != nil {
		t.Fatalf("create state dir: %v", err)
	}

	sentinel := []byte("user-authored memory that must survive the migration\n")
	if err := os.WriteFile(filepath.Join(sentinelDir, "keep.md"), sentinel, defs.FilePerm); err != nil {
		t.Fatalf("write sentinel: %v", err)
	}

	// A symlink pointing outside the tree. The copy must skip it rather than
	// follow it, which would write the linked content into the backup — and,
	// for a link to a directory, could pull in an arbitrary subtree.
	outsideTarget := filepath.Join(t.TempDir(), "outside.txt")
	outsideContent := []byte("content the backup must NOT reach through a symlink\n")
	if err := os.WriteFile(outsideTarget, outsideContent, defs.FilePerm); err != nil {
		t.Fatalf("write symlink target: %v", err)
	}
	symlinked := false
	if err := os.Symlink(outsideTarget, filepath.Join(legacyDir, "escape.txt")); err == nil {
		symlinked = true
	} // a platform that refuses symlinks simply skips this assertion

	var out bytes.Buffer
	if err := MigrateLegacyMemoryDir(root, &out); err != nil {
		t.Fatalf("MigrateLegacyMemoryDir: %v", err)
	}

	// The both-exist branch was taken: the legacy directory is gone.
	if _, err := os.Stat(legacyDir); !os.IsNotExist(err) {
		t.Fatalf("legacy dir still present (stat err = %v); the both-exist removal branch did not run, so this test would not be exercising REQ-UDS-008", err)
	}

	backupRoot := filepath.Join(root, defs.BackupsDir)
	found := findFileWithContent(t, backupRoot, sentinel)
	if found == "" {
		t.Fatalf("sentinel bytes not found anywhere under %s — .moai/memory/ was destroyed without a backup (REQ-UDS-008)", backupRoot)
	}
	t.Logf("sentinel bytes recovered from %s", found)

	if symlinked {
		if escaped := findFileWithContent(t, backupRoot, outsideContent); escaped != "" {
			t.Fatalf("the copy followed a symlink out of .moai/memory/ and wrote its target into the backup at %s", escaped)
		}
	}
}

// TestMigrateLegacyMemoryDir_AbortsRemovalOnBackupFailure pins the other half
// of REQ-UDS-008: a backup that FAILS must not be followed by the destruction
// it was taken to survive. Without this, a broken backup would degrade silently
// into exactly the data loss the requirement exists to prevent.
//
// The failure is injected, never raced: the backup root is pre-created as a
// regular file, so the copy's MkdirAll cannot succeed.
func TestMigrateLegacyMemoryDir_AbortsRemovalOnBackupFailure(t *testing.T) {
	root := t.TempDir()

	legacyDir := filepath.Join(root, defs.MoAIDir, "memory")
	stateDir := filepath.Join(root, defs.MoAIDir, defs.StateSubdir)
	if err := os.MkdirAll(legacyDir, defs.DirPerm); err != nil {
		t.Fatalf("create legacy memory dir: %v", err)
	}
	if err := os.MkdirAll(stateDir, defs.DirPerm); err != nil {
		t.Fatalf("create state dir: %v", err)
	}

	sentinelPath := filepath.Join(legacyDir, "keep.md")
	sentinel := []byte("must survive a failed backup\n")
	if err := os.WriteFile(sentinelPath, sentinel, defs.FilePerm); err != nil {
		t.Fatalf("write sentinel: %v", err)
	}

	// Occupy the backup root with a regular file so the copy cannot create
	// its destination directory.
	if err := os.WriteFile(filepath.Join(root, defs.BackupsDir), []byte("not a directory"), defs.FilePerm); err != nil {
		t.Fatalf("occupy backup root: %v", err)
	}

	var out bytes.Buffer
	err := MigrateLegacyMemoryDir(root, &out)
	if err == nil {
		t.Fatal("expected an error when the backup cannot be written, got nil")
	}

	// The legacy tree — and its content — must be untouched.
	got, readErr := os.ReadFile(sentinelPath)
	if readErr != nil {
		t.Fatalf("sentinel gone after a failed backup (REQ-UDS-008): %v", readErr)
	}
	if !bytes.Equal(got, sentinel) {
		t.Fatalf("sentinel content changed after a failed backup: got %q, want %q", got, sentinel)
	}
}

// findFileWithContent returns the path of the first regular file under root
// whose bytes equal want, or "" when there is none.
func findFileWithContent(t *testing.T, root string, want []byte) string {
	t.Helper()

	var found string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// A missing backup root is the failure this test reports, not a
			// walk error to escalate.
			if os.IsNotExist(err) {
				return filepath.SkipAll
			}
			return err
		}
		if d.IsDir() || found != "" {
			return nil
		}
		got, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if bytes.Equal(got, want) {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return found
}
