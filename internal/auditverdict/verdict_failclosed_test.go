package auditverdict

import (
	"strings"
	"testing"
)

// Card t2011 — the verdict predicate fails closed on decision fields it can read.
//
//   - sync phase: a PASS-family label with a non-zero must_pass_failed or
//     blocking_count, or with a decision field present but unparseable, is
//     refused. A verdict file that simply carries no such field (the legacy
//     sync shape) still admits on the label alone.
//   - a score outside [0, 1] never admits, in either phase.
//   - an applicable must-pass check recorded as unmet in the body's MP rows
//     (FAIL, any non-PASS/N/A state, or an N/A row that calls itself unverified)
//     refuses when it outnumbers the declared must_pass_failed.

func admitSync(raw string) (bool, string) {
	return Admit(Parse([]byte(raw)), PhaseSync, 0, false)
}

func TestAdmitSync_FieldsPresentFailClosed(t *testing.T) {
	cases := []struct {
		name       string
		raw        string
		want       bool
		reasonPart string
	}{
		{"legacy label only admits", "verdict: PASS\n", true, ""},
		{"legacy PASS-WITH-DEBT label only admits", "verdict: PASS-WITH-DEBT\n", true, ""},
		{"zero counts admit", "verdict: PASS\nmust_pass_failed: 0\nblocking_count: 0\nOverall Score: 0.9\n", true, ""},
		{"PASS with must_pass_failed 2", "verdict: PASS\nmust_pass_failed: 2\n", false, "must_pass_failed 2"},
		{"PASS with blocking_count 1", "verdict: PASS\nblocking_count: 1\n", false, "blocking_count 1"},
		{"PASS-WITH-DEBT with must_pass_failed 1", "verdict: PASS-WITH-DEBT\nmust_pass_failed: 1\n", false, "must_pass_failed 1"},
		{"must_pass_failed unparseable", "verdict: PASS\nmust_pass_failed: two\n", false, "must_pass_failed"},
		{"blocking_count unparseable", "verdict: PASS\nblocking_count: none\n", false, "blocking_count"},
		{"negative must_pass_failed", "verdict: PASS\nmust_pass_failed: -1\n", false, "must_pass_failed"},
		{"score 88 out of range", "verdict: PASS\nOverall Score: 88\n", false, "score"},
		{"score unparseable", "verdict: PASS\noverall_score: high\n", false, "score"},
		{"score exactly 1 admits", "verdict: PASS\noverall_score: 1\n", true, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, reason := admitSync(tc.raw)
			if ok != tc.want {
				t.Fatalf("admit = %v (reason %q), want %v", ok, reason, tc.want)
			}
			if !tc.want && !strings.Contains(reason, tc.reasonPart) {
				t.Fatalf("reason %q does not name %q", reason, tc.reasonPart)
			}
		})
	}
}

func TestAdmitPlan_ScoreOutsideUnitIntervalNeverAdmits(t *testing.T) {
	for _, bad := range []string{"88", "1.5", "-0.1", "NaN", "Inf"} {
		raw := strings.Replace(planPass, "Overall Score: 0.90", "Overall Score: "+bad, 1)
		ok, reason := admitPlan(raw, true)
		if ok {
			t.Fatalf("score %q admitted", bad)
		}
		if !strings.Contains(reason, "score") {
			t.Fatalf("score %q: reason %q does not name the score", bad, reason)
		}
	}
	if ok, reason := admitPlan(planPass, true); !ok {
		t.Fatalf("control: the unmodified PASS fixture must admit: %s", reason)
	}
}

const mpRowsBase = `verdict: PASS
Overall Score: 0.90
must_pass_failed: 0
blocking_count: 0
plan_artifact_hash: abc123
audited_sha: 0123456789ab

## Must-Pass Results
- [PASS] MP-1 REQ number consistency: cited
- [PASS] MP-2 EARS format compliance: cited
- [N/A] MP-4 Section 22 language neutrality: N/A: single-language SPEC
`

func TestAdmit_BodyMustPassRowsOutnumberingTheCountRefuse(t *testing.T) {
	withRow := func(row string) string { return mpRowsBase + row + "\n" }
	cases := []struct {
		name string
		raw  string
		want bool
	}{
		{"control: PASS and genuine N/A rows admit", mpRowsBase, true},
		{"FAIL row with count 0", withRow("- [FAIL] MP-5 D7 cross-SPEC reconciliation: FAIL (unverified), verb not run"), false},
		{"lowercase fail row", withRow("- [fail] mp-5 d7: not run"), false},
		{"star bullet FAIL row", withRow("* [FAIL] MP-6 D8: unverified"), false},
		{"unknown state row is unmet", withRow("- [UNVERIFIED] MP-7 clarification gate: plan.md absent"), false},
		{"N/A row calling itself unverified", withRow("- [N/A] MP-5 D7: unverified, verb not executable here"), false},
		{"N/A row naming NOT_RUN", withRow("- [N/A] MP-6 D8: NOT_RUN"), false},
		{"genuine N/A row without the unmet wording admits", withRow("- [N/A] MP-6 D8 cross-platform discipline: N/A: no cross-platform surface"), true},
		{"two distinct FAIL rows against count 1", strings.Replace(withRow("- [FAIL] MP-5 a: unverified\n- [FAIL] MP-6 b: unverified"), "must_pass_failed: 0", "must_pass_failed: 1", 1), false},
		{"the same MP id twice counts once", withRow("- [PASS] MP-8 first\n- [PASS] MP-8 again"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, reason := admitPlan(tc.raw, true)
			if ok != tc.want {
				t.Fatalf("admit = %v (reason %q), want %v", ok, reason, tc.want)
			}
		})
	}
}

func TestAdmit_BodyMustPassRowsAreCheckedInSyncToo(t *testing.T) {
	raw := "verdict: PASS\n- [FAIL] MP-5 D7: unverified\n"
	if ok, reason := admitSync(raw); ok || !strings.Contains(reason, "must-pass") {
		t.Fatalf("sync admitted a PASS label over an unmet MP row: ok=%v reason=%q", ok, reason)
	}
}
