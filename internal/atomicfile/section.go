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
	// One byte PAST the cap (review-gate finding on card t1606, round 5):
	// a silent truncation would let the disposal gate compare a valid
	// owner record's first 4096 bytes while garbage past the cap slipped
	// the byte-compare — an oversized file is not a valid owner record
	// anywhere, so refusing it is the conservative direction for every
	// reader (verdict, release, pre-check).
	raw, err := io.ReadAll(io.LimitReader(f, sectionOwnerReadMaxBytes+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > sectionOwnerReadMaxBytes {
		return nil, fmt.Errorf("section owner %s: %d bytes, over the %d-byte owner-record cap", path, len(raw), sectionOwnerReadMaxBytes)
	}
	return raw, nil
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

// maxReclaimGuardAttempts bounds the guard claim's attempts (card t1606):
// one initial O_EXCL and one more — after a blocked rival guard was
// disposed, or after a live-rival backoff, whichever consumed the first.
// A successful disposal claims immediately outside this budget. The
// former design claimed the guard through ClaimSection, whose contention
// path spawned the next .reclaim level for EVERY blocked rival — live ones
// included — so contention self-propagated the chain and the depth-3 cap
// refused it permanently: a single dead reclaimer wedged the lock for the
// life of the boot. The non-recursive guard claim kills the
// self-propagation: a live rival owns its disposal (at most ONE transient
// guard level spawns, and a fresh guard claims immediately), and a dead
// rival's disposal goes through the guarded path, so concurrent disposers
// stay serialized and any dead chain unwinds one level per walk.
const maxReclaimGuardAttempts = 2

// reclaimBackoff is the guard claim's pause between attempts — the same
// delay ClaimSection's callers pass for a guard claim, named here because
// claimGuard's live-rival backoff uses it inline.
const reclaimBackoff = 2 * time.Millisecond

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
			release, lerr := claimAndLabel(path, perm)
			if lerr != nil {
				return nil, lerr
			}
			return release, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("claim section %s: lock: %w", path, err)
		}
		// Contention: check whether the holder is a verified-dead owner. A
		// live owner blocks through the budget; a verified-dead one is
		// broken and the claim retried immediately. The caller's context
		// reaches the guard claim, so a cancelled claim does not re-walk
		// any guard chain.
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
		// The release read is BOUNDED (review gate finding, P2): the
		// release runs in the caller's defer, so no caller deadline reaches
		// it — a path swapped for a FIFO after the claim parked the read
		// forever. sectionRereadFn refuses a non-regular path without
		// opening it and costs one small capped read; an unreadable path is
		// remove-nothing, the same safe direction as a foreign owner.
		raw, rerr := sectionRereadFn(lockPath)
		if rerr != nil {
			return nil // gone, swapped, or unreadable: remove nothing
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

// claimGuard takes a break's guard marker WITHOUT the ClaimSection
// recursion that once self-propagated the chain (card t1606): one O_EXCL
// attempt, and on contention the rival guard's disposal runs through the
// GUARDED path (BreakStaleLockContext — the rival's own guard is claimed
// first, so two reclaimers of the same dead guard can never interleave a
// verdict with the other's live re-acquisition) and the claim retried. A
// live rival guard owns its disposal: the caller refuses and its own
// retry budget backs off, exactly as a live section holder blocks a
// claim.
func claimGuard(ctx context.Context, guardPath string) (func() error, bool) {
	for range maxReclaimGuardAttempts {
		if err := ctx.Err(); err != nil {
			return nil, false
		}
		err := Claim(guardPath, 0o600)
		if err == nil {
			release, lerr := claimAndLabel(guardPath, 0o600)
			if lerr != nil {
				return nil, false
			}
			return release, true
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, false
		}
		// A LIVE rival guard owns its disposal: refuse at once, without
		// entering the deeper reclaim walk (review-gate finding on card
		// t1606, round 2 — with concurrently-held guards the walk through
		// claimGuard → BreakStaleLockContext → claimGuard multiplied the
		// wait exponentially, 1.4s at 8 levels vs the base's 107ms). The
		// read is the same bounded one every verdict uses.
		if raw, rerr := sectionRereadFn(guardPath); rerr == nil {
			if owner, ok := OwnerFromBytes(raw); ok && !OwnerIsDead(owner) {
				select {
				case <-ctx.Done():
					return nil, false
				case <-time.After(reclaimBackoff):
				}
				continue
			}
		}
		// A rival guard that is not verifiably live blocks the claim. Its
		// disposal must be SERIALIZED (review-gate finding on card t1606,
		// P1): a bare verify-and-delete lets two reclaimers of the same
		// dead guard interleave so one's late delete removes the other's
		// LIVE re-acquired guard. Route the disposal through the guarded
		// path — the rival's OWN guard is claimed first, exactly like
		// every other delete here. That guarded claim is itself the one
		// place a deeper guard marker is created (transient: released as
		// soon as the disposal ends); a dead chain unwinds one level per
		// walk, and the guarded path's own verified-bytes gate aborts if
		// the rival re-acquired between this pre-check and the disposal.
		if !BreakStaleLockContext(ctx, guardPath) {
			// The disposal walk failed: a live owner sits at some depth, or
			// a rival reclaimer is mid-walk. Neither changes within this
			// function's backoff, and a retry here would re-walk the whole
			// failed sub-chain — 2^depth reads on a dead chain blocked by a
			// live tail (review-gate finding on card t1606, round 3:
			// 1,022 reads / 2.6s at 8 dead guards + a live tail). Return at
			// once; the caller's own retry budget re-enters later, against
			// whatever the state has become.
			slog.Warn("lock section: guard reclaim walk failed; refusing",
				"guard", guardPath)
			return nil, false
		}
		// The disposal cleared the path — claim it NOW, outside the attempt
		// budget: an attempt that ends here with a false would waste the
		// caller's budget round-trip on a path this call just cleared.
		// (review-gate finding on card t1606, round 4: the walk may have
		// consumed time while the caller's context was cancelled — a
		// cancelled caller stops here like every other retry boundary.)
		if cerr := ctx.Err(); cerr != nil {
			return nil, false
		}
		if err := Claim(guardPath, 0o600); err != nil {
			continue // a rival re-claimed between disposal and this claim
		}
		release, lerr := claimAndLabel(guardPath, 0o600)
		if lerr != nil {
			return nil, false
		}
		return release, true
	}
	slog.Warn("lock section: guard claim exhausted its attempts", "guard", guardPath)
	return nil, false
}

// claimAndLabel labels a just-claimed lock file, removing it on label
// failure. It is the acquire tail ClaimSection and claimGuard share — one
// label write, one conservative failure direction.
func claimAndLabel(path string, perm os.FileMode) (func() error, error) {
	if werr := writeOwnerLabel(path, perm); werr != nil {
		// The lock is HELD but unlabelled: release immediately and report —
		// an unlabelled lock could never be verified, and wedging on write
		// failure beats breaking the invariant.
		_ = os.Remove(path)
		return nil, fmt.Errorf("claim section %s: labelling lock: %w", path, werr)
	}
	return releaseSectionFunc(path), nil
}

// BreakStaleLockContext is BreakStaleLock under a caller's context: the
// cancellation reaches the guard claim's every retry boundary, so a
// cancelled caller's walk stops at the next claim boundary instead of
// burning its whole budget.
func BreakStaleLockContext(ctx context.Context, path string) bool {
	if strings.HasSuffix(path, reclaimSuffix) || strings.HasSuffix(path, breakingSuffix) {
		// Reclaiming a marker — a breaker's (.breaking) or a reclaimer's
		// (.reclaim): hold ITS OWN guard marker first — delete only on
		// creation success (review-gate residual: the .reclaim path's
		// verify-and-delete was bare, so a reclaimer pausing between its
		// check and its delete could remove a rival's LIVE re-acquired
		// guard). The guard is claimed through claimGuard, whose rival
		// disposal may itself claim ONE deeper guard level (transient —
		// released when that disposal ends): what is gone is the old
		// ClaimSection recursion that spawned a new level for EVERY
		// blocked contender, live ones included, until the depth cap
		// refused permanently. A dead chain now unwinds one level per
		// walk, bounded by the filesystem path length the suffix chain
		// can occupy.
		release, ok := claimGuard(ctx, path+reclaimSuffix)
		if !ok {
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
	// verdict was made on, immediately before the unlink. The re-check reads
	// through the same BOUNDED read as the verdict (review-gate finding on
	// card t1606, P2): a plain os.ReadFile here parked forever on a FIFO
	// swapped in after the first read — the caller's context never reached
	// it. sectionRereadFn refuses a non-regular path without opening it.
	now, err := sectionRereadFn(path)
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
