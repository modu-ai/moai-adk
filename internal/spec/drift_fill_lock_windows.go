//go:build windows

package spec

import (
	"fmt"
	"sync"

	"golang.org/x/sys/windows"
)

// driftFillLock is the Windows LockFileEx-based advisory lock guarding the
// out-of-band drift-cache fill. The pattern is copied from internal/sessionmsg
// lock_windows.go (itself copied from internal/session
// registry_lock_windows.go); neither of those packages is modified.
//
// Uses LockFileEx with LOCKFILE_EXCLUSIVE_LOCK | LOCKFILE_FAIL_IMMEDIATELY for
// non-blocking cross-process exclusion (parity with POSIX flock
// LOCK_EX | LOCK_NB). The lock is tied to the file HANDLE, so the kernel
// releases it when the process exits — including abnormal termination, which is
// REQ-DCF-008's platform-neutral requirement.
type driftFillLock struct {
	mu     sync.Mutex
	handle windows.Handle
}

func newDriftFillLock() *driftFillLock {
	return &driftFillLock{handle: windows.InvalidHandle}
}

// acquire opens the lock file and applies a NON-BLOCKING exclusive LockFileEx.
// Contention returns an error; it is never retried.
func (l *driftFillLock) acquire(lockPath string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	pathW, err := windows.UTF16PtrFromString(lockPath)
	if err != nil {
		return fmt.Errorf("drift fill lock utf16 %s: %w", lockPath, err)
	}

	handle, err := windows.CreateFile(
		pathW,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_ALWAYS,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return fmt.Errorf("drift fill lock CreateFile %s: %w", lockPath, err)
	}

	const (
		lockFlagsExclusive = 0x00000002 // LOCKFILE_EXCLUSIVE_LOCK
		lockFlagsImmediate = 0x00000001 // LOCKFILE_FAIL_IMMEDIATELY
		maxLen             = 0xFFFFFFFF
	)
	var overlapped windows.Overlapped
	if err := windows.LockFileEx(
		handle,
		lockFlagsExclusive|lockFlagsImmediate,
		0,
		maxLen,
		maxLen,
		&overlapped,
	); err != nil {
		_ = windows.CloseHandle(handle)
		return fmt.Errorf("drift fill lock LockFileEx %s: %w", lockPath, err)
	}

	l.handle = handle
	return nil
}

// release unlocks and closes the handle. Idempotent.
func (l *driftFillLock) release() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.handle == windows.InvalidHandle {
		return nil
	}
	const maxLen = 0xFFFFFFFF
	var overlapped windows.Overlapped
	unlockErr := windows.UnlockFileEx(l.handle, 0, maxLen, maxLen, &overlapped)
	closeErr := windows.CloseHandle(l.handle)
	// CloseHandle leaves the handle OPEN when it fails, so dropping our only
	// reference to it would leak the handle irrecoverably. Invalidate the field
	// only on a successful close; on failure keep it so a later release() can
	// retry.
	if closeErr == nil {
		l.handle = windows.InvalidHandle
	}
	if unlockErr != nil {
		return unlockErr
	}
	return closeErr
}
