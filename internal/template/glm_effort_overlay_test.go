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
		planType string // must have NO effect on the result
		want     bool
	}{
		// TRUE cases — a GLM backend signal is present.
		{"team_mode=glm (primary moai glm signal)", "", config.TeamModeGLM, "", true},
		{"team_mode=cg (moai cg signal)", "", config.TeamModeCG, "", true},
		{"mode=glm (defensive dormant-field OR)", config.LLMModeGLM, "", "", true},
		{"mode=glm AND team_mode=cg", config.LLMModeGLM, config.TeamModeCG, "", true},
		// FALSE cases — no GLM signal (legacy non-GLM team_mode values + empty).
		{"team_mode=claude (legacy non-GLM)", "", config.TeamModeClaude, "", false},
		{"team_mode=hybrid (legacy non-GLM)", "", config.TeamModeHybrid, "", false},
		{"no signal (both empty)", "", "", "", false},
		// plan_type has NO effect on the predicate result.
		{"team_mode=glm with plan_type=api → still TRUE", "", config.TeamModeGLM, config.PlanTypeAPI, true},
		{"team_mode=glm with plan_type=subscription → still TRUE", "", config.TeamModeGLM, config.PlanTypeSubscription, true},
		{"no signal with plan_type=api → still FALSE", "", "", config.PlanTypeAPI, false},
		{"no signal with plan_type=subscription → still FALSE", "", "", config.PlanTypeSubscription, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.LLMConfig{Mode: tt.mode, TeamMode: tt.teamMode, PlanType: tt.planType}
			if got := IsGLMBackend(cfg); got != tt.want {
				t.Errorf("IsGLMBackend(mode=%q, team_mode=%q, plan_type=%q) = %v, want %v",
					tt.mode, tt.teamMode, tt.planType, got, tt.want)
			}
		})
	}
}

