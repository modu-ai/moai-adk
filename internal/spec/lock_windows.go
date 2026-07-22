//go:build windows

// SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 — Windows per-SPEC close lock.
// Windows lacks fcntl-style advisory flock; we use atomic-create-file (O_CREATE|O_EXCL)
// per design.md §D.2 fallback. Stale lock detection (PID + timestamp embedded) is a
// post-MVP enhancement; M1 leaves stale-lock cleanup as a known-issue requiring
// manual `del .moai/state/spec-close-*.lock`.
package spec

import (
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/windows"
)

const (
	// Delete-pending window: os.Remove on Windows only *marks* a name for
	// deletion; the name lingers in a delete-pending state until the last handle
	// closes, and re-creating that name meanwhile fails with ERROR_ACCESS_DENIED
	// (surfaced as os.ErrPermission) — NOT ERROR_FILE_EXISTS. A search
	// indexer / virus scanner briefly holding the file surfaces the same way via
	// ERROR_SHARING_VIOLATION. Both clear in microseconds, so they need only a
	// tiny budget — NOT the per-holder contention budget. Mirrors migration's
	// lockTransientRetries / lockTransientDelay (~50ms).
	lockTransientRetries = 10
	lockTransientDelay   = 5 * time.Millisecond
)

// atomicFileSpecLock holds an exclusively-created lock file. Releasing the
// lock means removing the file.
type atomicFileSpecLock struct {
	lockPath string
	file     *os.File
}

func (f *atomicFileSpecLock) release() error {
	if f == nil {
		return nil
	}
	if f.file != nil {
		_ = f.file.Close()
		f.file = nil
	}
	if f.lockPath != "" {
		err := os.Remove(f.lockPath)
		f.lockPath = ""
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// acquireSpecCloseLockImpl creates lockPath with O_CREATE|O_EXCL — atomic on
// Windows NTFS.
//
// The three failure classes get distinct handling. Unlike the migration lock,
// genuine contention is NOT retried here: the spec-close contract (AC-LSG-010,
// TestAcquireSpecCloseLock_Contention) is that a lock owned by another holder is
// reported as held *immediately*, never blocked.
//
//   - os.IsExist(err) (another holder owns the lock) → ErrSpecCloseLockHeld,
//     returned immediately (no contention retry loop).
//   - a transient Windows condition (delete-pending from a just-released holder,
//     surfaced as os.ErrPermission / ERROR_ACCESS_DENIED, or an indexer/scanner
//     sharing conflict via ERROR_SHARING_VIOLATION) → a small bounded retry
//     (~50ms). A concurrent release() marks the old lock file delete-pending, so
//     a competing acquirer's O_CREATE|O_EXCL create fails with ERROR_ACCESS_DENIED
//     rather than ERROR_FILE_EXISTS until the delete completes; retrying lets that
//     window clear so the acquirer succeeds (or observes genuine contention).
//   - any other error → wrapped and returned immediately.
//
// This cannot weaken mutual exclusion: the lock is granted on exactly one path —
// the atomic O_CREATE|O_EXCL create, which the OS serializes — and that path is
// untouched. A transient retry only prolongs a spurious-failure wait; it can
// never admit a second holder.
func acquireSpecCloseLockImpl(lockPath string) (specCloseLockImpl, error) {
	transientLeft := lockTransientRetries
	for {
		file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o644)
		if err == nil {
			// Embed PID for future stale-lock detection (not yet acted upon in M1).
			_, _ = fmt.Fprintf(file, "pid=%d\n", os.Getpid())
			return &atomicFileSpecLock{lockPath: lockPath, file: file}, nil
		}
		switch {
		case os.IsExist(err):
			// Another holder owns the lock — report held immediately.
			return nil, ErrSpecCloseLockHeld
		case isTransientSpecLockError(err):
			if transientLeft--; transientLeft <= 0 {
				return nil, fmt.Errorf("open spec-close lock %s: %w", lockPath, err)
			}
			time.Sleep(lockTransientDelay)
		default:
			return nil, fmt.Errorf("open spec-close lock %s: %w", lockPath, err)
		}
	}
}

// isTransientSpecLockError reports whether a failed O_CREATE|O_EXCL acquire is a
// short-lived Windows condition (delete-pending after a concurrent release, or a
// sharing conflict from a search indexer / virus scanner) that a brief retry may
// clear. It deliberately does NOT cover os.IsExist / fs.ErrExist — genuine
// contention is reported as held immediately, never retried. Mirrors the
// transient classification in internal/migration/version_windows.go.
func isTransientSpecLockError(err error) bool {
	return errors.Is(err, os.ErrPermission) ||
		errors.Is(err, windows.ERROR_SHARING_VIOLATION)
}
