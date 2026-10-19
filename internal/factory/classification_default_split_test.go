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
