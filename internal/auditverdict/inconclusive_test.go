package auditverdict

import (
	"strings"
	"testing"
)

// SPEC-AUDIT-ENGINE-CONVERGENCE-001 M4 RED, sub-step S2a — the shared admission
// rule reads an INCONCLUSIVE verdict file (AC-AEC-002b, first sentence;
// REQ-AEC-007; spec.md section G; card t2017).
//
// An audit that ended with no backend reporting a defect and a required leg
// unrun is exported with the label INCONCLUSIVE, `convergence_overall:
// inconclusive`, and one `required_backend: <backend> inconclusive <class>` line
// per unrun leg. The admission rule refuses it exactly as it refuses FAIL, with a
// reason naming each unrun backend and its class. AdmitLabel stays unchanged.
//
// All rows go through the existing Parse, Admit and AdmitWithRequired.

// inconclusiveFile is the exported file of an audit whose codex leg hit a quota.
const inconclusiveFile = `# plan audit
verdict: INCONCLUSIVE
Overall Score: 0.70
must_pass_failed: 0
blocking_count: 0
plan_artifact_hash: abc123
audited_sha: abcdef0123456789abcdef0123456789abcdef01
convergence_overall: inconclusive
required_backend: claude pass
required_backend: codex inconclusive quota
`

func admitInconclusive(raw string, phase Phase, required []string) (bool, string) {
	f := Parse([]byte(raw))
	if required == nil {
		return Admit(f, phase, 0.85, true)
	}
	return AdmitWithRequired(f, phase, 0.85, true, required)
}

// AC-AEC-002b — the parser accepts convergence_overall: inconclusive and a class
// after an inconclusive backend line, for every member of the closed set.
func TestParseInconclusiveReceipt_IsWellFormed(t *testing.T) {
	f := Parse([]byte(inconclusiveFile))
	if !f.Receipt.Present || f.Receipt.ConvergenceOverall != "inconclusive" {
		t.Errorf("convergence_overall: inconclusive not read: %+v", f.Receipt)
	}
	if f.Receipt.Backends["codex"] != "inconclusive" || f.Receipt.Backends["claude"] != "pass" {
		t.Errorf("backend verdicts not read: %+v", f.Receipt.Backends)
	}
	if f.MalformedReceipts != 0 || len(f.DuplicateKeys) != 0 {
		t.Errorf("a well-formed inconclusive receipt was flagged: malformed=%d dups=%v", f.MalformedReceipts, f.DuplicateKeys)
	}
	for _, class := range []string{"quota", "timeout", "auth", "unavailable", "other"} {
		g := Parse([]byte("convergence_overall: inconclusive\nrequired_backend: codex inconclusive " + class + "\n"))
		if g.MalformedReceipts != 0 || g.Receipt.Backends["codex"] != "inconclusive" {
			t.Errorf("class %q after inconclusive was rejected: malformed=%d backends=%v", class, g.MalformedReceipts, g.Receipt.Backends)
		}
	}
}

// AC-AEC-002b — a third token is allowed ONLY after inconclusive and only from the
// closed set: a class after pass or fail, an unknown class, and a fourth token
// are malformed receipt lines (the mutant "accept any third token" fails here).
func TestParseInconclusiveReceipt_ThirdTokenRules(t *testing.T) {
	for _, bad := range []string{
		"required_backend: codex pass quota",
		"required_backend: codex fail quota",
		"required_backend: codex inconclusive lunch",
		"required_backend: codex inconclusive Quota",
		"required_backend: codex inconclusive quota extra",
	} {
		f := Parse([]byte("convergence_overall: inconclusive\n" + bad + "\n"))
		if f.MalformedReceipts == 0 {
			t.Errorf("%q must be a malformed receipt line", bad)
		}
		if ok, _ := AdmitWithRequired(f, PhaseSync, 0, false, []string{"codex"}); ok {
			t.Errorf("%q admitted", bad)
		}
	}
	// Control (passes today): the pre-existing two-token forms stay well formed,
	// including an inconclusive line that carries no class.
	for _, good := range []string{"required_backend: codex pass", "required_backend: codex fail", "required_backend: codex inconclusive"} {
		if f := Parse([]byte("convergence_overall: pass\n" + good + "\n")); f.MalformedReceipts != 0 {
			t.Errorf("%q was flagged malformed", good)
		}
	}
}

