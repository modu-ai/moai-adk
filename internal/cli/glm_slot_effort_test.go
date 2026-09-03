package cli

// glm_slot_effort_test.go — RC3 regression guards for wiring the per-slot
// llm.glm.effort.* preference into the GLM session launch (glm-settings-persist).
//
// The four glm.effort keys were write-only: settings.ApplySchemaEdits stored
// them, no runtime path read them, and the console labeled them stored-only.
// Now the launcher resolves the slot serving the MAIN session model and lets a
// non-empty stored glm.effort[slot] override the prefs/model_policy effort
// chain, ahead of the existing collapse overlay (which stays governing for the
// final wire value: stored high and max both reach z.ai as max; flash pins
// every effort to max).

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/template"
)

// tierEffortAll builds a GLMTierEffort with every slot set.
func tierEffortAll(high, medium, low, fable string) config.GLMTierEffort {
	return config.GLMTierEffort{High: high, Medium: medium, Low: low, Fable: fable}
}

// TestGLMSlotEffortForModel pins the alias→slot resolution: the pairing is the
// SAME mapping setGLMEnv uses to assign ANTHROPIC_DEFAULT_*_MODEL (opus feeds
// Models.High → effort.high, sonnet → Medium, haiku → Low, fable → Fable), with
// the [1m] suffix split first and canonical claude-* ids reverse-mapped through
// template.ModelAliasFromCanonicalID. Unknown or empty models resolve "" — the
// caller falls back to the prefs chain unchanged.
func TestGLMSlotEffortForModel(t *testing.T) {
	effort := tierEffortAll("e-high", "e-medium", "e-low", "e-fable")
	cases := []struct {
		model string
		want  string
	}{
		{"opus", "e-high"},
		{"sonnet", "e-medium"},
		{"haiku", "e-low"},
		{"fable", "e-fable"},
		{"opus[1m]", "e-high"},                           // 1M suffix split before lookup
		{"sonnet[1m]", "e-medium"},                       //
		{template.ModelIDOpus5, "e-high"},                // canonical id reverse-mapped
		{"claude-sonnet-5", "e-medium"},                  //
		{"", ""},                                         // no model pinned → no slot claim
		{config.DefaultGLM53, ""},                        // raw GLM id is not an alias → ""
		{template.ModelAliasCanonicalID("opusplan"), ""}, // routing alias owns no tier slot
	}
	for _, tc := range cases {
		if got := glmSlotEffortForModel(tc.model, effort); got != tc.want {
			t.Errorf("glmSlotEffortForModel(%q) = %q, want %q", tc.model, got, tc.want)
		}
	}
}

// TestLoadGLMConfig_CarriesTierEffort pins that loadGLMConfig no longer drops
// the persisted per-tier effort on the floor (the structural drop was RC3's
// root cause): the disk-loaded section's glm.effort map survives into the
// launcher-facing GLMConfigFromYAML.
func TestLoadGLMConfig_CarriesTierEffort(t *testing.T) {
	root := t.TempDir()
	sectionsDir := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	src := "llm:\n  glm:\n    effort:\n      high: " + template.GLMStateLow + "\n      fable: " + template.GLMStateMax + "\n"
	if err := os.WriteFile(filepath.Join(sectionsDir, "llm.yaml"), []byte(src), 0o644); err != nil {
		t.Fatalf("write llm.yaml: %v", err)
	}

	origDeps := deps
	deps = nil // force the disk path — the live launcher path
	defer func() { deps = origDeps }()

	cfg, err := loadGLMConfig(root)
	if err != nil {
		t.Fatalf("loadGLMConfig: %v", err)
	}
	if cfg.Effort.High != template.GLMStateLow {
		t.Errorf("Effort.High = %q, want %q — the stored per-tier effort was dropped", cfg.Effort.High, template.GLMStateLow)
	}
	if cfg.Effort.Fable != template.GLMStateMax {
		t.Errorf("Effort.Fable = %q, want %q", cfg.Effort.Fable, template.GLMStateMax)
	}
	if cfg.Effort.Medium != "" || cfg.Effort.Low != "" {
		t.Errorf("unset slots must stay empty, got %+v", cfg.Effort)
	}
}

