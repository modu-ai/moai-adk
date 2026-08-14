//go:build windows

// board_lock_clear_windows.go — the Windows stale-lock clear
// (SPEC-KANBAN-BOARD-001 REQ-KB-023). The Windows substrate is atomic-create:
// the artifact IS the lock, so a killed holder leaves a blocker and the
// clear is the bounded exit. The pre-removal re-read mitigates the
// check-then-unlink race (AP-29 residual stated in the doc comment below,
// not closed at this layer).
package kanban

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// interruptedAcquisitionGrace is how long the clear waits, when it finds an
// EMPTY lock artifact, before the pre-removal re-read — long enough to
// cover the create-to-write interval of a live acquirer mid-publication, so
// emptiness observed twice across the grace is evidence the writer is gone
// rather than merely slow (sync-audit F3).
const interruptedAcquisitionGrace = 100 * time.Millisecond

// isEmptyLockArtifact reports whether the artifact carries no bytes at all
// (whitespace aside): the shape a process killed between the O_EXCL create
// and the owner-write leaves behind — an identity that was never published.
func isEmptyLockArtifact(raw []byte) bool {
	return len(strings.TrimSpace(string(raw))) == 0
}

// processAlive is the liveness probe the Windows clear conditions on. It is
// a package-level indirection so the release-and-reacquire interleaving can
// be constructed in tests AT THE PROBE — the step the clear runs immediately
// before its pre-removal re-read (AC-KB-023 observation 3). It lives behind
// the windows tag with its only consumer: the unix clear is gated out (the
// kernel drops flock on exit), so a unix-side probe would be dead code
// (re-review ITEM 4).
var processAlive = defaultProcessAlive

// parseLockOwner decodes a lock artifact's recorded owner identity.
func parseLockOwner(raw []byte) (*BoardLockOwner, error) {
	var owner BoardLockOwner
	if err := json.Unmarshal(raw, &owner); err != nil {
		return nil, fmt.Errorf("parsing lock owner identity: %w", err)
	}
	if owner.PID <= 0 {
		return nil, fmt.Errorf("parsing lock owner identity: pid %d is not a process identity", owner.PID)
	}
	return &owner, nil
}

// ClearStaleBoardLock removes a stale board-wide lock artifact — an explicit,
// operator-visible, bounded act; never a step the acquire path takes on its
// own, and never conditioned on the artifact's age.
//
// The removal happens ONLY when the recorded process is positively observed
// absent, and a RE-READ of the recorded identity runs immediately before the
// unlink, aborting on any mismatch: between the inspection and the removal
// the lock can be legitimately released by its owner and re-acquired by a
// live process, and unlinking then would admit two writers to the critical
// section — the clearing operation causing precisely the concurrency the lock
// exists to prevent.
//
// TOCTOU RESIDUAL, stated rather than engineered away (AP-29): the re-read
// NARROWS the inspection-to-removal window, it does not close it. The removal
// primitive takes a PATH, resolves it at call time, and no portable
// handle-based unlink exists at this layer — a lock re-acquired in the
// residual interval between the re-read and the unlink can still be removed.
// The window is entered only by a clear already running against an artifact
// whose recorded owner was observed dead, and the operation is explicit and
// reports what it removed, so a rare bad outcome is attributable rather than
// silent.
func ClearStaleBoardLock(root string) (*ClearStaleReport, error) {
	path := boardLockPath(root)

	rawFirst, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ClearStaleReport{Removed: false, Reason: "no lock artifact present"}, nil
		}
		return nil, fmt.Errorf("clear stale board lock: reading artifact: %w", err)
	}
	ownerFirst, err := parseLockOwner(rawFirst)
	if err != nil {
		// A PUBLISHED-but-broken identity cannot be proven absent — refuse.
		// An EMPTY artifact is different: it is the interrupted-acquisition
		// shape (the creating process died between the O_EXCL create and the
		// owner-write), no identity was ever published, and there is no pid
		// to probe. AC-KB-023 observation 1 governs the shape — an artifact
		// left behind by a terminated process is removed — and the
		// pre-removal re-read provides the observation: after a grace
		// covering the create-to-write interval, emptiness observed AGAIN
		// means no live acquirer is publishing; content appearing in the
		// meantime means one is, and the clear aborts as changed hands
		// (sync-audit F3).
		if !isEmptyLockArtifact(rawFirst) {
			return nil, fmt.Errorf("clear stale board lock: artifact identity unparseable — cannot positively observe an unknown owner absent: %w", err)
		}
		time.Sleep(interruptedAcquisitionGrace)
		rawSecond, rerr := os.ReadFile(path)
		if rerr != nil {
			if os.IsNotExist(rerr) {
				return &ClearStaleReport{Removed: false, Reason: "artifact disappeared before removal"}, nil
			}
			return nil, fmt.Errorf("clear stale board lock: re-reading empty artifact: %w", rerr)
		}
		if !isEmptyLockArtifact(rawSecond) {
			return nil, fmt.Errorf("%w: an identity appeared in the empty artifact during the grace — a live acquirer is publishing", ErrBoardLockChangedHands)
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("clear stale board lock: removing interrupted-acquisition artifact: %w", err)
		}
		return &ClearStaleReport{
			Removed: true,
			Reason:  "removed empty lock artifact of an interrupted acquisition (no identity was ever published)",
		}, nil
	}

	if processAlive(ownerFirst.PID) {
		return &ClearStaleReport{
			Removed: false,
			PID:     ownerFirst.PID,
			Reason:  fmt.Sprintf("recorded owner pid %d is live; nothing removed", ownerFirst.PID),
		}, nil
	}

	// Re-read immediately before the removal; abort on any change of hands.
	rawSecond, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ClearStaleReport{
				Removed: false,
				PID:     ownerFirst.PID,
				Reason:  "artifact disappeared before removal",
			}, nil
		}
		return nil, fmt.Errorf("clear stale board lock: re-reading artifact: %w", err)
	}
	ownerSecond, err := parseLockOwner(rawSecond)
	if err != nil || *ownerSecond != *ownerFirst {
		return nil, fmt.Errorf("%w: recorded identity changed between inspection and removal", ErrBoardLockChangedHands)
	}

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return &ClearStaleReport{
				Removed: false,
				PID:     ownerFirst.PID,
				Reason:  "artifact disappeared before removal",
			}, nil
		}
		return nil, fmt.Errorf("clear stale board lock: removing artifact: %w", err)
	}
	return &ClearStaleReport{
		Removed: true,
		PID:     ownerFirst.PID,
		Reason:  fmt.Sprintf("removed stale lock artifact of terminated pid %d", ownerFirst.PID),
	}, nil
}
