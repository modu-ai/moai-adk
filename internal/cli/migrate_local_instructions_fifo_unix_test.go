//go:build unix

package cli

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestMigrateLocalInstructions_RefusesSymlinkSwap pins sync-audit F8: when
// the legacy file is moved OUT of the project and a SYMLINK to an outside
// regular file with the same mtime and content takes its place inside the
// read->write window, the re-check must refuse via the changed path — the
// final rename must never move the LINK into the backup and report
// "Migrated" (REQ-CIF-009: symlinks are rejected, never followed).
func TestMigrateLocalInstructions_RefusesSymlinkSwap(t *testing.T) {
	for _, pair := range migrationPairs {
		t.Run(pair.legacy, func(t *testing.T) {
			root := t.TempDir()
			outside := t.TempDir()
			legacyPath := filepath.Join(root, pair.legacy)
			body := []byte("legacy body\n")
			if err := os.WriteFile(legacyPath, body, 0o600); err != nil {
				t.Fatal(err)
			}
			info, err := os.Lstat(legacyPath)
			if err != nil {
				t.Fatal(err)
			}

			// Same-mtime, same-content symlink swap inside the window: mv the
			// original out (mtime preserved) and link an identical outside file.
			unlock := injectMigrateGuardHook(t, func() {
				moved := filepath.Join(outside, "moved-"+pair.legacy)
				if err := os.Rename(legacyPath, moved); err != nil {
					t.Fatal(err)
				}
				target := filepath.Join(outside, "target.txt")
				if err := os.WriteFile(target, body, 0o600); err != nil {
					t.Fatal(err)
				}
				if err := os.Chtimes(target, info.ModTime(), info.ModTime()); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, legacyPath); err != nil {
					t.Fatal(err)
				}
			})

			out, _, err := runMigrateLocalInstructions(t, root)
			unlock()
			if err == nil {
				t.Fatalf("migration over a symlink swap reported success (F8): %q", out)
			}
			if !strings.Contains(err.Error(), "changed") {
				t.Errorf("refusal = %q, want it to name the concurrent change", err.Error())
			}
			// The link stays in place, unfollowed; the canonical file was not
			// written; no backup entry carries the link.
			fi, serr := os.Lstat(legacyPath)
			if serr != nil || fi.Mode()&os.ModeSymlink == 0 {
				t.Errorf("the swapped link was altered or followed (err=%v mode=%v)", serr, fi.Mode())
			}
			if _, cerr := os.Lstat(filepath.Join(root, pair.canonical)); !os.IsNotExist(cerr) {
				t.Errorf("%s written despite the refusal (err=%v)", pair.canonical, cerr)
			}
		})
	}
}

// TestMigrateLocalInstructions_RefusesSameMtimeFIFOSwap pins audit P2-3:
// when the legacy file is swapped for a writer-less FIFO carrying the SAME
// mtime between the read and the re-check, the verb must refuse via the
// handle's own mode — an unguarded ReadFile would hang the migration
// forever. The guard-hook seam performs the swap at the exact window.
func TestMigrateLocalInstructions_RefusesSameMtimeFIFOSwap(t *testing.T) {
	for _, pair := range migrationPairs {
		t.Run(pair.legacy, func(t *testing.T) {
			root := t.TempDir()
			legacyPath := filepath.Join(root, pair.legacy)
			if err := os.WriteFile(legacyPath, []byte("legacy body\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			recorded := time.Now()
			// Same-mtime FIFO swap inside the read->write window.
			unlock := injectMigrateGuardHook(t, func() {
				info, err := os.Lstat(legacyPath)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(legacyPath); err != nil {
					t.Fatal(err)
				}
				if err := syscall.Mkfifo(legacyPath, 0o600); err != nil {
					t.Fatal(err)
				}
				if err := os.Chtimes(legacyPath, info.ModTime(), info.ModTime()); err != nil {
					t.Fatal(err)
				}
			})
			_ = recorded

			_, _, err := runMigrateLocalInstructions(t, root)
			unlock()
			if err == nil {
				t.Fatal("migration over a same-mtime FIFO swap exited 0, want a refusal")
			}
			if !strings.Contains(err.Error(), "changed") {
				t.Errorf("refusal = %q, want it to name the concurrent change", err.Error())
			}
			// No canonical file was written; the FIFO is untouched in place.
			if info, serr := os.Lstat(legacyPath); serr != nil || info.Mode()&os.ModeNamedPipe == 0 {
				t.Errorf("the swapped FIFO was altered (err=%v mode=%v)", serr, info.Mode())
			}
			if _, cerr := os.Lstat(filepath.Join(root, pair.canonical)); !os.IsNotExist(cerr) {
				t.Errorf("%s written despite the refusal (err=%v)", pair.canonical, cerr)
			}
		})
	}
}
