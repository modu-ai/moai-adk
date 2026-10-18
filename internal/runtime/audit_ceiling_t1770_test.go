package runtime

import (
	"strings"
	"testing"
)

// TestCeilingHoldReasonAdvertisesNoInertReleasePath (card t1770) — the hold
// refusal must not advertise an operator decision recorded in progress.md §G
// as a release path: no admission code reads a §G override record, so the
// advertised path is mechanically inert. The split/new-SPEC route stays named.
func TestCeilingHoldReasonAdvertisesNoInertReleasePath(t *testing.T) {
	f := newCeilingFixture(t)
	path := f.writeIter(t, 1, "FAIL")
	f.appendLine(t, path, "must_pass_failed: 1")
	fields, hashOK := f.fieldsFromHash(t, path)
	oc, _, err := EvaluateCeiling(f.input(), fields, hashOK, nil)
	if err != nil {
		t.Fatal(err)
	}
	if oc == nil || oc.Outcome != OutcomeHold {
		t.Fatalf("outcome %+v, want hold", oc)
	}
	joined := strings.Join(oc.Reasons, " ")
	for _, banned := range []string{"operator decision", "progress.md §G"} {
		if strings.Contains(joined, banned) {
			t.Errorf("hold reason advertises the inert path %q: %s", banned, joined)
		}
	}
	if !strings.Contains(joined, "split/new-SPEC route") {
		t.Errorf("hold reason no longer names the real release path: %s", joined)
	}
}

// TestPlanAuditCeilingRequiredBackendFailNeverDebtProceeds (card t1770) — a
// PASS-WITH-DEBT verdict whose receipt records a required backend as fail or
// inconclusive must not be admitted or recorded as debt-proceed by the
// ceiling evaluation entry (the path `moai spec ceiling` runs).
func TestPlanAuditCeilingRequiredBackendFailNeverDebtProceeds(t *testing.T) {
	t.Parallel()
	for _, v := range []string{"fail", "inconclusive"} {
		t.Run(v, func(t *testing.T) {
			t.Parallel()
			body := ceilingDebtBody + "convergence_overall: pass\nrequired_backend: claude pass\nrequired_backend: codex " + v + "\n"
			outcome, hit, err := EvaluatePlanAuditCeiling(ceilingInput("SPEC-CEIL-T1770", ceilingEvidenceFixture(t, body), "hold-and-split"))
			if err != nil {
				t.Fatalf("EvaluatePlanAuditCeiling: %v", err)
			}
			if !hit {
				t.Fatal("expected the ceiling to apply")
			}
			if outcome.Disposition == CeilingDispositionDebtProceed || outcome.VerdictAdmitted {
				t.Fatalf("required backend codex %s yielded a pass-family outcome: %+v", v, outcome)
			}
		})
	}
}
