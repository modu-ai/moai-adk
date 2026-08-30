// integration_lock_mutation.go — the SHORT-LIVED mutation lock that serializes
// the integration-window record's read-modify-write across processes
// (SPEC-INTEGRATION-LOCK-ATOMIC-001, card t336).
//
// TWO LIFETIMES, never conflated. The integration WINDOW is long-lived: it
// spans many CLI invocations, many turns, and minutes of human-paced work, so
// it is a RECORD whose validity is decided by the recorded holder's liveness
// (integration_lock.go). The lock in THIS file is short-lived: it spans one
// CLI invocation's critical section and dies with the process that took it.
// Nothing here is ever consulted to decide who holds the window, and the
// artifact is never persisted as the window.
//
// The artifact deliberately does NOT share the record's `integration-lock`
// filename stem. A future reader globbing `.moai/state/integration-lock*` would
// otherwise sweep both lifetimes into one set and re-derive exactly the
// conflation this file exists to keep apart.
//
// NO NEW PRIMITIVE. The platform substrate is the board lock's, taken as it is:
// acquireBoardLockImpl (flock on Unix, atomic-create on Windows) is already
// path-parameterized, and the bounded jittered contention policy is
// board_store.go's. What is added is a second SCOPE over the same substrate,
// not a second mechanism.
//
// Why a dedicated scope rather than borrowing AcquireBoardLock: the defect
// being repaired IS a scope mismatch — a lock whose stated scope did not match
// what it protected. Answering it with a lock scoped to the WHOLE BOARD repeats
// that category error one level up, leaving the next reader to work out why a
// board lock guards a release record.
package kanban

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// integrationMutationLockFileName names the mutation artifact. See the header:
// the stem is deliberately distinct from IntegrationLockFileName.
const integrationMutationLockFileName = "integration-mutation.lock"

// ErrIntegrationLockBusy is returned when the mutation lock stays contended for
// the whole wait budget.
//
// It is DISTINCT from ErrIntegrationLockHeld and must stay so. Held means
// another session owns the integration window — a statement about the board
// that sends a lane to ask a peer to release. Busy means a peer was mid-mutation
// for longer than the budget allows: transient, retry-me, and says nothing about
// who owns the window. Reporting one as the other tells a lane a false thing
// about the board.
var ErrIntegrationLockBusy = errors.New("release integration window record is busy: another process is mutating it")

// IsIntegrationLockBusy reports whether err is the transient-contention
// sentinel. IsIntegrationLockHeld(err) is false for it, by construction: the
// board's own sentinel is wrapped with %v, not %w, so neither this package's
// held predicate nor the board's leaks across the scope boundary.
func IsIntegrationLockBusy(err error) bool { return errors.Is(err, ErrIntegrationLockBusy) }

// integrationMutationLockPath resolves the mutation artifact beside the record,
// under the PRIMARY checkout's shared state directory.
func integrationMutationLockPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".moai", "state", integrationMutationLockFileName)
}

// withIntegrationLockMutation runs fn inside the critical section: at most one
// process at a time is inside the record's read → decide → write sequence for a
// given project root.
//
// fn performs its OWN read after entering (REQ-ILA-003). That placement is the
// point: a caller serialized behind another must decide against the state the
// previous mutation published, never against a read taken before the wait.
//
// The lock is released on every path, including a panic, via defer.
func withIntegrationLockMutation(projectRoot string, fn func() error) error {
	path := integrationMutationLockPath(projectRoot)
	// The record's own state directory — the same one AcquireIntegrationLock
	// creates. The lock cannot be taken before it exists.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("integration lock: %w", err)
	}
	impl, err := acquireIntegrationMutationLock(path)
	if err != nil {
		return err
	}
	defer func() { _ = impl.release() }()
	return fn()
}

// acquireIntegrationMutationLock takes the mutation lock, retrying contention
// within the shared elapsed budget (board_store.go: 10 supported writers ×
// 33ms CI mutation cost × 5 headroom = 1.65s) using the same jittered backoff.
// The budget is inherited rather than re-derived by guess; if integration-window
// mutations turn out to be slower in practice, it is re-derived in a later card.
func acquireIntegrationMutationLock(path string) (boardLockImpl, error) {
	var lastErr error
	deadline := time.Now().Add(boardLockWaitBudget)
	for attempt := 0; ; attempt++ {
		impl, err := acquireBoardLockImpl(path)
		if err == nil {
			return impl, nil
		}
		if !IsBoardLockHeld(err) {
			return nil, fmt.Errorf("integration lock: taking the mutation lock at %s: %w", path, err)
		}
		lastErr = err
		if !time.Now().Before(deadline) {
			// %v, never %w: the board's sentinel must not travel out of this
			// scope, or a caller's IsBoardLockHeld would report true for a
			// lock that has nothing to do with the board.
			return nil, fmt.Errorf("%w (waited %s): %v", ErrIntegrationLockBusy, boardLockWaitBudget, lastErr)
		}
		time.Sleep(boardLockRetryWait(attempt))
	}
}
