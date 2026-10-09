//go:build unix

// The model-calls budget read's blocking-file fixture is unix-only (FIFO
// semantics; the B1 split rule). The budget read runs INSIDE the queue-lock
// mutation, so a FIFO swapped in at the budget path didn't just stall the
// sender — it held the queue LOCK while parked, stalling every other queue
// operation too (review gate finding, P2). The read is bounded the same way
// the consent, spool, ledger, and queue reads are.

package outbox

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestModelCallsReadRefusesFifoWithoutBlocking(t *testing.T) {
	spoolFixture(t) // establishes the test's isolated MOAI_HOME
	path, err := StorePath(ModelCallsFileName)
	if err != nil {
		t.Fatalf("store path: %v", err)
	}
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Skipf("mkfifo unavailable on this platform: %v", err)
	}

	// FIFO open ordering: open the read end non-blocking first so the
	// writer's blocking open succeeds, then release our descriptor.
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
	allowed := AllowAndRecordModelCall(context.Background(), time.Now())
	elapsed := time.Since(start)
	if elapsed > 300*time.Millisecond {
		t.Fatalf("the model-calls judgment blocked %s on a FIFO budget file — while HOLDING the queue lock, stalling every queue operation", elapsed)
	}
	// The budget is a spend bound, not a safety gate: an unreadable budget
	// reads as empty spend (the established semantics), so the call is
	// allowed but the stall is bounded.
	if !allowed {
		t.Fatal("the FIFO budget file refused the call outright — an unreadable spend log reads as empty spend, not a refusal")
	}
}
