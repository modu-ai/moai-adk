package template

import (
	"io/fs"
	"slices"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestGLMTierDefaultsTemplateMatchesGoDefaults pins the distributed llm.yaml
// GLM block to the compiled defaults: every tier model equals the value
// config.NewDefaultLLMConfig() fills an absent key with (glm-5.3 for
// high/medium/fable, glm-5.3-flash for low), and every tier effort equals the
// reasoning-state default config.DefaultGLMEffort* names (high max, medium
// high, low max, fable high since card t1866) — the same default the web
// console preselects.
func TestGLMTierDefaultsTemplateMatchesGoDefaults(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}
	data, err := fs.ReadFile(fsys, ".moai/config/sections/llm.yaml")
	if err != nil {
		t.Fatalf("read embedded llm.yaml: %v", err)
	}
	var doc struct {
		LLM struct {
			GLM struct {
				Models map[string]string `yaml:"models"`
				Effort map[string]string `yaml:"effort"`
			} `yaml:"glm"`
		} `yaml:"llm"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse embedded llm.yaml: %v", err)
	}

	def := config.NewDefaultLLMConfig().GLM.Models
	wantModel := map[string]string{
		"high":   def.High,
		"medium": def.Medium,
		"low":    def.Low,
		"fable":  def.Fable,
	}
	wantEffort := map[string]string{
		"high":   config.DefaultGLMEffortHigh,
		"medium": config.DefaultGLMEffortMedium,
		"low":    config.DefaultGLMEffortLow,
		"fable":  config.DefaultGLMEffortFable,
	}
	for tier, want := range wantModel {
		if got := doc.LLM.GLM.Models[tier]; got != want {
			t.Errorf("template llm.glm.models.%s = %q, want the Go default %q", tier, got, want)
		}
		if got := doc.LLM.GLM.Effort[tier]; got != wantEffort[tier] {
			t.Errorf("template llm.glm.effort.%s = %q, want the Go default %q", tier, got, wantEffort[tier])
		}
		if !slices.Contains(GLMReasoningStateNames(), wantEffort[tier]) {
			t.Errorf("Go default effort for %s = %q is not a GLM reasoning state %v", tier, wantEffort[tier], GLMReasoningStateNames())
		}
	}
	// The Go defaults themselves (card t1808): the full glm-5.3 for the
	// Opus/Sonnet/Fable slots, glm-5.3-flash for the Haiku slot.
	for _, tier := range []string{"high", "medium", "fable"} {
		if wantModel[tier] != config.DefaultGLM53 {
			t.Errorf("Go default for %s = %q, want %q", tier, wantModel[tier], config.DefaultGLM53)
		}
	}
	if wantModel["low"] != config.DefaultGLM53Flash {
		t.Errorf("Go default for low = %q, want %q", wantModel["low"], config.DefaultGLM53Flash)
	}
}
