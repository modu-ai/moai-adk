//go:build !unix

package update

import "testing"

// makeFifo is the non-unix counterpart of reconcile_fifo_unix_test.go: a
// platform without Mkfifo cannot plant the FIFO fixture, so the helper
// skips — the calling test (TestClassifyNonRegularTargetDoesNotHang) skips
// on windows at runtime anyway, and the symbol stays resolvable so the
// windows/amd64 compile never sees an undefined reference.
func makeFifo(t *testing.T, _ string) {
	t.Skip("FIFOs are a POSIX fixture")
}
