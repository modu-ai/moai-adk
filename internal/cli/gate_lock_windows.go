//go:build windows

package cli

// gate_lock_windows.go — Windows substrate of the gate-run lock:
// atomic-create (O_CREATE|O_EXCL), mirroring internal/kanban
// board_lock_windows.go / board_lock_clear_windows.go. Windows lacks
// fcntl-style advisory flock, so the artifact IS the lock: a holder killed
// mid-run leaves an artifact that blocks every subsequent gate run until it
// is cleared — which is why the owner identity is recorded in the artifact
// and the clear below conditions on the holder being positively observed
// absent.

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

const (
	// Transient delete-pending / sharing-violation budget, mirroring
	// board_lock_windows.go's classification: genuine contention is reported
	// as held immediately and NEVER retried; only short-lived Windows
	// conditions (a just-released holder's delete-pending name, an
	// indexer/scanner sharing conflict) consume this budget.
	gateLockTransientRetries = 10
	gateLockTransientDelay   = 5 * time.Millisecond
)

// gateLockProcessAlive is the liveness probe the clear conditions on. It is a
// package-level indirection so the release-and-reacquire interleaving can be
// constructed in tests AT THE PROBE — the step the clear runs immediately
// before its pre-removal re-read.
var gateLockProcessAlive = kanban.FactoryProcessAlive

// interruptedAcquisitionGrace is how long the clear waits, when it finds an
// EMPTY lock artifact, before the pre-removal re-read — long enough to cover
// the create-to-write interval of a live acquirer mid-publication, so
// emptiness observed twice across the grace is evidence the writer is gone
// rather than merely slow.
const interruptedAcquisitionGrace = 100 * time.Millisecond

// isEmptyLockArtifact reports whether the artifact carries no bytes at all
// (whitespace aside): the shape a process killed between the O_EXCL create
// and the owner-write leaves behind — an identity that was never published.
func isEmptyLockArtifact(raw []byte) bool {
	return len(strings.TrimSpace(string(raw))) == 0
}

// atomicFileGateLock holds an exclusively-created lock file; releasing means
// closing and removing it.
type atomicFileGateLock struct {
	lockPath string
	file     *os.File
}

