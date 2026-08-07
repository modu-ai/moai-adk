// Package cli — `audit_multi` MCP tool handler
// (SPEC-AUDIT-MULTI-MODEL-001 M3, REQ-AMM-009 / REQ-AMM-010 / AC-AMM-012 /
// AC-AMM-013 / AC-AMM-014).
//
// mcp_audit_multi.go is a THIN WRAPPER over runMultiAudit (mcp_convergence.go).
// It maps the JSON-RPC tool params → (claude_verdict ReviewOutput, target,
// focus, MultiAuditConfig), calls runMultiAudit, and shapes the
// ConvergenceResult into the tool's declared output. The handler does NOT
// re-implement the codex/glm backends (C1 — additive, no fork), does NOT call
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

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/config"
)

// auditMultiToolName is the canonical MCP tool name. The cross-model audit
// Skill (moai-ref-cross-model-audit, M4) references it verbatim so a future
// rename propagates via grep — keeping it as a const makes that grep reliable.
const auditMultiToolName = "audit_multi"

// handleAuditMulti is the thin-wrapper MCP tool handler for `audit_multi`
// (REQ-AMM-010 / AC-AMM-013). It:
//  1. Assembles the claude_verdict ReviewOutput (the always-available anchor)
//     from the structured claude_verdict argument.
//  2. Reads the per-auditor audit_gate from the `gates` argument (with
//     distributed defaults applied for any gate the caller omits).
//  3. Fans out by calling runMultiAudit — which reuses the existing
//     codex/glm handler paths (NO backend re-implementation, AC-AMM-013).
//  4. Shapes the ConvergenceResult into the tool's declared output.
//
// The handler NEVER invokes AskUserQuestion (subagent boundary, REQ-AMM-018):
// a missing-input or inconclusive condition is returned as a structured
// ConvergenceResult and the orchestrator translates it through its own channel.
func handleAuditMulti(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	claudeVerdict, ok := readClaudeVerdict(req)
	if !ok {
		// Missing/malformed claude_verdict anchor — fall through with an empty
		// ReviewOutput so runMultiAudit applies its DQ-2 refusal (structured
		// overall=fail + a residual_risk_note). NEVER a hard error.
		claudeVerdict = ReviewOutput{}
	}

	target := req.GetString("target", "")
	focus := req.GetString("focus", "")
	gates := readGatesArgument(req)

	// The convergence engine writes its per-session state file under
	// .moai/state/audit-multi/<session>.json (DQ-1). The tool surface does NOT
	// invent a session id — it threads whatever the caller supplied (the
	// orchestrator resolves the authoritative id). An empty SessionID makes
	// persistence a no-op, which is fail-open safe.
	cfg := MultiAuditConfig{
		Gates:     gates,
		SessionID: req.GetString("session_id", ""),
	}

	result := runMultiAudit(ctx, claudeVerdict, target, focus, cfg)
	return convergenceToolResult(result)
}

// readClaudeVerdict extracts + assembles the ReviewOutput from the structured
// claude_verdict argument. Returns (zero, false) when the argument is absent OR
// not an object — the caller lets runMultiAudit's DQ-2 refusal handle it
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

// readGatesArgument reads the optional `gates` object argument and falls back
// to the distributed-default AuditGates when absent or partial. The defaults
// (claude required, codex required, glm advisory) are applied via the existing
// gateOr helper (mcp_convergence.go) so an explicit omission and a partial
// override both converge to the same behavior runMultiAudit sees.
func readGatesArgument(req mcp.CallToolRequest) config.AuditGates {
	args, ok := req.GetArguments()["gates"].(map[string]any)
	if !ok {
		args = map[string]any{}
	}
	return config.AuditGates{
		Claude: gateOr(stringGateArg(args["claude"]), config.AuditGateRequired),
		Codex:  gateOr(stringGateArg(args["codex"]), config.AuditGateRequired),
		GLM:    gateOr(stringGateArg(args["glm"]), config.AuditGateAdvisory),
	}
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
