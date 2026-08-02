// Package cli — update_preserve_partial_test.go
//
// Failure-branch coverage for mergeBackPreserveInventory (Step 6 of the
// canonical clean-reinstall order). Covers REQ-UDS-026 (the error names the
// stopping file AND the already-restored count), REQ-UDS-027 (each of the
// three failure returns is reached deterministically), and REQ-UDS-028 (the
// stat / MkdirAll / copyFile branches are distinguishable from one another).
//
// Every fixture uses t.TempDir() for filesystem isolation per
// CLAUDE.local.md §6 HARD.

package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// seedPartialRestoreFixture writes the given relative paths into a fresh
// backup directory and returns (projectRoot, backupDir). The inventory order
// is the caller's; mergeBackPreserveInventory iterates inv.Files in order, so
// placing the failing entry second guarantees a non-zero restored count at the
// moment of failure.
func seedPartialRestoreFixture(t *testing.T, rels ...string) (string, string) {
	t.Helper()

	root := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "preserve-backup")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		t.Fatalf("mkdir backupDir: %v", err)
	}
	for _, rel := range rels {
		writeTestFile(t, backupDir, rel, "backup content for "+rel+"\n")
	}
	return root, backupDir
}

// assertPartialRestoreError requires err to be non-nil and to carry BOTH the
// branch-distinguishing substring (REQ-UDS-028) and the restored-count detail
// (REQ-UDS-026). Asserting the specific substring — rather than merely
// err != nil — is what proves the intended branch was reached instead of an
// unrelated earlier failure.
func assertPartialRestoreError(t *testing.T, err error, wantBranch, wantCount string) {
	t.Helper()

	if err == nil {
		t.Fatalf("mergeBackPreserveInventory: expected error, got nil")
	}
	got := err.Error()
	if !strings.Contains(got, wantBranch) {
		t.Errorf("error does not name the expected branch + file:\n  got:  %s\n  want substring: %s", got, wantBranch)
	}
	if !strings.Contains(got, wantCount) {
		t.Errorf("error does not report the already-restored count:\n  got:  %s\n  want substring: %s", got, wantCount)
	}
}

// TestMergeBackPreserveInventory_PartialRestore drives each of the three
// failure returns of mergeBackPreserveInventory and asserts that the returned
// error is (a) attributable to that specific branch and (b) reports how many
// files had already been restored when it stopped.
func TestMergeBackPreserveInventory_PartialRestore(t *testing.T) {
	// Branch 1 — the stat of the backup source fails with something other than
	// os.ErrNotExist (a missing backup file is a documented `continue`, not a
	// failure). The failure is delivered through the injectable osStatFn seam
	// (REQ-UGE-001) rather than by chmod'ing the backup directory to 0o000.
	//
	// The permission-bit approach this replaces was unreachable on two real
	// execution environments: Windows does not model POSIX permission denial,
	// and root bypasses permission bits entirely, so both were t.Skip'd and the
	// branch was never exercised in the Windows CI job (REQ-UGE-003). Driving
	// the seam removes the platform dependency — no t.Skip remains.
	//
	// NFR-UGE-001: osStatFn is a package-level variable; this subtest reassigns
	// it, so neither it nor its parent may call t.Parallel().
	t.Run("stat_failure_injected_stat_error", func(t *testing.T) {
		root, backupDir := seedPartialRestoreFixture(t, "ok/first.md", "blocked/second.md")

		blockedSrc := filepath.Join(backupDir, filepath.FromSlash("blocked/second.md"))

		origStat := osStatFn
		t.Cleanup(func() { osStatFn = origStat })
		osStatFn = func(name string) (os.FileInfo, error) {
			if name == blockedSrc {
				// Deliberately NOT os.ErrNotExist — that is the `continue` path.
				return nil, errors.New("injected stat failure")
			}
			return origStat(name)
		}

		inv := PreserveInventory{Files: []string{"ok/first.md", "blocked/second.md"}}
		err := mergeBackPreserveInventory(root, inv, backupDir)

		assertPartialRestoreError(t, err, "stat backup blocked/second.md", "restored 1/2")

		// The entry preceding the failure must actually have landed, so the
		// reported count is a real restore boundary and not a formatting
		// coincidence.
		if _, statErr := os.Stat(filepath.Join(root, "ok", "first.md")); statErr != nil {
			t.Errorf("the entry preceding the injected stat failure was not restored: %v", statErr)
		}
	})

	// Branch 2 — os.MkdirAll on the destination's parent fails. A regular file
	// occupying a path component of the parent yields ENOTDIR on every
	// platform, so this subtest is never skipped.
	t.Run("mkdirall_failure_parent_is_regular_file", func(t *testing.T) {
		root, backupDir := seedPartialRestoreFixture(t, "ok/first.md", "blocker/second.md")

		// A regular FILE where the directory "blocker" must be created.
		if err := os.WriteFile(filepath.Join(root, "blocker"), []byte("not a directory\n"), 0o644); err != nil {
			t.Fatalf("write blocking regular file: %v", err)
		}

		inv := PreserveInventory{Files: []string{"ok/first.md", "blocker/second.md"}}
		err := mergeBackPreserveInventory(root, inv, backupDir)

		assertPartialRestoreError(t, err, "create restore parent for blocker/second.md", "restored 1/2")
	})

	// Branch 3 — copyFile fails. A DIRECTORY at the destination file path makes
	// the open-for-write fail (EISDIR) after MkdirAll has already succeeded, so
	// the failure is unambiguously the copy branch. Cross-platform; never
	// skipped. This subtest also carries AC-UDS-017: it is the one that must
	// name both the stopping file and a NON-ZERO already-restored count, which
	// is why "ok/first.md" is restored successfully before it.
	t.Run("reports_restored_count", func(t *testing.T) {
		root, backupDir := seedPartialRestoreFixture(t, "ok/first.md", "dir/second.md")

		// A DIRECTORY exactly where the destination FILE must be written.
		if err := os.MkdirAll(filepath.Join(root, "dir", "second.md"), 0o755); err != nil {
			t.Fatalf("mkdir blocking directory: %v", err)
		}

		inv := PreserveInventory{Files: []string{"ok/first.md", "dir/second.md"}}
		err := mergeBackPreserveInventory(root, inv, backupDir)

		assertPartialRestoreError(t, err, "restore dir/second.md", "restored 1/2")

		// The first entry must actually have landed — otherwise "restored 1/2"
		// would be a coincidence of formatting rather than a real count.
		if _, statErr := os.Stat(filepath.Join(root, "ok", "first.md")); statErr != nil {
			t.Errorf("the entry preceding the failure was not restored: %v", statErr)
		}
	})
}
