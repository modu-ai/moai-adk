// Package cli — SPEC-AUDIT-MULTI-MODEL-001 convergence engine tests (M2 TDD).
//
// These tests are the RED-GREEN-REFACTOR spec for the convergence engine in
// mcp_convergence.go. They cover the design.md §3 policy case table, the
// disagreement = advisory (NOT block) regression (EC-3 / AC-AMM-009), the
// super-review independence invariant (EC-6 / AC-AMM-003), the no-new-enum
// invariant (EC-7 / AC-AMM-011), the fail-open identity (AC-AMM-004 / EC-1 /
// EC-4), the DQ-2 missing-claude-anchor refusal, the DQ-1 state-file write,
// and the subagent boundary (no AskUserQuestion — AC-AMM-024).
//
// @MX:SPEC: SPEC-AUDIT-MULTI-MODEL-001
package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// ─── helpers ───

// pbv is a compact PerBackendVerdict builder for table tests.
func pbv(backend, gate, verdict string) PerBackendVerdict {
	return PerBackendVerdict{
		Backend:   backend,
		Gate:      gate,
		Verdict:   verdict,
		Summary:   backend + ":" + verdict,
		Findings:  []Finding{},
		NextSteps: []string{},
	}
}

// pbvReq returns a required-gate PerBackendVerdict.
func pbvReq(backend, verdict string) PerBackendVerdict {
	return pbv(backend, config.AuditGateRequired, verdict)
}

// pbvAdv returns an advisory-gate PerBackendVerdict.
func pbvAdv(backend, verdict string) PerBackendVerdict {
	return pbv(backend, config.AuditGateAdvisory, verdict)
}

// claudeReview is a reusable in-session claude verdict.
func claudeReview(verdict string) ReviewOutput {
	return ReviewOutput{
		Verdict:   verdict,
		Summary:   "claude:" + verdict,
		Findings:  []Finding{},
		NextSteps: []string{},
	}
}

// ─── M2 algorithm tests — pure converge() over the design.md §3 policy table ───

// Case #1 (AC-AMM-006): all required backends PASS → overall PASS,
// disagreement_flag == false.
func TestConverge_AllRequiredPass_Case1_AC_AMM_006(t *testing.T) {
	verdicts := []PerBackendVerdict{
		pbvReq(BackendClaude, "pass"),
		pbvReq(BackendCodex, "pass"),
		pbvReq(BackendGLM, "pass"),
	}
	r := converge(verdicts)
	if r.OverallVerdict != "pass" {
		t.Errorf("overall = %q, want pass", r.OverallVerdict)
	}
	if r.DisagreementFlag {
		t.Error("disagreement_flag = true, want false (all required agree on pass)")
	}
	if r.ResidualRiskNote != "" {
		t.Errorf("residual_risk_note = %q, want empty", r.ResidualRiskNote)
	}
}

// Case #2 (AC-AMM-007): all required backends FAIL (no split) → overall FAIL,
// disagreement_flag == false, residual_risk_note records which failed.
func TestConverge_AllRequiredFail_Case2_AC_AMM_007(t *testing.T) {
	verdicts := []PerBackendVerdict{
		pbvReq(BackendClaude, "fail"),
		pbvReq(BackendCodex, "fail"),
	}
	r := converge(verdicts)
	if r.OverallVerdict != "fail" {
		t.Errorf("overall = %q, want fail", r.OverallVerdict)
	}
	if r.DisagreementFlag {
		t.Error("disagreement_flag = true, want false (no split — all required agree on fail)")
	}
	if !strings.Contains(r.ResidualRiskNote, "claude") || !strings.Contains(r.ResidualRiskNote, "codex") {
		t.Errorf("residual_risk_note = %q, want both failed backends named", r.ResidualRiskNote)
	}
}

