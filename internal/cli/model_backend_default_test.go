package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// The codex and GLM tools resolve their model as audit pin > backend default
// (SPEC-AGENT-MODEL-INHERIT-001 REQ-AMI-019, design D5). A per-agent cell in
// llm.yaml — whether an agent_overrides entry or a profile matrix column — no
// longer reaches either backend: the per-agent assignment is gone, so a
// leftover cell must not act as a hidden pin.

func writeBackendDefaultLLM(t *testing.T, root, body string) {
	t.Helper()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "llm.yaml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestCodexResolution_IgnoresPerAgentLLMCells(t *testing.T) {
	root := t.TempDir()
	writeBackendDefaultLLM(t, root,
		"llm:\n  agent_overrides:\n    sync-auditor:\n      model: gpt-5-codex\n      effort: high\n")

	if got := resolveCodexModelEffort(map[string]any{"cwd": root}); got != (config.ModelEffort{}) {
		t.Errorf("codex task = %+v, want the zero value (codex applies its own default)", got)
	}
	// SPEC-MODEL-MATRIX-UPDATE-001 REQ-MMU-001: without a project pin the AUDIT
	// path lands on the terminal fallback pin — no longer the zero value
	// (REQ-AMP-005 neutrality superseded by operator directive 2026-09-30).
	if got := resolveCodexAuditModelEffort(map[string]any{"cwd": root}); got != (config.ModelEffort{Model: "gpt-6.1-sol", Effort: "high"}) {
		t.Errorf("codex audit without pin = %+v, want {gpt-6.1-sol high}", got)
	}
	// An explicit caller model still wins, sent verbatim.
	if got := resolveCodexModelEffort(map[string]any{"cwd": root, "model": " gpt-6.1-sol "}); got != (config.ModelEffort{Model: "gpt-6.1-sol"}) {
		t.Errorf("codex task with explicit model = %+v, want {gpt-6.1-sol, \"\"}", got)
	}
}

func TestGLMResolution_IgnoresPerAgentLLMCells(t *testing.T) {
	root := t.TempDir()
	writeBackendDefaultLLM(t, root,
		"llm:\n  team_mode: glm\n  agent_overrides:\n"+
			"    super-advisor:\n      model: glm-4.6\n      effort: low\n"+
			"    sync-auditor:\n      model: glm-4.6\n      effort: low\n")

	t.Setenv("CLAUDE_PROJECT_DIR", "")
	old := projectDirResolver
	projectDirResolver = func() string { return root }
	t.Cleanup(func() { projectDirResolver = old })

	if got := resolveGLMTaskModel(); got != config.DefaultGLMHigh {
		t.Errorf("glm task default = %q, want %q", got, config.DefaultGLMHigh)
	}
	if got := resolveGLMAuditModelEffort(root); got != (config.ModelEffort{Model: glmAuditDefaultModel, Effort: glmAuditDefaultEffort}) {
		t.Errorf("glm audit without pin = %+v, want the {%s, %s} default pin (REQ-TIER-004 — the former empty-effort backend default fell with the operator pin flip)", got, glmAuditDefaultModel, glmAuditDefaultEffort)
	}
}
