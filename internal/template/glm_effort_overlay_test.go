package template

import (
	"sort"
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
		{"team_mode=cg (moai cg signal)", "", config.TeamModeCG, true},
		{"mode=glm (defensive dormant-field OR)", config.LLMModeGLM, "", true},
		{"mode=glm AND team_mode=cg", config.LLMModeGLM, config.TeamModeCG, true},
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

// TestResolveGLMReasoning_CodingMaxOverride covers the REQ-MTP-028 coding-max
// override (AC-MTP-030): manager-develop resolves to reasoning-max REGARDLESS
// of the input effort; builder-harness (removed from the override set by
// SPEC-GLM-EFFORT-TUNE-001 P1, AC-GET-003) now follows the standard collapse;
// a non-override agent uses the un-overridden collapse result.
func TestResolveGLMReasoning_CodingMaxOverride(t *testing.T) {
	tests := []struct {
		name     string
		agent    string
		effort   string
		wantName string
	}{
		// Override set (now {manager-develop} only) → reasoning-max regardless of the collapse input.
		{"manager-develop input=low (would collapse to the low level) → override max", "manager-develop", EffortLevelLow, GLMStateMax},
		{"manager-develop input=high → override max", "manager-develop", EffortLevelHigh, GLMStateMax},
		{"manager-develop input=max → override agrees with collapse max", "manager-develop", EffortLevelMax, GLMStateMax},
		// builder-harness (removed from override by SPEC-GLM-EFFORT-TUNE-001 P1) → standard collapse.
		{"builder-harness input=low → low (AC-GET-003 make-or-break, re-anchored to low by SPEC-GLM-EFFORT-MAX-001)", "builder-harness", EffortLevelLow, GLMStateLow},
		{"builder-harness input=high → max (collapse of high, NOT override)", "builder-harness", EffortLevelHigh, GLMStateMax},
		{"builder-harness input=xhigh → max (collapse of xhigh, NOT override)", "builder-harness", EffortLevelXHigh, GLMStateMax},
		// Non-override agent → un-overridden collapse result.
		{"manager-git input=low → low (collapse, un-overridden)", "manager-git", EffortLevelLow, GLMStateLow},
		{"manager-spec input=high → max (collapse, un-overridden)", "manager-spec", EffortLevelHigh, GLMStateMax},
		{"super-advisor input=xhigh → max (collapse, not override)", "super-advisor", EffortLevelXHigh, GLMStateMax},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveGLMReasoning(tt.agent, tt.effort)
			if got.Name != tt.wantName {
				t.Errorf("ResolveGLMReasoning(%q, %q).Name = %q, want %q", tt.agent, tt.effort, got.Name, tt.wantName)
			}
		})
	}
}

// TestGLMCodingMaxOverrideAgents_ExactlyOne asserts the override set is EXACTLY
// {manager-develop} — the single code-producing run-phase agent (z.ai coding-task
// recommendation). builder-harness was removed by SPEC-GLM-EFFORT-TUNE-001 P1
// (AC-GET-001); it falls under the standard collapse (post SPEC-GLM-EFFORT-MAX-001:
// max for every effort above low).
func TestGLMCodingMaxOverrideAgents_ExactlyOne(t *testing.T) {
	got := GLMCodingMaxOverrideAgents()
	sort.Strings(got)
	want := []string{"manager-develop"}
	if len(got) != len(want) {
		t.Fatalf("GLMCodingMaxOverrideAgents() has %d members %v, want exactly 1 %v", len(got), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("GLMCodingMaxOverrideAgents()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
	// Membership predicate agrees with the (now singleton) set.
	if !IsGLMCodingMaxOverrideAgent("manager-develop") {
		t.Error("IsGLMCodingMaxOverrideAgent must be true for manager-develop")
	}
	if IsGLMCodingMaxOverrideAgent("builder-harness") {
		t.Error("IsGLMCodingMaxOverrideAgent must be false for builder-harness (removed by SPEC-GLM-EFFORT-TUNE-001 P1)")
	}
	if IsGLMCodingMaxOverrideAgent("manager-spec") || IsGLMCodingMaxOverrideAgent("sync-auditor") {
		t.Error("IsGLMCodingMaxOverrideAgent must be false for non-override agents")
	}
	// AC-GET-003 make-or-break behavioral assertion, re-anchored to the low effort
	// by SPEC-GLM-EFFORT-MAX-001 (plan B-2): at effort `high` the collapse and the
	// coding-max override now agree on max, so the not-overridden discrimination
	// is only observable at low — builder-harness stays at the low level (standard
	// collapse) where manager-develop is lifted to max (override).
	if got := ResolveGLMReasoning("builder-harness", EffortLevelLow); got.Name != GLMStateLow {
		t.Errorf("ResolveGLMReasoning(builder-harness, low).Name = %q, want %q (P1: no longer overridden)", got.Name, GLMStateLow)
	}
	if got := ResolveGLMReasoning("manager-develop", EffortLevelLow); got.Name != GLMStateMax {
		t.Errorf("ResolveGLMReasoning(manager-develop, low).Name = %q, want %q (coding-max override)", got.Name, GLMStateMax)
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

// TestResolveGLMReasoningForModel covers the model-aware per-agent resolution:
// under flash even a non-override agent at Claude effort low pins max (the
// low state does not exist on flash); under non-flash the coding-max override
// and the plain collapse keep their existing behavior.
func TestResolveGLMReasoningForModel(t *testing.T) {
	flash := config.DefaultGLM53Flash
	if got := ResolveGLMReasoningForModel(flash, "manager-spec", EffortLevelLow); got.Name != GLMStateMax || got.ReasoningEffort != GLMReasoningEffortMax {
		t.Errorf("ResolveGLMReasoningForModel(flash, manager-spec, low) = %+v, want max", got)
	}
	if got := ResolveGLMReasoningForModel(config.DefaultGLM53, "manager-spec", EffortLevelLow); got.Name != GLMStateLow {
		t.Errorf("ResolveGLMReasoningForModel(glm-5.3, manager-spec, low) = %+v, want low", got)
	}
	if got := ResolveGLMReasoningForModel(config.DefaultGLM53, "manager-develop", EffortLevelLow); got.Name != GLMStateMax {
		t.Errorf("ResolveGLMReasoningForModel(glm-5.3, manager-develop, low) = %+v, want max (coding-max override)", got)
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
