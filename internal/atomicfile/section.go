package atomicfile

// section.go — the owner-verified cross-process claim section
// (SPEC-FEEDBACK-PARTICIPATION-001 D37/D40). One implementation, two
// consumers: the feedback queue's Mutate lock and the bugreport capture
// spool's section lock. The primitive is Claim (exclusive create) dressed
// with what a bare O_EXCL lock lacks:
//
//   - an owner label (pid + boot identity) written at acquire time, so a
//     later caller can tell WHOSE lock it sees;
//   - an owner-verified stale-lock break on the contention path: the break
//     fires ONLY when the recorded owner is verifiably dead — a recorded
//     boot different from the current one, or a pid that no longer names a
//     live process. A live owner always blocks, however long its section
//     runs: there is NO age-based break and none may be added, because the
//     retry budget governs acquisition, not the hold — an age-only break
//     could discard a live slow owner's committed mutation, the lost-update
//     defect this repair closes. The invariant is absolute: verified owner
//     death, nothing else.
//   - a release that removes the lock only when the label at the path is
//     STILL this process's identity — a stale-armed reclaimer's remove
//     landing on this lock, another writer re-claiming, must not let this
//     release delete the OTHER writer's lock.
//
// An unlabelled lock (a crash between Claim and the label write) reads as
// LIVE: it wedges until the boot changes and is never broken on a
// possibly-live owner — the conservative direction.

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"
)

// sectionRereadFn is the verdict-read seam (tests inject a racing
// reclaimer's swap between the verdict and the gate).
var sectionRereadFn = os.ReadFile

// sectionRenameFn is the break's critical-section seam: the atomic rename
// that detaches the verified-dead lock from its shared path.
var sectionRenameFn = os.Rename

// sectionRemoveFn is the removal seam.
var sectionRemoveFn = os.Remove

// ClaimSection takes the advisory lock at path, returning its release func.
// Contention retries within the given budget, breaking the lock only on a
// verified-dead owner, then errors naming the path.
func ClaimSection(path string, perm os.FileMode, retries int, delay time.Duration) (func() error, error) {
	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		err := Claim(path, perm)
		if err == nil {
			if werr := writeOwnerLabel(path, perm); werr != nil {
				// The lock is HELD but unlabelled: release immediately and
				// report — an unlabelled lock could never be verified, and
				// wedging on write failure beats breaking the invariant.
				_ = os.Remove(path)
				return nil, fmt.Errorf("claim section %s: labelling lock: %w", path, werr)
			}
			return releaseSectionFunc(path), nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("claim section %s: lock: %w", path, err)
		}
		// Contention: check whether the holder is a verified-dead owner. A
		// live owner blocks through the budget; a verified-dead one is
		// broken and the claim retried immediately.
		if BreakStaleLock(path) {
			lastErr = err
			continue
		}
		lastErr = err
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("claim section %s: lock held: %w", path, lastErr)
}

// releaseSectionFunc returns the release func for a section just acquired
// at path. The release removes the lock only when the label at the path is
// STILL THIS PROCESS'S identity (pid + boot): the identity fields are what
// ownership means. CreatedAt is deliberately not compared.
func releaseSectionFunc(path string) func() error {
	lockPath := path
	return func() error {
		raw, rerr := os.ReadFile(lockPath)
		if rerr != nil {
			return nil // gone: nothing to remove
		}
		var atPath LockOwner
		if err := json.Unmarshal(raw, &atPath); err != nil ||
			atPath.PID != os.Getpid() || atPath.BootID != BootIDIdentity() {
			return nil // not ours: remove nothing
		}
		if rmErr := os.Remove(lockPath); rmErr != nil && !errors.Is(rmErr, os.ErrNotExist) {
			// A surviving artifact blocks every later writer, so the
			// failure is reported rather than swallowed.
			return fmt.Errorf("claim section %s: lock release failed: %w", lockPath, rmErr)
		}
		return nil
	}
}

// writeOwnerLabel labels a just-acquired lock with this process's identity.
func writeOwnerLabel(path string, perm os.FileMode) error {
	owner := LockOwner{
		PID:       os.Getpid(),
		BootID:    BootIDIdentity(),
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	raw, err := json.Marshal(owner)
	if err != nil {
		return err
	}
	// The file exists (Claim created it); write the label in place. The
	// window between Claim and this write is the crash window the break
	// cannot classify — an unlabelled lock reads as LIVE (conservative: it
	// wedges until the boot changes, never breaks a possibly-live owner).
	return os.WriteFile(path, raw, perm)
}

// BreakStaleLock removes the lock file at path when its recorded owner is
// verifiably dead, reporting whether a break happened.
//
// Verify-fresh and break are ONE cross-process critical section, and its
// boundary is an atomic RENAME: the breaker renames the dead lock to a
// unique break-name, verifies the MOVED bytes are still the ones its
// verdict was made on, and disposes of the detached file. The rename is the
// serialization: only one renamer succeeds (every other contender's rename
// fails because the file is gone), so no rival can interleave a reclaim
// between this breaker's verdict and its disposal, and the disposal
// targets the breaker's own detached copy — never whatever a rival may
// have acquired at the now-free path. The predecessor scheme (verify, then
// remove by path) left exactly that window: a rival's reclaim landing in it
// had its fresh, live lock deleted by the late remove, and the next acquirer
// entered the section beside the rival's still-running writer.
//
// A mismatch after the rename — the moved bytes differing from the verdict
// bytes — is impossible by construction under this scheme (changing the
// file's bytes at the path requires a prior successful rename, which would
// have made this breaker's rename fail), but it is handled rather than
// assumed: the moved lock is restored to the path best-effort and the break
// reports no break, never deleting bytes it cannot attribute.
//
// A breaker that dies between the rename and the remove leaves a tiny
// detached `*.break-*` debris file: it is unlabelled bytes nobody claims,
// harmless to every later acquirer, and bounded by the microsecond window
// it takes to crash in.
func BreakStaleLock(path string) bool {
	raw, err := sectionRereadFn(path)
	if err != nil {
		return false // unreadable: cannot verify death, never break
	}
	var owner LockOwner
	if err := json.Unmarshal(raw, &owner); err != nil {
		return false // unlabelled (a crash between Claim and label): live
	}
	if !OwnerIsDead(owner) {
		return false
	}

	// The critical-section boundary: atomically detach the verified-dead
	// lock from its shared path. Losing the rename means another breaker
	// (or a self-label release) already moved or removed the lock — this
	// breaker re-loops and finds whoever acquired at the path now.
	breakPath := path + ".break-" + fmt.Sprintf("%d-%d", os.Getpid(), time.Now().UnixNano())
	if err := sectionRenameFn(path, breakPath); err != nil {
		return false
	}

	// Verify INSIDE the section: the bytes this breaker moved must still be
	// the bytes its verdict was made on.
	moved, err := os.ReadFile(breakPath)
	if err != nil || string(moved) != string(raw) {
		// Not attributable to this verdict: restore best-effort and report
		// no break — deleting would dispose of a lock this breaker never
		// verified.
		_ = sectionRenameFn(breakPath, path)
		return false
	}

	if err := sectionRemoveFn(breakPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		// The dead lock is detached either way; a surviving break-file is
		// debris, not a wedge. Report the break: the path is free.
		slog.Warn("lock section: broke a stale lock, break-file removal failed",
			"lock", path, "break_file", breakPath, "owner_pid", owner.PID)
		return true
	}
	slog.Warn("lock section: broke a stale lock (verified-dead owner)",
		"lock", path, "owner_pid", owner.PID, "owner_boot", owner.BootID)
	return true
}
