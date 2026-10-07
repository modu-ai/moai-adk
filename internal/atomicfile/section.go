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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

// sectionOwnerReadMaxBytes caps the verdict read: an owner record is a
// short JSON object; anything longer is not one.
const sectionOwnerReadMaxBytes = 4096

// sectionRereadFn is the verdict-read seam (tests inject a racing
// reclaimer's swap between the verdict and the gate). The production read
// is BOUNDED (review gate finding, P2): a non-regular lock path is refused
// WITHOUT opening it — a FIFO swapped in at the lock path parked the
// reread past every deadline, beyond the caller's context — and the read
// costs one small capped allocation, never the file's size.
var sectionRereadFn = func(path string) ([]byte, error) {
	if info, serr := os.Stat(path); serr == nil && !info.Mode().IsRegular() {
		return nil, fmt.Errorf("section owner %s: not a regular file", path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return io.ReadAll(io.LimitReader(f, sectionOwnerReadMaxBytes))
}

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

// reclaimSuffix names a RECLAIMER's mutual-exclusion marker beside the
// breaker marker it is reclaiming (review-gate finding: the marker's own
// reclaim was a non-atomic read→delete — one reclaimer's late delete could
// remove a rival's LIVE re-acquired marker). The same CAS shape the queue
// break received applies one level down: create the replacement (the
// .reclaim marker) with O_EXCL first, and delete the breaker marker only
// on creation success. While this disposal holds .reclaim, no rival
// disposal can run, and a re-creation of the breaker marker (a rival
// claiming it as ITS live section) can only happen after this release —
// so a late delete can never land on a live marker.
const reclaimSuffix = ".reclaim"

// maxReclaimDepth bounds how deep a chain of nested dead guards is
// followed. Each level of a chain is one historical process death (a
// reclaimer that died holding its guard), so a chain deeper than a couple
// of levels is a pathological accumulation, not a working state. The cap
// is deliberately small for a second reason: each level's contention retry
// loop re-walks the chain below it (the recursion re-enters through
// ClaimSection's own attempts), so the walk cost grows exponentially with
// the cap — at 3 the worst case is a bounded handful of chain walks. Past
// the cap the reclaim REFUSES — no delete ever runs without its guard
// claim — and the caller's budget backs off; a wedge beats a race, and the
// store then needs an operator's cleanup. The depth is read from the path
// itself (the number of reclaimSuffix occurrences), because the recursion
// re-enters through ClaimSection's contention path.
const maxReclaimDepth = 3

// ClaimSection takes the advisory lock at path, returning its release
// func. Contention retries within the given budget, breaking the lock only
// on a verified-dead owner, then errors naming the path. The context is
// honored THROUGHOUT the contention loop: a caller's deadline or
// cancellation ends the wait at the next retry boundary instead of burning
// the whole budget — a lock wait that outlives its caller's context is
// precisely the stall the caller was trying to bound.
func ClaimSection(ctx context.Context, path string, perm os.FileMode, retries int, delay time.Duration) (func() error, error) {
	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("claim section %s: %w", path, err)
		}
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
		// broken and the claim retried immediately. The caller's context
		// reaches the recursive reclaim, so a cancelled claim does not
		// re-walk a deep guard chain.
		if BreakStaleLockContext(ctx, path) {
			lastErr = err
			continue
		}
		lastErr = err
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("claim section %s: %w", path, ctx.Err())
		case <-time.After(delay):
		}
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
// the same verified-dead rule (the .reclaim-guarded path below), so the
// marker cannot wedge the break.
// BreakStaleLock is BreakStaleLockContext under the caller's background
// context.
func BreakStaleLock(path string) bool {
	return BreakStaleLockContext(context.Background(), path)
}

// BreakStaleLockContext is BreakStaleLock under a caller's context: the
// cancellation reaches EVERY level of the recursive guard reclaim — each
// level's guard ClaimSection selects on it — so a cancelled caller's walk
// stops at the next claim boundary instead of re-walking the chain through
// the retry loops (each level's loop re-walks the chain below it, which is
// what makes an uncancelled deep walk expensive).
func BreakStaleLockContext(ctx context.Context, path string) bool {
	if strings.HasSuffix(path, reclaimSuffix) || strings.HasSuffix(path, breakingSuffix) {
		// Reclaiming a marker — a breaker's (.breaking) or a reclaimer's
		// (.reclaim): hold ITS OWN guard marker first — delete only on
		// creation success, at EVERY level (review-gate residual: the
		// .reclaim path's verify-and-delete was bare, so a reclaimer
		// pausing between its check and its delete could remove a rival's
		// LIVE re-acquired guard). The recursion re-enters through
		// ClaimSection's contention path when the guard is itself a dead
		// marker; the chain depth read from the path bounds it — past
		// maxReclaimDepth the reclaim refuses and nothing is deleted.
		if strings.Count(path, reclaimSuffix) >= maxReclaimDepth {
			slog.Warn("lock section: reclaim chain too deep; refusing to break",
				"lock", path, "depth", strings.Count(path, reclaimSuffix))
			return false
		}
		release, err := ClaimSection(ctx, path+reclaimSuffix, 0o600, 2, 2*time.Millisecond)
		if err != nil {
			return false // a live reclaimer owns the disposal, or the caller's context is done
		}
		defer func() { _ = release() }()
		return breakStaleLockBare(path)
	}
	release, err := ClaimSection(ctx, path+breakingSuffix, 0o600, 2, 2*time.Millisecond)
	if err != nil {
		return false // a live breaker owns the break, or the caller's context is done
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
