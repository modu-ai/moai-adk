package cli

// Card t1708 follow-up — a delta audit must not satisfy the full-review Stop
// gate. The gate reads audit-multi/<session>.json (loadConvergenceResult) and
// blocks on OverallVerdict == fail without checking scope, so a delta-only pass
// that overwrote that file after a failed full review would open the gate.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func runSessionAudit(t *testing.T, root, session, deltaBase, codexVerdict string) ConvergenceResult {
	t.Helper()
	prev := backendCall
	backendCall = func(_ context.Context, backend, _, _, _ string) ReviewOutput {
		if backend == BackendCodex {
			return ReviewOutput{Verdict: codexVerdict, Summary: "codex:" + codexVerdict, Findings: []Finding{}, NextSteps: []string{}}
		}
		return claudeReview("pass")
	}
	defer func() { backendCall = prev }()
	return runMultiAudit(context.Background(), ReviewOutput{}, codexTargetBaseBranch, "", MultiAuditConfig{
		Gates:       config.AuditGates{Claude: config.AuditGateRequired, Codex: config.AuditGateRequired, GLM: config.AuditGateOff},
		SessionID:   session,
		ProjectRoot: root,
		DeltaBase:   deltaBase,
	}, nil)
}

func TestRunMultiAudit_DeltaBase_LeavesTheSessionStateFileAlone(t *testing.T) {
	root := newReviewTargetRepo(t)
	const session = "sess-delta"
	statePath := filepath.Join(root, ".moai", "state", "audit-multi", session+".json")

	runSessionAudit(t, root, session, "", "fail") // the full review fails
	before, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("the full review must persist its state (%v)", err)
	}
	if r, ok := loadConvergenceResult(root, session); !ok || r.OverallVerdict != "fail" {
		t.Fatalf("control: gate loader sees %+v ok=%v, want the full-review fail", r.OverallVerdict, ok)
	}

	delta := runSessionAudit(t, root, session, reviewTargetGitOut(t, root, "rev-parse", "HEAD"), "pass")

	after, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("state file vanished: %v", err)
	}
	if string(after) != string(before) {
		t.Errorf("a delta call overwrote the full-review state file the Stop gate reads:\nbefore: %s\nafter:  %s", before, after)
	}
	if r, ok := loadConvergenceResult(root, session); !ok || r.OverallVerdict != "fail" {
		t.Errorf("gate loader sees %q ok=%v after a delta pass; the full-review fail must stand", r.OverallVerdict, ok)
	}
	if delta.OverallVerdict != "pass" {
		t.Errorf("the delta result is still returned to the caller: OverallVerdict = %q, want pass", delta.OverallVerdict)
	}
	if !strings.Contains(delta.ResidualRiskNote, "delta_base") {
		t.Errorf("delta result must carry its scope note; residual_risk_note = %q", delta.ResidualRiskNote)
	}
}

// Control: a call without delta_base persists exactly as before.
func TestRunMultiAudit_NoDeltaBase_StillPersistsTheSessionState(t *testing.T) {
	root := newReviewTargetRepo(t)
	const session = "sess-full"

	runSessionAudit(t, root, session, "", "fail")
	runSessionAudit(t, root, session, "", "pass")

	if r, ok := loadConvergenceResult(root, session); !ok || r.OverallVerdict != "pass" {
		t.Errorf("a full call must overwrite the session state as today; gate loader sees %q ok=%v", r.OverallVerdict, ok)
	}
}
