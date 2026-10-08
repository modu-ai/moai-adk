//go:build windows

// The windows half of the fold store's cross-process lock
// (REQ-DISPATCH-008): LockFileEx on a byte range of a lock file inside the
// store, exclusive and blocking. The OS releases the lock when the holder's
// process exits — a crashed fold leaves no stale lock to break.
package cli

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

// acquireFoldStoreLock takes the store's EXCLUSIVE lock, blocking until
// every other fold process on this store has released it. The returned
// release func unlocks and closes; it is called exactly once, by the
// caller's defer.
func acquireFoldStoreLock(dir string) (func(), error) {
	path := foldLockPath(dir)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o666)
	if err != nil {
		return nil, fmt.Errorf("memory fold: open the store lock %s: %w", path, err)
	}
	ol := new(windows.Overlapped)
	if err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK, 0, 1, 0, ol); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("memory fold: lock %s: %w", path, err)
	}
	return func() {
		_ = windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, ol)
		_ = f.Close()
	}, nil
}
