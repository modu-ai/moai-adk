// integration_lock.go — the release-integration holder lock (card t194).
//
// The doctrine this backs (Factory Dispatch Protocol § Integration into the release
// branch is self-served) serializes lanes by ANNOUNCEMENT: a lane tells the
// leader before entering the release worktree, the leader broadcasts the hold, and
// no other session enters until the completion report. Card t181 wrote that
// rule and named its own gap in the same breath — announcement is a social
// protocol, so nothing stops a lane that skips it, and the check such a lane
// still runs (`git rev-parse -q --verify MERGE_HEAD` printing nothing) is
// exactly the insufficient probe t181 exists to correct: MERGE_HEAD is equally
// absent while another lane is mid-resolution.
//
// This file is the mechanical layer under that rule. It is deliberately NOT
// the board lock next door (state_lock.go), and the difference is lifetime:
// the board lock spans one process's read-modify-write and is an flock, so it
// dies with the process that took it. An integration window spans many CLI
// invocations, many turns, and minutes of human-paced work — an fd cannot
// represent it. So the window is a RECORD whose validity is decided by the
// recorded holder's liveness, and the flock discipline is borrowed only to
// serialize mutations of that record.
//
// That last clause described nothing for the whole of this file's first life
// (card t194 through t298): no exclusion primitive was taken anywhere in it,
// and the read → decide → write below ran unserialized, so two processes could
// each be told they held the window. Anyone who read the sentence believed the
// record was already serialized, which is the specific way a comment describing
// an absent mechanism does damage — it is not merely uninformative, it stops
// the next reader looking. Card t336 built the mechanism the sentence names:
// integration_lock_mutation.go borrows exactly that discipline, at a scope of
// its own, and the clause is true from that commit onward.
//
// What this does NOT do: it does not make the release worktree unwritable, and
// it cannot. A lock that a determined caller may remove is a coordination
// signal, not a capability boundary. Its value is that skipping the
// announcement now requires a deliberate act (`--force`, recorded) rather than
// an honest mistake.
package factory

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

// Branch provenance values (card t637): which resolution tier decided the
// recorded Branch. Branch keeps its meaning — the integration TARGET — and the
// source only says how that value was reached, so a reader can tell a
// configured target from a caller-tree fallback.
const (
	// BranchSourceFlag — a non-blank --branch value decided it.
	BranchSourceFlag = "flag"
	// BranchSourceConfig — the configured git-flow develop branch decided it.
	BranchSourceConfig = "config"
	// BranchSourceCaller — the caller's own checked-out tree decided it.
	BranchSourceCaller = "caller"
)

// integrationLockMutationTestHook is a nil-by-default, TEST-ONLY interleaving
// point invoked once between the acquire decision and the write. It exists so
// the cross-process criterion can CONSTRUCT the read-modify-write interleaving
// instead of waiting for it: the unserialized window is one read, a branch, and
// one write — tens of microseconds — so a barrier-released pair hits it only by
// luck, and a criterion that waits for luck has no stop rule.
//
// It is unexported and package-level, so only `package factory` can assign it,
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

	// BranchSource names the resolution tier that decided Branch (one of the
	// BranchSource* values). Additive and optional exactly like PIDSource: a
	// record written before it existed carries no key and is read as it
	// always was, and no read path decides on it — it is for the human
	// reading status, never an input to the guard.
	BranchSource string `json:"branch_source,omitempty"`

	// SettingsDriftBypass and SettingsDriftPreserved record that the window
	// was taken over a REFUSED settings-drift verdict (card t488), and where
	// the drifted working copy was preserved.
	//
	// Both are written only when a refusal was actually bypassed — that is,
	// only when the refusal layer was on and --allow-settings-drift was given.
	// With the layer off there is no refusal to bypass, and stamping the flag
	// there would make the record assert something that did not happen.
	//
	// They are deliberately NOT driven by --force: that flag means "take the
	// window from a live holder", a different decision, and one flag carrying
	// two of them leaves the record unable to say which was intended.
	SettingsDriftBypass    bool   `json:"settings_drift_bypass,omitempty"`
	SettingsDriftPreserved string `json:"settings_drift_preserved,omitempty"`

	// The merge-window queue (card t1479, SPEC-MERGE-WINDOW-QUEUE-001).
	// Every field below is additive and optional exactly like BranchSource:
	// a record written before it existed carries no key and is read as it
	// always was (REQ-MWQ-001), and no read path outside the window verbs
	// decides on it.
	//
	// Queue is the FIFO of waiting sessions. omitempty is load-bearing, not
	// cosmetic: with no ticket, the record must carry no queue key at all,
	// and a legacy record re-written by the new code must stay byte-shape
	// compatible.
	Queue []IntegrationTicket `json:"queue,omitempty"`

	// LeaseExpiresAt stamps the holder's lease (REQ-MWQ-008), RFC3339. An
	// absent stamp is a legacy record or a disabled lease: validity is then
	// decided by owning-session liveness alone, exactly as before this SPEC.
	LeaseExpiresAt string `json:"lease_expires_at,omitempty"`

	// Displaced records the last holder this record took the window from
	// (stale takeover, --force, or promotion past a displaced holder), with
	// DisplacedReason naming why. Today's stale takeover returns the replaced
	// record to its caller; the queue needs the record to persist the
	// displacement so a later `status` can show who was displaced (REQ-MWQ-006).
	Displaced       *IntegrationLock `json:"displaced,omitempty"`
	DisplacedReason string           `json:"displaced_reason,omitempty"`
}

