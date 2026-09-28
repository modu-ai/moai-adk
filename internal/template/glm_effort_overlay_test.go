package template

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestIsGLMBackend covers the REQ-MTP-026 backend-detection predicate truth
// table (AC-MTP-028). The predicate reads the two llm.yaml intent signals only:
// team_mode ∈ {cg, glm} (the ACTUAL persisted GLM signals) OR mode == "glm"
// (the defensive OR for the currently-dormant llm.mode field).
func TestIsGLMBackend(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		teamMode string
		want     bool
	}{
		// TRUE cases — a GLM backend signal is present.
		{"team_mode=glm (primary moai glm signal)", "", config.TeamModeGLM, true},
		{"team_mode=cg is retired", "", config.LegacyTeamModeCG, false},
		{"mode=glm (defensive dormant-field OR)", config.LLMModeGLM, "", true},
		{"retired cg conflicts with mode=glm", config.LLMModeGLM, config.LegacyTeamModeCG, false},
		// FALSE cases — no GLM signal (legacy non-GLM team_mode values + empty).
		{"team_mode=claude (legacy non-GLM)", "", config.TeamModeClaude, false},
		{"team_mode=hybrid (legacy non-GLM)", "", config.TeamModeHybrid, false},
		{"no signal (both empty)", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.LLMConfig{Mode: tt.mode, TeamMode: tt.teamMode}
			if got := IsGLMBackend(cfg); got != tt.want {
				t.Errorf("IsGLMBackend(mode=%q, team_mode=%q) = %v, want %v",
					tt.mode, tt.teamMode, got, tt.want)
			}
		})
	}
}

// TestCollapseClaudeEffortToGLM covers the REQ-GEM-001 ceiling raise on the
// 5→3 collapse: low→low (thinking enabled); medium/high/xhigh/max→max; and the
// totality clause (unrecognized effort → documented GLM default state, max).
func TestCollapseClaudeEffortToGLM(t *testing.T) {
	tests := []struct {
		effort          string
		wantName        string
		wantThinking    bool
		wantReasoningEf string
	}{
		{EffortLevelLow, GLMStateLow, true, GLMReasoningEffortLow},
		{EffortLevelMedium, GLMStateMax, true, GLMReasoningEffortMax},
		{EffortLevelHigh, GLMStateMax, true, GLMReasoningEffortMax},
		{EffortLevelXHigh, GLMStateMax, true, GLMReasoningEffortMax},
		{EffortLevelMax, GLMStateMax, true, GLMReasoningEffortMax},
		// Totality: an unrecognized effort maps to the GLM default state
		// (reasoning-max = z.ai omit-default), no panic.
		{"bogus-unrecognized", GLMStateMax, true, GLMReasoningEffortMax},
		{"", GLMStateMax, true, GLMReasoningEffortMax},
	}
	for _, tt := range tests {
		t.Run(tt.effort, func(t *testing.T) {
			got := CollapseClaudeEffortToGLM(tt.effort)
			if got.Name != tt.wantName {
				t.Errorf("CollapseClaudeEffortToGLM(%q).Name = %q, want %q", tt.effort, got.Name, tt.wantName)
			}
			if got.ThinkingEnabled != tt.wantThinking {
				t.Errorf("CollapseClaudeEffortToGLM(%q).ThinkingEnabled = %v, want %v", tt.effort, got.ThinkingEnabled, tt.wantThinking)
			}
			if got.ReasoningEffort != tt.wantReasoningEf {
				t.Errorf("CollapseClaudeEffortToGLM(%q).ReasoningEffort = %q, want %q", tt.effort, got.ReasoningEffort, tt.wantReasoningEf)
			}
		})
	}
}

// TestSessionGLMReasoningState confirms the Branch-B session-global delivery
// value is the thinking-enabled max state (REQ-GEM-002, lead-ratified 2026-08-22:
// the session-global env is the only reasoning channel under Branch-B, t127
// measured trivial-spawn cost ≈ 0, and max is z.ai's own omit-default).
func TestSessionGLMReasoningState(t *testing.T) {
	got := SessionGLMReasoningState()
	if got.Name != GLMStateMax {
		t.Errorf("SessionGLMReasoningState().Name = %q, want %q (session default)", got.Name, GLMStateMax)
	}
	if !got.ThinkingEnabled || got.ReasoningEffort != GLMReasoningEffortMax {
		t.Errorf("SessionGLMReasoningState() = %+v, want thinking enabled + reasoning_effort=max", got)
	}
}

