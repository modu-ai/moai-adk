// factory_serial_slot_liveness_test.go — card t1828: an expired lease no
// longer frees the serial slot by itself. The slot frees when the lease
// holder's roster process is gone, and keeps holding while the roster binds
// the owner to a live pid of the SAME incarnation that last heartbeated the
// card (a label re-registered by a new process frees — card-review r1 P2) —
// for a first lease and a re-lease after the nominee's own expiry collection
// alike, and the refusal names the blocking card. Measured 2026-10-17 (run
// tn1p4e): t1825's lease expired at 10:57:43Z while its worker (pid 21334)
// stayed alive and committed at 11:22:39Z, and t1697 leased the slot at
// 10:59:37Z — two serial cards genuinely worked at once.
package cli

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// The two roster registration stamps the liveness arms play with, against
// t1's heartbeat (fcNow-90s, frozen before its lease expired — the measured
// shape). A registration BEFORE the heartbeat is the same incarnation that
// wrote it; one AFTER is a re-registered NEW process holding the label.
const (
	serialSlotIncarnationEarly = "2026-09-26T08:00:00Z"
	serialSlotIncarnationLater = "2026-09-26T09:00:30Z"
)

// expiredLeaseFixture builds the two-serial-card run the liveness arms read:
// t1 leased to lane-2 with its lease expired under the fixture clock and its
// heartbeat frozen before the expiry, t2 assigned to lane-1 with its queue
// item operator-picked (the shape the nominated lease of the lane's own
// assigned card takes), both worktrees isolated, lane-1 stamped.
func expiredLeaseFixture(t *testing.T) (string, *factory.BacklogStore) {
	t.Helper()
	root, store := flSerialPair(t)
	nmSetState(t, store, "t2", factory.BacklogStatePicked)
	nmIsolatedWorktrees(t, "t1", "t2")
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardLeased, OwnerLabel: "lane-2", LeaseHolder: "lane-2",
			LeaseExpiresAt: fcNow.Add(-time.Minute).Format(time.RFC3339Nano),
			HeartbeatAt:    fcNow.Add(-90 * time.Second).Format(time.RFC3339Nano), Stage: homestate.CardRun},
		homestate.Card{CardID: "t2", State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardRun},
	)
	nmLaneEnv(t, "lane-1", "")
	return root, store
}

// serialSlotWorkerPID points the roster row of label at pid and stamps its
// registration, so the holder-liveness read sees a chosen process without
// spawning one. pid <= 0 reads as a dead holder; os.Getpid() as a genuinely
// live one; registeredAt decides which incarnation holds the label.
func serialSlotWorkerPID(t *testing.T, root, label string, pid int, registeredAt string) {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	if _, err := db.DB.Exec(`UPDATE workers SET pid=?, registered_at=? WHERE label=?`, pid, registeredAt, label); err != nil {
		t.Fatalf("set %s pid: %v", label, err)
	}
}

// serialSlotWorkerRowDropped removes label's roster row entirely — the shape
// of a lane whose worker record was never written or was swept.
func serialSlotWorkerRowDropped(t *testing.T, root, label string) {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	if _, err := db.DB.Exec(`DELETE FROM workers WHERE label=?`, label); err != nil {
		t.Fatalf("drop %s roster row: %v", label, err)
	}
}

// TestFactoryNextRefusedBehindExpiredLeaseWithLiveHolder — card t1828: t1's
// lease expired but lane-2's worker process is alive, so t1 still holds the
// serial slot and lane-1's nomination of its own assigned card t2 is refused
// serial-slot (waitable), naming the blocking card and its holder. Before
// the fix the expiry alone freed the slot and the lease went through — the
// measured 10-17 double-lease.
func TestFactoryNextRefusedBehindExpiredLeaseWithLiveHolder(t *testing.T) {
	root, _ := expiredLeaseFixture(t)
	serialSlotWorkerPID(t, root, "lane-2", os.Getpid(), serialSlotIncarnationEarly)
	_, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t2")
	if err == nil {
		t.Fatalf("factory next leased a second serial card behind an expired lease whose holder is alive: stderr=%q", stderr)
	}
	if !strings.Contains(stderr, "refused serial-slot") {
		t.Fatalf("refusal is not serial-slot: %v stderr=%q", err, stderr)
	}
	if !strings.Contains(stderr, "t1 (holder lane-2)") {
		t.Fatalf("refusal does not name the blocking card and holder: %v stderr=%q", err, stderr)
	}
}

