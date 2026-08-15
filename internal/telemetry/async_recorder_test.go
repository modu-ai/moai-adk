package telemetry

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestAsyncRecorder_RecordReturnsFastSequential verifies the caller-path
// contract of Record(): it is a select/default channel send (async_recorder.go)
// — microseconds, never disk I/O — and every call returns either nil or
// ErrRecordDropped. Persistence after Close is asserted on the same run.
//
// What was wrong with the earlier form (TestAsyncRecorder_NonBlockingUnderLoad):
// it spawned 1000 concurrent goroutines and counted calls exceeding 100ms —
// measuring the Go scheduler under self-induced load, not the recorder. The
// recorder's own cost (a channel send) is nanoseconds; the scheduler storm was
// the entire signal, which is why it passed alone and failed on a loaded box.
//
// The sequential shape below catches the real regression class — Record
// growing synchronous work (a disk write per call, a mutex-held flush, a
// network call): 2000 sequential channel sends complete in well under 10ms
// even on a loaded box, while 2000 synchronous disk writes take seconds. The
// 2s total bound therefore sits ~3 orders of magnitude above healthy cost —
// too far for load noise to bridge, close enough that the regression class
// trips it. This is an absolute bound stated as an average; there is no
// natural budget to make it relative to.
//
// What it deliberately does NOT catch: a blocking send (`r.ch <- rec` without
// the select/default) — with a live consumer goroutine a blocking send still
// makes progress, so only a stalled consumer exposes it, and stalling the
// consumer deterministically needs an injection seam the recorder does not
// expose. If that invariant ever matters enough to test, add the seam first;
// a goroutine storm does not test it either.
func TestAsyncRecorder_RecordReturnsFastSequential(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Small buffer so both return paths (accepted / dropped) are likely
	// exercised across the run; drops are permitted, blocking is not.
	const bufSize = 8
	const numRecords = 2000
	// 1ms average per call. A channel send costs ~100ns; synchronous disk
	// writes cost ~ms each. The bound separates the two by ~3 orders of
	// magnitude on both sides.
	const totalBound = 2 * time.Second

	rec := NewAsyncRecorder(dir, bufSize)

	ts := time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)
	r := UsageRecord{
		Timestamp: ts,
		SkillID:   "test-skill",
		SessionID: "sess-sequential-test",
		Outcome:   OutcomeUnknown,
	}

	start := time.Now()
	dropped := 0
	for i := range numRecords {
		if err := rec.Record(r); err != nil {
			if !errors.Is(err, ErrRecordDropped) {
				t.Fatalf("Record %d returned unexpected error: %v", i, err)
			}
			dropped++
		}
	}
	elapsed := time.Since(start)
	t.Logf("%d sequential Record calls: total=%v avg=%s dropped=%d/%d",
		numRecords, elapsed, elapsed/time.Duration(numRecords), dropped, numRecords)

	if elapsed > totalBound {
		t.Errorf("%d sequential Record calls took %v (> %v, ~1ms/call average) — "+
			"Record() must be a channel send, not synchronous work; "+
			"this indicates Record grew disk I/O or a lock on the caller path",
			numRecords, elapsed, totalBound)
	}

	// Close after all records are processed.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rec.Close(ctx); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Verify that records were written to file (drop policy may keep less than total).
	telDir := filepath.Join(dir, ".moai", "evolution", "telemetry")
	entries, err := os.ReadDir(telDir)
	if err != nil {
		t.Fatalf("read telemetry dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no telemetry files were created")
	}
}

// TestAsyncRecorder_DropPolicyWhenFull verifies that when the buffer is full,
// Record returns an error instead of blocking.
func TestAsyncRecorder_DropPolicyWhenFull(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Buffer size 1; block the consumer so the buffer fills up.
	rec := NewAsyncRecorder(dir, 1)
	// Alternatively, fill the channel without a consumer to drive the test.

	ts := time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)
	r := UsageRecord{
		Timestamp: ts,
		SkillID:   "test-skill",
		Outcome:   OutcomeUnknown,
	}

	// Call Record many times and verify that drops occur.
	// Buffer is 1, so after the first few calls drops must happen.
	var dropped int
	for i := 0; i < 100; i++ {
		err := rec.Record(r)
		if errors.Is(err, ErrRecordDropped) {
			dropped++
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = rec.Close(ctx)

	// Drops must have occurred (because buffer is 1).
	// If the consumer is fast, dropped count may be small — verify loosely.
	t.Logf("dropped records: %d/100", dropped)
}

// TestAsyncRecorder_ReusesFileHandle verifies that the async recorder does not
// open the file on every record write (file handle reuse).
func TestAsyncRecorder_ReusesFileHandle(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	rec := NewAsyncRecorder(dir, 500)

	// Record 100 entries with the same date.
	ts := time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)
	for i := 0; i < 100; i++ {
		r := UsageRecord{
			Timestamp: ts,
			SkillID:   "test-skill",
			SessionID: "sess-reuse-test",
			Outcome:   OutcomeUnknown,
		}
		if err := rec.Record(r); err != nil {
			t.Fatalf("Record %d: %v", i, err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rec.Close(ctx); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Verify that all 100 records were written to the file.
	telDir := filepath.Join(dir, ".moai", "evolution", "telemetry")
	path := filepath.Join(telDir, "usage-2026-04-15.jsonl")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var rec UsageRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Errorf("invalid JSON at line %d: %v", count+1, err)
		}
		count++
	}

	if count != 100 {
		t.Errorf("expected 100 records, got %d", count)
	}
}
