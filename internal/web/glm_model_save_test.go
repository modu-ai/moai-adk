package web

// glm_model_save_test.go — card t1461 reproduction: selecting a GLM tier model
// on the 3rd Party LLM tab and pressing Save must leave that model selected.
// The submission is browser-faithful (rendered page → native form semantics →
// operator edit → POST /save → fresh GET), the same shape as
// web_save_fullform_repro_test.go.

import (
	"net/http"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// llmGLMSaveFixture mirrors the distributed llm.yaml shape for the GLM block:
// comments around the models map and no effort keys on disk.
const llmGLMSaveFixture = `llm:
  mode: ""
  team_mode: ""
  harness: "claude"
  glm_env_var: "GLM_API_KEY"
  profile: ""
  agent_overrides: {}

  # GLM backend configuration
  glm:
    base_url: "https://api.z.ai/api/anthropic"
    models:
      high: "glm-5.3-flash"   # 1M context — Opus slot
      medium: "glm-5.3-flash" # 1M context — Sonnet slot
      low: "glm-5.3-flash"    # 1M context — Haiku slot
      fable: "glm-5.3"        # 1M context — Fable tier
`

// llmNoGLMModelsFixture is an llm.yaml that carries no llm.glm.models block —
// a project initialized before the block existed, or one whose user trimmed
// it. The runtime resolves every tier from the compiled defaults.
const llmNoGLMModelsFixture = `llm:
  mode: ""
  team_mode: glm
  harness: "claude"
  glm_env_var: "GLM_API_KEY"
`

// TestGLMModelSelectionPersistsOnSave pins the stored-value path: switching
// one tier to the other model writes it and leaves the neighbours alone.
func TestGLMModelSelectionPersistsOnSave(t *testing.T) {
	root := t.TempDir()
	seedSectionFile(t, root, "llm", llmGLMSaveFixture)
	a := newAppWithRoot(t, root)

	form := extractBrowserSubmission(t, renderSettingsGET(a))
	if got := form.Get("llm.glm.models.high"); got != config.DefaultGLM53Flash {
		t.Fatalf("rendered form submits llm.glm.models.high=%q before the edit, want %q (stored value not preselected)", got, config.DefaultGLM53Flash)
	}
	form.Set("llm.glm.models.high", config.DefaultGLM53)

	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("save status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}

	got := readSeededSectionFile(t, root, "llm")
	if !strings.Contains(got, `high: "glm-5.3"`) && !strings.Contains(got, "high: glm-5.3\n") {
		t.Errorf("llm.glm.models.high selection not persisted; file:\n%s", got)
	}
	if !strings.Contains(got, `medium: "glm-5.3-flash"`) && !strings.Contains(got, "medium: glm-5.3-flash") {
		t.Errorf("llm.glm.models.medium changed by an unrelated edit; file:\n%s", got)
	}
}

// TestGLMModelSelectionSurvivesSaveWhenKeyAbsent is the card t1461 RED: with
// no llm.glm.models block on disk, every tier must render its effective model
// preselected, and a selection made and saved must still be the selected one
// on the next page load — for every tier and both models.
func TestGLMModelSelectionSurvivesSaveWhenKeyAbsent(t *testing.T) {
	tierDefault := map[string]string{
		"high":   config.DefaultGLMHigh,
		"medium": config.DefaultGLMMedium,
		"low":    config.DefaultGLMLow,
		"fable":  config.DefaultGLMFable,
	}
	for _, tier := range glmTierKeys {
		for _, model := range config.ValidGLMModels() {
			t.Run(tier+"/"+model, func(t *testing.T) {
				name := "llm.glm.models." + tier
				root := t.TempDir()
				seedSectionFile(t, root, "llm", llmNoGLMModelsFixture)
				a := newAppWithRoot(t, root)

				form := extractBrowserSubmission(t, renderSettingsGET(a))
				if got := form.Get(name); got != tierDefault[tier] {
					t.Errorf("%s: absent key renders preselected %q, want the effective default %q", name, got, tierDefault[tier])
				}
				form.Set(name, model)

				rec := servePost(t, a.routes(), "/save", form)
				if rec.Code != http.StatusOK {
					t.Fatalf("save status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
				}

				after := extractBrowserSubmission(t, renderSettingsGET(newAppWithRoot(t, root)))
				if got := after.Get(name); got != model {
					t.Errorf("%s: selected %q and saved, but the reloaded page selects %q; file:\n%s",
						name, model, got, readSeededSectionFile(t, root, "llm"))
				}
			})
		}
	}
}
