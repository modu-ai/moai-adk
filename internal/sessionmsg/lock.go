package sessionmsg

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// lockNameRegister is the advisory-lock name for the registration critical
// section (scan-for-kind+name then create). Distinct from per-agent mailbox
// locks so registrations never contend with message traffic.
const lockNameRegister = "register"

// Retry backoff for the non-blocking lock loop (internal/session registry
// precedent: base 5ms, cap 50ms, full jitter). The cap stays far below
// LockTimeout so one sleep can never consume the whole acquisition budget.
const (
	lockBackoffBase     = 5 * time.Millisecond
	lockBackoffCap      = 50 * time.Millisecond
	lockBackoffMaxShift = 8 // ceiling saturates well before this; guards the shift
)

// lockRetryDelay returns how long to sleep before retry number attempt
// (0-based) of the acquisition loop: full jitter over an exponentially
// growing ceiling, clamped to the time remaining before the deadline.
// NB-flock has no kernel fairness queue, so a fixed sleep would synchronize
// contenders; jitter breaks that cadence (cross-process mitigation —
// same-process contenders are serialized by the in-process mutex below).
func lockRetryDelay(attempt int, remaining time.Duration) time.Duration {
	if remaining <= 0 {
		return 0
	}
	shift := attempt
	if shift > lockBackoffMaxShift {
		shift = lockBackoffMaxShift
	}
	ceiling := lockBackoffBase << shift
	if ceiling > lockBackoffCap {
		ceiling = lockBackoffCap
	}
	delay := time.Duration(rand.Int64N(int64(ceiling)) + 1) // full jitter over (0, ceiling]
	if delay > remaining {
		delay = remaining
	}
	return delay
}

// inProcessMutexes is the path-keyed map of in-process mutexes
// (internal/session registry precedent, SPEC-V3R6-CI-FLAKY-STABILIZE-003
// pattern — copied, not imported). Each distinct absolute lock path gets its
// own *sync.Mutex so all Store instances sharing a path serialize against
// each other in-process; NB-flock remains the cross-process gate. Without
// this mutex, same-process goroutines contend on the unfair NB-flock and a
// repeated loser can starve past lockTimeout.
//
// @MX:WARN: [AUTO] package-global sync.Map grows with distinct lock paths over the process lifetime
// @MX:REASON: path-keyed in-process mutex registry; bounded in production (one path per agent mailbox per process) but unbounded in principle — mirrors the accepted internal/session registry tradeoff.
var inProcessMutexes sync.Map // map[string]*sync.Mutex

// withAgentLock acquires the advisory lock named lockName (an agentId for
// mailbox operations, lockNameRegister for registration), runs fn, and
// releases the lock. On contention it retries with jittered backoff until
// lockTimeout, then returns ErrLockTimeout. fn MUST NOT acquire another
// agent lock (no nesting — callers sequence lock scopes instead, which keeps
// the lock graph acyclic by construction).
func (s *Store) withAgentLock(lockName string, fn func() error) (err error) {
	lockPath := s.lockPath(lockName)
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return fmt.Errorf("sessionmsg: mkdir locks: %w", err)
	}

	// In-process mutex: serializes same-process contenders deterministically.
	// A failed Abs falls back to the path as given rather than the empty
	// string — an empty key would silently collapse every distinct lock onto
	// one shared mutex for the process lifetime.
	absLockPath, absErr := filepath.Abs(lockPath)
	if absErr != nil {
		absLockPath = lockPath
	}
	v, _ := inProcessMutexes.LoadOrStore(absLockPath, &sync.Mutex{})
	ipm := v.(*sync.Mutex)
	ipm.Lock()
	defer ipm.Unlock()

	lock := newAgentLock()
	timeout := s.lockTimeout
	if timeout <= 0 {
		timeout = LockTimeout
	}
	deadline := time.Now().Add(timeout)
	for attempt := 0; ; attempt++ {
		err := lock.acquire(lockPath)
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%w: %v", ErrLockTimeout, err)
		}
		time.Sleep(lockRetryDelay(attempt, time.Until(deadline)))
	}
	// A release failure is surfaced, not swallowed — but it must never mask
	// the outcome of fn(), which is the caller's real result. Only a clean fn
	// lets the release error become the returned error.
	defer func() {
		if relErr := lock.release(); relErr != nil && err == nil {
			err = fmt.Errorf("sessionmsg: release lock %s: %w", lockPath, relErr)
		}
	}()

	return fn()
}
