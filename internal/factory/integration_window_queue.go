// integration_window_queue.go — the merge-window queue's data model and its
// serialized mutation helper (card t1479, SPEC-MERGE-WINDOW-QUEUE-001).
//
// The window record itself stays where it always was
// (integration_lock.go): the holder fields are unchanged, and everything this
// file adds rides beside them as additive optional fields. What this file
// contributes is the QUEUE layer on top:
//
//   - UpdateIntegrationWindow — one serialized read-modify-write of the
//     record, reusing the mutation section this package already borrows for
//     acquire and release. Every queue mutation (enqueue, drop, promote,
//     withdraw, lease renewal, status's pre-print refresh) must run inside
//     it, because REQ-MWQ-002 decides enqueue order inside "the same
//     serialized record mutation that already guards acquire and release".
//
//   - the lease stamp/expiry helpers (REQ-MWQ-008) — the expiry is a string
//     field on the record, stamped by acquire and promotion and renewed by
//     every window verb the holder invokes.
//
//   - the policy sibling record (REQ-MWQ-012) — `open` by absence, written
//     beside the window record under the same state directory.
//
// The liveness probes the queue's drop rules consult (REQ-MWQ-003) are
// package-level seams here, NOT inline probes: a test constructs the exact
// liveness state a criterion describes (a dead waiter with a live owner, a
// live pid with a recycled start time) instead of arranging real processes
// to die. Production leaves both seams on their default implementations,
// which call the same probes the holder's own Stale() and the fingerprint
// reader use.
package factory

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// IntegrationLeaseDefault is the window lease duration REQ-MWQ-008 ships: 30
// minutes. M0's measurement (.moai/reports/t1479/m0-window-duration.md,
// median 534.7ms) found the in-window section sub-second, so the
// shorten-only rule lowers nothing — this default stays until a later
// measurement legitimately tightens it.
const IntegrationLeaseDefault = 30 * time.Minute

// Liveness values for tickets (leader decision Q14, REQ-MWQ-003): a waiting
// process refreshes its heartbeat every WaiterHeartbeatInterval, and any
// mutation drops a ticket whose heartbeat is older than WaiterHeartbeatWindow.
// M0 may only tighten these; the measurement did not, so they stay.
const (
	WaiterHeartbeatInterval = 15 * time.Second
	WaiterHeartbeatWindow   = 60 * time.Second
)

// IntegrationWindowPolicyFileName names the policy record beside the window
// record (REQ-MWQ-012). Its stem is distinct from the record's and from the
// mutation artifact's, for the same glob-safety reason the mutation file
// names its own.
const IntegrationWindowPolicyFileName = "integration-window-policy.json"

// PolicyOpen and PolicyHold are the two window policies. A missing policy
// record reads as Open; a malformed one is an error, not an open window —
// a policy gate that silently opened on corruption would convert a leader's
// hold into a queue that promotes right past it.
const (
	PolicyOpen = "open"
	PolicyHold = "hold"
)

// IntegrationWindowPolicy is the persisted window policy (REQ-MWQ-012).
type IntegrationWindowPolicy struct {
	Policy string `json:"policy"`
	Reason string `json:"reason,omitempty"`
	SetBy  string `json:"set_by,omitempty"`
	SetAt  string `json:"set_at,omitempty"`
}

// integrationWindowPolicyPath resolves the policy record's location under
// the project's state directory — the same directory the window record and
// the mutation artifact share.
func integrationWindowPolicyPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".moai", "state", IntegrationWindowPolicyFileName)
}

// ReadIntegrationWindowPolicy returns the persisted policy, or Open when no
// record exists. An unparseable record is an error rather than an open
// window: silently opening past a corrupted hold would put two lanes inside
// the window on the strength of a typo.
func ReadIntegrationWindowPolicy(projectRoot string) (IntegrationWindowPolicy, error) {
	data, err := os.ReadFile(integrationWindowPolicyPath(projectRoot))
	if errors.Is(err, os.ErrNotExist) {
		return IntegrationWindowPolicy{Policy: PolicyOpen}, nil
	}
	if err != nil {
		return IntegrationWindowPolicy{}, fmt.Errorf("read integration window policy: %w", err)
	}
	var policy IntegrationWindowPolicy
	if err := json.Unmarshal(data, &policy); err != nil {
		return IntegrationWindowPolicy{}, fmt.Errorf("integration window policy at %s is unreadable: %w", integrationWindowPolicyPath(projectRoot), err)
	}
	if policy.Policy == "" {
		policy.Policy = PolicyOpen
	}
	return policy, nil
}

// WriteIntegrationWindowPolicy replaces the policy record atomically, in the
// mutation section — a policy flip that lands while another process is
// deciding against the old policy is exactly the race the section exists to
// close. The gate-level lane-role refusal lives with the policy verb
// (internal/cli), not here: this package has no session identity to check.
func WriteIntegrationWindowPolicy(projectRoot string, policy IntegrationWindowPolicy) error {
	path := integrationWindowPolicyPath(projectRoot)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("integration window policy: %w", err)
	}
	var writeErr error
	if mutErr := withIntegrationLockMutation(projectRoot, func() error {
		writeErr = atomicWriteFile(path, policy)
		return writeErr
	}); mutErr != nil {
		return mutErr
	}
	return writeErr
}

