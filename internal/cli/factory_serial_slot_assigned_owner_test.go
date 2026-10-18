// factory_serial_slot_assigned_owner_test.go — card t1821: a lease-free,
// never-started `assigned` serial row (empty HeartbeatAt) holds the serial
// slot only while its owner lane is ALIVE and is not itself leasing another
// card. Before this card the row held whenever OwnerLabel was set, so dead
// lanes (lane-23/26/29) and busy lanes owning assigned serial cards (t1618,
// t1694, t1697, t1702, t1715) held the slot permanently and
// `moai todo --auto-lane` stopped at `serial-slot: another serial card is in
// flight` (brain ruling, board 10-16, option b). The ownerless free (t1407)
// is unchanged and pinned here as a control.
package cli

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// assignedOwnerFixture builds the run the arms read: queue t1 (candidate) and
// t2 (serial, picked, the pure-assigned sibling) plus t3 (parallelizable, the
// card the sibling's owner may be busy with). t2 is recorded `assigned` to
// lane-2 with no lease and no heartbeat; lane-1 is the nominating lane.
func assignedOwnerFixture(t *testing.T) (string, *factory.BacklogStore) {
	t.Helper()
	root, store := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued, factory.BacklogStateQueued)
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	nmSetState(t, store, "t2", factory.BacklogStatePicked)
	nmIsolatedWorktrees(t, "t1", "t2", "t3")
	fcPlace(t, root, homestate.Card{CardID: "t2", State: homestate.CardAssigned, OwnerLabel: "lane-2", Stage: homestate.CardRun})
	nmLaneEnv(t, "lane-1", "")
	return root, store
}

// assignedOwnerWorkerPID points the roster row of label at pid, so the owner
// liveness read sees a chosen process without spawning one. pid <= 0 reads as
// a dead owner; os.Getpid() as a genuinely live one.
func assignedOwnerWorkerPID(t *testing.T, root, label string, pid int) {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	if _, err := db.DB.Exec(`UPDATE workers SET pid=? WHERE label=?`, pid, label); err != nil {
		t.Fatalf("set %s pid: %v", label, err)
	}
}

// assignedOwnerWorkerDropped removes label's roster row entirely — the shape
// of a lane whose worker record was never written or was swept.
func assignedOwnerWorkerDropped(t *testing.T, root, label string) {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	if _, err := db.DB.Exec(`DELETE FROM workers WHERE label=?`, label); err != nil {
		t.Fatalf("drop %s roster row: %v", label, err)
	}
}

// assignedOwnerLeasing places t3 (parallelizable) leased to lane-2: the owner
// of the pure-assigned t2 is busy with another card. expiresAt decides whether
// that lease is live.
func assignedOwnerLeasing(t *testing.T, root string, expiresAt time.Time) {
	t.Helper()
	fcPlace(t, root, homestate.Card{CardID: "t3", State: homestate.CardLeased, OwnerLabel: "lane-2", LeaseHolder: "lane-2",
		LeaseExpiresAt: expiresAt.Format(time.RFC3339Nano),
		HeartbeatAt:    fcNow.Add(-time.Minute).Format(time.RFC3339Nano), Stage: homestate.CardRun})
}

// assignedOwnerRefused asserts the nomination of t1 is refused serial-slot,
// (the pure-assigned t2 holds the slot).
func assignedOwnerRefused(t *testing.T, root string) {
	t.Helper()
	_, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
	if err == nil {
		t.Fatalf("factory next leased a second serial card behind a held assigned row: stderr=%q", stderr)
	}
	if !strings.Contains(stderr, "refused serial-slot") {
		t.Fatalf("refusal is not serial-slot: %v stderr=%q", err, stderr)
	}
	if fcHasCard(t, root, "t1") {
		t.Fatalf("t1 gained a record row past a serial-slot refusal")
	}
}

// assignedOwnerLeases asserts the nomination of t1 goes through and lands
// leased by lane-1.
func assignedOwnerLeases(t *testing.T, root string) {
	t.Helper()
	if _, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1"); err != nil {
		t.Fatalf("factory next refused behind an assigned row that must not hold the slot: %v stderr=%q", err, stderr)
	}
	if st, owner := flRow(t, root, "t1"); st != homestate.CardLeased || owner != "lane-1" {
		t.Fatalf("t1 state=%s owner=%q, want leased by lane-1", st, owner)
	}
}

