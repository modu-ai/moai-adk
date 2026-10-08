package factory

import (
	"context"
	"errors"
	"testing"
)

// gtdProvenOwner is an operation owner whose readback already proved the
// assignment (the state a recovery pass starts from) and whose apply must
// never run again.
type gtdProvenOwner struct{}

func (gtdProvenOwner) Readback(context.Context, GTDOperation) (bool, error) { return true, nil }

func (gtdProvenOwner) Apply(context.Context, GTDOperation) error {
	return errors.New("apply must not run: the assignment was already proven")
}

// V1 (turn-end gate relay #2, card t1538): a reassignment refused for a valid
// lease must leave the dispatch binding exactly where it was. The binding and
// the owner move land in one factory transaction, so the refusal commits
// neither — the queue rolls the assignment back, and the binding must not be
// left naming a run the queue no longer points at.
func TestRefusedReassignmentLeavesTheBindingOnItsRun(t *testing.T) {
	root, _, card := gtdReconcileFixture(t)
	gtdPlaceRow(t, root, card, "run-old")
	gtdAssignRow(t, root, card, "run-old", "lane-1")
	gtdLeaseRow(t, root, card, "run-old", "lane-1")
	// run-current is the card's current engagement: its T2 wrote the binding last.
	gtdPlaceRow(t, root, card, "run-current")
	gtdAssignRow(t, root, card, "run-current", "lane-9")
	if got := gtdBindingRun(t, root, card); got != "run-current" {
		t.Fatalf("fixture binding = %q, want run-current", got)
	}
	if err := RecordFactoryCardAssignment(root, "run-old", card, "lane-2", ""); err == nil {
		t.Fatal("a reassignment over a valid lease was accepted")
	}
	if got := gtdBindingRun(t, root, card); got != "run-current" {
		t.Fatalf("binding = %q after the refused reassignment, want run-current — a past approval could close the current work", got)
	}
}

// V3: the owner of a dispatch comes from the queue's current-dispatch
// record, never from the runtime assignment row — RecordFactoryCardState
// overwrites that row with whatever owner a state report carries, so after a
// same-run re-dispatch lane-1 -> lane-2 a late lane-1 report would otherwise
// make recovery restore the superseded owner.
func TestRecoveryRestoresDispatchOwnerNotLateStateReportOwner(t *testing.T) {
	root, s, card := gtdReconcileFixture(t)
	gtdPlaceRow(t, root, card, "run-1")
	gtdAssignRow(t, root, card, "run-1", "lane-1")
	if err := RecordFactoryCardAssignment(root, "run-1", card, "lane-1", ""); err != nil {
		t.Fatal(err)
	}
	// Same-run re-dispatch to lane-2: the queue half lands (current dispatch
	// = lane-2), the factory half does not.
	gtdAssignWithSeveredFactoryHalf(t, s, "run-1", card, "lane-2")
	// A late state report from the superseded lane overwrites the runtime
	// assignment row's owner.
	if err := RecordFactoryCardState(root, "run-1", card, "lane-1", "", "picked", "card.state"); err != nil {
		t.Fatal(err)
	}
	op := GTDOperation{OperationID: "gtd-dispatch:run-1:" + card, MissionID: "run-1", Action: GTDActionDispatch, Target: card, SnapshotHash: "snap", ReceiptJSON: []byte(`{}`)}
	if _, err := ExecuteGTDOperation(context.Background(), s, op, gtdProvenOwner{}); err != nil {
		t.Fatalf("recovery failed: %v", err)
	}
	if row := gtdLoadRow(t, root, "run-1", card); row.OwnerLabel != "lane-2" {
		t.Fatalf("factory owner = %q after recovery, want lane-2 (the current dispatch's owner; lane-1 is a late state report's)", row.OwnerLabel)
	}
}
