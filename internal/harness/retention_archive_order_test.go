package harness

// SPEC-CAPACITY-RECLAIM-001 AC-CAP-009: the retention prune archives stale
// events BEFORE it rewrites the log (retention.go prune: partition → archive
// → overwrite). REQ-CAP-004's ordering guarantee is held by that existing
// sequence under the Q4 default — this is a characterization test of the
// default branch, not a behavior change: M1 changes neither the prune nor its
// lock. Observation (ii) is the guarantee's teeth: when the archive write
// fails, the log must be byte-identical to its pre-call content — the
// rewrite never happens on a failed archive.

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// TestRetentionPruneArchivesBeforeRewrite pins the prune's
// archive-before-rewrite ordering on fixture paths.
func TestRetentionPruneArchivesBeforeRewrite(t *testing.T) {
	t.Parallel()

	t.Run("archive holds stale events and log holds only fresh", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		logPath := filepath.Join(dir, "usage-log.jsonl")
		archiveDir := filepath.Join(dir, "archive")
		now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
		writeStaleLog(t, logPath, now, "stale-order-1")

		if err := NewRetention(logPath, archiveDir, func() time.Time { return now }).PruneStaleEntries(30); err != nil {
			t.Fatalf("prune: %v", err)
		}

		// (i) the stale event is present in archiveDir/<YYYY-MM>.jsonl.gz and
		// the rewritten log holds only the fresh events.
		archivePath := filepath.Join(archiveDir, "2026-08.jsonl.gz") // stale at now-40d
		subjects := archiveSubjects(t, archivePath)
		if subjects["stale-order-1"] != 1 {
			t.Fatalf("archived events = %v, want stale-order-1 once", subjects)
		}
		logSubjects := logSubjectsForUnit(t, logPath)
		if logSubjects["stale-order-1"] != 0 {
			t.Fatalf("stale event survived the rewrite: %v", logSubjects)
		}
		if logSubjects["fresh"] != 1 {
			t.Fatalf("fresh event lost in the rewrite: %v", logSubjects)
		}
	})

	t.Run("failed archive leaves log byte-identical", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("directory permission bits are not enforced on Windows")
		}
		if os.Geteuid() == 0 {
			t.Skip("root bypasses directory permissions")
		}
		t.Parallel()
		dir := t.TempDir()
		logPath := filepath.Join(dir, "usage-log.jsonl")
		archiveDir := filepath.Join(dir, "archive")
		now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
		writeStaleLog(t, logPath, now, "stale-order-2")
		if err := os.MkdirAll(archiveDir, 0o755); err != nil {
			t.Fatalf("mkdir archive: %v", err)
		}

		before, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("read log before prune: %v", err)
		}

		// Make the archive write fail: the archive directory refuses writes,
		// so the tmp+rename append cannot create its temp file.
		if err := os.Chmod(archiveDir, 0o555); err != nil {
			t.Fatalf("chmod archive dir: %v", err)
		}
		t.Cleanup(func() { _ = os.Chmod(archiveDir, 0o755) })

		pruneErr := NewRetention(logPath, archiveDir, func() time.Time { return now }).PruneStaleEntries(30)

		// (ii) the rewrite never happens on a failed archive: the log file is
		// byte-identical to its pre-call content, whether or not the caller
		// sees the archive error surface.
		after, readErr := os.ReadFile(logPath)
		if readErr != nil {
			t.Fatalf("read log after failed-archive prune: %v", readErr)
		}
		if !bytes.Equal(before, after) {
			t.Fatalf("log was rewritten despite the failed archive (pruneErr=%v): before %d bytes, after %d bytes",
				pruneErr, len(before), len(after))
		}
	})
}
