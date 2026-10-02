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
//
// SPEC-WEB-SETTINGS-SAVE-001 REQ-WSS-201 (operator-confirmed 2026-10-01): the
// audit fallback deliberately DIVERGES from the tier default now — {glm-5.3,
// max} (config.DefaultGLM53) while the launcher's tier default stays
// glm-5.3-flash. The former derivation guard
// (TestGLMAuditDefaultModel_DerivesFromTierDefault, RED observed against the
// new value) was rewritten into the divergence pin below.

// TestGLMAuditDefaultModel_PinsOperatorConfirmedAuditPin is the anti-drift
// guard, rewritten for the operator-confirmed audit pin: the fallback tracks
// config.DefaultGLM53 with the max reasoning state — NOT the tier default the
// launcher injects (glm-5.3-flash has no audit meaning here; a second literal
// was what let the fallback go stale once, so the pin stays a named-constant
// derivation rather than a restated literal).
func TestGLMAuditDefaultModel_PinsOperatorConfirmedAuditPin(t *testing.T) {
	t.Parallel()

	if glmAuditDefaultModel != config.DefaultGLM53 {
		t.Errorf("glmAuditDefaultModel = %q, want config.DefaultGLM53 (%q) — "+
			"the operator-confirmed audit pin (REQ-WSS-201), deliberately diverging from the tier default %q",
			glmAuditDefaultModel, config.DefaultGLM53, config.DefaultGLMHigh)
	}
	if glmAuditDefaultEffort != "max" {
		t.Errorf("glmAuditDefaultEffort = %q, want %q (the z.ai max reasoning state, REQ-WSS-201)", glmAuditDefaultEffort, "max")
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

	want := config.ModelEffort{Model: config.DefaultGLM53, Effort: "max"}
	if got := resolveGLMAuditModelEffort(""); got != want {
		t.Errorf("unreadable llm.yaml: resolveGLMAuditModelEffort() = %+v, want %+v", got, want)
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

	want := config.ModelEffort{Model: config.DefaultGLM53, Effort: "max"}
	if got := resolveGLMAuditModelEffort(""); got != want {
		t.Errorf("non-GLM session: resolveGLMAuditModelEffort() = %+v, want %+v", got, want)
	}
}
