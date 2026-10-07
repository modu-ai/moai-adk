package factory

// integration_window_ops_test.go — M3 tests for the queue operations
// (card t1479, SPEC-MERGE-WINDOW-QUEUE-001): liveness drops (REQ-MWQ-003),
// promotion with the owner-pid anchor and target copy (REQ-MWQ-006), hold
// suspension (REQ-MWQ-007), and the no-overtake invariants (REQ-MWQ-011).
// Every liveness state is constructed through the probe seam, never by
// arranging real processes.

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// opsClock returns a controllable clock: the returned setter moves it, the
// returned func is what the ops read.
func opsClock() (func(time.Time), func() time.Time) {
	state := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)
	set := func(at time.Time) { state = at }
	read := func() time.Time { return state }
	return set, read
}

// pinWindowClock pins the PRODUCTION clock — WindowClock, the clock the
// production paths (release, acquire, the merge step's refresh) read — at
// the fixture instant for the test's duration (card t1479 r3 F1). A fixture
// that stamps heartbeats from opsClock while the production path reads the
// real clock went red the moment wall time passed the stamp plus
// WaiterHeartbeatWindow; pinning the production clock to the same instant
// the stamps come from is what makes the test's liveness arrangement the
// one the production probe actually sees.
func pinWindowClock(t *testing.T, at time.Time) {
	t.Helper()
	prev := WindowClock
	WindowClock = func() time.Time { return at }
	t.Cleanup(func() { WindowClock = prev })
}

// liveProbe builds a probe from explicit liveness maps.
func liveProbe(owners map[int]bool, waiters map[string]bool) WindowProcProbe {
	return WindowProcProbe{
		OwnerAlive:  func(pid int) bool { return owners[pid] },
		WaiterAlive: func(pid int, start string) bool { return waiters[fmt.Sprintf("%d:%s", pid, start)] },
	}
}

func waiterKey(pid int, start string) string {
	return fmt.Sprintf("%d:%s", pid, start)
}

// baseHolder returns a live holder record with one lease stamp.
func baseHolder(pid int, at time.Time) *IntegrationLock {
	return &IntegrationLock{
		SessionID:      "sess-a",
		SessionName:    "lane-a",
		PID:            pid,
		PIDSource:      PIDSourceSessionOwner,
		Branch:         "develop",
		BranchSource:   BranchSourceConfig,
		Worktree:       "/repo/.claude/worktrees/develop",
		AcquiredAt:     at.Format(time.RFC3339),
		LeaseExpiresAt: StampLease(&IntegrationLock{}, at, IntegrationLeaseDefault).LeaseExpiresAt,
	}
}

// baseTicket returns a queued ticket for sess-b.
func baseTicket(at time.Time) IntegrationTicket {
	return IntegrationTicket{
		SessionID:    "sess-b",
		SessionName:  "lane-b",
		Card:         "t0002",
		OwnerPID:     5001,
		PIDSource:    PIDSourceSessionOwner,
		Branch:       "develop",
		BranchSource: BranchSourceConfig,
		Worktree:     "/repo/.claude/worktrees/develop",
		WaiterPID:    5002,
		WaiterStart:  "12345.678",
		Heartbeat:    at.Format(time.RFC3339),
		EnqueuedAt:   at.Format(time.RFC3339),
	}
}

func openPolicy() IntegrationWindowPolicy { return IntegrationWindowPolicy{Policy: PolicyOpen} }

