// integration_lock.go — the release-integration holder lock (card t194).
//
// The doctrine this backs (`kanban-dispatch.md` § Integration into the release
// branch is self-served) serializes lanes by ANNOUNCEMENT: a lane tells the
// lead before entering the release worktree, the lead broadcasts the hold, and
// no other session enters until the completion report. Card t181 wrote that
// rule and named its own gap in the same breath — announcement is a social
// protocol, so nothing stops a lane that skips it, and the check such a lane
// still runs (`git rev-parse -q --verify MERGE_HEAD` printing nothing) is
// exactly the insufficient probe t181 exists to correct: MERGE_HEAD is equally
// absent while another lane is mid-resolution.
//
// This file is the mechanical layer under that rule. It is deliberately NOT
// the board lock next door (board_lock.go), and the difference is lifetime:
// the board lock spans one process's read-modify-write and is an flock, so it
// dies with the process that took it. An integration window spans many CLI
// invocations, many turns, and minutes of human-paced work — an fd cannot
// represent it. So the window is a RECORD whose validity is decided by the
// recorded holder's liveness, and the flock discipline is borrowed only to
// serialize mutations of that record.
//
// What this does NOT do: it does not make the release worktree unwritable, and
// it cannot. A lock that a determined caller may remove is a coordination
// signal, not a capability boundary. Its value is that skipping the
// announcement now requires a deliberate act (`--force`, recorded) rather than
// an honest mistake.
package kanban

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// IntegrationLockFileName names the record inside the project's state dir.
const IntegrationLockFileName = "integration-lock.json"

// ErrIntegrationLockHeld is returned by AcquireIntegrationLock when a LIVE
// holder other than the caller already owns the window.
var ErrIntegrationLockHeld = errors.New("release integration window held by another session")

// IsIntegrationLockHeld reports whether err is the contention sentinel.
func IsIntegrationLockHeld(err error) bool { return errors.Is(err, ErrIntegrationLockHeld) }

// ErrIntegrationLockNotHeld is returned by ReleaseIntegrationLock when no
// record exists. Releasing nothing is reported rather than silently accepted:
// a lane that believes it released a window it never held has a broken model
// of the board, and the quiet success is what would preserve that belief.
var ErrIntegrationLockNotHeld = errors.New("no release integration window is held")

// IsIntegrationLockNotHeld reports whether err is the empty-release sentinel.
func IsIntegrationLockNotHeld(err error) bool { return errors.Is(err, ErrIntegrationLockNotHeld) }

// ErrIntegrationLockForeign is returned by ReleaseIntegrationLock when a
// DIFFERENT session holds the window. Releasing another lane's window is the
// same defect as entering it.
var ErrIntegrationLockForeign = errors.New("release integration window is held by a different session")

// IsIntegrationLockForeign reports whether err is the foreign-release sentinel.
func IsIntegrationLockForeign(err error) bool { return errors.Is(err, ErrIntegrationLockForeign) }

// PIDSourceSessionOwner marks a record whose PID names the OWNING SESSION
// rather than whichever process wrote the record. It is the discriminator
// between the two record shapes on disk: a record carrying it was written by a
// caller that resolved the owner (and pid 0 there means "owner unresolvable",
// not "no pid"), while a record without it predates the anchor and is read
// exactly as it always was.
const PIDSourceSessionOwner = "session-owner"

// integrationLockMutationTestHook is a nil-by-default, TEST-ONLY interleaving
// point invoked once between the acquire decision and the write. It exists so
// the cross-process criterion can CONSTRUCT the read-modify-write interleaving
// instead of waiting for it: the unserialized window is one read, a branch, and
// one write — tens of microseconds — so a barrier-released pair hits it only by
// luck, and a criterion that waits for luck has no stop rule.
//
// It is unexported and package-level, so only `package kanban` can assign it,
// and no non-test file does (the closure gate greps for the assignment). Every
// production path leaves it nil, and the call site below is nil-guarded — a
// nil func() invoked in Go panics, so the guard is what makes "with the hook
// nil, behavior is byte-for-byte unchanged" true rather than merely intended.
var integrationLockMutationTestHook func()