// TestFactoryNextPureAssignedRowFreesWhenOwnerGone — the measured defect: the
// owner lane of a pure-assigned serial row is dead (non-positive pid) or has
// no roster row at all, and the row still held the slot permanently.
func TestFactoryNextPureAssignedRowFreesWhenOwnerGone(t *testing.T) {
	t.Run("owner pid not alive", func(t *testing.T) {
		root, _ := assignedOwnerFixture(t)
		assignedOwnerWorkerPID(t, root, "lane-2", -1)
		assignedOwnerLeases(t, root)
	})
	t.Run("owner has no roster row", func(t *testing.T) {
		root, _ := assignedOwnerFixture(t)
		assignedOwnerWorkerDropped(t, root, "lane-2")
		assignedOwnerLeases(t, root)
	})
}

// TestFactoryNextPureAssignedRowHoldsWhileOwnerAliveAndIdle — the control
// arm: a live owner that leases nothing else keeps its pure assignment
// holding the slot (the leader's pre-assignment is real queued work).
func TestFactoryNextPureAssignedRowHoldsWhileOwnerAliveAndIdle(t *testing.T) {
	root, _ := assignedOwnerFixture(t)
	assignedOwnerWorkerPID(t, root, "lane-2", os.Getpid())
	assignedOwnerRefused(t, root)
}

// TestFactoryNextPureAssignedRowFreesWhileOwnerLeasesAnotherCard — a live
// owner that is currently leasing a DIFFERENT card (live lease) is not going
// to start the assigned one soon: the pure assignment stops holding. An
// EXPIRED lease on the other card is not "leasing" and the assignment holds.
func TestFactoryNextPureAssignedRowFreesWhileOwnerLeasesAnotherCard(t *testing.T) {
	t.Run("owner holds a live lease on another card", func(t *testing.T) {
		root, _ := assignedOwnerFixture(t)
		assignedOwnerWorkerPID(t, root, "lane-2", os.Getpid())
		assignedOwnerLeasing(t, root, fcNow.Add(time.Hour))
		assignedOwnerLeases(t, root)
	})
	t.Run("the other lease expired so the owner is idle", func(t *testing.T) {
		root, _ := assignedOwnerFixture(t)
		assignedOwnerWorkerPID(t, root, "lane-2", os.Getpid())
		assignedOwnerLeasing(t, root, fcNow.Add(-time.Minute))
		assignedOwnerRefused(t, root)
	})
}

// TestFactoryNextOwnerlessAssignedRowStillFrees — t1407/t1513: a row with no
// owner holds nothing, whatever the roster says.
func TestFactoryNextOwnerlessAssignedRowStillFrees(t *testing.T) {
	root, _ := assignedOwnerFixture(t)
	db := fcOpen(t, root)
	if _, err := db.DB.Exec(`UPDATE cards SET owner_label='' WHERE card_id='t2'`); err != nil {
		t.Fatalf("clear t2 owner: %v", err)
	}
	_ = db.Close()
	assignedOwnerLeases(t, root)
}

