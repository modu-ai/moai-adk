//go:build windows

package cli

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/sys/windows"
)

// foldStoreLockRetryDeadline bounds one fold store lock acquisition
// (plan.md D-2): the requesting fold retries a non-blocking LockFileEx
// until the deadline and then refuses cleanly, naming the contended store
// (REQ-MRR-003). Two seconds — the same magnitude as the close-path bound's
// production ceiling.
const foldStoreLockRetryDeadline = 2 * time.Second

// foldStoreLockRetryInterval is the pause between two non-blocking attempts
// inside one acquisition's bounded retry loop.
const foldStoreLockRetryInterval = 10 * time.Millisecond

// foldStoreLock is the per-store advisory lock the fold apply path holds
// for the whole applyFold span (REQ-MRR-001). One lock file per store
// directory, a deterministic name derived from the store (REQ-MRR-002);
// both write surfaces of a fold run under it. The Windows parity of
// internal/cli/fold_store_lock_unix.go: LockFileEx with
// LOCKFILE_EXCLUSIVE_LOCK | LOCKFILE_FAIL_IMMEDIATELY is the non-blocking
// cross-process exclusion (pattern copied from
// internal/sessionmsg/lock_windows.go; the frozen origins are not
// modified). The lock's lifetime is bound to the handle, so a killed
// holder releases the lock — no stale-lock sweep exists by construction.
type foldStoreLock struct {
	mu     sync.Mutex
	handle windows.Handle
}

// errFoldStoreContended marks a non-blocking LockFileEx refusal — the lock
// file exists and another holder keeps it. It is the one condition acquire
// retries; an open failure surfaces immediately.
var errFoldStoreContended = errors.New("store lock held by another writer")

// newFoldStoreLock returns a fresh foldStoreLock instance.
func newFoldStoreLock() *foldStoreLock {
	return &foldStoreLock{handle: windows.InvalidHandle}
}

// tryAcquire makes ONE non-blocking attempt to take the store's lock file
// at <dir>/.moai-fold.lock. A nil return means this lock object holds the
// byte-range lock; contention returns errFoldStoreContended immediately —
// the caller decides to retry or refuse. The lock file is created on first
// acquire (OPEN_ALWAYS) and never removed on release (plan.md D-3).
func (l *foldStoreLock) tryAcquire(dir string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	path := filepath.Join(dir, foldLockFileName)
	pathW, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return fmt.Errorf("memory fold: store lock utf16 %s: %w", path, err)
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
		return fmt.Errorf("memory fold: store lock open %s: %w", path, err)
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
		return fmt.Errorf("memory fold: %w (%s)", errFoldStoreContended, path)
	}
	l.handle = handle
	return nil
}

// acquire takes the store lock, retrying the non-blocking attempt until
// foldStoreLockRetryDeadline. On deadline expiry it returns the clean
// refusal of REQ-MRR-003 — write nothing, exit non-zero, name the contended
// store — never an unbounded block.
func (l *foldStoreLock) acquire(dir string) error {
	deadline := time.Now().Add(foldStoreLockRetryDeadline)
	for {
		err := l.tryAcquire(dir)
		if err == nil {
			return nil
		}
		if !errors.Is(err, errFoldStoreContended) {
			return err
		}
		if !time.Now().Before(deadline) {
			return fmt.Errorf("memory fold: the store lock on %s stayed held for %s — refusing to write (lock file %s)", dir, foldStoreLockRetryDeadline, filepath.Join(dir, foldLockFileName))
		}
		time.Sleep(foldStoreLockRetryInterval)
	}
}

// release unlocks the byte range and closes the handle. Idempotent.
func (l *foldStoreLock) release() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.handle == windows.InvalidHandle {
		return nil
	}
	const maxLen = 0xFFFFFFFF
	var overlapped windows.Overlapped
	unlockErr := windows.UnlockFileEx(l.handle, 0, maxLen, maxLen, &overlapped)
	closeErr := windows.CloseHandle(l.handle)
	// CloseHandle leaves the handle OPEN when it fails, so dropping our
	// only reference to it would leak the handle irrecoverably. Invalidate
	// the field only on a successful close; on failure keep the handle so a
	// later release() can retry it.
	if closeErr == nil {
		l.handle = windows.InvalidHandle
	}
	if unlockErr != nil {
		return unlockErr
	}
	return closeErr
}