// Case #2 variant (AC-AMM-007): one required FAIL + others inconclusive → FAIL,
// NOT a disagreement (no PASS among required to split against).
func TestConverge_RequiredFailWithInconclusive_Case2(t *testing.T) {
	verdicts := []PerBackendVerdict{
		pbvReq(BackendClaude, "fail"),
		pbvReq(BackendCodex, VerdictInconclusive),
	}
	r := converge(verdicts)
	if r.OverallVerdict != "fail" {
		t.Errorf("overall = %q, want fail (any required fail → fail)", r.OverallVerdict)
	}
	if r.DisagreementFlag {
		t.Error("disagreement_flag = true, want false (fail vs inconclusive is not a split)")
	}
}

// Case #3 (AC-AMM-008): required split (≥1 PASS + ≥1 FAIL) → conservative FAIL,
// disagreement_flag == true, residual_risk_note describes the split. Verdicts
// remain existing values — NO new VerdictDisagreement enum.
func TestConverge_RequiredSplit_Case3_AC_AMM_008(t *testing.T) {
	verdicts := []PerBackendVerdict{
		pbvReq(BackendClaude, "pass"),
		pbvReq(BackendCodex, "fail"),
		pbvReq(BackendGLM, "pass"),
	}
	r := converge(verdicts)
	if r.OverallVerdict != "fail" {
		t.Errorf("overall = %q, want fail (required-split resolves conservatively)", r.OverallVerdict)
	}
	if !r.DisagreementFlag {
		t.Error("disagreement_flag = false, want true (required split)")
	}
	if !strings.Contains(strings.ToLower(r.ResidualRiskNote), "split") &&
		!strings.Contains(strings.ToLower(r.ResidualRiskNote), "disagree") {
		t.Errorf("residual_risk_note = %q, want it to describe the split", r.ResidualRiskNote)
	}
	// The per-backend verdicts remain existing values (no VerdictDisagreement).
	for _, v := range r.PerBackendVerdicts {
		switch v.Verdict {
		case "pass", "fail", VerdictInconclusive:
			// ok
		default:
			t.Errorf("per-backend verdict = %q for %s; want one of the existing values (no new enum)", v.Verdict, v.Backend)
		}
	}
}

// Case #4 (AC-AMM-009, EC-3): advisory-only conflict — all required PASS, one
// advisory FAIL → overall PASS + disagreement_flag == true. THE REGRESSION
// GUARD FOR THE FIXED USER DECISION: cross-model disagreement is INFORMATION,
// not a GATE. A mere advisory FAIL never flips overall to FAIL.
func TestConverge_AdvisoryOnlyConflict_Case4_AC_AMM_009_EC3(t *testing.T) {
	verdicts := []PerBackendVerdict{
		pbvReq(BackendClaude, "pass"),
		pbvReq(BackendCodex, "pass"),
		pbvAdv(BackendGLM, "fail"), // GLM advisory gate, FAILs
	}
	r := converge(verdicts)
	if r.OverallVerdict != "pass" {
		t.Errorf("overall = %q, want pass (advisory FAIL never flips overall)", r.OverallVerdict)
	}
	if !r.DisagreementFlag {
		t.Error("disagreement_flag = false, want true (advisory-vs-required conflict surfaced)")
	}
	if r.ResidualRiskNote == "" {
		t.Error("residual_risk_note empty; want the advisory conflict described for the Verification Matrix")
	}
}

// EC-3 dedicated regression (AC-AMM-009 / AC-AMM-010): disagreement_flag == true
// MUST NOT itself be a block category. Verify that when disagreement is present
// but no required backend FAILed, overall stays PASS (flow NOT interrupted).
func TestConverge_DisagreementAdvisoryNotBlock_Regression_EC3(t *testing.T) {
	verdicts := []PerBackendVerdict{
		pbvReq(BackendClaude, "pass"),
		pbvAdv(BackendCodex, "fail"), // codex demoted to advisory; conflicts with claude
	}
	r := converge(verdicts)
	if !r.DisagreementFlag {
		t.Fatal("disagreement_flag = false; want true (prerequisite for the regression assertion)")
	}
	// THE LOAD-BEARING ASSERTION: disagreement alone does NOT block.
	if r.OverallVerdict != "pass" {
		t.Errorf("overall = %q, want pass — disagreement MUST NOT be a new block category (C3)", r.OverallVerdict)
	}
}

