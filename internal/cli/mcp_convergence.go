// Package cli — SPEC-AUDIT-MULTI-MODEL-001 multi-model audit convergence engine.
//
// mcp_convergence.go implements the parallel cross-backend fan-out + the
// disagreement-synthesis algorithm layered ON TOP of the single-backend
// infrastructure shipped by SPEC-MOAI-MCP-SERVER-001. It activates the `multi`
// audit_model token: when `audit_model: multi`, the engine runs codex + glm in
// parallel (per their audit_gate), accepts the in-session claude verdict as the
// always-available anchor, and converges the per-backend verdicts into a single
// ConvergenceResult.
//
// Design decisions locked at M0 (progress.md §D + design.md §1, §3, §5, §7):
//   - ConvergenceResult carries per_backend_verdicts[], overall_verdict,
//     disagreement_flag, residual_risk_note, fail_open_backends. The overall
//     verdict is ONE of the existing {pass, fail} values — NO new
//     VerdictDisagreement enum (REQ-AMM-008 / AC-AMM-011).
//   - Disagreement is INFORMATION, not a GATE (C3): a disagreement_flag is
//     surfaced as Verification Matrix residual-risk + advisory; it never
//     hard-blocks the flow on its own. A required-split resolves conservatively
//     to overall = fail via the per-backend required-gate contract (NOT a new
//     disagreement-block category).
//   - Super-review independence (REQ-AMM-003 / C4): claude_verdict is consumed
//     ONLY by the converge(...) synthesis step. The codex/glm goroutines receive
//     (target, focus) — NEVER the claude analysis. Enforced structurally by the
//     backendCaller signature.
//   - Fail-open identity (C2): a missing/unauthenticated/optional backend
//     returns VerdictInconclusive for that slot and convergence continues over
//     the rest. Evidence of absence ≠ evidence of failure.
//   - DQ-1: ConvergenceResult is written to .moai/state/audit-multi/<session>
//     .json on every call so the M5 multi-review-gate Stop hook reads the most
//     recent result rather than re-invoking convergence.
//   - DQ-2: claude_verdict absent → the engine REFUSES (overall = fail + a
//     residual_risk_note explaining the missing anchor). The refusal is a
//     structured result, NEVER a hard error — fail-open direction preserved.
//
// The engine NEVER invokes AskUserQuestion (subagent boundary, REQ-AMM-018 /
// C5): a missing-input or inconclusive condition is returned as a structured
// ConvergenceResult and the orchestrator translates it through its own channel.
//
// @MX:SPEC: SPEC-AUDIT-MULTI-MODEL-001
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/modu-ai/moai-adk/internal/config"
	"golang.org/x/sync/errgroup"
)

// Backend name constants — the three audit backends. These live here (NOT in
// mcp_codex.go / mcp_glm.go) because the convergence engine is the one surface
// that names all three uniformly. The single-backend handlers continue to use
// their own domain identifiers; the convergence engine maps onto these.
const (
	BackendClaude = "claude"
	BackendCodex  = "codex"
	BackendGLM    = "glm"
)

// overall verdict values — the existing pass/fail set. NO new enum (REQ-AMM-008).
const (
	overallVerdictPass = "pass"
	overallVerdictFail = "fail"
)

// convergenceStateDir is the directory under which per-session ConvergenceResult
// state files are written (DQ-1). Defaults to <projectDir>/.moai/state but
// overridable for tests. The file path is .moai/state/audit-multi/<session>.json
// — .moai/state/ is local-only (CLAUDE.local.md §2), so the state file never
// ships in the template and never reaches a distributed user's git tree.
//
// A string (not a func) so tests can override by direct assignment; the project
// dir is stable within a process so computing once at package load is correct.
var convergenceStateDir = defaultConvergenceStateDir()

// ─── data model (design.md §1) ───

