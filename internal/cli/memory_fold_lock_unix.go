//go:build !windows

// The unix half of the fold store's cross-process lock (REQ-DISPATCH-008):
// flock(2) on a lock file inside the store. The kernel releases the lock
// when the holder's process exits — a crashed fold leaves no stale lock to
// break.
package cli

import (
	"fmt"
	"os"
	"syscall"
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
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("memory fold: lock %s: %w", path, err)
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}, nil
}
