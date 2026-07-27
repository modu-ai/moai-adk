package web

// G3 — structural repoint of the sub-agent (agentfm) panel to the llm.yaml
// profile matrix as the single source of truth. These tests pin the intended
// behavior of the batch:
//
//	G3-1  the per-agent select's selected value derives from the runtime profile
//	      matrix (template.ResolveAgentModelEffort), NOT agent frontmatter or the
//	      badge-tier suggestions.
//	G3-2  a per-agent edit persists to llm.agent_overrides (round-trips a save →
//	      reload), and the agent .md frontmatter is NOT touched.
//	G3-3  a per-profile per-agent matrix JSON blob is emitted for the client
//	      tier-repopulation handler, and app.js wires it.
//	G3-4  the perf-tier control shows a preselected "Custom" pseudo-state when
//	      llm.agent_overrides is non-empty.

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// writeLLMProfileYAML writes root/.moai/config/sections/llm.yaml with a profile
// column and optional agent_overrides (name → [model, effort]).
func writeLLMProfileYAML(t *testing.T, root, profile string, overrides map[string][2]string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	var b strings.Builder
	b.WriteString("llm:\n")
	b.WriteString("    profile: \"" + profile + "\"\n")
	if len(overrides) > 0 {
		b.WriteString("    agent_overrides:\n")
		for name, me := range overrides {
			b.WriteString("        " + name + ":\n")
			b.WriteString("            model: " + me[0] + "\n")
			b.WriteString("            effort: " + me[1] + "\n")
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte(b.String()), 0o644); err != nil {
		t.Fatalf("write llm.yaml: %v", err)
	}
}

// TestG3ReadPathDerivesFromProfileMatrix (G3-1): the manager-spec select's
// selected model/effort comes from the max-profile spec_auditors cell (fable/low),
// NOT the seeded frontmatter (opus/xhigh) — proving the read path repointed off
// frontmatter onto the profile matrix.
func TestG3ReadPathDerivesFromProfileMatrix(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-spec", "opus", "xhigh") // frontmatter deliberately divergent
	writeLLMProfileYAML(t, root, "max", nil)

	body := renderAgentFMBody(t, root)

	if !strings.Contains(body, `<option value="fable" selected`) {
		t.Error("manager-spec model select did not select the max-profile cell (fable) — read path still reads frontmatter/badge-tier")
	}
	if !strings.Contains(body, `<option value="low" selected`) {
		t.Error("manager-spec effort select did not select the max-profile cell (low)")
	}
	if strings.Contains(body, `<option value="xhigh" selected`) {
		t.Error("manager-spec effort select shows the FRONTMATTER value (xhigh) — read path must derive from the profile matrix")
	}
}

// TestG3ReadPathOverrideWins (G3-1): an llm.agent_overrides entry wins over the
// profile cell in the rendered selects.
func TestG3ReadPathOverrideWins(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-spec", "opus", "xhigh")
	writeLLMProfileYAML(t, root, "max", map[string][2]string{"manager-spec": {"sonnet", "high"}})

	body := renderAgentFMBody(t, root)

	if !strings.Contains(body, `<option value="sonnet" selected`) {
		t.Error("agent_overrides model (sonnet) did not win over the profile cell in the render")
	}
	if !strings.Contains(body, `<option value="high" selected`) {
		t.Error("agent_overrides effort (high) did not win over the profile cell in the render")
	}
}

// TestG3SaveWritesAgentOverrideNotFrontmatter (G3-2): a per-agent edit persists
// to llm.agent_overrides (survives a save → LoadRaw reload) and leaves the agent
// frontmatter untouched.
func TestG3SaveWritesAgentOverrideNotFrontmatter(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-spec", "opus", "xhigh")
	writeLLMProfileYAML(t, root, "medium", nil)

	a := newPolicyTestApp(root)
	form := baseSaveForm()
	form.Set("performance_tier", "medium")
	// medium/spec_auditors default = opus/high; sonnet differs → a pin.
	form.Set("agentfm.manager-spec.model", "sonnet")
	form.Set("agentfm.manager-spec.effort", "high")

	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("save = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	cfg, err := config.NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	ov, ok := cfg.LLM.AgentOverrides["manager-spec"]
	if !ok {
		t.Fatalf("agent_overrides[manager-spec] not written; overrides=%#v", cfg.LLM.AgentOverrides)
	}
	if ov.Model != "sonnet" || ov.Effort != "high" {
		t.Errorf("override = %+v, want {sonnet high}", ov)
	}
	if m := agentFrontmatterValue(t, root, "manager-spec", "model"); m != "opus" {
		t.Errorf("frontmatter model mutated to %q — writes must go to agent_overrides, not frontmatter", m)
	}
	if e := agentFrontmatterValue(t, root, "manager-spec", "effort"); e != "xhigh" {
		t.Errorf("frontmatter effort mutated to %q — writes must go to agent_overrides, not frontmatter", e)
	}
}

// TestG3SaveDefaultValueClearsOverride (G3-2/G3-4): submitting the profile-default
// value for an agent clears a pre-existing override (reset-to-named-tier).
func TestG3SaveDefaultValueClearsOverride(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-spec", "opus", "xhigh")
	writeLLMProfileYAML(t, root, "medium", map[string][2]string{"manager-spec": {"sonnet", "high"}})

	a := newPolicyTestApp(root)
	form := baseSaveForm()
	form.Set("performance_tier", "medium")
	// medium/spec_auditors default = opus/high → submitting it clears the override.
	form.Set("agentfm.manager-spec.model", "opus")
	form.Set("agentfm.manager-spec.effort", "high")

	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("save = %d; body %s", rec.Code, rec.Body.String())
	}
	cfg, err := config.NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	if _, ok := cfg.LLM.AgentOverrides["manager-spec"]; ok {
		t.Errorf("override not cleared when the submitted value equals the profile default; overrides=%#v", cfg.LLM.AgentOverrides)
	}
}

// TestG3CustomTierShownWhenOverridesPresent (G3-4): the perf-tier control shows a
// preselected Custom pseudo-state when llm.agent_overrides is non-empty; the named
// tiers are NOT checked.
func TestG3CustomTierShownWhenOverridesPresent(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-spec", "opus", "xhigh")
	writeLLMProfileYAML(t, root, "medium", map[string][2]string{"manager-spec": {"sonnet", "high"}})

	body := renderAgentFMBody(t, root)
	if !strings.Contains(body, `value="custom" checked`) {
		t.Error("Custom tier radio not preselected when agent_overrides is non-empty")
	}
	if strings.Contains(body, `value="medium" checked`) {
		t.Error("a named tier is checked while Custom should be active (overrides present)")
	}
}

// TestG3CustomTierAbsentWhenNoOverrides (G3-4): with no overrides the named tier
// is preselected and Custom is not.
func TestG3CustomTierAbsentWhenNoOverrides(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-spec", "opus", "xhigh")
	writeLLMProfileYAML(t, root, "medium", nil)

	body := renderAgentFMBody(t, root)
	if strings.Contains(body, `value="custom" checked`) {
		t.Error("Custom tier preselected with no overrides present")
	}
	if !strings.Contains(body, `value="medium" checked`) {
		t.Error("named tier (medium) not preselected when no overrides present")
	}
}

// TestG3ProfileMatrixJSONEmitted (G3-3): the per-profile per-agent matrix JSON
// blob is emitted for the client tier-repopulation handler.
func TestG3ProfileMatrixJSONEmitted(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-spec", "opus", "xhigh")
	writeLLMProfileYAML(t, root, "medium", nil)

	body := renderAgentFMBody(t, root)
	if !strings.Contains(body, `id="moai-profile-matrix"`) {
		t.Error("profile-matrix JSON blob not emitted for the client tier-repopulation handler")
	}
	if !strings.Contains(body, `"manager-spec"`) {
		t.Error("profile-matrix JSON blob missing agent cells")
	}
	// The blob carries the three named tiers.
	for _, tier := range []string{`"max"`, `"medium"`, `"low"`} {
		if !strings.Contains(body, tier) {
			t.Errorf("profile-matrix JSON blob missing tier %s", tier)
		}
	}
}

// TestG3AppJSWiresProfileMatrix (G3-3): app.js reads the blob and wires the
// performance_tier radio change.
func TestG3AppJSWiresProfileMatrix(t *testing.T) {
	js := readEmbeddedAsset(t, "app.js")
	if !strings.Contains(js, "moai-profile-matrix") {
		t.Error("app.js does not read the profile-matrix blob")
	}
	if !strings.Contains(js, "performance_tier") {
		t.Error("app.js does not wire the performance_tier radio change")
	}
}