// TestCollapseClaudeEffortToGLM covers the REQ-MTP-027 5→3 collapse (AC-MTP-029):
// low→thinking-off; medium/high→reasoning-high; xhigh/max→reasoning-max; and the
// totality clause (unrecognized effort → documented GLM default state).
func TestCollapseClaudeEffortToGLM(t *testing.T) {
	tests := []struct {
		effort          string
		wantName        string
		wantThinking    bool
		wantReasoningEf string
	}{
		{EffortLevelLow, GLMStateThinkingOff, false, ""},
		{EffortLevelMedium, GLMStateReasoningHigh, true, GLMReasoningEffortHigh},
		{EffortLevelHigh, GLMStateReasoningHigh, true, GLMReasoningEffortHigh},
		{EffortLevelXHigh, GLMStateReasoningMax, true, GLMReasoningEffortMax},
		{EffortLevelMax, GLMStateReasoningMax, true, GLMReasoningEffortMax},
		// Totality: an unrecognized effort maps to the GLM default state
		// (reasoning-max = z.ai omit-default), no panic.
		{"bogus-unrecognized", GLMStateReasoningMax, true, GLMReasoningEffortMax},
		{"", GLMStateReasoningMax, true, GLMReasoningEffortMax},
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
		{"manager-develop input=low (would collapse thinking-off) → override max", "manager-develop", EffortLevelLow, GLMStateReasoningMax},
		{"manager-develop input=high → override max", "manager-develop", EffortLevelHigh, GLMStateReasoningMax},
		{"manager-develop input=max → override agrees with collapse max", "manager-develop", EffortLevelMax, GLMStateReasoningMax},
		// builder-harness (removed from override by SPEC-GLM-EFFORT-TUNE-001 P1) → standard collapse.
		{"builder-harness input=low → thinking-off (collapse, no longer overridden)", "builder-harness", EffortLevelLow, GLMStateThinkingOff},
		{"builder-harness input=high → reasoning-high (AC-GET-003 make-or-break)", "builder-harness", EffortLevelHigh, GLMStateReasoningHigh},
		{"builder-harness input=xhigh → reasoning-max (collapse of xhigh, NOT override)", "builder-harness", EffortLevelXHigh, GLMStateReasoningMax},
		// Non-override agent → un-overridden collapse result.
		{"manager-git input=low → thinking-off (collapse, un-overridden)", "manager-git", EffortLevelLow, GLMStateThinkingOff},
		{"manager-spec input=high → reasoning-high (collapse, un-overridden)", "manager-spec", EffortLevelHigh, GLMStateReasoningHigh},
		{"super-advisor input=xhigh → reasoning-max (collapse, not override)", "super-advisor", EffortLevelXHigh, GLMStateReasoningMax},
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
// (AC-GET-001); it falls under the standard collapse at reasoning-high.
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
	// AC-GET-003 make-or-break behavioral assertion: builder-harness now reaches
	// reasoning-high under a high Claude effort (standard collapse), NOT reasoning-max.
	if got := ResolveGLMReasoning("builder-harness", EffortLevelHigh); got.Name != GLMStateReasoningHigh {
		t.Errorf("ResolveGLMReasoning(builder-harness, high).Name = %q, want %q (P1: no longer overridden)", got.Name, GLMStateReasoningHigh)
	}
}

// TestApplyGLMEffortOverlay_EffortOnlyAndNoOp covers REQ-MTP-029 (AC-MTP-031):
// (a) under a GLM backend the overlay changes ONLY the effort representation —
// the model value is byte-identical; (b) under a non-GLM backend the overlay is
// an identity no-op.
func TestApplyGLMEffortOverlay_EffortOnlyAndNoOp(t *testing.T) {
	tests := []struct {
		name       string
		entry      TierProfileEntry
		agent      string
		glmBackend bool
		wantModel  string
		wantEffort string
	}{
		{
			name:       "GLM backend, non-override agent: model unchanged, effort → reasoning-high",
			entry:      TierProfileEntry{Model: "opus", Effort: EffortLevelHigh},
			agent:      "manager-spec",
			glmBackend: true,
			wantModel:  "opus", // byte-identical — the overlay never rewrites model
			wantEffort: GLMStateReasoningHigh,
		},
		{
			name:       "GLM backend, override agent: model unchanged, effort → reasoning-max",
			entry:      TierProfileEntry{Model: "sonnet", Effort: EffortLevelHigh},
			agent:      "manager-develop",
			glmBackend: true,
			wantModel:  "sonnet",
			wantEffort: GLMStateReasoningMax,
		},
		{
			name:       "non-GLM backend: identity no-op",
			entry:      TierProfileEntry{Model: "opus", Effort: EffortLevelHigh},
			agent:      "manager-spec",
			glmBackend: false,
			wantModel:  "opus",
			wantEffort: EffortLevelHigh, // unchanged Claude effort
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyGLMEffortOverlay(tt.entry, tt.agent, tt.glmBackend)
			if got.Model != tt.wantModel {
				t.Errorf("ApplyGLMEffortOverlay model = %q, want %q (overlay must never rewrite model)", got.Model, tt.wantModel)
			}
			if got.Effort != tt.wantEffort {
				t.Errorf("ApplyGLMEffortOverlay effort = %q, want %q", got.Effort, tt.wantEffort)
			}
		})
	}
}

// TestSessionGLMReasoningState confirms the Branch-B session-global delivery
// value derives from the coding-max override (reasoning-max), per the
// delivery-granularity limitation (research.md §D).
func TestSessionGLMReasoningState(t *testing.T) {
	got := SessionGLMReasoningState()
	if got.Name != GLMStateReasoningMax {
		t.Errorf("SessionGLMReasoningState().Name = %q, want %q (coding-max session default)", got.Name, GLMStateReasoningMax)
	}
	if !got.ThinkingEnabled || got.ReasoningEffort != GLMReasoningEffortMax {
		t.Errorf("SessionGLMReasoningState() = %+v, want thinking enabled + reasoning_effort=max", got)
	}
}

// TestSessionGLMReasoningStateForEffort verifies the MAIN-SESSION reasoning
// derivation that is driven by the web-set effort preference. Distinct from
// SessionGLMReasoningState() (the coding-max session default used for sub-agents
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
		{"low disables thinking", EffortLevelLow, GLMStateThinkingOff, false, ""},
		{"medium → reasoning high", EffortLevelMedium, GLMStateReasoningHigh, true, GLMReasoningEffortHigh},
		{"high → reasoning high", EffortLevelHigh, GLMStateReasoningHigh, true, GLMReasoningEffortHigh},
		{"xhigh → reasoning max", EffortLevelXHigh, GLMStateReasoningMax, true, GLMReasoningEffortMax},
		{"max → reasoning max", EffortLevelMax, GLMStateReasoningMax, true, GLMReasoningEffortMax},
		{"empty falls back to session default", "", GLMStateReasoningMax, true, GLMReasoningEffortMax},
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
