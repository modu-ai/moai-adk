package homestate

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

// factory_run_lease_rebound_test.go — REQ-FCR-019 / AC-FCR-023 at the
// homestate level (the cli test drives the same matrix through `factory
// next`): a lease abandons the card's earlier non-terminal row only when that
// row is the residue of a dead run.

var rbLeaseNow = time.Date(2026, 9, 26, 9, 0, 0, 0, time.UTC)

// rbLeaseFixture is the AC-FCR-023 matrix. Card id mapping to the SPEC's
// names: c9 -> dead run with a DEAD-pid worker, c7 -> retired dead run,
// c8 -> live run, c6a -> dead run, VALID lease, no worker, c6b -> dead run,
// EXPIRED lease, LIVE worker.
func rbLeaseFixture(t *testing.T) (*FactoryDB, ReconcileOptions) {
	t.Helper()
	db := openSandboxFactory(t)
	opts := rbOpts()
	rbRun(t, db, "run-now", rbPIDLive)
	rbRun(t, db, "run-dead", rbPIDDead)
	rbRun(t, db, "run-live", rbPIDThirdLive)
	rbRun(t, db, "run-ret", rbPIDDead)
	rbRun(t, db, "run-dead2", rbPIDDead)
	rbRun(t, db, "run-dead3", rbPIDDead)
	if _, err := db.DB.Exec(`UPDATE runs SET status='retired' WHERE run_id='run-ret'`); err != nil {
		t.Fatal(err)
	}
	rbCard(t, db, "run-dead", "c9", CardRun)
	rbWorker(t, db, "run-dead", "lane-ghost", rbPIDWorkerDead)
	rbCard(t, db, "run-live", "c8", CardRun)
	rbCard(t, db, "run-ret", "c7", CardAssigned)
	frPlace(t, db, Card{RunID: "run-dead2", CardID: "c6a", State: CardRun, Version: 3, LeaseHolder: "lane-x", LeaseExpiresAt: "2099-01-01T00:00:00Z"})
	frPlace(t, db, Card{RunID: "run-dead3", CardID: "c6b", State: CardRun, Version: 3, LeaseHolder: "lane-y", LeaseExpiresAt: "2020-01-01T00:00:00Z"})
	rbWorker(t, db, "run-dead3", "lane-y", rbPIDWorkerLive)
	return db, opts
}

func rbRowState(t *testing.T, db *FactoryDB, runID, cardID string) string {
	t.Helper()
	c, err := db.LoadCard(context.Background(), runID, cardID)
	if err != nil {
		t.Fatalf("load %s/%s: %v", runID, cardID, err)
	}
	return c.State
}

