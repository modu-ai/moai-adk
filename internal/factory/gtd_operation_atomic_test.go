package factory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// gtd_operation_atomic_test.go — REQ-FCR-024 atomicity / AC-FCR-017 (10b)
// (SPEC v0.3.6, ledger L9): the retired-by-consolidation check and every repair
// write live in ONE factory write transaction. A consolidation cannot slip in
// between them, because the repair already holds the write lock (the factory DSN
// opens transactions IMMEDIATE) from before the check to its commit.
//
// The seam (gtdRepairAfterCheck) runs a REAL RebindRun in the window after the
// check and before the first write, on its own database handle:
//   - with the check and the writes in separate transactions (the pre-fix code, and
//     mutant (j)) nothing holds the lock, the consolidation commits in the window,
//     and the writes then recreate the card's row and binding in the run it retired;
//   - in one transaction the consolidation blocks on the write lock until the repair
//     commits, then moves the repaired card to the survivor.
//
// Either outcome the AC allows is accepted; a row or binding in the retired source
// is not.

func TestRepairGTDDispatchNeverRecreatesWorkInASourceConsolidatedMidRepair(t *testing.T) {
	root, s, card := gtdReconcileFixture(t)
	ctx := context.Background()

	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	for id, pid := range map[string]int{"run-src": 20, "run-tgt": 21} {
		if err := db.RecordRun(ctx, homestate.FactoryRun{RunID: id, ManifestJSON: "{}", LeadPID: pid, LeadProcessStart: "fp"}); err != nil {
			t.Fatal(err)
		}
	}
	opts := homestate.ReconcileOptions{Classify: func(pid int, _ string) homestate.OwnerClassification {
		if pid == 21 {
			return homestate.OwnerLive
		}
		return homestate.OwnerDead
	}}
	owners, err := db.ClassifyRuns(ctx, opts)
	if err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	byID := map[string]homestate.RunOwner{}
	for _, o := range owners {
		byID[o.RunID] = o
	}
	// The queue half of an in-flight dispatch to run-src; the factory half never landed.
	gtdAssignWithSeveredFactoryHalf(t, s, "run-src", card, "worker-1")

	type rebindResult struct {
		res homestate.RebindResult
		err error
	}
	done := make(chan rebindResult, 1)
	started := false
	gtdRepairAfterCheck = func() {
		started = true
		go func() {
			rdb, err := homestate.OpenFactory(root)
			if err != nil {
				done <- rebindResult{err: err}
				return
			}
			defer func() { _ = rdb.Close() }()
			res, err := rdb.RebindRun(ctx, homestate.RebindRequest{Source: byID["run-src"], Target: byID["run-tgt"], RequireStampBasis: true}, opts)
			done <- rebindResult{res: res, err: err}
		}()
		// Separate transactions leave the write lock free here, so the
		// consolidation commits within moments. In one transaction it is blocked
		// behind the repair's lock: do not wait for it — the repair must finish first.
		select {
		case r := <-done:
			done <- r // hand the result back for the final collection
		case <-time.After(3 * time.Second):
		}
	}
	t.Cleanup(func() { gtdRepairAfterCheck = nil })

	repairErr := repairGTDDispatchRecord(ctx, s, gtdFactoryDBForStore(s), card, "run-src")
	if !started {
		t.Fatal("the seam never ran: the repair did not reach the point after the retired-run check")
	}
	var rb rebindResult
	select {
	case rb = <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("the consolidation never finished")
	}
	if rb.err != nil {
		t.Fatalf("the consolidation returned %v, want it to commit (before the repair's writes, or after the repair's commit)", rb.err)
	}
	if repairErr != nil && !errors.Is(repairErr, homestate.ErrRunRetiredByConsolidation) {
		t.Fatalf("repair err = %v, want nil or a typed refusal", repairErr)
	}

	// The invariant: nothing live was recreated in the retired source, and the
	// binding does not name it.
	vdb, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = vdb.Close() }()
	var status string
	if err := vdb.DB.QueryRow(`SELECT status FROM runs WHERE run_id='run-src'`).Scan(&status); err != nil || status != "retired" {
		t.Fatalf("run-src status = %q (%v), want retired by the consolidation", status, err)
	}
	var inSrc int
	if err := vdb.DB.QueryRow(`SELECT count(*) FROM cards WHERE run_id='run-src' AND card_id=? AND state NOT IN ('done','failed','abandoned')`, card).Scan(&inSrc); err != nil {
		t.Fatal(err)
	}
	var bound string
	switch err := vdb.DB.QueryRow(`SELECT run_id FROM card_dispatch WHERE card_id=?`, card).Scan(&bound); {
	case err != nil && err.Error() != "sql: no rows in result set":
		t.Fatal(err)
	}
	if inSrc != 0 || bound == "run-src" {
		t.Fatalf("work was recreated in the retired source: non-terminal rows in run-src=%d, binding=%q (repair err=%v)", inSrc, bound, repairErr)
	}
	if repairErr == nil {
		// The repair completed before the consolidation: its card now lives under the survivor.
		var inTgt int
		if err := vdb.DB.QueryRow(`SELECT count(*) FROM cards WHERE run_id='run-tgt' AND card_id=? AND state NOT IN ('done','failed','abandoned')`, card).Scan(&inTgt); err != nil || inTgt != 1 || bound != "run-tgt" {
			t.Fatalf("repair completed but the card is not under the survivor: rows in run-tgt=%d (%v), binding=%q", inTgt, err, bound)
		}
	}
}
