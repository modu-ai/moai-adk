package homestate

import (
	"context"
	"errors"
	"testing"
)

// Store-level coverage for the bundle columns (SPEC-TODO-CARD-ISSUANCE-001
// REQ-TCI-018): the schema-6 migration and the fresh DDL both serve
// bundle_id/bundle_order, and the fields round-trip through RecordPicked and
// SELECT. The command-level acceptance tests in internal/cli exercise the
// bundle chain end to end; this pins the store contract on its own.

func intp(i int) *int { return &i }

func TestFR_RecordBundleFieldsRoundTrip(t *testing.T) {
	db := frOpen(t)
	ctx := context.Background()
	c, err := db.RecordPicked(ctx, frRun, "b1", CardFields{BundleID: strp("bundle-a"), BundleOrder: intp(2)}, "bundle", frNow)
	if err != nil || c.State != CardPicked || c.BundleID != "bundle-a" || c.BundleOrder != 2 {
		t.Fatalf("create = %+v err=%v", c, err)
	}
	// The columns round-trip through SELECT: a fresh process would read the
	// row back through loadCard, so assert on the loaded copy.
	loaded, err := db.LoadCard(ctx, frRun, "b1")
	if err != nil || loaded.BundleID != "bundle-a" || loaded.BundleOrder != 2 {
		t.Fatalf("load = %+v err=%v", loaded, err)
	}
	// A card with no bundle fields reads as not-a-bundle-member.
	plain, err := db.RecordPicked(ctx, frRun, "b0", CardFields{}, "assign", frNow)
	if err != nil || plain.BundleID != "" || plain.BundleOrder != 0 {
		t.Fatalf("plain = %+v err=%v", plain, err)
	}
	// A negative member order is not a sequence position.
	if _, err := db.RecordPicked(ctx, frRun, "b2", CardFields{BundleID: strp("bundle-a"), BundleOrder: intp(-1)}, "bundle", frNow); !errors.Is(err, ErrInvalidCardInput) {
		t.Fatalf("negative order: err = %v, want ErrInvalidCardInput", err)
	}
}
