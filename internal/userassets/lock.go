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
func lockOwnerGone(path string) bool {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
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
// the caller assumed. rename(2) makes exactly one racing acquirer the
// winner; the losers see ENOENT and re-run their create loop (the lock
// file's F5 posture). Reports whether THIS caller reclaimed the marker.
func reclaimGuardMarker(markerPath string) bool {
	if !guardMarkerDead(markerPath) {
		return false // never move a marker whose owner is not PROVEN dead
	}
	reclaim := markerPath + ".reclaim-" + fmt.Sprintf("%d-%d", os.Getpid(), time.Now().UnixNano())
	if err := os.Rename(markerPath, reclaim); err != nil {
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
	GuardMarkerAbsent     GuardMarkerState = iota // no marker on disk
	GuardMarkerOwnerDead                          // pid record + owner proven dead — reclaimable
	GuardMarkerOwnerAlive                         // pid record + owner still running
	GuardMarkerOwnerless                          // no pid record — NEVER auto-reclaimed (design §3)
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
	}
	return "unknown"
}

// ClassifyGuardMarker classifies the marker at path for the doctor's
// visible-recovery row. The pid-less form is Ownerless — the suspended-
// process shape age alone can never refute — and the returned pid is 0
// whenever no readable pid record exists.
func ClassifyGuardMarker(markerPath string) (GuardMarkerState, int) {
	raw, err := os.ReadFile(markerPath)
	if err != nil {
		return GuardMarkerAbsent, 0
	}
	pid := 0
	hasPID := false
	for _, field := range strings.Fields(string(raw)) {
		if strings.HasPrefix(field, "pid=") {
			if parsed, convErr := strconv.Atoi(strings.TrimPrefix(field, "pid=")); convErr == nil && parsed > 0 && int64(parsed) <= 2147483647 {
				pid = parsed
				hasPID = true
			}
		}
	}
	if !hasPID {
		return GuardMarkerOwnerless, 0
	}
	if lockProcessGone(pid) {
		return GuardMarkerOwnerDead, pid
	}
	return GuardMarkerOwnerAlive, pid
}
