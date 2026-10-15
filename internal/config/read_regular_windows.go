//go:build windows

package config

import (
	"fmt"
	"io"
	"os"
)

// readNonBlockingRegularFile is the Windows arm of the regular-file guard
// (card t1729 audit round-5): named pipes exist on Windows too, so the
// opened handle is judged for regularity before any read. The O_NONBLOCK
// open trick is POSIX-only and unnecessary here — os.Open does not block
// on a Windows named pipe the way a POSIX open blocks on a FIFO.
func readNonBlockingRegularFile(path string) ([]byte, error) {
	f, err := os.Open(path)
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
	return io.ReadAll(io.LimitReader(f, 1<<20))
}