// atomicWriteFile replaces path with a JSON rendering of v, atomically: the
// staging-file discipline is writeIntegrationLock's, shared here rather than
// duplicated (the guard reads the policy on a hot path too).
func atomicWriteFile(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("integration window policy: %w", err)
	}
	data = append(data, '\n')
	tmp, err := os.CreateTemp(filepath.Dir(path), ".integration-window-*.tmp")
	if err != nil {
		return fmt.Errorf("integration window policy: %w", err)
	}
	staging := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(staging)
		return fmt.Errorf("integration window policy: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(staging)
		return fmt.Errorf("integration window policy: %w", err)
	}
	if err := os.Chmod(staging, 0o644); err != nil {
		_ = os.Remove(staging)
		return fmt.Errorf("integration window policy: %w", err)
	}
	if err := os.Rename(staging, path); err != nil {
		_ = os.Remove(staging)
		return fmt.Errorf("integration window policy: %w", err)
	}
	return nil
}

// UpdateIntegrationWindow runs mutate inside the record's mutation section,
// handing it the freshly read record and writing the result back when mutate
// returns nil. Every queue mutation routes through this one helper — the
// read INSIDE the section is the property that makes REQ-MWQ-002's
// same-mutation ordering real rather than aspirational.
//
// A nil mutate is allowed and means "rewrite the record unchanged": the M1
// legacy-compatibility path uses it to prove the round-trip keeps the
// holder and adds no queue key.
func UpdateIntegrationWindow(projectRoot string, mutate func(w *IntegrationLock) error) error {
	path := integrationLockPath(projectRoot)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("integration lock: %w", err)
	}
	return withIntegrationLockMutation(projectRoot, func() error {
		lock, err := ReadIntegrationLock(projectRoot)
		if err != nil {
			return err
		}
		if err := mutate(lock); err != nil {
			return err
		}
		return writeIntegrationLock(path, lock)
	})
}

// WindowClock is the clock the window verbs read. Package-level for the same
// test-geometry reason the probe seams are: the bound/heartbeat/lease
// criteria are stated against instants, and a test pins the clock instead of
// waiting on wall time.
var WindowClock = func() time.Time { return time.Now().UTC() }

// StampLease writes a lease expiry onto the holder record (REQ-MWQ-008): the
// expiry is now + duration, and a duration of zero (or less) DISABLES the
// lease by clearing any existing stamp — validity then falls back to
// owning-session liveness alone, exactly as LeaseExpired states. Mutates and
// returns the same record for chaining.
func StampLease(lock *IntegrationLock, now time.Time, duration time.Duration) *IntegrationLock {
	if duration <= 0 {
		lock.LeaseExpiresAt = ""
		return lock
	}
	lock.LeaseExpiresAt = now.Add(duration).Format(time.RFC3339)
	return lock
}

// LeaseExpired reports whether the record's lease stamp has lapsed at the
// given instant (or was never written). An absent stamp never reads
// expired: it names either a legacy record or a disabled lease, and both
// decide validity by owning-session liveness alone (REQ-MWQ-008, spec.md §D).
func (l *IntegrationLock) LeaseExpired(at time.Time) bool {
	if l == nil || l.LeaseExpiresAt == "" {
		return false
	}
	expires, err := time.Parse(time.RFC3339, l.LeaseExpiresAt)
	if err != nil {
		// An unparseable stamp is indeterminate, and the window's liveness
		// asymmetry reads indeterminate as LIVE: a corrupt stamp must not
		// evict a merging holder.
		return false
	}
	return at.After(expires)
}

// WindowProcProbe is the liveness seam the queue's drop rules consult
// (REQ-MWQ-003). OwnerAlive mirrors the holder's own probe
// (FactoryProcessAlive). WaiterAlive matches the waiter on id AND start
// time — the two-field identity tickets record — and treats an unprobeable
// process as live, the same asymmetry Stale() states for the holder.
type WindowProcProbe struct {
	OwnerAlive  func(pid int) bool
	WaiterAlive func(pid int, start string) bool
}

// DefaultWindowProcProbe is the production probe: the holder's own
// liveness function, and the process-identity fingerprint reader for
// waiters.
func DefaultWindowProcProbe() WindowProcProbe {
	return WindowProcProbe{
		OwnerAlive:  FactoryProcessAlive,
		WaiterAlive: waiterProcessAlive,
	}
}

// waiterProcessAlive reports whether (pid, start) names a live waiter. The
// start value a ticket records is homestate's process fingerprint — the
// kernel start instant the same reader uses for profile leases — so the
// match is on id AND start exactly as REQ-MWQ-003 states: a recycled pid
// with a different start instant is NOT this waiter. An unprobeable
// process reads LIVE, the asymmetry the holder's own Stale() states: an
// unprobeable waiter must not cost a lane its queue position.
func waiterProcessAlive(pid int, start string) bool {
	if pid <= 0 {
		return false
	}
	fp, state := homestate.ProbeProcessIdentity(pid)
	if state != homestate.ProcessIdentityLive {
		return false
	}
	if start == "" {
		// A ticket without a recorded start cannot be matched on the pair;
		// a live process with this pid is the best available answer and
		// reads live, as the asymmetry requires.
		return true
	}
	return fp == start
}
