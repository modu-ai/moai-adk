//go:build !windows

package harness

import (
	"os"
	"syscall"
)

// openStampReadOnly opens the state path for the lock-free stamp read. O_NONBLOCK keeps a FIFO with no
// writer from blocking the open (and so the prune hook) until its timeout; it adds no system call, and
// a regular file reads exactly as before.
func openStampReadOnly(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
}
