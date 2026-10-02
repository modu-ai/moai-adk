// Package cli — audit backend env hygiene
// (SPEC-MOAI-MCP-SERVER-001 M3, REQ-MCP-011/014).
//
// mcp_audit.go holds the backend env-reference surface shared by the codex
// (mcp_codex.go) and GLM (mcp_glm.go) backends. The audit_model token itself is
// validated and consumed by config.ResolveAuditPlan; the `multi` fan-out is
// orchestrated by the audit_multi tool (mcp_audit_multi.go), not here.
//   - buildAuditEnvBlock: the ONLY producer of backend env references for
//     provisioning. Every value is a ${VAR} literal expanded by the Claude Code
//     runtime at load — a resolved secret is NEVER serialized (C3 / REQ-MCP-011
//     / AC-MCP-013). The committed moai entry (buildMoaiMCPServerEntry) carries
//     no env block at all: the local stdio server reads ~/.moai/.env.glm and
//     ~/.moai/.env.codex directly via loadGLMKey at runtime, so no secret value
//     needs to reach the git-tracked host config.
//
// @MX:SPEC: SPEC-MOAI-MCP-SERVER-001
package cli

import (
	"github.com/modu-ai/moai-adk/internal/config"
)

// ${VAR} literal env-reference tokens. These are Claude Code host-runtime
// expansion tokens (the runtime substitutes them at .mcp.json load), NOT env
// var names moai reads via os.Getenv. They live here as domain constants —
// alongside the audit code that emits them — exactly as the codex domain
// identifiers live in mcp_codex.go. A resolved API key is NEVER written in
// their place (AC-MCP-013).
const (
	glmAPIKeyEnvLiteral   = "${GLM_API_KEY}"
	codexAPIKeyEnvLiteral = "${CODEX_API_KEY}"
	envKeyGLMName         = "GLM_API_KEY"
	envKeyCodexName       = "CODEX_API_KEY"
)

// buildAuditEnvBlock returns the backend env-reference block for a given
// audit_model, or nil when the model needs no backend key (claude).
//
// SECRET HYGIENE (C3 / REQ-MCP-011 / AC-MCP-013 — load-bearing): every value is
// the ${VAR} literal token (glmAPIKeyEnvLiteral / codexAPIKeyEnvLiteral), NEVER
// a resolved secret. The moai MCP server is local stdio and reads its keys
// directly via loadGLMKey at runtime, so this block is the ONLY surface a
// provisioning step would write; it is constructed so that even a caller that
// has a live key in hand cannot accidentally serialize it. The committed moai
// entry (buildMoaiMCPServerEntry) intentionally carries NO env block — this
// helper exists for callers (a future provisioning step, or a third-party
// server entry) that explicitly opt to emit backend env references.
func buildAuditEnvBlock(model string) map[string]string {
	switch model {
	case config.AuditModelGLM:
		return map[string]string{envKeyGLMName: glmAPIKeyEnvLiteral}
	case config.AuditModelCodex:
		return map[string]string{envKeyCodexName: codexAPIKeyEnvLiteral}
	case config.AuditModelMulti:
		return map[string]string{
			envKeyGLMName:   glmAPIKeyEnvLiteral,
			envKeyCodexName: codexAPIKeyEnvLiteral,
		}
	default:
		// claude (or unknown) needs no backend key.
		return nil
	}
}
