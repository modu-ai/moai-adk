package verify

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// claimLockTTL is the staleness threshold at which a held lockfile is
// considered abandoned (holder process crashed before releasing) and may be
// reclaimed. Generous by design: a healthy record completes in seconds, so a
// lock older than this is certainly stale, not merely contended.
const claimLockTTL = 5 * time.Minute

// claimLockRetries caps how many times acquireKeyLock retries the O_EXCL claim
// before failing open. With claimLockBackoff spacing, this bounds the worst-case
// wait to ~claimLockRetries × claimLockBackoff.
const claimLockRetries = 200

// claimLockBackoff is the sleep between O_EXCL retries when another consumer
// holds the lock.
const claimLockBackoff = 10 * time.Millisecond

// keyMutexSet provides in-process per-key mutual exclusion. This is the
// PRIMARY serialization layer: it makes RecordCheck's read-modify-write
// deterministic within one process (the directly-tested case — see
// TestRecordCheckConcurrentSameKeyNoDimensionDrop). The file lock below is the
// cross-process layer on top (REQ-004 ¶4 names the Stop-hook-vs-sync-auditor
// cross-process hazard).
type keyMutexSet struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

var keyMu = &keyMutexSet{locks: make(map[string]*sync.Mutex)}

// mutexFor returns the in-process mutex for the given key.
func (s *keyMutexSet) mutexFor(key string) *sync.Mutex {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.locks[key]
	if !ok {
		m = &sync.Mutex{}
		s.locks[key] = m
	}
	return m
}

// lockPath returns the on-disk claim-stamp path for a snapshot key. Sits
// alongside the snapshot file in the same directory.
func lockPath(projectRoot, key string) string {
	return SnapshotPath(projectRoot, key) + ".lock"
}

// acquireKeyLock takes the per-key serialization for a snapshot recording. It
// holds (1) the in-process per-key mutex (deterministic within one process)
// and (2) a cross-process file lock via O_EXCL claim-stamp with staleness
// reclaim (REQ-004 ¶4). Returns a release function the caller MUST defer.
//
// Fail-open policy: if the file lock cannot be acquired within
// claimLockRetries (e.g. a wedged holder on another process), the in-process
// mutex is still held and the recording proceeds WITHOUT the file lock — the
// worst case is the pre-fix read-modify-write race (a dropped dimension), not
// data corruption (Save is still atomic via rename). This matches the
// advisory-check fail-open discipline.
func acquireKeyLock(projectRoot, key string) func() {
	// Layer 1: in-process per-key mutex.
	m := keyMu.mutexFor(key)
	m.Lock()

	// Layer 2: cross-process O_EXCL claim-stamp.
	lp := lockPath(projectRoot, key)
	if err := os.MkdirAll(filepath.Dir(lp), 0o755); err != nil {
		// Cannot create the lock dir — fail open on the file lock (in-process
		// mutex still held). Best-effort, not fatal.
		return func() { m.Unlock() }
	}
	held := tryClaimStamp(lp)
	release := func() {
		if held {
			_ = os.Remove(lp) // best-effort; a missing file is fine (reclaimed by a waiter)
		}
		m.Unlock()
	}
	if held {
		return release
	}

	// Failed to claim on the first try — bounded retry with staleness reclaim.
	for i := 0; i < claimLockRetries; i++ {
		if isStaleClaim(lp) {
			_ = os.Remove(lp) // reclaim an abandoned lock
		}
		if tryClaimStampAt(lp) {
			held = true
			return release
		}
		time.Sleep(claimLockBackoff)
	}
	// Exhausted retries — fail open (in-process mutex still held).
	return release
}

// tryClaimStamp creates the claim file atomically (O_CREATE|O_EXCL) and stamps
// it with the holder PID + timestamp for staleness diagnosis. Returns true on
// success.
func tryClaimStamp(lp string) bool { return tryClaimStampAt(lp) }

// tryClaimStampAt is the per-path claim attempt, separated for the retry loop.
func tryClaimStampAt(lp string) bool {
	f, err := os.OpenFile(lp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return false
	}
	_, _ = fmt.Fprintf(f, "pid=%d os=%s stamped_at=%d\n", os.Getpid(), runtime.GOOS, time.Now().UnixNano())
	_ = f.Close()
	return true
}

// isStaleClaim reports whether the claim file at lp is older than claimLockTTL
// and thus abandoned (the holder crashed before releasing). A non-existent
// file is NOT stale (the caller will O_EXCL-create it).
func isStaleClaim(lp string) bool {
	info, err := os.Stat(lp)
	if err != nil {
		return false
	}
	return time.Since(info.ModTime()) > claimLockTTL
}
