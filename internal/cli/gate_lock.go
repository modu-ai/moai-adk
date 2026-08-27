package cli

// gate_lock.go — the gate-run advisory lock and its bounded wait
// (SPEC-GATE-THREE-AXES-001 axis 3, M3).
//
// Two concurrent `moai gate` runs in one checkout each run the full toolchain
// against the same working tree, competing for the same build cache and the
// same machine. This lock serializes them: a starting run acquires it before
// any step executes, waits a bounded budget while another run holds it, and
// degrades to an unserialized run when the budget expires — the lock never
// fails the gate and never blocks without bound.
//
// The substrate reuses the repository's existing cross-process per-scope lock
// PATTERN (internal/spec/lock.go and internal/kanban/board_lock.go with their
// platform counterparts: flock on Unix, atomic-create on Windows); the
// try-semantics give the bounded wait its natural form — attempt, sleep
// briefly, attempt again, until the budget expires. The lock lives in this
// package because this package is its only consumer, the same placement both
// precedents chose; a package of its own would add an import boundary no
// second consumer crosses.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// ErrGateLockHeld is returned by AcquireGateLock when another process holds
// the gate-run lock.
var ErrGateLockHeld = errors.New("gate-run lock held")

// IsGateLockHeld reports whether err is the contention sentinel.
func IsGateLockHeld(err error) bool {
	return errors.Is(err, ErrGateLockHeld)
}

// ErrGateLockChangedHands is returned by ClearStaleGateLock when the
// pre-removal re-read observes a different recorded identity than the
// inspection did — the artifact was released and re-acquired inside the
// window, and the clear aborts rather than unlinking a valid lock.
var ErrGateLockChangedHands = errors.New("gate-run lock changed hands between inspection and removal")

// IsGateLockChangedHands reports whether err is the changed-hands abort.
func IsGateLockChangedHands(err error) bool {
	return errors.Is(err, ErrGateLockChangedHands)
}

// gateLockFileName names the lock artifact inside the project's state dir.
const gateLockFileName = "gate-run.lock"

// gateLockRetryDelay is the brief sleep between acquisition attempts in the
// bounded wait loop — short enough that a run starting the moment the holder
// releases begins within a fraction of a second, long enough that the loop
// does not spin.
const gateLockRetryDelay = 100 * time.Millisecond

// GateLockOwner is the creating process's identity recorded IN the lock
// artifact. The identity is what a waiting run names in its notice and what
// makes a stale artifact distinguishable from a live holder's: without it,
// "the holder is gone" is a guess, and clearing on a guess unlinks a lock a
// live process may hold.
type GateLockOwner struct {
	PID       int    `json:"pid"`
	CreatedAt string `json:"created_at"`
}

// GateRunLock represents an acquired gate-run lock. Callers MUST call Release
// when the gate run completes.
type GateRunLock struct {
	path string
	impl gateLockImpl
}

// gateLockImpl is the platform-specific lock implementation (flock on Unix,
// atomic-create on Windows — mirroring internal/spec's substrate split).
type gateLockImpl interface {
	release() error
}

// AcquireGateLock acquires the gate-run lock at
// <projectDir>/.moai/state/gate-run.lock, creating the state directory if
// absent. Returns ErrGateLockHeld on contention; the caller retries, degrades,
// or reports — never blocks here.
//
// The acquiring process records its identity in the artifact as part of the
// acquisition, so an artifact always names its current owner.
func AcquireGateLock(projectDir string) (*GateRunLock, error) {
	dir := gateLockDir(projectDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("acquire gate-run lock: creating state dir: %w", err)
	}
	path := gateLockPath(projectDir)
	impl, err := acquireGateLockImpl(path)
	if err != nil {
		return nil, err
	}
	return &GateRunLock{path: path, impl: impl}, nil
}

// Release releases the gate-run lock. Safe to call multiple times.
func (l *GateRunLock) Release() error {
	if l == nil || l.impl == nil {
		return nil
	}
	err := l.impl.release()
	l.impl = nil
	return err
}

// Path returns the lock artifact's path (diagnostics).
func (l *GateRunLock) Path() string {
	if l == nil {
		return ""
	}
	return l.path
}

// gateLockDir returns the lock directory beneath projectDir.
func gateLockDir(projectDir string) string {
	return filepath.Join(projectDir, ".moai", "state")
}

// gateLockPath returns the lock artifact's path beneath projectDir.
func gateLockPath(projectDir string) string {
	return filepath.Join(gateLockDir(projectDir), gateLockFileName)
}

// ReadGateLockHolder reads the identity recorded in the lock artifact. A
// contender uses it to name the holder it is waiting on. Best-effort: an
// absent or unparseable artifact returns an error the caller treats as "no
// observable holder", never as a lock failure.
func ReadGateLockHolder(projectDir string) (GateLockOwner, error) {
	raw, err := os.ReadFile(gateLockPath(projectDir))
	if err != nil {
		return GateLockOwner{}, fmt.Errorf("reading gate-run lock holder: %w", err)
	}
	return parseGateLockOwner(raw)
}

