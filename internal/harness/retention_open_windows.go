//go:build windows

package harness

import "os"

// openStampReadOnly opens the state path for the lock-free stamp read. Windows has no FIFO to block on.
func openStampReadOnly(path string) (*os.File, error) {
	return os.Open(path)
}