// IntegrationTicket is one waiting session in the merge-window FIFO queue
// (card t1479, REQ-MWQ-001). The owning-session fields are a copy of the
// holder record's anchor — the pid resolved the same way acquire resolves it
// together with its session-owner source — so promotion can stamp them onto
// the holder record and the promoted holder behaves exactly like a directly
// acquired one (REQ-MWQ-006). The waiter fields are the ticket's own
// liveness: a waiting process refreshes Heartbeat every 15 seconds, and a
// ticket whose waiter process (matched on id AND start time) or owner is
// gone is dropped (REQ-MWQ-003).
type IntegrationTicket struct {
	SessionID    string `json:"session_id"`
	SessionName  string `json:"session_name,omitempty"`
	Card         string `json:"card,omitempty"`
	OwnerPID     int    `json:"owner_pid"`
	PIDSource    string `json:"pid_source,omitempty"`
	Branch       string `json:"branch,omitempty"`
	BranchSource string `json:"branch_source,omitempty"`
	Worktree     string `json:"worktree,omitempty"`
	WaiterPID    int    `json:"waiter_pid"`
	// WaiterStart is the waiter process's start instant (RFC3339Nano). The
	// pair (WaiterPID, WaiterStart) is the ticket's waiter identity: pids
	// recycle, so a live process with the same pid but a different start is
	// not this waiter.
	WaiterStart string `json:"waiter_start,omitempty"`
	Heartbeat   string `json:"heartbeat"`
	EnqueuedAt  string `json:"enqueued_at"`
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
// record — matched on the SESSION ID ALONE. A lane whose id rotated is
// therefore refused: `/clear` issues a new id to the same process, the record
// keeps the old one, and the refusal names the caller's own pid as the blocking
// holder. Such a lane releases first and then re-acquires; the release side
// admits it directly, via releasableBy.
//
// The two sides are deliberately asymmetric, and the asymmetry is the point
// rather than an oversight awaiting symmetry. They ask different questions:
// release asks "is this caller the holder ITSELF?", acquire asks "may this
// caller TAKE the window?". A pid shared by two sessions answers the first —
// it is evidence enough to let a caller free its OWN window — and does not
// answer the second, because one owning process can carry several sessions.
// Card t959 measured what happens when acquire borrows releasableBy anyway:
// TestIntegrationLockAcquire_SerializedAcrossProcesses records two lanes
// holding the window at once (successes=2, refusals=0), because both name the
// same owner pid. The cost of admitting a caller wrongly is not symmetric
// either — on release it frees a window that was the caller's, on acquire it is
// the concurrent merge this lock exists to prevent. Stale() states the same
// asymmetry on its own axis just above.
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
	return AcquireIntegrationWindow(projectRoot, want, force, nil)
}

