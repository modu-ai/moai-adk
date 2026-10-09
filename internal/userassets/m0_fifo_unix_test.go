//go:build unix

package userassets

// m0_fifo_unix_test.go — the unix half of the FIFO fixture helper
// (SPEC-USERASSET-DEPLOY-GUARD-001 M0.1, gate round 13 P1). The platform
// split is a COMPILE-TIME contract: syscall.Mkfifo does not exist on
// windows, and a runtime t.Skip cannot fix a package that fails to compile
// there (GOOS=windows go test -c must build the whole test package).

import (
	"os"
	"syscall"
	"testing"
)

// mkfifoOrSkip creates a FIFO at path, failing the test on unix errors.
func mkfifoOrSkip(t *testing.T, path string) {
	t.Helper()
	if err := syscall.Mkfifo(path, 0o644); err != nil {
		t.Fatalf("mkfifo %s: %v", path, err)
	}
}

// releaseFifoWriter unblocks a reader parked on the FIFO: opening the write
// end and closing it delivers EOF to the blocked read, so the install
// goroutine can finish before the test's TempDir cleanup runs (gate round
// 13: a leaked blocked reader survives past cleanup). The reader end is
// already open (that is where it parks), so the write-end open returns.
func releaseFifoWriter(t *testing.T, path string) {
	t.Helper()
	w, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		t.Logf("fifo write-end open failed (reader may have already exited): %v", err)
		return
	}
	_ = w.Close()
}
