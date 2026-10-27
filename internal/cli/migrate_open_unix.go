//go:build unix

package cli

import (
	"os"
	"syscall"
)

// openInstructionHandle opens path for reading WITHOUT blocking on a FIFO
// (audit P2-3): the post-write re-check must never wait on a writer-less
// FIFO swapped onto the legacy path — O_NONBLOCK makes the open itself
// fail-fast, and the caller verifies the handle's own Stat is a regular
// file before reading.
func openInstructionHandle(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
}
