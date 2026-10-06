package factory

// integration_window_release_test.go — M3 end-to-end record tests (card
// t1479): the release promotes the first live ticket onto the holder record
// in the same mutation (REQ-MWQ-006), and a hold keeps the queue intact
// (REQ-MWQ-007). Real record files under t.TempDir(), probes through the
// seam.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func writeWindowRecord(t *testing.T, root string, lock *IntegrationLock) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, ".moai", "state"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := UpdateIntegrationWindow(root, func(w *IntegrationLock) error {
		*w = *lock
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

// realTicket turns the fixture ticket's identity fields into THIS process's
// real values — pid and process start fingerprint — so the production probe
// reads it live. A unit test that arranges a ticket the production probe
// would drop tests the drop, not the promotion.
func realTicket(at time.Time) IntegrationTicket {
	ticket := baseTicket(at)
	ticket.OwnerPID = os.Getpid()
	ticket.WaiterPID = os.Getpid()
	ticket.WaiterStart = currentProcessFingerprint()
	return ticket
}

// currentProcessFingerprint reads this process's start fingerprint the same
// way the production waiter probe does.
func currentProcessFingerprint() string {
	fp, state := homestate.ProbeProcessIdentity(os.Getpid())
	if state != homestate.ProcessIdentityLive {
		return ""
	}
	return fp
}

func TestReleasePromotesFirstTicket(t *testing.T) {
	// AC-MWQ-006 scenario 1 through the real release path: A releases, B is
	// promoted in the SAME mutation, queue [], lease stamped.
	root := t.TempDir()
	setNow, now := opsClock()
	at := now()
	pinWindowClock(t, at) // F1: the production release path reads WindowClock
	holder := baseHolder(4001, at)
	holder.Queue = []IntegrationTicket{realTicket(at)}
	writeWindowRecord(t, root, holder)

	// A (sess-a) releases. The owner pid 4001 is this process — a real
	// pid the release path's releasableBy check accepts (pid source
	// session-owner matches the recorded owner).
	pid := os.Getpid()
	holder.PID = pid
	holder.Queue[0].OwnerPID = pid
	holder.Queue[0].WaiterPID = pid
	writeWindowRecord(t, root, holder)

	released, err := ReleaseIntegrationLock(root, "sess-a", pid, false)
	if err != nil {
		t.Fatalf("release must succeed: %v", err)
	}
	if released.SessionID != "sess-a" {
		t.Fatalf("the release answer names the holder it released, got %q", released.SessionID)
	}
	lock, err := ReadIntegrationLock(root)
	if err != nil {
		t.Fatal(err)
	}
	if lock.SessionID != "sess-b" || len(lock.Queue) != 0 {
		t.Fatalf("B must be promoted onto the holder record in the release mutation: %+v", lock)
	}
	if lock.LeaseExpiresAt == "" {
		t.Fatalf("the promoted holder must carry a lease stamp (REQ-MWQ-008)")
	}
	if lock.PID != pid || lock.PIDSource != PIDSourceSessionOwner {
		t.Fatalf("the promoted holder must carry the ticket's owner pid anchor: %+v", lock)
	}
	_ = setNow
}

func TestReleaseUnderHoldKeepsQueue(t *testing.T) {
	// AC-MWQ-007: release under hold leaves no holder, queue intact.
	root := t.TempDir()
	_, now := opsClock()
	at := now()
	pinWindowClock(t, at) // F1: the production release path reads WindowClock
	holder := baseHolder(os.Getpid(), at)
	holder.Queue = []IntegrationTicket{realTicket(at)}
	if err := WriteIntegrationWindowPolicy(root, IntegrationWindowPolicy{Policy: PolicyHold, Reason: "release-cut"}); err != nil {
		t.Fatal(err)
	}
	writeWindowRecord(t, root, holder)

	if _, err := ReleaseIntegrationLock(root, "sess-a", os.Getpid(), false); err != nil {
		t.Fatalf("release must succeed: %v", err)
	}
	lock, err := ReadIntegrationLock(root)
	if err != nil {
		t.Fatal(err)
	}
	if lock.SessionID != "" {
		t.Fatalf("release under hold must leave the window without a holder: %+v", lock)
	}
	if len(lock.Queue) != 1 || lock.Queue[0].SessionID != "sess-b" {
		t.Fatalf("the queue must stay intact: %+v", lock.Queue)
	}
}

