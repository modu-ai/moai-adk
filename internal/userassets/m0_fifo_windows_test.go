//go:build windows

package userassets

// m0_fifo_windows_test.go — the windows half of the FIFO fixture helper
// (SPEC-USERASSET-DEPLOY-GUARD-001 M0.1, gate round 13 P1). Named-pipe
// semantics are unix here; the windows axis of the collision family is the
// GOOS=windows BUILD gate, and this stub keeps the whole test package
// compilable there.

import "testing"

// mkfifoOrSkip skips: the FIFO precheck repro is a unix-execution
// observation.
func mkfifoOrSkip(t *testing.T, path string) {
	t.Helper()
	t.Skip("FIFO semantics are unix — the windows axis of this family is the GOOS=windows build gate")
}

// releaseFifoWriter is the windows stub: nothing parked on a FIFO.
func releaseFifoWriter(t *testing.T, path string) {
	t.Helper()
}
