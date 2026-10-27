//go:build unix

package cli

import (
	"os"
	"syscall"
)

// openInstructionHandle opens path for reading WITHOUT blocking on a FIFO
// (audit P2-3) and WITHOUT following a symlink (sync-audit F8): the
// post-write re-check must never wait on a writer-less FIFO swapped onto
// the legacy path — O_NONBLOCK makes the open itself fail-fast — and must
// never follow a link swapped into the window — O_NOFOLLOW makes opening
// a link fail (ELOOP), which the caller maps to "changed". The caller
// verifies the handle's own Stat is a regular file before reading.
func openInstructionHandle(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK|syscall.O_NOFOLLOW, 0)
}