func TestNoWaitAcquireRefusedBehindQueue(t *testing.T) {
	// REQ-MWQ-011: the window is not granted to a no-wait acquire while a
	// live ticket is queued. The record stays exactly as the refresh left
	// it (a refused acquire writes nothing).
	root := t.TempDir()
	_, now := opsClock()
	at := now()
	pinWindowClock(t, at) // F1: the production acquire refresh reads WindowClock
	pid := os.Getpid()
	lock := baseHolder(pid, at)
	ticket := realTicket(at)
	ticket.OwnerPID, ticket.WaiterPID = pid, pid
	lock.Queue = []IntegrationTicket{ticket}
	writeWindowRecord(t, root, lock)

	before, _ := ReadIntegrationLock(root)
	_, err := AcquireIntegrationWindow(root, IntegrationLock{
		SessionID: "sess-c",
		PID:       pid,
		PIDSource: PIDSourceSessionOwner,
		Branch:    "develop",
	}, false, nil)
	if err == nil || !IsIntegrationLockHeld(err) {
		t.Fatalf("a no-wait acquire behind a queued ticket must be refused: %v", err)
	}
	after, _ := ReadIntegrationLock(root)
	if after.SessionID != before.SessionID || len(after.Queue) != len(before.Queue) {
		t.Fatalf("a refused acquire must not change the record: %+v vs %+v", before, after)
	}
}

func TestAcquireUnderHoldRefusesNamingReason(t *testing.T) {
	// REQ-MWQ-012: acquire without --wait refuses naming the hold's reason.
	root := t.TempDir()
	if err := WriteIntegrationWindowPolicy(root, IntegrationWindowPolicy{Policy: PolicyHold, Reason: "release-cut"}); err != nil {
		t.Fatal(err)
	}
	_, err := AcquireIntegrationWindow(root, IntegrationLock{SessionID: "sess-c", PID: os.Getpid()}, false, nil)
	if err == nil || !IsIntegrationWindowHold(err) {
		t.Fatalf("a no-wait acquire under hold must refuse: %v", err)
	}
	if err != nil && !strings.Contains(err.Error(), "release-cut") {
		t.Fatalf("the refusal must name the reason: %v", err)
	}
}

func TestForceWithQueuePreservesOrder(t *testing.T) {
	// AC-MWQ-011 (last scenario): --force with a non-empty queue — the
	// forcer holds, the queue order is preserved, the displacement recorded.
	root := t.TempDir()
	_, now := opsClock()
	at := now()
	pinWindowClock(t, at) // F1: the force acquire's refresh reads WindowClock
	holder := baseHolder(4001, at)
	ticket := realTicket(at)
	holder.Queue = []IntegrationTicket{ticket}
	// The holder's owner is REALLY gone (pid 4001 will not be probed live by
	// the production probe, but the seam decides here) — force works on a
	// live holder regardless, and the promotion of the queued ticket runs
	// first per REQ-MWQ-006; to force against a LIVE holder the owner must
	// read live, so use this process.
	pid := os.Getpid()
	holder.PID = pid
	ticket.OwnerPID, ticket.WaiterPID = pid, pid
	holder.Queue[0] = ticket
	writeWindowRecord(t, root, holder)

	_, err := AcquireIntegrationWindow(root, IntegrationLock{
		SessionID: "sess-c",
		PID:       pid,
		PIDSource: PIDSourceSessionOwner,
		Branch:    "develop",
	}, true, &AcquireWindowOptions{Probe: WindowProcProbe{
		OwnerAlive:  func(int) bool { return true },
		WaiterAlive: func(int, string) bool { return true },
	}})
	if err != nil {
		t.Fatalf("force must take the window: %v", err)
	}
	lock, _ := ReadIntegrationLock(root)
	if lock.SessionID != "sess-c" {
		t.Fatalf("the forcer must hold: %+v", lock)
	}
	if len(lock.Queue) != 1 || lock.Queue[0].SessionID != "sess-b" {
		t.Fatalf("the queue order must be preserved: %+v", lock.Queue)
	}
	if lock.Displaced == nil || lock.Displaced.SessionID != "sess-a" {
		t.Fatalf("the displacement must be recorded: %+v", lock.Displaced)
	}
}

func TestLeaseDisabledByZeroConfig(t *testing.T) {
	// AC-MWQ-008: a configured duration of zero disables the lease — the
	// holder stays valid past any age on owner liveness alone.
	setNow, now := opsClock()
	at := now()
	holder := baseHolder(4001, at)
	StampLease(holder, at, 0) // duration 0 clears the stamp
	if holder.LeaseExpiresAt != "" {
		t.Fatalf("duration zero must clear the lease stamp")
	}
	setNow(at.Add(90 * 24 * time.Hour))
	if holder.LeaseExpired(now()) {
		t.Fatalf("a disabled lease never reads expired")
	}
}
