//go:build windows

// The windows half of the fold store's cross-process lock
// (REQ-DISPATCH-008): LockFileEx on a byte range of a lock file beside the
// store's index, exclusive. The OS releases the lock when the holder's
// process exits — a crashed fold leaves no stale lock to break. A non-nil
// forbidden makes the wait abandonment-aware: the acquisition polls a
// fail-immediately try instead of blocking forever, so a bounded fold whose
// caller timed out exits the wait promptly (the abandonment trio's third
// member) — C5's cross-process symmetry cuts both ways.
package cli

import (
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/windows"
)

// acquireFoldStoreLock takes the store's EXCLUSIVE lock. A nil forbidden
// blocks until every other fold process on this store has released it; a
// non-nil forbidden polls a fail-immediately try and abandons — releasing
// the descriptor, holding nothing — when the caller's deadline has passed.
// The returned release func unlocks and closes; it is called exactly once,
// by the caller's defer.
func acquireFoldStoreLock(dir string, forbidden func() bool) (func(), error) {
	path := foldLockPath(dir)
	// Windows: os.OpenFile passes no OPEN_REPARSE_POINT flag, so an open
	// WOULD follow a symlink at this path; the Lstat guard refuses one
	// first (a narrow TOCTOU tail remains between the lstat and the open —
	// named in the SPEC's run record; the unix half is O_NOFOLLOW-exact).
	if fi, statErr := os.Lstat(path); statErr == nil && fi.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("memory fold: the store lock %s is a symlink — refusing", path)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o666)
	if err != nil {
		return nil, fmt.Errorf("memory fold: open the store lock %s: %w", path, err)
	}
	ol := new(windows.Overlapped)
	if forbidden == nil {
		if err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK, 0, 1, 0, ol); err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("memory fold: lock %s: %w", path, err)
		}
		return releaseFoldStoreLockFunc(f, ol), nil
	}
	for {
		err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, ol)
		if err == nil {
			return releaseFoldStoreLockFunc(f, ol), nil
		}
		if !errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
			_ = f.Close()
			return nil, fmt.Errorf("memory fold: lock %s: %w", path, err)
		}
		if forbidden() {
			_ = f.Close()
			return nil, fmt.Errorf("memory fold: the step was abandoned while waiting for the store lock %s", path)
		}
		time.Sleep(foldAbandonedReadPoll)
	}
}

func releaseFoldStoreLockFunc(f *os.File, ol *windows.Overlapped) func() {
	return func() {
		_ = windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, ol)
		_ = f.Close()
	}
}
