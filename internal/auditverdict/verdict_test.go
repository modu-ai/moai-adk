package auditverdict

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const planPass = `# SPEC Review Report
verdict: PASS
Overall Score: 0.90
must_pass_failed: 0
blocking_count: 0
plan_artifact_hash: abc123
audited_sha: 0123456789ab
`

const planDebt = `verdict: PASS-WITH-DEBT
overall_score: 0.88
must_pass_failed: 0
blocking_count: 0
plan_artifact_hash: abc123
debts:
- debt: D1 dispose_in=run plan row omits an edit
- debt: D2 dispose_in=sync trace is indirect
`

func admitPlan(raw string, hashOK bool) (bool, string) {
	return Admit(Parse([]byte(raw)), PhasePlan, 0.85, hashOK, nil)
}

func TestAdmit_PlanPhaseFixtureTable(t *testing.T) {
	cases := []struct {
		name   string
		raw    string
		hashOK bool
		want   bool
	}{
		{"PASS meeting every check", planPass, true, true},
		{"PASS-WITH-DEBT with debts", planDebt, true, true},
		{"PASS-WITH-DEBT without debts", strings.Replace(strings.SplitN(planDebt, "debts:", 2)[0], "", "", 1), true, false},
		{"blocking_count 1", strings.Replace(planPass, "blocking_count: 0", "blocking_count: 1", 1), true, false},
		{"score below threshold with a PASS label", strings.Replace(planPass, "0.90", "0.80", 1), true, false},
		{"must_pass_failed 1 with a PASS label", strings.Replace(planPass, "must_pass_failed: 0", "must_pass_failed: 1", 1), true, false},
		{"hash mismatch", planPass, false, false},
		{"FAIL", strings.Replace(planPass, "verdict: PASS", "verdict: FAIL", 1), true, false},
		{"INCONCLUSIVE", strings.Replace(planPass, "verdict: PASS", "verdict: INCONCLUSIVE", 1), true, false},
		{"BYPASSED", strings.Replace(planPass, "verdict: PASS", "verdict: BYPASSED", 1), true, false},
		{"absent verdict", strings.Replace(planPass, "verdict: PASS\n", "", 1), true, false},
		{"must_pass_failed field missing", strings.Replace(planPass, "must_pass_failed: 0\n", "", 1), true, false},
		{"blocking_count field missing", strings.Replace(planPass, "blocking_count: 0\n", "", 1), true, false},
		{"score missing", strings.Replace(planPass, "Overall Score: 0.90\n", "", 1), true, false},
		{"debt with an invalid dispose_in", strings.Replace(planDebt, "dispose_in=run", "dispose_in=later", 1), true, false},
	}
	for _, c := range cases {
		got, reason := admitPlan(c.raw, c.hashOK)
		if got != c.want {
			t.Errorf("%s: admit=%v (%s), want %v", c.name, got, reason, c.want)
		}
		if !got && reason == "" {
			t.Errorf("%s: refusal carries no reason", c.name)
		}
	}
}

func TestAdmit_SyncPhaseKeepsTheLabelOnlyCheck(t *testing.T) {
	cases := map[string]bool{
		"verdict: PASS\n":           true,
		"verdict: PASS-WITH-DEBT\n": true,
		"verdict: FAIL\n":           false,
		"verdict: INCONCLUSIVE\n":   false,
		"verdict: BYPASSED\n":       false,
		"no verdict line\n":         false,
	}
	for raw, want := range cases {
		got, _ := Admit(Parse([]byte(raw)), PhaseSync, 0, false, nil)
		if got != want {
			t.Errorf("sync %q: admit=%v want %v", raw, got, want)
		}
	}
}

func TestParse_ReadsEveryField(t *testing.T) {
	f := Parse([]byte(planDebt))
	if f.Label != "PASS-WITH-DEBT" || !f.ScoreOK || f.Score != 0.88 || !f.MustPassKnown || f.MustPassFailed != 0 ||
		!f.BlockingKnown || f.BlockingCount != 0 || f.PlanArtifactHash != "abc123" || len(f.Debts) != 2 {
		t.Fatalf("parsed %+v", f)
	}
	if f.Debts[0].ID != "D1" || f.Debts[0].DisposeIn != "run" || f.Debts[0].Description != "plan row omits an edit" {
		t.Fatalf("debt %+v", f.Debts[0])
	}
}

