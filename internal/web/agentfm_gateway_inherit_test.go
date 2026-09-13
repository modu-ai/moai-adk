package web

// t840 (model matrix stage 1) — web half: under a gateway (gpt) backend the
// launcher pins every Claude alias slot (OPUS/SONNET/HAIKU/FABLE) to ONE gpt
// model id (internal/cli/gateway_prepare.go), so per-agent model selection is
// meaningless — sub-agents inherit the session model. The Agents panel must
// therefore render the model cell as a fixed "inherit" display (no editable
// select) and keep ONLY the effort select editable.
//
// The tests pin four properties:
//
//	G1  under a gateway backend (llm.yaml team_mode: gpt) no per-agent model
//	     select renders; each row carries a fixed inherit cell instead, and the
//	     effort select stays editable (never haiku-locked — the session model
//	     is a gpt id, not haiku).
//	G2  the panel states the gateway inheritance in a gateway-gated note
//	     (agentfm.gatewaynote), rendered ONLY under a gateway backend; the GLM
//	     surfaces (glmnote, reasoning chips) stay absent.
//	G3  the new note key exists in all four locale dictionaries (en/ko/ja/zh).
//	G4  save path: a submission carrying only the effort backfills the model
//	     from the resolved value, so the unsubmitted model cell never corrupts
//	     the override (the profile default stays cleared, a changed effort pins
//	     the resolved model).

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/settings/agentfm"
)

// TestAgentFMGatewayInheritCell (G1+G2): team_mode gpt renders the inherit
// cell, drops the model select, keeps the effort select editable.
func TestAgentFMGatewayInheritCell(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-develop", "opus", "medium")
	seedAgentFMFile(t, root, "moai", "Explore", "haiku", "low")
	writeLLMGLMYAML(t, root, "gpt", "medium", "")

	body := renderAgentFMGLMBody(t, root, "")

	for _, agent := range []string{"manager-develop", "Explore"} {
		if strings.Contains(body, `name="agentfm.`+agent+`.model"`) {
			t.Errorf("%s: a model select rendered under a gateway backend — the model axis is session-inherited", agent)
		}
		cell := `data-model-inherit="` + agent + `"`
		if !strings.Contains(body, cell) {
			t.Errorf("%s: rendered body lacks the fixed inherit cell (%s)", agent, cell)
		}
		if !strings.Contains(body, `name="agentfm.`+agent+`.effort"`) {
			t.Errorf("%s: the effort select is missing — effort must stay the editable axis", agent)
		}
	}
	// The Explore row resolves to haiku under a Claude backend, which locks its
	// effort select; under a gateway the session model is a gpt id, so the
	// lock must not apply.
	if strings.Contains(body, `name="agentfm.Explore.effort" id="agentfm.Explore.effort" aria-label="Effort for Explore" disabled`) {
		t.Error("Explore effort select is haiku-locked under a gateway backend — the session model is not haiku")
	}
	if !strings.Contains(body, `data-i18n="agentfm.gatewaynote"`) {
		t.Error("rendered body lacks the gateway-gated agentfm.gatewaynote")
	}
	if strings.Contains(body, `data-i18n="agentfm.glmnote"`) || strings.Contains(body, "data-glm-reasoning=") {
		t.Error("GLM surfaces rendered under a gateway backend — the two gates must be disjoint")
	}
}

// TestAgentFMGatewayInheritFromLaunchProviderEnv pins the production fold
// seam: llm.yaml carries no team_mode of its own and the launcher-owned
// MOAI_LAUNCH_PROVIDER=gpt env folds in (config.WithLaunchProvider), so the
// panel renders the gateway surface. This is the path a `moai gpt` session
// actually exercises — and the signal whose uncontrolled leak into the other
// render tests this file's helper now pins away (t840 env-isolation repair).
func TestAgentFMGatewayInheritFromLaunchProviderEnv(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-develop", "opus", "medium")
	writeLLMGLMYAML(t, root, "", "medium", "")

	body := renderAgentFMGLMBody(t, root, "gpt")

	if strings.Contains(body, `name="agentfm.manager-develop.model"`) {
		t.Error("a model select rendered although the launch provider folded a gateway backend in")
	}
	if !strings.Contains(body, `data-model-inherit="manager-develop"`) {
		t.Error("rendered body lacks the fixed inherit cell under a folded launch provider")
	}
	if !strings.Contains(body, `data-i18n="agentfm.gatewaynote"`) {
		t.Error("rendered body lacks the gateway-gated agentfm.gatewaynote under a folded launch provider")
	}
}