// IntegrationLock is the recorded holder of the release integration window.
//
// SessionID is the address a human and a peer both use; PID is what makes
// staleness decidable. Both are recorded because neither alone suffices: a
// session id cannot be probed for liveness, and a pid cannot be addressed in a
// dispatch.
//
// PIDSource records WHOSE pid the PID field is. It is additive and optional:
// its absence is a legacy record, and no read path branches on it — the probe
// below already answers correctly for both shapes. It exists so a human (and a
// future reader) can tell an unresolvable-owner record apart from one written
// before the anchor existed, which the pid alone cannot express.
type IntegrationLock struct {
	SessionID   string `json:"session_id"`
	SessionName string `json:"session_name,omitempty"`
	PID         int    `json:"pid"`
	PIDSource   string `json:"pid_source,omitempty"`
	Branch      string `json:"branch"`
	Worktree    string `json:"worktree"`
	AcquiredAt  string `json:"acquired_at"`
	Card        string `json:"card,omitempty"`
}

// Held reports whether the record names a holder at all.
func (l *IntegrationLock) Held() bool { return l != nil && l.SessionID != "" }

// Stale reports whether the recorded holder's process is gone.
//
// Indeterminate reads as LIVE, deliberately: treating an unprobeable holder as
// dead would clear a window that may still be in use, and the cost of a false
// "stale" (two lanes merging at once) is the failure this lock exists to
// prevent, while the cost of a false "live" is one operator asking the holder
// to release. The asymmetry is not close.
//
// A pid of 0 is the anchored form of that same indeterminacy — the acquirer
// could not resolve its owning session — and it takes the same answer: live,
// releasable only by an explicit release or a recorded --force. The single
// probe below serves both record shapes, so there is no marker-conditional
// branch here and none is wanted: a legacy record's dead pid still reads
// reclaimable exactly as it did before the anchor existed.
func (l *IntegrationLock) Stale() bool {
	if !l.Held() {
		return false
	}
	if l.PID <= 0 {
		return false
	}
	return !FactoryProcessAlive(l.PID)
}

// integrationLockPath resolves the record's location under the project's
// state directory. The caller passes the PRIMARY checkout's root: the record
// must be visible from every linked worktree, and only the primary's
// `.moai/state` is shared by all of them.
func integrationLockPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".moai", "state", IntegrationLockFileName)
}

// ReadIntegrationLock returns the recorded holder, or a nil-valued lock when
// no record exists.
//
// A record that cannot be parsed is reported as an error rather than treated
// as absent. "Unreadable" and "nobody holds it" are different states, and
// collapsing them would let a corrupted record read as a free window — the
// exact substitution (absence of a signal for evidence of freedom) that t181
// found in the MERGE_HEAD probe.
func ReadIntegrationLock(projectRoot string) (*IntegrationLock, error) {
	data, err := os.ReadFile(integrationLockPath(projectRoot))
	if errors.Is(err, os.ErrNotExist) {
		return &IntegrationLock{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read integration lock: %w", err)
	}
	var lock IntegrationLock
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("integration lock at %s is unreadable: %w", integrationLockPath(projectRoot), err)
	}
	return &lock, nil
}

