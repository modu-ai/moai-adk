// factory_serial_slot_liveness_test.go — card t1828: an expired lease no
// longer frees the serial slot by itself. The slot frees when the lease
// holder's roster process is gone, and keeps holding while the roster binds
// the owner to a live pid — for a first lease and a re-lease after the
// nominee's own expiry collection alike, and the refusal names the blocking
// card. Measured 2026-10-17 (run tn1p4e): t1825's lease expired at
// 10:57:43Z while its worker (pid 21334) stayed alive and committed at
// 11:22:39Z, and t1697 leased the slot at 10:59:37Z — two serial cards
// genuinely worked at once.
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

// expiredLeaseFixture builds the two-serial-card run the liveness arms read:
// t1 leased to lane-2 with its lease expired under the fixture clock, t2
// assigned to lane-1 with its queue item operator-picked (the shape the
// nominated lease of the lane's own assigned card takes), both worktrees
// isolated, lane-1 stamped.
func expiredLeaseFixture(t *testing.T) (string, *factory.BacklogStore) {
	t.Helper()
	root, store := flSerialPair(t)
	nmSetState(t, store, "t2", factory.BacklogStatePicked)
	nmIsolatedWorktrees(t, "t1", "t2")
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardLeased, OwnerLabel: "lane-2", LeaseHolder: "lane-2",
			LeaseExpiresAt: fcNow.Add(-time.Minute).Format(time.RFC3339Nano), Stage: homestate.CardRun},
		homestate.Card{CardID: "t2", State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardRun},
	)
	nmLaneEnv(t, "lane-1", "")
	return root, store
}

// serialSlotWorkerPID points the roster row of label at pid, so the
// holder-liveness read sees a chosen process without spawning one. pid <= 0
// reads as a dead holder; os.Getpid() as a genuinely live one.
func serialSlotWorkerPID(t *testing.T, root, label string, pid int) {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	if _, err := db.DB.Exec(`UPDATE workers SET pid=? WHERE label=?`, pid, label); err != nil {
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
	serialSlotWorkerPID(t, root, "lane-2", os.Getpid())
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
			LeaseExpiresAt: fcNow.Add(-time.Minute).Format(time.RFC3339Nano), Stage: homestate.CardRun},
		homestate.Card{CardID: "t2", State: homestate.CardLeased, OwnerLabel: "lane-1", LeaseHolder: "lane-1",
			LeaseExpiresAt: fcNow.Add(-time.Minute).Format(time.RFC3339Nano), Stage: homestate.CardRun},
	)
	serialSlotWorkerPID(t, root, "lane-2", os.Getpid())

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

// TestFactoryNextLeasesBehindExpiredLeaseWithDeadHolder — the t1407 no-wedge
// property, pinned on the far side of the t1828 refinement: an expired lease
// whose holder is GONE (no roster row; a non-positive pid) releases the slot,
// so the fleet never wedges behind a dead lane's lazily-collected row.
func TestFactoryNextLeasesBehindExpiredLeaseWithDeadHolder(t *testing.T) {
	t.Run("no roster row", func(t *testing.T) {
		root, _ := expiredLeaseFixture(t)
		serialSlotWorkerRowDropped(t, root, "lane-2")
		serialSlotDeadHolderLeases(t, root)
	})
	t.Run("non-positive pid", func(t *testing.T) {
		root, _ := expiredLeaseFixture(t)
		serialSlotWorkerPID(t, root, "lane-2", -1)
		serialSlotDeadHolderLeases(t, root)
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
