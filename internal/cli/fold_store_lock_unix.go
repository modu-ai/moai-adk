//go:build !windows

package cli

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

// foldStoreLockRetryDeadline bounds one fold store lock acquisition
// (plan.md D-2): the requesting fold retries a non-blocking flock until the
// deadline and then refuses cleanly, naming the contended store
// (REQ-MRR-003). Two seconds — the same magnitude as the close-path bound's
// production ceiling.
const foldStoreLockRetryDeadline = 2 * time.Second

// foldStoreLockRetryInterval is the pause between two non-blocking flock
// attempts inside one acquisition's bounded retry loop.
const foldStoreLockRetryInterval = 10 * time.Millisecond

// foldStoreLock is the per-store advisory lock the fold apply path holds
// for the whole applyFold span (REQ-MRR-001). One lock file per store
// directory, a deterministic name derived from the store (REQ-MRR-002);
// both write surfaces of a fold run under it. The pattern is copied from
// internal/sessionmsg/lock_unix.go (which copies internal/session); the
// frozen origins are not modified. flock's kernel lifetime means a killed
// holder releases the lock, so no stale-lock sweep exists by construction.
type foldStoreLock struct {
	mu sync.Mutex
	// fd is the lock file's descriptor once held. unacquiredFD (-1, never
	// 0) is the "holds nothing" sentinel: fd 0 is a valid descriptor the
	// kernel hands out whenever stdin is closed, and a 0 sentinel made
	// release() a silent no-op in exactly that case — leaking the flock
	// for the process lifetime (internal/sessionmsg's recorded lesson).
	fd int
}

const unacquiredFD = -1

// errFoldStoreContended marks a non-blocking flock refusal — the lock file
// exists and another holder keeps it. It is the one condition acquire
// retries; an open failure surfaces immediately.
var errFoldStoreContended = errors.New("store lock held by another writer")

// newFoldStoreLock returns a fresh foldStoreLock instance.
func newFoldStoreLock() *foldStoreLock {
	return &foldStoreLock{fd: unacquiredFD}
}

// tryAcquire makes ONE non-blocking attempt to take the store's lock file
// at <dir>/.moai-fold.lock. A nil return means this lock object holds the
// flock; contention returns errFoldStoreContended immediately — the caller
// decides to retry or refuse. The lock file is created on first acquire,
// mode 0644, O_CREAT|O_RDWR|O_CLOEXEC, and never removed on release
// (plan.md D-3).
func (l *foldStoreLock) tryAcquire(dir string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	path := filepath.Join(dir, foldLockFileName)
	fd, err := unix.Open(path, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC, 0o644)
	if err != nil {
		return fmt.Errorf("memory fold: store lock open %s: %w", path, err)
	}
	if ferr := unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB); ferr != nil {
		_ = unix.Close(fd)
		return fmt.Errorf("memory fold: %w (%s)", errFoldStoreContended, path)
	}
	l.fd = fd
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

// release releases the flock and closes the underlying fd. Idempotent.
func (l *foldStoreLock) release() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.fd < 0 {
		return nil
	}
	err := unix.Close(l.fd) // close releases the flock
	l.fd = unacquiredFD
	return err
}
