package harness

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPruneStampShorterOverLongerIsExact(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC) // formats without fractional digits
	writeStaleLog(t, logPath, now, "stale-1")
	statePath := logPath + stampSuffix
	old := "2026-09-01T00:00:00.123456789Z" // nine fractional digits, older than the interval
	if err := os.WriteFile(statePath, []byte(old), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := NewRetention(logPath, filepath.Join(dir, "archive"), func() time.Time { return now }).PruneStaleEntries(30); err != nil {
		t.Fatalf("prune: %v", err)
	}
	want := now.UTC().Format(time.RFC3339Nano)
	got, _ := os.ReadFile(statePath)
	if string(got) != want {
		t.Errorf("state file = %q, want exactly %q", got, want)
	}
	if !stampIsFresh(got, now.Add(time.Minute)) {
		t.Errorf("state file does not parse as a fresh stamp: %q", got)
	}
}