func TestRefreshWindowPromotesOnRelease(t *testing.T) {
	// AC-MWQ-006 scenario 1: A holds, B waits, A releases — B is the holder,
	// queue [], and B's record carries B's owner pid with the session-owner
	// source and B's integration target.
	setNow, now := opsClock()
	at := now()
	lock := baseHolder(4001, at)
	lock.Queue = []IntegrationTicket{baseTicket(at)}
	lock.SessionID = "sess-a" // the releaser: holder fields will be replaced
	_ = setNow                // the release path reads the clock once

	report := RefreshWindow(lock, openPolicy(), liveProbe(
		map[int]bool{4001: false, 5001: true}, // A's owner gone (release path), B's owner live
		map[string]bool{waiterKey(5002, "12345.678"): true},
	), now(), IntegrationLeaseDefault)
	if report.Promoted == nil || report.Promoted.SessionID != "sess-b" {
		t.Fatalf("B must be promoted on release, report=%+v lock=%+v", report, lock)
	}
	if lock.SessionID != "sess-b" || lock.SessionName != "lane-b" || lock.Card != "t0002" {
		t.Fatalf("holder fields must carry the ticket's identity: %+v", lock)
	}
	if lock.PID != 5001 || lock.PIDSource != PIDSourceSessionOwner {
		t.Fatalf("promoted holder must carry the ticket's owner pid anchor (REQ-MWQ-006): %+v", lock)
	}
	if lock.Branch != "develop" || lock.BranchSource != BranchSourceConfig || lock.Worktree != "/repo/.claude/worktrees/develop" {
		t.Fatalf("promoted holder must carry the ticket's integration target: %+v", lock)
	}
	if len(lock.Queue) != 0 {
		t.Fatalf("queue must be empty after the single ticket is promoted: %+v", lock.Queue)
	}
	// The promoted holder reads live and releasable: its own release succeeds
	// (AC-MWQ-006 scenario 2 is the CLI-level check; here the anchor fields).
	// The liveness read goes through the same seam the queue ops use —
	// Stale() probes real processes, which cannot be arranged in a unit test.
	if holderOwnerGone(lock, liveProbe(
		map[int]bool{4001: false, 5001: true},
		map[string]bool{},
	)) {
		t.Fatalf("a promoted holder with a live owner must not read stale")
	}
}

func TestRefreshWindowDropsByLiveness(t *testing.T) {
	// REQ-MWQ-003: every drop rule names the dropped ticket and the reason,
	// and a dropped ticket leaves the queue.
	setNow, now := opsClock()
	at := now()

	build := func() *IntegrationLock {
		lock := baseHolder(4001, at)
		lock.Queue = []IntegrationTicket{baseTicket(at)}
		return lock
	}
	probeAllLive := liveProbe(
		map[int]bool{4001: true, 5001: true},
		map[string]bool{waiterKey(5002, "12345.678"): true},
	)

	// Waiter gone (id match, but the process died): dropped, reason named.
	report := RefreshWindow(build(), openPolicy(), liveProbe(
		map[int]bool{4001: true, 5001: true},
		map[string]bool{}, // waiter gone
	), now(), IntegrationLeaseDefault)
	if len(report.Dropped) != 1 || !strings.Contains(report.Dropped[0], "lane-b") || !strings.Contains(report.Dropped[0], "waiter gone") {
		t.Fatalf("waiter-gone drop must name the ticket and reason: %+v", report.Dropped)
	}

	// Same pid, different start: NOT this waiter — dropped.
	report = RefreshWindow(build(), openPolicy(), liveProbe(
		map[int]bool{4001: true, 5001: true},
		map[string]bool{waiterKey(5002, "99999.1"): true}, // recycled pid, other start
	), now(), IntegrationLeaseDefault)
	if len(report.Dropped) != 1 || !strings.Contains(report.Dropped[0], "waiter gone") {
		t.Fatalf("a live pid with a different start must drop the ticket: %+v", report.Dropped)
	}

	// Heartbeat at −59s stays; at −61s it drops.
	lock := build()
	lock.Queue[0].Heartbeat = at.Add(-59 * time.Second).Format(time.RFC3339)
	report = RefreshWindow(lock, openPolicy(), probeAllLive, now(), IntegrationLeaseDefault)
	if len(report.Dropped) != 0 {
		t.Fatalf("a −59s heartbeat must survive: %+v", report.Dropped)
	}
	lock.Queue[0].Heartbeat = at.Add(-61 * time.Second).Format(time.RFC3339)
	report = RefreshWindow(lock, openPolicy(), probeAllLive, now(), IntegrationLeaseDefault)
	if len(report.Dropped) != 1 || !strings.Contains(report.Dropped[0], "heartbeat") {
		t.Fatalf("a −61s heartbeat must drop with a named reason: %+v", report.Dropped)
	}

	// Owner gone while the waiter lives: dropped as "owner gone".
	report = RefreshWindow(build(), openPolicy(), liveProbe(
		map[int]bool{4001: true, 5001: false},
		map[string]bool{waiterKey(5002, "12345.678"): true},
	), now(), IntegrationLeaseDefault)
	if len(report.Dropped) != 1 || !strings.Contains(report.Dropped[0], "owner gone") {
		t.Fatalf("owner-gone drop must name the reason: %+v", report.Dropped)
	}
	_ = setNow
}

