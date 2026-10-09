//go:build unix

// The outbox-log read's blocking-file fixture is unix-only (FIFO semantics;
// the B1 split rule, same as the ledger FIFO test). The duplicate-
// suppression history read was a plain os.ReadFile on the outbox log: a
// FIFO swapped in at that path parked the read past every deadline, hanging
// the auto-flush AND `moai update` — the same defect class the queue read
// was hardened against. The read must refuse a non-regular file without
// opening it and return inside its bound.

package outbox

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func TestSentHistoryReadRefusesFifoOutboxLog(t *testing.T) {
	spoolFixture(t) // establishes the test's isolated MOAI_HOME
	consentOn(t)

	logPath, err := StorePath(OutboxFileName)
	if err != nil {
		t.Fatalf("store path: %v", err)
	}
	if err := syscall.Mkfifo(logPath, 0o600); err != nil {
		t.Skipf("mkfifo unavailable on this platform: %v", err)
	}

	// FIFO open ordering (same as the ledger fixture): open the read end
	// non-blocking first so the writer's blocking open succeeds, then
	// release our descriptor — the history read's open succeeds and its
	// read would block forever.
	r, err := os.OpenFile(logPath, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		t.Fatalf("open fifo read end: %v", err)
	}
	opened := make(chan struct{})
	go func() {
		f, err := os.OpenFile(logPath, os.O_WRONLY, 0)
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
	got := SentHistoryHasFingerprintWithin("aaaaaaaaaaaaaaaa", time.Now(), 7)
	elapsed := time.Since(start)
	if elapsed > 300*time.Millisecond {
		t.Fatalf("the sent-history read blocked %s on a FIFO outbox log — a non-regular file must be refused without opening it", elapsed)
	}
	if got {
		t.Fatal("a FIFO outbox log reported a sent row — the reader must treat the refusal as no history")
	}
}

// The append path had the same defect the read path had: a plain blocking
// open on a FIFO parked the flush — and `moai update` with it — past every
// deadline (review gate finding, P2). The append must refuse a non-regular
// log without opening it and return inside its time box.
func TestAppendOutboxRefusesFifoOutboxLog(t *testing.T) {
	spoolFixture(t)
	consentOn(t)

	logPath, err := StorePath(OutboxFileName)
	if err != nil {
		t.Fatalf("store path: %v", err)
	}
	if err := syscall.Mkfifo(logPath, 0o600); err != nil {
		t.Skipf("mkfifo unavailable on this platform: %v", err)
	}

	r, err := os.OpenFile(logPath, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		t.Fatalf("open fifo read end: %v", err)
	}
	opened := make(chan struct{})
	go func() {
		f, err := os.OpenFile(logPath, os.O_WRONLY, 0)
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
	aerr := AppendOutbox(OutboxRow{Outcome: "sent", Fingerpr: "aaaaaaaaaaaaaaaa", At: time.Now().UTC().Format(time.RFC3339)})
	elapsed := time.Since(start)
	if elapsed > 300*time.Millisecond {
		t.Fatalf("the append blocked %s on a FIFO outbox log — a non-regular file must be refused without opening it", elapsed)
	}
	if aerr == nil {
		t.Fatal("a FIFO outbox log accepted an append — the writer must refuse it with an error")
	}
}
