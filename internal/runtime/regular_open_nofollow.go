//go:build linux || darwin || freebsd || netbsd || openbsd || dragonfly

package runtime

import (
	"errors"
	"io/fs"
	"os"
	"syscall"
)

// openRegularFile opens path with O_NOFOLLOW (a symlink at the path fails the
// open instead of being followed) and O_NONBLOCK (a FIFO opens at once instead
// of waiting for a writer), then requires the OPENED descriptor to be a
// regular file. No later step resolves the path again, so a swap between a
// check and the open has nothing left to race.
func openRegularFile(path string) (*os.File, error) {
	beforeRegularOpen(path)
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK|syscall.O_NOFOLLOW, 0)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			// ELOOP/EMLINK/EFTYPE differ by platform: name a link as a link.
			if li, lerr := os.Lstat(path); lerr == nil && li.Mode()&fs.ModeSymlink != 0 {
				return nil, notRegularError(path, li.Mode())
			}
		}
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if !info.Mode().IsRegular() {
		_ = f.Close()
		return nil, notRegularError(path, info.Mode())
	}
	return f, nil
}
