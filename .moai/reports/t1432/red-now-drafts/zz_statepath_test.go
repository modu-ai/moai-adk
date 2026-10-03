package harness

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// DRAFT of the M0 test for AC-HRH-001 (scratch, outside the tree).
func TestPruneStateSymlinkReplacedTargetUntouched(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symbolic links need privileges on Windows")
	}
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	archiveDir := filepath.Join(dir, "archive")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")

	victim := filepath.Join(dir, "victim.txt")
	victimBytes := bytes.Repeat([]byte("V"), 64)
	if err := os.WriteFile(victim, victimBytes, 0o644); err != nil {
		t.Fatalf("write victim: %v", err)
	}
	statePath := logPath + stampSuffix
	if err := os.Symlink(victim, statePath); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	err := NewRetention(logPath, archiveDir, func() time.Time { return now }).PruneStaleEntries(30)
	if err != nil {
		t.Errorf("PruneStaleEntries returned %v, want nil", err)
	}
	got, rerr := os.ReadFile(victim)
	if rerr != nil || !bytes.Equal(got, victimBytes) {
		t.Errorf("victim changed: err=%v content=%q", rerr, got)
	}
	fi, lerr := os.Lstat(statePath)
	if lerr != nil || !fi.Mode().IsRegular() {
		t.Errorf("state path is not a regular file: err=%v", lerr)
	}
	if !stampIsFresh(readStampFile(statePath), now.Add(time.Minute)) {
		t.Errorf("state path does not hold a fresh stamp: %q", readStampFile(statePath))
	}
	if logHasSubject(t, logPath, "stale-1") {
		t.Errorf("stale event still in the log")
	}
	if _, serr := os.Stat(filepath.Join(archiveDir, now.AddDate(0, 0, -40).UTC().Format("2006-01")+".jsonl.gz")); serr != nil {
		t.Errorf("stale event not archived: %v", serr)
	}
}

// DRAFT of the M0 test for AC-HRH-002.
func TestPruneStateUnwritableFileReplaced(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits are not enforced on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root bypasses file permissions")
	}
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	archiveDir := filepath.Join(dir, "archive")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")
	statePath := logPath + stampSuffix
	if err := os.WriteFile(statePath, []byte("2026-09-01T00:00:00Z"), 0o400); err != nil {
		t.Fatalf("write state: %v", err)
	}

	if err := NewRetention(logPath, archiveDir, func() time.Time { return now }).PruneStaleEntries(30); err != nil {
		t.Errorf("first prune returned %v, want nil", err)
	}
	if logHasSubject(t, logPath, "stale-1") {
		t.Errorf("first stale event still in the log")
	}

	later := now.AddDate(0, 0, 1)
	if err := appendEventsJSONL(logPath, []Event{{Timestamp: later.AddDate(0, 0, -40), EventType: EventTypeMoaiSubcommand, Subject: "stale-2", ContextHash: "h", SchemaVersion: LogSchemaVersion}}); err != nil {
		t.Fatalf("append stale-2: %v", err)
	}
	if err := NewRetention(logPath, archiveDir, func() time.Time { return later }).PruneStaleEntries(30); err != nil {
		t.Errorf("second prune returned %v, want nil", err)
	}
	if logHasSubject(t, logPath, "stale-2") {
		t.Errorf("second stale event still in the log")
	}
	f, oerr := os.OpenFile(statePath, os.O_RDWR, 0)
	if oerr != nil {
		t.Errorf("state file not openable read-write: %v", oerr)
	} else {
		_ = f.Close()
	}
}
