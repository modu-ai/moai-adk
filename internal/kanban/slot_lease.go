// slot_lease.go — the resource slot lease: an atomic acquire/status/release
// record for a named heavy resource (SPEC-RESOURCE-SLOT-LEASE-001, card t607).
//
// Why it exists. Lanes sharing one machine used probe-then-start to avoid
// overlapping heavy runs: look at the process list, start if it is empty. That
// is not mutual exclusion — two sessions can both observe "empty" between the
// probe and the start. An acquire here performs read → decide → write inside
// one cross-process critical section, so of any set of concurrent acquires for
// the same resource exactly one is admitted.
//
// Layout (plan.md §B2):
//
//	<primary>/.moai/state/slot-leases/<resource>.json           the record
//	<primary>/.moai/state/slot-leases/<resource>.mutation.lock  the mutation lock
//	<primary>/.moai/logs/slot-lease-audit.jsonl                 the audit log
//
// The record, the mutation lock, and every name here are separate from the
// integration window (integration_lock.go). Only the lock SUBSTRATE
// (acquireBoardLockImpl, flock on Unix / atomic-create on Windows) and its
// jittered wait budget are shared.
//
// Two lifetimes, as in the integration window: the LEASE is a record that
// outlives the CLI process that wrote it, and its validity is decided by the
// recorded owner's liveness and by the holder's own declared bound; the
// MUTATION LOCK spans one call's critical section and dies with its process.
//
// One deliberate difference from the integration window: a lease past its
// declared bound is taken over without --force even when its owner is alive
// (REQ-RSL-006). A false "free" here costs two overlapping heavy runs, not a
// corrupted tree, and the bound is the holder's own promise — enforcing it is
// not a judgement about someone else's work.
package kanban

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// SlotLeaseDirName names the per-resource record directory under the primary
// checkout's .moai/state.
const SlotLeaseDirName = "slot-leases"

// SlotLeaseAuditFileName names the audit log under the primary's .moai/logs.
const SlotLeaseAuditFileName = "slot-lease-audit.jsonl"

// Displacement reasons recorded when an acquire takes a resource over.
const (
	SlotDisplacedStale   = "stale"
	SlotDisplacedExpired = "expired"
	SlotDisplacedForce   = "force"
)

// slotResourceNameRE is the permitted resource-name alphabet. The name becomes
// a path segment, so it is a trust boundary: lowercase ASCII letters, digits,
// and '-', 1 to 64 characters — no separator, no dot, no traversal.
var slotResourceNameRE = regexp.MustCompile(`^[a-z0-9-]{1,64}$`)

// ErrSlotLeaseHeld is returned when a live, unexpired holder other than the
// caller owns the resource.
var ErrSlotLeaseHeld = errors.New("slot lease: resource held by another session")

// ErrSlotLeaseBusy is returned when the per-resource mutation lock stays
// contended for the whole wait budget.
//
// It is DISTINCT from ErrSlotLeaseHeld and must stay so: held is a statement
// about the board ("another session owns this resource"), busy says a peer was
// mid-mutation and the caller should retry. Reporting one as the other tells a
// lane a false thing.
var ErrSlotLeaseBusy = errors.New("slot lease: record is busy: another process is mutating it")

// ErrSlotLeaseNotHeld is returned when releasing a resource nobody holds.
var ErrSlotLeaseNotHeld = errors.New("slot lease: resource is not held")

// ErrSlotLeaseForeign is returned when a different session's release lacks force.
var ErrSlotLeaseForeign = errors.New("slot lease: resource is held by a different session")

// ErrSlotResourceNameInvalid is returned for a resource name outside the
// permitted character set or length.
var ErrSlotResourceNameInvalid = errors.New("slot lease: invalid resource name")

// ErrSlotLeaseBoundInvalid is returned for a declared maximum duration that is
// zero, negative, or unparseable.
var ErrSlotLeaseBoundInvalid = errors.New("slot lease: invalid declared maximum duration")

// IsSlotLeaseHeld reports whether err is the held sentinel.
func IsSlotLeaseHeld(err error) bool { return errors.Is(err, ErrSlotLeaseHeld) }

// IsSlotLeaseBusy reports whether err is the busy sentinel.
func IsSlotLeaseBusy(err error) bool { return errors.Is(err, ErrSlotLeaseBusy) }

// IsSlotLeaseNotHeld reports whether err is the empty-release sentinel.
func IsSlotLeaseNotHeld(err error) bool { return errors.Is(err, ErrSlotLeaseNotHeld) }

