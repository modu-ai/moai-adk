//go:build unix

// The spool read's blocking-file fixture is unix-only (FIFO semantics; the
// B1 split rule — a build tag, not a runtime skip). The spool READ was a
// plain os.ReadFile: a FIFO swapped in at the spool path parked the drain
// — open, then read — ignoring its deadline entirely (review gate finding,
// P2: a 20ms-deadline drain still hadn't returned at 250ms). The read is
// bounded the same way the consent reader is: non-regular files are refused
// without opening, and the open+read runs under a time box.

package bugreport

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// TestSpoolReadRejectsFifoWithoutBlocking: the drain's spool read must
// refuse a non-regular spool file WITHOUT opening it.
func TestSpoolReadRejectsFifoWithoutBlocking(t *testing.T) {
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	t.Setenv("CI", "")

	spoolDir := filepath.Join(home, filepath.FromSlash(BugreportStoreDir))
	if err := os.MkdirAll(spoolDir, 0o700); err != nil {
		t.Fatalf("mkdir store: %v", err)
	}
	fifoPath := filepath.Join(spoolDir, "spool.jsonl")
	if err := syscall.Mkfifo(fifoPath, 0o600); err != nil {
		t.Skipf("mkfifo unavailable on this platform: %v", err)
	}
	holdFifoWriterOpen(t, fifoPath)

	start := time.Now()
	_, _, err := ReadSpoolConsumable()
	elapsed := time.Since(start)
	if elapsed > 300*time.Millisecond {
		t.Fatalf("the spool read blocked %s on a FIFO spool file — a non-regular file must be refused without opening it", elapsed)
	}
	if err == nil {
		t.Fatal("a FIFO spool file read as an ordinary (empty) spool — the reader must refuse it with an error, not silently drain nothing")
	}
}

// holdFifoWriterOpen holds a writer on the FIFO open without writing — the
// reader's open succeeds but its read blocks forever. FIFO open ordering: a
// blocking O_WRONLY open parks until a reader exists, so the fixture opens
// the read end NON-BLOCKING first, lets the writer's open succeed, then
// closes its own descriptor. Cleanup-guaranteed: the writer is closed when
// the test ends so a leaked reader goroutine (on the pre-fix tree) can
// exit.
func holdFifoWriterOpen(t *testing.T, fifoPath string) {
	t.Helper()
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
}
