package spec

import (
	"fmt"
	"os"
	"path/filepath"
)

// drift_fill_lock.go — the cross-process single-flight lock for the out-of-band
// drift-cache fill (REQ-DCF-006 / REQ-DCF-007 / REQ-DCF-008).
//
// Two layers guard the fill, and neither replaces the other:
//
//   - The SessionStart handler claims a suppression RECORD before spawning a
//     child. That is the cheap spawn-suppressor: it stops a burst of sessions
//     from each starting a child, and it keeps suppressing when a child is so
//     broken it never runs.
//   - The child takes THIS lock before computing. That is the authoritative
//     compute-serialiser: whatever reaches the compute, at most one process
//     performs it.
//
// The lock is held by an OS handle — flock(LOCK_EX|LOCK_NB) on POSIX,
// LockFileEx(EXCLUSIVE|FAIL_IMMEDIATELY) on Windows — so its release is
// performed by process exit, INCLUDING abnormal termination. A pid-file lock
// would survive SIGKILL and wedge every later fill; that is the failure mode
// this shape exists to make impossible. The pair is copied from
// internal/sessionmsg (itself copied from internal/session) rather than written
// as a variant.
//
// Acquisition is NON-BLOCKING: a loser exits immediately having computed
// nothing. Blocking would serialise sessions instead of skipping them, which is
// the opposite of what a best-effort background fill wants.

// driftFillLockFilename sits beside the cache it guards, in the same
// .moai/state/ runtime-state directory.
const driftFillLockFilename = "drift-cache.lock"

// DriftFillLockPath returns the fill lock's location for a project root.
func DriftFillLockPath(baseDir string) string {
	return filepath.Join(baseDir, ".moai", "state", driftFillLockFilename)
}

// WithDriftFillLock runs fn while holding the exclusive cross-process fill lock.
//
// ran reports whether fn was executed: false means another process holds the
// lock and this caller computed nothing, which is a normal, successful outcome
// and NOT an error. err is non-nil only when the lock file itself could not be
// prepared (an unwritable state directory); callers on the best-effort fill path
// treat that as "do nothing" too.
func WithDriftFillLock(baseDir string, fn func()) (ran bool, err error) {
	path := DriftFillLockPath(baseDir)
	if mkErr := os.MkdirAll(filepath.Dir(path), 0o755); mkErr != nil {
		return false, fmt.Errorf("drift fill lock: prepare %s: %w", filepath.Dir(path), mkErr)
	}

	lock := newDriftFillLock()
	if acqErr := lock.acquire(path); acqErr != nil {
		// Contended (or unopenable): compute nothing and say so. Never retried
		// — the whole point is to skip rather than queue.
		return false, nil
	}
	defer func() { _ = lock.release() }()

	fn()
	return true, nil
}
