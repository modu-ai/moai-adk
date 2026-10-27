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
