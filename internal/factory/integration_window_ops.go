// integration_window_ops.go — the merge-window queue's operations (card
// t1479, SPEC-MERGE-WINDOW-QUEUE-001): the liveness refresh every queue
// mutation applies (REQ-MWQ-003's drops, REQ-MWQ-006/007's promotion), the
// enqueue, and the late-promotion release (REQ-MWQ-005).
//
// Every function here mutates a record the caller already holds INSIDE the
// mutation section — none of them takes the section itself, so a caller that
// composes them with its own read and write composes ONE serialized
// decision, which is what REQ-MWQ-002's same-mutation ordering and
// REQ-MWQ-005's same-mutation bound/promotion tie require.
package factory

import (
	"fmt"
	"time"
)

// WindowReport names what one refresh did, so the mutating command can print
// each dropped ticket and the reason (REQ-MWQ-003) and each promotion and
// displacement (REQ-MWQ-006) rather than doing any of it silently.
type WindowReport struct {
	// Dropped carries one entry per dropped ticket: "lane-b: waiter gone".
	Dropped []string
	// Promoted is the ticket promoted to the holder, when promotion happened.
	Promoted *IntegrationTicket
	// Displaced names the holder the promotion displaced, when it did.
	Displaced bool
}

// holderOwnerGone reports whether the record's owning-session process is
// gone. A pid of 0 (owner unresolvable) reads LIVE — the asymmetry Stale()
// states — and so does an unprobeable pid.
func holderOwnerGone(lock *IntegrationLock, probe WindowProcProbe) bool {
	if !lock.Held() {
		return false
	}
	if lock.PID <= 0 {
		return false
	}
	return !probe.OwnerAlive(lock.PID)
}

// ticketOwnerGone reports whether a ticket's owning-session process is gone.
// A ticket's owner pid is resolved at enqueue exactly as a holder's is, so
// the same pid-0 and unprobeable asymmetries apply.
func ticketOwnerGone(ticket IntegrationTicket, probe WindowProcProbe) bool {
	if ticket.OwnerPID <= 0 {
		return false
	}
	return !probe.OwnerAlive(ticket.OwnerPID)
}

// refreshQueue applies REQ-MWQ-003's three drop rules to the queue, naming
// each dropped ticket and reason. A dropped ticket leaves the queue: a still
// running waiter observes its own disappearance at its next poll and exits
// non-zero (REQ-MWQ-003's scenario 5 lives in the waiter loop).
func refreshQueue(lock *IntegrationLock, probe WindowProcProbe, now time.Time) []string {
	var dropped []string
	kept := lock.Queue[:0]
	for _, ticket := range lock.Queue {
		switch {
		case ticketOwnerGone(ticket, probe):
			dropped = append(dropped, fmt.Sprintf("%s: owner gone", ticketLabel(ticket)))
		case !probe.WaiterAlive(ticket.WaiterPID, ticket.WaiterStart):
			dropped = append(dropped, fmt.Sprintf("%s: waiter gone", ticketLabel(ticket)))
		case heartbeatStale(ticket, now):
			dropped = append(dropped, fmt.Sprintf("%s: heartbeat stale", ticketLabel(ticket)))
		default:
			kept = append(kept, ticket)
		}
	}
	lock.Queue = kept
	return dropped
}

// ticketLabel renders a ticket the way a dropped-ticket line names it: the
// human-facing name when recorded, else the session id.
func ticketLabel(ticket IntegrationTicket) string {
	if ticket.SessionName != "" {
		return ticket.SessionName
	}
	return ticket.SessionID
}

// heartbeatStale reports whether the ticket's heartbeat is older than
// WaiterHeartbeatWindow (REQ-MWQ-003).
func heartbeatStale(ticket IntegrationTicket, now time.Time) bool {
	beat, err := time.Parse(time.RFC3339, ticket.Heartbeat)
	if err != nil {
		// An unparseable heartbeat is indeterminate, and the window's
		// asymmetry reads indeterminate as live: a corrupt stamp must not
		// evict a waiting lane.
		return false
	}
	return now.Sub(beat) > WaiterHeartbeatWindow
}

