//go:build unix

// The ledger read's blocking-file fixture is unix-only (FIFO semantics; the
// B1 split rule). The ledger read was a plain os.ReadFile: a FIFO swapped
// in at the ledger path parked the drain past its deadline. The read is
// bounded the same way the consent and spool reads are.

package outbox

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestLedgerReadRefusesFifoWithoutBlocking(t *testing.T) {
	spoolFixture(t) // establishes the temporary MOAI_HOME
	ledgerPath, err := StorePath(LedgerFileName)
	if err != nil {
		t.Fatalf("store path: %v", err)
	}
	if err := syscall.Mkfifo(ledgerPath, 0o600); err != nil {
		t.Skipf("mkfifo unavailable on this platform: %v", err)
	}

	// FIFO open ordering: open the read end non-blocking first so the
	// writer's blocking open succeeds, then release our descriptor — the
	// ledger read's open will succeed and its read would block forever.
	r, err := os.OpenFile(ledgerPath, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		t.Fatalf("open fifo read end: %v", err)
	}
	opened := make(chan struct{})
	go func() {
		f, err := os.OpenFile(ledgerPath, os.O_WRONLY, 0)
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
	_, lerr := loadLedger()
	elapsed := time.Since(start)
	if elapsed > 300*time.Millisecond {
		t.Fatalf("the ledger read blocked %s on a FIFO ledger file — a non-regular file must be refused without opening it", elapsed)
	}
	if lerr == nil {
		t.Fatal("a FIFO ledger file loaded as an ordinary (empty) ledger — the reader must refuse it with an error")
	}
	_ = filepath.Join
}
