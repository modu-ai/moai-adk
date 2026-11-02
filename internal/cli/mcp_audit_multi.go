// Package cli — `audit_multi` MCP tool handler
// (SPEC-AUDIT-MULTI-MODEL-001 M3, REQ-AMM-009 / REQ-AMM-010 / AC-AMM-012 /
// AC-AMM-013 / AC-AMM-014).
//
// mcp_audit_multi.go is a THIN WRAPPER over runMultiAudit (mcp_convergence.go).
// It maps the JSON-RPC tool params → (optional claude_verdict ReviewOutput, target,
// focus, MultiAuditConfig), calls runMultiAudit, and shapes the
// ConvergenceResult into the tool's declared output. The handler does NOT
// re-implement the Claude/codex/GLM backends (C1 — additive, no fork), does NOT call
// AskUserQuestion (subagent boundary, REQ-AMM-018 / C5), and NEVER returns a
// hard Go error — every fail-open path produces a structured result so the
// orchestrator translates through its own channel.
//
// @MX:SPEC: SPEC-AUDIT-MULTI-MODEL-001
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/config"
)

// auditMultiToolName is the canonical MCP tool name. The cross-model audit
// Skill (moai-ref-cross-model-audit, M4) references it verbatim so a future
// rename propagates via grep — keeping it as a const makes that grep reliable.
const auditMultiToolName = "audit_multi"

