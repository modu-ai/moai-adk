//go:build windows

package migration

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"

	"golang.org/x/sys/windows"
)

// lockHandle holds the path of a Windows file mutex lock file.
// Windows has no equivalent to unix.Flock advisory lock, so we substitute a
// file mutex based on O_CREATE|O_EXCL (atomic create-or-fail).
// The REQ-V3R2-RT-007-031 advisory-lock semantics are simplified under the single-user assumption.
type lockHandle struct{ path string }

const (
	// Ordinary contention: another holder owns the lock (fs.ErrExist). Budget
	// ~1s, unchanged from the original design.
	lockMaxRetries = 100
	lockRetryDelay = 10 * time.Millisecond

	// Delete-pending window: os.Remove on Windows only *marks* a name for
	// deletion; it lingers in a delete-pending state until the last handle
	// closes, and re-creating that name meanwhile fails with
	// ERROR_ACCESS_DENIED (surfaced as os.ErrPermission) rather than
	// ERROR_FILE_EXISTS. This window clears in microseconds, so it needs only a
	// tiny budget — NOT the full contention budget. A genuinely unwritable lock
	// path (read-only dir, ACL denial) therefore fails in ~50ms, not seconds,
	// which is what kept slow permission-denied paths from ballooning package
	// test time on Windows CI.
	lockTransientRetries = 10
	lockTransientDelay   = 5 * time.Millisecond
)

// acquireLock acquires the mutex by creating the lock file with O_CREATE|O_EXCL.
//
// The two failure classes get different budgets so a genuine permission failure
// is not retried as long as real lock contention:
//   - fs.ErrExist (another holder) → lockMaxRetries (~1s)
//   - a transient Windows delete-pending / sharing conflict → lockTransientRetries (~50ms)
//
// Neither budget can weaken mutual exclusion: the lock is granted on exactly one
// path — the atomic O_CREATE|O_EXCL create, which the OS serializes — and that
// path is untouched. A retry only prolongs the wait; it can never admit a second
// holder. A non-retryable error returns immediately.
func acquireLock(lockPath string) (*lockHandle, error) {
	var lastErr error
	contentionLeft := lockMaxRetries
	transientLeft := lockTransientRetries
	for {
		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY|os.O_TRUNC, 0600)
		if err == nil {
			_ = f.Close()
			return &lockHandle{path: lockPath}, nil
		}
		lastErr = err
		switch {
		case errors.Is(err, fs.ErrExist):
			if contentionLeft--; contentionLeft <= 0 {
				return nil, fmt.Errorf("lock acquire timeout (windows) after %d contention retries: %w", lockMaxRetries, lastErr)
			}
			time.Sleep(lockRetryDelay)
		case isTransientLockError(err):
			if transientLeft--; transientLeft <= 0 {
				return nil, fmt.Errorf("lock acquire failed (windows) after %d transient retries: %w", lockTransientRetries, lastErr)
			}
			time.Sleep(lockTransientDelay)
		default:
			return nil, err
		}
	}
}

// isTransientLockError reports whether a failed O_CREATE|O_EXCL acquire is a
// short-lived Windows condition (delete-pending or a sharing conflict from a
// search indexer / virus scanner) that a brief retry may clear. It deliberately
// does NOT cover fs.ErrExist, which acquireLock handles on its own longer
// budget. Mirrors the transient classification in internal/atomicfile.
func isTransientLockError(err error) bool {
	return errors.Is(err, os.ErrPermission) ||
		errors.Is(err, windows.ERROR_SHARING_VIOLATION)
}

// releaseLock releases the mutex by deleting the lock file.
// Treats a missing file as success (idempotent).
func releaseLock(h *lockHandle) error {
	err := os.Remove(h.path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
