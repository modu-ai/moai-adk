// slot_lease.go — the resource slot lease: an atomic acquire/status/release
// record for a named heavy resource (SPEC-RESOURCE-SLOT-LEASE-001, card t607).
//
// M1 STATE: this file is the API surface only. Every operation returns
// errSlotLeaseNotImplemented and every predicate answers its zero value, so the
// M1 tests compile and fail on behaviour rather than on missing symbols. M2
// replaces the bodies; the names, the record schema, and the on-disk layout
// declared here are the decisions M1 pins.
//
// Layout (decided in plan.md §B2, pinned by slot_lease_test.go):
//
//	<primary>/.moai/state/slot-leases/<resource>.json           the record
//	<primary>/.moai/state/slot-leases/<resource>.mutation.lock  the mutation lock
//	<primary>/.moai/logs/slot-lease-audit.jsonl                 the audit log
//
// The record, the mutation lock, and every name here are separate from the
// integration window (integration_lock.go). Only the lock SUBSTRATE
// (acquireBoardLockImpl) is shared.
package kanban

import (
	"errors"
	"time"
)

// SlotLeaseDirName names the per-resource record directory under the primary
// checkout's .moai/state.
const SlotLeaseDirName = "slot-leases"

// Displacement reasons recorded when an acquire takes a resource over.
const (
	SlotDisplacedStale   = "stale"
	SlotDisplacedExpired = "expired"
	SlotDisplacedForce   = "force"
)

// errSlotLeaseNotImplemented is the M1 stub answer. It is deliberately none of
// the exported sentinels, so no predicate below can report true for it.
var errSlotLeaseNotImplemented = errors.New("slot lease: not implemented (M1 stub)")

// ErrSlotLeaseHeld is returned when a live, unexpired holder other than the
// caller owns the resource.
var ErrSlotLeaseHeld = errors.New("slot lease: resource held by another session")

// ErrSlotLeaseBusy is returned when the per-resource mutation lock stays
// contended for the whole wait budget. Distinct from ErrSlotLeaseHeld.
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
// per-resource mutation lock. Same contract as integrationLockMutationTestHook.
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
// the caller (0 when unresolvable); it is recorded verbatim. Now is the
// decision clock; zero means time.Now().
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

// Held reports whether the record names a holder.
func (l *SlotLease) Held() bool { return false }

// Stale reports whether the recorded owner process is gone (pid <= 0 reads live).
func (l *SlotLease) Stale() bool { return false }

// Expired reports whether the declared maximum duration has elapsed at now.
func (l *SlotLease) Expired(now time.Time) bool { return false }

// ValidateSlotResourceName reports whether name is a permitted resource name.
// @MX:TODO: [AUTO] M1 stub — implemented in M2 (AC-RSL-009).
func ValidateSlotResourceName(name string) error { return errSlotLeaseNotImplemented }

// ParseSlotLeaseMaxDuration parses a declared maximum duration.
func ParseSlotLeaseMaxDuration(s string) (time.Duration, error) {
	return 0, errSlotLeaseNotImplemented
}

// ResolveSlotLeaseRoot normalizes start (a hook root, CLAUDE_PROJECT_DIR, or a
// cwd) to the parent of its git common directory — the primary checkout — so
// the CLI and the PreToolUse guard resolve the SAME root through ONE function.
// @MX:TODO: [AUTO] M1 stub — implemented in M2 (AC-RSL-016, plan-audit N1).
func ResolveSlotLeaseRoot(start string) (string, error) {
	return "", errSlotLeaseNotImplemented
}

// ReadSlotLease returns the recorded holder of resource, or an empty lease.
// @MX:TODO: [AUTO] M1 stub — implemented in M2.
func ReadSlotLease(projectRoot, resource string) (*SlotLease, error) {
	return nil, errSlotLeaseNotImplemented
}

// AcquireSlotLease records req as the holder of req.Resource.
// @MX:TODO: [AUTO] M1 stub — implemented in M2 (AC-RSL-001..007).
func AcquireSlotLease(projectRoot string, req SlotLeaseRequest) (*SlotLease, error) {
	if slotLeaseMutationTestHook != nil {
		slotLeaseMutationTestHook()
	}
	return nil, errSlotLeaseNotImplemented
}

// ReleaseSlotLease removes the record when sessionID holds resource.
// @MX:TODO: [AUTO] M1 stub — implemented in M2 (AC-RSL-008).
func ReleaseSlotLease(projectRoot, resource, sessionID string, force bool) (*SlotLease, error) {
	return nil, errSlotLeaseNotImplemented
}