// PerBackendVerdict is one entry in ConvergenceResult.PerBackendVerdicts — the
// raw verdict of a single backend alongside its configured gate. Preserving
// per-backend verdicts (rather than collapsing upstream) gives the Verification
// Matrix the audit trail it needs: WHICH backend disagreed, not just that one
// did.
type PerBackendVerdict struct {
	Backend   string    `json:"backend"` // BackendClaude | BackendCodex | BackendGLM
	Gate      string    `json:"gate"`    // config.AuditGateOff | AuditGateAdvisory | AuditGateRequired
	Verdict   string    `json:"verdict"` // existing review-output.schema.json values (pass|fail|inconclusive); NO new enum
	Summary   string    `json:"summary"`
	Findings  []Finding `json:"findings"`
	NextSteps []string  `json:"next_steps"`
}

// ConvergenceResult is the synthesis output of converge(...) — the single
// structured result returned by the `audit_multi` MCP tool (M3) and read by the
// multi-review-gate Stop hook (M5). Shape locked at M0 (progress.md §D).
type ConvergenceResult struct {
	PerBackendVerdicts []PerBackendVerdict `json:"per_backend_verdicts"`
	// OverallVerdict ∈ {pass, fail} — the existing review-output.schema.json
	// values. NO VerdictDisagreement enum is ever produced (REQ-AMM-008).
	OverallVerdict   string   `json:"overall_verdict"`
	DisagreementFlag bool     `json:"disagreement_flag"`
	ResidualRiskNote string   `json:"residual_risk_note"`
	FailOpenBackends []string `json:"fail_open_backends"`
}

// ─── convergence algorithm (design.md §3) ───

// converge is the pure synthesis function: given the per-backend verdicts
// (already collected, with their gates attached), derive the ConvergenceResult
// per the 4-case policy table (REQ-AMM-006 #1–#4):
//
//  1. all required backends PASS (and no required FAIL)        → overall PASS
//  2. any required FAIL (no required PASS to split against)    → overall FAIL
//  3. required split (≥1 required PASS + ≥1 required FAIL)     → overall FAIL
//     (conservative — the per-backend required-gate contract holds), AND
//     disagreement_flag is set
//  4. advisory-only conflict (all required PASS, ≥1 advisory FAIL/PASS that
//     conflicts)                                               → overall PASS,
//     disagreement_flag is set (surfaced as advisory, NOT a block)
//
// Fail-open: when all required backends returned VerdictInconclusive (or a
// pass/inconclusive mix with no required FAIL), the overall verdict falls back
// to the claude verdict (AC-AMM-021 / EC-4). A missing optional backend is
// evidence-of-absence, NOT evidence-of-failure (C2).
//
// disagreement_flag is set when (a) the required set contains both pass and
// fail (a split), OR (b) an advisory backend's pass/fail verdict conflicts with
// the required set's verdicts (advisory-vs-required conflict, surfaced but not
// blocking). It is NEVER a new block category on its own (C3).
func converge(verdicts []PerBackendVerdict) ConvergenceResult {
	// Normalize empty slices so JSON serialization is uniform ([] not null).
	if verdicts == nil {
		verdicts = []PerBackendVerdict{}
	}

	required := filterRequired(verdicts)
	requiredFails := filterVerdict(required, "fail")

	// ── Step 1: overall_verdict derivation (REQ-AMM-006) ──
	var overall string
	switch {
	case len(requiredFails) > 0:
		// Case #2 / #3: any required FAIL → conservative fail. This is the
		// per-backend required-gate contract holding, NOT a disagreement-block.
		overall = overallVerdictFail
	case allPass(required):
		// Case #1 (all required pass) OR vacuous truth (no required backends —
		// nothing required, nothing blocks). Either way: overall pass.
		overall = overallVerdictPass
	default:
		// No required FAIL, but not all required PASS — i.e. some required are
		// inconclusive (missing/erroring optionals). Fail-OPEN to claude
		// (AC-AMM-021 / EC-1 / EC-4): the in-session claude verdict is the
		// always-available anchor.
		overall = claudeVerdictOrDefault(verdicts, overallVerdictPass)
	}

	// ── Step 2: disagreement_flag derivation ──
	distinctRequired := distinctVerdicts(required, "pass", "fail")
	disagreement := len(distinctRequired) > 1
	// Advisory-vs-required conflict: an advisory backend's pass/fail conflicts
	// with the required set. Surfaced (flag set) but NEVER blocks (C3 / case #4).
	if !disagreement && len(distinctRequired) > 0 {
		for _, v := range verdicts {
			if v.Gate != config.AuditGateAdvisory {
				continue
			}
			if v.Verdict != "pass" && v.Verdict != "fail" {
				continue // inconclusive advisories don't conflict
			}
			if !verdictIn(distinctRequired, v.Verdict) {
				disagreement = true
				break
			}
		}
	}

	// ── Step 3: residual_risk_note + fail_open_backends ──
	// The note is surfaced for the Verification Matrix residual-risk row whenever
	// there is something for a human reader to see: a disagreement (split or
	// advisory-vs-required conflict) OR a required-backend FAIL (AC-AMM-007:
	// "the residual_risk_note records which backend(s) failed"). Empty only when
	// the convergence is clean (all required PASS, no disagreement).
	note := ""
	switch {
	case disagreement:
		note = describeDisagreement(verdicts)
	case len(requiredFails) > 0:
		note = describeRequiredFails(requiredFails)
	}

	return ConvergenceResult{
		PerBackendVerdicts: verdicts,
		OverallVerdict:     overall,
		DisagreementFlag:   disagreement,
		ResidualRiskNote:   note,
		FailOpenBackends:   collectFailOpen(verdicts),
	}
}

