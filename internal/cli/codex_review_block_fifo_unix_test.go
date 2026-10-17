//go:build unix

package cli

// codex_review_block_fifo_unix_test.go — card t1619, codex review round 2 (P2): a FIFO at the
// block-count record must not stall the Stop-hook gate. The record read is non-blocking and
// bound to a regular file, and a record that is not one reads as unreadable (ok=false).

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestReviewGate_BlockRecordFIFODoesNotStall(t *testing.T) {
	root := m3GateFixture(t)
	m3ProduceFail(t, root, "Verdict: fail\n\nblocking finding on shared file.")
	state := m3ReceiptState(t, root)

	recordPath := m3BlockRecordPath(root, state)
	if err := os.MkdirAll(filepath.Dir(recordPath), 0o755); err != nil {
		t.Fatalf("mkdir block record dir: %v", err)
	}
	if err := syscall.Mkfifo(recordPath, 0o644); err != nil {
		t.Fatalf("mkfifo: %v", err)
	}

	done := make(chan bool, 1)
	go func() {
		_, ok := codexReviewBlockCount(root, state)
		done <- ok
	}()
	select {
	case ok := <-done:
		if ok {
			t.Fatal("a FIFO block record must read as unreadable (ok=false), never as a count")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("codexReviewBlockCount blocked on a FIFO block record: the gate's read must be non-blocking")
	}
}