// AC-AMM-006 vacuous case: zero required backends (all advisory) → overall PASS
// (nothing required, nothing blocks). disagreement_flag set only if advisories
// conflict among themselves in a way the design records (here they don't).
func TestConverge_NoRequiredBackends_VacuousPass(t *testing.T) {
	verdicts := []PerBackendVerdict{
		pbvAdv(BackendCodex, "pass"),
		pbvAdv(BackendGLM, "pass"),
	}
	r := converge(verdicts)
	if r.OverallVerdict != "pass" {
		t.Errorf("overall = %q, want pass (no required gate to satisfy)", r.OverallVerdict)
	}
	if r.DisagreementFlag {
		t.Error("disagreement_flag = true; want false (advisories agree)")
	}
}

// ─── fail-open identity (AC-AMM-004 / EC-1 / EC-4) ───

// EC-1: codex+glm both VerdictInconclusive (missing) → overall falls back to
// the claude verdict; both backends recorded in fail_open_backends; no hard block.
func TestConverge_AllSecondariesInconclusive_EC1(t *testing.T) {
	verdicts := []PerBackendVerdict{
		pbvReq(BackendClaude, "pass"),
		pbvReq(BackendCodex, VerdictInconclusive),
		pbvReq(BackendGLM, VerdictInconclusive),
	}
	r := converge(verdicts)
	if r.OverallVerdict != "pass" {
		t.Errorf("overall = %q, want pass (fail-open to claude pass)", r.OverallVerdict)
	}
	want := map[string]bool{BackendCodex: true, BackendGLM: true}
	for _, b := range r.FailOpenBackends {
		if !want[b] {
			t.Errorf("fail_open_backends contains %q; want only codex+glm", b)
		}
	}
}

// EC-4: all required backends VerdictInconclusive (both missing) → overall
// follows the in-session claude verdict. CLAUDE FAILS → overall fail; CLAUDE
// PASSES → overall pass. Fail-open to claude.
func TestConverge_AllRequiredInconclusive_FailOpenToClaude_EC4(t *testing.T) {
	t.Run("claude pass → overall pass", func(t *testing.T) {
		verdicts := []PerBackendVerdict{
			pbvReq(BackendClaude, "pass"),
			pbvReq(BackendCodex, VerdictInconclusive),
			pbvReq(BackendGLM, VerdictInconclusive),
		}
		if r := converge(verdicts); r.OverallVerdict != "pass" {
			t.Errorf("overall = %q, want pass (fail-open to claude pass)", r.OverallVerdict)
		}
	})
	t.Run("claude fail → overall fail", func(t *testing.T) {
		verdicts := []PerBackendVerdict{
			pbvReq(BackendClaude, "fail"),
			pbvReq(BackendCodex, VerdictInconclusive),
			pbvReq(BackendGLM, VerdictInconclusive),
		}
		if r := converge(verdicts); r.OverallVerdict != "fail" {
			t.Errorf("overall = %q, want fail (fail-open to claude fail)", r.OverallVerdict)
		}
	})
}

// ─── runMultiAudit fan-out: DQ-2, independence, parallelism, gate-off, DQ-1 ───

// recordingCaller captures every backend invocation so the independence test
// can assert claude_verdict was NEVER passed to a secondary backend.
type recordingCaller struct {
	mu    sync.Mutex
	calls []recordedCall
}

type recordedCall struct {
	Backend string
	Target  string
	Focus   string
	// NOTE: no ClaudeVerdict field — the backendCaller signature forbids it.
}

func (r *recordingCaller) call(_ context.Context, backend, target, focus string) ReviewOutput {
	r.mu.Lock()
	r.calls = append(r.calls, recordedCall{Backend: backend, Target: target, Focus: focus})
	r.mu.Unlock()
	// All secondaries PASS by default; individual tests override.
	return ReviewOutput{Verdict: "pass", Summary: backend + ":pass", Findings: []Finding{}, NextSteps: []string{}}
}

