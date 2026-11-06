//go:build unix

package agentemit_test

import (
	"syscall"
	"testing"
)

// restrictUmask sets a restrictive process umask for the test and restores the
// previous one on cleanup. The umask is process-global, so a test using it must
// not run in parallel with others. It reports whether the umask could be set.
func restrictUmask(t *testing.T) bool {
	t.Helper()
	old := syscall.Umask(0o077)
	t.Cleanup(func() { syscall.Umask(old) })
	return true
}
