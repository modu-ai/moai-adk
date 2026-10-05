package harness

// Unit tests for the on-disk prune stamp (card t1425). They use only the
// public Retention API and the documented "<log>.prune-state" file name, so
// they can be compiled against the pre-fix code to observe their RED state.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

const stampSuffix = ".prune-state"

// writeStaleLog writes one stale and one fresh event relative to now.
func writeStaleLog(t *testing.T, logPath string, now time.Time, staleSubject string) {
	t.Helper()
	events := []Event{
		{Timestamp: now.AddDate(0, 0, -40), EventType: EventTypeMoaiSubcommand, Subject: staleSubject, ContextHash: "h", SchemaVersion: LogSchemaVersion},
		{Timestamp: now.AddDate(0, 0, -1), EventType: EventTypeFeedback, Subject: "fresh", ContextHash: "h", SchemaVersion: LogSchemaVersion},
	}
	if err := writeEventsToFile(logPath, events); err != nil {
		t.Fatalf("write log: %v", err)
	}
}

func logHasSubject(t *testing.T, logPath, subject string) bool {
	t.Helper()
	return logSubjectsForUnit(t, logPath)[subject] > 0
}

func logSubjectsForUnit(t *testing.T, logPath string) map[string]int {
	t.Helper()
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	out := map[string]int{}
	for _, line := range splitNonEmptyLines(string(data)) {
		var evt Event
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			t.Fatalf("parse log line: %v", err)
		}
		out[evt.Subject]++
	}
	return out
}

// TestPruneStamp_PersistedAcrossInstances: a brand-new Retention (a new hook
// process) inside the skip interval must not prune; after the interval it must.
func TestPruneStamp_PersistedAcrossInstances(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	archiveDir := filepath.Join(dir, "archive")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")

	if err := NewRetention(logPath, archiveDir, func() time.Time { return now }).PruneStaleEntries(30); err != nil {
		t.Fatalf("first prune: %v", err)
	}
	if logHasSubject(t, logPath, "stale-1") {
		t.Fatalf("first prune left the stale event in the log")
	}
	if _, err := os.Stat(logPath + stampSuffix); err != nil {
		t.Fatalf("stamp file not written: %v", err)
	}

	// New stale event appears; a new instance 10 minutes later must skip.
	writeStaleLog(t, logPath, now, "stale-2")
	later := now.Add(10 * time.Minute)
	if err := NewRetention(logPath, archiveDir, func() time.Time { return later }).PruneStaleEntries(30); err != nil {
		t.Fatalf("second prune: %v", err)
	}
	if !logHasSubject(t, logPath, "stale-2") {
		t.Fatalf("a fresh instance inside the skip interval pruned: stamp not respected")
	}

	// 61 minutes after the first prune the stamp has expired.
	expired := now.Add(61 * time.Minute)
	if err := NewRetention(logPath, archiveDir, func() time.Time { return expired }).PruneStaleEntries(30); err != nil {
		t.Fatalf("third prune: %v", err)
	}
	if logHasSubject(t, logPath, "stale-2") {
		t.Fatalf("an instance past the skip interval did not prune")
	}
}

// TestPruneStamp_FutureStampTreatedAsExpired: a stamp ahead of the clock
// (clock skew, restored backup) must not suppress pruning indefinitely.
func TestPruneStamp_FutureStampTreatedAsExpired(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")
	future := now.Add(48 * time.Hour).Format(time.RFC3339Nano)
	if err := os.WriteFile(logPath+stampSuffix, []byte(future), 0o644); err != nil {
		t.Fatalf("write stamp: %v", err)
	}

	if err := NewRetention(logPath, filepath.Join(dir, "archive"), func() time.Time { return now }).PruneStaleEntries(30); err != nil {
		t.Fatalf("prune: %v", err)
	}
	if logHasSubject(t, logPath, "stale-1") {
		t.Fatalf("a future stamp suppressed pruning")
	}
}

