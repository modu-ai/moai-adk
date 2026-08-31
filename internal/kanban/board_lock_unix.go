//go:build !windows

// board_lock_unix.go — Unix substrate of the board-wide lock: flock(2) on an
// open descriptor, mirroring internal/spec/lock_unix.go's pattern. The kernel
// releases the flock when the descriptor closes, which it does on process
// exit — so a killed holder leaves an artifact that blocks nothing on this
// platform; the stale-lock requirement exists for the Windows substrate.
package kanban

import (
	"errors"
	"fmt"

	"golang.org/x/sys/unix"
)

// classifyBoardFlockErr maps a flock(2) failure to the board-lock error
// contract: EWOULDBLOCK/EAGAIN is contention, every other errno is a hard
// error that preserves the underlying errno for errors.Is inspection and
// names the lock path (SPEC-BOARDLOCK-ERRNO-001 REQ-BLE-001/002/003).
//
// This is defensive narrowing, NOT a repair of an observed misclassification:
// the measured misclassification count at the sole call site is ZERO. Of the
// three cases the plan-phase probe induced, only EWOULDBLOCK/EAGAIN was
// reachable through the real path, and it was already classified correctly.
//
// Both EWOULDBLOCK and EAGAIN are named although they share a value on linux
// and darwin — that is portability notation, not redundancy.
//
// EINTR falls on the non-contention side. Today it is absorbed by the
// contention retry budget in acquireBoardLockSerialized; here it becomes an
// immediate hard error. That is an accepted behaviour change on an input
// whose reachability is UNMEASURED, recorded as such in spec.md §1.3.1, and
// it sits outside REQ-BLE-005's measured-reachable scope.
func classifyBoardFlockErr(err error, lockPath string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN) {
		return ErrBoardLockHeld
	}
	return fmt.Errorf("lock board lock %s: %w", lockPath, err)
}

// flockBoardLock holds an open descriptor with flock(LOCK_EX|LOCK_NB) held.
type flockBoardLock struct {
	fd int
}

func (f *flockBoardLock) release() error {
	if f == nil || f.fd == 0 {
		return nil
	}
	// Close releases the flock atomically.
	err := unix.Close(f.fd)
	f.fd = 0
	return err
}

// acquireBoardLockImpl opens lockPath O_CREAT|O_RDWR and applies a
// non-blocking exclusive flock, then records this process's identity IN the
// artifact (REQ-KB-023). Only the lock holder writes, so the recorded
// identity always names the current owner; a contender that fails the flock
// writes nothing and the previous owner's record stands.
func acquireBoardLockImpl(lockPath string) (boardLockImpl, error) {
	fd, err := unix.Open(lockPath, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open board lock %s: %w", lockPath, err)
	}
	if err := unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = unix.Close(fd)
		return nil, classifyBoardFlockErr(err, lockPath)
	}

	record := newLockOwnerRecord()
	if err := unix.Ftruncate(fd, 0); err != nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("record board lock owner (truncate) %s: %w", lockPath, err)
	}
	if _, err := unix.Write(fd, record); err != nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("record board lock owner %s: %w", lockPath, err)
	}
	return &flockBoardLock{fd: fd}, nil
}