// filterRequired returns only the required-gate verdicts.
func filterRequired(vs []PerBackendVerdict) []PerBackendVerdict {
	var out []PerBackendVerdict
	for _, v := range vs {
		if v.Gate == config.AuditGateRequired {
			out = append(out, v)
		}
	}
	return out
}

// filterVerdict returns only entries whose Verdict matches want.
func filterVerdict(vs []PerBackendVerdict, want string) []PerBackendVerdict {
	var out []PerBackendVerdict
	for _, v := range vs {
		if v.Verdict == want {
			out = append(out, v)
		}
	}
	return out
}

// allPass reports whether vs is empty OR every entry is a pass. (Vacuous truth
// for the empty case — no required backends means nothing required blocks.)
func allPass(vs []PerBackendVerdict) bool {
	for _, v := range vs {
		if v.Verdict != "pass" {
			return false
		}
	}
	return true
}

// distinctVerdicts returns the set of distinct verdict values among vs that
// appear in allow. Used to detect a required-set split (pass+fail both present).
func distinctVerdicts(vs []PerBackendVerdict, allow ...string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range vs {
		if !verdictIn(allow, v.Verdict) {
			continue
		}
		if !seen[v.Verdict] {
			seen[v.Verdict] = true
			out = append(out, v.Verdict)
		}
	}
	return out
}

// verdictIn reports whether want appears in xs. Named to avoid colliding with
// the test-scope `contains` (update_yaml_test.go), `sliceContains`
// (tool_policy_test.go), and `containsString` (update_preserve_inventory_test.go)
// helpers — those live in _test.go files and are not visible to non-test source,
// so the convergence engine needs its own.
func verdictIn(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}

// claudeVerdictOrDefault returns the claude backend's verdict from vs, or dflt
// if claude is absent. Used for the fail-open-to-claude branch.
func claudeVerdictOrDefault(vs []PerBackendVerdict, dflt string) string {
	for _, v := range vs {
		if v.Backend == BackendClaude {
			return v.Verdict
		}
	}
	return dflt
}

// collectFailOpen returns the backends whose verdict is VerdictInconclusive
// (missing/unauthenticated/errored). Made explicit in the result so the
// orchestrator can surface "codex was unavailable, fell back to claude + glm"
// without re-deriving it from per_backend_verdicts.
func collectFailOpen(vs []PerBackendVerdict) []string {
	var out []string
	for _, v := range vs {
		if v.Verdict == VerdictInconclusive {
			out = append(out, v.Backend)
		}
	}
	if out == nil {
		out = []string{}
	}
	return out
}

// describeRequiredFails names the required backends that FAILED, for the
// AC-AMM-007 "residual_risk_note records which backend(s) failed" requirement.
// Used in the no-split case (all required agree on fail) where disagreement_flag
// is false but the Verification Matrix still needs to name the failures.
func describeRequiredFails(fails []PerBackendVerdict) string {
	names := make([]string, 0, len(fails))
	for _, v := range fails {
		names = append(names, v.Backend)
	}
	return "required-backend FAIL: " + strings.Join(names, ", ")
}

