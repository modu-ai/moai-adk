package template

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestResolveAgentModelEffort_MatrixAFidelity covers REQ-MPM-009/010/011/012 /
// AC-MPM-005: profile:max with no overrides resolves every agent to the Matrix A
// max column exactly.
func TestResolveAgentModelEffort_MatrixAFidelity(t *testing.T) {
	cfg := config.LLMConfig{Profile: "max"}
	want := map[string]config.ModelEffort{
		"manager-spec":    {Model: "fable", Effort: "medium"},
		"plan-auditor":    {Model: "fable", Effort: "medium"},
		"sync-auditor":    {Model: "fable", Effort: "medium"},
		"manager-develop": {Model: "fable", Effort: "low"},
		"super-advisor":   {Model: "fable", Effort: "medium"},
		"manager-design":  {Model: "opus", Effort: "high"},
		"builder-harness": {Model: "opus", Effort: "high"},
		"e2e-tester":      {Model: "opus", Effort: "high"},
		"manager-docs":    {Model: "sonnet", Effort: "medium"},
		"manager-git":     {Model: "sonnet", Effort: "low"},
	}
	for agent, exp := range want {
		got, hasGroup := ResolveAgentModelEffort(cfg, agent)
		if !hasGroup {
			t.Errorf("%s: expected group membership", agent)
			continue
		}
		if got != exp {
			t.Errorf("%s max: got %+v, want %+v", agent, got, exp)
		}
	}
}

// TestResolveAgentModelEffort_LowColumn covers AC-MPM-013 spot-checks on the low
// column.
func TestResolveAgentModelEffort_LowColumn(t *testing.T) {
	cfg := config.LLMConfig{Profile: "low"}
	cases := map[string]config.ModelEffort{
		"manager-spec":  {Model: "opus", Effort: "low"},
		"super-advisor": {Model: "opus", Effort: "high"},
		"manager-git":   {Model: "sonnet", Effort: "low"},
	}
	for agent, exp := range cases {
		got, _ := ResolveAgentModelEffort(cfg, agent)
		if got != exp {
			t.Errorf("%s low: got %+v, want %+v", agent, got, exp)
		}
	}
}

// TestResolveAgentModelEffort_OverridePrecedence covers REQ-MPM-012 / AC-MPM-006:
// an override wins for its agent and does not affect a sibling in the same group.
func TestResolveAgentModelEffort_OverridePrecedence(t *testing.T) {
	cfg := config.LLMConfig{
		Profile: "medium",
		AgentOverrides: map[string]config.ModelEffort{
			"manager-spec": {Model: "opus", Effort: "xhigh"},
		},
	}
	got, _ := ResolveAgentModelEffort(cfg, "manager-spec")
	if (got != config.ModelEffort{Model: "opus", Effort: "xhigh"}) {
		t.Errorf("override should win: got %+v", got)
	}
	// plan-auditor shares spec_auditors but is unaffected → medium group cell.
	got, _ = ResolveAgentModelEffort(cfg, "plan-auditor")
	if (got != config.ModelEffort{Model: "opus", Effort: "high"}) {
		t.Errorf("plan-auditor medium cell should be unaffected: got %+v", got)
	}
}

// TestResolveAgentModelEffort_Inherit covers REQ-MPM-013 / AC-MPM-007: Explore
// and unknown agents resolve to inherit with hasGroup=false.
func TestResolveAgentModelEffort_Inherit(t *testing.T) {
	cfg := config.LLMConfig{Profile: "max"}
	for _, agent := range []string{"Explore", "some-user-agent"} {
		got, hasGroup := ResolveAgentModelEffort(cfg, agent)
		if hasGroup {
			t.Errorf("%s should have no group", agent)
		}
		if got.Model != modelInherit {
			t.Errorf("%s should resolve to inherit, got %q", agent, got.Model)
		}
	}
}

// TestResolveAgentModelEffort_ConfigProfilesOverrideDefault covers REQ-MPM-010:
// a config llm.profiles cell overrides the Go default fallback.
func TestResolveAgentModelEffort_ConfigProfilesOverrideDefault(t *testing.T) {
	cfg := config.LLMConfig{
		Profile: "max",
		Profiles: map[string]map[string]config.ModelEffort{
			"max": {GroupDocs: {Model: "opus", Effort: "high"}},
		},
	}
	got, _ := ResolveAgentModelEffort(cfg, "manager-docs")
	if (got != config.ModelEffort{Model: "opus", Effort: "high"}) {
		t.Errorf("config profiles cell should override Go default: got %+v", got)
	}
	// manager-git absent from config profiles → Go default max cell.
	got, _ = ResolveAgentModelEffort(cfg, "manager-git")
	if (got != config.ModelEffort{Model: "sonnet", Effort: "low"}) {
		t.Errorf("absent cell should fall back to Go default: got %+v", got)
	}
}

// TestResolveAgentModelEffort_LegacyAlias covers AC-MPM-002: a legacy config with
// performance_tier and no profile resolves through the alias.
func TestResolveAgentModelEffort_LegacyAlias(t *testing.T) {
	cfg := config.LLMConfig{PerformanceTier: "max"} // no profile
	got, _ := ResolveAgentModelEffort(cfg, "manager-develop")
	if (got != config.ModelEffort{Model: "fable", Effort: "low"}) {
		t.Errorf("legacy perf_tier max should resolve to max column: got %+v", got)
	}
}

// TestDefaultProfileMatrix_NoHaiku covers AC-MPM-024: the matrix has zero haiku.
func TestDefaultProfileMatrix_NoHaiku(t *testing.T) {
	for profile, groups := range DefaultProfileMatrix() {
		for group, me := range groups {
			if me.Model == "haiku" {
				t.Errorf("haiku found at %s/%s — HaikuResidualRule violation", profile, group)
			}
		}
	}
}
