package harness

import (
	"path/filepath"
	"syscall"
	"testing"
)

// Control input for the Windows vet check: a FIFO test with no build constraint.
func TestZZUntaggedFifo(t *testing.T) {
	if err := syscall.Mkfifo(filepath.Join(t.TempDir(), "f"), 0o644); err != nil {
		t.Skipf("mkfifo unavailable: %v", err)
	}
}
