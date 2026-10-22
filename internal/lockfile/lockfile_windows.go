//go:build windows

package lockfile

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

// SPEC-WORKTREE-HOME-ROOT-001 adds cross-process migration and registry
// writers to the original team lock API. Use one OS-backed range for both
// acquisition forms: the previous process-local limitation cannot protect
// these writers, and descriptor/process exit must release the lock.

// Use a range beyond ordinary file contents. Windows range locks are
// mandatory for reads/writes on other handles, while this API is advisory:
// existing callers must still read their data file while holding its lock.
func advisoryRange() windows.Overlapped {
	return windows.Overlapped{OffsetHigh: 1 << 30}
}

// Lock acquires a blocking exclusive lock on the advisory coordination range.
func Lock(f *os.File) error {
	overlapped := advisoryRange()
	return windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK, 0, 1, 0, &overlapped)
}

// TryLock acquires the same exclusive range without waiting for its holder.
func TryLock(f *os.File) error {
	overlapped := advisoryRange()
	return windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &overlapped)
}

// Unlock releases the range, tolerating a descriptor that was never locked.
func Unlock(f *os.File) error {
	overlapped := advisoryRange()
	err := windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, &overlapped)
	if errors.Is(err, windows.ERROR_NOT_LOCKED) {
		return nil
	}
	return err
}
