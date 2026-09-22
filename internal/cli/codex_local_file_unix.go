//go:build unix

package cli

import (
	"os"

	"golang.org/x/sys/unix"
)

func openCodexLocalFile(path string) (*os.File, error) {
	// NONBLOCK prevents a FIFO open from hanging before fstat can reject it.
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}
	return os.NewFile(uintptr(fd), path), nil
}