// promoteFirst promotes the queue's first ticket onto the holder record
// (REQ-MWQ-006): the ticket's session identity, owner pid with its
// session-owner source, card, and integration target all copy through, the
// lease stamps, and the displaced holder is recorded so a later status can
// show who was displaced. The promoted holder's liveness, self-release, and
// target-ownership checks then behave exactly as a directly acquired
// holder's do — the fields are the anchor's, not a copy of some of them.
func (r *WindowReport) promoteFirst(lock *IntegrationLock, now time.Time, lease time.Duration) {
	ticket := lock.Queue[0]
	displaced := *lock
	if lock.Held() {
		lock.Displaced = &displaced
		lock.DisplacedReason = fmt.Sprintf("displaced %s when %s was promoted at %s", displaced.SessionID, ticket.SessionID, now.Format(time.RFC3339))
		r.Displaced = true
	}
	lock.SessionID = ticket.SessionID
	lock.SessionName = ticket.SessionName
	lock.Card = ticket.Card
	lock.PID = ticket.OwnerPID
	lock.PIDSource = ticket.PIDSource
	lock.Branch = ticket.Branch
	lock.BranchSource = ticket.BranchSource
	lock.Worktree = ticket.Worktree
	lock.AcquiredAt = now.Format(time.RFC3339)
	StampLease(lock, now, lease)
	lock.Queue = lock.Queue[1:]
	promoted := ticket
	r.Promoted = &promoted
}

// RefreshWindow is the refresh every queue mutation applies before deciding
// (REQ-MWQ-003 drops, REQ-MWQ-006/007 promotion): drop the dead tickets,
// then — while the policy is open and the holder is gone, stale, or past
// its lease — promote the first surviving ticket in queue order. Under
// hold, promotion is suspended entirely: a release leaves the window
// without a holder and the queue intact, and a stale holder is cleared
// with no successor (REQ-MWQ-007).
func RefreshWindow(lock *IntegrationLock, policy IntegrationWindowPolicy, probe WindowProcProbe, now time.Time, lease time.Duration) WindowReport {
	report := WindowReport{Dropped: refreshQueue(lock, probe, now)}
	if !lock.Held() {
		if policy.Policy == PolicyOpen && len(lock.Queue) > 0 {
			report.promoteFirst(lock, now, lease)
		}
		return report
	}
	// The holder is present. It is promotable-past when its owner is gone
	// or its lease has lapsed; both read as the stale takeover today's
	// acquire already performs, extended to stamp the queue's successor.
	stale := holderOwnerGone(lock, probe) || lock.LeaseExpired(now)
	if !stale {
		return report
	}
	if policy.Policy == PolicyHold {
		// REQ-MWQ-007: clear the stale holder with no successor, queue
		// intact — the displacement recorded, never silent (P2-9).
		clearHolder(lock, now, "stale holder cleared under hold")
		return report
	}
	if len(lock.Queue) > 0 {
		report.promoteFirst(lock, now, lease)
		return report
	}
	// No ticket to promote: the stale holder is cleared exactly as a
	// takeover would clear it (the caller's own acquire then takes the
	// free window) — with the displacement recorded.
	clearHolder(lock, now, "stale holder cleared")
	return report
}

// clearHolder empties the holder fields without touching the queue,
// RECORDING the displaced holder first (card-review r1 P2-9): the pre-queue
// stale takeover returned the replaced record to its caller, so the
// information existed; a clear that erased it silently lost who was
// displaced and why. A caller that wants the pre-queue shape (no displaced
// key on a takeover it reports itself) passes record=false.
func clearHolder(lock *IntegrationLock, now time.Time, why string) {
	if lock.Held() {
		displaced := *lock
		lock.Displaced = &displaced
		lock.DisplacedReason = fmt.Sprintf("%s at %s", why, now.Format(time.RFC3339))
	}
	lock.SessionID = ""
	lock.SessionName = ""
	lock.Card = ""
	lock.PID = 0
	lock.PIDSource = ""
	lock.Branch = ""
	lock.BranchSource = ""
	lock.Worktree = ""
	lock.AcquiredAt = ""
	lock.LeaseExpiresAt = ""
}

