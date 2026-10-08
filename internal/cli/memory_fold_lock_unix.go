//go:build !windows

// The unix half of the fold store's cross-process lock (REQ-DISPATCH-008):
// flock(2) on a lock file beside the store's index. The kernel releases the
// lock when the holder's process exits — a crashed fold leaves no stale
// lock to break. A non-nil forbidden makes the wait abandonment-aware: the
// acquisition polls a non-blocking try instead of blocking forever, so a
// bounded fold whose caller timed out exits the wait promptly (the
// abandonment trio's third member) — C5's cross-process symmetry cuts both
// ways.
package cli

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
)

// acquireFoldStoreLock takes the store's EXCLUSIVE lock. A nil forbidden
// blocks until every other fold process on this store has released it; a
// non-nil forbidden polls a non-blocking try and abandons — releasing the
// descriptor, holding nothing — when the caller's deadline has passed. The
// returned release func unlocks and closes; it is called exactly once, by
// the caller's defer.
//
// The open carries O_NOFOLLOW and the opened file is fstat-verified to be
// a REGULAR file: a symlinked `.moai-store-lock` is refused, never
// followed — the lock must never be taken on a file outside the store
// (post-close gate finding 4; the repro locked another repository's
// .git/index.lock into existence). Every other non-regular shape (a FIFO
// standing in for the lock, a device node) is refused the same way.
func acquireFoldStoreLock(dir string, forbidden func() bool) (func(), error) {
	path := foldLockPath(dir)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0o666)
	if err != nil {
		return nil, fmt.Errorf("memory fold: open the store lock %s: %w", path, err)
	}
	if st, statErr := f.Stat(); statErr != nil {
		_ = f.Close()
		return nil, fmt.Errorf("memory fold: stat the store lock %s: %w", path, statErr)
	} else if !st.Mode().IsRegular() {
		_ = f.Close()
		return nil, fmt.Errorf("memory fold: the store lock %s is not a regular file — refusing", path)
	}
	if forbidden == nil {
		if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("memory fold: lock %s: %w", path, err)
		}
		return releaseFoldStoreLockFunc(f), nil
	}
	for {
		err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return releaseFoldStoreLockFunc(f), nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) {
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

func releaseFoldStoreLockFunc(f *os.File) func() {
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}
}
