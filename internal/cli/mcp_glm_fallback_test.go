package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// resolveGLMAuditModelEffort once had three fallback returns — an unreadable llm.yaml, an
// unmapped SSOT lookup, and a non-GLM session whose resolved id is a Claude
// model z.ai cannot serve. All three handed back the same constant, and every
// unpinned call now does, so the value of that constant is the whole contract.
//
// It went stale once: the constant sat on a two-generation-old id while the tier
// defaults moved on, and the pre-existing test only asserted the result was
// non-empty and not a Claude id — both true of the stale value. These tests pin
// the value instead.

// TestGLMAuditDefaultModel_DerivesFromAuditPin is the anti-drift guard
// (updated by SPEC-AGENT-TIER-001 in the same commit as the default it pins:
// the audit fallback is the operator pin target — full glm-5.3 with effort
// max — while the glm_task delegation default keeps the flash slot default,
// REQ-AMP-008). Restating either as its own literal is what let the fallback
// drift once before.
func TestGLMAuditDefaultModel_DerivesFromAuditPin(t *testing.T) {
	t.Parallel()

	if glmAuditDefaultModel != config.DefaultGLM53 {
		t.Errorf("glmAuditDefaultModel = %q, want it to track config.DefaultGLM53 (%q) — "+
			"the audit pin targets full glm-5.3, not the flash slot default",
			glmAuditDefaultModel, config.DefaultGLM53)
	}
	if glmAuditDefaultEffort != "max" {
		t.Errorf("glmAuditDefaultEffort = %q, want %q (the pin effort, forwarded verbatim)", glmAuditDefaultEffort, "max")
	}
	if glmTaskDefaultModel != config.DefaultGLMHigh {
		t.Errorf("glmTaskDefaultModel = %q, want it to keep tracking config.DefaultGLMHigh (%q) — "+
			"the glm_task delegation default is unchanged (REQ-AMP-008)",
			glmTaskDefaultModel, config.DefaultGLMHigh)
	}
}

// TestResolveGLMAuditModel_UnreadableLLMYAML covers the load-error fallback: a
// present-but-unreadable llm.yaml. Written as a directory so os.Stat succeeds
// (the not-exist branch returns defaults instead) and os.ReadFile fails.
func TestResolveGLMAuditModel_UnreadableLLMYAML(t *testing.T) {
	root := t.TempDir()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(filepath.Join(sections, "llm.yaml"), 0o755); err != nil {
		t.Fatalf("seed unreadable llm.yaml: %v", err)
	}

	t.Setenv("CLAUDE_PROJECT_DIR", "")
	old := projectDirResolver
	projectDirResolver = func() string { return root }
	t.Cleanup(func() { projectDirResolver = old })

	got := resolveGLMAuditModelEffort("")
	if got.Model != config.DefaultGLM53 || got.Effort != "max" {
		t.Errorf("unreadable llm.yaml: resolveGLMAuditModelEffort() = {%s %s}, want {%s max}", got.Model, got.Effort, config.DefaultGLM53)
	}
}

// TestResolveGLMAuditModel_NonGLMSession covers the path a Claude session takes
// when it calls glm_audit for a cross-model second opinion — the common case.
// A leftover Claude-id cell never reaches z.ai: without a pin the backend default runs.
func TestResolveGLMAuditModel_NonGLMSession(t *testing.T) {
	root := t.TempDir()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	// mode/team_mode absent ⇒ not a GLM backend; the matrix maps sync-auditor to
	// a Claude model.
	llm := "llm:\n" +
		"  mode: \"\"\n" +
		"  team_mode: \"\"\n" +
		"  profile: \"medium\"\n" +
		"  profiles:\n" +
		"    medium:\n" +
		"      sync-auditor: { model: opus, effort: medium }\n"
	if err := os.WriteFile(filepath.Join(sections, "llm.yaml"), []byte(llm), 0o644); err != nil {
		t.Fatalf("write llm.yaml: %v", err)
	}

	t.Setenv("CLAUDE_PROJECT_DIR", "")
	old := projectDirResolver
	projectDirResolver = func() string { return root }
	t.Cleanup(func() { projectDirResolver = old })

	got := resolveGLMAuditModelEffort("")
	if got.Model != config.DefaultGLM53 || got.Effort != "max" {
		t.Errorf("non-GLM session: resolveGLMAuditModelEffort() = {%s %s}, want {%s max}", got.Model, got.Effort, config.DefaultGLM53)
	}
}
