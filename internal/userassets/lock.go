// lock.go — the user-level manifest read-modify-write lock (round-5 F4,
// REQ-006). Manifest mutation is serialized per user: the lock spans
// manifest READ → asset changes → manifest SAVE, so concurrent
// init/update/bundle runs from different projects of the same user cannot
// lose one another's writes (bundle-list entries, file records).
//
// The lock is a file under ~/.moai/ — a user-level concern, because the
// concurrent writers are different projects sharing one manifest. The file
// is created O_EXCL and carries the holder's PID; a lock whose mtime is
// older than staleAfter is taken over only after its holder is confirmed
// gone. Age alone never authorizes taking a live holder's lock.
package userassets

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// DefaultStaleAfter bounds how long a lock file may sit untouched before a
// later run takes it over. Generous by design: a run's install phase can
// legitimately take minutes on a cold tree.
const DefaultStaleAfter = 30 * time.Minute

// UserLock is a held user-level lock. Release unlocks and removes the file —
// ONLY if the lock file still carries this holder's identity token (the
// mid-run review fix RF8: after a stale reclaim, the original holder's
// late Release must not delete the NEW owner's lock).
type UserLock struct {
	path  string
	token string
}

// AcquireUserLock takes the user-level lock under the given moai home,
// retrying until timeout. It returns ErrLocked when the window elapses
// while a live holder keeps the lock.
func AcquireUserLock(home string, timeout time.Duration) (*UserLock, error) {
	return acquireUserLockStale(LockPath(home), timeout, DefaultStaleAfter)
}

