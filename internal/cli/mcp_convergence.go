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

	// SynthesisNote carries forward ReviewOutput.SynthesisNote — the record that
	// this backend's own verdict signals disagreed. It is forwarded rather than
	// re-derived because converge sees only verdicts, never review bodies.
	SynthesisNote string `json:"synthesis_note,omitempty"`
}

// ConvergenceResult is the synthesis output of converge(...) — the single
// structured result returned by the `audit_multi` MCP tool (M3) and read by the
// multi-review-gate Stop hook (M5). Shape locked at M0 (progress.md §D).
type ConvergenceResult struct {
	PerBackendVerdicts []PerBackendVerdict `json:"per_backend_verdicts"`
	// OverallVerdict ∈ {pass, fail} — the existing review-output.schema.json
	// values. NO VerdictDisagreement enum is ever produced (REQ-AMM-008).
	OverallVerdict string `json:"overall_verdict"`
	// DisagreementFlag is a *bool (SPEC-AUDIT-PARTICIPANT-COUNT-001): the
	// nullable third state is the point. nil = undetermined — fewer than 2
	// participants were compared, so neither "they agreed" nor "they
	// disagreed" is a grounded claim, and a bare `false` would assert a
	// comparison that never happened. Non-nil = the three-pass derivation's
	// boolean at ≥2 participants, or `true` below 2 when the intra-backend
	// synthesis pass directly observed a divergence (REQ-APC-003 carve-out —
	// observed information is never discarded, C3). NO omitempty: a nil
	// pointer must serialize as an explicit JSON null, never vanish — an
	// absent member is indistinguishable from an older binary's output
	// (REQ-APC-003 / AC-APC-003).
	DisagreementFlag *bool `json:"disagreement_flag"`
	// ParticipantCount is how many on-target backends contributed a
	// comparable verdict (REQ-APC-001): gate != off AND verdict pass|fail
	// (REQ-APC-002). Inconclusive entries — missing, unauthenticated,
	// errored — are evidence-of-absence, not participants (C2). Always
	// present, 0 included, so a one-backend result is distinguishable from a
	// three-backend agreement in the derived summary alone. This SPEC
	// reports the count; it never acts on it (no minimum-participant
	// policy, no gate — spec.md §E).
	ParticipantCount int      `json:"participant_count"`
	ResidualRiskNote string   `json:"residual_risk_note"`
	FailOpenBackends []string `json:"fail_open_backends"`

	// BuildCommit / BuildLag record the identity of the ONE binary that
	// serviced all three backends (SPEC-AUDIT-BUILD-IDENTITY-001) —
	// deliberately TOP-LEVEL, not on PerBackendVerdict: repeating the same
	// value per backend would be triple bookkeeping of a single fact.
	// Commit, not version: one version string names both a lagging build and
	// a current one (REQ-ABI-003). Flat siblings, additive + omitempty (the
	// SynthesisNote precedent): an absent identity changes no existing
	// consumer's JSON. Because persistConvergenceResult marshals THIS struct
	// verbatim, filling the returned result is all REQ-ABI-002 needs — the
	// state file follows automatically. Assembled ONLY by auditBuildIdentity
	// (REQ-ABI-007).
	BuildCommit string `json:"build_commit,omitempty"`
	BuildLag    string `json:"build_lag,omitempty"`
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

	// ── Step 2b: intra-backend signal divergence (REQ-CVS-003) ──
	// A backend whose OWN verdict signals contradicted each other is a
	// disagreement too — one inside a single review rather than between two
	// backends — and the reader of the Verification Matrix needs to see it.
	// Setting the flag here changes nothing about overall_verdict, which was
	// already derived in Step 1: disagreement is information, never a block (C3).
	synthesisNotes := collectSynthesisNotes(verdicts)
	if len(synthesisNotes) > 0 {
		disagreement = true
	}

	// ── Step 2c: participant-count narrowing (REQ-APC-003 / REQ-APC-004) ──
	// `false` is a positive claim — two or more participants were compared
	// and none disagreed — and a single participant cannot ground it. Below
	// 2 participants the flag becomes nil (undetermined), EXCEPT when the
	// Step 2b pass directly observed an intra-backend divergence, whose
	// `true` is kept rather than discarded (the REQ-APC-003 carve-out).
	// At 2+ participants the boolean is exactly what the three-pass
	// derivation above produced, unchanged (REQ-APC-004).
	participantCount := countParticipants(verdicts)
	var disagreementFlag *bool
	switch {
	case participantCount >= 2:
		disagreementFlag = &disagreement
	case len(synthesisNotes) > 0:
		diverged := true // carve-out: observed information, never discarded (C3)
		disagreementFlag = &diverged
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
	// The intra-backend notes are appended rather than substituted: a
	// cross-backend split and a backend contradicting itself can hold at once,
	// and dropping either of them loses the reason the other needs reading.
	if len(synthesisNotes) > 0 {
		if note != "" {
			note += " | "
		}
		note += strings.Join(synthesisNotes, " | ")
	}

	return ConvergenceResult{
		PerBackendVerdicts: verdicts,
		OverallVerdict:     overall,
		DisagreementFlag:   disagreementFlag,
		ParticipantCount:   participantCount,
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

// countParticipants returns how many entries actually contributed a
// comparable verdict (SPEC-AUDIT-PARTICIPANT-COUNT-001 REQ-APC-002): a
// participant is an entry whose gate is not `off` AND whose verdict is pass
// or fail. An `inconclusive` entry — missing, unauthenticated, or errored —
// is evidence-of-absence, not a participant, whatever its gate (C2); a
// gate-`off` entry never counts, whatever verdict it carries.
func countParticipants(vs []PerBackendVerdict) int {
	n := 0
	for _, v := range vs {
		if v.Gate == config.AuditGateOff {
			continue
		}
		if v.Verdict == "pass" || v.Verdict == "fail" {
			n++
		}
	}
	return n
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
// collectSynthesisNotes returns each backend's intra-review divergence note,
// prefixed with the backend it came from so the combined residual-risk note
// stays readable when more than one backend contradicted itself.
func collectSynthesisNotes(vs []PerBackendVerdict) []string {
	var out []string
	for _, v := range vs {
		if v.SynthesisNote != "" {
			out = append(out, v.Backend+": "+v.SynthesisNote)
		}
	}
	return out
}

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

	// ProjectRoot names the tree the secondary backends should read
	// (SPEC-MCP-WORKTREE-ROOT-001). Empty ⇒ each backend keeps whatever it
	// resolved before this parameter existed, so an unaware caller sees no
	// change. Validated by the handler, never here.
	ProjectRoot string
}

// backendCallFn is the injectable seam for the secondary-backend invocation.
// Production wires defaultBackendCaller (which reuses the codex/glm handler
// paths from mcp_codex.go / mcp_glm.go); tests swap it to record calls,
// simulate slow backends, or fail-open specific backends.
//
// INDEPENDENCE-INVARIANT (REQ-AMM-003 / C4 — load-bearing): the signature carries
// NO verdict of any kind. claude_verdict is structurally forbidden from reaching a
// secondary backend — it is consumed ONLY by converge, and a future edit that tried
// to thread it in would not compile against this signature.
//
// The parameter list is (ctx, backend, target, focus, projectRoot). projectRoot was
// added by SPEC-MCP-WORKTREE-ROOT-001 so a caller inside a worktree can say which
// tree the backends should read; it names a directory, carries no analysis, and so
// leaves the independence guarantee intact. The invariant is stated as "no verdict
// crosses this seam" rather than "nothing else crosses it", because the latter
// wording stopped being true when this parameter landed — and a comment left
// asserting a guarantee the signature no longer expresses is how a real invariant
// stops being enforced without anyone noticing.
type backendCallFn func(ctx context.Context, backend, target, focus, projectRoot string) ReviewOutput

// backendCall is the package-level seam (the type is backendCallFn). Tests swap
// it with t.Cleanup.
var backendCall backendCallFn = defaultBackendCaller

// defaultBackendCaller routes to the existing single-backend handlers. It reuses
// (does NOT re-implement) the codex binary shellout from mcp_codex.go and the
// GLM z.ai API call from mcp_glm.go (C1 — additive to MOAI-MCP-SERVER).
func defaultBackendCaller(ctx context.Context, backend, target, focus, projectRoot string) ReviewOutput {
	switch backend {
	case BackendCodex:
		return performCodexAudit(ctx, target, focus, projectRoot)
	case BackendGLM:
		// Both sides of this seam changed at once, and the merge is what makes
		// them work: the GLM path gained a projectRoot so it can collect a real
		// diff to review, and this seam gained one so the caller can say which
		// tree that is. Passing "" here would leave GLM resolving its own root
		// while codex used the caller's — the two backends reviewing different
		// trees, which is the failure the parameter exists to prevent.
		return performGLMAudit(ctx, target, focus, projectRoot)
	default:
		// Unknown backend — fail-open to inconclusive (never a hard error).
		return inconclusiveReview("unknown backend: " + backend)
	}
}

// performCodexAudit wraps the existing codex handler path (binary resolution +
// the codexReviewRPC seam) as a plain function callable without the MCP request
// ceremony. It uses adversarial mode (turn/start + the adversarial-review
// prompt) — the super-review secondary backend produces an uncorrelated second
// opinion, not a re-sample of the claude analysis.
//
// Reuses (does NOT fork) mcp_codex.go: codexLookPath, the codexReviewRPC seam
// (wired to runCodexAuditReviewRPC — the pin-aware audit resolution per
// SPEC-V3R6-AUDIT-MODEL-PIN-001 M2; the audit_multi leg IS an audit entry
// point, so the workflow.audit.codex pin applies here),
// codexAdversarialReviewPrompt, inconclusiveReview.
// projectRoot (SPEC-MCP-WORKTREE-ROOT-001 REQ-1 / AC-1b) names the tree codex
// should read. This path carried NO `cwd` at all before — unlike the
// single-backend handler at mcp_codex.go, which has always passed one — so the
// repair ADDS the key rather than redirecting it, and the two codex paths were
// divergent until now. The key is written only when a root was actually named:
// substituting a default for an absent parameter would change what an existing
// caller's backend receives, which REQ-2 forbids.
func performCodexAudit(ctx context.Context, target, focus, projectRoot string) ReviewOutput {
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
	if root := strings.TrimSpace(projectRoot); root != "" {
		params["cwd"] = root
	}
	out, _ := codexReviewRPC(ctx, binaryPath, codexMethodTurnStart, params) // fail-open inside
	return out
}

// performGLMAudit wraps the existing GLM handler path (key load + model resolve
// + callGLMAudit). Reuses (does NOT fork) mcp_glm.go: glmKeyLoader,
// resolveGLMAuditModelEffort, callGLMAudit, glmInconclusive.
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
	me := resolveGLMAuditModelEffort(root) // pin > SSOT, from the SAME tree as the diff (CR #8)
	return callGLMAudit(ctx, key, me.Model, me.Effort, focus, diff, nil)
}

// runMultiAudit is the fan-out entry point invoked by the `audit_multi` MCP tool
// (M3) and read by the multi-review-gate Stop hook (M5). It:
//  1. DQ-2: refuses if claude_verdict is absent (the always-available anchor).
//  2. Fans out across the active secondary backends (codex, glm) in parallel
//     via errgroup — each goroutine receives (target, focus, cfg.ProjectRoot)
//     and NEVER the claude verdict (super-review independence).
//  3. Assembles per_backend_verdicts (claude anchor + secondary results).
//  4. converge()s them into a ConvergenceResult.
//  5. DQ-1: persists the result to .moai/state/audit-multi/<session>.json.
//
// It NEVER returns a hard error — every path produces a structured
// ConvergenceResult (fail-open identity, C2). The orchestrator translates any
// inconclusive condition through its own AskUserQuestion channel (C5).
func runMultiAudit(ctx context.Context, claudeVerdict ReviewOutput, target, focus string, cfg MultiAuditConfig, token mcp.ProgressToken) ConvergenceResult {
	// Build identity is assembled ONCE for the whole convergence (REQ-ABI-007)
	// and rides every result below — including the DQ-2 refusal, which is a
	// structured verdict like any other and must name its binary. The
	// comparison uses cfg.ProjectRoot when the caller named a tree; an absent
	// one falls back to the process cwd INSIDE auditBuildIdentity, for the
	// comparison only — the backends still receive exactly what
	// resolveOptionalToolProjectRoot resolved (nothing, on an omitted
	// project_root).
	buildCommit, buildLag := auditBuildIdentity(ctx, cfg.ProjectRoot)

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
			// Zero participants: nobody was compared, so the flag is the
			// zero-value nil (undetermined) rather than a `false` no
			// comparison grounds, and participant_count rides as 0
			// (SPEC-AUDIT-PARTICIPANT-COUNT-001, acceptance §B).
			ResidualRiskNote: "claude_verdict anchor missing — refusing to synthesize (fail-open direction preserved; the in-session claude verdict is the always-available anchor)",
			FailOpenBackends: []string{},
			BuildCommit:      buildCommit,
			BuildLag:         buildLag,
		}
	}

	// ── assemble claude's per-backend entry ──
	claudeGate := cfg.Gates.Claude
	if claudeGate == "" {
		claudeGate = config.AuditGateRequired // distributed default
	}
	verdicts := []PerBackendVerdict{
		{
			Backend:       BackendClaude,
			Gate:          claudeGate,
			Verdict:       claudeVerdict.Verdict,
			Summary:       claudeVerdict.Summary,
			Findings:      claudeVerdict.Findings,
			NextSteps:     claudeVerdict.NextSteps,
			SynthesisNote: claudeVerdict.SynthesisNote,
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
			// The secondary backend receives (target, focus, projectRoot) — no
			// verdict of any kind. claude_verdict stays structurally excluded by
			// the backendCaller signature; projectRoot names a directory and
			// carries no analysis, so the independence guarantee is unchanged.
			// Any error inside defaultBackendCaller is already fail-opened to a
			// VerdictInconclusive ReviewOutput, so this never returns a hard error.
			notifyMCPProgress(gctx, token, 0.3, s.name+" 백엔드 리뷰 실행 중...")
			out := backendCall(gctx, s.name, target, focus, cfg.ProjectRoot)
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
					Backend:       p.backend,
					Gate:          p.gate,
					Verdict:       p.out.Verdict,
					Summary:       p.out.Summary,
					Findings:      p.out.Findings,
					NextSteps:     p.out.NextSteps,
					SynthesisNote: p.out.SynthesisNote,
				})
			}
		}
	}

	result := converge(verdicts)
	// Set BEFORE persist: persistConvergenceResult marshals this struct
	// verbatim, so filling the returned result is what makes the state file
	// carry the same commit the verdict carried (REQ-ABI-002 — no separate
	// persistence-side code).
	result.BuildCommit, result.BuildLag = buildCommit, buildLag

	// ── DQ-1: persist to .moai/state/audit-multi/<session>.json ──
	// Best-effort: a write failure is logged via the returned error but MUST NOT
	// block the flow (fail-open). The convergence result is valid regardless of
	// whether the state file landed.
	if cfg.SessionID != "" {
		_ = persistConvergenceResult(result, cfg.SessionID, cfg.ProjectRoot)
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
// <state dir>/audit-multi/<sessionID>.json. Creates the directory if needed.
// Returns any disk error so the caller can log it; the caller MUST continue
// with the in-memory result regardless (fail-open).
//
// projectRoot names the tree the audit was run against. It is honored because
// the reader side already is: loadConvergenceResult takes the caller's own
// projectDir (multi_review_gate.go), so a run inside a worktree that wrote to
// the primary checkout's .moai/state left its verdict where that worktree's
// gate never looks — and mixed several worktrees' verdicts into one directory.
// Empty ⇒ the package-level convergenceStateDir, so an unaware caller and the
// existing tests that override that variable see no change.
func persistConvergenceResult(r ConvergenceResult, sessionID, projectRoot string) error {
	dir := filepath.Join(convergenceStateDirFor(projectRoot), "audit-multi")
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

// convergenceStateDirFor returns the .moai/state directory the convergence
// result belongs under: the named project root when the caller supplied one,
// otherwise the process-wide convergenceStateDir resolved at package load.
func convergenceStateDirFor(projectRoot string) string {
	if root := strings.TrimSpace(projectRoot); root != "" {
		return filepath.Join(root, ".moai", "state")
	}
	return convergenceStateDir
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
