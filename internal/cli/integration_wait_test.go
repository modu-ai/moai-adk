package cli

// integration_wait_test.go — M3 CLI-level tests for the acquire --wait loop
// (card t1479): promotion ends the wait (AC-MWQ-002 scenario 1's B), a
// bound that elapses first withdraws the ticket and exits non-zero naming
// the holder, position, and bound (REQ-MWQ-004), and a dropped ticket exits
// non-zero without re-enqueueing (REQ-MWQ-003 scenario 5). The poll interval
// is shortened and the clock is the factory's WindowClock seam — no test
// sleeps its production interval.

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
)

func waitTestRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	t.Setenv("MOAI_FACTORY_ROLE", "")
	return root
}

func waitHolder(t *testing.T, root, sessionID string) {
	t.Helper()
	_, err := factory.AcquireIntegrationWindow(root, factory.IntegrationLock{
		SessionID: sessionID,
		PID:       os.Getpid(),
		PIDSource: factory.PIDSourceSessionOwner,
		Branch:    "develop",
	}, false, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWaitLoopPromotedWhenHolderReleases(t *testing.T) {
	// AC-MWQ-002 scenario 1: B waits behind A; A releases; B is promoted and
	// the loop returns success. The clock is pinned BEFORE the acquire so
	// the lease stamps against the same fixed instant — a stamp taken at the
	// real wall clock and then judged at 09:00 would read expired.
	root := waitTestRoot(t)
	oldInterval := integrationWaitPollInterval
	integrationWaitPollInterval = 10 * time.Millisecond
	t.Cleanup(func() { integrationWaitPollInterval = oldInterval })
	oldClock := factory.WindowClock
	clockAt := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)
	factory.WindowClock = func() time.Time { return clockAt }
	t.Cleanup(func() { factory.WindowClock = oldClock })
	waitHolder(t, root, "sess-a")

	done := make(chan error, 1)
	go func() {
		done <- integrationWaitInQueue(root, "sess-b", factory.IntegrationTicket{
			SessionID: "sess-b", SessionName: "lane-b", OwnerPID: os.Getpid(),
			WaiterPID: os.Getpid(),
		}, 60*time.Minute, nil)
	}()
	// Give the waiter a moment to enqueue, then release.
	deadline := time.Now().Add(5 * time.Second)
	for {
		lock, _ := factory.ReadIntegrationLock(root)
		if len(lock.Queue) > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("the waiter never enqueued")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if _, err := factory.ReleaseIntegrationLock(root, "sess-a", os.Getpid(), false); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("the waiter must succeed once promoted: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("the waiter did not return after the release")
	}
	lock, _ := factory.ReadIntegrationLock(root)
	if lock.SessionID != "sess-b" {
		t.Fatalf("B must hold the window after promotion: %+v", lock)
	}
}

func TestWaitLoopTimesOutNamingHolderAndPosition(t *testing.T) {
	// REQ-MWQ-004: the bound elapses before promotion — exit non-zero naming
	// the holder, the last queue position, and the bound, and the ticket is
	// gone.
	root := waitTestRoot(t)
	oldInterval := integrationWaitPollInterval
	integrationWaitPollInterval = 5 * time.Millisecond
	t.Cleanup(func() { integrationWaitPollInterval = oldInterval })
	oldClock := factory.WindowClock
	clockAt := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)
	factory.WindowClock = func() time.Time { return clockAt }
	t.Cleanup(func() { factory.WindowClock = oldClock })
	waitHolder(t, root, "sess-a")

	done := make(chan error, 1)
	go func() {
		done <- integrationWaitInQueue(root, "sess-b", factory.IntegrationTicket{
			SessionID: "sess-b", OwnerPID: os.Getpid(), WaiterPID: os.Getpid(),
		}, 2*time.Minute, nil)
	}()
	// Advance the clock past the bound while the waiter polls.
	go func() {
		time.Sleep(100 * time.Millisecond)
		clockAt = clockAt.Add(3 * time.Minute)
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatalf("a lapsed bound must exit non-zero")
		}
		for _, want := range []string{"sess-a", "position 1", "2m0s"} {
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("the timeout must name %q: %v", want, err)
			}
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("the waiter did not time out")
	}
	lock, _ := factory.ReadIntegrationLock(root)
	if factory.TicketPosition(lock, "sess-b") != 0 {
		t.Fatalf("the timed-out ticket must be withdrawn: %+v", lock.Queue)
	}
}

func TestWaitLoopExitsWhenTicketDropped(t *testing.T) {
	// REQ-MWQ-003 scenario 5: the waiter finds its own ticket dropped — it
	// exits non-zero naming the reason and does not re-enqueue itself.
	root := waitTestRoot(t)
	oldInterval := integrationWaitPollInterval
	integrationWaitPollInterval = 10 * time.Millisecond
	t.Cleanup(func() { integrationWaitPollInterval = oldInterval })
	oldClock := factory.WindowClock
	clockAt := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)
	factory.WindowClock = func() time.Time { return clockAt }
	t.Cleanup(func() { factory.WindowClock = oldClock })
	waitHolder(t, root, "sess-a")

	done := make(chan error, 1)
	go func() {
		// A waiter whose OWNER is a dead pid: any mutation drops its ticket.
		done <- integrationWaitInQueue(root, "sess-b", factory.IntegrationTicket{
			SessionID: "sess-b", OwnerPID: 1 << 20, WaiterPID: os.Getpid(),
		}, 60*time.Minute, nil)
	}()
	// The drop is the next queue mutation's job (REQ-MWQ-003): any mutation
	// — here a third lane's status — refreshes liveness and drops the dead
	// owner's ticket, naming it in that command's output.
	enqueueDeadline := time.Now().Add(5 * time.Second)
	for {
		if lock, _ := factory.ReadIntegrationLock(root); factory.TicketPosition(lock, "sess-b") > 0 {
			break
		}
		if time.Now().After(enqueueDeadline) {
			t.Fatalf("the waiter never enqueued")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if _, err := factory.RefreshIntegrationWindowAt(root); err != nil {
		t.Fatal(err)
	}
	if dropped, _ := factory.RefreshIntegrationWindowAt(root); len(dropped.Dropped) != 0 {
		t.Fatalf("the drop happened once; the second refresh must find nothing: %+v", dropped.Dropped)
	}
	select {
	case err := <-done:
		if err == nil {
			t.Fatalf("a dropped ticket must exit non-zero")
		}
		if !strings.Contains(err.Error(), "dropped") || !strings.Contains(err.Error(), "owner gone") {
			t.Fatalf("the drop must name the reason: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("the waiter did not observe the drop")
	}
	lock, _ := factory.ReadIntegrationLock(root)
	if factory.TicketPosition(lock, "sess-b") != 0 {
		t.Fatalf("a dropped waiter must not re-enqueue itself: %+v", lock.Queue)
	}
}

func TestParseAcquireWait(t *testing.T) {
	if _, requested, _ := parseAcquireWait(""); requested {
		t.Fatalf("absent --wait is not a wait")
	}
	bound, requested, err := parseAcquireWait("true")
	if err != nil || !requested || bound != AcquireWaitDefaultBound {
		t.Fatalf("bare --wait must be the 60m default: bound=%v requested=%v err=%v", bound, requested, err)
	}
	bound, _, err = parseAcquireWait("2m")
	if err != nil || bound != 2*time.Minute {
		t.Fatalf("--wait=2m must parse to two minutes: %v %v", bound, err)
	}
	if _, _, err := parseAcquireWait("banana"); err == nil {
		t.Fatalf("a non-duration --wait value must be refused")
	}
}