func acquireUserLockStale(path string, timeout, staleAfter time.Duration) (*UserLock, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("userassets: mkdir lock home: %w", err)
	}
	// A3/B2 (review-fix round 2 addendum): the whole reclaim-then-acquire
	// sequence runs under an OS-level guard (flock on unix) so two racing
	// callers serialize — the rename-based reclaim alone did not give
	// manifest-mutation mutual exclusion (21-32 concurrent owners
	// reproduced by the gate).
	guardPath := strings.TrimSuffix(path, ".lock") + ".acquire-guard"
	releaseGuard, guardErr := acquireGuard(guardPath, timeout)
	if guardErr != nil {
		return nil, fmt.Errorf("userassets: acquire guard: %w", guardErr)
	}
	defer releaseGuard()
	deadline := time.Now().Add(timeout)
	for {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			token := fmt.Sprintf("%d-%s", os.Getpid(), time.Now().UTC().Format(time.RFC3339Nano))
			_, _ = fmt.Fprintf(f, "pid=%d token=%s acquired=%s\n", os.Getpid(), token, time.Now().UTC().Format(time.RFC3339))
			_ = f.Close()
			return &UserLock{path: path, token: token}, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("userassets: acquire lock: %w", err)
		}
		// Held. Take over a stale lock — a crashed holder must not wedge
		// the user's manifest forever. F5 (review-fix round 2): the
		// takeover is ATOMIC — rename(2) moves the stale lock to a unique
		// reclaim name; exactly one racing caller succeeds, the losers see
		// ENOENT (someone else reclaimed) and re-run the create loop. The
		// former os.Remove-based takeover let a second caller delete the
		// lock the first had just ACQUIRED.
		if info, statErr := os.Stat(path); statErr == nil && time.Since(info.ModTime()) > staleAfter && lockOwnerGone(path) {
			reclaim := path + ".reclaim-" + fmt.Sprintf("%d-%d", os.Getpid(), time.Now().UnixNano())
			if renameErr := os.Rename(path, reclaim); renameErr == nil {
				_ = os.Remove(reclaim)
				continue
			} else if os.IsNotExist(renameErr) {
				continue
			}
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("%w: %s", ErrLocked, path)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// Release removes the lock file — only when it still carries THIS holder's
// identity token. A lock reclaimed by another owner (the stale-reclaim path)
// is left alone: the original holder's late release neither deletes the new
// owner's lock nor opens a third-run window.
func (l *UserLock) Release() error {
	if l == nil {
		return nil
	}
	raw, err := os.ReadFile(l.path)
	if err == nil && !strings.Contains(string(raw), "token="+l.token) {
		return fmt.Errorf("userassets: lock at %s was reclaimed by another owner — release skipped", l.path)
	}
	if err := os.Remove(l.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("userassets: release lock: %w", err)
	}
	return nil
}

// lockOwnerGone fails closed for unreadable or malformed ownership records.
// PID reuse can delay recovery, but cannot authorize a second live writer.
// Gate round 19: the record's file TYPE is judged before any read — a FIFO
// at a marker path would hang os.ReadFile forever waiting for a writer
// (the install.go:211 hazard class), so a non-regular file fails closed.
func lockOwnerGone(path string) bool {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	raw, readErr := os.ReadFile(path)
	if readErr != nil {
		return false
	}
	return ownerRecordGone(raw)
}

// ownerRecordGone parses one pid-record's death evidence from raw bytes.
func ownerRecordGone(raw []byte) bool {
	for _, field := range strings.Fields(string(raw)) {
		if strings.HasPrefix(field, "pid=") {
			pid, err := strconv.Atoi(strings.TrimPrefix(field, "pid="))
			return err == nil && pid > 0 && int64(pid) <= 2147483647 && lockProcessGone(pid)
		}
	}
	return false
}

// guardMarkerDead reports whether the guard marker at path records an owner
// PROVEN dead — the reclaim license of REQ-LOCK-001 (SPEC-USERASSET-
// DEPLOY-GUARD-001 M2). It is the lockOwnerGone posture applied to the
// marker: a marker carrying no pid record (a legacy or foreign marker) is
// never proven dead, and unreadable or malformed records fail closed. The
// platform-neutral home is deliberate: both platform guards decide their
// reclaim through this one predicate, so the marker-state × owner-liveness
// table judges the shared policy on every platform.
func guardMarkerDead(markerPath string) bool {
	return lockOwnerGone(markerPath)
}

// reclaimGuardMarker moves a death-proven marker aside. It re-proves the
// owner itself (fail closed): an unproven marker is never moved, whatever
// the caller assumed.
//
// Gate round 18 P1 — the reclaim IDENTITY race: two callers may verify the
// same dead owner; the loser's rename would then act on the WINNER'S newly
// created live marker and its delete would remove a live owner's marker
// (gate-reproduced). The rename is therefore IDENTITY-CHECKED: the bytes
// read before the rename must be byte-identical to the renamed file's —
// the marker record carries a nanosecond stamp, so a new owner's marker
// never matches. On a mismatch the displaced file is restored to its
// original path and the caller re-runs its loop.
func reclaimGuardMarker(markerPath string) bool {
	info, err := os.Lstat(markerPath)
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	proven, readErr := os.ReadFile(markerPath)
	if readErr != nil || !ownerRecordGone(proven) {
		return false // never move a marker whose owner is not PROVEN dead
	}
	reclaim := markerPath + ".reclaim-" + fmt.Sprintf("%d-%d", os.Getpid(), time.Now().UnixNano())
	if renameErr := os.Rename(markerPath, reclaim); renameErr != nil {
		return false
	}
	displaced, readErr := os.ReadFile(reclaim)
	if readErr != nil || string(displaced) != string(proven) {
		// Not the marker whose death we proved — a racing winner's live
		// marker took the path. Put it back exactly where it was.
		_ = os.Rename(reclaim, markerPath)
		return false
	}
	_ = os.Remove(reclaim)
	return true
}

// GuardMarkerState is the doctor-visible classification of a lock marker
// (REQ-LOCK-001's visible-recovery path — the doctor reports; only a
// death-PROVEN marker is auto-reclaimed, by the next acquisition).
type GuardMarkerState int

const (
	GuardMarkerAbsent     GuardMarkerState = iota // no marker on disk (or a clean leftover with no recovery meaning)
	GuardMarkerOwnerDead                          // pid record + owner proven dead — reclaimable
	GuardMarkerOwnerAlive                         // pid record + owner still running (or a flock-held unix guard)
	GuardMarkerOwnerless                          // no pid record on a platform where the record is the truth — NEVER auto-reclaimed (design §3)
	GuardMarkerIrregular                          // present but not a readable regular file — surfaced, never read
)

// String renders the state for reports.
func (s GuardMarkerState) String() string {
	switch s {
	case GuardMarkerAbsent:
		return "absent"
	case GuardMarkerOwnerDead:
		return "owner-dead"
	case GuardMarkerOwnerAlive:
		return "owner-alive"
	case GuardMarkerOwnerless:
		return "ownerless"
	case GuardMarkerIrregular:
		return "irregular"
	}
	return "unknown"
}

// ClassifyGuardMarker classifies the GUARD marker at path for the doctor's
// visible-recovery row. Gate rounds 18/19: only os.IsNotExist reads as
// absent; the file TYPE is judged before any read (a FIFO must never be
// read — it hangs); and the pid-less residual is PLATFORM-DECIDED — on
// unix the guard file survives a clean release by design (flock is the
// truth), so a free file is a clean leftover while a held file is alive;
// on windows the marker's existence is the held evidence and a pid-less
// marker is ownerless. For the .LOCK file use ClassifyLockFile — the lock
// file takes no flock itself, so the flock-based absence judgment does not
// apply to it (gate round 20).
func ClassifyGuardMarker(markerPath string) (GuardMarkerState, int) {
	state, pid, has := classifyLockRecord(markerPath)
	if !has {
		return classifyPidlessMarker(markerPath)
	}
	return state, pid
}

// ClassifyLockFile classifies the .lock FILE for the doctor's row (gate
// round 20): the lock file is O_EXCL-guarded and carries no flock, so the
// guard file's flock-based absence judgment must not be applied to it — a
// pid-less leftover .lock is a genuine manual-recovery state (acquisition
// refuses with ErrLocked) on EVERY platform.
func ClassifyLockFile(lockPath string) (GuardMarkerState, int) {
	state, pid, has := classifyLockRecord(lockPath)
	if !has {
		return GuardMarkerOwnerless, 0
	}
	return state, pid
}

// classifyLockRecord reads and pid-classifies one lock-family record. The
// boolean reports whether a usable classification was reached (false = the
// pid-less residual, which the caller decides per file kind).
func classifyLockRecord(path string) (state GuardMarkerState, pid int, has bool) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return GuardMarkerAbsent, 0, true
		}
		return GuardMarkerIrregular, 0, true // access error — surfaced, not absence
	}
	if !info.Mode().IsRegular() {
		return GuardMarkerIrregular, 0, true
	}
	raw, readErr := os.ReadFile(path)
	if readErr != nil {
		return GuardMarkerIrregular, 0, true
	}
	for _, field := range strings.Fields(string(raw)) {
		if strings.HasPrefix(field, "pid=") {
			if parsed, convErr := strconv.Atoi(strings.TrimPrefix(field, "pid=")); convErr == nil && parsed > 0 && int64(parsed) <= 2147483647 {
				if lockProcessGone(parsed) {
					return GuardMarkerOwnerDead, parsed, true
				}
				return GuardMarkerOwnerAlive, parsed, true
			}
		}
	}
	return GuardMarkerOwnerless, 0, false
}
