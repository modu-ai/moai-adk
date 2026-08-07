package cli

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// SPEC-MOAI-MCP-SERVER-001 M3 — 3-way audit selection + secret hygiene
// (REQ-MCP-010/011/014, AC-MCP-012/013/016/017). RED until
// internal/cli/mcp_audit.go exists.

func TestActiveAuditBackend_SingleBackends(t *testing.T) {
	// AC-MCP-017: claude/codex/glm each resolve to exactly one active backend.
	for _, model := range []string{config.AuditModelClaude, config.AuditModelCodex, config.AuditModelGLM} {
		got, err := activeAuditBackend(model)
		if err != nil {
			t.Errorf("model %q: unexpected error %v", model, err)
			continue
		}
		if got != model {
			t.Errorf("activeAuditBackend(%q) = %q, want %q", model, got, model)
		}
	}
}

func TestActiveAuditBackend_MultiTokenAccepted(t *testing.T) {
	// AC-MCP-017: `multi` is accepted as a stored value. Its convergence logic
	// was deferred by SPEC-MOAI-MCP-SERVER-001 M3 (AP-8) and IMPLEMENTED by
	// SPEC-AUDIT-MULTI-MODEL-001 M1 (multiConvergenceImplemented flipped
	// false → true in the same commit as internal/cli/mcp_convergence.go).
	got, err := activeAuditBackend(config.AuditModelMulti)
	if err != nil {
		t.Fatalf("multi token rejected: %v (must be accepted)", err)
	}
	if got != config.AuditModelMulti {
		t.Errorf("activeAuditBackend(multi) = %q, want multi (stored verbatim)", got)
	}
	// SPEC-AUDIT-MULTI-MODEL-001 M1: the sentinel MUST now be true (the engine
	// exists). A false value here would mean the sentinel lies — a missing-engine
	// hazard (R3).
	if !multiConvergenceImplemented {
		t.Error("multiConvergenceImplemented = false; want true (SPEC-AUDIT-MULTI-MODEL-001 M1 engine is present)")
	}
}

func TestActiveAuditBackend_RejectsUnknown(t *testing.T) {
	if _, err := activeAuditBackend("grok"); err == nil {
		t.Error("activeAuditBackend accepted unknown model 'grok' (must reject)")
	}
}

// TestBuildAuditEnvBlock_SecretHygiene_NegativeTest is the LOAD-BEARING
// AC-MCP-013 guard. Even when a real GLM key is resolvable in the environment,
// the env block written for the audit backends MUST carry only the ${VAR}
// literal — never the resolved secret value. And the committed moai entry
// (buildMoaiMCPServerEntry) carries NO env block at all.
func TestBuildAuditEnvBlock_SecretHygiene_NegativeTest(t *testing.T) {
	const fakeSecret = "GLM-sk-DO-NOT-LEAK-1234567890abcdef"
	withGLMSeams(t, fakeSecret, nil) // key loader now returns a real-looking secret

	t.Run("env block uses ${VAR} literal, never the resolved key", func(t *testing.T) {
		env := buildAuditEnvBlock(config.AuditModelGLM)
		if env == nil {
			t.Fatal("buildAuditEnvBlock(glm) = nil; want a ${GLM_API_KEY} literal env")
		}
		val, ok := env["GLM_API_KEY"]
		if !ok {
			t.Fatalf("env = %v; want GLM_API_KEY entry", env)
		}
		if val != "${GLM_API_KEY}" {
			t.Errorf("GLM_API_KEY = %q, want literal ${GLM_API_KEY}", val)
		}
		// The resolved secret MUST NOT appear anywhere in the serialized env.
		b, _ := json.Marshal(env)
		if strings.Contains(string(b), fakeSecret) {
			t.Errorf("resolved secret leaked into env block: %s", b)
		}
	})

	t.Run("committed moai entry has no env block (local stdio, no secrets)", func(t *testing.T) {
		entry := buildMoaiMCPServerEntry()
		if _, ok := entry["env"]; ok {
			t.Errorf("committed moai entry must not carry an env block, got: %v", entry["env"])
		}
		b, _ := json.Marshal(entry)
		if strings.Contains(string(b), fakeSecret) {
			t.Errorf("resolved secret leaked into committed moai entry: %s", b)
		}
		if strings.Contains(string(b), "${") {
			t.Errorf("committed moai entry should carry no ${VAR} literal either (no env block): %s", b)
		}
	})

	t.Run("multi env block covers both backends with literals", func(t *testing.T) {
		env := buildAuditEnvBlock(config.AuditModelMulti)
		if env["GLM_API_KEY"] != "${GLM_API_KEY}" {
			t.Errorf("multi GLM_API_KEY = %q, want ${GLM_API_KEY}", env["GLM_API_KEY"])
		}
		if env["CODEX_API_KEY"] != "${CODEX_API_KEY}" {
			t.Errorf("multi CODEX_API_KEY = %q, want ${CODEX_API_KEY}", env["CODEX_API_KEY"])
		}
		b, _ := json.Marshal(env)
		if strings.Contains(string(b), fakeSecret) {
			t.Errorf("resolved secret leaked into multi env block: %s", b)
		}
	})

	t.Run("claude env block is empty (no secrets to provision)", func(t *testing.T) {
		if env := buildAuditEnvBlock(config.AuditModelClaude); env != nil {
			t.Errorf("claude env = %v; want nil (claude needs no backend key)", env)
		}
	})
}

// TestMCPAudit_NoAskUserQuestion is the package-wide subagent-boundary guard
// (AC-MCP-016 / C-HRA-008). It greps the M3 audit handler sources for any
// AskUserQuestion / mcp__askuser reference outside tests + comments — 0 actual
// calls required.
func TestMCPAudit_NoAskUserQuestion(t *testing.T) {
	files := []string{
		"mcp_glm.go",
		"mcp_audit.go",
	}
	for _, f := range files {
		out, err := exec.Command("grep", "-n", "-E", "AskUserQuestion|mcp__askuser", f).Output()
		if err != nil && err.Error() != "" && !strings.Contains(err.Error(), "exit status 1") {
			// grep exit 1 = no matches (the desired outcome). Any other exec
			// error is a test-infra failure.
			t.Fatalf("grep %s: %v", f, err)
		}
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, ln := range lines {
			ln = strings.TrimSpace(ln)
			if ln == "" {
				continue
			}
			if strings.Contains(ln, "//") || strings.Contains(ln, "_test.go") {
				continue
			}
			t.Errorf("%s: unexpected AskUserQuestion reference: %s", f, ln)
		}
	}
}

// TestMCPAudit_NoDirectFrontmatterRead (AC-MCP-015) — the model/effort SSOT
// invariant: no direct agent-frontmatter or llm.agent_overrides read in the
// MCP audit package. Resolution MUST go through template.ResolveAgentModelEffort.
func TestMCPAudit_NoDirectFrontmatterRead(t *testing.T) {
	files := []string{"mcp_glm.go", "mcp_audit.go", "mcp_codex.go"}
	for _, f := range files {
		out, _ := exec.Command("grep", "-n", "-E", "AgentOverrides|agentfm|ReadFrontmatter|ParseFrontmatter", f).Output()
		for _, ln := range strings.Split(string(out), "\n") {
			ln = strings.TrimSpace(ln)
			if ln == "" || strings.Contains(ln, "//") {
				continue
			}
			t.Errorf("%s: direct frontmatter/override read forbidden (use ResolveAgentModelEffort): %s", f, ln)
		}
	}
}
