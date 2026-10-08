//go:build unix

// The owner-read FIFO fixture is unix-only (FIFO semantics; the B1 split
// rule). The stale-lock verdict read was a plain os.ReadFile: a FIFO
// swapped in at the lock path parked the reread past every deadline, and
// the caller's context could not reach a read blocked in the kernel. The
// reread must refuse a non-regular file without opening it and cost one
// small capped read, never the file's size.

package atomicfile

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestSectionOwnerReadRefusesFifoWithoutBlocking(t *testing.T) {
	path := filepath.Join(t.TempDir(), "section.lock")
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Skipf("mkfifo unavailable on this platform: %v", err)
	}

	// FIFO open ordering: open the read end non-blocking first so the
	// writer's blocking open succeeds, then release our descriptor — the
	// verdict read's open would otherwise block forever.
	r, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		t.Fatalf("open fifo read end: %v", err)
	}
	opened := make(chan struct{})
	go func() {
		f, err := os.OpenFile(path, os.O_WRONLY, 0)
		if err != nil {
			close(opened)
			return
		}
		defer func() { _ = f.Close() }()
		close(opened)
		time.Sleep(10 * time.Second)
	}()
	t.Cleanup(func() {
		select {
		case <-opened:
		case <-time.After(100 * time.Millisecond):
		}
	})
	select {
	case <-opened:
	case <-time.After(2 * time.Second):
		t.Skip("writer could not open the fifo")
	}
	_ = r.Close()

	start := time.Now()
	_, rerr := sectionRereadFn(path)
	elapsed := time.Since(start)
	if elapsed > 300*time.Millisecond {
		t.Fatalf("the section owner read blocked %s on a FIFO lock path — a non-regular file must be refused without opening it", elapsed)
	}
	if rerr == nil {
		t.Fatal("a FIFO lock path read back as a valid owner — the reread must refuse it with an error")
	}
}

// TestDisposalGateRereadRefusesFifoWithoutBlocking pins the review-gate
// finding on card t1606 (P2): the disposal gate's byte re-check read the
// path with a plain os.ReadFile — a FIFO swapped in AFTER the verdict read
// parked that read forever, past the caller's context. The re-check must
// go through the same bounded read as the verdict and refuse a
// non-regular path without opening it.
func TestDisposalGateRereadRefusesFifoWithoutBlocking(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "queue.lock")
	previousBootFixture(t, path) // a dead owner: the verdict read succeeds

	// Interpose on the read seam: the FIRST (verdict) read returns the dead
	// owner's bytes, and right after it the path becomes a FIFO — exactly
	// what a hostile swap between verdict and disposal gate looks like.
	reads := 0
	prevRead := sectionRereadFn
	t.Cleanup(func() { sectionRereadFn = prevRead })
	sectionRereadFn = func(p string) ([]byte, error) {
		reads++
		raw, err := prevRead(p)
		if reads == 1 && err == nil {
			if rmErr := os.Remove(p); rmErr == nil {
				if mkErr := syscall.Mkfifo(p, 0o600); mkErr != nil {
					t.Skipf("mkfifo unavailable on this platform: %v", mkErr)
				}
				t.Cleanup(func() { _ = os.Remove(p) })
			}
		}
		return raw, err
	}

	start := time.Now()
	if BreakStaleLock(path) {
		t.Fatal("a FIFO at the disposal path read back as the verified bytes — the gate must refuse it")
	}
	if elapsed := time.Since(start); elapsed > 300*time.Millisecond {
		t.Fatalf("the disposal gate blocked %s on a FIFO swapped in after the verdict — the re-check must use the bounded read", elapsed)
	}
}
