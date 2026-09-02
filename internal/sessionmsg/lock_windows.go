//go:build windows

package sessionmsg

import (
	"fmt"
	"sync"

	"golang.org/x/sys/windows"
)

// agentLock is the Windows LockFileEx-based advisory lock guarding one agent
// mailbox (or the registration critical section) in the session messaging
// broker. The pattern is copied from internal/session
// registry_lock_windows.go (REQ-CSM-009); the frozen internal/session
// package is not modified.
//
// Uses LockFileEx with LOCKFILE_EXCLUSIVE_LOCK |
// LOCKFILE_FAIL_IMMEDIATELY for non-blocking cross-process exclusion
// (parity with POSIX flock LOCK_EX | LOCK_NB).
type agentLock struct {
	mu     sync.Mutex
	handle windows.Handle
}

// newAgentLock returns a fresh agentLock instance.
func newAgentLock() *agentLock {
	return &agentLock{handle: windows.InvalidHandle}
}

// acquire opens the lock companion file and applies a non-blocking
// exclusive LockFileEx. Returns an error on contention so the caller
// (withAgentLock) can retry within the LockTimeout window.
func (l *agentLock) acquire(lockPath string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	pathW, err := windows.UTF16PtrFromString(lockPath)
	if err != nil {
		return fmt.Errorf("sessionmsg lock utf16 %s: %w", lockPath, err)
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
		return fmt.Errorf("sessionmsg lock CreateFile %s: %w", lockPath, err)
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
		return fmt.Errorf("sessionmsg lock LockFileEx %s: %w", lockPath, err)
	}

	l.handle = handle
	return nil
}

// release unlocks and closes the handle. Idempotent.
func (l *agentLock) release() error {
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
	// reference to it would leak the handle irrecoverably. Invalidate the
	// field only on a successful close; on failure keep the handle so a
	// later release() can retry it.
	if closeErr == nil {
		l.handle = windows.InvalidHandle
	}
	if unlockErr != nil {
		return unlockErr
	}
	return closeErr
}