// TestAutoLaneCycleSerialSlotAssignedOwnerDeadOrBusy — the card's surface: the
// auto-lane cycle takes the queued serial candidate when the pure-assigned
// sibling's owner is dead or busy, and still stops at serial-slot when the
// owner is alive and idle.
func TestAutoLaneCycleSerialSlotAssignedOwnerDeadOrBusy(t *testing.T) {
	run := func(t *testing.T, root string) (out, errOut string) {
		t.Helper()
		laneAutoSeams(t)
		laneAutoSeedEvidence(t, root, "t1")
		tick := 0
		sleep, now := laneAutoTickClock(&tick)
		var o, e bytes.Buffer
		if err := runAutoLaneCycle(context.Background(), root, &o, &e, laneAutoOpts(t, 5*time.Minute, sleep, now)); err != nil {
			t.Fatalf("lane cycle: %v (stderr %q)", err, e.String())
		}
		return o.String(), e.String()
	}
	leased := func(t *testing.T, root, out, errOut string) {
		t.Helper()
		if !strings.Contains(out, "t1 stage=") {
			t.Fatalf("the cycle did not take the queued serial candidate:\nout=%s\nstderr=%s", out, errOut)
		}
		nmAssertLeased(t, root, "t1", "lane-1")
	}
	t.Run("owner dead", func(t *testing.T) {
		root, _ := assignedOwnerFixture(t)
		assignedOwnerWorkerPID(t, root, "lane-2", -1)
		out, errOut := run(t, root)
		leased(t, root, out, errOut)
	})
	t.Run("owner busy with another card", func(t *testing.T) {
		root, _ := assignedOwnerFixture(t)
		assignedOwnerWorkerPID(t, root, "lane-2", os.Getpid())
		assignedOwnerLeasing(t, root, fcNow.Add(time.Hour))
		out, errOut := run(t, root)
		leased(t, root, out, errOut)
	})
	t.Run("owner alive and idle still stops at serial-slot", func(t *testing.T) {
		root, _ := assignedOwnerFixture(t)
		assignedOwnerWorkerPID(t, root, "lane-2", os.Getpid())
		out, errOut := run(t, root)
		if !strings.Contains(errOut, "serial-slot") {
			t.Errorf("the serial-slot refusal is not narrated: %q", errOut)
		}
		if strings.Contains(out, "stage=leased") || fcHasCard(t, root, "t1") {
			t.Errorf("a card was leased while a live idle owner's assignment held the slot:\n%s", out)
		}
	})
}

// TestAutoLaneCycleLeasesOwnAssignedBehindForeignAssignedRows — arm (a) of the
// card text, the exact story: lane-1 holds an assigned serial card (the
// t1757 shape: here t1) while the queue holds a further serial candidate (t2)
// and other lanes own pure-assigned serial rows (t3 alive and idle, t4 dead).
// The lane leases its OWN assigned card before it walks the queue order, and
// the foreign assigned rows do not stand in front of it.
func TestAutoLaneCycleLeasesOwnAssignedBehindForeignAssignedRows(t *testing.T) {
	root, store := nmBase(t, factory.BacklogStatePicked, factory.BacklogStateQueued, factory.BacklogStatePicked, factory.BacklogStatePicked)
	for _, id := range []string{"t1", "t2", "t3", "t4"} {
		fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	}
	sdRegisterLane(t, root, "lane-3")
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardRun},
		homestate.Card{CardID: "t3", State: homestate.CardAssigned, OwnerLabel: "lane-3", Stage: homestate.CardRun},
		homestate.Card{CardID: "t4", State: homestate.CardAssigned, OwnerLabel: "lane-2", Stage: homestate.CardRun},
	)
	assignedOwnerWorkerPID(t, root, "lane-3", os.Getpid())
	assignedOwnerWorkerPID(t, root, "lane-2", -1)
	nmLaneEnv(t, "lane-1", "")
	nmIsolatedWorktrees(t, "t1", "t2", "t3", "t4")
	laneAutoSeams(t)
	laneAutoSeedEvidence(t, root, "t1", "t2")

	tick := 0
	sleep, now := laneAutoTickClock(&tick)
	var out, errOut bytes.Buffer
	if err := runAutoLaneCycle(context.Background(), root, &out, &errOut, laneAutoOpts(t, 5*time.Minute, sleep, now)); err != nil {
		t.Fatalf("lane cycle: %v (stderr %q)", err, errOut.String())
	}
	got := out.String()
	t1At := strings.Index(got, "t1 stage=")
	t2At := strings.Index(got, "t2 stage=")
	if t1At < 0 {
		t.Fatalf("the lane's own assigned card was never leased behind foreign assigned rows:\nout=%s\nstderr=%s", got, errOut.String())
	}
	if t2At >= 0 && t2At < t1At {
		t.Errorf("a queued candidate was ranked ahead of the lane's own assigned card:\n%s", got)
	}
	nmAssertLeased(t, root, "t1", "lane-1")
}
