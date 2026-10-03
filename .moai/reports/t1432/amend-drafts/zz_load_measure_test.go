//go:build !windows

package harness

// Measurement instrument for Definition of Done 15 of SPEC-HARNESS-RETENTION-HARDEN-001 (card t1432):
// the wall time of PruneStaleEntries(30) when the heal path runs and another descriptor holds the heal
// lock for about 1.8 s, i.e. the sum of the heal-lock wait, the heal and the prune. Not part of the
// change and not in the tree: it is injected with `go test -overlay`.
//
// Fixture per observation: a synthetic usage log of about 65.8 MB, about 12.5 percent of its events
// older than the 30-day cut; a symbolic link owned by the user at the state path (so the heal runs);
// a test-owned descriptor holding the heal lock exclusively and releasing it 1.8 s after the call starts.
// Three observations with no load (control), then three with CPU-busy goroutines inside this process,
// one per available CPU, bounded by a 200 s context deadline and stopped by t.Cleanup. Run the whole
// go test under `timeout 330` with `-timeout 300s` (raised from 120/150/180 after the first run's
// observations reached 23 s each). Prune-only observations (no heal, no holder) decompose the sum.
// Every observation logs the returned error; an
// observation with a non-nil error is invalid (the call failed closed and the prune did not run).

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

func buildLoadMeasureLog(t *testing.T, now time.Time) []byte {
	t.Helper()
	const target = 65_800_000
	var sb strings.Builder
	sb.Grow(target + 1024)
	pad := strings.Repeat("x", 80)
	for i := 0; sb.Len() < target; i++ {
		ts := now.AddDate(0, 0, -1)
		if i%8 == 0 { // 12.5 percent older than the 30-day cut
			ts = now.AddDate(0, 0, -40)
		}
		evt := Event{Timestamp: ts, EventType: EventTypeFeedback, Subject: fmt.Sprintf("subject-%d-%s", i, pad), ContextHash: "h", SchemaVersion: LogSchemaVersion}
		b, err := json.Marshal(evt)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		sb.Write(b)
		sb.WriteByte('\n')
	}
	return []byte(sb.String())
}

func TestZZLoadMeasureHealWaitPlusPrune(t *testing.T) {
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	logBytes := buildLoadMeasureLog(t, now)
	t.Logf("MEASURE fixture: log_bytes=%d cpus=%d", len(logBytes), runtime.NumCPU())

	// healAndWait false: a healthy state file with an expired stamp and no heal-lock holder, so the wall
	// time is the prune alone (decomposition control, added after the first run exceeded 5 s).
	observe := func(label string, n int, healAndWait bool) {
		dir := t.TempDir()
		logPath := filepath.Join(dir, "usage-log.jsonl")
		f, err := os.Create(logPath)
		if err != nil {
			t.Fatalf("create log: %v", err)
		}
		w := bufio.NewWriterSize(f, 1<<20)
		if _, err := w.Write(logBytes); err != nil {
			t.Fatalf("write log: %v", err)
		}
		if err := w.Flush(); err != nil {
			t.Fatalf("flush log: %v", err)
		}
		if err := f.Close(); err != nil {
			t.Fatalf("close log: %v", err)
		}
		victim := filepath.Join(dir, "victim")
		if err := os.WriteFile(victim, []byte("V"), 0o644); err != nil {
			t.Fatalf("write victim: %v", err)
		}
		var hf *os.File
		if healAndWait {
			if err := os.Symlink(victim, logPath+stampSuffix); err != nil {
				t.Fatalf("symlink state: %v", err)
			}
			if hf, err = os.OpenFile(logPath+".prune-heal", os.O_RDWR|os.O_CREATE, 0o600); err != nil {
				t.Fatalf("open heal lock: %v", err)
			}
			defer func() { _ = hf.Close() }()
			if err := syscall.Flock(int(hf.Fd()), syscall.LOCK_EX); err != nil {
				t.Fatalf("flock heal lock: %v", err)
			}
		} else if err := os.WriteFile(logPath+stampSuffix, []byte("2026-09-01T00:00:00Z"), 0o644); err != nil {
			t.Fatalf("write state: %v", err)
		}
		ret := NewRetention(logPath, filepath.Join(dir, "archive"), func() time.Time { return now })
		start := time.Now()
		if hf != nil {
			timer := time.AfterFunc(1800*time.Millisecond, func() { _ = syscall.Flock(int(hf.Fd()), syscall.LOCK_UN) })
			defer timer.Stop()
		}
		perr := ret.PruneStaleEntries(30)
		wall := time.Since(start)
		after, _ := os.Stat(logPath)
		size := int64(-1)
		if after != nil {
			size = after.Size()
		}
		valid := perr == nil && size >= 0 && size < int64(len(logBytes))
		t.Logf("MEASURE %s #%d: wall=%v err=%v log_bytes_after=%d valid=%v over_5s=%v", label, n, wall, perr, size, valid, wall > 5*time.Second)
	}

	for i := 1; i <= 3; i++ {
		observe("no-load prune-only", i, false)
	}
	for i := 1; i <= 3; i++ {
		observe("no-load", i, true)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Second)
	var wg sync.WaitGroup
	t.Cleanup(func() {
		cancel()
		wg.Wait()
	})
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			x := 0
			for ctx.Err() == nil {
				for j := 0; j < 100000; j++ {
					x += j
				}
			}
			_ = x
		}()
	}
	for i := 1; i <= 3; i++ {
		observe("load", i, true)
	}
	for i := 1; i <= 3; i++ {
		observe("load prune-only", i, false)
	}
}
