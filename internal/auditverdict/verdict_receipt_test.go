package auditverdict

import (
	"strings"
	"testing"
)

// Receipt tests for SPEC-AUDIT-CEILING-001 (REQ-ACE-008..010): the
// convergence receipt lines a required-backend tree's verdict file carries,
// the refusal causes they add to the shared admission predicate, and the
// error-vs-empty gate-set resolution contract (D21).

// planPassReceipt is planPass plus the receipt lines a required-backend
// tree's export must carry.
const planPassReceipt = `# SPEC Review Report
verdict: PASS
Overall Score: 0.90
must_pass_failed: 0
blocking_count: 0
plan_artifact_hash: abc123
convergence_overall: pass
required_backend: claude pass
`

// admitPlanGates mirrors admitPlan with the tree's required-backend set.
func admitPlanGates(raw string, hashOK bool, required []string) (bool, string) {
	return Admit(Parse([]byte(raw)), PhasePlan, 0.85, hashOK, required)
}

// TestParseReceipt covers REQ-ACE-008's Parse arm: the receipt keys are read,
// the single-occurrence and per-backend repeat rules hold (design.md §3,
// D18), and malformed or duplicated receipt lines are inadmissible
// (acceptance.md §C edges 2 and 3).
func TestParseReceipt(t *testing.T) {
	f := Parse([]byte(planPassReceipt))
	if !f.Receipt.Present || f.Receipt.ConvergenceOverall != "pass" {
		t.Fatalf("receipt %+v", f.Receipt)
	}
	if v := f.Receipt.Backends["claude"]; v != "pass" {
		t.Fatalf("backend claude=%q, want pass", v)
	}
	if len(f.DuplicateKeys) != 0 || f.MalformedReceipts != 0 {
		t.Fatalf("clean receipt flagged: dups=%v malformed=%d", f.DuplicateKeys, f.MalformedReceipts)
	}

	// Case-insensitive keys with space-fold, matching the parser convention.
	f = Parse([]byte("Convergence Overall: FAIL\nRequired Backend: codex inconclusive\n"))
	if f.Receipt.ConvergenceOverall != "fail" || f.Receipt.Backends["codex"] != "inconclusive" {
		t.Fatalf("case/space-folded keys parsed %+v", f.Receipt)
	}

	// Edge 2: a second, differing convergence_overall is a duplicate key —
	// never last-wins.
	f = Parse([]byte(planPassReceipt + "convergence_overall: fail\n"))
	if len(f.DuplicateKeys) == 0 {
		t.Fatal("differing duplicate convergence_overall admitted")
	}

	// An identical repeat (report header plus its machine line) stays
	// admissible.
	f = Parse([]byte(planPassReceipt + "Convergence Overall: pass\n"))
	if len(f.DuplicateKeys) != 0 {
		t.Fatalf("identical duplicate convergence_overall flagged: %v", f.DuplicateKeys)
	}

	// A second line for the same backend is a duplicate only when the verdict
	// differs (design.md §3 per-key repeat rule).
	f = Parse([]byte(planPassReceipt + "required_backend: claude pass\n"))
	if len(f.DuplicateKeys) != 0 || f.MalformedReceipts != 0 {
		t.Fatalf("identical backend repeat flagged: %v / %d", f.DuplicateKeys, f.MalformedReceipts)
	}
	f = Parse([]byte(planPassReceipt + "required_backend: claude fail\n"))
	if len(f.DuplicateKeys) == 0 {
		t.Fatal("differing backend repeat admitted")
	}

	// Edge 3: a malformed required_backend line (empty or unparseable verdict
	// value) is inadmissible.
	for _, bad := range []string{"required_backend: claude", "required_backend: claude yes", "required_backend:"} {
		f = Parse([]byte(planPassReceipt + bad + "\n"))
		if f.MalformedReceipts == 0 {
			t.Fatalf("malformed line %q not flagged", bad)
		}
	}

	// A value outside the receipt vocabulary stays malformed — the projection
	// exists so the producer never emits one, not so the parser accepts one.
	f = Parse([]byte(planPassReceipt + "required_backend: claude FAIL\n"))
	if f.MalformedReceipts == 0 {
		t.Fatal("out-of-vocabulary backend verdict accepted")
	}
}

