package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
)

// TestResolveModelProfileReport_GatewayInherit (t840): under a gateway (gpt)
// backend the launcher pins every alias slot to one gpt model id, so the
// report states backend=gpt and folds each mapped agent's gateway_model to
// the explicit inherit sentinel (the same shape convention as the GLM
// columns); effort stays the only per-agent axis, and no GLM field leaks.
func TestResolveModelProfileReport_GatewayInherit(t *testing.T) {
	llm := config.LLMConfig{Profile: "medium", TeamMode: config.TeamModeGPT}
	rpt := resolveModelProfileReport(llm)
	if rpt.Backend != "gpt" {
		t.Fatalf("expected gpt backend, got %s", rpt.Backend)
	}
	if rpt.WireNote == "" {
		t.Error("gateway backend should state the session-inheritance note")
	}
	for _, e := range rpt.Agents {
		if e.Group == "-" {
			continue
		}
		if e.GatewayModel != template.ModelInherit {
			t.Errorf("%s gateway model: got %q, want the explicit inherit sentinel", e.Agent, e.GatewayModel)
		}
		if e.GLMModel != "" || e.GLMReasoning != "" {
			t.Errorf("%s carries GLM fields under a gateway backend: %+v", e.Agent, e)
		}
		if e.Effort == "" {
			t.Errorf("%s effort empty under a gateway backend — effort is the per-agent axis", e.Agent)
		}
		b, err := json.Marshal(e)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(b), `"gateway_model":"inherit"`) {
			t.Errorf("%s: JSON %s lacks the explicit gateway_model inherit field", e.Agent, b)
		}
	}
}

// TestResolveModelProfileReport_ClaudeHasNoGatewayField: the gateway column
// is populated only under a gateway backend (omitempty elsewhere).
func TestResolveModelProfileReport_ClaudeHasNoGatewayField(t *testing.T) {
	rpt := resolveModelProfileReport(config.LLMConfig{Profile: "medium"})
	for _, e := range rpt.Agents {
		if e.GatewayModel != "" {
			t.Errorf("%s: gateway_model %q populated under a Claude backend", e.Agent, e.GatewayModel)
		}
	}
}
