package homestate

import (
	"context"
	"errors"
	"testing"
)

// factory_run_rebind_faults_test.go — the failure edges of the consolidation
// operation and of the lease-path abandon (card t1963): bad arguments, rows that
// are not there, and a store that breaks mid-operation. Each failure surfaces as
// an error and leaves the tables that still exist exactly as they were — the
// transaction rolled back, nothing half-moved.

func TestRebindRejectsMalformedRequests(t *testing.T) {
	db, opts := rbBase(t)
	good := rbRequest(t, db, opts, true)
	for _, tc := range []struct {
		name string
		mut  func(*RebindRequest)
	}{
		{"same run for source and target", func(r *RebindRequest) { r.Target = r.Source }},
		{"empty source id", func(r *RebindRequest) { r.Source.RunID = "" }},
		{"empty target id", func(r *RebindRequest) { r.Target.RunID = "  " }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := good
			tc.mut(&req)
			before := rbDump(t, db)
			_, err := db.RebindRun(context.Background(), req, opts)
			rbRefusedWith(t, err)
			if after := rbDump(t, db); after != before {
				t.Fatalf("dump changed after the refusal")
			}
		})
	}
}

// A run the caller classified but the store no longer holds is refused at the
// stamp re-read, for the source and for the target alike.
func TestRebindRefusesRunsThatAreGone(t *testing.T) {
	for _, gone := range []string{"run-old", "run-new"} {
		t.Run(gone, func(t *testing.T) {
			db, opts := rbBase(t)
			req := rbRequest(t, db, opts, true)
			if _, err := db.DB.Exec(`DELETE FROM runs WHERE run_id=?`, gone); err != nil {
				t.Fatal(err)
			}
			_, err := db.RebindRun(context.Background(), req, opts)
			rbRefusedWith(t, err, gone, "does not exist")
		})
	}
}

// A store that fails mid-operation fails the operation and rolls it back: the
// workers, the dispatch bindings and the events are each broken in turn, and the
// cards and runs the operation had already touched are back where they were.
func TestRebindRollsBackWhenTheStoreBreaks(t *testing.T) {
	for _, table := range []string{"workers", "card_dispatch", "events"} {
		t.Run(table, func(t *testing.T) {
			db, opts := rbBase(t)
			req := rbRequest(t, db, opts, true)
			cardsBefore := rbDumpTable(t, db, "cards")
			runsBefore := rbDumpTable(t, db, "runs")
			// SQL: the table name comes from the fixed list above.
			if _, err := db.DB.Exec(`DROP TABLE ` + table); err != nil {
				t.Fatal(err)
			}
			_, err := db.RebindRun(context.Background(), req, opts)
			if err == nil {
				t.Fatalf("operation succeeded with the %s table gone", table)
			}
			if errors.Is(err, ErrRebindRefused) {
				t.Fatalf("a broken store reported as a refusal: %v", err)
			}
			if got := rbDumpTable(t, db, "cards"); got != cardsBefore {
				t.Errorf("cards changed although the operation failed:\n got %s\nwant %s", got, cardsBefore)
			}
			if got := rbDumpTable(t, db, "runs"); got != runsBefore {
				t.Errorf("runs changed although the operation failed:\n got %s\nwant %s", got, runsBefore)
			}
		})
	}
}

// A write the store refuses part-way — the re-key, the abandon, the flip of the
// source to retired — fails the operation and the transaction rolls every
// earlier write back. The triggers stand in for a failing disk or a constraint:
// they abort exactly the statement named.
func TestRebindRollsBackWhenAWriteIsRefused(t *testing.T) {
	for _, tc := range []struct{ name, trigger string }{
		{"re-key", `CREATE TRIGGER t_fail BEFORE UPDATE OF run_id ON cards BEGIN SELECT RAISE(ABORT, 'injected re-key failure'); END`},
		{"abandon", `CREATE TRIGGER t_fail BEFORE UPDATE OF state ON cards WHEN NEW.state='abandoned' BEGIN SELECT RAISE(ABORT, 'injected abandon failure'); END`},
		{"flip", `CREATE TRIGGER t_fail BEFORE UPDATE OF status ON runs WHEN NEW.status='retired' BEGIN SELECT RAISE(ABORT, 'injected flip failure'); END`},
		{"cards table gone", `DROP TABLE cards`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, opts := rbBase(t)
			req := rbRequest(t, db, opts, true)
			var before string
			if tc.name != "cards table gone" {
				before = rbDump(t, db)
			}
			if _, err := db.DB.Exec(tc.trigger); err != nil {
				t.Fatal(err)
			}
			_, err := db.RebindRun(context.Background(), req, opts)
			if err == nil || errors.Is(err, ErrRebindRefused) {
				t.Fatalf("err = %v, want a store failure (not a refusal)", err)
			}
			if before != "" {
				if after := rbDump(t, db); after != before {
					t.Fatalf("dump changed after the failed operation:\n--- before\n%s\n--- after\n%s", before, after)
				}
			}
		})
	}
}

