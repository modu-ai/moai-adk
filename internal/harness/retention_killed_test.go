//go:build !windows

package harness

// Killed-pruner contract (card t1425, sync-audit F1): the attempt stamp must
// be on disk BEFORE the prune does its work, so that a pruner killed by the
// hook timeout after archiving and before the rename is not repeated by every
// later hook. The test needs no production seam: the archive path is a FIFO,
// so the pruner blocks inside its archive step until the test reads from it.

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestPruneStamp_StampExistsBeforeTheWork(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	archiveDir := filepath.Join(dir, "archive")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")

	// The stale event lands in this month's archive; make that archive path a FIFO.
	month := now.AddDate(0, 0, -40).UTC().Format("2006-01")
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		t.Fatalf("mkdir archive: %v", err)
	}
	fifo := filepath.Join(archiveDir, month+".jsonl.gz")
	if err := syscall.Mkfifo(fifo, 0o644); err != nil {
		t.Skipf("mkfifo unavailable: %v", err)
	}

	// drain unblocks the pruner: opening the FIFO for reading rendezvouses with its blocked open. It
	// never blocks on a pruner that did not reach its archive step: the blocking read open runs in a
	// goroutine, and once archiveStepWait has passed the goroutine is released by opening the write
	// side without blocking (that open succeeds only while a reader is blocked in its open). drain
	// reports whether a pruner opened the FIFO for writing, that is, reached its archive step.
	const archiveStepWait = 10 * time.Second
	var drainOnce sync.Once
	archiveReached := false
	drain := func() bool {
		drainOnce.Do(func() {
			opened := make(chan struct{})
			go func() {
				defer close(opened)
				f, err := os.OpenFile(fifo, os.O_RDONLY, 0)
				if err != nil {
					return
				}
				_, _ = io.Copy(io.Discard, f)
				_ = f.Close()
			}()
			select {
			case <-opened:
				archiveReached = true
				return
			case <-time.After(archiveStepWait):
			}
			release := time.NewTicker(10 * time.Millisecond)
			defer release.Stop()
			for {
				if w, err := os.OpenFile(fifo, os.O_WRONLY|syscall.O_NONBLOCK, 0); err == nil {
					_ = w.Close()
				}
				select {
				case <-opened:
					return
				case <-release.C:
				}
			}
		})
		return archiveReached
	}
	defer drain()

	pruner := NewRetention(logPath, archiveDir, func() time.Time { return now })
	done := make(chan error, 1)
	go func() { done <- pruner.PruneStaleEntries(30) }()

	// While the pruner is stuck in its archive step the stamp must already be on disk.
	stamped := false
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if stampIsFresh(readStampFile(logPath+stampSuffix), now) {
			stamped = true
			break
		}
		time.Sleep(time.Millisecond)
	}
	if !stamped {
		drain()
		select {
		case <-done:
		case <-time.After(time.Second):
		}
		t.Fatalf("no stamp on disk while the pruner was still working: a killed pruner would be repeated")
	}

	// A hook arriving now skips at once instead of piling onto the stuck pruner.
	other := NewRetention(logPath, archiveDir, func() time.Time { return now.Add(time.Minute) })
	if err := waitOrFail(t, 2*time.Second, "second hook during a stuck prune", func() error {
		return other.PruneStaleEntries(30)
	}); err != nil {
		t.Fatalf("second hook: %v", err)
	}
	if !logHasSubject(t, logPath, "stale-1") {
		t.Fatalf("the log changed before the pruner finished its archive step")
	}

	// Let the pruner finish and confirm it completes normally. A pruner that never reached its
	// archive step fails here, within archiveStepWait, instead of waiting for the go test timeout.
	if !drain() {
		t.Fatalf("the pruner never reached its archive step: nothing opened the archive FIFO %s for writing within %s", fifo, archiveStepWait)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("pruner failed after being unblocked: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("pruner did not finish after the archive step was unblocked")
	}
	if logHasSubject(t, logPath, "stale-1") {
		t.Fatalf("pruner finished but the stale event is still in the log")
	}
}
