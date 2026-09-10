package cli

// required_gate_block_test.go — GH #1632 item 3, card t580: an audit gate the
// project EXPLICITLY configures `required` in workflow.audit.gates must fail
// the convergence overall_verdict when its backend returned no verdict
// (fail-open inconclusive). A gate the project never configured keeps the
// pre-existing annotate-only fail-open behavior byte-for-byte. The three
// branches the card names are the first three tests below; the rest pin the
// adjacent contracts the enforcement must not disturb.
//
// The discriminator is the RAW workflow.yaml value — the same one applyGateUnmet
// reads at the single-backend surface — never the engine's distributed default:
// the engine defaults codex to `required` when the key is absent, and treating
// that default as an explicit opt-in would flip every existing project to
// fail-closed.

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// gateBlockCaller returns a backendCall stub whose per-backend verdicts are
// fixed by the map; a backend missing from the map passes. Backends whose gate
// is off are never invoked, so they never reach the stub.
func gateBlockCaller(verdicts map[string]string) backendCallFn {
	return func(_ context.Context, backend, _, _, _ string) ReviewOutput {
		v, ok := verdicts[backend]
		if !ok {
			v = "pass"
		}
		return ReviewOutput{Verdict: v, Summary: backend + ": " + v, Findings: []Finding{}, NextSteps: []string{}}
	}
}

// withBackendCall swaps the package-level backendCall seam for the duration of
// the test (the established t.Cleanup pattern of this package's fan-out tests).
func withBackendCall(t *testing.T, fn backendCallFn) {
	t.Helper()
	orig := backendCall
	backendCall = fn
	t.Cleanup(func() { backendCall = orig })
}

// auditGatesDefault returns the effective gate map readGatesArgument produces
// when the caller omits the gates argument: claude+codex required,
// glm advisory.
func auditGatesDefault() config.AuditGates {
	return config.AuditGates{
		Claude: config.AuditGateRequired,
		Codex:  config.AuditGateRequired,
		GLM:    config.AuditGateAdvisory,
	}
}

// claudePassAnchor is the in-session claude verdict every branch starts from —
// a passing anchor is exactly what the fail-open fall-through used to absorb
// an unmet required gate into an overall pass.
func claudePassAnchor() ReviewOutput {
	return ReviewOutput{Verdict: "pass", Summary: "claude: pass", Findings: []Finding{}, NextSteps: []string{}}
}

