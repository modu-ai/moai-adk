// Package cli — 3-way audit backend selection + secret hygiene
// (SPEC-MOAI-MCP-SERVER-001 M3, REQ-MCP-010/011/014).
//
// mcp_audit.go holds the audit_model/audit_gate contract surface shared by the
// codex (mcp_codex.go), GLM (mcp_glm.go), and multi-convergence
// (mcp_convergence.go) backends:
//   - activeAuditBackend: validates audit_model ∈ {claude,codex,glm,multi} and
//     returns the single active backend. `multi` activates the parallel fan-out
//     + disagreement-synthesis engine in mcp_convergence.go
//     (SPEC-AUDIT-MULTI-MODEL-001 M1).
//   - buildAuditEnvBlock: the ONLY producer of backend env references for
//     provisioning. Every value is a ${VAR} literal expanded by the Claude Code
//     runtime at load — a resolved secret is NEVER serialized (C3 / REQ-MCP-011
//     / AC-MCP-013). The committed moai entry (buildMoaiMCPServerEntry) carries
//     no env block at all: the local stdio server reads ~/.moai/.env.glm and
//     ~/.moai/.env.codex directly via loadGLMKey at runtime, so no secret value
//     needs to reach the git-tracked host config.
//
// @MX:SPEC: SPEC-MOAI-MCP-SERVER-001
// @MX:SPEC: SPEC-AUDIT-MULTI-MODEL-001
package cli

import (
	"fmt"

	"github.com/modu-ai/moai-adk/internal/config"
)

// multiConvergenceImplemented documents that the `multi` audit_model token's
// convergence logic (parallel fan-out + disagreement synthesis) is implemented.
// Flipped from false → true by SPEC-AUDIT-MULTI-MODEL-001 M1, IN THE SAME COMMIT
// as the engine (internal/cli/mcp_convergence.go) — there is no window where
// the sentinel lies (R3 / Definition of Done sentinel-flip same-commit invariant).
// The engine activates `audit_model: multi` end-to-end: parallel errgroup fan-out
// across codex + glm, super-review independence (claude_verdict consumed only by
// the synthesis step), disagreement = advisory (NOT block), fail-open identity.
const multiConvergenceImplemented = true

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

// activeAuditBackend validates audit_model and returns the single active
// backend (AC-MCP-017). For claude/codex/glm the backend is the model itself
// (exactly one executes, per its audit_gate). `multi` activates the convergence
// engine in mcp_convergence.go (multiConvergenceImplemented = true since
// SPEC-AUDIT-MULTI-MODEL-001 M1) — a caller that needs a concrete single
// backend should still treat `multi` as "fan out, then converge".
func activeAuditBackend(model string) (string, error) {
	switch model {
	case config.AuditModelClaude, config.AuditModelCodex, config.AuditModelGLM, config.AuditModelMulti:
		return model, nil
	default:
		return "", fmt.Errorf("audit_model %q unknown (want one of claude|codex|glm|multi)", model)
	}
}

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
