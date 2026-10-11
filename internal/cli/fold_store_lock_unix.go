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
// mode 0644, O_CREAT|O_RDWR|O_CLOEXEC|O_NOFOLLOW (a symlink is refused), and
// never removed on release
// (plan.md D-3).
func (l *foldStoreLock) tryAcquire(dir string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	path := filepath.Join(dir, foldLockFileName)
	// O_NOFOLLOW: a symlinked lock file is refused, never followed, so the lock
	// is never taken on a file outside the store (post-close gate finding 4).
	fd, err := unix.Open(path, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0o644)
	if err != nil {
		return fmt.Errorf("memory fold: store lock open %s: %w", path, err)
	}
	// Every other non-regular shape (a FIFO standing in for the lock, a device
	// node) is refused the same way.
	var st unix.Stat_t
	if err := unix.Fstat(fd, &st); err != nil {
		_ = unix.Close(fd)
		return fmt.Errorf("memory fold: stat the store lock %s: %w", path, err)
	}
	if st.Mode&unix.S_IFMT != unix.S_IFREG {
		_ = unix.Close(fd)
		return fmt.Errorf("memory fold: the store lock %s is not a regular file — refusing", path)
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
	return l.acquireUntil(dir, nil)
}

// acquireUntil is acquire with an abandonment-aware wait: a non-nil forbidden
// is polled on every contended retry, and a true reading ends the wait at
// once, holding nothing. A bounded step whose caller gave up therefore never
// keeps waiting for a hold its holder may not release soon. The deadline
// still bounds the wait.
func (l *foldStoreLock) acquireUntil(dir string, forbidden func() bool) error {
	deadline := time.Now().Add(foldStoreLockRetryDeadline)
	for {
		err := l.tryAcquire(dir)
		if err == nil {
			return nil
		}
		if !errors.Is(err, errFoldStoreContended) {
			return err
		}
		if forbidden != nil && forbidden() {
			return fmt.Errorf("memory fold: the step was abandoned while waiting for the store lock %s", filepath.Join(dir, foldLockFileName))
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
