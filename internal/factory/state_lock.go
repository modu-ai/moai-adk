// state_lock.go — the shared cross-process file-lock substrate (the type, its
// contention sentinels, and the owner-identity record) that the todo queue
// (backlog_store.go), the integration lock (integration_lock_mutation.go), and
// the slot lease (slot_lease.go) each take over their own lock path
// (SPEC-KANBAN-BOARD-001 REQ-KB-019/023, kept after the board went:
// SPEC-LAUNCHER-ENTRY-FLAGS-001 M6).
//
// The substrate reuses the repository's existing cross-process per-scope lock
// PATTERN (internal/spec/lock.go and its platform counterparts: flock on Unix,
// atomic-create on Windows); internal/lockfile's in-process mutex is neither
// used nor upgraded.
package factory

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

// ErrStateLockHeld is returned by a state-lock acquisition when another
// process holds the lock at that path.
var ErrStateLockHeld = errors.New("factory state lock held")

// IsStateLockHeld reports whether err is the contention sentinel.
func IsStateLockHeld(err error) bool {
	return errors.Is(err, ErrStateLockHeld)
}

// ErrStateLockChangedHands is returned by the Windows stale-lock clear when the
// pre-removal re-read observes a different recorded identity than the
// inspection did — the artifact was released and re-acquired inside the
// window, and the clear aborts rather than unlinking a valid lock.
var ErrStateLockChangedHands = errors.New("factory state lock changed hands between inspection and removal")

// IsStateLockChangedHands reports whether err is the changed-hands abort.
func IsStateLockChangedHands(err error) bool {
	return errors.Is(err, ErrStateLockChangedHands)
}

// StateLockOwner is the creating process's identity recorded IN the lock
// artifact (REQ-KB-023). The identity is what makes a stale artifact
// distinguishable from a live holder's: without it, "the holder is gone" is a
// guess, and clearing on a guess unlinks a lock a live process may hold.
type StateLockOwner struct {
	PID       int    `json:"pid"`
	CreatedAt string `json:"created_at"`
}

// StateLock represents an acquired state lock. Callers MUST call Release when
// the guarded read-modify-write completes.
type StateLock struct {
	path string
	impl stateLockImpl
}

// stateLockImpl is the platform-specific lock implementation (flock on Unix,
// atomic-create on Windows — mirroring internal/spec's substrate split).
type stateLockImpl interface {
	release() error
}

// Release releases the state lock. Safe to call multiple times.
func (l *StateLock) Release() error {
	if l == nil || l.impl == nil {
		return nil
	}
	err := l.impl.release()
	l.impl = nil
	return err
}

// Path returns the lock artifact's path (diagnostics).
func (l *StateLock) Path() string {
	if l == nil {
		return ""
	}
	return l.path
}

// ClearStaleReport is what a clear operation observed and did — the operation
// is explicit and operator-visible, so it REPORTS rather than acting silently
// (REQ-KB-023).
type ClearStaleReport struct {
	// Removed is true only when the artifact was unlinked by this call.
	Removed bool
	// PID names the recorded owner the decision was made about.
	PID int
	// Reason names the observation that decided the outcome.
	Reason string
}

// newLockOwnerRecord builds the owner identity block written at acquisition.
func newLockOwnerRecord() []byte {
	owner := StateLockOwner{
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