// IsSlotLeaseForeign reports whether err is the foreign-release sentinel.
func IsSlotLeaseForeign(err error) bool { return errors.Is(err, ErrSlotLeaseForeign) }

// IsSlotResourceNameInvalid reports whether err is the invalid-name sentinel.
func IsSlotResourceNameInvalid(err error) bool { return errors.Is(err, ErrSlotResourceNameInvalid) }

// slotLeaseMutationTestHook is a nil-by-default, TEST-ONLY interleaving point
// invoked once between the acquire decision and the write, inside the
// per-resource mutation lock. Same contract as integrationLockMutationTestHook:
// only package kanban can assign it, no non-test file does, and the nil guard
// at the call site keeps production behaviour byte-for-byte unchanged.
var slotLeaseMutationTestHook func()

// SlotLeaseDisplacement records the holder an acquire took the resource from.
type SlotLeaseDisplacement struct {
	SessionID   string `json:"session_id"`
	SessionName string `json:"session_name"`
	PID         int    `json:"pid"`
	AcquiredAt  string `json:"acquired_at"`
	Reason      string `json:"reason"`
	At          string `json:"at"`
}

// SlotLease is the recorded holder of one named resource. SessionName and
// Command carry no omitempty: an omitted value is recorded as an empty string,
// so the record always says whether the caller gave one.
type SlotLease struct {
	Resource    string                 `json:"resource"`
	SessionID   string                 `json:"session_id"`
	SessionName string                 `json:"session_name"`
	PID         int                    `json:"pid"`
	PIDSource   string                 `json:"pid_source"`
	Command     string                 `json:"command"`
	AcquiredAt  string                 `json:"acquired_at"`
	MaxDuration string                 `json:"max_duration"`
	ExpiresAt   string                 `json:"expires_at"`
	Displaced   *SlotLeaseDisplacement `json:"displaced,omitempty"`
}

// SlotLeaseRequest is one acquire. PID is the owning session's pid resolved by
// the caller (0 when unresolvable); it is recorded verbatim, never replaced by
// the acquiring process's own pid, which is dead by the time anyone probes it.
// Now is the decision clock; zero means time.Now().
type SlotLeaseRequest struct {
	Resource    string
	SessionID   string
	SessionName string
	PID         int
	Command     string
	MaxDuration time.Duration
	Force       bool
	Now         time.Time
}

// SlotLeaseAuditEntry is one line of the audit log. Event is the vocabulary
// the tests pin: acquire, refuse, takeover, release (the lease core) and
// guard-deny, allow-self, allow-stale, allow-expired, allow-unheld, fail-open
// (the PreToolUse guard). Reason qualifies takeover (stale|expired|force) and
// fail-open (the uncertainty that was met).
type SlotLeaseAuditEntry struct {
	Time            string `json:"ts"`
	Event           string `json:"event"`
	Reason          string `json:"reason,omitempty"`
	Resource        string `json:"resource,omitempty"`
	SessionID       string `json:"session_id,omitempty"`
	HolderSessionID string `json:"holder_session_id,omitempty"`
	HolderPID       int    `json:"holder_pid,omitempty"`
}

// Held reports whether the record names a holder.
func (l *SlotLease) Held() bool { return l != nil && l.SessionID != "" }

// Stale reports whether the recorded owner process is gone.
//
// A pid <= 0 reads LIVE: it is the recorded form of "owner unresolvable", and
// the conservative reading is that the holder may still be running. Such a
// lease leaves by an explicit release, a recorded --force, or its declared
// bound.
func (l *SlotLease) Stale() bool {
	if !l.Held() || l.PID <= 0 {
		return false
	}
	return !FactoryProcessAlive(l.PID)
}

// Expired reports whether the declared maximum duration has elapsed at now.
// An unparseable expiry reads as NOT expired, for the same conservative reason
// Stale reads an unknown owner as live.
func (l *SlotLease) Expired(now time.Time) bool {
	if !l.Held() {
		return false
	}
	expires, err := time.Parse(time.RFC3339, l.ExpiresAt)
	if err != nil {
		return false
	}
	return !now.Before(expires)
}

// holderLabel prefers the human-facing session name and falls back to the id.
func (l *SlotLease) holderLabel() string {
	switch {
	case l == nil:
		return "unknown"
	case l.SessionName != "":
		return l.SessionName
	case l.SessionID != "":
		return l.SessionID
	default:
		return "unknown"
	}
}

