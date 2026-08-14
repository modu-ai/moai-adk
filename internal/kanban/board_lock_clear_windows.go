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
)

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
		return nil, fmt.Errorf("clear stale board lock: artifact identity unparseable — cannot positively observe an unknown owner absent: %w", err)
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
