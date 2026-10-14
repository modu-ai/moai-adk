//go:build unix

package backup

import (
	"os"
	"syscall"
)

// openRunMarker opens the run marker non-blocking: a FIFO with no writer returns at once
// instead of waiting for one, and readRunMarker's regular-file check then rejects it.
func openRunMarker(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
}
