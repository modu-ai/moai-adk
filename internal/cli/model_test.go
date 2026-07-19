package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestResolveModelProfileReport_MaxClaude covers REQ-MPM-025 / AC-MPM-016: the
// resolver emits the Matrix A max column under a Claude backend, with the model
// as the per-spawn arg and effort as documented intent.
func TestResolveModelProfileReport_MaxClaude(t *testing.T) {
	llm := config.LLMConfig{Profile: "max"}
	rpt := resolveModelProfileReport(llm)
	if rpt.Profile != "max" || rpt.Backend != "claude" {
		t.Fatalf("expected profile=max backend=claude, got %s/%s", rpt.Profile, rpt.Backend)
	}
	if rpt.WireNote != "" {
		t.Errorf("Claude backend should carry no GLM wire note")
	}
	got := map[string]modelProfileEntry{}
	for _, e := range rpt.Agents {
		got[e.Agent] = e
	}
	if e := got["manager-develop"]; e.Model != "fable" || e.Effort != "low" {
		t.Errorf("manager-develop max: got %s/%s, want fable/low", e.Model, e.Effort)
	}
	if e := got["Explore"]; e.Model != "inherit" || e.Group != "-" {
		t.Errorf("Explore: got %s group=%s, want inherit group=-", e.Model, e.Group)
	}
	if len(rpt.Agents) != 11 {
		t.Errorf("expected 11 agents, got %d", len(rpt.Agents))
	}
}

// TestResolveModelProfileReport_GLMOverlay covers REQ-MPM-029/030/031 /
// AC-MPM-019: under GLM, fable maps to glm-5.2 and manager-develop's effort
// collapses to reasoning-max (coding-max override), with a wire-honesty note.
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
	// manager-develop max = fable/low → glm-5.2 + coding-max override → reasoning-max.
	md := got["manager-develop"]
	if md.GLMModel != "glm-5.2" {
		t.Errorf("manager-develop glm model: got %s, want glm-5.2", md.GLMModel)
	}
	if md.GLMReasoning != "reasoning-max" {
		t.Errorf("manager-develop glm reasoning: got %s, want reasoning-max (coding-max override)", md.GLMReasoning)
	}
	// manager-git = sonnet/low → glm-4.7 + low collapses to thinking-off.
	mg := got["manager-git"]
	if mg.GLMModel != "glm-4.7" || mg.GLMReasoning != "thinking-off" {
		t.Errorf("manager-git glm: got %s/%s, want glm-4.7/thinking-off", mg.GLMModel, mg.GLMReasoning)
	}
	// Explore stays inherit — no GLM injection.
	if e := got["Explore"]; e.GLMModel != "" {
		t.Errorf("Explore should not be GLM-injected, got %s", e.GLMModel)
	}
}
