package factory

import (
	"context"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// gtdLeaseRow puts the card's run-row under a valid lease held by holder.
func gtdLeaseRow(t *testing.T, root, card, runID, holder string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	exp := time.Now().UTC().Add(time.Hour).Format(time.RFC3339Nano)
	holder = canonicalOwnerLabel(holder)
	if _, err := db.DB.Exec(`UPDATE cards SET state='leased', owner_label=?, lease_holder=?, lease_expires_at=? WHERE run_id=? AND card_id=?`, holder, holder, exp, runID, card); err != nil {
		t.Fatal(err)
	}
}

func gtdLoadRow(t *testing.T, root, runID, card string) homestate.Card {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	row, err := db.LoadCard(context.Background(), runID, card)
	if err != nil {
		t.Fatal(err)
	}
	return row
}

// gtdAssignRow moves the placed row to assigned under owner (the T2 edge).
func gtdAssignRow(t *testing.T, root, card, runID, owner string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	row, err := db.LoadCard(context.Background(), runID, card)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Transition(context.Background(), homestate.TransitionRequest{RunID: runID, CardID: card, To: homestate.CardAssigned, ExpectedVersion: row.Version, Actor: "dispatch", Owner: owner, Now: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
}

// P1-2 (turn-end gate, card t1538): a same-run reassignment must move the
// factory row's owner (and so its version) in the assignment's own critical
// section — otherwise the prior lane's approval still passes a completion
// that lands before reconciliation.
func TestSameRunReassignmentInvalidatesPriorOwnerApproval(t *testing.T) {
	root, _, card := gtdReconcileFixture(t)
	gtdPlaceRow(t, root, card, "run-1")
	gtdAssignRow(t, root, card, "run-1", "worker-1")
	if err := RecordFactoryCardAssignment(root, "run-1", card, "worker-1", ""); err != nil {
		t.Fatal(err)
	}
	before := gtdLoadRow(t, root, "run-1", card)
	if err := RecordFactoryCardAssignment(root, "run-1", card, "worker-2", ""); err != nil {
		t.Fatal(err)
	}
	after := gtdLoadRow(t, root, "run-1", card)
	if after.OwnerLabel != NormalizeOwnerLabel("worker-2") || after.Version <= before.Version {
		t.Fatalf("after reassignment owner=%q version=%d (before owner=%q version=%d) — the prior owner's approval is still valid", after.OwnerLabel, after.Version, before.OwnerLabel, before.Version)
	}
}

// P1-4: an owner replacement never lands on a row that still holds a valid
// lease — the old lane would keep renewing while the new owner cannot
// continue. The assignment is refused whole (it rolls back with the hook).
func TestReassignmentRefusedWhileValidLeaseHeld(t *testing.T) {
	root, s, card := gtdReconcileFixture(t)
	gtdPlaceRow(t, root, card, "run-1")
	gtdAssignRow(t, root, card, "run-1", "worker-1")
	if err := RecordFactoryCardAssignment(root, "run-1", card, "worker-1", ""); err != nil {
		t.Fatal(err)
	}
	gtdLeaseRow(t, root, card, "run-1", "worker-1")
	if err := RecordFactoryCardAssignment(root, "run-1", card, "worker-2", ""); err == nil {
		t.Fatal("reassignment over a valid lease was accepted")
	}
	row := gtdLoadRow(t, root, "run-1", card)
	if row.OwnerLabel != NormalizeOwnerLabel("worker-1") || row.LeaseHolder != NormalizeOwnerLabel("worker-1") {
		t.Fatalf("row owner=%q holder=%q — the refused reassignment still moved the row", row.OwnerLabel, row.LeaseHolder)
	}
	record, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range record.Runtime.Assignments {
		if a.CardID == card && a.OwnerLabel == NormalizeOwnerLabel("worker-2") {
			t.Fatalf("the queue kept the refused assignment: %+v", a)
		}
	}
}

// P1-4 (repair path): reconciling a queue assignment that names another
// owner must not swap the owner of a leased row either.
func TestReconcileDoesNotSwapOwnerUnderValidLease(t *testing.T) {
	root, s, card := gtdReconcileFixture(t)
	ctx := context.Background()
	gtdPlaceRow(t, root, card, "run-1")
	gtdAssignRow(t, root, card, "run-1", "worker-1")
	gtdLeaseRow(t, root, card, "run-1", "worker-1")
	gtdAssignWithSeveredFactoryHalf(t, s, "run-1", card, "worker-2")
	op := GTDOperation{OperationID: "gtd-dispatch:run-1:" + card, MissionID: "run-1", Action: GTDActionDispatch, Target: card, SnapshotHash: "snap", ReceiptJSON: []byte(`{}`)}
	got, err := ExecuteGTDOperation(ctx, s, op, gtdDispatchOwner{store: s, root: root, runID: "run-1", card: card, owner: "worker-2"})
	if err == nil && got.State == "reconciled" {
		t.Fatalf("operation reconciled over a lease held by another owner")
	}
	row := gtdLoadRow(t, root, "run-1", card)
	if row.OwnerLabel != NormalizeOwnerLabel("worker-1") {
		t.Fatalf("owner swapped to %q under a valid lease", row.OwnerLabel)
	}
}

// P1-3: every assignment path records the queue's current-dispatch identity.
// An aborted old-run operation, retried after a direct assign moved the card
// to a newer run, must not drag the binding back.
func TestDirectAssignUpdatesCurrentDispatchSoOldRetryIsSuperseded(t *testing.T) {
	root, s, card := gtdReconcileFixture(t)
	ctx := context.Background()
	gtdPlaceRow(t, root, card, "run-old")
	gtdAssignWithSeveredFactoryHalf(t, s, "run-old", card, "worker-1") // aborted old op: queue half only
	gtdPlaceRow(t, root, card, "run-new")
	gtdAssignRow(t, root, card, "run-new", "worker-2")
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.RecordDispatchBinding(ctx, card, "run-new", time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	// What the `factory assign` surface does under its queue lock.
	if err := s.WithLock(func(l *LockedBacklog) error { return l.RefreshDispatchCurrent(card, "run-new", "worker-2") }); err != nil {
		t.Fatal(err)
	}
	oldOp := GTDOperation{OperationID: "gtd-dispatch:run-old:" + card, MissionID: "run-old", Action: GTDActionDispatch, Target: card, SnapshotHash: "snap", ReceiptJSON: []byte(`{}`)}
	_, _ = ExecuteGTDOperation(ctx, s, oldOp, gtdDispatchOwner{store: s, root: root, runID: "run-old", card: card, owner: "worker-1"})
	if got := gtdBindingRun(t, root, card); got != "run-new" {
		t.Fatalf("binding = %q after retrying the old operation, want run-new", got)
	}
}
