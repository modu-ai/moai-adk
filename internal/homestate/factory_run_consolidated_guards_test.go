package homestate

import (
	"context"
	"errors"
	"testing"
	"time"
)

// factory_run_consolidated_guards_test.go — REQ-FCR-024 at the STORAGE layer
// (card t1963, codex review P1 on 8b10f67066): creating a card row in a run a
// consolidation retired, and pointing a dispatch binding at it, are refused in
// the write's own transaction — so every writer is covered, not the repair path
// alone. Only creation and re-pointing are guarded: reading an existing row, an
// ordinary (non-consolidation) retirement, and the consolidation's own writes are
// untouched.

func rbConsolidated(t *testing.T) *FactoryDB {
	t.Helper()
	db, opts := rbBase(t)
	if _, err := db.RebindRun(context.Background(), rbRequest(t, db, opts, true), opts); err != nil {
		t.Fatalf("rebind: %v", err)
	}
	return db
}

func TestStorageWritesRefuseAConsolidatedSource(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name string
		call func(db *FactoryDB) error
	}{
		{"RecordPicked of a brand-new card", func(db *FactoryDB) error {
			_, err := db.RecordPicked(ctx, "run-old", "c-new", CardFields{}, "dispatch", now)
			return err
		}},
		{"RecordPicked of a card the consolidation moved away", func(db *FactoryDB) error {
			_, err := db.RecordPicked(ctx, "run-old", "c1", CardFields{}, "dispatch", now)
			return err
		}},
		{"RecordPickedReconciled creating a row", func(db *FactoryDB) error {
			_, err := db.RecordPickedReconciled(ctx, "run-old", "c-new", CardFields{}, "lane-1", now, rbOpts(), nil)
			return err
		}},
		{"RecordDispatchBinding", func(db *FactoryDB) error {
			return db.RecordDispatchBinding(ctx, "c1", "run-old", now)
		}},
		{"RecordDispatchEngagement", func(db *FactoryDB) error {
			return db.RecordDispatchEngagement(ctx, "c1", "run-old", "worker-1", nil, now)
		}},
		{"RecordDispatchEngagement, binding only", func(db *FactoryDB) error {
			return db.RecordDispatchEngagement(ctx, "c1", "run-old", "", nil, now)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := rbConsolidated(t)
			before := rbDump(t, db)
			if err := tc.call(db); !errors.Is(err, ErrRunRetiredByConsolidation) {
				t.Fatalf("err = %v, want ErrRunRetiredByConsolidation", err)
			}
			if after := rbDump(t, db); after != before {
				t.Fatalf("dump changed after the refusal:\n--- before\n%s\n--- after\n%s", before, after)
			}
		})
	}
}

// What stays allowed: reading a row the retired run still holds (c5, done), and
// the whole of an ORDINARY retirement — the marker needs rebound_to.
func TestStorageGuardsLeaveOrdinaryRetirementAndReadsAlone(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC)
	t.Run("RecordPicked of a row the retired run still holds returns it", func(t *testing.T) {
		db := rbConsolidated(t)
		c, err := db.RecordPicked(ctx, "run-old", "c5", CardFields{}, "dispatch", now)
		if err != nil || c.State != CardDone {
			t.Fatalf("RecordPicked(existing c5) = (%+v, %v), want the done row and no error", c, err)
		}
	})
	t.Run("an ordinarily retired run still accepts the writes", func(t *testing.T) {
		db := openSandboxFactory(t)
		opts := rbOpts()
		rbRun(t, db, "run-plain", rbPIDDead) // no cards: the reconciler retires it
		if _, err := db.ReconcileActiveRuns(ctx, opts); err != nil {
			t.Fatal(err)
		}
		if got := runStatus(t, db, "run-plain"); got != "retired" {
			t.Fatalf("fixture: status = %s, want retired", got)
		}
		if _, err := db.RecordPicked(ctx, "run-plain", "c1", CardFields{}, "dispatch", now); err != nil {
			t.Fatalf("RecordPicked in an ordinarily retired run: %v", err)
		}
		if err := db.RecordDispatchBinding(ctx, "c1", "run-plain", now); err != nil {
			t.Fatalf("RecordDispatchBinding to an ordinarily retired run: %v", err)
		}
	})
	t.Run("the consolidation itself, an abandon and the lease path are not blocked", func(t *testing.T) {
		db, opts := rbBase(t)
		// The consolidation re-keys and repoints bindings with UPDATEs and abandons
		// c4 with a pointer: none of it goes through the guarded writes.
		if _, err := db.RebindRun(ctx, rbRequest(t, db, opts, true), opts); err != nil {
			t.Fatalf("rebind: %v", err)
		}
		// The lease path records into the LIVE target as before.
		if _, err := db.RecordPickedReconciled(ctx, "run-new", "c-lease", CardFields{}, "lane-1", now, opts, nil); err != nil {
			t.Fatalf("lease into the surviving run: %v", err)
		}
	})
}
