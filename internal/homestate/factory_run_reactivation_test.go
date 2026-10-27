package homestate

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// factory_run_reactivation_test.go — SPEC-FACTORY-COMPLETION-RECOVERY-001
// REQ-FCR-024 (card t1963, ledger L12): a run a consolidation retired can never
// be re-activated, because re-activation would switch the storage guard's
// marker off (the guard reads 'retired' + run.retired.rebound_to). RecordRun
// (an upsert to 'active') and ResumeRun (re-activates any non-active row) both
// refuse it, in the same transaction as their write.

// lceConsolidated returns a store where run-old was consolidated into run-new.
func lceConsolidated(t *testing.T) *FactoryDB {
	t.Helper()
	db, opts := rbBase(t)
	if _, err := db.RebindRun(context.Background(), rbRequest(t, db, opts, true), opts); err != nil {
		t.Fatalf("rebind: %v", err)
	}
	return db
}

func lceRefusedReactivation(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, ErrRunRetiredByConsolidation) {
		t.Fatalf("err = %v, want ErrRunRetiredByConsolidation", err)
	}
	if !strings.Contains(err.Error(), "run-old") || !strings.Contains(err.Error(), "run-new") {
		t.Fatalf("refusal %q must name the consolidated run and the run its cards went to", err)
	}
}

func TestRecordRunRefusesAConsolidatedSource(t *testing.T) {
	db := lceConsolidated(t)
	before := rbDump(t, db)
	err := db.RecordRun(context.Background(), FactoryRun{RunID: "run-old", ManifestJSON: "{}", LeadPID: rbPIDLive, LeadProcessStart: "fp"})
	lceRefusedReactivation(t, err)
	if got := runStatus(t, db, "run-old"); got != "retired" {
		t.Fatalf("run-old status = %q after the refusal, want retired", got)
	}
	if after := rbDump(t, db); after != before {
		t.Fatalf("a refused RecordRun changed stored rows (a run.started event included):\nbefore:\n%s\nafter:\n%s", before, after)
	}
	lceGuardsStillRefuse(t, db)
}

func TestResumeRunRefusesAConsolidatedSource(t *testing.T) {
	db := lceConsolidated(t)
	before := rbDump(t, db)
	err := db.ResumeRun(context.Background(), "run-old", 4242, "fp", "test")
	lceRefusedReactivation(t, err)
	if got := runStatus(t, db, "run-old"); got != "retired" {
		t.Fatalf("run-old status = %q after the refusal, want retired", got)
	}
	if after := rbDump(t, db); after != before {
		t.Fatalf("a refused ResumeRun changed stored rows (a run.resumed event included):\nbefore:\n%s\nafter:\n%s", before, after)
	}
	lceGuardsStillRefuse(t, db)
}

// After the refused re-activation, row creation in the source is still refused.
func lceGuardsStillRefuse(t *testing.T, db *FactoryDB) {
	t.Helper()
	if _, err := db.RecordPicked(context.Background(), "run-old", "c-new", CardFields{}, "op", time.Now().UTC()); !errors.Is(err, ErrRunRetiredByConsolidation) {
		t.Fatalf("RecordPicked into the consolidated source: err = %v, want ErrRunRetiredByConsolidation", err)
	}
	if n := rbCount(t, db, `SELECT count(*) FROM cards WHERE run_id='run-old' AND card_id='c-new'`); n != 0 {
		t.Fatalf("%d card rows were created in the consolidated source", n)
	}
}

// A run retired the ordinary way (no rebound_to) stays re-recordable and
// resumable exactly as before.
func TestReactivationOfAnOrdinarilyRetiredRunUnchanged(t *testing.T) {
	for _, tc := range []struct {
		name   string
		revive func(*FactoryDB) error
		kind   string
	}{
		{"RecordRun", func(db *FactoryDB) error {
			return db.RecordRun(context.Background(), FactoryRun{RunID: "run-dead", ManifestJSON: "{}", LeadPID: rbPIDLive, LeadProcessStart: "fp"})
		}, "run.started"},
		{"ResumeRun", func(db *FactoryDB) error {
			return db.ResumeRun(context.Background(), "run-dead", 4242, "fp", "test")
		}, "run.resumed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := openSandboxFactory(t)
			opts := rbOpts()
			rbRun(t, db, "run-dead", rbPIDDead)
			if _, err := db.RetireRunIfDead(context.Background(), "run-dead", opts); err != nil {
				t.Fatalf("ordinary retire: %v", err)
			}
			if got := runStatus(t, db, "run-dead"); got != "retired" {
				t.Fatalf("fixture: run-dead status = %q, want retired", got)
			}
			eventsBefore := rbCount(t, db, `SELECT count(*) FROM events WHERE run_id='run-dead' AND kind=?`, tc.kind)
			if err := tc.revive(db); err != nil {
				t.Fatalf("%s on an ordinarily retired run: %v", tc.name, err)
			}
			if got := runStatus(t, db, "run-dead"); got != "active" {
				t.Fatalf("status after %s = %q, want active", tc.name, got)
			}
			if n := rbCount(t, db, `SELECT count(*) FROM events WHERE run_id='run-dead' AND kind=?`, tc.kind); n != eventsBefore+1 {
				t.Fatalf("%d %s events, want %d (one more than before)", n, tc.kind, eventsBefore+1)
			}
		})
	}
}
