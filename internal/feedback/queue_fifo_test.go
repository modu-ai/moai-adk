//go:build unix

// The queue read's blocking-file fixture is unix-only (FIFO semantics; the
// B1 split rule). QueueStore.Load was a plain read: a FIFO swapped in at
// the queue path parked the sender AND `moai update` past their deadlines
// (review gate finding, P2: a 20ms deadline still hadn't returned at
// 150ms). The read is bounded the same way the consent, spool, and ledger
// reads are: non-regular files are refused without opening, and the
// open+read runs under a time box.

package feedback

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestQueueLoadRefusesFifoWithoutBlocking(t *testing.T) {
	fifoPath := filepath.Join(t.TempDir(), "queue.json")
	if err := syscall.Mkfifo(fifoPath, 0o600); err != nil {
		t.Skipf("mkfifo unavailable on this platform: %v", err)
	}

	// FIFO open ordering: open the read end non-blocking first so the
	// writer's blocking open succeeds, then release our descriptor — the
	// queue read's open will succeed and its read would block forever.
	r, err := os.OpenFile(fifoPath, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		t.Fatalf("open fifo read end: %v", err)
	}
	opened := make(chan struct{})
	go func() {
		f, err := os.OpenFile(fifoPath, os.O_WRONLY, 0)
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

	store := NewQueueStore(fifoPath)
	start := time.Now()
	_, err = store.Load()
	elapsed := time.Since(start)
	if elapsed > 300*time.Millisecond {
		t.Fatalf("the queue read blocked %s on a FIFO queue file — a non-regular file must be refused without opening it", elapsed)
	}
	if err == nil {
		t.Fatal("a FIFO queue file loaded as an ordinary (empty) queue — the reader must refuse it with an error")
	}
}
