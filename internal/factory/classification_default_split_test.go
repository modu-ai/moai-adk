// classification_default_split_test.go — SPEC-FACTORY-DEFAULT-MODE-001 M1/M3
// acceptance tests on the default split: the absent-field read derivation
// produces parallelizable (AC-FDM-001, REQ-FDM-001) while the decider-failure
// fail-safe stays a separate serial value (REQ-FDM-003 — its cli-side guard
// is TestTodoAddDeciderFailureFailsafeSerial), and a recorded judgment is
// never rewritten by either default (AC-FDM-010).
package factory

import "testing"

// TestEffectiveClassificationAbsentDefaultsParallelizable — AC-FDM-001: an
// absent classification field derives mode parallelizable (priority normal,
// blocked false, decider default) through the read seam, with a reason that
// names the absence derivation. The fail-safe constructor stays a SEPARATE
// serial value in the same assertion: the two defaults are two values by the
// 2026-10-19 operator ruling (plan D2), and a silent re-merge fails here.
func TestEffectiveClassificationAbsentDefaultsParallelizable(t *testing.T) {
	got := EffectiveCardClassification(BacklogItem{ID: "t1"})
	want := CardClassification{
		Priority: ClassPriorityNormal,
		Blocked:  false,
		Mode:     ClassModeParallelizable,
		Decider:  DeciderIdentityDefault,
	}
	if got.Priority != want.Priority || got.Blocked != want.Blocked || got.Mode != want.Mode || got.Decider != want.Decider {
		t.Errorf("effective classification of an absent field = %+v, want %+v", got, want)
	}
	if got.Reason == "" {
		t.Errorf("absent-field derivation carries no reason; the derivation must name the absence")
	}
	if failsafe := DefaultCardClassification(); failsafe.Mode != ClassModeSerial {
		t.Errorf("fail-safe default mode = %q, want serial (the fail-safe is a separate value from the absent-read default)", failsafe.Mode)
	}
}

// TestRecordedClassificationNotRewritten — AC-FDM-010 (regression guard): a
// card carrying an explicit classification recorded by the old default
// decider (decider default, mode serial) reads verbatim through the
// derivation seam after the flip, and the store round-trip keeps it verbatim
// — no default, migration, or read path rewrites it; only the classify verb
// changes it (C3).
func TestRecordedClassificationNotRewritten(t *testing.T) {
	recorded := CardClassification{
		Priority:     ClassPriorityNormal,
		Blocked:      false,
		Mode:         ClassModeSerial,
		Decider:      DeciderIdentityDefault,
		ClassifiedAt: "2026-09-29T00:00:00Z",
		Reason:       "deterministic default: no classification judgment supplied",
	}
	if got := EffectiveCardClassification(BacklogItem{ID: "t29", Classification: &recorded}); got != recorded {
		t.Errorf("recorded serial-default judgment = %+v, want verbatim %+v", got, recorded)
	}
	store := NewBacklogStore(BacklogPathForRoot(t.TempDir()))
	if _, _, err := store.Add("old-default card"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := store.Mutate(func(rec *BacklogRecord) error {
		for i := range rec.Items {
			c := recorded
			rec.Items[i].Classification = &c
		}
		return nil
	}); err != nil {
		t.Fatalf("record: %v", err)
	}
	loaded, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Items[0].Classification == nil || *loaded.Items[0].Classification != recorded {
		t.Errorf("stored record = %+v, want the recorded judgment verbatim", loaded.Items[0].Classification)
	}
}