// @MX:ANCHOR: [AUTO] ValidateSlotResourceName is the trust boundary for every path built from a resource name
// @MX:REASON: [AUTO] fan_in >= 3 (ReadSlotLease, AcquireSlotLease, ReleaseSlotLease, the CLI); a laxer check lets a name escape the lease directory
// ValidateSlotResourceName reports whether name is a permitted resource name.
func ValidateSlotResourceName(name string) error {
	if !slotResourceNameRE.MatchString(name) {
		return fmt.Errorf("%w: %q (want 1-64 characters of a-z, 0-9, '-')", ErrSlotResourceNameInvalid, name)
	}
	return nil
}

// ParseSlotLeaseMaxDuration parses a declared maximum duration. A value that
// does not parse, or parses to zero or less, is ErrSlotLeaseBoundInvalid.
func ParseSlotLeaseMaxDuration(s string) (time.Duration, error) {
	d, err := time.ParseDuration(strings.TrimSpace(s))
	if err != nil {
		return 0, fmt.Errorf("%w: %q: %v", ErrSlotLeaseBoundInvalid, s, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("%w: %q must be positive", ErrSlotLeaseBoundInvalid, s)
	}
	return d, nil
}

// @MX:ANCHOR: [AUTO] ResolveSlotLeaseRoot is the single root resolver shared by the CLI and the PreToolUse guard
// @MX:REASON: [AUTO] plan-audit N1 — two code paths resolving the shared root separately is exactly how the CLI and the guard drift apart (CLI writes one tree, guard reads another)
// ResolveSlotLeaseRoot normalizes start (a hook root, CLAUDE_PROJECT_DIR, or a
// cwd) to the parent of its git common directory — the primary checkout — so
// the CLI and the PreToolUse guard resolve the SAME root through ONE function.
//
// It reports an error rather than falling back to start: a caller inside a
// linked worktree that silently used its own tree would read an empty lease
// directory and conclude nobody holds anything. What to do on the error is the
// caller's decision (the guard fails open and says so).
func ResolveSlotLeaseRoot(start string) (string, error) {
	if strings.TrimSpace(start) == "" {
		return "", errors.New("slot lease: cannot resolve the shared root from an empty start directory")
	}
	out, err := exec.Command("git", "-C", start, "rev-parse", "--git-common-dir").Output()
	if err != nil {
		return "", fmt.Errorf("slot lease: %s is not inside a git repository (git rev-parse --git-common-dir: %v)", start, err)
	}
	common := strings.TrimSpace(string(out))
	if common == "" {
		return "", fmt.Errorf("slot lease: git reported no common directory for %s", start)
	}
	if !filepath.IsAbs(common) {
		common = filepath.Join(start, common)
	}
	common = filepath.Clean(common)
	if filepath.Base(common) != ".git" {
		return "", fmt.Errorf("slot lease: git common directory %s is not a checkout's .git; cannot derive the primary root", common)
	}
	return filepath.Dir(common), nil
}

func slotLeaseDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".moai", "state", SlotLeaseDirName)
}

func slotLeasePath(projectRoot, resource string) string {
	return filepath.Join(slotLeaseDir(projectRoot), resource+".json")
}

func slotLeaseMutationPath(projectRoot, resource string) string {
	return filepath.Join(slotLeaseDir(projectRoot), resource+".mutation.lock")
}

// ReadSlotLease returns the recorded holder of resource, or an empty lease
// when no record exists. An unparseable record is an ERROR, never a free
// resource: "unreadable" and "nobody holds it" are different states.
func ReadSlotLease(projectRoot, resource string) (*SlotLease, error) {
	if err := ValidateSlotResourceName(resource); err != nil {
		return nil, err
	}
	path := slotLeasePath(projectRoot, resource)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &SlotLease{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read slot lease: %w", err)
	}
	var lease SlotLease
	if err := json.Unmarshal(data, &lease); err != nil {
		return nil, fmt.Errorf("slot lease at %s is unreadable: %w", path, err)
	}
	return &lease, nil
}