// handleAuditMulti is the thin-wrapper MCP tool handler for `audit_multi`
// (REQ-AMM-010 / AC-AMM-013). It:
//  1. Assembles the optional claude_verdict ReviewOutput. It is an anchor only
//     in a Claude-origin session; GPT/GLM/unknown origins ignore it and invoke
//     the independent subscription-backed Claude backend.
//  2. Reads the per-auditor audit_gate the caller supplied in the `gates`
//     argument and resolves the audited tree's plan around it: a supplied gate
//     wins, then the tree's audit.gates, then its audit.model token, then the
//     distributed default (config.ResolveAuditPlan).
//  3. Fans out by calling runMultiAudit — which reuses the existing
//     Claude/codex/GLM handler paths (NO backend re-implementation, AC-AMM-013).
//  4. Shapes the ConvergenceResult into the tool's declared output.
//
// The handler NEVER invokes AskUserQuestion (subagent boundary, REQ-AMM-018):
// a missing-input or inconclusive condition is returned as a structured
// ConvergenceResult and the orchestrator translates it through its own channel.
func handleAuditMulti(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	claudeVerdict, ok := readClaudeVerdict(req)
	if !ok {
		// Missing/malformed claude_verdict — fall through with an empty
		// ReviewOutput. All origins then use the independent Claude backend;
		// GPT/GLM/unknown origins would do so even if a verdict were supplied.
		// NEVER a hard error.
		claudeVerdict = ReviewOutput{}
	}

	target := req.GetString("target", "")
	focus := req.GetString("focus", "")
	gates := readGatesArgument(req)
	// A4 card argument (REQ-CLOSURE-012): binds the fan-out to the card.
	// Empty ⇒ byte-identical pre-change behavior.
	cardID := req.GetString("card_id", "")

	// The convergence engine writes its per-session state file under
	// .moai/state/audit-multi/<session>.json (DQ-1). The tool surface does NOT
	// invent a session id — it threads whatever the caller supplied (the
	// orchestrator resolves the authoritative id). An empty SessionID makes
	// persistence a no-op, which is fail-open safe.
	// SPEC-MCP-WORKTREE-ROOT-001 REQ-1/REQ-3: the caller may name the tree the
	// backends should read. Absent ⇒ empty ⇒ each backend resolves exactly as it
	// did before the parameter existed. An unusable path is REJECTED rather than
	// replaced by the default: a fallback would fan out over the primary checkout
	// while reporting success, which is the defect this parameter exists to fix.
	// This is the one hard-error path in a handler that is otherwise fail-open —
	// fail-open covers an absent or broken BACKEND, not a caller input the caller
	// can correct, and swallowing a mistyped path would make the mistake invisible.
	projectRoot, rootErr := resolveOptionalToolProjectRoot(req)
	if rootErr != nil {
		return toolErr(auditMultiToolName, rootErr), nil
	}

	// Card t1708: the optional delta scope, validated once before any backend
	// runs. Like the unusable project_root above it is a caller input the caller
	// can correct, so it is a hard error rather than a fallback to the full diff.
	rawDeltaBase, deltaErr := deltaBaseArg(req)
	if deltaErr != nil {
		return toolErr(auditMultiToolName, deltaErr), nil
	}
	// SPEC-AUDIT-ENGINE-CONVERGENCE-001 REQ-AEC-001/002: delta_base "full" asks for
	// the full resolved base; a card call without a delta_base (baseBranch target)
	// proposes the audited_sha of the card's newest FAIL verdict, which the chain
	// gate in runMultiAudit then admits or refuses. The note says what happened.
	var deltaDefaultNote string
	if strings.EqualFold(strings.TrimSpace(rawDeltaBase), deltaBaseFull) {
		rawDeltaBase = ""
		if cardID != "" {
			deltaDefaultNote = deltaFullNote()
		}
	} else if cardID != "" && strings.TrimSpace(rawDeltaBase) == "" && target == codexTargetBaseBranch {
		proposeRoot := projectRoot
		if proposeRoot == "" {
			proposeRoot = resolveProjectDir()
		}
		rawDeltaBase, deltaDefaultNote = defaultDeltaProposal(proposeRoot, cardID)
	}
	deltaRoot := projectRoot
	if deltaRoot == "" && rawDeltaBase != "" {
		deltaRoot = resolveProjectDir()
	}
	deltaBase, deltaErr := readDeltaBase(deltaRoot, rawDeltaBase, target)
	if deltaErr != nil {
		return toolErr(auditMultiToolName, deltaErr), nil
	}
	// SPEC-AUDIT-ENGINE-CONVERGENCE-001 REQ-AEC-004 (ii): the optional BLOCKING-
	// closure confirmation of a card delta call. Malformed input is a caller-
	// correctable error, not an absent confirmation: reading it as absent would
	// quietly widen the call to the full scope.
	blockingClosed, closedErr := readBlockingClosed(req)
	if closedErr != nil {
		return toolErr(auditMultiToolName, closedErr), nil
	}

	// SPEC-AUDIT-MODEL-CONVERGE-001 REQ-ACV-003/006: the audited tree's plan —
	// resolved from its RAW workflow.audit section and the gates this call
	// supplied — decides every backend's gate. A configured value outside the
	// closed sets is the second hard-error path of this handler for the same
	// reason as the unusable project_root above: a mistyped token silently
	// becoming the default is the failure the resolver exists to remove. Since
	// SPEC-AUDIT-CEILING-002 REQ-ACR-006, a workflow.yaml that cannot be read
	// or parsed is the third hard-error path: the error surfaces, it never
	// reads as an absent audit configuration.
	planRoot := projectRoot
	if planRoot == "" {
		planRoot = resolveProjectDir()
	}
	var (
		audit               config.AuditConfig
		primaryUnidentified bool
	)
	if planRoot != "" {
		var sectionErr error
		audit, primaryUnidentified, sectionErr = auditSectionForRoot(planRoot)
		if sectionErr != nil {
			return toolErr(auditMultiToolName, sectionErr), nil
		}
	}
	plan, planErr := config.ResolveAuditPlan(audit, gates)
	if planErr != nil {
		return toolErr(auditMultiToolName, planErr), nil
	}

	// The same call-start reading fixes the enforcement (REQ-ACV-006/008): the
	// configuration is read once, here, and the fan-out and the unmet-gate check
	// both follow it — not a second read after the backends have run.
	enforcement, enforcementNote := callStartEnforcement(audit, primaryUnidentified, gates)
	cfg := MultiAuditConfig{
		Gates:            planGates(plan),
		SessionID:        req.GetString("session_id", ""),
		ProjectRoot:      projectRoot,
		OriginProvider:   os.Getenv(config.EnvMoaiLaunchProvider),
		CardID:           cardID,
		DeltaBase:        deltaBase,
		BlockingClosed:   blockingClosed,
		DeltaDefaultNote: deltaDefaultNote,
		EnforcementGates: &enforcement,
		EnforcementNote:  enforcementNote,
	}
	if plan.FromConfig() {
		cfg.PlanSource = planSourceConfig
	}

	token := extractProgressToken(req)
	notifyMCPProgress(ctx, token, 0, "audit_multi start - converging the claude/codex/glm backends...")
	result := runMultiAudit(ctx, claudeVerdict, target, focus, cfg, token)
	notifyMCPProgress(ctx, token, 1, "audit_multi done - convergence result assembled")
	return convergenceToolResult(result)
}