// TestFactoryNextReNominationAfterOwnExpiryStillRefused — case (b) of card
// t1828, the re-lease asymmetry observed live (first lease passed at
// 10:59Z, re-lease refused after 11:30Z): the nominee's own card has itself
// cycled through lease expiry and the T27/T28 collection back to `assigned`
// (owner kept, lease cleared), and the lane re-nominates it — the
// assignedHere read, whose ignoreAssigned skips only sibling rows that are
// MERELY assigned. The sibling's expired lease with a live holder is not
// "merely assigned" and keeps holding, so the re-lease is refused
// serial-slot naming the blocker — the two refusals stay symmetric.
func TestFactoryNextReNominationAfterOwnExpiryStillRefused(t *testing.T) {
	root, store := flSerialPair(t)
	nmSetState(t, store, "t2", factory.BacklogStatePicked)
	nmIsolatedWorktrees(t, "t1", "t2")
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardLeased, OwnerLabel: "lane-2", LeaseHolder: "lane-2",
			LeaseExpiresAt: fcNow.Add(-time.Minute).Format(time.RFC3339Nano),
			HeartbeatAt:    fcNow.Add(-90 * time.Second).Format(time.RFC3339Nano), Stage: homestate.CardRun},
		homestate.Card{CardID: "t2", State: homestate.CardLeased, OwnerLabel: "lane-1", LeaseHolder: "lane-1",
			LeaseExpiresAt: fcNow.Add(-time.Minute).Format(time.RFC3339Nano), Stage: homestate.CardRun},
	)
	serialSlotWorkerPID(t, root, "lane-2", os.Getpid(), serialSlotIncarnationEarly)

	// The nominee's own lease expired too; any transition attempt collects
	// it back to `assigned` — owner kept, lease cleared (T27/T28).
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	if _, err := db.Transition(context.Background(), homestate.TransitionRequest{
		RunID: fcRun, CardID: "t2", To: homestate.CardRun,
		ExpectedVersion: fcCard(t, root, "t2").Version, Actor: "test", Now: fcNow,
	}); !errors.Is(err, homestate.ErrLeaseExpired) {
		t.Fatalf("collecting t2's expired lease: err=%v, want ErrLeaseExpired", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close factory: %v", err)
	}
	if c := fcCard(t, root, "t2"); c.State != homestate.CardAssigned || c.OwnerLabel != "lane-1" || c.LeaseExpiresAt != "" {
		t.Fatalf("t2 after collection: state=%s owner=%q lease=%q, want assigned/lane-1/cleared", c.State, c.OwnerLabel, c.LeaseExpiresAt)
	}

	nmLaneEnv(t, "lane-1", "")
	_, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t2")
	if err == nil {
		t.Fatalf("factory next re-leased behind a live holder's expired lease: stderr=%q", stderr)
	}
	if !strings.Contains(stderr, "refused serial-slot") {
		t.Fatalf("refusal is not serial-slot: %v stderr=%q", err, stderr)
	}
	if !strings.Contains(stderr, "t1 (holder lane-2)") {
		t.Fatalf("refusal does not name the blocking card and holder: %v stderr=%q", err, stderr)
	}
}