// TestSessionGLMReasoningStateForEffort verifies the MAIN-SESSION reasoning
// derivation that is driven by the web-set effort preference. Distinct from
// SessionGLMReasoningState() (the session default used for sub-agents
// and the empty-effort fallback), this helper collapses the user's prefs.EffortLevel
// onto z.ai's 3-state reasoning control so a web-set effort actually reaches z.ai.
func TestSessionGLMReasoningStateForEffort(t *testing.T) {
	cases := []struct {
		name             string
		effort           string
		wantName         string
		wantThinking     bool
		wantReasoningVal string
	}{
		{"low → reasoning low", EffortLevelLow, GLMStateLow, true, GLMReasoningEffortLow},
		{"medium → max", EffortLevelMedium, GLMStateMax, true, GLMReasoningEffortMax},
		{"high → max", EffortLevelHigh, GLMStateMax, true, GLMReasoningEffortMax},
		{"xhigh → max", EffortLevelXHigh, GLMStateMax, true, GLMReasoningEffortMax},
		{"max → max", EffortLevelMax, GLMStateMax, true, GLMReasoningEffortMax},
		{"empty falls back to session default (max)", "", GLMStateMax, true, GLMReasoningEffortMax},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SessionGLMReasoningStateForEffort(tc.effort)
			if got.Name != tc.wantName {
				t.Errorf("SessionGLMReasoningStateForEffort(%q).Name = %q, want %q",
					tc.effort, got.Name, tc.wantName)
			}
			if got.ThinkingEnabled != tc.wantThinking {
				t.Errorf("SessionGLMReasoningStateForEffort(%q).ThinkingEnabled = %v, want %v",
					tc.effort, got.ThinkingEnabled, tc.wantThinking)
			}
			if got.ReasoningEffort != tc.wantReasoningVal {
				t.Errorf("SessionGLMReasoningStateForEffort(%q).ReasoningEffort = %q, want %q",
					tc.effort, got.ReasoningEffort, tc.wantReasoningVal)
			}
		})
	}
}

// TestIsGLMFlashModel covers the flash-model predicate: exact id, decorated
// id, case-insensitive input; and the non-flash negatives (glm-5.3 itself,
// glm-5.1, empty).
func TestIsGLMFlashModel(t *testing.T) {
	tests := []struct {
		model string
		want  bool
	}{
		{config.DefaultGLM53Flash, true},
		{"GLM-5.3-FLASH", true},
		{"glm-5.3-flash[1m]", true},
		{config.DefaultGLM53, false},
		{config.DefaultGLM51, false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsGLMFlashModel(tt.model); got != tt.want {
			t.Errorf("IsGLMFlashModel(%q) = %v, want %v", tt.model, got, tt.want)
		}
	}
}

// TestCollapseClaudeEffortToGLMForModel covers the model-aware collapse:
// under glm-5.3-flash EVERY Claude effort (including low) resolves to the max
// state — flash accepts reasoning_effort: max only, so the low state must not
// be emitted; under any non-flash model the existing collapse is unchanged
// (low→low, above-low→max, unrecognized→max).
func TestCollapseClaudeEffortToGLMForModel(t *testing.T) {
	flash := config.DefaultGLM53Flash
	// Flash: every effort → max (thinking enabled, reasoning_effort=max).
	for _, effort := range []string{
		EffortLevelLow, EffortLevelMedium, EffortLevelHigh,
		EffortLevelXHigh, EffortLevelMax, "bogus-unrecognized", "",
	} {
		got := CollapseClaudeEffortToGLMForModel(flash, effort)
		if got.Name != GLMStateMax || !got.ThinkingEnabled || got.ReasoningEffort != GLMReasoningEffortMax {
			t.Errorf("CollapseClaudeEffortToGLMForModel(flash, %q) = %+v, want the max state", effort, got)
		}
	}
	// Mirror-image regression: non-flash collapse EXACTLY unchanged.
	for _, tc := range []struct {
		model, effort, wantName, wantEffort string
	}{
		{config.DefaultGLM53, EffortLevelLow, GLMStateLow, GLMReasoningEffortLow},
		{config.DefaultGLM53, EffortLevelMedium, GLMStateMax, GLMReasoningEffortMax},
		{config.DefaultGLM53, "bogus", GLMStateMax, GLMReasoningEffortMax},
		{config.DefaultGLM51, EffortLevelLow, GLMStateLow, GLMReasoningEffortLow},
		{"", EffortLevelLow, GLMStateLow, GLMReasoningEffortLow},
	} {
		got := CollapseClaudeEffortToGLMForModel(tc.model, tc.effort)
		if got.Name != tc.wantName || got.ReasoningEffort != tc.wantEffort {
			t.Errorf("CollapseClaudeEffortToGLMForModel(%q, %q) = %+v, want %s/%s",
				tc.model, tc.effort, got, tc.wantName, tc.wantEffort)
		}
	}
}

// TestSessionGLMReasoningStateForModel covers the model-aware main-session
// derivation: under flash a web-set low effort still pins max and the empty
// fallback stays max; under non-flash the prefs-driven collapse is unchanged.
func TestSessionGLMReasoningStateForModel(t *testing.T) {
	flash := config.DefaultGLM53Flash
	if got := SessionGLMReasoningStateForModel(flash, EffortLevelLow); got.Name != GLMStateMax || got.ReasoningEffort != GLMReasoningEffortMax {
		t.Errorf("SessionGLMReasoningStateForModel(flash, low) = %+v, want max", got)
	}
	if got := SessionGLMReasoningStateForModel(flash, ""); got.Name != GLMStateMax {
		t.Errorf("SessionGLMReasoningStateForModel(flash, \"\") = %+v, want max", got)
	}
	if got := SessionGLMReasoningStateForModel(config.DefaultGLM53, EffortLevelLow); got.Name != GLMStateLow || got.ReasoningEffort != GLMReasoningEffortLow {
		t.Errorf("SessionGLMReasoningStateForModel(glm-5.3, low) = %+v, want low", got)
	}
	if got := SessionGLMReasoningStateForModel(config.DefaultGLM53, ""); got.Name != GLMStateMax {
		t.Errorf("SessionGLMReasoningStateForModel(glm-5.3, \"\") = %+v, want max", got)
	}
}
