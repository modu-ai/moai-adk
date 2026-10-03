//go:build !windows

package harness

// Tests for events appended while a prune is in flight (SPEC-HARNESS-
// RETENTION-HARDEN-001, AC-HRH-007 and AC-HRH-008 cases b and c). The pruner
// is held in its archive step by making the month archive path a FIFO, so a
// second writer can append after the pruner has read the log. The file uses
// syscall.Mkfifo and therefore carries the !windows build constraint.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// blockedPruner starts a pruner whose archive step blocks on a FIFO and
// returns once the stamp is on disk and a short delay has let the pruner read
// the log. release() unblocks the pruner and returns the prune error.
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
	var once sync.Once
	var result error
	release = func() error {
		once.Do(func() { result = releaseBlockedPruner(t, fifo, done) })
		return result
	}
	// A test that fails before calling release still unblocks the pruner and the FIFO.
	t.Cleanup(func() { _ = release() })
	return release
}

// releaseBlockedPruner drains the FIFO the pruner's archive step writes to and returns the prune
// error. The whole wait is bounded: the blocking read open runs in a goroutine, and if the pruner
// returns without ever opening the FIFO for writing (it failed before its archive step) or does not
// reach that step in time, the reader is released by a non-blocking write open (which succeeds only
// while a reader is blocked in its open), so the real error is reported instead of a test timeout.
func releaseBlockedPruner(t *testing.T, fifo string, done <-chan error) error {
	t.Helper()
	const wait = 10 * time.Second
	read := make(chan struct{})
	go func() {
		defer close(read)
		f, err := os.OpenFile(fifo, os.O_RDONLY, 0)
		if err != nil {
			return
		}
		_, _ = io.Copy(io.Discard, f)
		_ = f.Close()
	}()
	unblockReader := func() {
		tick := time.NewTicker(10 * time.Millisecond)
		defer tick.Stop()
		for {
			if w, err := os.OpenFile(fifo, os.O_WRONLY|syscall.O_NONBLOCK, 0); err == nil {
				_ = w.Close()
			}
			select {
			case <-read:
				return
			case <-tick.C:
			}
		}
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case perr := <-done:
		unblockReader()
		return perr
	case <-read:
	case <-timer.C:
		unblockReader()
		t.Errorf("the pruner did not reach its archive step within %v", wait)
		return fmt.Errorf("pruner did not reach its archive step within %v", wait)
	}
	select {
	case perr := <-done:
		return perr
	case <-timer.C:
		t.Errorf("the pruner did not finish within %v of the release", wait)
		return fmt.Errorf("pruner did not finish within %v", wait)
	}
}

// TestPruneCarriesLateEvents: an event appended while the pruner is blocked in
// its archive step survives the rewrite, after the fresh event and verbatim.
func TestPruneCarriesLateEvents(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")
	release := blockedPruner(t, dir, now, logPath)
	lateEvt := Event{Timestamp: now, EventType: EventTypeFeedback, Subject: "late-event", ContextHash: "h", SchemaVersion: LogSchemaVersion}
	if err := appendEventsJSONL(logPath, []Event{lateEvt}); err != nil {
		t.Fatalf("append late: %v", err)
	}
	if err := release(); err != nil {
		t.Errorf("pruner returned %v, want nil", err)
	}
	subjects := logSubjectsForUnit(t, logPath)
	if subjects["late-event"] != 1 {
		t.Errorf("late-event count = %d, want 1 (subjects after prune: %v)", subjects["late-event"], subjects)
	}
	if subjects["fresh"] != 1 {
		t.Errorf("fresh count = %d, want 1", subjects["fresh"])
	}
	if subjects["stale-1"] != 0 {
		t.Errorf("stale-1 count = %d, want 0", subjects["stale-1"])
	}
	lines := readLogLines(t, logPath)
	if len(lines) != 2 || !strings.Contains(lines[0], `"fresh"`) || lines[1] != marshalEvent(t, lateEvt) {
		t.Errorf("want fresh then the late event verbatim, got %q", lines)
	}
}

