// Package spec — cross-platform per-SPEC file lock primitive.
//
// SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 REQ-LSG-010 + AC-LSG-010 + AC-LSG-021 (NFR-LSG-005):
// `moai spec close SPEC-XXX` acquires .moai/state/spec-close-<SPEC-ID>.lock to
// prevent concurrent close operations on the same SPEC. Lock scope is per-SPEC,
// not global (different SPECs may close concurrently).
//
// Cross-platform: Unix uses flock(2) advisory lock; Windows uses atomic-create-file
// pattern (O_CREATE|O_EXCL). Per CLAUDE.local.md §14, no naked syscall in body —
// the platform impl lives in lock_unix.go / lock_windows.go.
package spec

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ErrSpecCloseLockHeld is returned by AcquireSpecCloseLock when another process
// holds the per-SPEC lock. The orchestrator translates this to exit code 1 with
// stderr message "Error: another close operation in progress (lock held)" per
// AC-LSG-010.
var ErrSpecCloseLockHeld = errors.New("spec-close lock held")

// SpecCloseLock represents an acquired per-SPEC file lock.
// Callers MUST call Release() (or use the release callback from AcquireSpecCloseLock).
type SpecCloseLock struct {
	specID   string
	lockPath string
	impl     specCloseLockImpl
}

// specCloseLockImpl is the platform-specific lock implementation (interface
// implemented by lock_unix.go's flockImpl + lock_windows.go's atomicFileImpl).
type specCloseLockImpl interface {
	release() error
}

// AcquireSpecCloseLock acquires the per-SPEC close lock at
// .moai/state/spec-close-<SPEC-ID>.lock under stateRoot (project root by default).
//
// Returns:
//   - (*SpecCloseLock, nil) on success — caller MUST call Release()
//   - (nil, ErrSpecCloseLockHeld) when another process holds the lock
//   - (nil, other error) on I/O failure (cannot create lock directory, etc.)
//
// stateRoot is the project root (usually `.` or absolute path). The function
// creates .moai/state/ if absent.
func AcquireSpecCloseLock(stateRoot, specID string) (*SpecCloseLock, error) {
	if specID == "" {
		return nil, fmt.Errorf("AcquireSpecCloseLock: empty specID")
	}

	lockDir := filepath.Join(stateRoot, ".moai", "state")
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		return nil, fmt.Errorf("create lock dir %s: %w", lockDir, err)
	}
	lockPath := filepath.Join(lockDir, fmt.Sprintf("spec-close-%s.lock", specID))

	impl, err := acquireSpecCloseLockImpl(lockPath)
	if err != nil {
		return nil, err
	}

	return &SpecCloseLock{
		specID:   specID,
		lockPath: lockPath,
		impl:     impl,
	}, nil
}

// Release releases the per-SPEC lock and cleans up the lock file.
// Safe to call multiple times (idempotent — subsequent calls return nil).
func (l *SpecCloseLock) Release() error {
	if l == nil || l.impl == nil {
		return nil
	}
	err := l.impl.release()
	l.impl = nil
	return err
}

// LockPath returns the absolute path of the underlying lock file (for diagnostics).
func (l *SpecCloseLock) LockPath() string {
	if l == nil {
		return ""
	}
	return l.lockPath
}

// SpecID returns the SPEC ID this lock is scoped to.
func (l *SpecCloseLock) SpecID() string {
	if l == nil {
		return ""
	}
	return l.specID
}

// IsLockHeldError reports whether err indicates the per-SPEC close lock is
// currently held by another process.
func IsLockHeldError(err error) bool {
	return errors.Is(err, ErrSpecCloseLockHeld)
}
