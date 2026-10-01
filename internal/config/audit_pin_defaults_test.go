package config

import "testing"

// TestDefaultWorkflowConfig_AuditPinsOperatorTable pins the Go default audit
// pins byte-exact against the operator table (SPEC-AGENT-TIER-001 REQ-TIER-004
// / AC-TIER-002): claude {claude-opus-5-5, high}, codex {gpt-6.1-sol, high},
// glm {glm-5.3, max}. The GLM pin ships NON-EMPTY from this SPEC on
// (REQ-TIER-006 — it replaces the empty default); it is audit-only and never
// task delegation (REQ-AMP-008 lineage).
//
// RED-now (run start): the claude cell carried effort "medium" and the GLM
// cell was absent — this test failed on both before M1 flipped the block.
func TestDefaultWorkflowConfig_AuditPinsOperatorTable(t *testing.T) {
	a := NewDefaultWorkflowConfig().Audit

	if a.Claude.Model != "claude-opus-5-5" || a.Claude.Effort != "high" {
		t.Errorf("default Claude pin = {%s %s}, want {claude-opus-5-5 high}", a.Claude.Model, a.Claude.Effort)
	}
	if a.Codex.Model != "gpt-6.1-sol" || a.Codex.Effort != "high" {
		t.Errorf("default Codex pin = {%s %s}, want {gpt-6.1-sol high}", a.Codex.Model, a.Codex.Effort)
	}
	if a.GLM.Model != "glm-5.3" || a.GLM.Effort != "max" {
		t.Errorf("default GLM pin = {%s %s}, want {glm-5.3 max} (non-empty per REQ-TIER-006)", a.GLM.Model, a.GLM.Effort)
	}
}