// describeDisagreement produces a human-readable residual-risk note describing
// the split for the Verification Matrix. It names the disagreeing backends and
// their verdicts so a human reader of the Completion Report can see at a glance
// which backend disagreed with which.
func describeDisagreement(vs []PerBackendVerdict) string {
	var passList, failList []string
	for _, v := range vs {
		switch v.Verdict {
		case "pass":
			passList = append(passList, v.Backend+"("+v.Gate+")")
		case "fail":
			failList = append(failList, v.Backend+"("+v.Gate+")")
		}
	}
	if len(passList) == 0 || len(failList) == 0 {
		// No pass-vs-fail disagreement — e.g. an advisory inconclusive vs a
		// required pass. Surface a generic note rather than an empty one.
		return "cross-model disagreement detected; see per_backend_verdicts for details"
	}
	return fmt.Sprintf("cross-model disagreement (advisory, NOT a block): pass=[%s] fail=[%s]",
		strings.Join(passList, ", "), strings.Join(failList, ", "))
}

// ─── fan-out: errgroup parallel invocation of the active backends ───

// MultiAuditConfig carries the per-auditor gate map + optional session id used
// for DQ-1 state-file persistence. The fan-out reads the gates to decide which
// backends to invoke (gate == off ⇒ skip).
type MultiAuditConfig struct {
	Gates     config.AuditGates
	SessionID string // for .moai/state/audit-multi/<session>.json (DQ-1)
}

// backendCallFn is the injectable seam for the secondary-backend invocation.
// Production wires defaultBackendCaller (which reuses the codex/glm handler
// paths from mcp_codex.go / mcp_glm.go); tests swap it to record calls,
// simulate slow backends, or fail-open specific backends.
//
// INDEPENDENCE-INVARIANT (REQ-AMM-003 / C4 — load-bearing): the signature takes
// (ctx, backend, target, focus) and NOTHING ELSE. claude_verdict is structurally
// forbidden from reaching a secondary backend — it is consumed ONLY by converge.
// A future edit that tried to thread claude_verdict in would not compile against
// this signature.
type backendCallFn func(ctx context.Context, backend, target, focus string) ReviewOutput

// backendCall is the package-level seam (the type is backendCallFn). Tests swap
// it with t.Cleanup.
var backendCall backendCallFn = defaultBackendCaller

// defaultBackendCaller routes to the existing single-backend handlers. It reuses
// (does NOT re-implement) the codex binary shellout from mcp_codex.go and the
// GLM z.ai API call from mcp_glm.go (C1 — additive to MOAI-MCP-SERVER).
func defaultBackendCaller(ctx context.Context, backend, target, focus string) ReviewOutput {
	switch backend {
	case BackendCodex:
		return performCodexAudit(ctx, target, focus)
	case BackendGLM:
		return performGLMAudit(ctx, target, focus, "")
	default:
		// Unknown backend — fail-open to inconclusive (never a hard error).
		return inconclusiveReview("unknown backend: " + backend)
	}
}

// performCodexAudit wraps the existing codex handler path (binary resolution +
// runCodexReviewRPC) as a plain function callable without the MCP request
// ceremony. It uses adversarial mode (turn/start + the adversarial-review
// prompt) — the super-review secondary backend produces an uncorrelated second
// opinion, not a re-sample of the claude analysis.
//
// Reuses (does NOT fork) mcp_codex.go: codexLookPath, runCodexReviewRPC,
// codexAdversarialReviewPrompt, inconclusiveReview.
func performCodexAudit(ctx context.Context, target, focus string) ReviewOutput {
	binaryPath, err := codexLookPath(codexBinaryName)
	if err != nil {
		return inconclusiveReview("codex binary not found in PATH")
	}
	params := map[string]any{
		"target": target,
		"model":  "", // model/effort resolution is delegated to the SSOT at the single-backend layer
		"prompt": codexAdversarialReviewPrompt(focus),
	}
	if strings.TrimSpace(focus) != "" {
		params["focus"] = focus
	}
	out, _ := runCodexReviewRPC(ctx, binaryPath, codexMethodTurnStart, params) // fail-open inside
	return out
}

