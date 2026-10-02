package harness

// Contract tests for the prune stamp ordering, the failure paths and the lock
// behaviour (card t1425, sync-audit F1-F3). Each test pins one decision of
// pruneExclusive so that a wrong variant fails here instead of surviving.

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/lockfile"
)

// failingArchiveDir returns an archive directory that cannot be created (its
// parent is a regular file), so the archive step of a prune always fails.
func failingArchiveDir(t *testing.T, dir string) string {
	t.Helper()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	return filepath.Join(blocker, "archive")
}

// waitOrFail runs fn in a goroutine and fails the test if it does not return
// within the bound. It returns fn's error.
func waitOrFail(t *testing.T, bound time.Duration, what string, fn func() error) error {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- fn() }()
	select {
	case err := <-done:
		return err
	case <-time.After(bound):
		t.Fatalf("%s did not return within %s", what, bound)
		return nil
	}
}

// TestPruneStamp_FailedPruneStillStampsAndIsNotRetried: an attempt whose prune
// FAILS keeps its stamp, returns the error, and a new instance inside the
// interval does not retry (the original storm after a persistent failure).
func TestPruneStamp_FailedPruneStillStampsAndIsNotRetried(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	archiveDir := failingArchiveDir(t, dir)
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")

	first := NewRetention(logPath, archiveDir, func() time.Time { return now })
	if err := first.PruneStaleEntries(30); err == nil {
		t.Fatalf("expected the archive failure to be returned")
	}
	stamp := readStampFile(logPath + stampSuffix)
	if !stampIsFresh(stamp, now.Add(time.Minute)) {
		t.Fatalf("a failed prune left no fresh stamp on disk: %q", stamp)
	}
	if !logHasSubject(t, logPath, "stale-1") {
		t.Fatalf("a failed prune must leave the log untouched")
	}

	// Inside the interval a new instance must skip: no error, no retry.
	second := NewRetention(logPath, archiveDir, func() time.Time { return now.Add(10 * time.Minute) })
	if err := second.PruneStaleEntries(30); err != nil {
		t.Fatalf("a new instance retried a failed prune inside the interval: %v", err)
	}

	// After the interval the prune is attempted again (and fails again).
	third := NewRetention(logPath, archiveDir, func() time.Time { return now.Add(61 * time.Minute) })
	if err := third.PruneStaleEntries(30); err == nil {
		t.Fatalf("an instance past the interval did not retry")
	}
}

// TestPruneStamp_StateNotOpenableSkipsPruneAndRecordSucceeds: the state path
// is a directory (open for write fails) while the log directory stays
// writable. The prune must be skipped, never done unlocked, the error is
// returned, and RecordEvent still succeeds.
func TestPruneStamp_StateNotOpenableSkipsPruneAndRecordSucceeds(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	archiveDir := filepath.Join(dir, "archive")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")
	if err := os.Mkdir(logPath+stampSuffix, 0o755); err != nil {
		t.Fatalf("mkdir state path: %v", err)
	}

	r := NewRetention(logPath, archiveDir, func() time.Time { return now })
	if err := r.PruneStaleEntries(30); err == nil {
		t.Fatalf("expected an error when the state file cannot be opened")
	}
	if !logHasSubject(t, logPath, "stale-1") {
		t.Fatalf("the log was pruned without the lock")
	}
	if _, err := os.Stat(archiveDir); !os.IsNotExist(err) {
		t.Fatalf("an archive was written without the lock: err=%v", err)
	}

	obs := NewObserverWithRetention(logPath, NewRetention(logPath, archiveDir, func() time.Time { return now }))
	if err := obs.RecordEvent(EventTypeFeedback, "after", "ctx"); err != nil {
		t.Fatalf("RecordEvent failed because of the prune error: %v", err)
	}
}

// TestPruneStamp_FreshStampNeedsNoLock: with a fresh stamp on disk the hot
// path must return without taking the lock, even while another holder owns it.
func TestPruneStamp_FreshStampNeedsNoLock(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")
	statePath := logPath + stampSuffix
	if err := os.WriteFile(statePath, []byte(now.Add(-10*time.Minute).Format(time.RFC3339Nano)), 0o644); err != nil {
		t.Fatalf("write stamp: %v", err)
	}

	// A second open file description holds the exclusive lock for the whole call.
	holder, err := os.OpenFile(statePath, os.O_RDWR, 0o644)
	if err != nil {
		t.Fatalf("open holder: %v", err)
	}
	defer func() { _ = holder.Close() }()
	if err := lockfile.Lock(holder); err != nil {
		t.Fatalf("lock holder: %v", err)
	}
	unlocked := false
	unlock := func() {
		if !unlocked {
			unlocked = true
			_ = lockfile.Unlock(holder)
		}
	}
	defer unlock()

	r := NewRetention(logPath, filepath.Join(dir, "archive"), func() time.Time { return now })
	err = waitOrFail(t, 2*time.Second, "PruneStaleEntries behind a held lock with a fresh stamp", func() error {
		return r.PruneStaleEntries(30)
	})
	unlock() // lets a wrongly blocked call finish before the temp dir is removed
	if err != nil {
		t.Fatalf("hot path returned an error: %v", err)
	}
	if !logHasSubject(t, logPath, "stale-1") {
		t.Fatalf("the log was pruned inside the interval")
	}
}

// TestPruneStamp_RecheckUsesFreshClock: the clock read before the lock sees the
// stamp as a future one (expired), the clock read after the lock shows it is
// fresh. The re-check must use the post-lock reading and so must not prune.
func TestPruneStamp_RecheckUsesFreshClock(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	stampTime := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, stampTime, "stale-1")
	if err := os.WriteFile(logPath+stampSuffix, []byte(stampTime.Format(time.RFC3339Nano)), 0o644); err != nil {
		t.Fatalf("write stamp: %v", err)
	}

	calls := 0
	clock := func() time.Time {
		calls++
		if calls == 1 {
			return stampTime.Add(-5 * time.Minute) // pre-lock: the stamp looks like the future
		}
		return stampTime.Add(5 * time.Minute) // post-lock: the stamp is 5 minutes old
	}
	r := NewRetention(logPath, filepath.Join(dir, "archive"), clock)
	if err := r.PruneStaleEntries(30); err != nil {
		t.Fatalf("PruneStaleEntries: %v", err)
	}
	if calls != 2 {
		t.Errorf("expected exactly two clock readings (pre-lock, post-lock), got %d", calls)
	}
	if !logHasSubject(t, logPath, "stale-1") {
		t.Fatalf("pruned although the post-lock reading shows a fresh stamp")
	}
}
