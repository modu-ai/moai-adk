package auditverdict

import (
	"strings"
	"testing"
)

// Card t1571: the receipt's own overall verdict is the exporting auditor's
// bottom line — a recorded "fail" refuses through the shared predicate (every
// site, the sync phase and the plan gate included), even when every backend
// line reads pass and the label reads PASS.
func TestAdmitRefusesFailedConvergenceOverall(t *testing.T) {
	raw := "verdict: PASS\naudited_sha: " + strings.Repeat("a", 40) + "\n" +
		"convergence_overall: fail\nrequired_backend: codex pass\n"
	if ok, reason := Admit(Parse([]byte(raw)), PhaseSync, 0, false); ok {
		t.Error("sync admission approved a convergence_overall: fail record")
	} else if !strings.Contains(reason, "overall fail") {
		t.Errorf("sync refusal names no overall fail: %q", reason)
	}
	if ok, _ := AdmitWithRequired(Parse([]byte(raw)), PhasePlan, 0.5, true, []string{"codex"}); ok {
		t.Error("plan admission with required backends approved a convergence_overall: fail record")
	}
}
