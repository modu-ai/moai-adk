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
//
// The owner-record write truncates and rewrites the artifact, so the opener
// first proves the artifact is the project's own file: no symlinked parent
// directory (checkBoardLockAncestors), no symlinked lock file (Lstat, then
// O_NOFOLLOW to close the check-to-open window on the final component), and,
// on the opened descriptor and BEFORE the flock-then-truncate/write, a regular
// file with a single link. A refusal wraps ErrBoardLockUnsafePath.
//
// @MX:ANCHOR: [AUTO] Shared opener for board.lock and the factory worktree-step lock; the owner-record write follows the safety checks.
// @MX:REASON: the truncate+write must never run on a descriptor not proven a single-link regular file inside the project; moving the fstat check after the write reintroduces the outside-file overwrite (card t1458 P1).
// @MX:SPEC: SPEC-FACTORY-ATOMIC-LEASE-001
func acquireBoardLockImpl(lockPath string) (boardLockImpl, error) {
	if err := checkBoardLockAncestors(lockPath); err != nil {
		return nil, err
	}
	if err := checkBoardLockArtifact(lockPath); err != nil {
		return nil, err
	}
	// O_NONBLOCK keeps a FIFO planted at the path from blocking the open; it
	// has no effect on a regular file.
	fd, err := unix.Open(lockPath, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0o644)
	if err != nil {
		if errors.Is(err, unix.ELOOP) {
			return nil, unsafeBoardLockPath(lockPath, "lock file is a symlink")
		}
		if errors.Is(err, unix.EISDIR) {
			return nil, unsafeBoardLockPath(lockPath, "lock file is not a regular file")
		}
		return nil, fmt.Errorf("open board lock %s: %w", lockPath, err)
	}
	var st unix.Stat_t
	if err := unix.Fstat(fd, &st); err != nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("inspect board lock %s: %w", lockPath, err)
	}
	if st.Mode&unix.S_IFMT != unix.S_IFREG {
		_ = unix.Close(fd)
		return nil, unsafeBoardLockPath(lockPath, "lock file is not a regular file")
	}
	if uint64(st.Nlink) != 1 {
		_ = unix.Close(fd)
		return nil, unsafeBoardLockPath(lockPath, "lock file has more than one hard link")
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
