//go:build windows

package lockfile

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

// @MX:NOTE: [AUTO] Windows locks are LockFileEx byte-range locks, so they exclude
// across processes as well as across handles in one process (card t2001,
// user decision R162). The earlier implementation was a process-local mutex
// map (SPEC-AGENT-TEAM-RETIRE-001 REQ-ATR-001 kept it verbatim as "behavior
// preservation") and could not exclude a second PROCESS; the state-file lock
// of SPEC-WORKTREE-HOME-ROOT-001 REQ-WHR-017 and its cross-process tests need
// that exclusion, so the Windows arm now matches what flock gives on Unix.
//
// The locked region is ONE BYTE FAR BEYOND ANY FILE DATA (offset
// 0xFFFFFFFF00000000), not byte 0. Win32 byte-range locks are mandatory, so a
// lock on bytes 0.. makes another handle's ReadFile fail with
// ERROR_LOCK_VIOLATION; call sites that lock a data file and read it back
// through a separate handle (internal/cli/taskledger, internal/harness
// retention) depend on the lock being advisory, as flock is. A region past the
// end of the file is permitted by LockFileEx and overlaps no data, so lockers
// exclude each other while data I/O is unaffected. A lock belongs to the file
// handle: another handle on the same path — in this process or another —
// cannot take it until it is released, and closing the handle (or the holder
// dying) releases it. The exported contract is unchanged: Lock blocks, TryLock
// returns at once, Unlock of a file that holds no lock is not an error.
//
// COMPATIBILITY: a process running the earlier byte-0 build and one running
// this build do not exclude each other (they lock different regions), so the
// two must not be mixed against the same lock file.

// errLockHeld reports a try-acquisition refused because another holder owns
// the lock (Windows TryLock arm).
var errLockHeld = errors.New("lock already held")

// lockOffsetHigh puts the one-byte lock region at offset 0xFFFFFFFF00000000.
const lockOffsetHigh = 0xFFFFFFFF

// lockOverlapped names the lock region. Lock, TryLock and Unlock must all use
// it: an unlock on a different region than the lock would fail to release.
func lockOverlapped() *windows.Overlapped {
	return &windows.Overlapped{OffsetHigh: lockOffsetHigh}
}

// lockRegion locks or tries to lock the one-byte region with the given flags.
func lockRegion(f *os.File, flags uint32) error {
	return windows.LockFileEx(windows.Handle(f.Fd()), flags, 0, 1, 0, lockOverlapped())
}

// Lock acquires an exclusive byte-range lock on the file, blocking until
// another holder (in any process) releases it.
func Lock(f *os.File) error {
	return lockRegion(f, windows.LOCKFILE_EXCLUSIVE_LOCK)
}

// TryLock acquires the exclusive lock without blocking, returning an error
// when another holder owns it (SPEC-WORKTREE-HOME-ROOT-001 M5, design.md D-12:
// the state-file lock is try-and-bounded-wait). A refusal wraps errLockHeld.
func TryLock(f *os.File) error {
	err := lockRegion(f, windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY)
	if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
		return errLockHeld
	}
	return err
}

// Unlock releases the lock on the file. Releasing a file that holds no lock is
// a no-op, as flock(LOCK_UN) is on an unlocked descriptor: call sites unlock
// from deferred cleanup paths that can run before a successful Lock.
func Unlock(f *os.File) error {
	err := windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, lockOverlapped())
	if errors.Is(err, windows.ERROR_NOT_LOCKED) {
		return nil
	}
	return err
}