// TestPruneStamp_CorruptOrPartialStampIgnored: garbage, empty, and truncated
// stamps count as "no stamp" and must not block pruning or return an error.
func TestPruneStamp_CorruptOrPartialStampIgnored(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	full := now.Add(-time.Minute).Format(time.RFC3339Nano)
	cases := map[string]string{
		"garbage":   "not a timestamp\n",
		"empty":     "",
		"truncated": full[:len(full)-4],
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			logPath := filepath.Join(dir, "usage-log.jsonl")
			writeStaleLog(t, logPath, now, "stale-1")
			if err := os.WriteFile(logPath+stampSuffix, []byte(content), 0o644); err != nil {
				t.Fatalf("write stamp: %v", err)
			}
			if err := NewRetention(logPath, filepath.Join(dir, "archive"), func() time.Time { return now }).PruneStaleEntries(30); err != nil {
				t.Fatalf("prune with %s stamp: %v", name, err)
			}
			if logHasSubject(t, logPath, "stale-1") {
				t.Fatalf("a %s stamp suppressed pruning", name)
			}
		})
	}
}

// TestPruneStamp_StampWrittenWhenNothingToPrune: an attempt that finds no
// stale event still stamps, so the next process skips the log read.
func TestPruneStamp_StampWrittenWhenNothingToPrune(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	if err := writeEventsToFile(logPath, []Event{
		{Timestamp: now.AddDate(0, 0, -1), EventType: EventTypeFeedback, Subject: "fresh", ContextHash: "h", SchemaVersion: LogSchemaVersion},
	}); err != nil {
		t.Fatalf("write log: %v", err)
	}
	if err := NewRetention(logPath, filepath.Join(dir, "archive"), func() time.Time { return now }).PruneStaleEntries(30); err != nil {
		t.Fatalf("prune: %v", err)
	}
	if _, err := os.Stat(logPath + stampSuffix); err != nil {
		t.Fatalf("no stamp after a no-op prune: %v", err)
	}
}

// TestPruneStamp_RecheckUnderLockSkips pins the double check: a process that
// won the lock only after another process stamped (the stamp appeared between
// its lock-free read and the lock) must return without touching the log.
func TestPruneStamp_RecheckUnderLockSkips(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")
	if err := os.WriteFile(logPath+stampSuffix, []byte(now.Format(time.RFC3339Nano)), 0o644); err != nil {
		t.Fatalf("write stamp: %v", err)
	}

	r := NewRetention(logPath, filepath.Join(dir, "archive"), func() time.Time { return now })
	if err := r.pruneExclusive(logPath+stampSuffix, 30); err != nil {
		t.Fatalf("pruneExclusive: %v", err)
	}
	if !logHasSubject(t, logPath, "stale-1") {
		t.Fatalf("pruneExclusive pruned although a fresh stamp was already on disk")
	}
}

// TestPruneStamp_MissingLogCreatesNoState: with no log there is nothing to
// prune; the call must succeed without creating a state file (the directory
// may not even exist).
func TestPruneStamp_MissingLogCreatesNoState(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "no-such-dir", "usage-log.jsonl")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	if err := NewRetention(logPath, filepath.Join(dir, "archive"), func() time.Time { return now }).PruneStaleEntries(30); err != nil {
		t.Fatalf("prune with missing log: %v", err)
	}
	if _, err := os.Stat(logPath + stampSuffix); !os.IsNotExist(err) {
		t.Fatalf("state file exists for a missing log: err=%v", err)
	}
}

// TestPruneStamp_UnwritableStateSkipsPruneAndRecordSucceeds: when the state
// file cannot be created the prune is skipped with an error (never an
// unguarded prune, which could rebuild the storm), the log is untouched, and
// RecordEvent still succeeds because the observer ignores the prune error.
func TestPruneStamp_UnwritableStateSkipsPruneAndRecordSucceeds(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permission bits are not enforced on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")

	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	r := NewRetention(logPath, filepath.Join(dir, "archive"), func() time.Time { return now })
	if err := r.PruneStaleEntries(30); err == nil {
		t.Fatalf("expected an error when the state file cannot be created")
	}
	if !logHasSubject(t, logPath, "stale-1") {
		t.Fatalf("log was pruned without the lock")
	}

	obs := NewObserverWithRetention(logPath, NewRetention(logPath, filepath.Join(dir, "archive"), func() time.Time { return now }))
	if err := obs.RecordEvent(EventTypeFeedback, "after", "ctx"); err != nil {
		t.Fatalf("RecordEvent failed because of the prune error: %v", err)
	}
}
