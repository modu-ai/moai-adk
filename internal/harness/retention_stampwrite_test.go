//go:build !windows

package harness

// Pins the stamp-write failure path of the locked prune phase (SPEC-HARNESS-
// RETENTION-HARDEN-001, AC-HRH-014): when the attempt stamp cannot be recorded
// the prune must be skipped. A state file opened read-only makes the stamp write
// fail without any file-system fault. The file asserts POSIX read-only-handle
// behaviour and therefore carries the !windows build constraint.

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPruneStampWriteFailureSkipsPrune(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	archiveDir := filepath.Join(dir, "archive")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")
	statePath := logPath + stampSuffix
	old := []byte("2026-09-01T00:00:00Z") // older than the interval, so the stamp re-check does not skip
	if err := os.WriteFile(statePath, old, 0o644); err != nil {
		t.Fatalf("seed state: %v", err)
	}
	sf, err := os.Open(statePath) // read-only: Truncate and WriteAt fail
	if err != nil {
		t.Fatalf("open state read-only: %v", err)
	}
	defer func() { _ = sf.Close() }()
	logBefore, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}

	perr := NewRetention(logPath, archiveDir, func() time.Time { return now }).pruneLocked(sf, 30)
	if perr == nil {
		t.Errorf("pruneLocked returned nil, want the stamp-write error")
	}
	if logAfter, _ := os.ReadFile(logPath); !bytes.Equal(logBefore, logAfter) {
		t.Errorf("log changed although the stamp could not be written")
	}
	if _, serr := os.Stat(archiveDir); !os.IsNotExist(serr) {
		t.Errorf("archive directory exists although the stamp could not be written: %v", serr)
	}
	if got, _ := os.ReadFile(statePath); !bytes.Equal(got, old) {
		t.Errorf("state file changed: %q", got)
	}
}
