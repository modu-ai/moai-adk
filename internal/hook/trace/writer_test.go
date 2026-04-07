package trace

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// makeEntry builds a test TraceEntry with the given event name.
func makeEntry(event string) TraceEntry {
	return TraceEntry{
		Timestamp:  time.Now(),
		Event:      event,
		Handler:    "testHandler",
		DurationMs: 10,
		Decision:   "allow",
		SessionID:  "sess-test",
	}
}

func TestTraceWriter_Write_BasicJSONLine(t *testing.T) {
	dir := t.TempDir()
	w := NewTraceWriter(dir, "basic")
	defer w.Close()

	entry := TraceEntry{
		Timestamp:  time.Date(2026, 4, 7, 10, 30, 0, 0, time.UTC),
		Event:      "PreToolUse",
		Handler:    "preToolHandler",
		Tool:       "Bash",
		DurationMs: 45,
		Decision:   "allow",
		SessionID:  "sess-basic",
	}
	w.Write(entry)
	// Close to flush.
	_ = w.Close()

	data, err := os.ReadFile(filepath.Join(dir, "trace-basic.jsonl"))
	if err != nil {
		t.Fatalf("read trace file: %v", err)
	}

	lines := nonEmptyLines(string(data))
	if len(lines) != 1 {
		t.Fatalf("want 1 line, got %d", len(lines))
	}

	var got TraceEntry
	if err := json.Unmarshal([]byte(lines[0]), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Event != "PreToolUse" {
		t.Errorf("event: want PreToolUse, got %q", got.Event)
	}
	if got.Tool != "Bash" {
		t.Errorf("tool: want Bash, got %q", got.Tool)
	}
	if got.DurationMs != 45 {
		t.Errorf("duration_ms: want 45, got %d", got.DurationMs)
	}
	if got.Decision != "allow" {
		t.Errorf("decision: want allow, got %q", got.Decision)
	}
	if got.SessionID != "sess-basic" {
		t.Errorf("session_id: want sess-basic, got %q", got.SessionID)
	}
}

func TestTraceWriter_AsyncNonBlocking(t *testing.T) {
	dir := t.TempDir()
	w := NewTraceWriter(dir, "async")

	start := time.Now()
	// Write 10 entries; each should return immediately.
	for range 10 {
		w.Write(makeEntry("TestEvent"))
	}
	elapsed := time.Since(start)

	// Writes should complete in under 50ms regardless of disk speed.
	if elapsed > 50*time.Millisecond {
		t.Errorf("Write() took %v, expected non-blocking", elapsed)
	}

	_ = w.Close()
}

func TestTraceWriter_MultipleEntries(t *testing.T) {
	dir := t.TempDir()
	w := NewTraceWriter(dir, "multi")

	for i := range 5 {
		w.Write(TraceEntry{
			Timestamp:  time.Now(),
			Event:      "Event",
			Handler:    "h",
			DurationMs: int64(i),
			SessionID:  "multi",
		})
	}
	_ = w.Close()

	f, err := os.Open(filepath.Join(dir, "trace-multi.jsonl"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()

	count := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if strings.TrimSpace(sc.Text()) != "" {
			count++
		}
	}
	if count != 5 {
		t.Errorf("want 5 lines, got %d", count)
	}
}

func TestTraceWriter_FileRotation(t *testing.T) {
	dir := t.TempDir()
	w := NewTraceWriter(dir, "rotate")
	// Set a tiny maxSize to trigger rotation.
	w.maxSize = 50

	// Write enough entries to exceed 50 bytes.
	for range 5 {
		w.Write(makeEntry("Rotate"))
	}
	_ = w.Close()

	// The rotated file should exist.
	rotated := filepath.Join(dir, "trace-rotate.1.jsonl")
	if _, err := os.Stat(rotated); err != nil {
		t.Errorf("rotated file not found: %v", err)
	}
}

func TestTraceWriter_GracefulDegradation_UnwritableDir(t *testing.T) {
	dir := t.TempDir()
	// Make the directory read-only so file creation fails.
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Skip("cannot chmod dir (may be running as root)")
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	w := NewTraceWriter(dir, "unwritable")
	// Write should not panic or block; errors are logged internally.
	w.Write(makeEntry("TestEvent"))
	// Close should complete without returning an error to the caller.
	if err := w.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}

func TestTraceWriter_Close_FlushPending(t *testing.T) {
	dir := t.TempDir()
	w := NewTraceWriter(dir, "flush")

	const count = 20
	for range count {
		w.Write(makeEntry("Flush"))
	}
	// Close must drain all pending entries.
	_ = w.Close()

	data, err := os.ReadFile(filepath.Join(dir, "trace-flush.jsonl"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	got := len(nonEmptyLines(string(data)))
	if got != count {
		t.Errorf("want %d lines after Close, got %d", count, got)
	}
}

func TestTraceWriter_ConcurrentWrites(t *testing.T) {
	dir := t.TempDir()
	w := NewTraceWriter(dir, "concurrent")

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w.Write(makeEntry("Concurrent"))
		}()
	}
	wg.Wait()
	_ = w.Close()

	data, _ := os.ReadFile(filepath.Join(dir, "trace-concurrent.jsonl"))
	// All 50 writes should have been queued and persisted (channel capacity 100).
	lines := nonEmptyLines(string(data))
	if len(lines) != 50 {
		t.Errorf("want 50 lines, got %d", len(lines))
	}
}

func TestTraceWriter_CloseTwice(t *testing.T) {
	dir := t.TempDir()
	w := NewTraceWriter(dir, "twice")
	// Second Close must not panic.
	_ = w.Close()
	_ = w.Close()
}

// nonEmptyLines splits s by newline and returns non-empty lines.
func nonEmptyLines(s string) []string {
	var result []string
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}
	return result
}
