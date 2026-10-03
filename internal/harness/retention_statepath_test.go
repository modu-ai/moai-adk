package harness

// Tests for the prune state-file path (SPEC-HARNESS-RETENTION-HARDEN-001):
// a symbolic link at the state path (AC-HRH-001), an unwritable state file the
// user owns (AC-HRH-002), and an entry that cannot be replaced (AC-HRH-003 b).
// They use only the public Retention API, the documented "<log>.prune-state"
// file name and the existing helpers, so they compile against the unmodified
// package.

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// TestPruneStateSymlinkReplacedTargetUntouched: a symbolic link owned by the
// current user at the state path is replaced by a regular state file, the link
// target is never written, and the prune completes.
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
	archive := filepath.Join(archiveDir, now.AddDate(0, 0, -40).UTC().Format("2006-01")+".jsonl.gz")
	if _, serr := os.Stat(archive); serr != nil {
		t.Errorf("stale event not archived: %v", serr)
	}
}

// TestPruneStateUnwritableFileReplaced: a state file of mode 0400 owned by the
// current user is replaced (a new file identity, not a permission change), and
// retention keeps working on a second instance one day later.
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
	// Hold the 0400 file open so its inode cannot be freed and reused by the
	// replacement: the identity comparison below is then decisive.
	held, err := os.Open(statePath)
	if err != nil {
		t.Fatalf("open held: %v", err)
	}
	defer func() { _ = held.Close() }()
	heldInfo, err := held.Stat()
	if err != nil {
		t.Fatalf("stat held: %v", err)
	}

	if err := NewRetention(logPath, archiveDir, func() time.Time { return now }).PruneStaleEntries(30); err != nil {
		t.Errorf("first prune returned %v, want nil", err)
	}
	if logHasSubject(t, logPath, "stale-1") {
		t.Errorf("first stale event still in the log")
	}

	later := now.AddDate(0, 0, 1)
	late := Event{Timestamp: later.AddDate(0, 0, -40), EventType: EventTypeMoaiSubcommand, Subject: "stale-2", ContextHash: "h", SchemaVersion: LogSchemaVersion}
	if err := appendEventsJSONL(logPath, []Event{late}); err != nil {
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
		return
	}
	defer func() { _ = f.Close() }()
	info, serr := f.Stat()
	if serr != nil {
		t.Errorf("stat state file: %v", serr)
		return
	}
	if !info.Mode().IsRegular() {
		t.Errorf("state path is not a regular file: %v", info.Mode())
	}
	if os.SameFile(heldInfo, info) {
		t.Errorf("state file is the same file as the 0400 original: replacement expected, not a permission change")
	}
}

// TestPruneStateUnreplaceableInReadOnlyDirSkips: an unwritable state file in a
// directory the user cannot write to cannot be replaced, so the prune skips
// with an error, leaves everything untouched, and RecordEvent still succeeds.
func TestPruneStateUnreplaceableInReadOnlyDirSkips(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permission bits are not enforced on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root bypasses permissions")
	}
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	archiveDir := filepath.Join(dir, "archive")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")
	statePath := logPath + stampSuffix
	old := []byte("2026-09-01T00:00:00Z")
	if err := os.WriteFile(statePath, old, 0o400); err != nil {
		t.Fatalf("write state: %v", err)
	}
	logBefore, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	stateBefore, err := os.Lstat(statePath)
	if err != nil {
		t.Fatalf("lstat state: %v", err)
	}
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatalf("chmod dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	if err := NewRetention(logPath, archiveDir, func() time.Time { return now }).PruneStaleEntries(30); err == nil {
		t.Errorf("expected an error: the state file cannot be opened or replaced")
	}
	logAfter, _ := os.ReadFile(logPath)
	if !bytes.Equal(logBefore, logAfter) {
		t.Errorf("log changed")
	}
	if _, err := os.Stat(archiveDir); !os.IsNotExist(err) {
		t.Errorf("archive directory exists: %v", err)
	}
	stateAfter, lerr := os.Lstat(statePath)
	if lerr != nil || !os.SameFile(stateBefore, stateAfter) || stateAfter.Mode() != stateBefore.Mode() {
		t.Errorf("state-path entry changed: err=%v", lerr)
	}
	if got, _ := os.ReadFile(statePath); !bytes.Equal(got, old) {
		t.Errorf("state file content changed: %q", got)
	}

	obs := NewObserverWithRetention(logPath, NewRetention(logPath, archiveDir, func() time.Time { return now }))
	if err := obs.RecordEvent(EventTypeFeedback, "after", "ctx"); err != nil {
		t.Errorf("RecordEvent failed because of the prune error: %v", err)
	}
}