func TestPlanThreshold_ReadsTheSpecTier(t *testing.T) {
	dir := t.TempDir()
	if got := PlanThreshold(dir); got != 0.85 {
		t.Fatalf("absent spec.md threshold %v, want Tier L 0.85", got)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte("---\ntier: M\n---\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := PlanThreshold(dir); got != 0.80 {
		t.Fatalf("Tier M threshold %v, want 0.80", got)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte("tier: \"S\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := SpecTier(dir); got != "S" {
		t.Fatalf("quoted tier read as %q", got)
	}
}

// Sync-audit F1: a non-finite or out-of-range score must never pass the
// threshold comparison.
func TestAdmit_RefusesNonFiniteAndOutOfRangeScores(t *testing.T) {
	for _, s := range []string{"NaN", "nan", "+Inf", "-Inf", "Inf", "1.5", "-0.1"} {
		raw := strings.Replace(planPass, "Overall Score: 0.90", "Overall Score: "+s, 1)
		if ok, _ := admitPlan(raw, true); ok {
			t.Errorf("score %q admitted", s)
		}
	}
}

// Sync re-audit N1: a duplicated decision key makes the file inadmissible —
// never last-wins.
func TestAdmit_RefusesDuplicatedDecisionKeys(t *testing.T) {
	cases := map[string]string{
		"verdict PASS appended after FAIL":   strings.Replace(planPass, "verdict: PASS", "verdict: FAIL", 1) + "Verdict: PASS\n",
		"second, different must_pass_failed": strings.Replace(planPass, "must_pass_failed: 0", "must_pass_failed: 2", 1) + "must_pass_failed: 0\n",
		"second, different blocking_count":   planPass + "blocking_count: 1\n",
		"second, different score":            planPass + "overall_score: 0.95\n",
		"second, different plan hash":        planPass + "plan_artifact_hash: def456\n",
	}
	for name, raw := range cases {
		if ok, reason := admitPlan(raw, true); ok {
			t.Errorf("%s: admitted (%s)", name, reason)
		}
	}
	if ok, _ := Admit(Parse([]byte("verdict: FAIL\nverdict: PASS\n")), PhaseSync, 0, false, nil); ok {
		t.Errorf("sync phase: duplicated verdict admitted")
	}
}

// Sync round 3 B1: the plan-auditor report repeats keys legitimately (a
// `Verdict:` header plus a machine `verdict:` line). Equal duplicates are
// admitted; only disagreeing or unparseable duplicates refuse.
func TestParse_RealReportShapeEqualDuplicatesAdmitted(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "plan-audit-report.md"))
	if err != nil {
		t.Fatal(err)
	}
	f := Parse(raw)
	if len(f.DuplicateKeys) != 0 {
		t.Fatalf("equal duplicates in a real report flagged: %v", f.DuplicateKeys)
	}
	if f.Label != LabelPass {
		t.Fatalf("label %q", f.Label)
	}
	full := planPass + "Verdict: PASS\nOverall Score: 0.900\n"
	if ok, reason := admitPlan(full, true); !ok {
		t.Fatalf("equal duplicate verdict/score refused: %s", reason)
	}
	if ok, _ := admitPlan(strings.Replace(planPass, "verdict: PASS", "verdict: FAIL", 1)+"Verdict: PASS\n", true); ok {
		t.Fatalf("FAIL then appended PASS admitted")
	}
	if ok, _ := admitPlan(planPass+"overall_score: banana\n", true); ok {
		t.Fatalf("unparseable duplicate score admitted")
	}
}

func TestAdmitLabel(t *testing.T) {
	for label, want := range map[string]bool{"PASS": true, "PASS-WITH-DEBT": true, "FAIL": false, "": false, "BYPASSED": false} {
		if AdmitLabel(label) != want {
			t.Errorf("AdmitLabel(%q) != %v", label, want)
		}
	}
}
