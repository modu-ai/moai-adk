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
	"strings"
	"time"
)

// sectionRereadFn is the verdict-read seam (tests inject a racing
// reclaimer's swap between the verdict and the gate).
var sectionRereadFn = os.ReadFile

// sectionRemoveFn is the removal seam.
var sectionRemoveFn = os.Remove

// breakingSuffix names a breaker's mutual-exclusion marker beside the lock
// it is breaking: one O_EXCL file per lock path, claimed through
// ClaimSection itself (owner-labelled, owner-reclaimable). The marker is
// the breaker-vs-breaker critical section: while one process spans
// verdict-to-disposal, no rival breaker can enter, so the path bytes a
// verdict was made on cannot be replaced by another breaker's reclaim.
// Claimers cannot mutate the path either — a Claim needs the path ABSENT,
// and the verified-stale lock is present until the disposal; a release
// removes only its own label, and a dead owner issues none.
const breakingSuffix = ".breaking"

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
// Verify-fresh and disposal are ONE cross-process critical section, and the
// section is an O_EXCL BREAKER MARKER (path+breakingSuffix) claimed through
// ClaimSection itself — the reviewer-gate CAS: create the replacement
// (the marker) with O_EXCL, then dispose of the verified-stale original
// only while that creation holds. RENAME is not a compare-and-swap: a
// rename-based breaker moved whatever sat at the path at rename time, so a
// rival that re-acquired between the verdict and the rename had its fresh
// live lock MOVED — the path freed under a live holder, a third entrant
// acquired, and the post-hoc byte-compare could not close the window. The
// marker closes it structurally: while this breaker holds the marker, no
// rival breaker can enter its own span, claimers cannot mutate the path (a
// Claim needs it absent, and the verified-stale lock is present until the
// disposal), and a release removes only its own label — so the bytes the
// verdict was made on cannot change between the verdict and the unlink.
// The disposal still re-reads immediately before the unlink and aborts on
// any mismatch: belt, never assumption.
//
// A breaker that dies holding the marker leaves an owner-labelled marker
// file; the next breaker's ClaimSection contention path reclaims it through
// the same verified-dead rule (the bare path below), so the marker cannot
// wedge the break.
func BreakStaleLock(path string) bool {
	if strings.HasSuffix(path, breakingSuffix) {
		// Reclaiming an orphaned breaker marker: no nested marker — a
		// marker's break needs no breaker-vs-breaker serialization beyond
		// the rename-free verify-then-dispose, and nesting one would
		// recurse without bound.
		return breakStaleLockBare(path)
	}
	release, err := ClaimSection(path+breakingSuffix, 0o600, 2, 2*time.Millisecond)
	if err != nil {
		return false // a live breaker owns the break; the caller's retry loop re-runs
	}
	defer func() { _ = release() }()
	return breakStaleLockBare(path)
}

// breakStaleLockBare is the verify-then-dispose body without a marker —
// the marker holder's critical body, and the whole break for a marker
// file's own reclaim.
func breakStaleLockBare(path string) bool {
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
	// The disposal gate: the bytes at the path must STILL be the bytes the
	// verdict was made on, immediately before the unlink.
	now, err := os.ReadFile(path)
	if err != nil || string(now) != string(raw) {
		return false // someone replaced the lock between verdict and disposal
	}
	if err := sectionRemoveFn(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return false
	}
	slog.Warn("lock section: broke a stale lock (verified-dead owner)",
		"lock", path, "owner_pid", owner.PID, "owner_boot", owner.BootID)
	return true
}
