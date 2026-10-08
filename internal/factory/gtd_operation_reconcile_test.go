package factory

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// gtdReconcileFixture is a project root with a queue card and an open-able
// factory database — the minimal stage for the dispatch-reconciliation
// contract tests (review round-21, card t1538).
func gtdReconcileFixture(t *testing.T) (string, *BacklogStore, string) {
	t.Helper()
	root, s, card := runtimeFixture(t)
	if err := s.Mutate(func(r *BacklogRecord) error {
		for i := range r.Items {
			if r.Items[i].ID == card {
				r.Items[i].State = BacklogStatePicked
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return root, s, card
}

// gtdPlaceRow inserts a factory card row directly (a fixture placement).
func gtdPlaceRow(t *testing.T, root, card, runID string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.RecordPicked(context.Background(), runID, card, homestate.CardFields{}, "dispatch", time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
}

// gtdBindingRun reads the card's recorded dispatch binding run.
func gtdBindingRun(t *testing.T, root, card string) string {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	var run string
	if err := db.DB.QueryRow(`SELECT run_id FROM card_dispatch WHERE card_id=?`, card).Scan(&run); err != nil {
		t.Fatalf("binding read: %v", err)
	}
	return run
}

// gtdDispatchOwner is a GTDOperationOwner whose readback checks the runtime
// assignment record and whose apply re-records it (the idempotent dispatch
// owner shape the CLI closures mirror).
type gtdDispatchOwner struct {
	store *BacklogStore
	root  string
	runID string
	card  string
	owner string
}

func (o gtdDispatchOwner) Readback(context.Context, GTDOperation) (bool, error) {
	record, err := o.store.LoadPure()
	if err != nil {
		return false, err
	}
	want := NormalizeOwnerLabel(o.owner)
	for _, a := range record.Runtime.Assignments {
		if a.RunID == o.runID && a.CardID == o.card && a.OwnerLabel == want {
			return true, nil
		}
	}
	return false, nil
}

func (o gtdDispatchOwner) Apply(context.Context, GTDOperation) error {
	return RecordFactoryCardAssignment(o.root, o.runID, o.card, o.owner, "")
}

// A dispatch op whose Target/MissionID name the supervisor lineage's
// vocabulary (the goal mission's gtd item ref and session) reconciles
// against the identifiers its owner reveals — the assigned card and its run
// (review round-21, card t1538). The repair runs against the revealed pair
// even though the op's own identifiers resolve to no factory record at all.
type revealingDispatchOwner struct {
	GTDOperationOwner
	cardID string
	runID  string
}

func (o revealingDispatchOwner) DispatchReconcileIdentifiers() (string, string) {
	return o.cardID, o.runID
}

func TestGTDDispatchReconcileUsesRevealedIdentifiers(t *testing.T) {
	root, s, card := gtdReconcileFixture(t)
	// The card is factory-engaged from a prior run: the scope sentence
	// applies, and the re-dispatch's factory half (row for run-1) is the
	// part the reconciliation must repair.
	gtdPlaceRow(t, root, card, "run-0")
	if err := RecordFactoryCardAssignment(root, "run-1", card, "worker-7", ""); err != nil {
		t.Fatal(err)
	}
	// No card row for (run-1, card): the reconciliation must repair it.
	op := GTDOperation{OperationID: "goal-dispatch:run-1:" + card, MissionID: "session-1", Action: GTDActionDispatch, Target: "gtd:item-7", SnapshotHash: "snap", ReceiptJSON: []byte(`{}`)}
	owner := revealingDispatchOwner{
		GTDOperationOwner: gtdDispatchOwner{store: s, root: root, runID: "run-1", card: card, owner: "worker-7"},
		cardID:            card,
		runID:             "run-1",
	}
	if _, err := ExecuteGTDOperation(context.Background(), s, op, owner); err != nil {
		t.Fatal(err)
	}

	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.LoadCard(context.Background(), "run-1", card); err != nil {
		t.Fatalf("revealed-identifier reconciliation left no factory row: %v", err)
	}
	if got := gtdBindingRun(t, root, card); got != "run-1" {
		t.Fatalf("binding = %q, want run-1", got)
	}
}

// A dispatch operation that already reconciled must not move the dispatch
// binding when it is replayed: its assignment row persists in the queue
// record, and re-running the reconcile would regress the binding to the
// operation's run — arming the old run's approval against the new run's
// work (review round-21 P1, card t1538).
func TestGTDReconciledDispatchReplayDoesNotRegressBinding(t *testing.T) {
	root, s, card := gtdReconcileFixture(t)
	ctx := context.Background()
	gtdPlaceRow(t, root, card, "run-old")

	oldOp := GTDOperation{OperationID: "gtd-dispatch:run-old:" + card, MissionID: "run-old", Action: GTDActionDispatch, Target: card, SnapshotHash: "snap", ReceiptJSON: []byte(`{}`)}
	owner := gtdDispatchOwner{store: s, root: root, runID: "run-old", card: card, owner: "worker-1"}
	if _, err := ExecuteGTDOperation(ctx, s, oldOp, owner); err != nil {
		t.Fatal(err)
	}

	// The card re-dispatches to a new run: assignment + binding move there.
	if err := RecordFactoryCardAssignment(root, "run-new", card, "worker-2", ""); err != nil {
		t.Fatal(err)
	}
	gtdPlaceRow(t, root, card, "run-new")
	if got := gtdBindingRun(t, root, card); got != "run-new" {
		t.Fatalf("pre-replay binding = %q, want run-new", got)
	}

	// Replaying the OLD (reconciled) operation must leave the binding on
	// the new run.
	if _, err := ExecuteGTDOperation(ctx, s, oldOp, owner); err != nil {
		t.Fatal(err)
	}
	if got := gtdBindingRun(t, root, card); got != "run-new" {
		t.Fatalf("reconciled-op replay regressed the binding to %q, want run-new", got)
	}
}

// The dispatch-binding check must see the factory database that is the
// queue store's state-directory sibling — the layout a home-resolved queue
// (<home>/db/<key>/{todo,factory}) actually produces — instead of reading
// the axis as vacuous because no project .moai sits above the queue
// (review round-21, card t1538).
func TestGTDBindingCurrentResolvesHomeStateSibling(t *testing.T) {
	home := t.TempDir()
	queueDir := filepath.Join(home, "db", "projkey", "todo")
	if err := os.MkdirAll(queueDir, 0o755); err != nil {
		t.Fatal(err)
	}
	store := NewBacklogStore(filepath.Join(queueDir, "backlog.json"))

	db, err := homestate.OpenFactoryPath(filepath.Join(home, "db", "projkey", "factory", "factory.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.RecordDispatchBinding(context.Background(), "t1", "run-old", time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	current, err := gtdDispatchBindingCurrent(context.Background(), store, "t1", "run-new")
	if err != nil {
		t.Fatal(err)
	}
	if current {
		t.Fatal("home-state sibling factory db invisible: stale binding read as current")
	}
}

// The partial-failure repair restores the assigned state with the
// authoritative assignment's owner — a row parked at picked with no owner
// is not the dispatch the operation read back (review round-21, card
// t1538).
func TestGTDDispatchRepairRestoresAssignedOwner(t *testing.T) {
	root, s, card := gtdReconcileFixture(t)
	if err := RecordFactoryCardAssignment(root, "run-1", card, "worker-9", ""); err != nil {
		t.Fatal(err)
	}
	// A stale picked row from a prior cycle: the row exists, the mirror's
	// assignment half never landed.
	gtdPlaceRow(t, root, card, "run-1")

	op := GTDOperation{OperationID: "gtd-dispatch:run-1:" + card, MissionID: "run-1", Action: GTDActionDispatch, Target: card, SnapshotHash: "snap", ReceiptJSON: []byte(`{}`)}
	owner := gtdDispatchOwner{store: s, root: root, runID: "run-1", card: card, owner: "worker-9"}
	if _, err := ExecuteGTDOperation(context.Background(), s, op, owner); err != nil {
		t.Fatal(err)
	}

	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	row, err := db.LoadCard(context.Background(), "run-1", card)
	if err != nil {
		t.Fatal(err)
	}
	if row.State != homestate.CardAssigned {
		t.Fatalf("repaired row state = %q, want assigned", row.State)
	}
	if row.OwnerLabel != NormalizeOwnerLabel("worker-9") {
		t.Fatalf("repaired row owner = %q, want the authoritative assignment's owner", row.OwnerLabel)
	}
}

// The assignment rolls BACK when its binding hook fails (review round-22
// P1, card t1538): a dispatch that cannot bind must not leave the record
// pointing at the new run — that partial state let the completion gate
// close the card on the previous run's approval.
func TestGTDAssignmentRollsBackWhenBindingHookFails(t *testing.T) {
	root, s, card := gtdReconcileFixture(t)
	gtdPlaceRow(t, root, card, "run-0") // the card is factory-engaged: the hook attempts the binding

	p, err := homestate.FactoryDBPath(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(p, p+".preserved"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(p, 0o700); err != nil {
		t.Fatal(err)
	}
	assignErr := RecordFactoryCardAssignment(root, "run-1", card, "worker-1", "")
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(p+".preserved", p); err != nil {
		t.Fatal(err)
	}
	if assignErr == nil {
		t.Fatal("the binding failure did not fail the assignment")
	}
	rec, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range rec.Runtime.Assignments {
		if a.RunID == "run-1" && a.CardID == card {
			t.Fatalf("the assignment survived the failed binding: %+v", a)
		}
	}
}

// A dispatch op stuck at invoking (its assignment applied, its completion
// never verified) must not drag the binding back to its run when replayed
// after the card moved on — its run is no longer the card's latest
// assignment (review round-22 P1, card t1538).
type failingApplyOwner struct {
	gtdDispatchOwner
}

func (o failingApplyOwner) Apply(ctx context.Context, op GTDOperation) error {
	if err := o.gtdDispatchOwner.Apply(ctx, op); err != nil {
		return err
	}
	return errors.New("injected mid-flight failure")
}

func TestGTDSupersededInvokingOpDoesNotRegressBinding(t *testing.T) {
	root, s, card := gtdReconcileFixture(t)
	ctx := context.Background()
	gtdPlaceRow(t, root, card, "run-old")
	op := GTDOperation{OperationID: "gtd-dispatch:run-old:" + card, MissionID: "run-old", Action: GTDActionDispatch, Target: card, SnapshotHash: "snap", ReceiptJSON: []byte(`{}`)}
	owner := failingApplyOwner{gtdDispatchOwner{store: s, root: root, runID: "run-old", card: card, owner: "worker-1"}}
	if _, err := ExecuteGTDOperation(ctx, s, op, owner); err == nil {
		t.Fatal("the injected mid-flight failure did not fail the dispatch")
	}

	// A new run takes the card over: assignment, row, and binding.
	if err := RecordFactoryCardAssignment(root, "run-new", card, "worker-2", ""); err != nil {
		t.Fatal(err)
	}
	gtdPlaceRow(t, root, card, "run-new")

	// Replaying the stuck old operation leaves the binding on the new run.
	if _, err := ExecuteGTDOperation(ctx, s, op, owner); err != nil {
		t.Fatal(err)
	}
	if got := gtdBindingRun(t, root, card); got != "run-new" {
		t.Fatalf("superseded-op replay regressed the binding to %q, want run-new", got)
	}
}

// A row parked at picked with no owner is the mirror's failed T2 — the
// dispatch reconciliation must not read it as current factory state; the
// repair restores the assignment half (review round-22, card t1538).
func TestGTDBindingCurrentRejectsPickedOwnerlessRow(t *testing.T) {
	root, s, card := gtdReconcileFixture(t)
	gtdPlaceRow(t, root, card, "run-1") // T1 landed; the T2 half never did
	if err := RecordFactoryCardAssignment(root, "run-1", card, "worker-9", ""); err != nil {
		t.Fatal(err)
	}
	current, err := gtdDispatchBindingCurrent(context.Background(), s, card, "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if current {
		t.Fatal("a picked, ownerless row read as reconciled factory state")
	}
	// The full operation flow repairs the half-landed row.
	op := GTDOperation{OperationID: "gtd-dispatch:run-1:" + card, MissionID: "run-1", Action: GTDActionDispatch, Target: card, SnapshotHash: "snap", ReceiptJSON: []byte(`{}`)}
	owner := gtdDispatchOwner{store: s, root: root, runID: "run-1", card: card, owner: "worker-9"}
	if _, err := ExecuteGTDOperation(context.Background(), s, op, owner); err != nil {
		t.Fatal(err)
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	row, err := db.LoadCard(context.Background(), "run-1", card)
	if err != nil {
		t.Fatal(err)
	}
	if row.State != homestate.CardAssigned || row.OwnerLabel != NormalizeOwnerLabel("worker-9") {
		t.Fatalf("repaired row = %s/%q, want assigned/worker-9", row.State, row.OwnerLabel)
	}
}
