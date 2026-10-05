package factory

import (
	"testing"
)

// card t1454 card-review r2c finding C4: the disposition recorder refuses a
// self-pair. subject==related made the two Names tests identical, so the
// pair matched EVERY finding naming the card — `relate t1 t1 --disposition
// reject` re-stamped the t1·t2 and t1·t3 findings alike. The exact pair,
// either endpoint order, still records.
func TestRecordFindingDispositionRefusesSelfPair(t *testing.T) {
	_, store := todoLikeFixture(t)
	if err := store.Mutate(func(rec *BacklogRecord) error {
		rec.Findings = append(rec.Findings, BacklogFinding{
			SubjectID: "t1", RelatedID: "t3", Relation: "blocks", Source: BacklogSourceAgent,
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	if err := RecordFindingDisposition(store, "t1", "t1", "reject"); err == nil {
		t.Fatal("a self-pair disposition was accepted — subject==related matches every finding naming the card")
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range rec.Findings {
		if f.Disposition != nil {
			t.Fatalf("the refused self-pair still stamped %q on the %s/%s finding", *f.Disposition, f.SubjectID, f.RelatedID)
		}
	}
	// The exact pair in the other endpoint order still records.
	if err := RecordFindingDisposition(store, "t3", "t1", "reject"); err != nil {
		t.Fatalf("the exact pair in the other order was refused: %v", err)
	}
	rec, err = store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	stamped := 0
	for _, f := range rec.Findings {
		if f.Disposition != nil {
			stamped++
		}
	}
	if stamped != 1 {
		t.Fatalf("exact-pair recording stamped %d findings, want 1", stamped)
	}
}