func TestRefreshWindowPromotesPastStaleHolder(t *testing.T) {
	// AC-MWQ-006 scenario 3: A's owning process gone, B queued — any
	// mutation (here a refresh) promotes B and records A as displaced.
	setNow, now := opsClock()
	at := now()
	lock := baseHolder(4001, at) // A's owner pid 4001
	lock.Queue = []IntegrationTicket{baseTicket(at)}

	report := RefreshWindow(lock, openPolicy(), liveProbe(
		map[int]bool{4001: false, 5001: true},
		map[string]bool{waiterKey(5002, "12345.678"): true},
	), now(), IntegrationLeaseDefault)
	if report.Promoted == nil || lock.SessionID != "sess-b" {
		t.Fatalf("B must be promoted past the stale holder: %+v", report)
	}
	if lock.Displaced == nil || lock.Displaced.SessionID != "sess-a" {
		t.Fatalf("the displaced holder must be recorded: %+v", lock.Displaced)
	}
	if lock.DisplacedReason == "" {
		t.Fatalf("the displacement must name why")
	}
	_ = setNow
}

func TestRefreshWindowPromotesPastExpiredLease(t *testing.T) {
	// AC-MWQ-006 scenario 4: A's lease expired with a live owner and B
	// queued — the refresh promotes B and records A as displaced.
	setNow, now := opsClock()
	at := now()
	lock := baseHolder(4001, at)
	lock.LeaseExpiresAt = at.Add(-time.Minute).Format(time.RFC3339)
	lock.Queue = []IntegrationTicket{baseTicket(at)}

	report := RefreshWindow(lock, openPolicy(), liveProbe(
		map[int]bool{4001: true, 5001: true},
		map[string]bool{waiterKey(5002, "12345.678"): true},
	), now(), IntegrationLeaseDefault)
	if report.Promoted == nil || lock.SessionID != "sess-b" {
		t.Fatalf("an expired lease must promote the first live ticket: %+v", report)
	}
	_ = setNow
}

func TestHoldSuspendsPromotion(t *testing.T) {
	// AC-MWQ-007: under hold, a release leaves no holder and the queue
	// intact; a stale holder is cleared with no successor; promotion resumes
	// on open at that same mutation.
	setNow, now := opsClock()
	at := now()
	probeB := liveProbe(
		map[int]bool{4001: false, 5001: true},
		map[string]bool{waiterKey(5002, "12345.678"): true},
	)

	hold := IntegrationWindowPolicy{Policy: PolicyHold, Reason: "release-cut"}

	// Release under hold: no holder, queue [B].
	lock := baseHolder(4001, at)
	lock.Queue = []IntegrationTicket{baseTicket(at)}
	report := RefreshWindow(lock, hold, probeB, now(), IntegrationLeaseDefault)
	if report.Promoted != nil {
		t.Fatalf("hold must suspend promotion: %+v", report)
	}
	if lock.SessionID != "" {
		t.Fatalf("release under hold must leave the window without a holder: %+v", lock)
	}
	if len(lock.Queue) != 1 {
		t.Fatalf("the queue must stay intact under hold: %+v", lock.Queue)
	}

	// Stale holder under hold: cleared, no successor.
	lock = baseHolder(4001, at)
	lock.Queue = []IntegrationTicket{baseTicket(at)}
	report = RefreshWindow(lock, hold, probeB, now(), IntegrationLeaseDefault)
	if lock.SessionID != "" || report.Promoted != nil || len(lock.Queue) != 1 {
		t.Fatalf("a stale holder under hold is cleared with no successor and the queue intact: %+v", lock)
	}

	// Policy returns to open: B promoted at that same mutation.
	report = RefreshWindow(lock, openPolicy(), probeB, now(), IntegrationLeaseDefault)
	if report.Promoted == nil || lock.SessionID != "sess-b" {
		t.Fatalf("promotion must resume on open in queue order: %+v", report)
	}
	_ = setNow
}