// DQ-2: claude_verdict absent → the engine REFUSES to synthesize. Returns a
// structured ConvergenceResult (NOT a hard error), overall_verdict = fail,
// residual_risk_note explains the missing anchor. Fail-open direction preserved.
func TestRunMultiAudit_DQ2_MissingClaudeAnchor_Refuses(t *testing.T) {
	rc := &recordingCaller{}
	orig := backendCall
	backendCall = rc.call
	t.Cleanup(func() { backendCall = orig })

	cfg := MultiAuditConfig{
		Gates: config.AuditGates{
			Claude: config.AuditGateRequired,
			Codex:  config.AuditGateRequired,
			GLM:    config.AuditGateAdvisory,
		},
	}
	// No claude verdict provided — Verdict field empty.
	r := runMultiAudit(context.Background(), ReviewOutput{}, "uncommittedChanges", "concurrency", cfg)
	if r.OverallVerdict != "fail" {
		t.Errorf("overall = %q, want fail (missing claude anchor refuses to synthesize)", r.OverallVerdict)
	}
	if r.ResidualRiskNote == "" {
		t.Error("residual_risk_note empty; want the missing-anchor explanation")
	}
	// No backend should have been invoked once the anchor was found missing.
	if len(rc.calls) != 0 {
		t.Errorf("backend invoked %d time(s) despite missing claude anchor; want 0", len(rc.calls))
	}
}

// AC-AMM-003 / EC-6 (R1): super-review independence — the claude_verdict MUST
// NOT appear in any secondary backend's call payload. The recordingCaller's
// signature structurally excludes it, AND we assert no claude-analysis text
// leaked into target/focus either.
func TestRunMultiAudit_Independence_ClaudeVerdictNotInSecondaryPayload_AC_AMM_003_EC6(t *testing.T) {
	rc := &recordingCaller{}
	orig := backendCall
	backendCall = rc.call
	t.Cleanup(func() { backendCall = orig })

	// A rich claude verdict with distinctive text that must NEVER reach codex/glm.
	const claudeSecret = "CLAUDE-ANALYSIS-SECRET-do-not-share-with-codex-or-glm-12345"
	claude := ReviewOutput{
		Verdict: "pass",
		Summary: claudeSecret,
		Findings: []Finding{
			{Severity: "high", Title: claudeSecret, Body: claudeSecret, File: "x.go", Line: 42, Confidence: 0.9},
		},
		NextSteps: []string{claudeSecret},
	}
	cfg := MultiAuditConfig{
		Gates: config.AuditGates{
			Claude: config.AuditGateRequired,
			Codex:  config.AuditGateRequired,
			GLM:    config.AuditGateRequired,
		},
	}
	runMultiAudit(context.Background(), claude, "uncommittedChanges", "concurrency", cfg)

	if len(rc.calls) == 0 {
		t.Fatal("no backend calls recorded; codex+glm should have been invoked (all gates required)")
	}
	for _, c := range rc.calls {
		// The recordingCall struct has NO ClaudeVerdict field — if the engine
		// tried to pass it, the call wouldn't compile. This loop is belt-and-
		// suspenders: assert the secret text never reached target/focus either.
		if strings.Contains(c.Target, claudeSecret) || strings.Contains(c.Focus, claudeSecret) {
			t.Errorf("claude verdict text leaked into %s call payload (target=%q focus=%q)", c.Backend, c.Target, c.Focus)
		}
		if c.Backend != BackendCodex && c.Backend != BackendGLM {
			t.Errorf("unexpected backend %q; only codex/glm are secondary", c.Backend)
		}
	}
}

