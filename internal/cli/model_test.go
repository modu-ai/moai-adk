package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
)

// TestResolveModelProfileReport_MaxClaude covers REQ-MPM-025 / AC-MPM-016: the
// resolver emits the top (high) column under a Claude backend, with the model as
// the per-spawn arg and effort as documented intent. The seeded "max" exercises
// the superseded top-column alias.
func TestResolveModelProfileReport_MaxClaude(t *testing.T) {
	llm := config.LLMConfig{Profile: "max"}
	rpt := resolveModelProfileReport(llm)
	if rpt.Profile != "high" || rpt.Backend != "claude" {
		t.Fatalf("expected profile=high (alias of max) backend=claude, got %s/%s", rpt.Profile, rpt.Backend)
	}
	if rpt.WireNote != "" {
		t.Errorf("Claude backend should carry no GLM wire note")
	}
	got := map[string]modelProfileEntry{}
	for _, e := range rpt.Agents {
		got[e.Agent] = e
	}
	if e := got["manager-develop"]; e.Model != "opus" || e.Effort != "max" {
		t.Errorf("manager-develop high: got %s/%s, want opus/max", e.Model, e.Effort)
	}
	if e := got["Explore"]; e.Model != "sonnet" || e.Effort != "low" || e.Group != "explore" {
		t.Errorf("Explore: got %s/%s group=%s, want sonnet/low group=explore", e.Model, e.Effort, e.Group)
	}
	if len(rpt.Agents) != 11 {
		t.Errorf("expected 11 agents, got %d", len(rpt.Agents))
	}
}

// TestResolveModelProfileReport_GLMOverlay covers REQ-MPM-029/030/031 /
// AC-MPM-019 plus the t66 GLM session-model inheritance fold: under GLM the
// per-agent glm_model column carries the explicit "inherit" sentinel (the
// sub-agent model is session-inherited via llm.glm.models →
// ANTHROPIC_DEFAULT_*_MODEL, so per-agent model cells carry no differentiation),
// while manager-develop's effort still collapses to the max level (coding-max
// override), with a wire-honesty note.
func TestResolveModelProfileReport_GLMOverlay(t *testing.T) {
	llm := config.LLMConfig{
		Profile:  "max",
		TeamMode: "glm",
		GLM: config.GLMSettings{Models: config.GLMModels{
			High: "glm-5.2", Medium: "glm-4.7", Low: "glm-4.5-air", Fable: "glm-5.2",
			Opus: "glm-5.2", Sonnet: "glm-4.7", Haiku: "glm-4.5-air",
		}},
	}
	rpt := resolveModelProfileReport(llm)
	if rpt.Backend != "glm" {
		t.Fatalf("expected glm backend, got %s", rpt.Backend)
	}
	if rpt.WireNote == "" {
		t.Errorf("GLM backend must carry the wire-honesty note (AC-MPM-023)")
	}
	got := map[string]modelProfileEntry{}
	for _, e := range rpt.Agents {
		got[e.Agent] = e
	}
	// Every mapped agent's glm_model column states the explicit inherit
	// sentinel — the GLM model is session-carried, never per-agent routed
	// (measured 2026-08-16: all agents glm_model identical; only reasoning
	// differs). The tier→model map above deliberately DIFFERS per tier, so a
	// pass here proves the fold is unconditional, not an artifact of equal
	// tier values.
	for _, e := range rpt.Agents {
		if e.GLMModel != "inherit" {
			t.Errorf("%s glm model: got %q, want the explicit inherit sentinel", e.Agent, e.GLMModel)
		}
	}
	// manager-develop = opus/max → coding-max override → the max level.
	md := got["manager-develop"]
	if md.GLMReasoning != template.GLMStateMax {
		t.Errorf("manager-develop glm reasoning: got %s, want %s (coding-max override)", md.GLMReasoning, template.GLMStateMax)
	}
	// manager-git = sonnet/low → low collapses to the low level.
	mg := got["manager-git"]
	if mg.GLMReasoning != template.GLMStateLow {
		t.Errorf("manager-git glm reasoning: got %s, want %s", mg.GLMReasoning, template.GLMStateLow)
	}
	// Explore = sonnet/low → low collapses to the low level (now that
	// Explore has an explicit group it IS GLM-injected).
	if e := got["Explore"]; e.GLMReasoning != template.GLMStateLow {
		t.Errorf("Explore glm reasoning: got %s, want %s", e.GLMReasoning, template.GLMStateLow)
	}
}

// TestResolveModelProfileReport_GLMModelExplicitInJSON pins the explicit
// representation requirement of the t66 fold: the glm_model field survives JSON
// serialization with the non-empty inherit sentinel (never a silent absence via
// omitempty), so a JSON consumer can distinguish "states inheritance" from
// "column dropped".
func TestResolveModelProfileReport_GLMModelExplicitInJSON(t *testing.T) {
	llm := config.LLMConfig{Profile: "medium", TeamMode: "glm"}
	rpt := resolveModelProfileReport(llm)
	for _, e := range rpt.Agents {
		b, err := json.Marshal(e)
		if err != nil {
			t.Fatalf("marshal entry: %v", err)
		}
		if !strings.Contains(string(b), `"glm_model":"inherit"`) {
			t.Errorf("%s: JSON %s lacks the explicit glm_model inherit field", e.Agent, b)
		}
	}
}
