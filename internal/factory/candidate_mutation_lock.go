// candidate_mutation_lock.go — the per-card candidate mutation lock (card
// t1478, SPEC-CANDIDATE-CI-001 M2 repair): serializes ONE card's
// push→record-write section across processes.
//
// The defect it closes: two candidate invocations of the same pinned SHA
// can interleave as A.push → B.push+record → A.record, leaving the remote
// naming B's candidate commit while the record names A's — the verdict the
// landing check later consumes then describes a commit the branch does not
// carry. The section is therefore atomic per card: push and record move
// together or not at all.
//
// NO NEW PRIMITIVE, and NEVER the integration window. The substrate is the
// shared state lock's, taken exactly as integration_lock_mutation.go takes
// it: acquireStateLockImpl (flock on Unix, atomic-create on Windows) over a
// per-card path, with the bounded jittered contention policy. The candidate
// path is deliberately DISTINCT from the integration-mutation artifact — a
// reader globbing one lifetime must never sweep the other, and a candidate
// push must never queue behind a window mutation (the candidate verb is
// window-free by design, design.md D4).
//
// The artifact's filename stem (`candidate-mutation-`) is likewise distinct
// from the record store's directory (`candidate/`), for the same reason the
// integration file cites: one stem per lifetime.
package factory

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// candidateMutationLockFilePrefix names the mutation artifact, one per
// card: candidate-mutation-<card>.lock beside the record store.
const candidateMutationLockFilePrefix = "candidate-mutation-"

// ErrCandidateMutationBusy is returned when the per-card mutation lock
// stays contended for the whole wait budget — another candidate
// invocation of the SAME card is mid-push. Transient, retry-me; it says
// nothing about the integration window, and is DISTINCT from every
// integration-lock sentinel for the same reason ErrIntegrationLockBusy is
// distinct from ErrIntegrationLockHeld.
var ErrCandidateMutationBusy = errors.New("candidate record is busy: another process is pushing this card's candidate")

// candidateMutationLockPath resolves the per-card artifact under the
// project's shared state directory. cardID is a validated path segment
// (the same guard the record path uses), so the filename cannot escape
// the directory.
func candidateMutationLockPath(projectRoot, cardID string) (string, error) {
	if err := validCandidateKeyPart(cardID); err != nil {
		return "", fmt.Errorf("candidate lock: card id: %w", err)
	}
	return filepath.Join(projectRoot, ".moai", "state", candidateMutationLockFilePrefix+cardID+".lock"), nil
}

// WithCandidateMutation runs fn inside the per-card critical section: at
// most one process at a time is inside one card's push → read → write
// sequence for a given project root. fn performs its OWN reads after
// entering — a caller serialized behind another must decide against the
// state the previous section published, never against a read taken before
// the wait.
//
// The lock is released on every path, including a panic, via defer. A push
// inside the section can legitimately take longer than the wait budget on
// a slow remote; a contender that expires receives ErrCandidateMutationBusy
// and leaves every byte untouched — the honest transient failure, never a
// divergent record.
func WithCandidateMutation(projectRoot, cardID string, fn func() error) error {
	path, err := candidateMutationLockPath(projectRoot, cardID)
	if err != nil {
		return err
	}
	// The state directory exists before the lock is taken — the same
	// ordering the integration mutation lock uses.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("candidate lock: %w", err)
	}
	impl, err := acquireCandidateMutationLock(path)
	if err != nil {
		return err
	}
	defer func() { _ = impl.release() }()
	return fn()
}

// acquireCandidateMutationLock takes the per-card lock, retrying contention
// within the shared elapsed budget (state_lock_wait.go) using the same
// jittered backoff the integration mutation lock inherits.
func acquireCandidateMutationLock(path string) (stateLockImpl, error) {
	var lastErr error
	deadline := time.Now().Add(stateLockWaitBudget)
	for attempt := 0; ; attempt++ {
		impl, err := acquireStateLockImpl(path)
		if err == nil {
			return impl, nil
		}
		if !IsStateLockHeld(err) {
			return nil, fmt.Errorf("candidate lock: taking the mutation lock at %s: %w", path, err)
		}
		lastErr = err
		if !time.Now().Before(deadline) {
			// Budget exhausted. The Windows wedge-recovery mirrors the
			// integration mutation lock's: the artifact IS the lock on
			// Windows, so a holder killed inside the section wedges every
			// later push of this card; the recovery clears the artifact only
			// when its recorded owner is positively observed absent. One
			// retry, never a loop — uncertainty resolves toward busy.
			if report, clearErr := clearWedgedCandidateMutationLock(path); clearErr == nil && report != nil && report.Removed {
				if impl, retryErr := acquireStateLockImpl(path); retryErr == nil {
					return impl, nil
				}
			}
			return nil, fmt.Errorf("%w (waited %s at %s): %v", ErrCandidateMutationBusy, stateLockWaitBudget, path, lastErr)
		}
		time.Sleep(stateLockRetryWait(attempt))
	}
}
