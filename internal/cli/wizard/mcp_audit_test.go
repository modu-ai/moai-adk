package wizard

// mcp_audit_test.go — SPEC-MOAI-MCP-SERVER-001 M4 (REQ-MCP-015 / AC-MCP-020).
//
// These tests pin the audit + MCP opt-in selection surfaced on page 3 of the
// `moai init` wizard. The selection reuses the M3 typed config vocabulary
// (AuditModel* / AuditGate* constants from internal/config) so the wizard and
// the audit backend share one interpreter — no fork (AC-MCP-021 shares the
// same interpreter via the web console schema fields).

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestPage3Questions_AuditSelectionSurfaced verifies AC-MCP-020: page 3
// presents audit_model, per-auditor audit_gate (claude/codex/glm),
// codex_audit_enabled, and mcp_tools_opt_in.
func TestPage3Questions_AuditSelectionSurfaced(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	questions := Page3Questions(tmpDir)

	wantIDs := []string{
		"audit_model",
		"audit_gate_claude",
		"audit_gate_codex",
		"audit_gate_glm",
		"codex_audit_enabled",
		"mcp_tools_opt_in",
	}
	for _, id := range wantIDs {
		q := QuestionByID(questions, id)
		if q == nil {
			t.Errorf("page-3 question %q is missing (AC-MCP-020)", id)
			continue
		}
		// The audit questions MUST be ungated (visible on every page-3 pass).
		if q.Condition != nil {
			t.Errorf("page-3 question %q carries a Condition — audit selection must be ungated", id)
		}
	}
}

// TestAuditModelQuestion_UsesM3EnumVocabulary verifies the wizard reuses the
// M3 typed-config enum constants (the identical interpreter) rather than
// redeclaring the audit_model values. No fork (REQ-MCP-015).
func TestAuditModelQuestion_UsesM3EnumVocabulary(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	q := QuestionByID(Page3Questions(tmpDir), "audit_model")
	if q == nil {
		t.Fatal("audit_model question missing")
	}
	wantValues := []string{
		config.AuditModelClaude,
		config.AuditModelCodex,
		config.AuditModelGLM,
		config.AuditModelMulti,
	}
	if len(q.Options) != len(wantValues) {
		t.Fatalf("audit_model has %d options, want %d", len(q.Options), len(wantValues))
	}
	got := make(map[string]bool, len(q.Options))
	for _, o := range q.Options {
		got[o.Value] = true
	}
	for _, w := range wantValues {
		if !got[w] {
			t.Errorf("audit_model options missing M3 enum value %q", w)
		}
	}
}

// TestAuditGateQuestions_UsesM3EnumVocabulary verifies each per-auditor
// audit_gate question reuses the M3 AuditGate* constants (off/advisory/required).
func TestAuditGateQuestions_UsesM3EnumVocabulary(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	wantValues := []string{config.AuditGateOff, config.AuditGateAdvisory, config.AuditGateRequired}
	for _, id := range []string{"audit_gate_claude", "audit_gate_codex", "audit_gate_glm"} {
		q := QuestionByID(Page3Questions(tmpDir), id)
		if q == nil {
			t.Fatalf("%s question missing", id)
		}
		if len(q.Options) != len(wantValues) {
			t.Fatalf("%s has %d options, want %d", id, len(q.Options), len(wantValues))
		}
		got := make(map[string]bool, len(q.Options))
		for _, o := range q.Options {
			got[o.Value] = true
		}
		for _, w := range wantValues {
			if !got[w] {
				t.Errorf("%s options missing M3 enum value %q", id, w)
			}
		}
	}
}

// TestAuditModelQuestion_DefaultIsClaude verifies the distributed default
// audit_model is claude (the locked M3 default profile — REQ-MCP-010 / §G.3).
func TestAuditModelQuestion_DefaultIsClaude(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	q := QuestionByID(Page3Questions(tmpDir), "audit_model")
	if q == nil {
		t.Fatal("audit_model question missing")
	}
	if q.Default != config.AuditModelClaude {
		t.Errorf("audit_model default = %q, want %q (claude)", q.Default, config.AuditModelClaude)
	}
}

// TestAuditGateQuestions_DefaultProfile verifies the per-auditor gate defaults
// match the locked M3 default profile (§G.3): claude + codex required, glm
// advisory. glm ships advisory (NOT required) so a distributed user with no GLM
// key is never hard-blocked (C2 fail-open).
func TestAuditGateQuestions_DefaultProfile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cases := []struct {
		id        string
		wantDefault string
	}{
		{"audit_gate_claude", config.AuditGateRequired},
		{"audit_gate_codex", config.AuditGateRequired},
		{"audit_gate_glm", config.AuditGateAdvisory},
	}
	for _, c := range cases {
		q := QuestionByID(Page3Questions(tmpDir), c.id)
		if q == nil {
			t.Fatalf("%s question missing", c.id)
		}
		if q.Default != c.wantDefault {
			t.Errorf("%s default = %q, want %q", c.id, q.Default, c.wantDefault)
		}
	}
}

// TestOptInFlags_DefaultOff verifies codex_audit_enabled + mcp_tools_opt_in
// default to false (C6 opt-in default-off — REQ-MCP-002). A fresh project that
// accepts the wizard defaults ships BOTH surfaces inert.
func TestOptInFlags_DefaultOff(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	for _, id := range []string{"codex_audit_enabled", "mcp_tools_opt_in"} {
		q := QuestionByID(Page3Questions(tmpDir), id)
		if q == nil {
			t.Fatalf("%s question missing", id)
		}
		if q.Type != QuestionTypeConfirm {
			t.Errorf("%s type = %v, want Confirm", id, q.Type)
		}
		if q.Default != "false" {
			t.Errorf("%s default = %q, want \"false\" (opt-in default-off, C6)", id, q.Default)
		}
	}
}