// AcquireWindowOptions carries the merge-window queue's extensions to the
// acquire decision (card t1479). A nil options is the pre-queue behavior —
// AcquireIntegrationLock passes nil, which is why no pre-queue caller
// changes.
type AcquireWindowOptions struct {
	// ViaWait marks a --wait acquire: REQ-MWQ-011 refuses a no-wait acquire
	// while a live ticket is queued, and REQ-MWQ-012 refuses it under hold.
	ViaWait bool
	// LeaseDuration stamps the holder's lease (REQ-MWQ-008). Zero means the
	// shipped default; a NEGATIVE value disables the lease entirely (the
	// configured-zero reading — StampLease clears the stamp).
	LeaseDuration time.Duration
	// Probe overrides the liveness seam; nil is the production probe.
	Probe WindowProcProbe
}

func (o *AcquireWindowOptions) lease() time.Duration {
	if o == nil || o.LeaseDuration == 0 {
		return IntegrationLeaseDefault
	}
	return o.LeaseDuration
}

func (o *AcquireWindowOptions) probe() WindowProcProbe {
	if o == nil || o.Probe.OwnerAlive == nil {
		return DefaultWindowProcProbe()
	}
	return o.Probe
}

// ErrIntegrationWindowHold is returned by AcquireIntegrationWindow when the
// window policy is hold (REQ-MWQ-012): a no-wait acquire refuses naming the
// reason, and a --wait acquire is told to enqueue. The reason text travels
// in the error's message.
var ErrIntegrationWindowHold = errors.New("release integration window policy is hold")

// IsIntegrationWindowHold reports whether err is the hold sentinel.
func IsIntegrationWindowHold(err error) bool { return errors.Is(err, ErrIntegrationWindowHold) }