// TestResolveGLMMainSessionEffort pins the launch precedence: a non-empty
// stored glm.effort[slot] wins over the prefs/model_policy chain; an empty
// stored value (or a model with no slot claim) falls back to it unchanged —
// byte-identical to the pre-RC3 launch behavior.
func TestResolveGLMMainSessionEffort(t *testing.T) {
	stored := config.GLMTierEffort{High: template.GLMStateLow}

	t.Run("non-empty stored slot overrides the prefs chain", func(t *testing.T) {
		if got := resolveGLMMainSessionEffort("opus", stored, "xhigh"); got != template.GLMStateLow {
			t.Errorf("resolveGLMMainSessionEffort(opus, high=low, fallback=xhigh) = %q, want the stored %q", got, template.GLMStateLow)
		}
	})
	t.Run("empty stored slot keeps the fallback unchanged", func(t *testing.T) {
		if got := resolveGLMMainSessionEffort("sonnet", stored, "medium"); got != "medium" {
			t.Errorf("resolveGLMMainSessionEffort(sonnet, high=low, fallback=medium) = %q, want %q", got, "medium")
		}
	})
	t.Run("model with no slot claim keeps the fallback unchanged", func(t *testing.T) {
		if got := resolveGLMMainSessionEffort("", stored, "high"); got != "high" {
			t.Errorf("resolveGLMMainSessionEffort(\"\", ...) = %q, want %q", got, "high")
		}
		if got := resolveGLMMainSessionEffort(config.DefaultGLM53, stored, "high"); got != "high" {
			t.Errorf("resolveGLMMainSessionEffort(glm-5.3, ...) = %q, want %q", got, "high")
		}
	})
	t.Run("empty everything stays empty", func(t *testing.T) {
		if got := resolveGLMMainSessionEffort("opus", config.GLMTierEffort{}, ""); got != "" {
			t.Errorf("resolveGLMMainSessionEffort with no stored and no fallback = %q, want \"\"", got)
		}
	})
}

// TestBuildEnvForGLMLaunch_OverlayGovernsStoredEffort documents that the
// downstream collapse overlay stays governing for the final wire value: a
// stored `low` reaches z.ai as reasoning_effort=low under glm-5.3, but under
// glm-5.3-flash EVERY stored effort (including low) pins to max — flash accepts
// reasoning_effort: max only. Stored high and max both collapse to max under
// any model (SPEC-GLM-EFFORT-MAX-001).
func TestBuildEnvForGLMLaunch_OverlayGovernsStoredEffort(t *testing.T) {
	base := []string{"PATH=/usr/bin"}

	t.Run("non-flash model honors stored low", func(t *testing.T) {
		env := buildEnvForGLMLaunch(config.GLMModels{High: config.DefaultGLM53}, "opus", template.GLMStateLow, base)
		found := ""
		for _, e := range env {
			if len(e) > len(config.EnvAnthropicReasoningEffort) && e[:len(config.EnvAnthropicReasoningEffort)] == config.EnvAnthropicReasoningEffort {
				found = e[len(config.EnvAnthropicReasoningEffort)+1:]
			}
		}
		if found != template.GLMReasoningEffortLow {
			t.Errorf("%s = %q, want %q under %s", config.EnvAnthropicReasoningEffort, found, template.GLMReasoningEffortLow, config.DefaultGLM53)
		}
	})
	t.Run("flash model pins every stored effort to max", func(t *testing.T) {
		for _, effort := range []string{template.GLMStateLow, template.GLMStateHigh, template.GLMStateMax} {
			env := buildEnvForGLMLaunch(config.GLMModels{High: config.DefaultGLM53Flash}, "opus", effort, base)
			found := ""
			for _, e := range env {
				if len(e) > len(config.EnvAnthropicReasoningEffort) && e[:len(config.EnvAnthropicReasoningEffort)] == config.EnvAnthropicReasoningEffort {
					found = e[len(config.EnvAnthropicReasoningEffort)+1:]
				}
			}
			if found != template.GLMReasoningEffortMax {
				t.Errorf("flash + stored %q: %s = %q, want %q (flash accepts max only)", effort, config.EnvAnthropicReasoningEffort, found, template.GLMReasoningEffortMax)
			}
		}
	})
}

// TestResolveGLMBackendForLaunch_ReturnsTierEffort pins the launcher-side
// resolver hands the persisted per-tier effort to launchClaude alongside the
// backend flag and high-slot model.
func TestResolveGLMBackendForLaunch_ReturnsTierEffort(t *testing.T) {
	root := t.TempDir()
	sectionsDir := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	src := "llm:\n  team_mode: glm\n  glm:\n    models:\n      high: " + config.DefaultGLM53 + "\n    effort:\n      medium: " + template.GLMStateLow + "\n"
	if err := os.WriteFile(filepath.Join(sectionsDir, "llm.yaml"), []byte(src), 0o644); err != nil {
		t.Fatalf("write llm.yaml: %v", err)
	}

	glmBackend, models, effort := resolveGLMBackendForLaunch(root)
	if !glmBackend {
		t.Fatalf("resolveGLMBackendForLaunch = false, want true (team_mode: glm)")
	}
	if models.High != config.DefaultGLM53 {
		t.Errorf("high model = %q, want %q", models.High, config.DefaultGLM53)
	}
	if effort.Medium != template.GLMStateLow {
		t.Errorf("effort.Medium = %q, want %q — the stored tier effort did not reach the launcher", effort.Medium, template.GLMStateLow)
	}
}