// seedWorkflowGates writes a workflow.audit.gates block with the given
// backend→token map into the named tree's workflow.yaml, the config surface
// GH #1632 item 3 reports as unenforced.
func seedWorkflowGates(t *testing.T, root string, gates map[string]string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	keys := make([]string, 0, len(gates))
	for k := range gates {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	body := "workflow:\n  audit:\n    gates:\n"
	for _, k := range keys {
		body += "      " + k + ": " + gates[k] + "\n"
	}
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

// Branch 1 — the new contract: workflow.audit.gates.codex is EXPLICITLY
// `required` and codex returns no verdict → overall_verdict FAILS even though
// the claude anchor passed. The enforcement moves ONLY the overall verdict and
// the note: codex's own entry keeps its true inconclusive verdict and its
// fail_open_backends listing, so the audit trail still says the backend never
// ran.
func TestRunMultiAudit_ExplicitRequiredGateUnmet_FailsOverall(t *testing.T) {
	root := newProbeProject(t, "SPEC-REQGATEUNMET-901")
	seedWorkflowGates(t, root, map[string]string{"codex": config.AuditGateRequired})
	withBackendCall(t, gateBlockCaller(map[string]string{BackendCodex: VerdictInconclusive}))

	r := runMultiAudit(context.Background(), claudePassAnchor(), "uncommittedChanges", "", MultiAuditConfig{
		Gates:       auditGatesDefault(),
		ProjectRoot: root,
	}, nil)

	if r.OverallVerdict != overallVerdictFail {
		t.Errorf("overall_verdict = %q, want %q — an explicitly-required gate left unmet must fail the overall verdict, not ride the claude anchor to pass", r.OverallVerdict, overallVerdictFail)
	}
	codexVerdict := ""
	for _, v := range r.PerBackendVerdicts {
		if v.Backend == BackendCodex {
			codexVerdict = v.Verdict
		}
	}
	if codexVerdict != VerdictInconclusive {
		t.Errorf("codex per-backend verdict = %q, want %q — the enforcement must not rewrite the backend's own verdict", codexVerdict, VerdictInconclusive)
	}
	open := false
	for _, b := range r.FailOpenBackends {
		if b == BackendCodex {
			open = true
		}
	}
	if !open {
		t.Errorf("fail_open_backends = %v, want codex still named — the fail-open listing is the audit trail", r.FailOpenBackends)
	}
	if !strings.Contains(r.ResidualRiskNote, BackendCodex) {
		t.Errorf("residual_risk_note = %q, want it to name codex as the unmet required gate", r.ResidualRiskNote)
	}
}

// Branch 2 — gate met: the same explicit `required` config with codex actually
// producing a verdict passes. A satisfied gate must not leak the enforcement
// into a healthy convergence.
func TestRunMultiAudit_ExplicitRequiredGateMet_Passes(t *testing.T) {
	root := newProbeProject(t, "SPEC-REQGATEMET-902")
	seedWorkflowGates(t, root, map[string]string{"codex": config.AuditGateRequired})
	withBackendCall(t, gateBlockCaller(map[string]string{BackendCodex: "pass"}))

	r := runMultiAudit(context.Background(), claudePassAnchor(), "uncommittedChanges", "", MultiAuditConfig{
		Gates:       auditGatesDefault(),
		ProjectRoot: root,
	}, nil)

	if r.OverallVerdict != overallVerdictPass {
		t.Errorf("overall_verdict = %q, want %q — a required gate satisfied by a real verdict passes", r.OverallVerdict, overallVerdictPass)
	}
	if strings.Contains(r.ResidualRiskNote, "unmet") {
		t.Errorf("residual_risk_note = %q, want no unmet-gate claim on a satisfied gate", r.ResidualRiskNote)
	}
}

// Branch 2b — the explicit-required gate FAILS: overall fails through the
// pre-existing required-FAIL contract, and the enforcement adds no unmet-gate
// claim on top (the backend ran and gave its verdict — the gate is met).
func TestRunMultiAudit_ExplicitRequiredGateFail_BlockedByExistingContract(t *testing.T) {
	root := newProbeProject(t, "SPEC-REQGATEFAIL-903")
	seedWorkflowGates(t, root, map[string]string{"codex": config.AuditGateRequired})
	withBackendCall(t, gateBlockCaller(map[string]string{BackendCodex: "fail"}))

	r := runMultiAudit(context.Background(), claudePassAnchor(), "uncommittedChanges", "", MultiAuditConfig{
		Gates:       auditGatesDefault(),
		ProjectRoot: root,
	}, nil)

	if r.OverallVerdict != overallVerdictFail {
		t.Errorf("overall_verdict = %q, want %q — a required FAIL blocks on the existing contract", r.OverallVerdict, overallVerdictFail)
	}
	if strings.Contains(r.ResidualRiskNote, "unmet") {
		t.Errorf("residual_risk_note = %q, want no unmet-gate claim — a real fail verdict satisfies the gate", r.ResidualRiskNote)
	}
}

// Branch 3 — the regression guard: NO gates block in workflow.yaml (the common
// project; the distributed template ships none). The same fail-open
// inconclusive must still produce overall pass — the engine's default codex
// gate is `required`, but a default is not an explicit configuration, and
// flipping it would change behavior for every existing project.
func TestRunMultiAudit_UnsetGate_KeepsAnnotateOnlyFailOpen(t *testing.T) {
	root := newProbeProject(t, "SPEC-REQGATEUNSET-904") // no workflow.yaml seeded
	withBackendCall(t, gateBlockCaller(map[string]string{BackendCodex: VerdictInconclusive}))

	r := runMultiAudit(context.Background(), claudePassAnchor(), "uncommittedChanges", "", MultiAuditConfig{
		Gates:       auditGatesDefault(),
		ProjectRoot: root,
	}, nil)

	if r.OverallVerdict != overallVerdictPass {
		t.Errorf("overall_verdict = %q, want %q — an UNSET gate keeps the fail-open behavior byte-for-byte (a default is not an explicit opt-in)", r.OverallVerdict, overallVerdictPass)
	}
	open := false
	for _, b := range r.FailOpenBackends {
		if b == BackendCodex {
			open = true
		}
	}
	if !open {
		t.Errorf("fail_open_backends = %v, want codex named — the annotate-only surface is unchanged for unset gates", r.FailOpenBackends)
	}
}

// Advisory stays advisory: an EXPLICITLY `advisory` codex gate left
// inconclusive never blocks, whatever the engine's default for codex is.
func TestRunMultiAudit_ExplicitAdvisoryGateUnmet_StillPasses(t *testing.T) {
	root := newProbeProject(t, "SPEC-REQGATEADV-905")
	seedWorkflowGates(t, root, map[string]string{"codex": config.AuditGateAdvisory})
	withBackendCall(t, gateBlockCaller(map[string]string{BackendCodex: VerdictInconclusive}))

	r := runMultiAudit(context.Background(), claudePassAnchor(), "uncommittedChanges", "", MultiAuditConfig{
		Gates:       auditGatesDefault(),
		ProjectRoot: root,
	}, nil)

	if r.OverallVerdict != overallVerdictPass {
		t.Errorf("overall_verdict = %q, want %q — an advisory gate never blocks, even explicitly configured", r.OverallVerdict, overallVerdictPass)
	}
}

// The claude column of the gates map is enforced symmetrically: an explicitly
// required claude anchor that returns no verdict fails the overall verdict.
// (Pre-change, an inconclusive claude anchor propagated verbatim into
// overall_verdict — a value outside the declared {pass, fail} set.)
func TestRunMultiAudit_ExplicitRequiredClaudeAnchorUnmet_FailsOverall(t *testing.T) {
	root := newProbeProject(t, "SPEC-REQGATECLAUSE-906")
	seedWorkflowGates(t, root, map[string]string{"claude": config.AuditGateRequired})
	withBackendCall(t, gateBlockCaller(nil)) // both secondaries pass

	claude := ReviewOutput{Verdict: VerdictInconclusive, Summary: "claude: no verdict", Findings: []Finding{}, NextSteps: []string{}}
	r := runMultiAudit(context.Background(), claude, "uncommittedChanges", "", MultiAuditConfig{
		Gates:       auditGatesDefault(),
		ProjectRoot: root,
	}, nil)

	if r.OverallVerdict != overallVerdictFail {
		t.Errorf("overall_verdict = %q, want %q — an explicitly-required claude anchor left unmet fails the overall verdict", r.OverallVerdict, overallVerdictFail)
	}
}

// The enforced verdict reaches the persisted state file the multi-review-gate
// Stop hook reads, so the block survives past the tool call (DQ-1 consumption).
func TestRunMultiAudit_ExplicitRequiredGateUnmet_PersistedResultCarriesFail(t *testing.T) {
	root := newProbeProject(t, "SPEC-REQGATEPERSIST-907")
	seedWorkflowGates(t, root, map[string]string{"codex": config.AuditGateRequired})
	withBackendCall(t, gateBlockCaller(map[string]string{BackendCodex: VerdictInconclusive}))

	const sessionID = "gate-unmet-session"
	runMultiAudit(context.Background(), claudePassAnchor(), "uncommittedChanges", "", MultiAuditConfig{
		Gates:       auditGatesDefault(),
		SessionID:   sessionID,
		ProjectRoot: root,
	}, nil)

	stored, ok := loadConvergenceResult(root, sessionID)
	if !ok {
		t.Fatal("no convergence state file at the named tree — the Stop hook would read nothing")
	}
	if stored.OverallVerdict != overallVerdictFail {
		t.Errorf("persisted overall_verdict = %q, want %q — the state file must carry the enforced verdict so the Stop hook blocks on it", stored.OverallVerdict, overallVerdictFail)
	}
}

// An explicitly-required gate in a DIFFERENT tree must not reach into this
// audit: the enforcement keys on the audited tree (project_root), the same
// tree-scoped rule applyGateUnmet follows at the single-backend surface.
func TestRunMultiAudit_GateIsTreeScoped(t *testing.T) {
	configured := newProbeProject(t, "SPEC-REQGATESCOPE-908")
	seedWorkflowGates(t, configured, map[string]string{"codex": config.AuditGateRequired})
	other := newProbeProject(t, "SPEC-REQGATESCOPE-909") // no gates block
	withBackendCall(t, gateBlockCaller(map[string]string{BackendCodex: VerdictInconclusive}))

	r := runMultiAudit(context.Background(), claudePassAnchor(), "uncommittedChanges", "", MultiAuditConfig{
		Gates:       auditGatesDefault(),
		ProjectRoot: other,
	}, nil)

	if r.OverallVerdict != overallVerdictPass {
		t.Errorf("overall_verdict = %q, want %q — a gate configured in another tree must not enforce here", r.OverallVerdict, overallVerdictPass)
	}
	_ = configured
}
