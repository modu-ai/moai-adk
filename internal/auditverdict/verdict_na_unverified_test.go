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
	// An N/A body row whose text does not call the check unverified or not run
	// does not stop admission (see the CHANGED note below for the FAIL row).
	if ok, reason := admitPlan(planPassWithBody, true); !ok {
		t.Fatalf("a PASS verdict with must_pass_failed 0 and an N/A body line must admit (the consumer does not parse the body): %s", reason)
	}
	// CHANGED (card t2011, leader order (c)): a body MP row recorded FAIL while
	// must_pass_failed declares 0 used to admit because the body was ignored. The
	// predicate now reads the existing "- [STATE] MP-n" row format and refuses
	// when unmet rows outnumber the declared count. The [N/A] row above still
	// admits: a genuine non-applicability and an unrun applicable check that is
	// worded as N/A without the words "unverified" / "not run" remain
	// indistinguishable to the consumer — that residual is held by the
	// plan-auditor instruction text only.
	failedBody := strings.Replace(planPassWithBody, "- [N/A] MP-5 D7 cross-SPEC reconciliation: verb not executable", "- [FAIL] MP-5 D7 cross-SPEC reconciliation: unverified, verb not run", 1)
	if ok, reason := admitPlan(failedBody, true); ok || !strings.Contains(reason, "must-pass row") {
		t.Fatalf("a FAIL body row with must_pass_failed 0 must now refuse with the row reason, got admit=%v reason=%q", ok, reason)
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