// @MX:ANCHOR: [AUTO] AcquireSlotLease is the atomic acquire every caller (CLI, tests, future wrappers) goes through
// @MX:REASON: [AUTO] REQ-RSL-003 — the read, the decision, and the write MUST stay inside one withSlotLeaseMutation call; moving the read out reintroduces probe-then-start
// AcquireSlotLease records req as the holder of req.Resource.
//
// Decision table (inside the per-resource mutation lock):
//   - no holder, or the caller's own session → acquire (a re-acquire restarts
//     the declared bound);
//   - a foreign holder whose owner is gone → take over, displaced reason stale;
//   - a foreign holder past its declared bound → take over, reason expired;
//   - a live, unexpired foreign holder with Force → take over, reason force;
//   - otherwise → ErrSlotLeaseHeld, record unchanged, one refuse audit line.
func AcquireSlotLease(projectRoot string, req SlotLeaseRequest) (*SlotLease, error) {
	if err := ValidateSlotResourceName(req.Resource); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.SessionID) == "" {
		return nil, errors.New("slot lease: session id is required")
	}
	if req.MaxDuration <= 0 {
		return nil, fmt.Errorf("%w: %s must be positive", ErrSlotLeaseBoundInvalid, req.MaxDuration)
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC().Truncate(time.Second)

	var result *SlotLease
	decide := func() error {
		current, readErr := ReadSlotLease(projectRoot, req.Resource)
		if readErr != nil {
			if !req.Force {
				return readErr
			}
			current = &SlotLease{}
		}
		lease := &SlotLease{
			Resource:    req.Resource,
			SessionID:   req.SessionID,
			SessionName: req.SessionName,
			PID:         req.PID,
			PIDSource:   PIDSourceSessionOwner,
			Command:     req.Command,
			AcquiredAt:  now.Format(time.RFC3339),
			MaxDuration: req.MaxDuration.String(),
			ExpiresAt:   now.Add(req.MaxDuration).Format(time.RFC3339),
		}
		if current.Held() && current.SessionID != req.SessionID {
			reason := ""
			switch {
			case current.Stale():
				reason = SlotDisplacedStale
			case current.Expired(now):
				reason = SlotDisplacedExpired
			case req.Force:
				reason = SlotDisplacedForce
			default:
				appendSlotLeaseAuditBestEffort(projectRoot, SlotLeaseAuditEntry{
					Event: "refuse", Resource: req.Resource, SessionID: req.SessionID,
					HolderSessionID: current.SessionID, HolderPID: current.PID,
				})
				return fmt.Errorf("%w: %s holds %q (session %s, pid %d) since %s, bound ends %s",
					ErrSlotLeaseHeld, current.holderLabel(), req.Resource, current.SessionID, current.PID,
					current.AcquiredAt, current.ExpiresAt)
			}
			lease.Displaced = &SlotLeaseDisplacement{
				SessionID:   current.SessionID,
				SessionName: current.SessionName,
				PID:         current.PID,
				AcquiredAt:  current.AcquiredAt,
				Reason:      reason,
				At:          now.Format(time.RFC3339),
			}
		}
		// Between the decision and the write — the window the mutation lock
		// exists to close. Nil in production.
		if slotLeaseMutationTestHook != nil {
			slotLeaseMutationTestHook()
		}
		if err := writeSlotLease(slotLeasePath(projectRoot, req.Resource), lease); err != nil {
			return err
		}
		entry := SlotLeaseAuditEntry{Event: "acquire", Resource: req.Resource, SessionID: req.SessionID}
		if lease.Displaced != nil {
			entry.Event = "takeover"
			entry.Reason = lease.Displaced.Reason
			entry.HolderSessionID = lease.Displaced.SessionID
			entry.HolderPID = lease.Displaced.PID
		}
		appendSlotLeaseAuditBestEffort(projectRoot, entry)
		result = lease
		return nil
	}
	if err := withSlotLeaseMutation(projectRoot, req.Resource, decide); err != nil {
		return nil, err
	}
	return result, nil
}

// ReleaseSlotLease removes the record when sessionID holds resource. Releasing
// another session's lease needs force; releasing nothing is an error.
func ReleaseSlotLease(projectRoot, resource, sessionID string, force bool) (*SlotLease, error) {
	if err := ValidateSlotResourceName(resource); err != nil {
		return nil, err
	}
	var released *SlotLease
	if err := withSlotLeaseMutation(projectRoot, resource, func() error {
		current, readErr := ReadSlotLease(projectRoot, resource)
		if readErr != nil {
			if !force {
				return readErr
			}
			// A forced release of an unreadable record clears it; there is
			// no holder to report.
			current = &SlotLease{Resource: resource, SessionID: "(unreadable)"}
		}
		if !current.Held() {
			return fmt.Errorf("%w: %q", ErrSlotLeaseNotHeld, resource)
		}
		if current.SessionID != sessionID && !force {
			return fmt.Errorf("%w: %s (session %s, pid %d) holds %q", ErrSlotLeaseForeign, current.holderLabel(), current.SessionID, current.PID, resource)
		}
		if err := os.Remove(slotLeasePath(projectRoot, resource)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("slot lease: %w", err)
		}
		entry := SlotLeaseAuditEntry{Event: "release", Resource: resource, SessionID: sessionID,
			HolderSessionID: current.SessionID, HolderPID: current.PID}
		if current.SessionID != sessionID {
			entry.Reason = SlotDisplacedForce
		}
		appendSlotLeaseAuditBestEffort(projectRoot, entry)
		released = current
		return nil
	}); err != nil {
		return nil, err
	}
	return released, nil
}