// blockingClosedArg is the audit_multi argument carrying the BLOCKING-closure
// confirmation.
const blockingClosedArg = "blocking_closed"

// blockingClosedDescription is the tool-schema text of the blocking_closed input.
const blockingClosedDescription = "Optional BLOCKING-closure confirmation for a card_id delta review: an object mapping each `blocking_defect:` id of the card's FAIL verdict files to the commit that closed it (a commit SHA inside delta_base..HEAD). An id whose closure an earlier chain round already confirmed may map to an empty string. A delta review needs the confirmation to name every id; omitted or incomplete ⇒ the full scope is reviewed and the result says why. Ignored without card_id and delta_base."

// auditMultiDeltaBaseDescription is delta_base's schema text for audit_multi:
// the shared text says audit_multi files no record for a delta call, which is
// no longer true for a delta the chain gate admits.
var auditMultiDeltaBaseDescription = strings.Replace(strings.Replace(deltaBaseDescription,
	"audit_multi writes no second-review record for a delta call.",
	"With a card_id the SHA is a proposal: the delta is reviewed only when the card's second-review records form an unbroken, reviewed chain ending at it and blocking_closed confirms every BLOCKING defect; otherwise the full scope is reviewed and the result says why, and an admitted delta files one second-review record. Without a card_id no record is written. With a card_id and no delta_base, the audited_sha of the card's newest FAIL verdict file is proposed the same way, and the result names the file and SHA or says why none was proposed; \"full\" reviews the full resolved base instead.", 1),
	"codex adversarial (and the audit_multi codex leg) is told in the prompt only, and the result says so.",
	"codex adversarial is told in the prompt only, and the result says so; the audit_multi codex leg carries the delta diff the server collected in its adversarial turn.", 1)

// readBlockingClosed reads the optional blocking_closed argument: a JSON object
// of id -> closing commit SHA ("" allowed). Absent ⇒ nil. A present value that
// is not such an object, an empty or oversized id, an id with a control
// character (the id is echoed into the legs' focus text), or a value that is not
// a commit SHA is an error.
func readBlockingClosed(req mcp.CallToolRequest) (map[string]string, error) {
	raw, present := req.GetArguments()[blockingClosedArg]
	if !present {
		return nil, nil
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must be an object mapping each BLOCKING id to its closing commit SHA, got %T", blockingClosedArg, raw)
	}
	out := make(map[string]string, len(obj))
	for id, v := range obj {
		id = strings.TrimSpace(id)
		if id == "" || len(id) > maxClosureIDBytes || strings.ContainsFunc(id, unicode.IsControl) {
			return nil, fmt.Errorf("%s has an unusable id %q (1-%d characters, no control characters)", blockingClosedArg, id, maxClosureIDBytes)
		}
		commit, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("%s[%s] must be a commit SHA string (or \"\" for an id confirmed in an earlier round), got %T", blockingClosedArg, id, v)
		}
		commit = strings.TrimSpace(commit)
		if commit != "" && !deltaBaseHex.MatchString(commit) {
			return nil, fmt.Errorf("%s[%s] %q is not a commit SHA (7-64 hex digits)", blockingClosedArg, id, commit)
		}
		out[id] = commit
	}
	return out, nil
}

// readClaudeVerdict extracts + assembles the ReviewOutput from the structured
// claude_verdict argument. Returns (zero, false) when the argument is absent OR
// not an object. runMultiAudit applies the origin-sensitive fallback
// structurally (NOT a hard error).
func readClaudeVerdict(req mcp.CallToolRequest) (ReviewOutput, bool) {
	raw, ok := req.GetArguments()["claude_verdict"]
	if !ok || raw == nil {
		return ReviewOutput{}, false
	}
	// Round-trip through JSON so a map[string]any literal AND a typed struct
	// both decode cleanly — the MCP runtime delivers arguments as map[string]any.
	b, err := json.Marshal(raw)
	if err != nil {
		return ReviewOutput{}, false
	}
	var out ReviewOutput
	if err := json.Unmarshal(b, &out); err != nil {
		return ReviewOutput{}, false
	}
	return out, true
}

// planSourceConfig is the plan_source value of a result whose gates came, in
// whole or in part, from the audited tree's configuration.
const planSourceConfig = "config"

