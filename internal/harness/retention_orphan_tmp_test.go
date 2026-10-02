package harness

// Regression tests for card t1431: a pruner killed while rewriting the log leaves a
// usage-log-*.tmp behind, and the next prune cycle must sweep the old ones. They use only
// the public Retention API and the documented temp-file naming, so they compile against
// the pre-fix code and fail there (RED) instead of failing to build.

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeTmpAged creates a file at path whose mtime is age before now.
func writeTmpAged(t *testing.T, path string, now time.Time, age time.Duration) {
	t.Helper()
	if err := os.WriteFile(path, []byte("orphan"), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	mt := now.Add(-age)
	if err := os.Chtimes(path, mt, mt); err != nil {
		t.Fatalf("chtimes %s: %v", path, err)
	}
}

func pathExists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

// TestPruneSweepsOldOrphanTmp: the pruner that wins the lock deletes usage-log-*.tmp files
// that are old, whether or not the log has anything stale, and leaves everything else.
func TestPruneSweepsOldOrphanTmp(t *testing.T) {
	t.Parallel()
	cases := map[string]bool{ // name -> log has a stale event
		"log-with-stale-event": true,
		"log-nothing-stale":    false,
	}
	for name, hasStale := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			logPath := filepath.Join(dir, "usage-log.jsonl")
			// The injected clock runs 72 h ahead of the real one, so entries created below without
			// a chosen mtime (the directory and the symlink) already read as old; only their file
			// type keeps the sweep away from them.
			now := time.Now().Add(72 * time.Hour)
			if hasStale {
				writeStaleLog(t, logPath, now, "stale")
			} else {
				fresh := []Event{{Timestamp: now.AddDate(0, 0, -1), EventType: EventTypeFeedback, Subject: "fresh", ContextHash: "h", SchemaVersion: LogSchemaVersion}}
				if err := writeEventsToFile(logPath, fresh); err != nil {
					t.Fatalf("write log: %v", err)
				}
			}

			oldA := filepath.Join(dir, "usage-log-111111.tmp")
			oldB := filepath.Join(dir, "usage-log-222222.tmp")
			young := filepath.Join(dir, "usage-log-333333.tmp")
			writeTmpAged(t, oldA, now, time.Hour)
			writeTmpAged(t, oldB, now, 48*time.Hour)
			writeTmpAged(t, young, now, time.Minute)

			// Files that must never be swept: wrong prefix, wrong suffix, the log's own siblings.
			keepers := []string{
				filepath.Join(dir, "other-usage-log-1.tmp"),
				filepath.Join(dir, "usage-log-1.tmp.bak"),
				filepath.Join(dir, "usage-log.jsonl.bak"),
			}
			for _, k := range keepers {
				writeTmpAged(t, k, now, 48*time.Hour)
			}
			// An old directory that matches the name is not ours to remove (its mtime is the real
			// creation time, 72 h behind the injected clock).
			dirMatch := filepath.Join(dir, "usage-log-dir.tmp")
			if err := os.Mkdir(dirMatch, 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			// An old symlink that matches the name is skipped, and its target is untouched.
			target := filepath.Join(t.TempDir(), "elsewhere.txt")
			writeTmpAged(t, target, now, 48*time.Hour)
			link := filepath.Join(dir, "usage-log-link.tmp")
			haveLink := os.Symlink(target, link) == nil

			if err := NewRetention(logPath, filepath.Join(dir, "archive"), func() time.Time { return now }).PruneStaleEntries(30); err != nil {
				t.Fatalf("prune: %v", err)
			}

			for _, gone := range []string{oldA, oldB} {
				if pathExists(gone) {
					t.Errorf("old orphan %s survived the prune", filepath.Base(gone))
				}
			}
			if !pathExists(young) {
				t.Errorf("a young tmp (1 minute old) was swept; it may belong to a live rewrite")
			}
			for _, k := range append(keepers, dirMatch) {
				if !pathExists(k) {
					t.Errorf("%s was swept but is not a prune temp file", filepath.Base(k))
				}
			}
			if haveLink {
				if !pathExists(link) {
					t.Errorf("the symlink %s was removed", filepath.Base(link))
				}
				if !pathExists(target) {
					t.Errorf("the symlink target was removed")
				}
			}
		})
	}
}

// TestPruneSweepRidesTheLockHolderOnly: a hook process that finds a fresh stamp returns
// after one small read; it must not scan the directory, so the sweep waits for the next cycle.
func TestPruneSweepRidesTheLockHolderOnly(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale")

	if err := os.WriteFile(logPath+stampSuffix, []byte(now.Add(-10*time.Minute).Format(time.RFC3339Nano)), 0o644); err != nil {
		t.Fatalf("write stamp: %v", err)
	}
	orphan := filepath.Join(dir, "usage-log-444444.tmp")
	writeTmpAged(t, orphan, now, 48*time.Hour)

	if err := NewRetention(logPath, filepath.Join(dir, "archive"), func() time.Time { return now }).PruneStaleEntries(30); err != nil {
		t.Fatalf("prune: %v", err)
	}
	if !pathExists(orphan) {
		t.Fatalf("a process that skipped on a fresh stamp swept the directory")
	}

	// The next cycle (stamp expired) is the sweeper.
	later := now.Add(2 * time.Hour)
	if err := NewRetention(logPath, filepath.Join(dir, "archive"), func() time.Time { return later }).PruneStaleEntries(30); err != nil {
		t.Fatalf("next-cycle prune: %v", err)
	}
	if pathExists(orphan) {
		t.Fatalf("the next prune cycle did not sweep the old orphan")
	}
}

// TestPruneSweepAgeBoundary pins the threshold between "may belong to a live rewrite" and
// "orphan": 9 minutes old is kept, 11 minutes old is swept.
func TestPruneSweepAgeBoundary(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale")
	nine := filepath.Join(dir, "usage-log-nine.tmp")
	eleven := filepath.Join(dir, "usage-log-eleven.tmp")
	writeTmpAged(t, nine, now, 9*time.Minute)
	writeTmpAged(t, eleven, now, 11*time.Minute)

	if err := NewRetention(logPath, filepath.Join(dir, "archive"), func() time.Time { return now }).PruneStaleEntries(30); err != nil {
		t.Fatalf("prune: %v", err)
	}
	if !pathExists(nine) {
		t.Errorf("a 9-minute-old tmp was swept")
	}
	if pathExists(eleven) {
		t.Errorf("an 11-minute-old tmp survived")
	}
}

// TestPruneSweepRunsWhenPruneFails: the sweep is housekeeping that does not depend on the prune
// succeeding; a prune that errors (here: the archive directory path is occupied by a file) still
// returns its error and still sweeps.
func TestPruneSweepRunsWhenPruneFails(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	archiveDir := filepath.Join(dir, "archive")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale")
	if err := os.WriteFile(archiveDir, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("occupy archive path: %v", err)
	}
	orphan := filepath.Join(dir, "usage-log-555555.tmp")
	writeTmpAged(t, orphan, now, time.Hour)

	if err := NewRetention(logPath, archiveDir, func() time.Time { return now }).PruneStaleEntries(30); err == nil {
		t.Fatalf("prune with an unusable archive directory returned no error")
	}
	if pathExists(orphan) {
		t.Fatalf("the orphan survived a prune that failed")
	}
}
