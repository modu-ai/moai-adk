package homestate

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// zombieShape seeds today's shape: one run led by a live owner and n runs whose
// owner is dead, each holding one unfinished card and one registered worker
// whose pid probes as workerPID.
func zombieShape(t *testing.T, db *FactoryDB, n, workerPID int) {
	t.Helper()
	mustRecordRun(t, db, FactoryRun{RunID: "run-live", ManifestJSON: "{}", LeadPID: rbPIDLive, LeadProcessStart: "live-fp"})
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("run-dead-%d", i)
		mustRecordRun(t, db, FactoryRun{RunID: id, ManifestJSON: "{}", LeadPID: rbPIDDead, LeadProcessStart: fmt.Sprintf("gone-%d", i)})
		mustInsertCard(t, db, id, fmt.Sprintf("c-%d", i), "run")
		rbWorker(t, db, id, fmt.Sprintf("lane-%d", i), workerPID)
	}
}

func countWhere(t *testing.T, db *FactoryDB, query string, args ...any) int {
	t.Helper()
	var n int
	if err := db.DB.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return n
}

// Fix 1+2+5: 1 live + 5 dead runs, each dead run holding a card and a dead
// worker row. The reconciler retires all five, moves their cards to the live
// run and removes the dead worker rows.
func TestReconcileRetiresDeadRunsHoldingCardsIntoTheLiveRun(t *testing.T) {
	db := openSandboxFactory(t)
	zombieShape(t, db, 5, rbPIDWorkerDead)

	rec, err := db.ReconcileActiveRuns(context.Background(), rbOpts())
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if len(rec.Retired) != 5 || len(rec.Preserved) != 0 {
		t.Fatalf("retired=%d preserved=%d, want 5 and 0", len(rec.Retired), len(rec.Preserved))
	}
	if len(rec.Remaining) != 1 || rec.Remaining[0].RunID != "run-live" {
		t.Fatalf("remaining = %+v, want only run-live", rec.Remaining)
	}
	for i := 0; i < 5; i++ {
		if got := runStatus(t, db, fmt.Sprintf("run-dead-%d", i)); got != "retired" {
			t.Errorf("run-dead-%d status = %q", i, got)
		}
		if n := countWhere(t, db, `SELECT COUNT(*) FROM cards WHERE run_id='run-live' AND card_id=?`, fmt.Sprintf("c-%d", i)); n != 1 {
			t.Errorf("card c-%d not in run-live", i)
		}
	}
	if n := countWhere(t, db, `SELECT COUNT(*) FROM workers WHERE pid=?`, rbPIDWorkerDead); n != 0 {
		t.Errorf("%d dead worker rows remain", n)
	}
	if n := countWhere(t, db, `SELECT COUNT(*) FROM events WHERE kind='worker.reaped'`); n != 5 {
		t.Errorf("worker.reaped events = %d, want 5", n)
	}
}

// A worker whose pid cannot be decided is never removed and keeps its run, its
// cards and its row (safe direction).
func TestReconcileKeepsRunWithUndecidedWorker(t *testing.T) {
	db := openSandboxFactory(t)
	zombieShape(t, db, 1, rbPIDWorkerUnknown)
	if _, err := db.ReconcileActiveRuns(context.Background(), rbOpts()); err != nil {
		t.Fatal(err)
	}
	if got := runStatus(t, db, "run-dead-0"); got != "active" {
		t.Fatalf("status = %q, want active", got)
	}
	if countWhere(t, db, `SELECT COUNT(*) FROM cards WHERE run_id='run-dead-0'`) != 1 || countWhere(t, db, `SELECT COUNT(*) FROM workers`) != 1 {
		t.Fatal("cards and worker row must be untouched")
	}
}

// With no live run there is nowhere to move the cards: the run stays preserved.
func TestReconcileKeepsCardsWhenNoLiveRunExists(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-dead", ManifestJSON: "{}", LeadPID: rbPIDDead, LeadProcessStart: "gone"})
	mustInsertCard(t, db, "run-dead", "c", "run")
	rec, err := db.ReconcileActiveRuns(context.Background(), rbOpts())
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.Retired) != 0 || len(rec.Preserved) != 1 || runStatus(t, db, "run-dead") != "active" {
		t.Fatalf("rec = %+v", rec)
	}
}

// Two live runs make the target ambiguous: nothing moves.
func TestReconcileKeepsCardsWhenTwoLiveRunsExist(t *testing.T) {
	db := openSandboxFactory(t)
	zombieShape(t, db, 1, rbPIDWorkerDead)
	mustRecordRun(t, db, FactoryRun{RunID: "run-live-2", ManifestJSON: "{}", LeadPID: rbPIDThirdLive, LeadProcessStart: "fp2"})
	if _, err := db.ReconcileActiveRuns(context.Background(), rbOpts()); err != nil {
		t.Fatal(err)
	}
	if runStatus(t, db, "run-dead-0") != "active" || countWhere(t, db, `SELECT COUNT(*) FROM cards WHERE run_id='run-dead-0'`) != 1 {
		t.Fatal("run with cards must stay when the target is ambiguous")
	}
}

// A refused consolidation (the target already holds a diverging row of the same
// card) leaves the run preserved and the rest of the reconcile going.
func TestReconcileRefusedConsolidationLeavesRunPreserved(t *testing.T) {
	db := openSandboxFactory(t)
	zombieShape(t, db, 2, rbPIDWorkerDead)
	mustInsertCard(t, db, "run-live", "c-0", "done") // terminal row in the target
	rec, err := db.ReconcileActiveRuns(context.Background(), rbOpts())
	if err != nil {
		t.Fatalf("a refusal must not fail the pass: %v", err)
	}
	if runStatus(t, db, "run-dead-0") != "active" || runStatus(t, db, "run-dead-1") != "retired" {
		t.Fatalf("statuses: %s / %s (preserved=%d)", runStatus(t, db, "run-dead-0"), runStatus(t, db, "run-dead-1"), len(rec.Preserved))
	}
}