// EnqueueTicket appends one ticket at the tail of the queue inside the
// caller's mutation (REQ-MWQ-002), after applying the liveness refresh UNDER
// THE REAL POLICY — a hardcoded open here let the wait path's own mutations
// promote tickets through a hold (card-review r1 P1-2). The order is
// decided inside the same serialized record mutation as every other queue
// decision. A live ticket for the same session id is idempotent: a
// re-invoking waiter keeps exactly one ticket at its position (REQ-MWQ-011,
// "shall not hold two tickets for one session"). The enqueue identity —
// EnqueuedAt and the first Heartbeat — is stamped HERE from the mutation's
// clock, never left to the caller to remember (card-review r1 P2-5).
func EnqueueTicket(lock *IntegrationLock, ticket IntegrationTicket, probe WindowProcProbe, now time.Time, policy IntegrationWindowPolicy) error {
	RefreshWindow(lock, policy, probe, now, WindowLeaseDuration)
	ticket.EnqueuedAt = now.Format(time.RFC3339)
	ticket.Heartbeat = ticket.EnqueuedAt
	for i, existing := range lock.Queue {
		if existing.SessionID == ticket.SessionID {
			if probe.WaiterAlive(existing.WaiterPID, existing.WaiterStart) {
				// Refresh the heartbeat in place: the caller's waiter is the
				// same lane, still waiting.
				lock.Queue[i].Heartbeat = now.Format(time.RFC3339)
				return nil
			}
			// A stale entry for this session is replaced at its position.
			lock.Queue[i] = ticket
			return nil
		}
	}
	lock.Queue = append(lock.Queue, ticket)
	return nil
}

// WithdrawTicket removes the caller's ticket (REQ-MWQ-004's bound elapse).
func WithdrawTicket(lock *IntegrationLock, sessionID string) {
	kept := lock.Queue[:0]
	for _, ticket := range lock.Queue {
		if ticket.SessionID != sessionID {
			kept = append(kept, ticket)
		}
	}
	lock.Queue = kept
}

// TicketPosition returns the caller's queue position (1-based), 0 when not
// queued.
func TicketPosition(lock *IntegrationLock, sessionID string) int {
	for i, ticket := range lock.Queue {
		if ticket.SessionID == sessionID {
			return i + 1
		}
	}
	return 0
}

// RefreshIntegrationWindowAt is the one-mutation refresh a reporting verb
// (status) applies before it prints (REQ-MWQ-009): read, apply the liveness
// drops and the promotion, write back, and hand the caller the report so the
// mutating command can name each dropped ticket (REQ-MWQ-003). Absent
// record, nothing to do — a free window with no queue is not a mutation.
func RefreshIntegrationWindowAt(projectRoot string) (WindowReport, error) {
	var report WindowReport
	if err := UpdateIntegrationWindow(projectRoot, func(w *IntegrationLock) error {
		policy, policyErr := ReadIntegrationWindowPolicy(projectRoot)
		if policyErr != nil {
			return policyErr
		}
		report = RefreshWindow(w, policy, DefaultWindowProcProbe(), WindowClock(), WindowLeaseDuration)
		return nil
	}); err != nil {
		return WindowReport{}, err
	}
	return report, nil
}

// LatePromotion is what PromotedAfterBound decided.
type LatePromotion struct {
	// Released is true when the caller was promoted past its bound and has
	// released the window onward.
	Released bool
	// BoundElapsed is true when the caller's bound had elapsed before the
	// promotion was observed — the reason for the release.
	BoundElapsed bool
}

// PromotedAfterBound is REQ-MWQ-005's second half, decided inside the
// caller's mutation: when the session was promoted as the holder but its
// bound had already elapsed at that same observation, the promotion is
// handed onward immediately — the record refreshes (dropping this holder and
// promoting the next live ticket) — and the caller reports that it released.
// The same-mutation decision is the point: neither the promotion nor the
// bound outcome may depend on a later read.
func PromotedAfterBound(lock *IntegrationLock, sessionID string, boundAt time.Time, probe WindowProcProbe, now time.Time, lease time.Duration, policy IntegrationWindowPolicy) (LatePromotion, error) {
	if lock.Held() && lock.SessionID == sessionID {
		if now.After(boundAt) {
			// Promoted past the bound: release onward at once — under the
			// real policy, so a hold suspends the onward promotion too.
			clearHolder(lock, now, "bound elapsed, released onward")
			report := RefreshWindow(lock, policy, probe, now, lease)
			_ = report
			return LatePromotion{Released: true, BoundElapsed: true}, nil
		}
		return LatePromotion{BoundElapsed: false}, nil
	}
	// Not the holder (still queued, dropped, or bound-withdrawn elsewhere).
	pos := TicketPosition(lock, sessionID)
	if pos == 0 {
		return LatePromotion{}, fmt.Errorf("integration window: ticket for %s is not in the queue", sessionID)
	}
	return LatePromotion{}, nil
}
