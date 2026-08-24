package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// resolveGLMAuditModelEffort has three fallback returns — an unreadable llm.yaml, an
// unmapped SSOT lookup, and a non-GLM session whose resolved id is a Claude
// model z.ai cannot serve. All three hand back the same constant, so the value
// of that constant is the whole contract.
//
// It went stale once: the constant sat on a two-generation-old id while the tier
// defaults moved on, and the pre-existing test only asserted the result was
// non-empty and not a Claude id — both true of the stale value. These tests pin
// the value instead.

// TestGLMAuditDefaultModel_DerivesFromTierDefault is the anti-drift guard. The
// fallback is not an independent choice; it is whatever the launcher injects.
// Restating it as its own literal is what let the two diverge.
func TestGLMAuditDefaultModel_DerivesFromTierDefault(t *testing.T) {
	t.Parallel()

	if glmAuditDefaultModel != config.DefaultGLMHigh {
		t.Errorf("glmAuditDefaultModel = %q, want it to track config.DefaultGLMHigh (%q) — "+
			"a separate literal drifts away from the model the launcher actually injects",
			glmAuditDefaultModel, config.DefaultGLMHigh)
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

	if got := resolveGLMAuditModelEffort().Model; got != config.DefaultGLMHigh {
		t.Errorf("unreadable llm.yaml: resolveGLMAuditModelEffort().Model = %q, want %q", got, config.DefaultGLMHigh)
	}
}

// TestResolveGLMAuditModel_NonGLMSession covers the path a Claude session takes
// when it calls glm_audit for a cross-model second opinion — the common case.
// The SSOT resolves a Claude id, which z.ai cannot serve, so the fallback runs.
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

	if got := resolveGLMAuditModelEffort().Model; got != config.DefaultGLMHigh {
		t.Errorf("non-GLM session: resolveGLMAuditModelEffort().Model = %q, want %q", got, config.DefaultGLMHigh)
	}
}