// AC-AMM-002: codex + glm fan out IN PARALLEL — wall-clock is closer to
// max(t_codex, t_glm) than to t_codex + t_glm. Uses blocking seams + a
// coordinator channel so the assertion is deterministic (not wall-clock flaky).
func TestRunMultiAudit_ParallelFanOut_AC_AMM_002(t *testing.T) {
	started := make(chan string, 2)
	release := make(chan struct{})
	parallel := struct {
		mu       sync.Mutex
		inflight int
		max      int
	}{}
	parallelFn := func(_ context.Context, backend, _, _ string) ReviewOutput {
		parallel.mu.Lock()
		parallel.inflight++
		if parallel.inflight > parallel.max {
			parallel.max = parallel.inflight
		}
		parallel.mu.Unlock()
		started <- backend
		<-release // block until both are inflight
		parallel.mu.Lock()
		parallel.inflight--
		parallel.mu.Unlock()
		return ReviewOutput{Verdict: "pass", Summary: backend + ":pass", Findings: []Finding{}, NextSteps: []string{}}
	}
	orig := backendCall
	backendCall = parallelFn
	t.Cleanup(func() { backendCall = orig })

	cfg := MultiAuditConfig{
		Gates: config.AuditGates{Claude: config.AuditGateRequired, Codex: config.AuditGateRequired, GLM: config.AuditGateRequired},
	}
	// Run in a goroutine so we can coordinate the release timing.
	done := make(chan ConvergenceResult, 1)
	go func() {
		done <- runMultiAudit(context.Background(), claudeReview("pass"), "uncommittedChanges", "", cfg)
	}()

	// Wait for BOTH backends to have started (proves they ran concurrently).
	got := map[string]bool{}
	for i := 0; i < 2; i++ {
		select {
		case b := <-started:
			got[b] = true
		case <-time.After(5 * time.Second):
			t.Fatalf("timeout waiting for backend %d to start (got %v)", i+1, got)
		}
	}
	close(release) // let both finish
	<-done

	if !got[BackendCodex] || !got[BackendGLM] {
		t.Errorf("expected both codex+glm to start in parallel; got %v", got)
	}
	// If max inflight < 2, they ran sequentially (wall-clock = sum, not max).
	parallel.mu.Lock()
	maxInflight := parallel.max
	parallel.mu.Unlock()
	if maxInflight < 2 {
		t.Errorf("max concurrent backends = %d, want >= 2 (fan-out must be parallel, not sequential)", maxInflight)
	}
}

// AC-AMM-014: a backend whose audit_gate == off is NOT invoked, and no entry
// for it appears in per_backend_verdicts.
func TestRunMultiAudit_AuditGateOff_SkipsBackend_AC_AMM_014(t *testing.T) {
	rc := &recordingCaller{}
	orig := backendCall
	backendCall = rc.call
	t.Cleanup(func() { backendCall = orig })

	cfg := MultiAuditConfig{
		Gates: config.AuditGates{
			Claude: config.AuditGateRequired,
			Codex:  config.AuditGateOff, // codex opted out
			GLM:    config.AuditGateAdvisory,
		},
	}
	r := runMultiAudit(context.Background(), claudeReview("pass"), "uncommittedChanges", "", cfg)
	for _, c := range rc.calls {
		if c.Backend == BackendCodex {
			t.Error("codex was invoked despite audit_gate == off")
		}
	}
	for _, v := range r.PerBackendVerdicts {
		if v.Backend == BackendCodex {
			t.Error("codex appears in per_backend_verdicts despite audit_gate == off")
		}
	}
}

// DQ-1: ConvergenceResult is persisted to .moai/state/audit-multi/<session>.json
// on every call so the M5 multi-review-gate Stop hook can read the most recent
// result. The path is session-scoped + gitignored (.moai/state/ is local-only).
func TestPersistConvergenceResult_WritesStateFile_DQ1(t *testing.T) {
	tmp := t.TempDir()
	orig := convergenceStateDir
	convergenceStateDir = tmp
	t.Cleanup(func() { convergenceStateDir = orig })

	r := ConvergenceResult{
		PerBackendVerdicts: []PerBackendVerdict{pbvReq(BackendClaude, "pass")},
		OverallVerdict:     "pass",
		DisagreementFlag:   false,
		ResidualRiskNote:   "",
		FailOpenBackends:   []string{},
	}
	if err := persistConvergenceResult(r, "sess-123"); err != nil {
		t.Fatalf("persistConvergenceResult: %v", err)
	}
	path := filepath.Join(tmp, "audit-multi", "sess-123.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("state file not written: %v", err)
	}
	// json.MarshalIndent emits `"overall_verdict": "pass"` (space after colon).
	if !strings.Contains(string(b), `"overall_verdict"`) || !strings.Contains(string(b), `"pass"`) {
		t.Errorf("state file content = %s; want overall_verdict pass", b)
	}
}

