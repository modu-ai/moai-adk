package cli

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// SPEC-MOAI-MCP-SERVER-001 M3 — audit secret hygiene
// (REQ-MCP-011/014, AC-MCP-013/016). The audit_model token is validated by
// config.ResolveAuditPlan (see internal/config/audit_plan_test.go).

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

// TestMCPAudit_NoDirectFrontmatterRead (AC-MCP-015) — no agent-frontmatter or
// llm.agent_overrides read in the MCP audit package: MoAI assigns no per-agent
// model, so resolution is audit pin > backend default
// (SPEC-AGENT-MODEL-INHERIT-001 design D5).
func TestMCPAudit_NoDirectFrontmatterRead(t *testing.T) {
	files := []string{"mcp_glm.go", "mcp_audit.go", "mcp_codex.go"}
	for _, f := range files {
		out, _ := exec.Command("grep", "-n", "-E", "AgentOverrides|agentfm|ReadFrontmatter|ParseFrontmatter", f).Output()
		for _, ln := range strings.Split(string(out), "\n") {
			ln = strings.TrimSpace(ln)
			if ln == "" || strings.Contains(ln, "//") {
				continue
			}
			t.Errorf("%s: frontmatter/override read forbidden (resolution is audit pin > backend default): %s", f, ln)
		}
	}
}
