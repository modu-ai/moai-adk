//go:build windows

package session

import (
	"sync"
	"time"
)

// fileLock is the cross-platform advisory-lock interface (Windows variant).
// SPEC-V3R2-RT-004 REQ-040: lock primitive that prevents concurrent checkpoint writes.
type fileLock interface {
	acquire(path string, retries int, backoff time.Duration) error
	release() error
}

// newFileLock returns the platform-specific lock implementation.
// On Windows this is a mutex-only same-process fallback (cross-process support
// will be replaced by LockFileEx later).
// SPEC-V3R2-RT-004 WIP: a native Windows advisory lock is deferred to M4-M5 or a separate SPEC.
func newFileLock() fileLock {
	return &mutexLock{}
}

// mutexLock is the process-level mutex-only fallback used on Windows.
// SPEC-V3R2-RT-004 WIP M3 stage: serializes only goroutine races within the
// same process. Cross-process races are not guaranteed at this stage and will
// be replaced by a LockFileEx-based native implementation in M4-M5 or a
// follow-up SPEC.
type mutexLock struct {
	mu     sync.Mutex
	locked bool
}

// acquire immediately obtains an exclusive lock at the same-process level.
// The retries / backoff arguments are preserved for interface parity with the
// Unix flock implementation but are not used.
func (l *mutexLock) acquire(path string, retries int, backoff time.Duration) error {
	l.mu.Lock()
	l.locked = true
	return nil
}

// release releases the mutex.
func (l *mutexLock) release() error {
	if !l.locked {
		return nil
	}
	l.locked = false
	l.mu.Unlock()
	return nil
}

// acquireWithRetry delegates to immediate mutex acquisition on the Windows fallback.
// The Unix retry policy (3-retry / 10ms-backoff) will be applied once the
// native implementation lands.
func acquireWithRetry(lock fileLock, path string, retries int, backoff time.Duration) error {
	return lock.acquire(path, retries, backoff)
}