// TestParseReceiptPassWithDebtProjection covers the PASS-WITH-DEBT
// single-model receipt end to end (V4-D2): the receipt projects the own label
// per design.md §3 — PASS-WITH-DEBT is a passing label per AdmitLabel, so the
// receipt reads pass — while the raw own label stays readable in the verdict
// body.
func TestParseReceiptPassWithDebtProjection(t *testing.T) {
	const singleModelDebt = `verdict: PASS-WITH-DEBT
overall_score: 0.88
must_pass_failed: 0
blocking_count: 0
plan_artifact_hash: abc123
- debt: D1 dispose_in=run fixture debt
convergence_overall: pass
required_backend: claude pass
`
	f := Parse([]byte(singleModelDebt))
	if f.Label != LabelPassWithDebt {
		t.Fatalf("raw own label %q not preserved", f.Label)
	}
	if f.Receipt.ConvergenceOverall != "pass" || f.Receipt.Backends["claude"] != "pass" {
		t.Fatalf("PASS-WITH-DEBT projection read %+v", f.Receipt)
	}
	if !AdmitLabel(f.Label) {
		t.Fatal("PASS-WITH-DEBT is not a passing label per AdmitLabel")
	}
}

// TestAdmitRequiredBackendFail covers REQ-ACE-009: a required backend
// recorded fail refuses regardless of the auditor's own PASS label, with a
// reason naming the backend.
func TestAdmitRequiredBackendFail(t *testing.T) {
	raw := strings.Replace(planPassReceipt, "required_backend: claude pass", "required_backend: claude fail", 1)
	ok, reason := admitPlanGates(raw, true, []string{"claude"})
	if ok {
		t.Fatal("PASS admitted over a required-backend fail receipt")
	}
	if !strings.Contains(reason, "claude") {
		t.Fatalf("reason %q does not name the backend", reason)
	}

	// The receipt refusal is not weakened by any other field: a fail receipt
	// with an unbound hash refuses too.
	if ok, _ := admitPlanGates(raw, false, []string{"claude"}); ok {
		t.Fatal("PASS admitted over a fail receipt with an unbound hash")
	}
}

// TestAdmitRequiredBackendInconclusive covers AC-ACE-019: a required backend
// recorded inconclusive refuses regardless of the auditor's own label.
func TestAdmitRequiredBackendInconclusive(t *testing.T) {
	raw := strings.Replace(planPassReceipt, "required_backend: claude pass", "required_backend: claude inconclusive", 1)
	ok, reason := admitPlanGates(raw, true, []string{"claude"})
	if ok {
		t.Fatal("PASS admitted over a required-backend inconclusive receipt")
	}
	if !strings.Contains(reason, "inconclusive") || !strings.Contains(reason, "claude") {
		t.Fatalf("reason %q does not name the backend and its verdict", reason)
	}
}

// TestAdmitReceiptAbsent covers REQ-ACE-010's first two arms: a required
// backend configured with no receipt at all refuses (fail-closed, Q4), and
// no required backend configured admits a receipt-less verdict (C4; edge 5).
func TestAdmitReceiptAbsent(t *testing.T) {
	ok, reason := admitPlanGates(planPass, true, []string{"claude"})
	if ok {
		t.Fatal("receipt-less verdict admitted on a required-backend tree")
	}
	if reason == "" || !strings.Contains(reason, "claude") {
		t.Fatalf("reason %q does not name the configured backend", reason)
	}
	if ok, reason := admitPlanGates(planPass, true, nil); !ok {
		t.Fatalf("receipt-less verdict refused on a gate-less tree: %s", reason)
	}
}

// TestAdmitRequiredBackendAbsent covers AC-ACE-018: a receipt that records
// the others but omits one configured backend's line refuses with a reason
// naming the missing backend.
func TestAdmitRequiredBackendAbsent(t *testing.T) {
	raw := strings.Replace(planPassReceipt, "required_backend: claude pass", "required_backend: codex pass", 1)
	ok, reason := admitPlanGates(raw, true, []string{"claude", "codex"})
	if ok {
		t.Fatal("receipt omitting a configured backend admitted")
	}
	if !strings.Contains(reason, "claude") {
		t.Fatalf("reason %q does not name the missing backend", reason)
	}
}
