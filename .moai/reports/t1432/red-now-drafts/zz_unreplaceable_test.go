package harness

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// DRAFT of the AC-HRH-003 case (b): an unreadable state file in a directory
// the current user cannot write to. The pruner cannot replace it, so it skips.
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
	logBefore, _ := os.ReadFile(logPath)
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
	if got, _ := os.ReadFile(statePath); !bytes.Equal(got, old) {
		t.Errorf("state file changed: %q", got)
	}
}
