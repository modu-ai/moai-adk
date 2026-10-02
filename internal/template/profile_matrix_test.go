package template

// profile_matrix_test.go — SPEC-WEB-AGENTFM-RESTORE-001 M1: tests for the
// re-ported per-agent profile matrix (SPEC-MODEL-PROFILE-MATRIX-001 machinery,
// restored with cells RE-DERIVED from the CURRENT config defaults per plan
// §B-1(b)/§F M1). The blueprint cells (opus/sonnet, 2026-09-27 snapshot) are
// stale; these tests pin the cells to config.DefaultClaudeTier* as the single
// source so a future tier bump flows through without a cell edit.

import (
	"slices"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// effortRank orders the effort levels by reasoning depth so the monotonicity
// assertion (high >= medium >= low per row) is a number compare.
func effortRank(effort string) int {
	switch effort {
	case EffortLevelHigh:
		return 2
	case EffortLevelMedium:
		return 1
	case EffortLevelLow:
		return 0
	}
	return -1
}

// profileColumns are the three canonical profile keys in display order.
var profileColumns = []string{config.ProfileHigh, config.ProfileMedium, config.ProfileLow}

// matrixModels are the only model aliases a default cell may carry — the
// CURRENT config claude_models defaults (High/Medium columns). ClaudeModels.Low
// (haiku) is deliberately absent: the No-Haiku policy keeps it out of the
// matrix (available only as an explicit override).
var matrixModels = map[string]bool{
	defaultMatrixModelHigh:   true,
	defaultMatrixModelMedium: true,
}

// TestProfileMatrixAgents_CurrentRoster pins the display roster's derivation
// contract: ProfileMatrixAgents is the CANONICAL retained roster
// (template.RetainedAgents — the single roster literal, rosterguard-asserted)
// filtered to the console rows, so the names here are never restated as a
// second literal. Semantic pins: mission-governor absent, Explore off-console,
// every definition-file agent present.
func TestProfileMatrixAgents_CurrentRoster(t *testing.T) {
	agents := ProfileMatrixAgents()
	var want []string
	for _, name := range RetainedAgents() {
		if name == "Explore" {
			continue
		}
		want = append(want, name)
	}
	if !slices.Equal(agents, want) {
		t.Errorf("ProfileMatrixAgents = %v, want the canonical roster minus Explore: %v", agents, want)
	}
	if slices.Contains(agents, "mission-governor") {
		t.Error("mission-governor must not appear (retired from the catalog)")
	}
	if slices.Contains(agents, "Explore") {
		t.Error("Explore must not appear (no definition file — off the console surface)")
	}
	if !slices.Contains(agents, "manager-todo") {
		t.Error("manager-todo must appear (current catalog)")
	}
}

// TestDefaultProfileMatrix_CellsAreCurrentConfigDefaults asserts every cell's
// MODEL is one of the current config claude_models defaults (High/Medium —
// the named re-derivation source, plan §B-1(b)) and its EFFORT is one of the
// 5-level vocabulary's policy levels the matrix uses. A cell carrying any
// other value is a stale restatement.
func TestDefaultProfileMatrix_CellsAreCurrentConfigDefaults(t *testing.T) {
	matrix := DefaultProfileMatrix()
	if len(matrix) == 0 {
		t.Fatal("DefaultProfileMatrix is empty — the matrix machine did not re-port")
	}
	validEfforts := map[string]bool{EffortLevelLow: true, EffortLevelMedium: true, EffortLevelHigh: true}
	for _, col := range profileColumns {
		agents, ok := matrix[col]
		if !ok {
			t.Fatalf("matrix has no %q column", col)
		}
		if len(agents) == 0 {
			t.Fatalf("column %q is empty", col)
		}
		for agent, cell := range agents {
			if !matrixModels[cell.Model] {
				t.Errorf("matrix[%s][%s].model = %q — not a current config claude_models default (and haiku/fable/inherit never enter the matrix)", col, agent, cell.Model)
			}
			if !validEfforts[cell.Effort] {
				t.Errorf("matrix[%s][%s].effort = %q — outside the matrix policy levels {low, medium, high}", col, agent, cell.Effort)
			}
		}
	}
}

// TestDefaultProfileMatrix_RowMonotonicityAndNoSentinels asserts the old
// matrix's invariants hold under the re-derived cells: per-agent depth never
// increases as the profile column descends, and no haiku / fable / inherit
// model ever appears inside the matrix (inherit survives only as the
// unmapped-agent fallback).
func TestDefaultProfileMatrix_RowMonotonicityAndNoSentinels(t *testing.T) {
	matrix := DefaultProfileMatrix()
	agents := ProfileMatrixAgents()
	for _, agent := range agents {
		prev := 3
		for _, col := range profileColumns {
			cell, ok := matrix[col][agent]
			if !ok {
				t.Fatalf("matrix[%s] lacks agent %q — every display agent needs a cell in every column", col, agent)
			}
			if rank := effortRank(cell.Effort); rank > prev {
				t.Errorf("row %q breaks monotonicity at column %s: %+v deeper than the column above", agent, col, cell)
			} else {
				prev = rank
			}
			switch cell.Model {
			case ModelInherit, "fable", "haiku":
				t.Errorf("matrix[%s][%s].model = %q — sentinel/fable/haiku must not appear inside the matrix", col, agent, cell.Model)
			}
		}
	}
}

// TestResolveAgentModelEffort_Precedence asserts the D2 precedence on the
// re-ported resolver: an llm.agent_overrides entry wins; else the Go-default
// cell under the effective profile; an unknown profile resolves the medium
// column; an unmapped agent resolves the inherit sentinel unmapped.
func TestResolveAgentModelEffort_Precedence(t *testing.T) {
	t.Run("override wins", func(t *testing.T) {
		ov := config.ModelEffort{Model: "haiku", Effort: "low"}
		cfg := config.LLMConfig{Profile: config.ProfileHigh, AgentOverrides: map[string]config.ModelEffort{"manager-develop": ov}}
		me, mapped := ResolveAgentModelEffort(cfg, "manager-develop")
		if me != ov || !mapped {
			t.Errorf("override resolution = %+v,%v; want %+v,true", me, mapped, ov)
		}
	})
	t.Run("default cell under each profile", func(t *testing.T) {
		for _, col := range profileColumns {
			cfg := config.LLMConfig{Profile: col}
			me, mapped := ResolveAgentModelEffort(cfg, "super-advisor")
			want := DefaultProfileMatrix()[col]["super-advisor"]
			if me != want || !mapped {
				t.Errorf("profile %s: = %+v,%v; want %+v,true", col, me, mapped, want)
			}
		}
	})
	t.Run("unrecognized profile falls to the medium column", func(t *testing.T) {
		cfg := config.LLMConfig{Profile: "bogus"}
		me, mapped := ResolveAgentModelEffort(cfg, "manager-git")
		want := DefaultProfileMatrix()[config.ProfileMedium]["manager-git"]
		if me != want || !mapped {
			t.Errorf("= %+v,%v; want medium-column %+v,true", me, mapped, want)
		}
	})
	t.Run("unmapped agent inherits", func(t *testing.T) {
		me, mapped := ResolveAgentModelEffort(config.LLMConfig{}, "totally-user-agent")
		if mapped || me.Model != ModelInherit || me.Effort != "" {
			t.Errorf("= %+v,%v; want {inherit \"\"},false", me, mapped)
		}
	})
}

// TestValidPerformanceTiers_SelectorVocabulary pins the restored console
// selector's wire vocabulary to {max, medium, low} (REQ-AFR-003 closed set;
// AC-AFR-002 persists a max submission verbatim to llm.profile). The client-only
// "custom" pseudo-state is NOT a member.
func TestValidPerformanceTiers_SelectorVocabulary(t *testing.T) {
	want := []string{PerformanceTierMax, PerformanceTierMedium, PerformanceTierLow}
	if !slices.Equal(ValidPerformanceTiers(), want) {
		t.Errorf("ValidPerformanceTiers = %v, want %v", ValidPerformanceTiers(), want)
	}
	for _, v := range want {
		if !IsValidPerformanceTier(v) {
			t.Errorf("IsValidPerformanceTier(%q) = false, want true", v)
		}
	}
	for _, v := range []string{"custom", "", "bogus", "high"} {
		if IsValidPerformanceTier(v) {
			t.Errorf("IsValidPerformanceTier(%q) = true, want false — outside the selector wire set", v)
		}
	}
}

// TestAgentGroup_CurrentMembership pins the group layer to the current roster:
// manager-todo is mapped, mission-governor is not, Explore keeps its mapped
// cell (old-config overrides for it stay resolvable).
func TestAgentGroup_CurrentMembership(t *testing.T) {
	if g, ok := AgentGroup("manager-todo"); !ok || g == "" {
		t.Errorf("AgentGroup(manager-todo) = %q,%v; want a mapped group", g, ok)
	}
	if _, ok := AgentGroup("mission-governor"); ok {
		t.Error("AgentGroup(mission-governor) mapped — the retired agent must not be a member")
	}
	if g, ok := AgentGroup("Explore"); !ok || g != GroupExplore {
		t.Errorf("AgentGroup(Explore) = %q,%v; want %q,true", g, ok, GroupExplore)
	}
}