// readGatesArgument reads the optional `gates` object argument as SUPPLIED-ONLY:
// a backend the caller did not name, or named with a value outside
// off|advisory|required, reads as the empty string ("not supplied"). The
// resolver then decides that backend's gate from the tree's configuration or the
// distributed default; filling the defaults here would make every call look like
// it supplied all three gates and erase the difference between "the caller said
// required" and "the caller said nothing".
//
// An invalid supplied value is deliberately NOT an error: the hard-error surface
// of this handler is the unusable project_root and an invalid CONFIGURATION, and
// a value the call itself carries has never been one.
func readGatesArgument(req mcp.CallToolRequest) config.AuditGates {
	args, ok := req.GetArguments()["gates"].(map[string]any)
	if !ok {
		args = map[string]any{}
	}
	return config.AuditGates{
		Claude: suppliedGate(args["claude"]),
		Codex:  suppliedGate(args["codex"]),
		GLM:    suppliedGate(args["glm"]),
	}
}

// suppliedGate returns v as a trimmed gate token when it is a string in
// off|advisory|required, and "" (not supplied) otherwise.
func suppliedGate(v any) string {
	g := strings.TrimSpace(stringGateArg(v))
	for _, valid := range config.ValidAuditGates() {
		if g == valid {
			return g
		}
	}
	return ""
}

// planGates flattens a resolved plan into the per-backend gates the fan-out
// reads. Every backend carries a gate, so the fan-out's own defaulting never
// applies to a call that went through the resolver.
func planGates(plan config.AuditPlan) config.AuditGates {
	var g config.AuditGates
	for _, e := range plan.Backends {
		switch e.Backend {
		case BackendClaude:
			g.Claude = e.Gate
		case BackendCodex:
			g.Codex = e.Gate
		case BackendGLM:
			g.GLM = e.Gate
		}
	}
	return g
}

// callStartEnforcement returns the gates the unmet-gate enforcement keys on,
// from the audit section read at call start: the gates the operator wrote
// (audit.gates or the audit.model token — never the distributed default), with
// a gate the call supplied taking precedence. A supplied off or advisory
// replaces a configured required, so the entry's gate and the enforcement agree
// (REQ-ACV-006). A supplied required is not itself an opt-in and leaves the
// configured reading in place: it enforces where the tree wrote required and
// stays fail-open where it wrote nothing (design.md §D.2). A config-orphaned
// root whose primary cannot be identified keeps the codex gate assumed required
// and returns the note that says so (REQ-MWU-011/012); the same supplied-gate
// precedence applies to it.
func callStartEnforcement(audit config.AuditConfig, primaryUnidentified bool, supplied config.AuditGates) (config.AuditGates, string) {
	var (
		gates config.AuditGates
		note  string
	)
	if primaryUnidentified {
		gates, note = config.AuditGates{Codex: config.AuditGateRequired}, gateAssumedRequiredNote
	} else if configured, err := config.ResolveAuditPlan(audit, config.AuditGates{}); err == nil {
		// An invalid configured value was rejected by the handler's own resolution
		// before this runs; the error arm is the fail-open reading.
		gates = configured.ExplicitGates()
	}
	for _, g := range []struct{ suppliedGate, enforced *string }{
		{&supplied.Claude, &gates.Claude},
		{&supplied.Codex, &gates.Codex},
		{&supplied.GLM, &gates.GLM},
	} {
		if *g.suppliedGate != "" && *g.suppliedGate != config.AuditGateRequired {
			*g.enforced = *g.suppliedGate
		}
	}
	return gates, note
}

// stringGateArg coerces an `any` argument (string, or nil) into a string gate.
// Non-string values (numbers, bools) return "" so the gate defaults downstream.
func stringGateArg(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// convergenceToolResult shapes a ConvergenceResult as the schema-typed MCP
// result. It reuses the SDK's NewToolResultJSON so the result carries BOTH the
// JSON in TextContent (for the chat-readable channel) AND the typed struct in
// StructuredContent (for callers that want to read fields directly). On a
// marshal failure it degrades to a text summary rather than losing the verdict.
func convergenceToolResult(r ConvergenceResult) (*mcp.CallToolResult, error) {
	res, err := mcp.NewToolResultJSON(r)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("%s: overall=%s (%s)", auditMultiToolName, r.OverallVerdict, r.ResidualRiskNote)), nil
	}
	return res, nil
}
