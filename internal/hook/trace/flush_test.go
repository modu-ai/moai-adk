package trace

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// generousFlushBudget is long enough that a healthy drain always completes
// within it, so a timeout in these tests signals a real defect, not flake.
const generousFlushBudget = 5 * time.Second

// tightFlushBudget is short enough that a blocked drain always exhausts it.
const tightFlushBudget = 20 * time.Millisecond

// undrainedFromLog scans JSON log lines for the undrained_entries field emitted
// when the flush budget is exhausted, and returns its value.
func undrainedFromLog(t *testing.T, logs string) int {
	t.Helper()
	for _, line := range strings.Split(logs, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue
		}
		v, ok := rec["undrained_entries"]
		if !ok {
			continue
		}
		n, ok := v.(float64)
		if !ok {
			t.Fatalf("undrained_entries is %T, want a number: %v", v, v)
		}
		return int(n)
	}
	t.Fatalf("no log record carried an undrained_entries field; logs were:\n%s", logs)
	return -1
}

// TestTraceWriterCloseWithTimeoutFlushesAll asserts that every entry enqueued
// before teardown reaches disk when the flush budget is generous (AC-HTF-001).
func TestTraceWriterCloseWithTimeoutFlushesAll(t *testing.T) {
	dir := t.TempDir()
	w := NewTraceWriter(dir, "flushall")

	const count = 25
	for range count {
		w.Write(makeEntry("FlushAll"))
	}

	if err := w.CloseWithTimeout(generousFlushBudget); err != nil {
		t.Fatalf("CloseWithTimeout returned %v, want nil", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "trace-flushall.jsonl"))
	if err != nil {
		t.Fatalf("read trace file: %v", err)
	}
	if got := len(nonEmptyLines(string(data))); got != count {
		t.Errorf("flushed %d lines, want %d", got, count)
	}
}

// TestTraceWriterCloseWithTimeoutAbandonsOnBudget asserts that an exhausted
// budget returns a distinguishable timeout signal without waiting forever
// (AC-HTF-002 — REQ-HTF-003 and REQ-HTF-004).
func TestTraceWriterCloseWithTimeoutAbandonsOnBudget(t *testing.T) {
	dir := t.TempDir()
	release := make(chan struct{})
	w := newTraceWriter(dir, "budget", func() { <-release })
	t.Cleanup(func() {
		close(release)
		<-w.done
	})

	for range 10 {
		w.Write(makeEntry("Budget"))
	}

	start := time.Now()
	err := w.CloseWithTimeout(tightFlushBudget)
	elapsed := time.Since(start)

	if !errors.Is(err, ErrFlushTimeout) {
		t.Fatalf("CloseWithTimeout returned %v, want an error matching ErrFlushTimeout", err)
	}
	// A regression to unbounded waiting (REQ-HTF-004) hangs here until the
	// package test timeout fires, so this assertion is falsifiable.
	if elapsed > 20*tightFlushBudget {
		t.Errorf("CloseWithTimeout waited %v, want bounded near %v", elapsed, tightFlushBudget)
	}
}

// TestCloseWithTimeoutReportsUndrainedCount asserts that abandoning the drain
// emits an observable structured field carrying the number of entries left
// unwritten, so residual loss is measurable rather than silent (AC-HTF-014 —
// REQ-HTF-013).
func TestCloseWithTimeoutReportsUndrainedCount(t *testing.T) {
	dir := t.TempDir()
	release := make(chan struct{})
	w := newTraceWriter(dir, "undrained", func() { <-release })
	t.Cleanup(func() {
		close(release)
		<-w.done
	})

	// The drain goroutine is gated before it consumes anything, so every
	// enqueued entry is still buffered when the budget expires.
	const pending = 12
	for range pending {
		w.Write(makeEntry("Undrained"))
	}

	var logs bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelWarn})))
	err := w.CloseWithTimeout(tightFlushBudget)
	slog.SetDefault(prev)

	if !errors.Is(err, ErrFlushTimeout) {
		t.Fatalf("CloseWithTimeout returned %v, want an error matching ErrFlushTimeout", err)
	}
	if got := undrainedFromLog(t, logs.String()); got != pending {
		t.Errorf("emitted undrained_entries=%d, want %d", got, pending)
	}
}

// TestCloseWithTimeoutIsIdempotent asserts that repeated teardown is safe and
// that a second call does not re-pay the budget.
func TestCloseWithTimeoutIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	w := NewTraceWriter(dir, "idem")
	w.Write(makeEntry("Idem"))

	if err := w.CloseWithTimeout(generousFlushBudget); err != nil {
		t.Fatalf("first CloseWithTimeout returned %v, want nil", err)
	}

	start := time.Now()
	if err := w.CloseWithTimeout(generousFlushBudget); err != nil {
		t.Fatalf("second CloseWithTimeout returned %v, want nil", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("second CloseWithTimeout waited %v, want a prompt return", elapsed)
	}

	// Mixing the two teardown entry points must stay safe.
	if err := w.Close(); err != nil {
		t.Errorf("Close after CloseWithTimeout returned %v, want nil", err)
	}
}
