//go:build unix

package backup

// retention_prune_unix_test.go — AC-UBR-11 (SPEC-UPDATE-BACKUP-RETENTION-001):
// unclassifiable content fails closed and never blocks the prune.

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestAC_UBR_11_UnclassifiableContentFailsClosedAndNeverBlocks(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission bits do not restrict root")
	}
	root := t.TempDir()
	t1733RunWith(t, root, 4, map[string]string{"x.md": "w\n"})
	// Redundant: its only content is the window's version.
	redundant := t1733RunWith(t, root, 3, map[string]string{"x.md": "w\n"})
	// FIFO-only: no protected key, and the FIFO must never be opened.
	fifoRun := t1733RunWith(t, root, 2, map[string]string{"keep.txt": "k\n"})
	if err := os.Remove(filepath.Join(fifoRun, "keep.txt")); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(filepath.Join(fifoRun, "pipe"), 0o644); err != nil {
		t.Fatalf("mkfifo: %v", err)
	}
	// Removing and creating entries moved the directory's modification time, which
	// orders unmarked runs; put it back on the day its name encodes.
	stamp := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(fifoRun, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	// Unreadable file.
	unreadable := t1733RunWith(t, root, 1, map[string]string{"secret.txt": "s\n"})
	secret := filepath.Join(unreadable, "secret.txt")
	if err := os.Chmod(secret, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(secret, 0o644) })

	done := make(chan PruneReport, 1)
	go func() { done <- PruneBackups(root, 1, 3) }()
	var rep PruneReport
	select {
	case rep = <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("AC-UBR-11: the prune did not return; a FIFO was opened")
	}

	for _, run := range []string{fifoRun, unreadable} {
		it, ok := t1733Item(rep, run)
		if !ok || it.Action != RetentionActionRetained || it.Reason != RetentionReasonUnclassifiable {
			t.Errorf("AC-UBR-11: %s = %+v (found=%v), want retained for %s", run, it, ok, RetentionReasonUnclassifiable)
		}
		if it.Detail == "" {
			t.Errorf("AC-UBR-11: %s names no unclassifiable entry", run)
		}
		if !t1733Exists(run) {
			t.Errorf("AC-UBR-11: exempt run %s was removed", run)
		}
	}
	if it, ok := t1733Item(rep, redundant); !ok || it.Action != RetentionActionRemoved {
		t.Errorf("AC-UBR-11: redundant readable run = %+v (found=%v), want removed", it, ok)
	}
	if t1733Exists(redundant) {
		t.Errorf("AC-UBR-11: the redundant run remains on disk")
	}
}
