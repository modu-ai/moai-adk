//go:build windows

package decision

import (
	"os"

	"golang.org/x/sys/windows"
)

// lockBoard takes an exclusive lock on the board's companion lock file and
// returns its release.
func lockBoard(path string) (func(), error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	var ov windows.Overlapped
	if err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK, 0, 1, 0, &ov); err != nil {
		_ = f.Close()
		return nil, err
	}
	return func() {
		_ = windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, &ov)
		_ = f.Close()
	}, nil
}
