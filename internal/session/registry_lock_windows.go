//go:build windows

package session

import (
	"fmt"
	"sync"

	"golang.org/x/sys/windows"
)

// registryLock is the Windows LockFileEx-based advisory lock used
// exclusively by the multi-session coordination registry. Distinct from
// mutexLock in lock_windows.go which serves SPEC-V3R2-RT-004 checkpoints
// with a same-process-only fallback.
//
// This implementation uses windows.LockFileEx with LOCKFILE_EXCLUSIVE_LOCK
// | LOCKFILE_FAIL_IMMEDIATELY for non-blocking cross-process exclusion
// (parity with POSIX flock LOCK_EX | LOCK_NB).
//
// SPEC-V3R6-MULTI-SESSION-COORD-001 REQ-COORD-008, REQ-COORD-022.
type registryLock struct {
	mu     sync.Mutex
	handle windows.Handle
}

// newRegistryLock returns a fresh registryLock instance.
func newRegistryLock() *registryLock {
	return &registryLock{handle: windows.InvalidHandle}
}

// acquire opens the lock companion file and applies a non-blocking
// exclusive LockFileEx. Returns an error on contention so the caller
// can retry within LockTimeout.
func (l *registryLock) acquire(lockPath string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	pathW, err := windows.UTF16PtrFromString(lockPath)
	if err != nil {
		return fmt.Errorf("registry lock utf16 %s: %w", lockPath, err)
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
		return fmt.Errorf("registry lock CreateFile %s: %w", lockPath, err)
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
		return fmt.Errorf("registry lock LockFileEx %s: %w", lockPath, err)
	}

	l.handle = handle
	return nil
}

// release unlocks and closes the handle. Idempotent.
func (l *registryLock) release() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.handle == windows.InvalidHandle {
		return nil
	}
	const maxLen = 0xFFFFFFFF
	var overlapped windows.Overlapped
	unlockErr := windows.UnlockFileEx(l.handle, 0, maxLen, maxLen, &overlapped)
	closeErr := windows.CloseHandle(l.handle)
	l.handle = windows.InvalidHandle
	if unlockErr != nil {
		return unlockErr
	}
	return closeErr
}