func (f *atomicFileGateLock) release() error {
	if f == nil {
		return nil
	}
	if f.file != nil {
		_ = f.file.Close()
		f.file = nil
	}
	if f.lockPath != "" {
		err := os.Remove(f.lockPath)
		f.lockPath = ""
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// acquireGateLockImpl creates lockPath with O_CREATE|O_EXCL — atomic on NTFS
// — and records this process's identity IN the artifact.
func acquireGateLockImpl(lockPath string) (gateLockImpl, error) {
	transientLeft := gateLockTransientRetries
	for {
		file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o644)
		if err == nil {
			if _, werr := file.Write(newGateLockOwnerRecord()); werr != nil {
				_ = file.Close()
				_ = os.Remove(lockPath)
				return nil, fmt.Errorf("record gate-run lock owner %s: %w", lockPath, werr)
			}
			return &atomicFileGateLock{lockPath: lockPath, file: file}, nil
		}
		switch {
		case os.IsExist(err):
			return nil, ErrGateLockHeld
		case errors.Is(err, os.ErrPermission):
			if transientLeft--; transientLeft <= 0 {
				return nil, fmt.Errorf("open gate-run lock %s: %w", lockPath, err)
			}
			time.Sleep(gateLockTransientDelay)
		default:
			return nil, fmt.Errorf("open gate-run lock %s: %w", lockPath, err)
		}
	}
}

// ClearStaleGateLock removes a stale gate-run lock artifact — an explicit,
// bounded act on the Windows substrate, where a killed holder's artifact
// blocks every subsequent run until it is cleared.
//
// The removal happens ONLY when the recorded process is positively observed
// absent, and a RE-READ of the recorded identity runs immediately before the
// unlink, aborting on any mismatch: between the inspection and the removal
// the lock can be legitimately released by its owner and re-acquired by a
// live process, and unlinking then would admit two serialized-section
// breakers at once — the clearing operation causing precisely the concurrency
// the lock exists to prevent.
//
// TOCTOU RESIDUAL, stated rather than engineered away: the re-read NARROWS
// the inspection-to-removal window, it does not close it. The removal
// primitive takes a PATH, resolves it at call time, and no portable
// handle-based unlink exists at this layer — a lock re-acquired in the
// residual interval between the re-read and the unlink can still be removed.
// The window is entered only by a clear already running against an artifact
// whose recorded owner was observed dead, and the operation is explicit and
// reports what it removed, so a rare bad outcome is attributable rather than
// silent.
func ClearStaleGateLock(projectDir string) (*GateLockClearReport, error) {
	path := gateLockPath(projectDir)

	rawFirst, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &GateLockClearReport{Removed: false, Reason: "no lock artifact present"}, nil
		}
		return nil, fmt.Errorf("clear stale gate-run lock: reading artifact: %w", err)
	}
	ownerFirst, err := parseGateLockOwner(rawFirst)
	if err != nil {
		// A PUBLISHED-but-broken identity cannot be proven absent — refuse.
		// An EMPTY artifact is different: it is the interrupted-acquisition
		// shape (the creating process died between the O_EXCL create and the
		// owner-write), no identity was ever published, and there is no pid
		// to probe. After a grace covering the create-to-write interval,
		// emptiness observed AGAIN means no live acquirer is publishing;
		// content appearing in the meantime means one is, and the clear
		// aborts as changed hands.
		if !isEmptyLockArtifact(rawFirst) {
			return nil, fmt.Errorf("clear stale gate-run lock: artifact identity unparseable — cannot positively observe an unknown owner absent: %w", err)
		}
		time.Sleep(interruptedAcquisitionGrace)
		rawSecond, rerr := os.ReadFile(path)
		if rerr != nil {
			if os.IsNotExist(rerr) {
				return &GateLockClearReport{Removed: false, Reason: "artifact disappeared before removal"}, nil
			}
			return nil, fmt.Errorf("clear stale gate-run lock: re-reading empty artifact: %w", rerr)
		}
		if !isEmptyLockArtifact(rawSecond) {
			return nil, fmt.Errorf("%w: an identity appeared in the empty artifact during the grace — a live acquirer is publishing", ErrGateLockChangedHands)
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("clear stale gate-run lock: removing interrupted-acquisition artifact: %w", err)
		}
		return &GateLockClearReport{
			Removed: true,
			Reason:  "removed empty lock artifact of an interrupted acquisition (no identity was ever published)",
		}, nil
	}

	if gateLockProcessAlive(ownerFirst.PID) {
		return &GateLockClearReport{
			Removed: false,
			PID:     ownerFirst.PID,
			Reason:  fmt.Sprintf("recorded owner pid %d is live; nothing removed", ownerFirst.PID),
		}, nil
	}

	// Re-read immediately before the removal; abort on any change of hands.
	rawSecond, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &GateLockClearReport{
				Removed: false,
				PID:     ownerFirst.PID,
				Reason:  "artifact disappeared before removal",
			}, nil
		}
		return nil, fmt.Errorf("clear stale gate-run lock: re-reading artifact: %w", err)
	}
	ownerSecond, err := parseGateLockOwner(rawSecond)
	if err != nil || ownerSecond != ownerFirst {
		return nil, fmt.Errorf("%w: recorded identity changed between inspection and removal", ErrGateLockChangedHands)
	}

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return &GateLockClearReport{
				Removed: false,
				PID:     ownerFirst.PID,
				Reason:  "artifact disappeared before removal",
			}, nil
		}
		return nil, fmt.Errorf("clear stale gate-run lock: removing artifact: %w", err)
	}
	return &GateLockClearReport{
		Removed: true,
		PID:     ownerFirst.PID,
		Reason:  fmt.Sprintf("removed stale lock artifact of terminated pid %d", ownerFirst.PID),
	}, nil
}
