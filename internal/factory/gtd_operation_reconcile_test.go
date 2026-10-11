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

// gtdCurrentDispatchRun reads the queue's own current-dispatch record for a
// card — the identity the reconciliation adjudicates on (review round-23).
func gtdCurrentDispatchRun(t *testing.T, s *BacklogStore, card string) (string, bool) {
	t.Helper()
	record, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	return gtdRecordDispatchCurrent(record, card)
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

	current, err := gtdDispatchBindingCurrent(context.Background(), store, "t1", "run-new", "")
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
	current, err := gtdDispatchBindingCurrent(context.Background(), s, card, "run-1", NormalizeOwnerLabel("worker-9"))
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

// A→B→A re-dispatch: the binding follows the LAST dispatch, not a row
// timestamp. Each re-dispatch's IfEngaged hook re-points the binding
// (upsert), so after A→B→A the binding names run-a again and run-a's
// in-flight repair proceeds against its picked, ownerless row
// (review round-22 cont., card t1538).
func TestGTDRedispatchRetakesTheBinding(t *testing.T) {
	root, s, card := gtdReconcileFixture(t)
	gtdPlaceRow(t, root, card, "run-a") // picked/ownerless row: the repair's target
	for _, run := range []string{"run-a", "run-b", "run-a"} {
		if err := RecordFactoryCardAssignment(root, run, card, "worker-1", ""); err != nil {
			t.Fatal(err)
		}
		if run == "run-b" {
			if got := gtdBindingRun(t, root, card); got != "run-b" {
				t.Fatalf("mid-sequence binding = %q, want run-b", got)
			}
		}
	}
	if got := gtdBindingRun(t, root, card); got != "run-a" {
		t.Fatalf("post-sequence binding = %q, want run-a", got)
	}
	op := GTDOperation{OperationID: "gtd-dispatch:run-a:" + card, MissionID: "run-a", Action: GTDActionDispatch, Target: card, SnapshotHash: "snap", ReceiptJSON: []byte(`{}`)}
	owner := gtdDispatchOwner{store: s, root: root, runID: "run-a", card: card, owner: "worker-1"}
	if _, err := ExecuteGTDOperation(context.Background(), s, op, owner); err != nil {
		t.Fatal(err)
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	row, err := db.LoadCard(context.Background(), "run-a", card)
	if err != nil {
		t.Fatal(err)
	}
	if row.State != homestate.CardAssigned || row.OwnerLabel != NormalizeOwnerLabel("worker-1") {
		t.Fatalf("re-dispatched run repaired to %s/%q, want assigned/worker-1", row.State, row.OwnerLabel)
	}
}

// A stale run's late completion report must not move the engagement: the
// binding AND the queue's current-dispatch identity are dispatch records —
// only a dispatch's assignment write records them; a state report never
// does. After run-new takes the card, run-old's RecordFactoryCardState
// report leaves both on run-new, even though the report inserts its own
// assignment row (review round-22 cont. + round-23, card t1538).
func TestGTDStateReportNeverMovesTheBinding(t *testing.T) {
	root, s, card := gtdReconcileFixture(t)
	gtdPlaceRow(t, root, card, "run-new") // the card is factory-engaged: the hook binds
	if err := RecordFactoryCardAssignment(root, "run-new", card, "worker-2", ""); err != nil {
		t.Fatal(err)
	}
	if got := gtdBindingRun(t, root, card); got != "run-new" {
		t.Fatalf("pre-report binding = %q, want run-new", got)
	}
	if got, ok := gtdCurrentDispatchRun(t, s, card); !ok || got != "run-new" {
		t.Fatalf("pre-report current dispatch = %q (ok=%v), want run-new", got, ok)
	}
	// The old run's state report arrives AFTER the new dispatch and inserts
	// its own row — a report writes assignment data only, so neither the
	// binding nor the queue's engagement identity moves.
	if err := RecordFactoryCardState(root, "run-old", card, "worker-1", "", "merged", "card.merged"); err != nil {
		t.Fatal(err)
	}
	if got := gtdBindingRun(t, root, card); got != "run-new" {
		t.Fatalf("late state report moved the binding to %q, want run-new", got)
	}
	if got, ok := gtdCurrentDispatchRun(t, s, card); !ok || got != "run-new" {
		t.Fatalf("late state report moved the current dispatch to %q (ok=%v), want run-new", got, ok)
	}
}

// gtdAssignWithSeveredFactoryHalf stores a dispatch assignment whose queue
// half lands and whose factory half does not: the hook slot is occupied by
// a no-op, so the assignment (and everything riding its transaction) commits
// while the binding write the real hook performs never happens. This is the
// "partial failure stored only the new assignment" shape — the dispatch
// writers' factory half is the part that went missing.
func gtdAssignWithSeveredFactoryHalf(t *testing.T, s *BacklogStore, runID, card, owner string) {
	t.Helper()
	if err := s.recordRuntimeHook(TodoRuntimeRun{RunID: runID, ManifestJSON: "{}"}, &TodoRuntimeAssignment{
		RunID: runID, CardID: card, OwnerLabel: owner, ReportedState: "picked", EventKind: "card.assigned", ProvenanceJSON: "{}",
	}, func() error { return nil }); err != nil {
		t.Fatal(err)
	}
}

// gtdLegacyAssign stores a runtime assignment with no hook at all — the
// pre-binding-write shape an older binary leaves: the queue half lands, no
// binding write is attempted, and no dispatch-identity record is kept. Used
// to construct the states only mixed-version fleets produce.
func gtdLegacyAssign(t *testing.T, s *BacklogStore, runID, card, owner string) {
	t.Helper()
	if err := s.recordRuntime(TodoRuntimeRun{RunID: runID, ManifestJSON: "{}"}, &TodoRuntimeAssignment{
		RunID: runID, CardID: card, OwnerLabel: owner, ReportedState: "picked", EventKind: "card.assigned", ProvenanceJSON: "{}",
	}); err != nil {
		t.Fatal(err)
	}
}

// P1-a (review round-23 gate, card t1538): reconciliation must compare the
// dispatch's actual owner against the factory row's owner — a factory row
// carrying a DIFFERENT owner than the queue's authoritative assignment is a
// stale engagement's residue, not the dispatch the operation read back. The
// old non-empty check blessed it: runtime lane-2 vs factory lane-1 read
// reconciled, and the old lane's approval stayed valid at the gate.
func TestGTDReconcileMatchesFactoryOwnerToDispatchOwner(t *testing.T) {
	root, s, card := gtdReconcileFixture(t)
	ctx := context.Background()
	gtdPlaceRow(t, root, card, "run-1")
	if err := RecordFactoryCardAssignment(root, "run-1", card, "worker-1", ""); err != nil {
		t.Fatal(err)
	}
	firstOp := GTDOperation{OperationID: "gtd-dispatch:run-1:" + card, MissionID: "run-1", Action: GTDActionDispatch, Target: card, SnapshotHash: "snap", ReceiptJSON: []byte(`{}`)}
	if _, err := ExecuteGTDOperation(ctx, s, firstOp, gtdDispatchOwner{store: s, root: root, runID: "run-1", card: card, owner: "worker-1"}); err != nil {
		t.Fatal(err)
	}
	// The card re-dispatches inside the same run under a different owner:
	// the queue's authoritative assignment now names lane-2, while the
	// factory row still carries the first dispatch's owner.
	if err := RecordFactoryCardAssignment(root, "run-1", card, "worker-2", ""); err != nil {
		t.Fatal(err)
	}
	secondOp := GTDOperation{OperationID: "gtd-dispatch:run-1:re:" + card, MissionID: "run-1", Action: GTDActionDispatch, Target: card, SnapshotHash: "snap", ReceiptJSON: []byte(`{}`)}
	if _, err := ExecuteGTDOperation(ctx, s, secondOp, gtdDispatchOwner{store: s, root: root, runID: "run-1", card: card, owner: "worker-2"}); err != nil {
		t.Fatalf("re-dispatch under a new owner failed: %v", err)
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	row, err := db.LoadCard(ctx, "run-1", card)
	if err != nil {
		t.Fatal(err)
	}
	if row.OwnerLabel != NormalizeOwnerLabel("worker-2") {
		t.Fatalf("factory owner = %q after reconciling owner-%s's dispatch — the owner identity was never compared, so the prior owner's approval stays armed", row.OwnerLabel, NormalizeOwnerLabel("worker-2"))
	}
}

// P1-b (review round-23 gate, card t1538): a prior binding is not evidence
// of a subsequent dispatch. When a partial failure stored only the new
// assignment — the queue half landed, the factory binding write did not —
// the binding stays on the prior run, and the new dispatch must REPAIR it,
// not read the stale binding as "superseded" and reconcile over it: that
// skip left the old run's approval armed against the new run's work.
func TestGTDStalePriorBindingIsNotALaterDispatch(t *testing.T) {
	root, s, card := gtdReconcileFixture(t)
	ctx := context.Background()
	gtdPlaceRow(t, root, card, "run-old")
	if err := RecordFactoryCardAssignment(root, "run-old", card, "worker-1", ""); err != nil {
		t.Fatal(err)
	}
	oldOp := GTDOperation{OperationID: "gtd-dispatch:run-old:" + card, MissionID: "run-old", Action: GTDActionDispatch, Target: card, SnapshotHash: "snap", ReceiptJSON: []byte(`{}`)}
	if _, err := ExecuteGTDOperation(ctx, s, oldOp, gtdDispatchOwner{store: s, root: root, runID: "run-old", card: card, owner: "worker-1"}); err != nil {
		t.Fatal(err)
	}
	if got := gtdBindingRun(t, root, card); got != "run-old" {
		t.Fatalf("pre-partial binding = %q, want run-old", got)
	}
	// The new run's dispatch: queue half lands, factory half severed — the
	// binding keeps naming the PRIOR engagement.
	gtdAssignWithSeveredFactoryHalf(t, s, "run-new", card, "worker-2")

	newOp := GTDOperation{OperationID: "gtd-dispatch:run-new:" + card, MissionID: "run-new", Action: GTDActionDispatch, Target: card, SnapshotHash: "snap", ReceiptJSON: []byte(`{}`)}
	if _, err := ExecuteGTDOperation(ctx, s, newOp, gtdDispatchOwner{store: s, root: root, runID: "run-new", card: card, owner: "worker-2"}); err != nil {
		t.Fatalf("the new dispatch failed over a stale prior binding: %v", err)
	}
	if got := gtdBindingRun(t, root, card); got != "run-new" {
		t.Fatalf("binding = %q after the new dispatch read reconciled — the prior run's binding was treated as evidence of a LATER dispatch, arming the old approval", got)
	}
}

// The unadjudicable shape fails CLOSED (review round-23, card t1538): a
// binding naming another run on a queue that carries NO dispatch-identity
// record cannot be told apart from a genuine later engagement by identity —
// skipping would bless a possibly-stale binding (P1-b), repairing would
// possibly drag a genuine later engagement back. The reconciliation refuses
// loudly instead: the operation does not reconcile, the binding does not
// move, and the recovery is one re-dispatch, which records the identity the
// next retry adjudicates on.
func TestGTDUnadjudicatedPriorBindingFailsClosed(t *testing.T) {
	root, s, card := gtdReconcileFixture(t)
	ctx := context.Background()
	// A prior engagement written entirely pre-identity-record: an assigned
	// row (T2 writes the binding with it) and no queue assignment of its own.
	gtdPlaceRow(t, root, card, "run-old")
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	prior, err := db.LoadCard(ctx, "run-old", card)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Transition(ctx, homestate.TransitionRequest{RunID: "run-old", CardID: card, To: homestate.CardAssigned, ExpectedVersion: prior.Version, Actor: "dispatch", Owner: "worker-1", Now: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	// The new run's dispatch arrives as a bare legacy assignment: no binding
	// write, no identity record — the binding still names run-old.
	gtdLegacyAssign(t, s, "run-new", card, "worker-2")

	newOp := GTDOperation{OperationID: "gtd-dispatch:run-new:" + card, MissionID: "run-new", Action: GTDActionDispatch, Target: card, SnapshotHash: "snap", ReceiptJSON: []byte(`{}`)}
	op, err := ExecuteGTDOperation(ctx, s, newOp, gtdDispatchOwner{store: s, root: root, runID: "run-new", card: card, owner: "worker-2"})
	if err == nil {
		t.Fatalf("an unadjudicable prior binding reconciled silently (state %s) — identity cannot decide this shape, so it must refuse", op.State)
	}
	if got := gtdBindingRun(t, root, card); got != "run-old" {
		t.Fatalf("the refusal moved the binding to %q; a fail-closed refusal touches nothing", got)
	}
}

// A card that PROGRESSED past assigned between its dispatch and the
// reconciliation (lease acquired, work started) still reconciles: the row
// carries the dispatch's owner, and refusing it would leave the operation
// permanently unfinished (review round-22 cont., card t1538).
func TestGTDReconcileAdmitsProgressedRow(t *testing.T) {
	root, s, card := gtdReconcileFixture(t)
	gtdPlaceRow(t, root, card, "run-1")
	if err := RecordFactoryCardAssignment(root, "run-1", card, "worker-9", ""); err != nil {
		t.Fatal(err)
	}
	// The card progressed: lease acquired under the same owner.
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.Exec(`UPDATE cards SET state='leased', owner_label='worker-9', lease_holder='worker-9' WHERE run_id='run-1' AND card_id=?`, card); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	op := GTDOperation{OperationID: "gtd-dispatch:run-1:" + card, MissionID: "run-1", Action: GTDActionDispatch, Target: card, SnapshotHash: "snap", ReceiptJSON: []byte(`{}`)}
	owner := gtdDispatchOwner{store: s, root: root, runID: "run-1", card: card, owner: "worker-9"}
	if _, err := ExecuteGTDOperation(context.Background(), s, op, owner); err != nil {
		t.Fatalf("reconciliation refused a progressed card: %v", err)
	}
	if got := gtdBindingRun(t, root, card); got != "run-1" {
		t.Fatalf("binding = %q, want run-1", got)
	}
}
