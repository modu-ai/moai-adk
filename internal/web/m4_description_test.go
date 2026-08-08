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

// TestGitStrategySectionIDNotATab verifies the git_strategy SECTION id is not
// itself a tab or panel id. SPEC-WEB-CONSOLE-REDESIGN-001 M1 restored the
// git_strategy render surface (F1 — the FieldDefs existed with no UI), but it
// lands on the `git-worktree` panel, which mixes git_strategy and
// workflow.worktree fields and therefore owns its own tab.* i18n namespace.
// The controls themselves are asserted present by TestGitStrategyRendered.
func TestGitStrategySectionIDNotATab(t *testing.T) {
	body := renderConsolePage(t)

	for _, marker := range []string{
		`data-tab="git_strategy"`,
		`data-panel="git_strategy"`,
		`data-i18n="sec.git_strategy.title"`,
	} {
		if strings.Contains(body, marker) {
			t.Errorf("rendered page uses the section id %q as a tab/panel id — the panel id is git-worktree", marker)
		}
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

// TestM4QualityExtrasToggleOnLaunch verifies the launch tab no longer renders
// the quality_extras_enabled toggle: the control was removed from the UI and the
// field is now forced to true unconditionally in the persistence layer
// (sectionapply.go applyTypedEdits). Both the toggle input and its hidden
// __present companion MUST be absent from the rendered HTML.
func TestM4QualityExtrasToggleOnLaunch(t *testing.T) {
	body := renderConsolePage(t)

	// The toggle control is no longer rendered on the launch tab.
	if strings.Contains(body, `name="quality.quality_extras_enabled"`) {
		t.Error(`launch tab still renders the removed quality_extras toggle (name="quality.quality_extras_enabled" — should be absent after UI removal)`)
	}
	// The hidden __present companion is also gone (the toggle was the only submitter).
	if strings.Contains(body, `name="quality.quality_extras_enabled__present"`) {
		t.Error(`quality_extras __present hidden companion still rendered (should be absent after toggle removal)`)
	}
}