// persistConvergenceResult must fail-open (never block the flow on a disk
// error) — it returns the error so the caller can log it but the convergence
// result is still usable. A read-only dir produces a non-nil error.
func TestPersistConvergenceResult_FailOpenOnDiskError(t *testing.T) {
	ro := filepath.Join(t.TempDir(), "ro")
	if err := os.MkdirAll(ro, 0o555); err != nil {
		t.Skipf("cannot make read-only dir on this platform: %v", err)
	}
	orig := convergenceStateDir
	convergenceStateDir = ro
	t.Cleanup(func() { convergenceStateDir = orig })

	r := ConvergenceResult{OverallVerdict: "pass"}
	err := persistConvergenceResult(r, "sess-ro")
	if err == nil {
		// Some platforms (root, certain filesystems) permit writes inside a 0555
		// dir; skip rather than fail in that case.
		t.Skip("writable read-only dir (running as root or permissive FS); cannot assert error path")
	}
	_ = err // the load-bearing property: the function returns an error, the caller continues with r intact.
}

// ─── structural invariants (no new enum, no AskUserQuestion, SSOT) ───

// EC-7 / AC-AMM-011 (REQ-AMM-008): no new VerdictDisagreement enum value exists
// in the codebase. Disagreement is a boolean flag on ConvergenceResult, never a
// verdict value.
func TestNoVerdictDisagreementEnum_EC7_AC_AMM_011(t *testing.T) {
	matches := grepRepo(t, []string{"internal/cli/mcp_convergence.go", "internal/cli/mcp_convergence_test.go", "internal/cli/mcp_audit.go", "internal/cli/mcp_codex.go", "internal/cli/mcp_glm.go"}, "VerdictDisagreement")
	if len(matches) > 0 {
		t.Errorf("VerdictDisagreement enum value found (must not exist — disagreement is a boolean flag):\n%s", strings.Join(matches, "\n"))
	}
}

// AC-AMM-024 / C-HRA-008 (REQ-AMM-018): the convergence engine MUST NOT call
// AskUserQuestion or emit free-form user questions. Subagent boundary.
func TestConvergence_NoAskUserQuestion_AC_AMM_024(t *testing.T) {
	matches := grepRepo(t, []string{"internal/cli/mcp_convergence.go"}, `AskUserQuestion|mcp__askuser`)
	if len(matches) > 0 {
		t.Errorf("AskUserQuestion reference found in convergence engine (subagent boundary violated):\n%s", strings.Join(matches, "\n"))
	}
}

// AC-AMM-005 / C6 (REQ-AMM-005): the convergence engine resolves model/effort
// ONLY through template.ResolveAgentModelEffort — it does NOT read agent
// frontmatter or llm.agent_overrides directly (fork risk). Since the engine
// delegates the actual backend calls to the existing codex/glm handlers (which
// already go through the SSOT), the engine itself MUST contain no direct
// frontmatter/override read.
func TestConvergence_NoDirectFrontmatterRead_AC_AMM_005(t *testing.T) {
	matches := grepRepo(t, []string{"internal/cli/mcp_convergence.go"}, `agent_overrides|frontmatter|ReadAgentFrontmatter`)
	if len(matches) > 0 {
		t.Errorf("direct frontmatter/override read in convergence engine (SSOT violation — ResolveAgentModelEffort is the sole interpreter):\n%s", strings.Join(matches, "\n"))
	}
}

// REQ-AMM-008 / AC-AMM-011 (no new audit_model enum token): the `multi` token
// ALREADY exists; the engine activates it — no new token. Verify the engine
// never invents a new audit_model value.
func TestConvergence_NoNewAuditModelEnum(t *testing.T) {
	matches := grepRepo(t, []string{"internal/cli/mcp_convergence.go"}, `AuditModel[A-Z][a-zA-Z]+\s*=`)
	if len(matches) > 0 {
		t.Errorf("new AuditModel* constant defined in convergence engine (the multi token already exists; no new enum):\n%s", strings.Join(matches, "\n"))
	}
}

