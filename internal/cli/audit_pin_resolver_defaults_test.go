package cli

import (
	"testing"
)

// TestAuditResolverTerminalFallbacks_OperatorTable pins the three audit
// resolvers' terminal fallbacks (no user pin, no explicit caller override)
// against the operator table (SPEC-AGENT-TIER-001 REQ-TIER-004 / AC-TIER-004):
//
//	claude → {claude-opus-5-5, high}
//	codex  → {gpt-6.1-sol, high}   (unchanged by this SPEC)
//	glm    → {glm-5.3, max}        (effort forwarded verbatim, REQ-AMP-006)
//
// The GLM model is the FULL glm-5.3 (config.DefaultGLM53), NOT the flash slot
// default (config.DefaultGLMHigh) that bound here before this SPEC — the pin
// targets full glm-5.3 (AC-TIER-015).
//
// RED-now (run start): the claude fallback returned effort "medium" and the
// GLM fallback returned {glm-5.3-flash, ""} — this test failed on both before
// M1 flipped the fallback constants.
func TestAuditResolverTerminalFallbacks_OperatorTable(t *testing.T) {
	root := newGLMReviewTree(t, false) // a tree with NO workflow.yaml → absent pin

	claude := resolveClaudeAuditModelEffort(root, "", "")
	if claude.Model != "claude-opus-5-5" || claude.Effort != "high" {
		t.Errorf("claude terminal fallback = {%s %s}, want {claude-opus-5-5 high}", claude.Model, claude.Effort)
	}

	glme := resolveGLMAuditModelEffort(root)
	if glme.Model != "glm-5.3" || glme.Effort != "max" {
		t.Errorf("glm terminal fallback = {%s %s}, want {glm-5.3 max}", glme.Model, glme.Effort)
	}

	codex := resolveCodexAuditModelEffort(map[string]any{"cwd": root})
	if codex.Model != "gpt-6.1-sol" || codex.Effort != "high" {
		t.Errorf("codex terminal fallback = {%s %s}, want {gpt-6.1-sol high}", codex.Model, codex.Effort)
	}
}