// parseGateLockOwner decodes a lock artifact's recorded owner identity.
func parseGateLockOwner(raw []byte) (GateLockOwner, error) {
	var owner GateLockOwner
	if err := json.Unmarshal(raw, &owner); err != nil {
		return GateLockOwner{}, fmt.Errorf("parsing gate-run lock owner identity: %w", err)
	}
	if owner.PID <= 0 {
		return GateLockOwner{}, fmt.Errorf("parsing gate-run lock owner identity: pid %d is not a process identity", owner.PID)
	}
	return owner, nil
}

// newGateLockOwnerRecord builds the owner identity block written at
// acquisition.
func newGateLockOwnerRecord() []byte {
	owner := GateLockOwner{
		PID:       os.Getpid(),
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	encoded, err := json.MarshalIndent(owner, "", "  ")
	if err != nil {
		// Marshalling a two-field struct cannot fail; fall back to the
		// minimal record rather than failing the acquisition.
		return []byte(fmt.Sprintf("{\"pid\":%d}\n", owner.PID))
	}
	return append(encoded, '\n')
}

// GateLockClearReport is what a clear operation observed and did — the clear
// is explicit and operator-visible, so it REPORTS rather than acting
// silently.
type GateLockClearReport struct {
	// Removed is true only when the artifact was unlinked by this call.
	Removed bool
	// PID names the recorded owner the decision was made about.
	PID int
	// Reason names the observation that decided the outcome.
	Reason string
}

// gateLockWaitResult reports how a bounded wait ended. Exactly one of
// lock-acquired, budget-expired, and machinery-failure holds; the line is the
// verdict text runGate surfaces for the two outcomes that did not acquire.
type gateLockWaitResult struct {
	// lock is non-nil when the wait acquired the lock.
	lock *GateRunLock
	// holderPID names the holder the run waited on (0 when none was
	// observed). It is what the waiting notice prints.
	holderPID int
	// waited is the wall-clock span spent in the wait loop.
	waited time.Duration
	// line is the lock-outcome verdict line for stderr; empty when the lock
	// was acquired. It is a REPORT — the gate's own verdict, not this line,
	// decides the exit code.
	line string
}

// waitForGateLock runs the bounded acquisition loop: attempt, sleep briefly,
// attempt again, until the budget expires. On expiry the result carries no
// lock and the caller proceeds unserialized — a one-way degradation, which
// the loop's shape guarantees: once it returns, nothing re-attempts the
// acquisition for the remainder of that run.
//
// Where the artifact records a holder process that is no longer alive, the
// stale artifact is cleared (with the changed-hands abort) and the retry is
// immediate, so a dead holder does not cost the full budget.
//
// A lock-machinery failure (the state directory cannot be created, the
// artifact cannot be opened) is reported through the result and the caller
// proceeds unserialized; it never fails the run.
func waitForGateLock(projectDir string, budget time.Duration, w io.Writer) gateLockWaitResult {
	started := time.Now()
	deadline := started.Add(budget)
	lastNotifiedPID := 0
	for {
		lock, err := AcquireGateLock(projectDir)
		if err == nil {
			return gateLockWaitResult{lock: lock, waited: time.Since(started)}
		}
		if !IsGateLockHeld(err) {
			return gateLockWaitResult{
				waited: time.Since(started),
				line:   fmt.Sprintf("gate-run lock: unavailable (%v) — running unserialized", err),
			}
		}

		// Contention. Name the holder in the notice — a PID, knowable only
		// by reading the artifact, is what distinguishes this from a generic
		// "another run is in progress".
		holder, herr := ReadGateLockHolder(projectDir)
		if herr == nil && holder.PID > 0 {
			resHolder := holder.PID
			if resHolder != lastNotifiedPID {
				lastNotifiedPID = resHolder
				_, _ = fmt.Fprintf(w, "gate-run lock: held by pid %d — waiting (budget %s)\n", resHolder, budget)
			}
			if !kanban.FactoryProcessAlive(resHolder) {
				// The recorded holder is gone. Clear the stale artifact and
				// retry immediately rather than sleeping the delay first; a
				// changed-hands abort means a live process holds it now, and
				// the retry observes that as ordinary contention.
				report, cerr := ClearStaleGateLock(projectDir)
				switch {
				case cerr != nil && !IsGateLockChangedHands(cerr):
					_, _ = fmt.Fprintf(w, "gate-run lock: stale clear failed (%v); continuing to wait\n", cerr)
				case cerr == nil && report != nil && report.Removed:
					_, _ = fmt.Fprintf(w, "gate-run lock: cleared stale artifact of terminated pid %d\n", resHolder)
				}
				if time.Now().After(deadline) {
					return gateLockWaitResult{
						holderPID: resHolder,
						waited:    time.Since(started),
						line:      fmt.Sprintf("gate-run lock: wait budget %s expired — running unserialized", budget),
					}
				}
				continue
			}
		}

		// The budget is waited out in full, not rounded down to the retry
		// delay: a run that degrades at W−delay started its unserialized work
		// before the budget the config promised had elapsed. The final sleep
		// is the remaining slice.
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return gateLockWaitResult{
				holderPID: lastNotifiedPID,
				waited:    time.Since(started),
				line:      fmt.Sprintf("gate-run lock: wait budget %s expired — running unserialized", budget),
			}
		}
		sleepFor := gateLockRetryDelay
		if remaining < sleepFor {
			sleepFor = remaining
		}
		time.Sleep(sleepFor)
	}
}
