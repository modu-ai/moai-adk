//go:build !windows

package harness

// Card t1467 M1, rename-failure arm. The injection needs a directory whose write bit is
// off, which is what makes the tmp+rename write fail at temp creation; the pre-fix direct
// append opens the EXISTING archive file instead and needs no directory write permission,
// so this test is RED there (it succeeds and modifies the archive). Windows gives a
// directory's read-only attribute no creation-blocking meaning, so the injection does not
// exist there and the test is unix-only.

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestPruneFailedRenamePreservesPreviousArchive: when the archive write cannot land (here:
// temp creation denied by a read-only archiveDir), the prune returns the error and the
// previous archive is byte-identical afterwards — a failed write never damages what is
// already archived.
func TestPruneFailedRenamePreservesPreviousArchive(t *testing.T) {
	t.Parallel()
	if os.Geteuid() == 0 {
		t.Skip("root bypasses permissions")
	}
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	archiveDir := filepath.Join(dir, "archive")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		t.Fatalf("mkdir archive: %v", err)
	}
	archivePath := filepath.Join(archiveDir, now.AddDate(0, 0, -40).UTC().Format("2006-01")+".jsonl.gz")
	writeArchiveFixture(t, archivePath, "archived-1", now.AddDate(0, 0, -40))

	before, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("read archive before: %v", err)
	}

	if err := os.Chmod(archiveDir, 0o500); err != nil {
		t.Fatalf("chmod archive dir read-only: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(archiveDir, 0o755) })

	perr := NewRetention(logPath, archiveDir, func() time.Time { return now }).PruneStaleEntries(30)
	if perr == nil {
		t.Fatalf("a prune whose archive write cannot land returned no error")
	}

	after, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("read archive after: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Errorf("the previous archive changed across a failed archive write (%d -> %d bytes)", len(before), len(after))
	}
}
