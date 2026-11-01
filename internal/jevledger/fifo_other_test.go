//go:build !(linux || darwin || freebsd || netbsd || openbsd || dragonfly)

package jevledger

import "testing"

// The counterpart of fifo_unix_test.go for every GOOS the Unix constraint
// excludes (windows included). syscall.Mkfifo and syscall.O_NOFOLLOW do not
// exist there, so the FIFO cases cannot be compiled, and a run-time skip
// inside an untagged file would not help: the package would fail to compile
// before any test ran. This named skip is the continued-firing signal — the
// FIFO cases show as absent on those platforms instead of silently missing.
func TestLedger_FIFOCasesUnsupportedHere(t *testing.T) {
	t.Skip("named pipes unsupported on this platform; FIFO cases run on the unix set only")
}