// AcquireIntegrationLock records sessionID as the holder of the release
// integration window.
//
// Re-acquiring a window the caller already holds succeeds and refreshes the
// record, so a lane that re-enters after a `/clear` is not locked out of its
// own window.
//
// A STALE record (recorded holder's process gone) is taken over, and the
// takeover is returned in `replaced` so the caller can report what it cleared
// rather than silently overwriting another lane's trace.
//
// force takes over a LIVE holder. It exists because a wedged session must not
// be able to block the batch forever, and it is never the quiet path: the
// caller is expected to surface `replaced`.
//
// want.PID is recorded verbatim, including zero. This function deliberately
// does NOT fill an unset pid with os.Getpid(): a window outlives the process
// that records it, so this process's pid is dead by the time any reader probes
// it, and filling the field with it made every record read as abandoned the
// instant it was written. Resolving the owner is the CALLER's job (the acquire
// verb uses session.ResolveOwnerPID); an unset pid arriving here means the
// caller could not resolve one, and the conservative reading of that — live
// until released — is Stale()'s.
func AcquireIntegrationLock(projectRoot string, want IntegrationLock, force bool) (replaced *IntegrationLock, err error) {
	if want.SessionID == "" {
		return nil, errors.New("integration lock: session id is required")
	}
	path := integrationLockPath(projectRoot)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("integration lock: %w", err)
	}

	// The read → decide → write sequence runs inside the mutation lock, so at
	// most one process at a time is inside it for a given project root
	// (integration_lock_mutation.go). The READ is INSIDE the section, not
	// duplicated outside it: a caller serialized behind another must decide
	// against the state the previous mutation published, never against a read
	// taken before the wait.
	if mutErr := withIntegrationLockMutation(projectRoot, func() error {
		current, readErr := ReadIntegrationLock(projectRoot)
		if readErr != nil && !force {
			return readErr
		}
		if current.Held() && current.SessionID != want.SessionID {
			switch {
			case force:
				replaced = current
			case current.Stale():
				replaced = current
			default:
				return fmt.Errorf("%w: %s (pid %d) since %s on %s",
					ErrIntegrationLockHeld, current.holderLabel(), current.PID, current.AcquiredAt, current.Branch)
			}
		}

		if want.AcquiredAt == "" {
			want.AcquiredAt = time.Now().UTC().Format(time.RFC3339)
		}
		// Between the decision and the write — the exact window this file's
		// serialization discipline is about. Nil in production; see the hook's
		// declaration for why the guard is load-bearing rather than defensive.
		if integrationLockMutationTestHook != nil {
			integrationLockMutationTestHook()
		}
		return writeIntegrationLock(path, &want)
	}); mutErr != nil {
		return nil, mutErr
	}
	return replaced, nil
}

// ReleaseIntegrationLock removes the record when sessionID is its holder.
//
// force releases a foreign window, for the same wedged-holder reason acquire
// carries it.
func ReleaseIntegrationLock(projectRoot, sessionID string, force bool) (released *IntegrationLock, err error) {
	// Same critical section as acquire, for the same reason: read → holder
	// check → remove is a read-modify-write too. Every sentinel and every
	// message below is byte-identical to what it was before the section
	// existed — this changes WHEN mutations interleave, never WHAT a given
	// single-threaded call decides or says.
	if mutErr := withIntegrationLockMutation(projectRoot, func() error {
		current, readErr := ReadIntegrationLock(projectRoot)
		if readErr != nil && !force {
			return readErr
		}
		if current == nil || !current.Held() {
			return ErrIntegrationLockNotHeld
		}
		if current.SessionID != sessionID && !force {
			return fmt.Errorf("%w: %s (pid %d) holds it", ErrIntegrationLockForeign, current.holderLabel(), current.PID)
		}
		if remErr := os.Remove(integrationLockPath(projectRoot)); remErr != nil && !errors.Is(remErr, os.ErrNotExist) {
			return fmt.Errorf("integration lock: %w", remErr)
		}
		released = current
		return nil
	}); mutErr != nil {
		return nil, mutErr
	}
	return released, nil
}

// holderLabel prefers the human-facing session name and falls back to the id.
func (l *IntegrationLock) holderLabel() string {
	if l == nil {
		return "unknown"
	}
	if l.SessionName != "" {
		return l.SessionName
	}
	if l.SessionID != "" {
		return l.SessionID
	}
	return "unknown"
}

// writeIntegrationLock replaces the record atomically: the reader is a guard
// on a hot path, and a torn read there would surface as "unreadable" — which
// this package deliberately treats as a hard error rather than as a free
// window, so a non-atomic write would convert a routine acquire into a blocked
// integration.
func writeIntegrationLock(path string, lock *IntegrationLock) error {
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return fmt.Errorf("integration lock: %w", err)
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("integration lock: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("integration lock: %w", err)
	}
	return nil
}