// AcquireIntegrationWindow is the queue-aware acquire (card t1479): inside
// the one mutation it applies the liveness refresh (drops + promotion), then
// decides against the refreshed record. The holder refusals keep their
// pre-queue byte shape (REQ-MWQ-010): the refresh runs FIRST, so a record it
// changes is already written back before the refusal formats, and a record
// it does not change formats the refusal exactly as before.
func AcquireIntegrationWindow(projectRoot string, want IntegrationLock, force bool, opts *AcquireWindowOptions) (replaced *IntegrationLock, err error) {
	if want.SessionID == "" {
		return nil, errors.New("integration lock: session id is required")
	}
	viaWait, lease, probe := false, opts.lease(), opts.probe()
	if opts != nil {
		viaWait = opts.ViaWait
	}
	path := integrationLockPath(projectRoot)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("integration lock: %w", err)
	}

	if mutErr := withIntegrationLockMutation(projectRoot, func() error {
		current, readErr := ReadIntegrationLock(projectRoot)
		if readErr != nil {
			if !force {
				return readErr
			}
			// P2-8 (card-review r1): --force IS the wedged-window recovery —
			// an unreadable record is exactly when a leader reaches for it —
			// so the failed read yields an EMPTY record here instead of a
			// nil dereference at the decision below.
			current = &IntegrationLock{}
		}
		policy, policyErr := ReadIntegrationWindowPolicy(projectRoot)
		if policyErr != nil {
			return policyErr
		}
		// The queue refresh every mutation applies (REQ-MWQ-003 drops,
		// REQ-MWQ-006/007 promotion) — the same serialized mutation the
		// decision below runs in, which is the ordering REQ-MWQ-002 asks
		// the enqueue to decide inside.
		report := RefreshWindow(current, policy, probe, WindowClock(), lease)
		// P2-9: a stale holder the refresh cleared is recorded ON the
		// record now — surface it as this acquire's takeover so the
		// "never silent" promise of the pre-queue takeover holds.
		if replaced == nil && !current.Held() && current.Displaced != nil {
			replaced = current.Displaced
		}
		// F8 (card-review r3): a refusal return used to abort BEFORE the
		// write, so a record the refresh just changed stayed stale on disk
		// while the refusal named the refreshed state (first's expired hold
		// with the ticket still queued, under an error that promoted that
		// ticket). REQ-MWQ-010's invariant — a record the refresh changes is
		// written back before the refusal formats — is made true here,
		// conditionally: a record the refresh did NOT change is not
		// rewritten, and its refusal formats exactly as before.
		persistRefreshed := func() error {
			if len(report.Dropped) == 0 && report.Promoted == nil && !report.Displaced {
				return nil
			}
			return writeIntegrationLock(path, current)
		}

		// REQ-MWQ-012 (card-review r3 F2): under hold, EVERY acquire of a
		// window the caller does not already hold refuses naming the reason.
		// The former viaWait carve-out let a --wait acquire on an EMPTY
		// window through to the grant path, voiding the hold exactly where
		// nothing else stands guard. The verb enqueues on this sentinel, so
		// a --wait caller comes back to the queue through it either way.
		if policy.Policy == PolicyHold && (!current.Held() || current.SessionID != want.SessionID) {
			if err := persistRefreshed(); err != nil {
				return err
			}
			return fmt.Errorf("%w: held by policy (%s)", ErrIntegrationWindowHold, policy.Reason)
		}
		// REQ-MWQ-011: the window is not granted to a no-wait acquire while
		// a live ticket is queued — the queue decides first, always. --force
		// is the recorded exception: the forcer takes the window and the
		// queue order survives it.
		if !viaWait && !force && len(current.Queue) > 0 && (!current.Held() || current.SessionID != want.SessionID) {
			if err := persistRefreshed(); err != nil {
				return err
			}
			return fmt.Errorf("%w: %s (pid %d) since %s on %s — %d live ticket(s) queued; acquire --wait enqueues behind them",
				ErrIntegrationLockHeld, current.holderLabel(), current.PID, current.AcquiredAt, current.Branch, len(current.Queue))
		}

		if current.Held() && current.SessionID != want.SessionID {
			switch {
			case force:
				// The recorded --force seizure (pre-queue shape kept): the
				// queue order survives untouched (REQ-MWQ-011) and the
				// displaced holder is recorded. The snapshot keeps the LAST
				// holder only (t1576 review round 9's class) — the force
				// path nested the previous Displaced chain the same way.
				want.Queue = current.Queue
				displaced := *current
				displaced.Displaced = nil
				displaced.Queue = nil
				want.Displaced = &displaced
				want.DisplacedReason = fmt.Sprintf("taken by force from %s at %s", displaced.SessionID, WindowClock().Format(time.RFC3339))
				replaced = current
			case current.Stale():
				// The stale takeover. After the refresh above, a stale
				// holder with queued tickets was already replaced by the
				// first ticket's promotion, so what remains here is the
				// queue-less shape — the pre-queue takeover, unchanged.
				replaced = current
			default:
				if err := persistRefreshed(); err != nil {
					return err
				}
				return fmt.Errorf("%w: %s (pid %d) since %s on %s",
					ErrIntegrationLockHeld, current.holderLabel(), current.PID, current.AcquiredAt, current.Branch)
			}
		}

		if want.AcquiredAt == "" {
			want.AcquiredAt = time.Now().UTC().Format(time.RFC3339)
		}
		// P2-1 (card-review r1): a RE-ACQUIRE (the holder refreshing) must
		// carry the queue through — want carries none, so the rewrite below
		// would wipe every queued ticket.
		if want.Queue == nil && len(current.Queue) > 0 {
			want.Queue = current.Queue
		}
		if want.LeaseExpiresAt == "" {
			// REQ-MWQ-008: every acquire stamps the holder's lease.
			StampLease(&want, WindowClock(), lease)
		}
		if integrationLockMutationTestHook != nil {
			integrationLockMutationTestHook()
		}
		return writeIntegrationLock(path, &want)
	}); mutErr != nil {
		return nil, mutErr
	}
	return replaced, nil
}

// releasableBy reports whether the caller is the holder of this record.
//
// Two keys, and the second exists because the first one rotates. A session id
// is the address a human and a peer use, but it is NOT stable across the
// holder's lifetime: `/clear` issues a new id to the SAME long-lived process,
// and the record keeps the old one. Deciding ownership on the id alone
// therefore refused a lane's own release after a clear — observed on card t791,
// where the refusal named pid 48258 as "a different session" while 48258 was
// the refused process itself (card t951).
//
// The pid key is admitted only where the record SAYS its pid names the owning
// session (PIDSourceSessionOwner). A record without that marker predates the
// anchor and its pid means whatever its writer meant; promoting it to an
// ownership key would re-interpret records already on disk. A pid of 0 on
// either side is "owner unresolvable", and two unknowns are not the same owner
// — matching them would let any session release any unresolvable-owner window,
// a wider hole than the one being closed.
//
// This is deliberately NOT what --force means. force takes a window from a
// DIFFERENT holder and is recorded as such; routing a self-release through it
// would stamp the ledger with a seizure that never happened.
func (l *IntegrationLock) releasableBy(sessionID string, callerOwnerPID int) bool {
	if l == nil {
		return false
	}
	if l.SessionID == sessionID {
		return true
	}
	return l.PIDSource == PIDSourceSessionOwner && l.PID > 0 && l.PID == callerOwnerPID
}

