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
