//go:build unix

package factory

import (
	"fmt"
	"os"
	"syscall"
)

// openCandidateRecordFile opens a candidate record with the non-blocking,
// no-follow flags and verifies the OPENED descriptor is a regular file
// before any read (card t1478 M2 repair — the TOCTOU tail of the Lstat
// precheck): a record swapped from a regular file to a FIFO between the
// precheck and the open would otherwise park os.ReadFile inside the
// candidate mutation lock, wedging every later candidate call of the card.
// O_NONBLOCK makes even a FIFO open return immediately; O_NOFOLLOW refuses
// a symlink swapped in the same window. The descriptor — not the path —
// is what the verdict inspects, so what was verified is what is read.
func openCandidateRecordFile(path string) (*os.File, error) {
	fd, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_NONBLOCK|syscall.O_NOFOLLOW, 0)
	if err != nil {
		if err == syscall.ENOENT {
			return nil, os.ErrNotExist
		}
		if err == syscall.ELOOP {
			return nil, fmt.Errorf("candidate record at %s is a symbolic link — refusing to open it", path)
		}
		return nil, fmt.Errorf("candidate store: open %s: %w", path, err)
	}
	var st syscall.Stat_t
	if err := syscall.Fstat(fd, &st); err != nil {
		_ = syscall.Close(fd)
		return nil, fmt.Errorf("candidate store: fstat %s: %w", path, err)
	}
	if st.Mode&syscall.S_IFMT != syscall.S_IFREG {
		_ = syscall.Close(fd)
		return nil, fmt.Errorf("candidate record at %s is not a regular file — refusing to open it", path)
	}
	return os.NewFile(uintptr(fd), path), nil
}
