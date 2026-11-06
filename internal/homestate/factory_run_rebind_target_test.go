package homestate

import (
	"context"
	"testing"
)

// factory_run_rebind_target_test.go — SPEC-FACTORY-COMPLETION-RECOVERY-001
// REQ-FCR-020 (card t1963, ledger L13): the fresh classification RebindRun
// takes just before BEGIN is applied to the TARGET too. A process dying does
// not change the run row's owner stamp, so requireRunStamp inside the
// transaction cannot see it; without this check cards are moved onto a dead
// run while the source is retired.

// lceReclassified runs RebindRun on a request classified with rbOpts (target
// live) under opts, which is the classification at the moment of the call.
func lceReclassified(t *testing.T, mutate func(db *FactoryDB), opts ReconcileOptions) (*FactoryDB, string, error) {
	t.Helper()
	db, req := rbBase(t)
	request := rbRequest(t, db, req, true)
	if mutate != nil {
		mutate(db)
	}
	before := rbDump(t, db)
	_, err := db.RebindRun(context.Background(), request, opts)
	return db, before, err
}

func lceAssertNothingMoved(t *testing.T, db *FactoryDB, before string) {
	t.Helper()
	if after := rbDump(t, db); after != before {
		t.Fatalf("a refused consolidation changed stored rows:\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if got := runStatus(t, db, "run-old"); got != "active" {
		t.Fatalf("source run-old status = %q after the refusal, want active", got)
	}
	if n := rbCount(t, db, `SELECT count(*) FROM events WHERE kind='run.retired'`); n != 0 {
		t.Fatalf("%d run.retired events after a refused consolidation, want 0", n)
	}
	if n := rbCount(t, db, `SELECT count(*) FROM cards WHERE run_id='run-old' AND card_id='c1'`); n != 1 {
		t.Fatalf("c1 no longer in run-old")
	}
}

// The target's owner died between classification and the call: refuse.
func TestRebindRefusesTargetThatDiedBeforeTheTransaction(t *testing.T) {
	// Both owners now classify dead (the target's pid, 21, is in the dead set).
	db, before, err := lceReclassified(t, nil, lceOpts(rbPIDDead, rbPIDLive))
	rbRefusedWith(t, err, "run-new")
	lceAssertNothingMoved(t, db, before)
}

// The target's owner can no longer be classified live OR dead between
// classification and the call (REQ-FCR-020, step 2): indeterminate is not live,
// so the consolidation is refused and nothing moves. Pins the `!= OwnerLive`
// predicate against a weaker `== OwnerDead` one, which would let an
// undecidable owner receive the source's cards.
func TestRebindRefusesTargetWhoseOwnerBecameIndeterminate(t *testing.T) {
	indeterminateTarget := ReconcileOptions{
		Classify: func(pid int, _ string) OwnerClassification {
			if pid == rbPIDLive {
				return OwnerIndeterminate
			}
			return rbOpts().Classify(pid, "")
		},
		WorkerProbe: rbOpts().WorkerProbe,
	}
	db, before, err := lceReclassified(t, nil, indeterminateTarget)
	rbRefusedWith(t, err, "run-new", string(OwnerIndeterminate))
	lceAssertNothingMoved(t, db, before)
}

// A target that is no longer active (retired in between) is absent from the
// freshly classified active set: refuse.
func TestRebindRefusesTargetRetiredBeforeTheTransaction(t *testing.T) {
	retireTarget := func(db *FactoryDB) {
		if _, err := db.DB.Exec(`UPDATE runs SET status='retired' WHERE run_id='run-new'`); err != nil {
			t.Fatal(err)
		}
	}
	db, before, err := lceReclassified(t, retireTarget, rbOpts())
	rbRefusedWith(t, err, "run-new")
	lceAssertNothingMoved(t, db, before)
}

// A live target is unchanged: the consolidation proceeds.
func TestRebindProceedsWhenTargetStaysLive(t *testing.T) {
	db, _, err := lceReclassified(t, nil, rbOpts())
	if err != nil {
		t.Fatalf("rebind with a live target: %v", err)
	}
	if got := runStatus(t, db, "run-old"); got != "retired" {
		t.Fatalf("run-old status = %q, want retired", got)
	}
}
