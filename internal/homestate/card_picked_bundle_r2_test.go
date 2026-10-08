package homestate

import (
	"context"
	"testing"
)

// card t1454 card-review r2 P1-1: the review claimed updateCardRow's SET
// list omitting bundle_id/bundle_order makes a re-record drop the stored
// bundle fields. The mechanical reproduction REFUTES the consequence: an
// UPDATE that omits columns leaves them unchanged, so the re-recorded card
// keeps its bundle identity and position. What the omission did leave was a
// row writer whose SET list was not column-total against the INSERT — a
// future field-clearing write would have silently failed. The SET list now
// carries the same columns the INSERT does, and this test pins the observed
// preservation either way.
func TestFR_RecordPickedUpdateKeepsBundleFields(t *testing.T) {
	db := frOpen(t)
	ctx := context.Background()
	first, err := db.RecordPicked(ctx, frRun, "b7", CardFields{BundleID: strp("bundle-x"), BundleOrder: intp(1)}, "bundle", frNow)
	if err != nil || first.BundleID != "bundle-x" || first.BundleOrder != 1 {
		t.Fatalf("create = %+v err=%v", first, err)
	}
	// A second record on the still-picked member — the update path. The
	// fields the caller passes do not touch the bundle; the stored row must
	// keep what the first record wrote.
	spec := "SPEC-FIXTURE-001"
	if _, err := db.RecordPicked(ctx, frRun, "b7", CardFields{SpecID: &spec}, "assign", frNow); err != nil {
		t.Fatalf("re-record: %v", err)
	}
	got, err := db.LoadCard(ctx, frRun, "b7")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.BundleID != "bundle-x" || got.BundleOrder != 1 {
		t.Fatalf("after re-record: bundle=%q order=%d, want bundle-x/1", got.BundleID, got.BundleOrder)
	}
}