func TestEnqueueAppendsAtTailWithoutOvertaking(t *testing.T) {
	// REQ-MWQ-002/011: B then C enqueue in order; B re-invoking with its
	// waiter alive keeps exactly one ticket at its position.
	setNow, now := opsClock()
	at := now()
	lock := baseHolder(4001, at)
	probeLiveB := liveProbe(
		map[int]bool{4001: true, 5001: true},
		map[string]bool{waiterKey(5002, "12345.678"): true},
	)
	ticketB := baseTicket(at)
	if err := EnqueueTicket(lock, ticketB, probeLiveB, now(), openPolicy()); err != nil {
		t.Fatalf("B's first enqueue must succeed: %v", err)
	}
	if len(lock.Queue) != 1 || lock.Queue[0].SessionID != "sess-b" {
		t.Fatalf("B must sit at position 1: %+v", lock.Queue)
	}
	// B re-invokes: exactly one ticket, still at its position.
	if err := EnqueueTicket(lock, ticketB, probeLiveB, now(), openPolicy()); err != nil {
		t.Fatalf("B's re-invoke must be idempotent: %v", err)
	}
	if len(lock.Queue) != 1 || lock.Queue[0].SessionID != "sess-b" {
		t.Fatalf("a live waiter must not hold two tickets: %+v", lock.Queue)
	}
	// C enqueues behind B.
	ticketC := ticketB
	ticketC.SessionID, ticketC.SessionName, ticketC.OwnerPID, ticketC.WaiterPID = "sess-c", "lane-c", 6001, 6002
	ticketC.WaiterStart = "88888.0"
	probeC := liveProbe(
		map[int]bool{4001: true, 5001: true, 6001: true},
		map[string]bool{waiterKey(5002, "12345.678"): true, waiterKey(6002, "88888.0"): true},
	)
	if err := EnqueueTicket(lock, ticketC, probeC, now(), openPolicy()); err != nil {
		t.Fatalf("C's enqueue must succeed: %v", err)
	}
	if len(lock.Queue) != 2 || lock.Queue[1].SessionID != "sess-c" {
		t.Fatalf("C must sit behind B: %+v", lock.Queue)
	}
	// C is never promoted while B is live.
	report := RefreshWindow(lock, openPolicy(), probeC, now(), IntegrationLeaseDefault)
	if report.Promoted != nil && report.Promoted.SessionID == "sess-c" {
		t.Fatalf("C must never be promoted past a live B: %+v", report)
	}
	_ = setNow
}

func TestPromotedOnwardReleaseWhenBoundElapsed(t *testing.T) {
	// REQ-MWQ-005: a waiter observing its promotion after its bound elapsed
	// releases immediately (promoting onward) and reports that it released.
	setNow, now := opsClock()
	at := now()
	// B IS the holder — the promotion already happened (the fixture of the
	// REQ-MWQ-005 scenario) — and C waits behind.
	lock := baseHolder(4001, at)
	lock.SessionID = "sess-b"
	lock.SessionName = "lane-b"
	ticketC := baseTicket(at)
	ticketC.SessionID, ticketC.SessionName, ticketC.OwnerPID, ticketC.WaiterPID = "sess-c", "lane-c", 6001, 6002
	ticketC.WaiterStart = "88888.0"

	// B's bound elapsed before the promotion is observed. C's heartbeat is
	// fresh AT the observation instant — set BEFORE the queue assignment,
	// since ticketC is a value copy and a post-assignment edit would leave
	// the queued copy three minutes stale (dropped before the promotion).
	bound := at.Add(2 * time.Minute)
	setNow(bound.Add(time.Minute))
	ticketC.Heartbeat = now().Format(time.RFC3339)
	ticketC.EnqueuedAt = ticketC.Heartbeat
	lock.Queue = []IntegrationTicket{ticketC}
	probe := liveProbe(
		map[int]bool{4001: true, 5001: true, 6001: true},
		map[string]bool{waiterKey(5002, "12345.678"): true, waiterKey(6002, "88888.0"): true},
	)
	observed, err := PromotedAfterBound(lock, "sess-b", bound, probe, now(), IntegrationLeaseDefault, openPolicy())
	if err != nil {
		t.Fatalf("the late-promotion release must not error: %v", err)
	}
	if !observed.Released || lock.SessionID != "sess-c" {
		t.Fatalf("B must release onward and C must be promoted: released=%v holder=%q", observed.Released, lock.SessionID)
	}
}