// ReleaseIntegrationLock removes the record when the caller is its holder.
//
// callerOwnerPID is the caller's OWN owning-session pid (session.ResolveOwnerPID),
// or 0 when it could not be resolved. Resolving it is the caller's job for the
// same reason acquire records it rather than inventing one: this package cannot
// see the caller's process ancestry, and a pid guessed here would be this
// process's, which is dead the moment the record is read.
//
// force releases a foreign window, for the same wedged-holder reason acquire
// carries it.
func ReleaseIntegrationLock(projectRoot, sessionID string, callerOwnerPID int, force bool) (released *IntegrationLock, err error) {
	if mutErr := withIntegrationLockMutation(projectRoot, func() error {
		current, readErr := ReadIntegrationLock(projectRoot)
		if readErr != nil && !force {
			return readErr
		}
		if current == nil || !current.Held() {
			return ErrIntegrationLockNotHeld
		}
		if !current.releasableBy(sessionID, callerOwnerPID) && !force {
			return fmt.Errorf("%w: %s (pid %d) holds it", ErrIntegrationLockForeign, current.holderLabel(), current.PID)
		}
		// The queue-aware release (card t1479, REQ-MWQ-006/007): with live
		// tickets queued the record SURVIVES the release — the holder is
		// cleared first (this IS the release), and then the same mutation
		// promotes the first live ticket while the policy is open; under
		// hold the window is left without a holder with the queue intact.
		// Without a queue the pre-queue shape stands: the record file is
		// removed, byte-for-byte as before.
		if len(current.Queue) > 0 {
			// The caller's answer names the holder it released, not the one
			// the promotion put in its place.
			releasedSnapshot := *current
			policy, policyErr := ReadIntegrationWindowPolicy(projectRoot)
			if policyErr != nil {
				return policyErr
			}
			clearHolder(current, WindowClock(), "released by holder")
			RefreshWindow(current, policy, DefaultWindowProcProbe(), WindowClock(), WindowLeaseDuration)
			if err := writeIntegrationLock(integrationLockPath(projectRoot), current); err != nil {
				return err
			}
			released = &releasedSnapshot
			return nil
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

	// The staging path is unique per call. It used to be a fixed sibling name
	// derived from the record's own path, shared by every concurrent writer;
	// now no two writers can ever share one staging file. (The retired literal
	// is deliberately not quoted here: AC-ILA-008 greps this file for it, and a
	// comment naming it would make that check answer about prose rather than
	// about code.)
	//
	// Stated honestly (REQ-ILA-010): under the mutation lock two concurrent
	// writers cannot reach this function at all, so this is defence in depth
	// against a FUTURE caller that writes outside the lock — not a defect
	// anyone observed, and no criterion here claims to have seen a torn record.
	// What is observable is the property: a unique staging name, and no residue
	// left behind on any path.
	tmp, err := os.CreateTemp(filepath.Dir(path), ".integration-lock-*.tmp")
	if err != nil {
		return fmt.Errorf("integration lock: %w", err)
	}
	staging := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(staging)
		return fmt.Errorf("integration lock: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(staging)
		return fmt.Errorf("integration lock: %w", err)
	}
	// CreateTemp opens at 0600; the record is read by every other session's
	// guard, so restore the 0644 the fixed-path write used.
	if err := os.Chmod(staging, 0o644); err != nil {
		_ = os.Remove(staging)
		return fmt.Errorf("integration lock: %w", err)
	}
	if err := os.Rename(staging, path); err != nil {
		_ = os.Remove(staging)
		return fmt.Errorf("integration lock: %w", err)
	}
	return nil
}