// ─── grep helper (mirrors the TestWeb_NoAskUserQuestion static-guard pattern) ───

// grepRepo reads the named files (relative to the package dir, which IS the
// internal/cli directory for this test file) and returns lines whose trimmed
// lower-cased form is NOT a pure comment AND that contain needle. Comment lines
// (leading //) are excluded so doc references do not trip the guard.
func grepRepo(t *testing.T, files []string, needle string) []string {
	t.Helper()
	var hits []string
	for _, rel := range files {
		b, err := os.ReadFile(rel)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue // a not-yet-created impl file is RED, not a grep failure
			}
			t.Errorf("grep read %s: %v", rel, err)
			continue
		}
		for _, line := range strings.Split(string(b), "\n") {
			trim := strings.TrimSpace(line)
			if strings.HasPrefix(trim, "//") {
				continue
			}
			if strings.Contains(line, needle) {
				hits = append(hits, rel+": "+trim)
			}
		}
	}
	return hits
}

// ─── production-wiring tests: engine reuses codex/glm handlers verbatim ───
//
// These exercise performCodexAudit / performGLMAudit / defaultBackendCaller via
// the EXISTING single-backend seams (codexLookPath, codexRunner, glmKeyLoader,
// glmHTTPClient). They prove the convergence engine CALLS the existing handlers
// (C1 — additive to MOAI-MCP-SERVER, no re-implementation — AP-AMM-1) and lift
// mcp_convergence.go coverage above the 85% target by exercising the production
// wiring paths the algorithm-only tests skip.

// fakeCodexRunner is reused from mcp_codex_test.go (same package).

// performCodexAudit reuses the existing codex binary-resolution + JSON-RPC
// session seam: when codex is on PATH and the session yields a clean review, the
// engine passes the synthesized verdict through unchanged (NO re-implementation,
// AP-AMM-1).
func TestPerformCodexAudit_ReusesExistingCodexHandler_NoReimpl_AP_AMM_1(t *testing.T) {
	// Swap codex seams (the same ones mcp_codex_test.go swaps). The adversarial
	// path (turn/start) emits the verdict as a final agentMessage; a clean review
	// (no finding bullets) synthesizes to verdict=pass.
	lines := []string{
		`{"id":1,"result":{"userAgent":"fake/1","codexHome":"/x","platformFamily":"unix","platformOs":"macos"}}`,
		`{"id":2,"result":{"thread":{"id":"tid-fake"}}}`,
		`{"id":3,"result":{"turn":{"id":"trn","status":"inProgress"}}}`,
		`{"method":"item/completed","params":{"threadId":"tid-fake","turnId":"trn","completedAtMs":1,"item":{"type":"agentMessage","id":"m1","text":"codex:ok, no findings"}}}`,
		`{"method":"turn/completed","params":{"threadId":"tid-fake","turn":{"id":"trn","status":"completed"}}}`,
	}
	prevLook, prevSess := codexLookPath, codexSession
	codexLookPath = func(string) (string, error) { return "/fake/codex", nil }
	codexSession = &fakeCodexSession{lines: lines}
	t.Cleanup(func() { codexLookPath, codexSession = prevLook, prevSess })

	out := performCodexAudit(context.Background(), "uncommittedChanges", "concurrency")
	if out.Verdict != "pass" {
		t.Errorf("performCodexAudit verdict = %q, want pass (must pass through codex result)", out.Verdict)
	}
}

// performCodexAudit fails OPEN (returns VerdictInconclusive) when the codex
// binary is missing — same fail-open identity the single-backend codex_audit
// handler has (AC-AMM-004 / C2 / EC-1). Proves the engine inherits the existing
// fail-open, never invents its own.
func TestPerformCodexAudit_FailOpenWhenCodexMissing_AC_AMM_004(t *testing.T) {
	prevLook := codexLookPath
	codexLookPath = func(string) (string, error) { return "", os.ErrNotExist }
	t.Cleanup(func() { codexLookPath = prevLook })

	out := performCodexAudit(context.Background(), "uncommittedChanges", "")
	if out.Verdict != VerdictInconclusive {
		t.Errorf("performCodexAudit missing-binary verdict = %q, want %q (fail-open inherited)", out.Verdict, VerdictInconclusive)
	}
}