// TestPruneTailAlreadyTerminatedGetsNoExtraNewline compares raw bytes: a late event that already
// ends with a newline is carried exactly, so the replacement log is the kept line followed by the
// late bytes and nothing else (no blank line is added after the tail).
func TestPruneTailAlreadyTerminatedGetsNoExtraNewline(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")
	original, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read original log: %v", err)
	}
	staleLine, rest, found := bytes.Cut(original, []byte("\n"))
	if !found || len(staleLine) == 0 || !bytes.HasSuffix(rest, []byte("\n")) {
		t.Fatalf("unexpected original log layout: %q", original)
	}
	release := blockedPruner(t, dir, now, logPath)
	lateEvt := Event{Timestamp: now, EventType: EventTypeFeedback, Subject: "late-event", ContextHash: "h", SchemaVersion: LogSchemaVersion}
	if err := appendEventsJSONL(logPath, []Event{lateEvt}); err != nil {
		t.Fatalf("append late: %v", err)
	}
	appended, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read appended log: %v", err)
	}
	late := appended[len(original):]
	if len(late) == 0 || late[len(late)-1] != '\n' {
		t.Fatalf("precondition: the late event must already end with a newline, got %q", late)
	}
	if err := release(); err != nil {
		t.Errorf("pruner returned %v, want nil", err)
	}
	got, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read pruned log: %v", err)
	}
	want := append(append([]byte(nil), rest...), late...)
	if !bytes.Equal(got, want) {
		t.Errorf("pruned log bytes = %q, want %q (the kept line plus the late bytes, no extra blank line)", got, want)
	}
}

// TestPruneTailPartialLineCarriedAndTerminated: a fragment without a
// terminating newline that arrives during the prune is carried exactly once,
// the replacement log ends with a newline, and the next append starts on its
// own line.
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
		t.Errorf("pruner returned %v, want nil", err)
	}
	data, rerr := os.ReadFile(logPath)
	if rerr != nil {
		t.Fatalf("read log: %v", rerr)
	}
	if n := bytes.Count(data, []byte(partial)); n != 1 {
		t.Errorf("partial fragment occurs %d times, want 1", n)
	}
	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Errorf("replacement log does not end with a newline")
	}
	after := Event{Timestamp: now, EventType: EventTypeFeedback, Subject: "after-prune", ContextHash: "h", SchemaVersion: LogSchemaVersion}
	if err := appendEventsJSONL(logPath, []Event{after}); err != nil {
		t.Fatalf("append after prune: %v", err)
	}
	lines := readLogLines(t, logPath)
	last := lines[len(lines)-1]
	var parsed Event
	if err := json.Unmarshal([]byte(last), &parsed); err != nil || parsed.Subject != "after-prune" {
		t.Errorf("the next append is not on its own line and parseable: last=%q err=%v", last, err)
	}
}

// TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval: a stale event that is
// the final line of the log with no terminating newline is not classified in
// this prune (it may be a write still in flight); the rewrite keeps it and the
// next interval's prune archives it.
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
		t.Fatalf("open log: %v", err)
	}
	if _, err := f.WriteString(raw); err != nil {
		t.Fatalf("write final line: %v", err)
	}
	_ = f.Close()
	if err := NewRetention(logPath, archiveDir, func() time.Time { return now }).PruneStaleEntries(30); err != nil {
		t.Errorf("prune returned %v, want nil", err)
	}
	subjects := logSubjectsForUnit(t, logPath)
	if subjects["stale-final"] != 1 {
		t.Errorf("stale-final count after first prune = %d, want 1 (kept in the log this interval; subjects: %v)", subjects["stale-final"], subjects)
	}
	if subjects["stale-1"] != 0 {
		t.Errorf("stale-1 count = %d, want 0", subjects["stale-1"])
	}
	if data, rerr := os.ReadFile(logPath); rerr != nil || len(data) == 0 || data[len(data)-1] != '\n' {
		t.Errorf("rewritten log does not end with a newline: err=%v", rerr)
	}
	later := now.Add(61 * time.Minute)
	if err := NewRetention(logPath, archiveDir, func() time.Time { return later }).PruneStaleEntries(30); err != nil {
		t.Errorf("second prune returned %v, want nil", err)
	}
	if logHasSubject(t, logPath, "stale-final") {
		t.Errorf("stale-final still in the log after the next interval")
	}
}

// N2: a prune with no late event leaves the replacement log exactly the kept lines: it does not end
// with a blank line (the log reader drops blank lines, so only a raw count sees it).
func TestPruneWithoutLateEventAddsNoBlankLine(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")
	if err := NewRetention(logPath, filepath.Join(dir, "archive"), func() time.Time { return now }).PruneStaleEntries(30); err != nil {
		t.Fatalf("prune returned %v, want nil", err)
	}
	got, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if n := strings.Count(string(got), "\n"); n != 1 || !bytes.HasSuffix(got, []byte("}\n")) {
		t.Errorf("replacement log = %q: want exactly one newline-terminated kept line and no blank line", got)
	}
}
