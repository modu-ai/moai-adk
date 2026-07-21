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
	lockMaxRetries = 500 // ~5s @ 10ms/retry
	lockRetryDelay = 10 * time.Millisecond
)

// acquireLock acquires the mutex by creating the lock file with O_CREATE|O_EXCL,
// retrying on transient contention until the budget is exhausted.
func acquireLock(lockPath string) (*lockHandle, error) {
	var lastErr error
	for i := 0; i < lockMaxRetries; i++ {
		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY|os.O_TRUNC, 0600)
		if err == nil {
			_ = f.Close()
			return &lockHandle{path: lockPath}, nil
		}
		if !isLockContention(err) {
			return nil, err
		}
		lastErr = err
		time.Sleep(lockRetryDelay)
	}
	return nil, fmt.Errorf("lock acquire timeout (windows) after %d attempts: %w", lockMaxRetries, lastErr)
}

// isLockContention reports whether a failed O_CREATE|O_EXCL acquire is a
// transient condition a later retry may clear, rather than a hard failure:
//
//   - fs.ErrExist — another holder owns the lock. The ordinary contention path.
//   - os.ErrPermission (ERROR_ACCESS_DENIED) — the previous holder's releaseLock
//     removed the file, but os.Remove on Windows only *marks* a name for
//     deletion: it stays in a "delete pending" state until the last handle
//     closes, and creating that name meanwhile fails with ERROR_ACCESS_DENIED
//     rather than ERROR_FILE_EXISTS. Treating it as fatal lets an ordinary
//     release make the very next acquirer fail outright — which is precisely
//     what made this lock fail under contention on Windows CI.
//   - ERROR_SHARING_VIOLATION — another handle (search indexer, virus scanner)
//     is briefly open on the lock file.
//
// Retrying these can NEVER weaken mutual exclusion. The lock is granted on
// exactly one path — the atomic O_CREATE|O_EXCL create succeeding (CREATE_NEW
// on Windows, which the OS serializes) — and that path is untouched here. A
// retry only prolongs the wait; it can never admit a second holder.
//
// Mirrors the same classification used for lock acquisition in
// internal/cli/preference/toggle.go and for rename/read in internal/atomicfile.
// The os.ErrPermission arm makes a genuinely unwritable lock path (read-only
// directory, ACL denial) fail after the budget instead of immediately; the
// error is still surfaced, wrapped, never swallowed.
func isLockContention(err error) bool {
	return errors.Is(err, fs.ErrExist) ||
		errors.Is(err, os.ErrPermission) ||
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
