//go:build !windows

package harness

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// blockedPruner starts a pruner whose archive step is blocked on a FIFO and
// returns once the stamp is on disk and a short delay has let the pruner read
// the log. release() unblocks it and returns the prune error.
func blockedPruner(t *testing.T, dir string, now time.Time, logPath string) (release func() error) {
	t.Helper()
	archiveDir := filepath.Join(dir, "archive")
	month := now.AddDate(0, 0, -40).UTC().Format("2006-01")
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		t.Fatalf("mkdir archive: %v", err)
	}
	fifo := filepath.Join(archiveDir, month+".jsonl.gz")
	if err := syscall.Mkfifo(fifo, 0o644); err != nil {
		t.Skipf("mkfifo unavailable: %v", err)
	}
	done := make(chan error, 1)
	go func() {
		done <- NewRetention(logPath, archiveDir, func() time.Time { return now }).PruneStaleEntries(30)
	}()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && !stampIsFresh(readStampFile(logPath+stampSuffix), now) {
		time.Sleep(time.Millisecond)
	}
	time.Sleep(300 * time.Millisecond)
	return func() error {
		f, err := os.OpenFile(fifo, os.O_RDONLY, 0)
		if err == nil {
			_, _ = io.Copy(io.Discard, f)
			_ = f.Close()
		}
		select {
		case perr := <-done:
			return perr
		case <-time.After(10 * time.Second):
			t.Fatalf("pruner did not finish")
			return nil
		}
	}
}

func TestPruneCarriesLateEvents(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")
	release := blockedPruner(t, dir, now, logPath)
	if err := appendEventsJSONL(logPath, []Event{{Timestamp: now, EventType: EventTypeFeedback, Subject: "late-event", ContextHash: "h", SchemaVersion: LogSchemaVersion}}); err != nil {
		t.Fatalf("append late: %v", err)
	}
	if err := release(); err != nil {
		t.Errorf("pruner returned %v", err)
	}
	subjects := logSubjectsForUnit(t, logPath)
	t.Logf("subjects after prune: %v", subjects)
	if subjects["late-event"] != 1 {
		t.Errorf("late-event count = %d, want 1", subjects["late-event"])
	}
	if subjects["fresh"] != 1 {
		t.Errorf("fresh count = %d, want 1", subjects["fresh"])
	}
	if subjects["stale-1"] != 0 {
		t.Errorf("stale-1 count = %d, want 0", subjects["stale-1"])
	}
	lines := readLogLines(t, logPath)
	if len(lines) < 2 || !strings.Contains(lines[0], `"fresh"`) || !strings.Contains(lines[len(lines)-1], "late-event") {
		t.Errorf("order wrong: %v", lines)
	}
}

func TestPruneTailPartialLineCarriedAndTerminated(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")
	release := blockedPruner(t, dir, now, logPath)
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open log: %v", err)
	}
	const partial = `{"timestamp":"2026-10-02T00:00:00Z","subject":"par`
	if _, err := f.WriteString(partial); err != nil {
		t.Fatalf("write partial: %v", err)
	}
	_ = f.Close()
	if err := release(); err != nil {
		t.Errorf("pruner returned %v", err)
	}
	data, _ := os.ReadFile(logPath)
	if n := bytes.Count(data, []byte(partial)); n != 1 {
		t.Errorf("partial fragment occurs %d times, want 1", n)
	}
	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Errorf("replacement log does not end with a newline")
	}
	if err := appendEventsJSONL(logPath, []Event{{Timestamp: now, EventType: EventTypeFeedback, Subject: "after-prune", ContextHash: "h", SchemaVersion: LogSchemaVersion}}); err != nil {
		t.Fatalf("append after prune: %v", err)
	}
	lines := readLogLines(t, logPath)
	if last := lines[len(lines)-1]; !strings.Contains(last, `"after-prune"`) || strings.Contains(last, `"par`) {
		t.Errorf("the next append is not on its own line: last=%q", last)
	}
}

func TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	archiveDir := filepath.Join(dir, "archive")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")
	raw := marshalEvent(t, Event{Timestamp: now.AddDate(0, 0, -41), EventType: EventTypeMoaiSubcommand, Subject: "stale-final", ContextHash: "h", SchemaVersion: LogSchemaVersion})
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString(raw)
	_ = f.Close()
	if err := NewRetention(logPath, archiveDir, func() time.Time { return now }).PruneStaleEntries(30); err != nil {
		t.Errorf("prune returned %v", err)
	}
	subjects := logSubjectsForUnit(t, logPath)
	t.Logf("subjects after first prune: %v", subjects)
	if subjects["stale-final"] != 1 {
		t.Errorf("stale-final count after first prune = %d, want 1 (kept in the log this interval)", subjects["stale-final"])
	}
	if subjects["stale-1"] != 0 {
		t.Errorf("stale-1 count = %d, want 0", subjects["stale-1"])
	}
	later := now.Add(61 * time.Minute)
	if err := NewRetention(logPath, archiveDir, func() time.Time { return later }).PruneStaleEntries(30); err != nil {
		t.Errorf("second prune returned %v", err)
	}
	if logHasSubject(t, logPath, "stale-final") {
		t.Errorf("stale-final still in the log after the next interval")
	}
}
