//go:build !windows

package config

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/sys/unix"
)

// readNonBlockingRegularFile reads the file at path with the POSIX TOCTOU
// guard (card t1729 audit round-5 — the same shape as the hook package's
// rule reads): the path is opened NON-BLOCKING first — a plain open BLOCKS
// on a FIFO until a writer appears — the OPENED HANDLE is fstat-ed for
// regularity, and only a regular file is read (O_NONBLOCK cleared so the
// read itself blocks on data, not on the open). Reads are bounded: a
// one-key config never needs more.
func readNonBlockingRegularFile(path string) ([]byte, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("opened-handle stat failed: %w", err)
	}
	if !fi.Mode().IsRegular() {
		return nil, fmt.Errorf("config read: %s is not a regular file", path)
	}
	if err := unix.SetNonblock(int(f.Fd()), false); err != nil {
		return nil, fmt.Errorf("clear non-blocking mode: %w", err)
	}
	return io.ReadAll(io.LimitReader(f, 1<<20))
}