// TestFactoryNextRefusedBehindCollectedAssignedOfLiveHolder — sync-audit
// cold F1 (card t1828, repair round 3): factoryStageCard does not verify the
// requesting lane owns the cardID, so a FOREIGN lane's stage on an expired
// card triggers the lazy T27/T28 collection to `assigned` (owner and
// heartbeat kept); the walk's ignoreAssigned skip then excluded that row, so
// a lane nominating its own assigned card leased a second serial card while
// the first holder's session was still live. An expiry-collected assigned
// row is not "merely assigned" — its heartbeat names the work session — and
// the same generation-aware liveness reader decides it.
func TestFactoryNextRefusedBehindCollectedAssignedOfLiveHolder(t *testing.T) {
	root, store := flSerialPair(t)
	nmSetState(t, store, "t2", factory.BacklogStatePicked)
	nmIsolatedWorktrees(t, "t1", "t2")
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardLeased, OwnerLabel: "lane-2", LeaseHolder: "lane-2",
			LeaseExpiresAt: fcNow.Add(-time.Minute).Format(time.RFC3339Nano),
			HeartbeatAt:    fcNow.Add(-90 * time.Second).Format(time.RFC3339Nano), Stage: homestate.CardRun},
		homestate.Card{CardID: "t2", State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardRun},
	)
	serialSlotWorkerPID(t, root, "lane-2", os.Getpid(), serialSlotIncarnationEarly)

	// A foreign lane's stage on the expired card collects the lease to
	// `assigned` — owner kept, heartbeat kept (sync-audit F1 leg 1).
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	if _, err := db.Transition(context.Background(), homestate.TransitionRequest{
		RunID: fcRun, CardID: "t1", To: homestate.CardRun,
		ExpectedVersion: fcCard(t, root, "t1").Version, Actor: "lane-13", Now: fcNow,
	}); !errors.Is(err, homestate.ErrLeaseExpired) {
		t.Fatalf("collecting t1's expired lease: err=%v, want ErrLeaseExpired", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close factory: %v", err)
	}
	if c := fcCard(t, root, "t1"); c.State != homestate.CardAssigned || c.OwnerLabel != "lane-2" || c.HeartbeatAt == "" {
		t.Fatalf("t1 after collection: state=%s owner=%q heartbeat=%q, want assigned/lane-2/kept", c.State, c.OwnerLabel, c.HeartbeatAt)
	}

	nmLaneEnv(t, "lane-1", "")
	_, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t2")
	if err == nil {
		t.Fatalf("factory next leased a second serial card behind a collected assigned row whose session is live: stderr=%q", stderr)
	}
	if !strings.Contains(stderr, "refused serial-slot") {
		t.Fatalf("refusal is not serial-slot: %v stderr=%q", err, stderr)
	}
	if !strings.Contains(stderr, "t1 (holder lane-2)") {
		t.Fatalf("refusal does not name the blocking card and holder: %v stderr=%q", err, stderr)
	}
}