// TestAgentFMGatewayCellHiddenUnderClaudeAndGLM (G1/G2 negative): neither a
// Claude nor a GLM backend renders the inherit cell or the gateway note; the
// model select stays editable there.
func TestAgentFMGatewayCellHiddenUnderClaudeAndGLM(t *testing.T) {
	for _, teamMode := range []string{"", "glm"} {
		t.Run("team_mode="+teamMode, func(t *testing.T) {
			root := t.TempDir()
			seedAgentFMFile(t, root, "moai", "manager-develop", "opus", "medium")
			writeLLMGLMYAML(t, root, teamMode, "medium", "glm-5.3")

			body := renderAgentFMGLMBody(t, root, "")

			if strings.Contains(body, "data-model-inherit=") {
				t.Error("inherit cell rendered outside a gateway backend")
			}
			if strings.Contains(body, `data-i18n="agentfm.gatewaynote"`) {
				t.Error("agentfm.gatewaynote rendered outside a gateway backend")
			}
			if !strings.Contains(body, `name="agentfm.manager-develop.model"`) {
				t.Error("model select missing outside a gateway backend")
			}
		})
	}
}

// TestAgentFMGatewayNoteKeyInFourLocales (G3): agentfm.gatewaynote exists in
// en/ko/ja/zh (same guard shape as TestAgentFMGLMNoteKeyInFourLocales).
func TestAgentFMGatewayNoteKeyInFourLocales(t *testing.T) {
	blocks := localeBlocks(t, readEmbeddedAsset(t, "i18n.js"))
	for _, loc := range []string{"en", "ko", "ja", "zh"} {
		if !strings.Contains(blocks[loc], `"agentfm.gatewaynote":`) {
			t.Errorf("i18n.js locale %q is missing the agentfm.gatewaynote key", loc)
		}
	}
}

// TestAgentFMGatewaySaveBackfillsModel (G4): with no model field submitted
// (the gateway render has none), an effort-only submission backfills the
// model from the resolved value — a changed effort pins {resolved model,
// new effort}; an unchanged effort clears (profile default).
func TestAgentFMGatewaySaveBackfillsModel(t *testing.T) {
	llm := config.LLMConfig{Profile: "medium", TeamMode: config.TeamModeGPT}
	agents := []agentfm.AgentInfo{{Name: "manager-develop", ParseOK: true}}
	resolvedModel := agentResolvedModel(llm, "manager-develop")
	resolvedEffort := agentResolvedEffort(llm, "manager-develop")
	newEffort := "low"
	if resolvedEffort == newEffort {
		newEffort = "high"
	}

	post := func(effort string) *http.Request {
		form := url.Values{"agentfm.manager-develop.effort": {effort}}
		req := httptest.NewRequest(http.MethodPost, "/settings", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return req
	}

	pins, _, errs := parseAgentFMForm(post(newEffort), agents, llm, "")
	if len(errs) != 0 {
		t.Fatalf("unexpected field errors: %v", errs)
	}
	got, ok := pins["manager-develop"]
	if !ok {
		t.Fatalf("effort-only submission did not pin an override; pins=%v", pins)
	}
	if got.Model != resolvedModel || got.Effort != newEffort {
		t.Errorf("pin = %+v, want model backfilled to %q and effort %q", got, resolvedModel, newEffort)
	}

	pins, _, _ = parseAgentFMForm(post(resolvedEffort), agents, llm, "")
	if _, pinned := pins["manager-develop"]; pinned {
		t.Errorf("unchanged effort pinned an override %+v — the backfilled model must compare equal to the profile default", pins["manager-develop"])
	}
}
