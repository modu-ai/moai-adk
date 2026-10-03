package harness

// Scratch probe for audit finding B3 (card t1432): a FIFO at the prune state path.
// Injected through go test -overlay only.

import (
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestZZProbeB3FifoAtStatePath(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")
	if err := syscall.Mkfifo(logPath+stampSuffix, 0o644); err != nil {
		t.Skipf("mkfifo: %v", err)
	}
	done := make(chan error, 1)
	go func() {
		done <- NewRetention(logPath, filepath.Join(dir, "archive"), func() time.Time { return now }).PruneStaleEntries(30)
	}()
	select {
	case err := <-done:
		t.Logf("PROBE-B3 returned err=%v", err)
	case <-time.After(3 * time.Second):
		t.Logf("PROBE-B3 BLOCKED: PruneStaleEntries did not return within 3s on a FIFO state path")
	}
}
