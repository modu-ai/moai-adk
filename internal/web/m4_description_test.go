package web

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// Tests for SPEC-WEBCONF-SIMPLIFY-001 M4: per-option description rendering
// mechanism (REQ-WC-015, design.md §H.1 option (a)) + simplified surviving-tab
// surfaces (REQ-WC-016 git_strategy; llm GLM tier mapping) + quality_extras
// enable/disable toggle on the launch tab (REQ-WC-004 / AC-WC-004).

// TestM4DescriptionElementRenders verifies the REQ-WC-015 description mechanism:
// a FieldDef with a non-empty Description renders a .field-description element
// carrying the fieldDesc.<sectionID>.<fieldID> i18n key (design.md §H.1/§H.3).
// git_strategy.mode has Description="fieldDesc.git_strategy.mode" set in M4.
func TestM4DescriptionElementRenders(t *testing.T) {
	body := renderConsolePage(t)

	// The .field-description element renders for git_strategy.mode (Description set).
	if !strings.Contains(body, `class="field-description"`) {
		t.Error(`rendered page missing .field-description element (FieldDef.Description mechanism — REQ-WC-015)`)
	}
	if !strings.Contains(body, `data-i18n="fieldDesc.git_strategy.mode"`) {
		t.Error(`rendered page missing fieldDesc.git_strategy.mode i18n key on the description element`)
	}
	// Per-option title attribute (data-i18n-title) on git_strategy.mode options.
	if !strings.Contains(body, `data-i18n-title="fieldDesc.git_strategy.mode.option.manual"`) {
		t.Error(`rendered page missing per-option data-i18n-title for git_strategy.mode.manual (REQ-WC-015 per-option)`)
	}
}

// TestM4GitStrategySurface verifies REQ-WC-016 / AC-WC-024: the git_strategy tab
// surfaces EXACTLY mode + merge_method (×3 profiles) + hooks.pre_push (×3
// profiles). No per-provider nesting (branch_creation/automation/commit_style/etc.)
// is exposed.
func TestM4GitStrategySurface(t *testing.T) {
	names := map[string]bool{}
	for _, f := range settings.SectionFields(settings.SectionGitStrategy) {
		names[f.Name] = true
	}
	// The 4 core fields: mode + {manual,personal,team}.merge_method (radio widgets, pre_push removed).
	want := []string{
		"git_strategy.mode",
	}
	for _, w := range want {
		if !names[w] {
			t.Errorf("git_strategy surface missing field %q (REQ-WC-016)", w)
		}
	}
	// Per-provider nesting MUST NOT be exposed (baked/hidden, M2).
	for _, hidden := range []string{
		"git_strategy.manual.branch_creation",
		"git_strategy.team.draft_dr",
		"git_strategy.personal.commit_style",
		"git_strategy.team.required_reviews",
	} {
		if names[hidden] {
			t.Errorf("git_strategy surface exposes per-provider nesting field %q (must stay UI-hidden — REQ-WC-016)", hidden)
		}
	}
}

// TestM4LLMSurface verifies the llm tab surfaces ONLY the GLM tier mapping
// (glm.models.{high,medium,low,fable}). mode/team_mode are read-only display
// (not editable); performance_tier/plan_type/claude_models are baked/hidden.
func TestM4LLMSurface(t *testing.T) {
	names := map[string]bool{}
	for _, f := range settings.SectionFields(settings.SectionLLM) {
		names[f.Name] = true
	}
	for _, w := range []string{
		"llm.glm.models.high", "llm.glm.models.medium",
		"llm.glm.models.low", "llm.glm.models.fable",
	} {
		if !names[w] {
			t.Errorf("llm surface missing GLM tier field %q", w)
		}
	}
	for _, hidden := range []string{
		"llm.mode", "llm.team_mode", // read-only display (not editable fields)
		"llm.performance_tier", "llm.plan_type", // baked/hidden
	} {
		if names[hidden] {
			t.Errorf("llm surface exposes non-tier field %q (must stay hidden — glm tier mapping only)", hidden)
		}
	}
}

// TestM4QualityExtrasToggleOnLaunch verifies REQ-WC-004 / AC-WC-004: a single
// enable/disable toggle for the quality-extras feature is rendered on the launch
// tab (OQ-1 resolution). The detailed DDD-gate fields stay baked/hidden.
func TestM4QualityExtrasToggleOnLaunch(t *testing.T) {
	body := renderConsolePage(t)

	// The toggle control renders on the launch tab.
	if !strings.Contains(body, `name="quality.quality_extras_enabled"`) {
		t.Error(`launch tab missing the quality_extras enable/disable toggle (name="quality.quality_extras_enabled" — REQ-WC-004 / AC-WC-004)`)
	}
	// The toggle persists via the hidden __present companion (bool disambiguation).
	if !strings.Contains(body, `name="quality.quality_extras_enabled__present"`) {
		t.Error(`quality_extras toggle missing the __present hidden companion (bool save disambiguation)`)
	}
}
