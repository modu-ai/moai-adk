package template

import (
	"io/fs"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestGLMTierDefaultsTemplateMatchesGoDefaults pins the distributed llm.yaml
// GLM block to the compiled defaults: every tier model equals the value
// config.NewDefaultLLMConfig() fills an absent key with (glm-5.3-flash for
// high/medium/low, glm-5.3 for fable), and every tier effort is the max
// reasoning state — the same default the web console preselects.
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
	for tier, want := range wantModel {
		if got := doc.LLM.GLM.Models[tier]; got != want {
			t.Errorf("template llm.glm.models.%s = %q, want the Go default %q", tier, got, want)
		}
		if got := doc.LLM.GLM.Effort[tier]; got != GLMStateMax {
			t.Errorf("template llm.glm.effort.%s = %q, want %q", tier, got, GLMStateMax)
		}
	}
	// The Go defaults themselves: flash for the three Claude-named tiers, the
	// full glm-5.3 for the Fable slot.
	for _, tier := range []string{"high", "medium", "low"} {
		if wantModel[tier] != config.DefaultGLM53Flash {
			t.Errorf("Go default for %s = %q, want %q", tier, wantModel[tier], config.DefaultGLM53Flash)
		}
	}
	if wantModel["fable"] != config.DefaultGLM53 {
		t.Errorf("Go default for fable = %q, want %q", wantModel["fable"], config.DefaultGLM53)
	}
}