// AC-AEC-002b — the admission rule refuses an INCONCLUSIVE file, in the plan and
// the sync phase, with a reason naming the unrun backend and its class.
func TestAdmitInconclusiveFile_RefusesNamingTheBackendAndItsClass(t *testing.T) {
	for _, phase := range []Phase{PhasePlan, PhaseSync} {
		for name, required := range map[string][]string{"Admit": nil, "AdmitWithRequired": {"claude", "codex"}} {
			ok, reason := admitInconclusive(inconclusiveFile, phase, required)
			if ok {
				t.Errorf("phase %d %s admitted an INCONCLUSIVE file", phase, name)
				continue
			}
			for _, want := range []string{"codex", "quota"} {
				if !strings.Contains(reason, want) {
					t.Errorf("phase %d %s reason %q does not name %q", phase, name, reason, want)
				}
			}
			if strings.Contains(reason, "overall fail") {
				t.Errorf("an unrun leg must not read as an overall fail: %q", reason)
			}
		}
	}
}

// AC-AEC-002b — every unrun backend is named with its own class.
func TestAdmitInconclusiveFile_NamesEachUnrunBackendAndClass(t *testing.T) {
	raw := strings.Replace(inconclusiveFile, "required_backend: claude pass", "required_backend: glm inconclusive timeout", 1)
	ok, reason := admitInconclusive(raw, PhaseSync, nil)
	if ok {
		t.Fatal("admitted")
	}
	for _, want := range []string{"codex", "quota", "glm", "timeout"} {
		if !strings.Contains(reason, want) {
			t.Errorf("reason %q does not name %q", reason, want)
		}
	}
}

// AC-AEC-002b — the inconclusive overall refuses regardless of the own label (the
// step precedes the label check), and a class-less inconclusive line is named too.
func TestAdmitInconclusiveOverall_RefusesWhateverTheLabel(t *testing.T) {
	passLabel := strings.Replace(inconclusiveFile, "verdict: INCONCLUSIVE", "verdict: PASS", 1)
	if ok, reason := admitInconclusive(passLabel, PhasePlan, nil); ok || !strings.Contains(reason, "codex") {
		t.Errorf("a PASS label over convergence_overall: inconclusive must refuse naming codex; ok=%v reason=%q", ok, reason)
	}
	noClass := strings.Replace(inconclusiveFile, "required_backend: codex inconclusive quota", "required_backend: codex inconclusive", 1)
	if ok, reason := admitInconclusive(noClass, PhaseSync, nil); ok || !strings.Contains(reason, "codex") {
		t.Errorf("an inconclusive line without a class must still refuse naming codex; ok=%v reason=%q", ok, reason)
	}
}

// AC-AEC-002b / spec.md section G — the admission order: duplicates, then
// required_backend_fail, then convergence_overall fail, then convergence_overall
// inconclusive. A recorded backend fail is reported before an unrun leg.
// Control: passes today (the earlier steps already exist).
func TestAdmitOrder_BackendFailAndOverallFailPrecedeTheInconclusiveStep(t *testing.T) {
	withFail := inconclusiveFile + "required_backend_fail: glm\n"
	if ok, reason := admitInconclusive(withFail, PhaseSync, nil); ok || !strings.Contains(reason, "glm") || !strings.Contains(reason, "recorded fail") {
		t.Errorf("required_backend_fail must be reported first; ok=%v reason=%q", ok, reason)
	}
	overallFail := strings.Replace(inconclusiveFile, "convergence_overall: inconclusive", "convergence_overall: fail", 1)
	if ok, reason := admitInconclusive(overallFail, PhaseSync, nil); ok || !strings.Contains(reason, "overall fail") {
		t.Errorf("convergence_overall: fail must still read as overall fail; ok=%v reason=%q", ok, reason)
	}
	dup := inconclusiveFile + "convergence_overall: pass\n"
	if ok, reason := admitInconclusive(dup, PhaseSync, nil); ok || !strings.Contains(reason, "duplicated") {
		t.Errorf("a duplicated decision key must be reported first; ok=%v reason=%q", ok, reason)
	}
}

// AC-AEC-002b — AdmitLabel stays unchanged: INCONCLUSIVE is not a passing label.
// Control: passes today.
func TestAdmitLabel_INCONCLUSIVEIsNotAPassingLabel(t *testing.T) {
	for label, want := range map[string]bool{"PASS": true, "PASS-WITH-DEBT": true, "FAIL": false, "INCONCLUSIVE": false, "": false} {
		if got := AdmitLabel(label); got != want {
			t.Errorf("AdmitLabel(%q) = %v, want %v", label, got, want)
		}
	}
}

// Control (passes today): an ordinary PASS receipt is still admitted, so the new
// steps cannot be satisfied by refusing everything.
func TestAdmitPassReceipt_StillAdmitted_Control(t *testing.T) {
	const pass = `verdict: PASS
Overall Score: 0.90
must_pass_failed: 0
blocking_count: 0
plan_artifact_hash: abc123
convergence_overall: pass
required_backend: claude pass
required_backend: codex pass
`
	if ok, reason := admitInconclusive(pass, PhasePlan, []string{"claude", "codex"}); !ok {
		t.Errorf("an all-pass receipt was refused: %s", reason)
	}
}
