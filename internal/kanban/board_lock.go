// board_lock.go — the board-wide advisory lock and its bounded stale clear
// (SPEC-KANBAN-BOARD-001 REQ-KB-019/023, M1).
//
// The lock spans the ENTIRE read-modify-write of the WHOLE board, not a card:
// with WIP 2, two concurrent transitions of two different cards each holding
// only their own card's lock would each observe the bound satisfied and each
// write, landing at WIP 3 — the bound is only sound beneath board-wide
// exclusion. The substrate reuses the repository's existing cross-process
// per-scope lock PATTERN (internal/spec/lock.go and its platform
// counterparts: flock on Unix, atomic-create on Windows); internal/lockfile's
// in-process mutex is neither used nor upgraded.
package kanban

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ErrBoardLockHeld is returned by AcquireBoardLock when another process holds
// the board-wide lock.
var ErrBoardLockHeld = errors.New("kanban board lock held")

// IsBoardLockHeld reports whether err is the contention sentinel.
func IsBoardLockHeld(err error) bool {
	return errors.Is(err, ErrBoardLockHeld)
}

// ErrBoardLockChangedHands is returned by ClearStaleBoardLock when the
// pre-removal re-read observes a different recorded identity than the
// inspection did — the artifact was released and re-acquired inside the
// window, and the clear aborts rather than unlinking a valid lock.
var ErrBoardLockChangedHands = errors.New("kanban board lock changed hands between inspection and removal")

// IsBoardLockChangedHands reports whether err is the changed-hands abort.
func IsBoardLockChangedHands(err error) bool {
	return errors.Is(err, ErrBoardLockChangedHands)
}

// boardLockFileName names the lock artifact inside the board directory.
const boardLockFileName = "board.lock"

// BoardLockOwner is the creating process's identity recorded IN the lock
// artifact (REQ-KB-023). The identity is what makes a stale artifact
// distinguishable from a live holder's: without it, "the holder is gone" is a
// guess, and clearing on a guess unlinks a lock a live process may hold.
type BoardLockOwner struct {
	PID       int    `json:"pid"`
	CreatedAt string `json:"created_at"`
}

// BoardLock represents an acquired board-wide lock. Callers MUST call
// Release when the read-modify-write completes.
type BoardLock struct {
	path string
	impl boardLockImpl
}

// boardLockImpl is the platform-specific lock implementation (flock on Unix,
// atomic-create on Windows — mirroring internal/spec's substrate split).
type boardLockImpl interface {
	release() error
}

// processAlive is the liveness probe the clear conditions on. It is a
// package-level indirection so the release-and-reacquire interleaving can be
// constructed in tests AT THE PROBE — the step the clear runs immediately
// before its pre-removal re-read (AC-KB-023 observation 3).
var processAlive = defaultProcessAlive

// AcquireBoardLock acquires the board-wide lock at
// <root>/.moai/state/kanban-board/board.lock, creating the board directory if
// absent. Returns ErrBoardLockHeld on contention; the caller retries or
// reports, never blocks.
//
// The acquiring process records its identity in the artifact as part of the
// acquisition, so an artifact always names its current owner.
func AcquireBoardLock(root string) (*BoardLock, error) {
	dir := BoardDir(root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("acquire board lock: creating board dir: %w", err)
	}
	path := boardLockPath(root)
	impl, err := acquireBoardLockImpl(path)
	if err != nil {
		return nil, err
	}
	return &BoardLock{path: path, impl: impl}, nil
}

// Release releases the board-wide lock. Safe to call multiple times.
func (l *BoardLock) Release() error {
	if l == nil || l.impl == nil {
		return nil
	}
	err := l.impl.release()
	l.impl = nil
	return err
}

// Path returns the lock artifact's path (diagnostics).
func (l *BoardLock) Path() string {
	if l == nil {
		return ""
	}
	return l.path
}

// boardLockPath returns the lock artifact's path beneath root.
func boardLockPath(root string) string {
	return filepath.Join(BoardDir(root), boardLockFileName)
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

// newLockOwnerRecord builds the owner identity block written at acquisition.
func newLockOwnerRecord() []byte {
	owner := BoardLockOwner{
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