// TestFactoryNextLeasesBehindAReassignedCollectedCard — sync-audit round 2
// F4 (card t1828, repair round 4): ReassignCard kept HeartbeatAt when the
// owner changed, so the reassigned row carried the OLD owner's heartbeat and
// the round-3 "started work" discriminator read the not-yet-started row as
// work-started — two expiry-collected serial cards reassigned cross-wise
// wedge each lane's self-nomination, unrecoverably. The reassignment clears
// the heartbeat — the row is a pure assignment for its new owner — and the
// nomination SUCCEEDS.
func TestFactoryNextLeasesBehindAReassignedCollectedCard(t *testing.T) {
	root, store := flSerialPair(t)
	nmSetState(t, store, "t2", factory.BacklogStatePicked)
	nmIsolatedWorktrees(t, "t1", "t2")
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardLeased, OwnerLabel: "lane-2", LeaseHolder: "lane-2",
			LeaseExpiresAt: fcNow.Add(-time.Minute).Format(time.RFC3339Nano),
			HeartbeatAt:    fcNow.Add(-90 * time.Second).Format(time.RFC3339Nano), Stage: homestate.CardRun},
		homestate.Card{CardID: "t2", State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardRun},
	)
	serialSlotWorkerPID(t, root, "lane-2", os.Getpid(), serialSlotIncarnationEarly)

	// Collect t1's expired lease to `assigned` (heartbeat kept) — the F1
	// shape — then reassign the collected card cross-wise to lane-1 through
	// the real reassign path.
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	if _, err := db.Transition(context.Background(), homestate.TransitionRequest{
		RunID: fcRun, CardID: "t1", To: homestate.CardRun,
		ExpectedVersion: fcCard(t, root, "t1").Version, Actor: "lane-13", Now: fcNow,
	}); !errors.Is(err, homestate.ErrLeaseExpired) {
		t.Fatalf("collecting t1's expired lease: err=%v, want ErrLeaseExpired", err)
	}
	if _, err := db.ReassignCard(context.Background(), fcRun, "t1", "lane-1", "leader", fcNow, nil); err != nil {
		t.Fatalf("reassign t1: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close factory: %v", err)
	}
	if c := fcCard(t, root, "t1"); c.State != homestate.CardAssigned || c.OwnerLabel != "lane-1" {
		t.Fatalf("t1 after reassign: state=%s owner=%q, want assigned/lane-1", c.State, c.OwnerLabel)
	}

	// The stale heartbeat names the OLD owner's session, but it rides the row
	// into lane-1's own roster generation and reads as lane-1's live session
	// — the self-unrecoverable wedge this repair closes.
	serialSlotWorkerPID(t, root, "lane-1", os.Getpid(), serialSlotIncarnationEarly)
	nmLaneEnv(t, "lane-1", "")
	if _, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t2"); err != nil {
		t.Fatalf("factory next refused behind a reassigned card carrying its old owner's heartbeat: %v stderr=%q", err, stderr)
	}
	if st, owner := flRow(t, root, "t2"); st != homestate.CardLeased || owner != "lane-1" {
		t.Fatalf("t2 state=%s owner=%q, want leased by lane-1", st, owner)
	}
}

// TestFactoryNextLeasesBehindExpiredLeaseOfReRegisteredHolder — the
// card-review r1 P2 repair of card t1828: factory_slots deletes a dead
// worker's roster row and REASSIGNS the same label to a new process (a lane
// relaunch after reboot is routine), so a live pid behind the label is not
// the process that held the lease. The liveness read is generation-aware —
// the roster row counts only when its registration precedes the card's last
// heartbeat (the incarnation that wrote the heartbeat is the one
// registered). A re-registered label (registered_at AFTER the card's
// heartbeat, pid genuinely alive) must free the slot and the nomination must
// SUCCEED; before the repair the new process's liveness was read as the old
// row's and the slot stayed blocked indefinitely.
func TestFactoryNextLeasesBehindExpiredLeaseOfReRegisteredHolder(t *testing.T) {
	root, _ := expiredLeaseFixture(t)
	serialSlotWorkerPID(t, root, "lane-2", os.Getpid(), serialSlotIncarnationLater)
	serialSlotDeadHolderLeases(t, root)
}

// TestFactoryNextLeasesBehindExpiredLeaseWithDeadHolder — the t1407 no-wedge
// property, pinned on the far side of the t1828 refinement: an expired lease
// whose holder is GONE (no roster row; a non-positive pid; an empty
// heartbeat; a re-registered label) releases the slot, so the fleet never
// wedges behind a dead lane's lazily-collected row.
func TestFactoryNextLeasesBehindExpiredLeaseWithDeadHolder(t *testing.T) {
	t.Run("no roster row", func(t *testing.T) {
		root, _ := expiredLeaseFixture(t)
		serialSlotWorkerRowDropped(t, root, "lane-2")
		serialSlotDeadHolderLeases(t, root)
	})
	t.Run("non-positive pid", func(t *testing.T) {
		root, _ := expiredLeaseFixture(t)
		serialSlotWorkerPID(t, root, "lane-2", -1, serialSlotIncarnationEarly)
		serialSlotDeadHolderLeases(t, root)
	})
	t.Run("empty heartbeat", func(t *testing.T) {
		// A lease-holding row that never carried a heartbeat cannot name its
		// incarnation; leased cards always carry one, so empty frees.
		root, _ := expiredLeaseFixture(t)
		db := fcOpen(t, root)
		if _, err := db.DB.Exec(`UPDATE cards SET heartbeat_at='' WHERE card_id='t1'`); err != nil {
			t.Fatalf("clear t1 heartbeat: %v", err)
		}
		_ = db.Close()
		serialSlotWorkerPID(t, root, "lane-2", os.Getpid(), serialSlotIncarnationEarly)
		serialSlotDeadHolderLeases(t, root)
	})
}

