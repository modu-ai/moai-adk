//go:build !windows

package harness

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// TestPruneStateFIFODoesNotHang: a FIFO at the state path must not block the lock-free stamp
// pre-check. A plain read-only open of a FIFO with no writer blocks, which would hold a prune hook
// until its timeout. The prune returns promptly (skipping with an error is acceptable).
func TestPruneStateFIFODoesNotHang(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")

	statePath := logPath + stampSuffix
	if err := syscall.Mkfifo(statePath, 0o644); err != nil {
		t.Skipf("mkfifo unsupported: %v", err)
	}

	r := NewRetention(logPath, filepath.Join(dir, "archive"), func() time.Time { return now })
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = r.PruneStaleEntries(30)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		// Release the blocked open so the goroutine and the temp dir do not leak.
		if f, err := os.OpenFile(statePath, os.O_RDWR, 0); err == nil {
			_ = f.Close()
		}
		<-done
		t.Fatal("PruneStaleEntries blocked on a FIFO at the state path")
	}
}
