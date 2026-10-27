package homestate

import (
	"context"
	"errors"
	"testing"
	"time"
)

// factory_dispatch_repair_test.go — RepairDispatchAtomic (REQ-FCR-024
// atomicity, card t1963): the retired-run check and every repair write are one
// transaction, with the engagement semantics of RecordDispatchEngagement kept
// exactly (owner restamp, valid-lease refusal, binding upsert).

var rbRepairNow = time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC)

func TestRepairDispatchAtomicRecordsAssignsAndBinds(t *testing.T) {
	db := openSandboxFactory(t)
	rbRun(t, db, "run-x", rbPIDDead)
	if err := db.RepairDispatchAtomic(context.Background(), "c1", "run-x", "worker-1", nil, rbRepairNow, nil); err != nil {
		t.Fatalf("repair: %v", err)
	}
	c, err := db.LoadCard(context.Background(), "run-x", "c1")
	if err != nil {
		t.Fatal(err)
	}
	if c.State != CardAssigned || c.OwnerLabel != "worker-1" {
		t.Fatalf("row = %s owner %q, want assigned under worker-1", c.State, c.OwnerLabel)
	}
	if got := rbBound(t, db, "c1"); got != "run-x" {
		t.Fatalf("binding = %s, want run-x", got)
	}
	// Idempotent: a second repair of the same dispatch changes nothing.
	before := rbDump(t, db)
	if err := db.RepairDispatchAtomic(context.Background(), "c1", "run-x", "worker-1", nil, rbRepairNow, nil); err != nil {
		t.Fatalf("second repair: %v", err)
	}
	if after := rbDump(t, db); after != before {
		t.Fatalf("a repeated repair changed the store:\n--- before\n%s\n--- after\n%s", before, after)
	}
}

// All or nothing: a binding write the store refuses rolls the row creation and
// the picked -> assigned edge back too (the separate transactions left a picked
// row behind).
func TestRepairDispatchAtomicRollsBackEveryWrite(t *testing.T) {
	db := openSandboxFactory(t)
	rbRun(t, db, "run-x", rbPIDDead)
	before := rbDump(t, db)
	if _, err := db.DB.Exec(`CREATE TRIGGER t_fail BEFORE INSERT ON card_dispatch BEGIN SELECT RAISE(ABORT, 'injected binding failure'); END`); err != nil {
		t.Fatal(err)
	}
	if err := db.RepairDispatchAtomic(context.Background(), "c1", "run-x", "worker-1", nil, rbRepairNow, nil); err == nil {
		t.Fatal("the repair succeeded although the binding write was refused")
	}
	if _, err := db.LoadCard(context.Background(), "run-x", "c1"); !errors.Is(err, ErrCardNotFound) {
		t.Fatalf("LoadCard after the failed repair: err = %v, want ErrCardNotFound (no half-done row)", err)
	}
	if after := rbDump(t, db); after != before {
		t.Fatalf("dump changed after the failed repair")
	}
}

// RecordDispatchEngagement's lease refusal is preserved: a row that still holds
// a valid lease under another owner refuses the whole repair, binding included.
func TestRepairDispatchAtomicKeepsTheValidLeaseRefusal(t *testing.T) {
	db := openSandboxFactory(t)
	rbRun(t, db, "run-x", rbPIDDead)
	frPlace(t, db, Card{RunID: "run-x", CardID: "c1", State: CardLeased, OwnerLabel: "worker-2", LeaseHolder: "worker-2", LeaseExpiresAt: "2099-01-01T00:00:00Z", Version: 3})
	before := rbDump(t, db)
	err := db.RepairDispatchAtomic(context.Background(), "c1", "run-x", "worker-1", nil, rbRepairNow, nil)
	if !errors.Is(err, ErrLeaseHolder) {
		t.Fatalf("err = %v, want ErrLeaseHolder", err)
	}
	if after := rbDump(t, db); after != before {
		t.Fatalf("dump changed after the refused repair")
	}
}

// A row under a different owner with no valid lease is restamped to the
// dispatch's owner in the same commit that binds the card.
func TestRepairDispatchAtomicRestampsAStaleOwner(t *testing.T) {
	db := openSandboxFactory(t)
	rbRun(t, db, "run-x", rbPIDDead)
	frPlace(t, db, Card{RunID: "run-x", CardID: "c1", State: CardAssigned, OwnerLabel: "worker-2", Version: 3})
	if err := db.RepairDispatchAtomic(context.Background(), "c1", "run-x", "worker-1", nil, rbRepairNow, nil); err != nil {
		t.Fatalf("repair: %v", err)
	}
	c, err := db.LoadCard(context.Background(), "run-x", "c1")
	if err != nil {
		t.Fatal(err)
	}
	if c.OwnerLabel != "worker-1" || c.Version != 4 {
		t.Fatalf("row owner %q version %d, want worker-1 at version 4", c.OwnerLabel, c.Version)
	}
	if got := rbBound(t, db, "c1"); got != "run-x" {
		t.Fatalf("binding = %s, want run-x", got)
	}
}

// The retired-run refusal inside the transaction: nothing is written.
func TestRepairDispatchAtomicRefusesAConsolidatedRun(t *testing.T) {
	db, opts := rbBase(t)
	if _, err := db.RebindRun(context.Background(), rbRequest(t, db, opts, true), opts); err != nil {
		t.Fatalf("rebind: %v", err)
	}
	before := rbDump(t, db)
	err := db.RepairDispatchAtomic(context.Background(), "c1", "run-old", "worker-1", nil, rbRepairNow, nil)
	if !errors.Is(err, ErrRunRetiredByConsolidation) {
		t.Fatalf("err = %v, want ErrRunRetiredByConsolidation", err)
	}
	if after := rbDump(t, db); after != before {
		t.Fatalf("dump changed after the refusal")
	}
}

func TestRepairDispatchAtomicRejectsEmptyIDs(t *testing.T) {
	db := openSandboxFactory(t)
	for _, tc := range [][2]string{{"", "run-x"}, {"c1", " "}} {
		if err := db.RepairDispatchAtomic(context.Background(), tc[0], tc[1], "worker-1", nil, rbRepairNow, nil); !errors.Is(err, ErrInvalidCardInput) {
			t.Errorf("RepairDispatchAtomic(%q, %q) err = %v, want ErrInvalidCardInput", tc[0], tc[1], err)
		}
	}
}
