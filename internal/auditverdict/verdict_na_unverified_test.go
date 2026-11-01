package auditverdict

import (
	"strings"
	"testing"
)

// The admission rule reads the machine fields of a plan-audit verdict, never the
// per-criterion lines of its prose body. These cases pin what that means for an
// applicable must-pass check the auditor could not run (card t2011):
//
//   - a body line saying N/A, FAIL, or "unverified" is not parsed, so the only
//     thing that keeps an unverified applicable check from admitting is the
//     must_pass_failed count the auditor writes;
//   - a genuine non-applicability (single-language SPEC) leaves the count at 0 and
//     admits.
//
// The consumer cannot tell a real N/A from an unrun applicable check; that
// discrimination lives in the auditor's instruction text, which a separate
// template test pins.
const planPassWithBody = `# SPEC Review Report
verdict: PASS
Overall Score: 0.90
must_pass_failed: 0
blocking_count: 0
plan_artifact_hash: abc123
audited_sha: 0123456789ab

## Must-Pass Results
- [PASS] MP-1 REQ number consistency: cited
- [N/A] MP-4 Section 22 language neutrality: N/A: single-language SPEC
- [N/A] MP-5 D7 cross-SPEC reconciliation: verb not executable
`

func TestAdmit_BodyCriterionLinesAreNotParsed(t *testing.T) {
	// A body line that says N/A for an applicable-but-unrun check does not stop
	// admission: nothing in the consumer looks at it.
	if ok, reason := admitPlan(planPassWithBody, true); !ok {
		t.Fatalf("a PASS verdict with must_pass_failed 0 and an N/A body line must admit (the consumer does not parse the body): %s", reason)
	}
	// Even a body line that says FAIL does not refuse by itself.
	failedBody := strings.Replace(planPassWithBody, "- [N/A] MP-5 D7 cross-SPEC reconciliation: verb not executable", "- [FAIL] MP-5 D7 cross-SPEC reconciliation: unverified, verb not run", 1)
	if ok, _ := admitPlan(failedBody, true); !ok {
		t.Fatal("the consumer reads only the machine fields; a FAIL body line with must_pass_failed 0 is not a refusal signal")
	}
}

func TestAdmit_UnverifiedApplicableCheckRefusesOnlyThroughTheCount(t *testing.T) {
	counted := strings.Replace(planPassWithBody, "must_pass_failed: 0", "must_pass_failed: 1", 1)
	counted = strings.Replace(counted, "verdict: PASS", "verdict: FAIL", 1)
	if ok, reason := admitPlan(counted, true); ok || !strings.Contains(reason, "verdict FAIL") {
		t.Fatalf("a FAIL verdict that counts the unverified check must refuse with the label reason, got admit=%v reason=%q", ok, reason)
	}
	// A PASS label with the unverified check counted is still refused, by the count.
	passCounted := strings.Replace(planPassWithBody, "must_pass_failed: 0", "must_pass_failed: 1", 1)
	if ok, reason := admitPlan(passCounted, true); ok || !strings.Contains(reason, "must_pass_failed 1") {
		t.Fatalf("a PASS label with must_pass_failed 1 must refuse with the count reason, got admit=%v reason=%q", ok, reason)
	}
}
