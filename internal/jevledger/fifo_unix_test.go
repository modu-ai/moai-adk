//go:build linux || darwin || freebsd || netbsd || openbsd || dragonfly

package jevledger

import (
	"errors"
	"os"
	"syscall"
	"testing"
)

// AC-JEVL-012 (ii) and (iv): a FIFO at the ledger path, with and without a
// reader. Every Unix-only symbol (syscall.Mkfifo, O_NONBLOCK on a reader) lives
// in this file, under the same build constraint as the production open in
// open_unix.go, so the fixture never runs where the production path is the
// Lstat variant. The 2 s guard in appendWithin is a hang guard for the test,
// not a timing claim about the recorder.

func makeLedgerFIFO(t *testing.T, root string) string {
	t.Helper()
	mustMkdirAll(t, ledgerDir(root))
	path := ledgerPath(root)
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Skipf("syscall.Mkfifo unavailable here: %v", err)
	}
	return path
}

func assertStillFIFO(t *testing.T, path string) {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeNamedPipe == 0 {
		t.Errorf("the object at the ledger path is no longer a FIFO: err=%v mode=%v", err, info)
	}
}

func TestLedger_FIFODropped(t *testing.T) {
	t.Run("no-reader", func(t *testing.T) {
		root := t.TempDir()
		path := makeLedgerFIFO(t, root)
		// Registered BEFORE the call: if a mutant blocks in a write open, this
		// opens the read side and releases the blocked goroutine so the test
		// binary can exit.
		t.Cleanup(func() {
			if fd, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_NONBLOCK, 0); err == nil {
				_ = syscall.Close(fd)
			}
		})

		var res Result
		stdout, stderr := captureStd(t, func() { res = appendWithin(t, New(root), baseRow()) })
		if res.Landed || res.Reason == "" {
			t.Errorf("result = %+v, want a dropped result that names a reason", res)
		}
		if stdout != "" || stderr != "" {
			t.Errorf("a drop must be silent: stdout=%q stderr=%q", stdout, stderr)
		}
		assertStillFIFO(t, path)
	})

	t.Run("with-reader", func(t *testing.T) {
		root := t.TempDir()
		path := makeLedgerFIFO(t, root)
		// The reader is opened BEFORE the call and held through it, so a zero
		// read afterwards is attributable to the recorder, not to a pipe that
		// nobody could have written.
		fd, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_NONBLOCK, 0)
		if err != nil {
			t.Fatalf("open the FIFO read side: %v", err)
		}
		t.Cleanup(func() { _ = syscall.Close(fd) })

		var res Result
		stdout, stderr := captureStd(t, func() { res = appendWithin(t, New(root), baseRow()) })
		if res.Landed || res.Reason == "" {
			t.Errorf("result = %+v, want a dropped result that names a reason", res)
		}
		if stdout != "" || stderr != "" {
			t.Errorf("a drop must be silent: stdout=%q stderr=%q", stdout, stderr)
		}
		assertStillFIFO(t, path)

		buf := make([]byte, 4096)
		n, err := syscall.Read(fd, buf)
		if err != nil && !errors.Is(err, syscall.EAGAIN) {
			t.Fatalf("read the FIFO: %v", err)
		}
		if n > 0 {
			t.Errorf("%d bytes reached the pipe: %q", n, buf[:n])
		}
	})
}
