package homestate

import (
	"context"
	"testing"
)

// factory_run_consolidated_test.go — the reader behind REQ-FCR-024 (card t1963):
// a run is "retired by a consolidation" only when its status is retired AND its
// run.retired event carries rebound_to. An ordinary retirement (retireRun, the
// reconciler) writes the same event kind without rebound_to and must read false.

func TestRunRetiredByConsolidation(t *testing.T) {
	ctx := context.Background()

	t.Run("a consolidated source reads true with its target", func(t *testing.T) {
		db, opts := rbBase(t)
		if _, err := db.RebindRun(ctx, rbRequest(t, db, opts, true), opts); err != nil {
			t.Fatalf("rebind: %v", err)
		}
		into, retired, err := db.RunRetiredByConsolidation(ctx, "run-old")
		if err != nil || !retired || into != "run-new" {
			t.Fatalf("RunRetiredByConsolidation(run-old) = (%q, %v, %v), want (run-new, true, nil)", into, retired, err)
		}
	})

	t.Run("an ordinary retirement reads false", func(t *testing.T) {
		db := openSandboxFactory(t)
		opts := rbOpts()
		rbRun(t, db, "run-plain", rbPIDDead) // no cards: the reconciler retires it
		if _, err := db.ReconcileActiveRuns(ctx, opts); err != nil {
			t.Fatalf("reconcile: %v", err)
		}
		if got := runStatus(t, db, "run-plain"); got != "retired" {
			t.Fatalf("fixture: run-plain status = %s, want retired", got)
		}
		into, retired, err := db.RunRetiredByConsolidation(ctx, "run-plain")
		if err != nil || retired || into != "" {
			t.Fatalf("RunRetiredByConsolidation(run-plain) = (%q, %v, %v), want (\"\", false, nil)", into, retired, err)
		}
	})

	t.Run("an active run, and a run that does not exist, read false", func(t *testing.T) {
		db, _ := rbBase(t)
		for _, id := range []string{"run-new", "run-missing"} {
			into, retired, err := db.RunRetiredByConsolidation(ctx, id)
			if err != nil || retired || into != "" {
				t.Errorf("RunRetiredByConsolidation(%s) = (%q, %v, %v), want (\"\", false, nil)", id, into, retired, err)
			}
		}
	})

	t.Run("a closed store is an error, not a false", func(t *testing.T) {
		db, _ := rbBase(t)
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
		if _, retired, err := db.RunRetiredByConsolidation(ctx, "run-old"); err == nil || retired {
			t.Fatalf("closed store: retired=%v err=%v, want an error", retired, err)
		}
	})
}
