//go:build unix

package update

import (
	"syscall"
	"testing"
)

// makeFifo plants a POSIX FIFO at path — the fixture the classifier's
// non-regular-target guard is tested against (gate round 10, finding 5).
// The syscall lives in a unix-tagged file (the established
// audit_plan_cmd_fifo_unix_test.go pattern): syscall.Mkfifo does not exist
// on Windows, and the calling test skips there at runtime — the symbol must
// never reach the windows/amd64 compile.
func makeFifo(t *testing.T, path string) {
	t.Helper()
	if err := syscall.Mkfifo(path, 0o644); err != nil {
		t.Fatalf("mkfifo: %v", err)
	}
}