func ownerOf(t *testing.T, db *FactoryDB, runID string) RunOwner {
	t.Helper()
	owners, err := db.ClassifyRuns(context.Background(), rbOpts())
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range owners {
		if o.RunID == runID {
			return o
		}
	}
	t.Fatalf("run %s not classified", runID)
	return RunOwner{}
}

// A dead run with no card but a dead worker row retires and the row is reaped
// by the records-free retirement path (retireRunUnlessLive), not the rebind.
func TestReconcileReapsDeadWorkerOfCardlessRun(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-dead", ManifestJSON: "{}", LeadPID: rbPIDDead, LeadProcessStart: "gone"})
	rbWorker(t, db, "run-dead", "lane-x", rbPIDWorkerDead)
	if _, err := db.ReconcileActiveRuns(context.Background(), rbOpts()); err != nil {
		t.Fatal(err)
	}
	if runStatus(t, db, "run-dead") != "retired" || countWhere(t, db, `SELECT COUNT(*) FROM workers`) != 0 {
		t.Fatal("the run must retire and its dead worker row must go")
	}
	if n := countWhere(t, db, `SELECT COUNT(*) FROM events WHERE kind='worker.reaped' AND run_id='run-dead'`); n != 1 {
		t.Fatalf("worker.reaped events = %d, want 1", n)
	}
}

// Transfer needs exactly one Live owner and no Indeterminate one: an
// indeterminate owner may be a live leader, so the dead run's cards stay put.
func TestReconcileKeepsCardsWhileAnOwnerIsIndeterminate(t *testing.T) {
	db := openSandboxFactory(t)
	zombieShape(t, db, 1, rbPIDWorkerDead)
	mustRecordRun(t, db, FactoryRun{RunID: "run-unk", ManifestJSON: "{}", LeadPID: rbPIDIndeterminate, LeadProcessStart: "fp"})
	if _, err := db.ReconcileActiveRuns(context.Background(), rbOpts()); err != nil {
		t.Fatal(err)
	}
	if runStatus(t, db, "run-dead-0") != "active" || countWhere(t, db, `SELECT COUNT(*) FROM cards WHERE run_id='run-dead-0'`) != 1 {
		t.Fatal("cards must stay in place while another owner is indeterminate")
	}
}

// The transfer names the moved card ids in the source's run.retired event, so
// the loss of the (non-following) approval is visible.
func TestReconcileTransferEventNamesTheCards(t *testing.T) {
	db := openSandboxFactory(t)
	zombieShape(t, db, 1, rbPIDWorkerDead)
	if _, err := db.ReconcileActiveRuns(context.Background(), rbOpts()); err != nil {
		t.Fatal(err)
	}
	var payload string
	if err := db.DB.QueryRow(`SELECT payload_json FROM events WHERE run_id='run-dead-0' AND kind='run.retired'`).Scan(&payload); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(payload, `"moved_cards":["c-0"]`) || !strings.Contains(payload, `"rebound_to":"run-live"`) {
		t.Fatalf("payload = %s", payload)
	}
}

// consolidateIntoLive requires the source's own stamp: a source proved dead by
// the peer or boot proof is refused, the refusal is swallowed (false, nil) and
// nothing moves.
func TestConsolidateIntoLiveRefusesNonStampBasis(t *testing.T) {
	for _, basis := range []ProofBasis{BasisPeer, BasisBoot} {
		db := openSandboxFactory(t)
		zombieShape(t, db, 1, rbPIDWorkerDead)
		src, tgt := ownerOf(t, db, "run-dead-0"), ownerOf(t, db, "run-live")
		src.Basis = basis
		ok, err := db.consolidateIntoLive(context.Background(), src, tgt, rbOpts())
		if ok || err != nil {
			t.Fatalf("basis %s: ok=%v err=%v, want false/nil", basis, ok, err)
		}
		if countWhere(t, db, `SELECT COUNT(*) FROM cards WHERE run_id='run-dead-0'`) != 1 {
			t.Fatalf("basis %s: the card moved", basis)
		}
	}
}

// Only ErrRebindRefused is swallowed: a store failure propagates, and a failure
// at the last injection point rolls back the card move, the worker deletion and
// the reaped event together.
func TestConsolidateIntoLiveRollsBackAndPropagatesStoreFailure(t *testing.T) {
	db := openSandboxFactory(t)
	zombieShape(t, db, 1, rbPIDWorkerDead)
	src, tgt := ownerOf(t, db, "run-dead-0"), ownerOf(t, db, "run-live")
	boom := errors.New("disk full")
	rebindFault = func(p string) error {
		if p == rebindPointBeforeCommit {
			return boom
		}
		return nil
	}
	t.Cleanup(func() { rebindFault = nil })
	ok, err := db.consolidateIntoLive(context.Background(), src, tgt, rbOpts())
	if ok || !errors.Is(err, boom) {
		t.Fatalf("ok=%v err=%v, want the store failure propagated", ok, err)
	}
	if runStatus(t, db, "run-dead-0") != "active" ||
		countWhere(t, db, `SELECT COUNT(*) FROM cards WHERE run_id='run-dead-0'`) != 1 ||
		countWhere(t, db, `SELECT COUNT(*) FROM workers`) != 1 ||
		countWhere(t, db, `SELECT COUNT(*) FROM events WHERE kind='worker.reaped'`) != 0 {
		t.Fatal("a failure before commit must roll back the cards, the worker deletion and the reaped event")
	}
}