// withSlotLeaseMutation runs fn inside the per-resource critical section: at
// most one process at a time is inside a resource's read → decide → write.
// fn performs its OWN read after entering, so a caller serialized behind
// another decides against the state the previous mutation published.
func withSlotLeaseMutation(projectRoot, resource string, fn func() error) error {
	path := slotLeaseMutationPath(projectRoot, resource)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("slot lease: %w", err)
	}
	impl, err := acquireSlotLeaseMutationLock(path)
	if err != nil {
		return err
	}
	defer func() { _ = impl.release() }()
	return fn()
}

// acquireSlotLeaseMutationLock takes the per-resource mutation lock, retrying
// contention within the shared elapsed budget (boardLockWaitBudget) with the
// board's jittered backoff. On budget exhaustion the Windows wedge clear runs
// once (a no-op on Unix), then the caller gets ErrSlotLeaseBusy.
func acquireSlotLeaseMutationLock(path string) (boardLockImpl, error) {
	var lastErr error
	deadline := time.Now().Add(boardLockWaitBudget)
	for attempt := 0; ; attempt++ {
		impl, err := acquireBoardLockImpl(path)
		if err == nil {
			return impl, nil
		}
		if !IsBoardLockHeld(err) {
			return nil, fmt.Errorf("slot lease: taking the mutation lock at %s: %w", path, err)
		}
		lastErr = err
		if !time.Now().Before(deadline) {
			if report, clearErr := clearWedgedSlotLeaseMutationLock(path); clearErr == nil && report != nil && report.Removed {
				if impl, retryErr := acquireBoardLockImpl(path); retryErr == nil {
					return impl, nil
				}
			}
			// %v, never %w: the board sentinel must not leak out of this scope.
			return nil, fmt.Errorf("%w (waited %s): %v", ErrSlotLeaseBusy, boardLockWaitBudget, lastErr)
		}
		time.Sleep(boardLockRetryWait(attempt))
	}
}

// writeSlotLease replaces the record atomically through a unique staging file.
func writeSlotLease(path string, lease *SlotLease) error {
	data, err := json.MarshalIndent(lease, "", "  ")
	if err != nil {
		return fmt.Errorf("slot lease: %w", err)
	}
	data = append(data, '\n')
	tmp, err := os.CreateTemp(filepath.Dir(path), ".slot-lease-*.tmp")
	if err != nil {
		return fmt.Errorf("slot lease: %w", err)
	}
	staging := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(staging)
		return fmt.Errorf("slot lease: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(staging)
		return fmt.Errorf("slot lease: %w", err)
	}
	if err := os.Chmod(staging, 0o644); err != nil {
		_ = os.Remove(staging)
		return fmt.Errorf("slot lease: %w", err)
	}
	if err := os.Rename(staging, path); err != nil {
		_ = os.Remove(staging)
		return fmt.Errorf("slot lease: %w", err)
	}
	return nil
}

// AppendSlotLeaseAudit appends one line to <root>/.moai/logs/slot-lease-audit.jsonl.
// Time is filled when empty.
func AppendSlotLeaseAudit(projectRoot string, entry SlotLeaseAuditEntry) error {
	if strings.TrimSpace(projectRoot) == "" {
		return errors.New("slot lease audit: no project root")
	}
	if entry.Time == "" {
		entry.Time = time.Now().UTC().Format(time.RFC3339)
	}
	line, err := json.Marshal(&entry)
	if err != nil {
		return fmt.Errorf("slot lease audit: %w", err)
	}
	dir := filepath.Join(projectRoot, ".moai", "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("slot lease audit: %w", err)
	}
	f, err := os.OpenFile(filepath.Join(dir, SlotLeaseAuditFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("slot lease audit: %w", err)
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		_ = f.Close()
		return fmt.Errorf("slot lease audit: %w", err)
	}
	return f.Close()
}

// appendSlotLeaseAuditBestEffort writes an audit line from inside a lease
// mutation. The record decision has already been made (or refused); failing
// the whole operation because the log could not be appended would turn a
// logging fault into a lock fault, so the failure is reported on stderr instead
// of being swallowed.
func appendSlotLeaseAuditBestEffort(projectRoot string, entry SlotLeaseAuditEntry) {
	if err := AppendSlotLeaseAudit(projectRoot, entry); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[moai:slot-lease] advisory: %v\n", err)
	}
}
