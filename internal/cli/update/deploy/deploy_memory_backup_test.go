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