// performGLMAudit wraps the existing GLM handler path (key load + model resolve
// + callGLMAudit). Reuses (does NOT fork) mcp_glm.go: glmKeyLoader,
// resolveGLMAuditModel, callGLMAudit, glmInconclusive.
//
// It now also collects the change under review. codex is a subprocess inside the
// tree and reads it for itself; GLM is an HTTPS call to z.ai with no filesystem,
// so the only code it can review is the code this function puts in the request.
// Until it did, GLM was reviewing nothing and saying so with confidence — one
// live run returned a `fail` citing a repository whitelist this codebase does
// not contain (card t178).
//
// projectRoot names the tree to read the change from; empty falls back to the
// resolver the rest of this package uses. When no diff can be produced the
// result is inconclusive and z.ai is NOT called: the fail-open direction here
// points at "could not tell", never at a verdict about unseen code.
func performGLMAudit(ctx context.Context, target, focus, projectRoot string) ReviewOutput {
	key := glmKeyLoader()
	if key == "" {
		return glmInconclusive("GLM API key not configured (~/.moai/.env.glm)")
	}
	root := strings.TrimSpace(projectRoot)
	if root == "" {
		root = resolveProjectDir()
	}
	diff, err := collectReviewDiff(root, target)
	if err != nil {
		return glmInconclusive("no reviewable change: " + err.Error())
	}
	if strings.TrimSpace(diff) == "" {
		return glmInconclusive("no reviewable change: target " + target + " produced an empty diff")
	}
	return callGLMAudit(ctx, key, resolveGLMAuditModel(), focus, diff, nil)
}

