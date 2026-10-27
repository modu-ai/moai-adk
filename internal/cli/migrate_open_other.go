//go:build !unix

package cli

import "os"

// openInstructionHandle is the non-unix counterpart of the O_NONBLOCK
// open (audit P2-3): windows named-pipe opens do not block the way unix
// FIFOs do, so a plain read-only open suffices; the caller still verifies
// the handle's Stat is a regular file before reading.
func openInstructionHandle(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY, 0)
}