// TestFactoryNextPeekSeesLiveHolderBehindExpiredLease — the reconciliation
// of card t1828 with main's --wait peek (factoryNextMayLease, sync-audit
// absorb F1): the peek's serial reads must take the same generation-aware
// holder-liveness read the lease arms take. Behind a live holder's expired
// lease the peek answers false — or the --wait loop would poll on a "may
// lease" the lease attempt then refuses serial-slot; behind a dead holder it
// answers true, so the loop never wedges on a lazily-collected row either.
func TestFactoryNextPeekSeesLiveHolderBehindExpiredLease(t *testing.T) {
	ctx := context.Background()
	t.Run("nominated, live holder", func(t *testing.T) {
		root, _ := expiredLeaseFixture(t)
		serialSlotWorkerPID(t, root, "lane-2", os.Getpid(), serialSlotIncarnationEarly)
		if factoryNextMayLease(ctx, root, fcRun, "lane-1", "t2", false) {
			t.Fatalf("peek = true behind an expired lease whose holder is alive; want false — the --wait loop would poll a lease the attempt refuses serial-slot")
		}
	})
	t.Run("nominated, dead holder", func(t *testing.T) {
		root, _ := expiredLeaseFixture(t)
		serialSlotWorkerRowDropped(t, root, "lane-2")
		if !factoryNextMayLease(ctx, root, fcRun, "lane-1", "t2", false) {
			t.Fatalf("peek = false behind an expired lease whose holder is gone; want true — the peek must not wedge the wait loop on a dead lane's row")
		}
	})
	t.Run("bare candidate, live holder", func(t *testing.T) {
		root := fcSlotFixture(t, homestate.Card{
			State: homestate.CardLeased, OwnerLabel: "lane-2", LeaseHolder: "lane-2",
			Stage:          homestate.CardRun,
			LeaseExpiresAt: fcNow.Add(-time.Minute).Format(time.RFC3339Nano),
			HeartbeatAt:    fcNow.Add(-90 * time.Second).Format(time.RFC3339Nano),
		}, 1)
		serialSlotWorkerPID(t, root, "lane-2", os.Getpid(), serialSlotIncarnationEarly)
		if factoryNextMayLease(ctx, root, fcRun, "lane-1", "", false) {
			t.Fatalf("bare peek = true while a queued serial candidate sits behind an expired lease whose holder is alive; want false")
		}
	})
	t.Run("bare candidate, dead holder", func(t *testing.T) {
		root := fcSlotFixture(t, homestate.Card{
			State: homestate.CardLeased, OwnerLabel: "lane-2", LeaseHolder: "lane-2",
			Stage:          homestate.CardRun,
			LeaseExpiresAt: fcNow.Add(-time.Minute).Format(time.RFC3339Nano),
			HeartbeatAt:    fcNow.Add(-90 * time.Second).Format(time.RFC3339Nano),
		}, 1)
		serialSlotWorkerRowDropped(t, root, "lane-2")
		if !factoryNextMayLease(ctx, root, fcRun, "lane-1", "", false) {
			t.Fatalf("bare peek = false behind an expired lease whose holder is gone; want true")
		}
	})
}

// serialSlotDeadHolderLeases asserts the lease of t2 goes through behind a
// dead holder and lands leased by lane-1.
func serialSlotDeadHolderLeases(t *testing.T, root string) {
	t.Helper()
	if _, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t2"); err != nil {
		t.Fatalf("factory next refused behind an expired lease whose holder is dead: %v stderr=%q", err, stderr)
	}
	if st, owner := flRow(t, root, "t2"); st != homestate.CardLeased || owner != "lane-1" {
		t.Fatalf("t2 state=%s owner=%q, want leased by lane-1", st, owner)
	}
}