// runMultiAudit is the fan-out entry point invoked by the `audit_multi` MCP tool
// (M3) and read by the multi-review-gate Stop hook (M5). It:
//  1. DQ-2: refuses if claude_verdict is absent (the always-available anchor).
//  2. Fans out across the active secondary backends (codex, glm) in parallel
//     via errgroup — each goroutine receives (target, focus) and NEVER the
//     claude verdict (super-review independence).
//  3. Assembles per_backend_verdicts (claude anchor + secondary results).
//  4. converge()s them into a ConvergenceResult.
//  5. DQ-1: persists the result to .moai/state/audit-multi/<session>.json.
//
// It NEVER returns a hard error — every path produces a structured
// ConvergenceResult (fail-open identity, C2). The orchestrator translates any
// inconclusive condition through its own AskUserQuestion channel (C5).
func runMultiAudit(ctx context.Context, claudeVerdict ReviewOutput, target, focus string, cfg MultiAuditConfig, token mcp.ProgressToken) ConvergenceResult {
	// ── DQ-2: claude_verdict anchor presence ──
	if strings.TrimSpace(claudeVerdict.Verdict) == "" {
		// REFUSE: claude_verdict is the always-available anchor per the fail-open
		// identity; proceeding without it would invert the fail-open direction
		// (synthesizing over secondary-only verdicts when the anchor that
		// guarantees a claude fallback is missing). The refusal is a STRUCTURED
		// result (overall = fail + a note), never a hard error.
		return ConvergenceResult{
			PerBackendVerdicts: []PerBackendVerdict{},
			OverallVerdict:     overallVerdictFail,
			DisagreementFlag:   false,
			ResidualRiskNote:   "claude_verdict anchor missing — refusing to synthesize (fail-open direction preserved; the in-session claude verdict is the always-available anchor)",
			FailOpenBackends:   []string{},
		}
	}

	// ── assemble claude's per-backend entry ──
	claudeGate := cfg.Gates.Claude
	if claudeGate == "" {
		claudeGate = config.AuditGateRequired // distributed default
	}
	verdicts := []PerBackendVerdict{
		{
			Backend:   BackendClaude,
			Gate:      claudeGate,
			Verdict:   claudeVerdict.Verdict,
			Summary:   claudeVerdict.Summary,
			Findings:  claudeVerdict.Findings,
			NextSteps: claudeVerdict.NextSteps,
		},
	}

	// ── fan out across the active secondary backends ──
	type secondaryResult struct {
		backend string
		gate    string
		out     ReviewOutput
	}
	var secondaries = []struct {
		name, gate string
	}{
		{BackendCodex, gateOr(cfg.Gates.Codex, config.AuditGateRequired)},
		{BackendGLM, gateOr(cfg.Gates.GLM, config.AuditGateAdvisory)},
	}

	var (
		mu    sync.Mutex
		parts []secondaryResult
	)
	eg, gctx := errgroup.WithContext(ctx)
	for _, s := range secondaries {
		s := s
		if s.gate == config.AuditGateOff {
			continue // gate off ⇒ backend NOT invoked (AC-AMM-014)
		}
		eg.Go(func() error {
			// The secondary backend receives (target, focus) ONLY — claude_verdict
			// is structurally excluded by the backendCaller signature. Any error
			// inside defaultBackendCaller is already fail-opened to a
			// VerdictInconclusive ReviewOutput, so this never returns a hard error.
			notifyMCPProgress(gctx, token, 0.3, s.name+" 백엔드 리뷰 실행 중...")
			out := backendCall(gctx, s.name, target, focus)
			notifyMCPProgress(gctx, token, 0.7, s.name+" 백엔드 응답 수신")
			mu.Lock()
			parts = append(parts, secondaryResult{backend: s.name, gate: s.gate, out: out})
			mu.Unlock()
			return nil
		})
	}
	_ = eg.Wait() // errors are already fail-opened inside the callers; ignore the errgroup error

	// Append secondary results in canonical order (codex before glm) so the
	// per_backend_verdicts array reads deterministically.
	for _, wantName := range []string{BackendCodex, BackendGLM} {
		for _, p := range parts {
			if p.backend == wantName {
				verdicts = append(verdicts, PerBackendVerdict{
					Backend:   p.backend,
					Gate:      p.gate,
					Verdict:   p.out.Verdict,
					Summary:   p.out.Summary,
					Findings:  p.out.Findings,
					NextSteps: p.out.NextSteps,
				})
			}
		}
	}

	result := converge(verdicts)

	// ── DQ-1: persist to .moai/state/audit-multi/<session>.json ──
	// Best-effort: a write failure is logged via the returned error but MUST NOT
	// block the flow (fail-open). The convergence result is valid regardless of
	// whether the state file landed.
	if cfg.SessionID != "" {
		_ = persistConvergenceResult(result, cfg.SessionID)
	}
	return result
}

// gateOr returns g unless empty, in which case it returns dflt. Used to honor
// the distributed default gate when the configured gate string is unset.
func gateOr(g, dflt string) string {
	if strings.TrimSpace(g) == "" {
		return dflt
	}
	return g
}

// ─── DQ-1: state-file persistence ───

// persistConvergenceResult writes the ConvergenceResult to
// <convergenceStateDir>/audit-multi/<sessionID>.json. Creates the directory if
// needed. Returns any disk error so the caller can log it; the caller MUST
// continue with the in-memory result regardless (fail-open).
func persistConvergenceResult(r ConvergenceResult, sessionID string) error {
	dir := filepath.Join(convergenceStateDir, "audit-multi")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("convergence state dir: %w", err)
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("convergence state marshal: %w", err)
	}
	path := filepath.Join(dir, sessionID+".json")
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("convergence state write: %w", err)
	}
	return nil
}

// defaultConvergenceStateDir resolves the project root via the existing
// resolveProjectDir seam (the same one the codex/glm handlers use to locate
// .moai/config/) and returns its .moai/state subdirectory. Falls back to an
// empty string (which makes persistConvergenceResult a no-op via MkdirAll
// failure) if the project dir cannot be resolved — fail-open identity preserved.
func defaultConvergenceStateDir() string {
	projectDir := resolveProjectDir()
	if projectDir == "" {
		return ""
	}
	return filepath.Join(projectDir, ".moai", "state")
}

var _ = defaultConvergenceStateDir // retained for readability of the initializer above