// AC-FCR-023 — residue of a dead run is abandoned with a pointer event;
// everything else is left alone and the lease still records the new row.
func TestRecordPickedAbandonsDeadRunDuplicate(t *testing.T) {
	db, opts := rbLeaseFixture(t)
	cardsBefore := rbCount(t, db, `SELECT count(*) FROM cards`)
	c9Before, err := db.LoadCard(context.Background(), "run-dead", "c9")
	if err != nil {
		t.Fatal(err)
	}

	for _, id := range []string{"c9", "c7", "c8", "c6a", "c6b"} {
		got, err := db.RecordPickedReconciled(context.Background(), "run-now", id, CardFields{}, "lane-1", rbLeaseNow, opts, nil)
		if err != nil {
			t.Fatalf("record %s: %v", id, err)
		}
		if got.RunID != "run-now" || got.State != CardPicked {
			t.Fatalf("record %s = %+v, want a picked row in run-now", id, got)
		}
	}

	if got := rbRowState(t, db, "run-dead", "c9"); got != CardAbandoned {
		t.Errorf("run-dead c9 = %s, want abandoned (a dead-pid worker row does not hold the run)", got)
	}
	if got := rbRowState(t, db, "run-ret", "c7"); got != CardAbandoned {
		t.Errorf("run-ret c7 = %s, want abandoned (a retired run's dead-owner row)", got)
	}
	if got := rbRowState(t, db, "run-live", "c8"); got != CardRun {
		t.Errorf("run-live c8 = %s, want run (a live owner's row is untouched)", got)
	}
	if got := rbRowState(t, db, "run-dead2", "c6a"); got != CardRun {
		t.Errorf("run-dead2 c6a = %s, want run (a valid lease protects the row)", got)
	}
	if got := rbRowState(t, db, "run-dead3", "c6b"); got != CardRun {
		t.Errorf("run-dead3 c6b = %s, want run (a live worker protects the row)", got)
	}
	if n := rbCount(t, db, `SELECT count(*) FROM cards`); n != cardsBefore+5 {
		t.Errorf("cards rows = %d, want %d (only the five new run-now rows)", n, cardsBefore+5)
	}

	events := frEvents(t, db, "card.rebound")
	if len(events) != 2 {
		t.Fatalf("card.rebound events = %d, want 2 (c9, c7): %v", len(events), events)
	}
	var first map[string]any
	if err := json.Unmarshal([]byte(events[0]), &first); err != nil {
		t.Fatal(err)
	}
	for k, want := range map[string]any{"card_id": "c9", "from_run": "run-dead", "to_run": "run-now", "state": CardRun, "version": float64(c9Before.Version), "stage": c9Before.Stage, "evidence_sha": c9Before.EvidenceSHA} {
		if first[k] != want {
			t.Errorf("card.rebound[%s] = %v, want %v (payload %v)", k, first[k], want, first)
		}
	}
	var second map[string]any
	if err := json.Unmarshal([]byte(events[1]), &second); err != nil {
		t.Fatal(err)
	}
	if second["card_id"] != "c7" || second["from_run"] != "run-ret" || second["to_run"] != "run-now" {
		t.Errorf("second card.rebound payload = %v, want c7 from run-ret to run-now", second)
	}
}

// One transaction: a failure after the T1 record and before the abandon rolls
// the record back too — neither the new run-now row nor the abandon survives.
func TestRecordPickedAbandonIsOneTransaction(t *testing.T) {
	db, opts := rbLeaseFixture(t)
	before := rbDump(t, db)
	boom := errors.New("injected failure between the record and the abandon")
	_, err := db.RecordPickedReconciled(context.Background(), "run-now", "c9", CardFields{}, "lane-1", rbLeaseNow, opts, func() error { return boom })
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want the injected failure", err)
	}
	if after := rbDump(t, db); after != before {
		t.Fatalf("dump changed after the failed lease:\n--- before\n%s\n--- after\n%s", before, after)
	}
}

// The classified stamp is re-checked inside the transaction: a run whose owner
// stamp changed after classification (another leader adopted it) is not
// touched.
func TestRecordPickedLeavesRunRestampedAfterClassification(t *testing.T) {
	db, opts := rbLeaseFixture(t)
	leaseReboundBeforeTx = func() {
		if _, err := db.DB.Exec(`UPDATE runs SET lead_pid=?, lead_process_start='adopter' WHERE run_id='run-dead'`, rbPIDLive); err != nil {
			t.Error(err)
		}
	}
	t.Cleanup(func() { leaseReboundBeforeTx = nil })
	if _, err := db.RecordPickedReconciled(context.Background(), "run-now", "c9", CardFields{}, "lane-1", rbLeaseNow, opts, nil); err != nil {
		t.Fatalf("record: %v", err)
	}
	if got := rbRowState(t, db, "run-dead", "c9"); got != CardRun {
		t.Fatalf("run-dead c9 = %s, want run (the run was adopted after classification)", got)
	}
	if n := len(frEvents(t, db, "card.rebound")); n != 0 {
		t.Fatalf("card.rebound events = %d, want 0", n)
	}
}

// A card no other run holds takes the plain RecordPicked path.
func TestRecordPickedReconciledWithoutRivalIsPlainRecord(t *testing.T) {
	db, opts := rbLeaseFixture(t)
	got, err := db.RecordPickedReconciled(context.Background(), "run-now", "c-fresh", CardFields{}, "lane-1", rbLeaseNow, opts, func() error { return errors.New("must not run") })
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	if got.State != CardPicked || got.RunID != "run-now" {
		t.Fatalf("record = %+v", got)
	}
}
