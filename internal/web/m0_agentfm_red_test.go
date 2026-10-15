// m0_agentfm_red_test.go — SPEC-USERASSET-DEPLOY-GUARD-001 M0, web family:
// AC-015 (same-name agent consolidated single-row end-to-end).
//
// M0 discipline: observation only — no production change.
package web

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestAgentFormSameNameConsolidatedEndToEnd — AC-015 (ledger 8a,
// REQ-SRF-004). Given the SAME agent name in the user scope
// (~/.claude/agents) and the project scope (.claude/agents/moai):
// (1) the form renders ONE consolidated row carrying the scope provenance —
// not duplicate rows; (2) one edit round-trips parse (pins[name]) → save
// (applyAgentOverrides → llm.agent_overrides) → re-read consistently, and
// the two-step store-reread procedure never leaves two values surviving.
// RED-now reason: agentDirsFor yields BOTH scopes and the merge concatenates
// them (agentfm.go:123-132, :200-206), so the same name renders twice; the
// form keys are name-only (:282-283), so both duplicate rows submit under
// one key.
func TestAgentFormSameNameConsolidatedEndToEnd(t *testing.T) {
	a, root := seedAgentOverridesProject(t)
	writeLLMYAML(t, root, "llm:\n  mode: \"\"\n  glm_env_var: GLM_API_KEY\n")

	// The user scope carries the same name as the project row: a pre-flight
	// state the JD-20 migration creates whenever a pre-migration project
	// copy outlives its user counterpart.
	home := t.TempDir()
	userAgents := filepath.Join(home, ".claude", "agents")
	if err := os.MkdirAll(userAgents, 0o755); err != nil {
		t.Fatal(err)
	}
	userBody := "---\nname: manager-develop\ndescription: User-scope twin of the project row\n---\nbody\n"
	if err := os.WriteFile(filepath.Join(userAgents, "manager-develop.md"), []byte(userBody), 0o644); err != nil {
		t.Fatal(err)
	}

	prevHome := homeAgentsDir
	homeAgentsDir = userAgents
	t.Cleanup(func() { homeAgentsDir = prevHome })

	llm := readLLMForTest(t, root)

	// Arm 1 — the render contract: one consolidated row, provenance carried.
	rows, err := a.listAllAgentFMs(root, llm)
	if err != nil {
		t.Fatalf("listAllAgentFMs: %v", err)
	}
	count := 0
	for _, row := range rows {
		if row.Name == "manager-develop" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("RED (intended): the same-name agent rendered %d rows (want 1 consolidated row with scope provenance) — agentDirsFor concatenates both scopes without dedup", count)
	}

	// Arm 2 — the store-reread procedure: two sequential submissions with
	// DIFFERENT values; the store must always equal the LAST submission.
	submit := func(model string) {
		form := map[string][]string{
			"agentfm.manager-develop.model":  {model},
			"agentfm.manager-develop.effort": {"xhigh"},
		}
		req := httptest.NewRequest("POST", "/settings/agent-overrides", nil)
		req.PostForm = form
		req.Form = form
		agents := rows
		pins, submitted, errs := parseAgentFMForm(req, agents, llm, "")
		if len(errs) > 0 {
			t.Fatalf("parse errors: %v", errs)
		}
		if err := applyAgentOverrides(root, pins, submitted); err != nil {
			t.Fatalf("applyAgentOverrides(%s): %v", model, err)
		}
	}
	submit("opus")
	afterFirst := readStoredOverride(t, root)
	submit("sonnet")
	afterSecond := readStoredOverride(t, root)

	if afterSecond != "sonnet" {
		t.Fatalf("RED-observed: after the two-step store-reread procedure the stored override is %q (want the last submission %q); first step stored %q — two values surviving or first-row precedence", afterSecond, "sonnet", afterFirst)
	}
}

// readLLMForTest loads the project's LLM config the way the app's read seam does.
func readLLMForTest(t *testing.T, root string) config.LLMConfig {
	t.Helper()
	cfg, err := config.NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	return cfg.LLM
}

// readStoredOverride reads llm.agent_overrides["manager-develop"].Model from
// the project config — the re-read arm of the store contract.
func readStoredOverride(t *testing.T, root string) string {
	t.Helper()
	cfg, err := config.NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("re-read config: %v", err)
	}
	me, ok := cfg.LLM.AgentOverrides["manager-develop"]
	if !ok {
		return "" // absent override reads as the profile default, not a second value
	}
	return me.Model
}
