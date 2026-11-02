package auditverdict

import (
	"strings"
	"testing"
)

// Card t2011, audit sync-audit-1 defect D2 — the sync phase must keep admitting
// the stored report shapes it admitted before the fail-closed change.
//
// Measured on the stored reports (read-only scan of the primary checkout's
// .moai/reports: label-passing sync reports with no duplicate key, backend
// fail, or convergence fail): the strict score rule newly refused 133 of 2043
// for a score written on the 0-100 scale or with prose after it ("Score:
// 90/100 (harmonic mean ...)", "Overall Score: 92%", "Score: 90.9 (harmonic
// mean)", "Score: 0.93 (Tier S ...)"), and 2 more for must_pass_failed written
// as an empty list ("[]", "none"). The sync-auditor authoring contract writes
// the {n}/100 form and was not changed, so the reader accepts it.
//
// The plan phase is NOT relaxed: it compares the score with a threshold, so it
// keeps reading only a bare number in [0, 1].

func TestAdmitSync_LegacyScoreShapesAdmit(t *testing.T) {
	for _, score := range []string{
		"90/100",
		"90/100 (harmonic mean of Functionality 93, Security 92, Craft 84, Consistency 92)",
		"88/100 (weighted harmonic mean; weighted arithmetic 87.6)",
		"92%",
		"90.9 (harmonic mean)",
		"0.93 (Tier S PASS threshold 0.75)",
		"1.0 (round-8 re-audit)",
		// A landing-confirmation round that states it did not re-score.
		"none (landing-confirmation round; no re-score)",
	} {
		for _, key := range []string{"Score", "Overall Score", "overall_score"} {
			raw := "verdict: PASS\n" + key + ": " + score + "\nmust_pass_failed: 0\nblocking_count: 0\n"
			if ok, reason := admitSync(raw); !ok {
				t.Errorf("%s: %s -> refused (%s), want admitted (stored legacy shape)", key, score, reason)
			}
		}
	}
}

func TestAdmitSync_ScoreGarbageStillRefuses(t *testing.T) {
	for _, score := range []string{
		"high", "NaN", "Inf", "-3", "-0.2", "101", "188", "120/100", "150%", "90/0", "90/50", "/100", "%", "nonexistent", "nope",
	} {
		raw := "verdict: PASS\nOverall Score: " + score + "\n"
		ok, reason := admitSync(raw)
		if ok {
			t.Errorf("score %q admitted, want refused", score)
			continue
		}
		if !strings.Contains(reason, "score") {
			t.Errorf("score %q: reason %q does not name the score", score, reason)
		}
	}
}

func TestAdmitSync_EmptyMustPassListMeansZero(t *testing.T) {
	for _, v := range []string{"[]", "none", "None"} {
		raw := "verdict: PASS\nmust_pass_failed: " + v + "\nblocking_count: 0\n"
		if ok, reason := admitSync(raw); !ok {
			t.Errorf("must_pass_failed: %s -> refused (%s), want admitted", v, reason)
		}
	}
	// A real count or an unreadable word still refuses; the empty-list forms
	// are an exact vocabulary, not a prefix match.
	for _, v := range []string{"2", "[1]", "none yet", "nothing", "two"} {
		raw := "verdict: PASS\nmust_pass_failed: " + v + "\n"
		if ok, _ := admitSync(raw); ok {
			t.Errorf("must_pass_failed: %s admitted, want refused", v)
		}
	}
}

func TestAdmitPlan_LegacyShapesStayRefused(t *testing.T) {
	// The relaxation is the sync phase only.
	for _, bad := range []string{"90/100", "92%", "90.9"} {
		raw := strings.Replace(planPass, "Overall Score: 0.90", "Overall Score: "+bad, 1)
		if ok, _ := admitPlan(raw, true); ok {
			t.Errorf("plan phase admitted score %q", bad)
		}
	}
	raw := strings.Replace(planPass, "must_pass_failed: 0", "must_pass_failed: none", 1)
	if ok, _ := admitPlan(raw, true); ok {
		t.Errorf("plan phase admitted must_pass_failed: none")
	}
}