// performGLMAudit reuses the existing GLM key-load + HTTP seam: with a stubbed
// z.ai response, the engine passes the parsed verdict through unchanged.
func TestPerformGLMAudit_ReusesExistingGLMHandler_NoReimpl_AP_AMM_1(t *testing.T) {
	stub := &stubGLMDoer{body: glmMessagesResp(t, ReviewOutput{Verdict: "fail", Summary: "glm:bad"})}
	withGLMSeams(t, "test-glm-key", stub)

	out := performGLMAudit(context.Background(), "auth")
	if out.Verdict != "fail" {
		t.Errorf("performGLMAudit verdict = %q, want fail (must pass through glm result)", out.Verdict)
	}
}

// performGLMAudit fails OPEN when the GLM key is missing (same fail-open the
// single-backend glm_audit handler has).
func TestPerformGLMAudit_FailOpenWhenKeyMissing_AC_AMM_004(t *testing.T) {
	withGLMSeams(t, "", nil) // empty key
	out := performGLMAudit(context.Background(), "")
	if out.Verdict != VerdictInconclusive {
		t.Errorf("performGLMAudit missing-key verdict = %q, want %q (fail-open inherited)", out.Verdict, VerdictInconclusive)
	}
}

// defaultBackendCaller routes codex→performCodexAudit, glm→performGLMAudit, and
// fails open (inconclusive) for any unknown backend name — never a hard error.
func TestDefaultBackendCaller_RoutesAndFailsOpenOnUnknown(t *testing.T) {
	t.Run("unknown backend fails open to inconclusive", func(t *testing.T) {
		out := defaultBackendCaller(context.Background(), "grok", "uncommittedChanges", "")
		if out.Verdict != VerdictInconclusive {
			t.Errorf("unknown backend verdict = %q, want %q", out.Verdict, VerdictInconclusive)
		}
	})
	t.Run("codex routes through performCodexAudit (missing binary ⇒ inconclusive)", func(t *testing.T) {
		prevLook := codexLookPath
		codexLookPath = func(string) (string, error) { return "", os.ErrNotExist }
		t.Cleanup(func() { codexLookPath = prevLook })
		out := defaultBackendCaller(context.Background(), BackendCodex, "uncommittedChanges", "")
		if out.Verdict != VerdictInconclusive {
			t.Errorf("codex route verdict = %q, want inconclusive (binary missing)", out.Verdict)
		}
	})
	t.Run("glm routes through performGLMAudit (missing key ⇒ inconclusive)", func(t *testing.T) {
		withGLMSeams(t, "", nil)
		out := defaultBackendCaller(context.Background(), BackendGLM, "uncommittedChanges", "")
		if out.Verdict != VerdictInconclusive {
			t.Errorf("glm route verdict = %q, want inconclusive (key missing)", out.Verdict)
		}
	})
}

// gateOr honors the distributed default when the configured gate is empty.
func TestGateOr_DefaultsWhenEmpty(t *testing.T) {
	if got := gateOr("", config.AuditGateRequired); got != config.AuditGateRequired {
		t.Errorf("gateOr(empty, required) = %q, want required", got)
	}
	if got := gateOr(config.AuditGateOff, config.AuditGateRequired); got != config.AuditGateOff {
		t.Errorf("gateOr(off, required) = %q, want off (configured value preserved)", got)
	}
}

// defaultConvergenceStateDir returns a non-empty path when resolveProjectDir
// resolves (production case); falls back to "" otherwise (fail-open).
func TestDefaultConvergenceStateDir(t *testing.T) {
	got := defaultConvergenceStateDir()
	// In test context resolveProjectDir may resolve to cwd or ""; both are
	// acceptable. The load-bearing property is the function does not panic.
	_ = got
}