// A closed handle fails every entry point with an error and no panic.
func TestRebindAndReconciledRecordFailOnAClosedStore(t *testing.T) {
	db, opts := rbLeaseFixture(t)
	req := RebindRequest{Source: rbOwner(t, db, opts, "run-dead"), Target: rbOwner(t, db, opts, "run-now"), RequireStampBasis: true}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := db.RebindRun(context.Background(), req, opts); err == nil {
		t.Error("RebindRun on a closed store succeeded")
	}
	if _, err := db.RecordPickedReconciled(context.Background(), "run-now", "c9", CardFields{}, "lane-1", rbLeaseNow, opts, nil); err == nil {
		t.Error("RecordPickedReconciled on a closed store succeeded")
	}
	if _, err := db.PlanLeaderEntry(context.Background(), opts); err == nil {
		t.Error("PlanLeaderEntry on a closed store succeeded")
	}
}

func TestRecordPickedReconciledRejectsBadArguments(t *testing.T) {
	db, opts := rbLeaseFixture(t)
	for _, tc := range []struct{ run, card string }{{"", "c9"}, {"run-now", ""}, {"run-now", "bad card id"}} {
		if _, err := db.RecordPickedReconciled(context.Background(), tc.run, tc.card, CardFields{}, "lane-1", rbLeaseNow, opts, nil); !errors.Is(err, ErrInvalidCardInput) {
			t.Errorf("RecordPickedReconciled(%q, %q) err = %v, want ErrInvalidCardInput", tc.run, tc.card, err)
		}
	}
	// A rival row exists for c9, so the field validation of the reconciling
	// path (not the plain path) rejects a malformed field.
	badSpec := "not-a-spec"
	if _, err := db.RecordPickedReconciled(context.Background(), "run-now", "c9", CardFields{SpecID: &badSpec}, "lane-1", rbLeaseNow, opts, nil); !errors.Is(err, ErrInvalidCardInput) {
		t.Errorf("a malformed field on the reconciling path: err = %v, want ErrInvalidCardInput", err)
	}
}

// A store whose events table is gone fails the abandon and rolls the T1 record
// back with it; a store whose workers table is gone cannot prove the dead run
// has no live worker, so it abandons nothing.
func TestRecordPickedAbandonFailsClosedWhenTheStoreBreaks(t *testing.T) {
	t.Run("events table gone: the lease fails as a whole", func(t *testing.T) {
		db, opts := rbLeaseFixture(t)
		cardsBefore := rbDumpTable(t, db, "cards")
		if _, err := db.DB.Exec(`DROP TABLE events`); err != nil {
			t.Fatal(err)
		}
		if _, err := db.RecordPickedReconciled(context.Background(), "run-now", "c9", CardFields{}, "lane-1", rbLeaseNow, opts, nil); err == nil {
			t.Fatal("the lease succeeded with no events table")
		}
		if got := rbDumpTable(t, db, "cards"); got != cardsBefore {
			t.Errorf("cards changed after the failed lease:\n got %s\nwant %s", got, cardsBefore)
		}
	})
	t.Run("workers table gone: nothing is abandoned", func(t *testing.T) {
		db, opts := rbLeaseFixture(t)
		if _, err := db.DB.Exec(`DROP TABLE workers`); err != nil {
			t.Fatal(err)
		}
		if _, err := db.RecordPickedReconciled(context.Background(), "run-now", "c9", CardFields{}, "lane-1", rbLeaseNow, opts, nil); err != nil {
			t.Fatalf("lease: %v", err)
		}
		if got := rbRowState(t, db, "run-dead", "c9"); got != CardRun {
			t.Fatalf("run-dead c9 = %s, want run (the worker check could not be made)", got)
		}
	})
}
