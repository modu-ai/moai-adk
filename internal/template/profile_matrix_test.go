package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// writeProfileLLM writes a llm.yaml under root and returns its path.
func writeProfileLLM(t *testing.T, root, body string) string {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	p := filepath.Join(dir, "llm.yaml")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return p
}

// TestApplyProfile_RoundTripStripsRetiredKeys covers REQ-MPM-005 / AC-MPM-003: a
// write updating the llm section removes plan_type + the claude_models block and
// writes profile:.
func TestApplyProfile_RoundTripStripsRetiredKeys(t *testing.T) {
	root := t.TempDir()
	p := writeProfileLLM(t, root, `llm:
    plan_type: "subscription"
    performance_tier: "max"
    profile: "medium"
    claude_models:
        high: "opus"
        medium: "sonnet"
        low: "sonnet"
    glm:
        base_url: "https://api.z.ai/api/anthropic"
`)
	// "max" is the superseded top-column name: readable, but normalized to "high"
	// on write so it is never persisted again.
	if err := ApplyProfile(root, "max"); err != nil {
		t.Fatalf("ApplyProfile: %v", err)
	}
	got, _ := os.ReadFile(p)
	s := string(got)
	if strings.Contains(s, "plan_type") {
		t.Errorf("plan_type not stripped:\n%s", s)
	}
	if strings.Contains(s, "claude_models") || strings.Contains(s, `high: "opus"`) {
		t.Errorf("claude_models block not stripped:\n%s", s)
	}
	if !strings.Contains(s, "profile: high") {
		t.Errorf("profile: high (normalized from max) not written:\n%s", s)
	}
	// The glm block (a sibling AFTER claude_models) must survive the strip.
	if !strings.Contains(s, "base_url:") {
		t.Errorf("glm block was over-stripped:\n%s", s)
	}
}

// TestApplyProfile_InsertsProfileWhenAbsent covers the migration insert path — a
// legacy config with no profile: key gains one.
func TestApplyProfile_InsertsProfileWhenAbsent(t *testing.T) {
	root := t.TempDir()
	p := writeProfileLLM(t, root, "llm:\n    performance_tier: \"low\"\n")
	if err := ApplyProfile(root, "low"); err != nil {
		t.Fatalf("ApplyProfile: %v", err)
	}
	got, _ := os.ReadFile(p)
	if !strings.Contains(string(got), "profile: low") {
		t.Errorf("profile not inserted:\n%s", got)
	}
}

// TestResolveAgentModelEffort_MatrixAFidelity covers REQ-MPM-009/010/011/012 /
// AC-MPM-005: profile:high with no overrides resolves every one of the 11 mapped
// agents to the high column exactly.
func TestResolveAgentModelEffort_MatrixAFidelity(t *testing.T) {
	cfg := config.LLMConfig{Profile: "high"}
	want := map[string]config.ModelEffort{
		"manager-spec":    {Model: "opus", Effort: "high"},
		"plan-auditor":    {Model: "opus", Effort: "high"},
		"sync-auditor":    {Model: "opus", Effort: "high"},
		"manager-develop": {Model: "opus", Effort: "max"},
		"super-advisor":   {Model: "opus", Effort: "max"},
		"manager-design":  {Model: "opus", Effort: "high"},
		"builder-harness": {Model: "opus", Effort: "high"},
		"e2e-tester":      {Model: "opus", Effort: "medium"},
		"manager-docs":    {Model: "opus", Effort: "low"},
		"manager-git":     {Model: "sonnet", Effort: "low"},
		"Explore":         {Model: "sonnet", Effort: "low"},
	}
	for agent, exp := range want {
		got, mapped := ResolveAgentModelEffort(cfg, agent)
		if !mapped {
			t.Errorf("%s: expected matrix membership", agent)
			continue
		}
		if got != exp {
			t.Errorf("%s high: got %+v, want %+v", agent, got, exp)
		}
	}
}

// TestResolveAgentModelEffort_LowColumn covers AC-MPM-013 spot-checks on the low
// column, including super-advisor whose former low>medium inversion is fixed.
func TestResolveAgentModelEffort_LowColumn(t *testing.T) {
	cfg := config.LLMConfig{Profile: "low"}
	cases := map[string]config.ModelEffort{
		"manager-spec":    {Model: "opus", Effort: "low"},
		"super-advisor":   {Model: "opus", Effort: "medium"},
		"manager-develop": {Model: "opus", Effort: "low"},
		"manager-docs":    {Model: "sonnet", Effort: "low"},
		"manager-git":     {Model: "sonnet", Effort: "low"},
	}
	for agent, exp := range cases {
		got, _ := ResolveAgentModelEffort(cfg, agent)
		if got != exp {
			t.Errorf("%s low: got %+v, want %+v", agent, got, exp)
		}
	}
}

// TestDefaultProfileMatrix_Shape asserts the structural invariants of the
// per-agent matrix: 3 profiles x 11 agents = 33 cells, models restricted to
// {opus, sonnet} (fable is retired from the matrix — it is dominated by Opus 5
// on the coding axis at every effort), efforts restricted to
// {low, medium, high, max} (no `xhigh` cell — on Opus 5 xhigh scores the same
// as high at materially higher cost), and no `inherit` inside the matrix.
func TestDefaultProfileMatrix_Shape(t *testing.T) {
	m := DefaultProfileMatrix()
	wantProfiles := []string{PerformanceTierHigh, PerformanceTierMedium, PerformanceTierLow}
	if len(m) != len(wantProfiles) {
		t.Fatalf("profile count = %d, want %d", len(m), len(wantProfiles))
	}
	okModel := map[string]bool{"opus": true, "sonnet": true}
	okEffort := map[string]bool{
		EffortLevelLow: true, EffortLevelMedium: true,
		EffortLevelHigh: true, EffortLevelMax: true,
	}
	total := 0
	for _, profile := range wantProfiles {
		agents, ok := m[profile]
		if !ok {
			t.Fatalf("profile %q missing from matrix", profile)
		}
		if len(agents) != len(ProfileMatrixAgents()) {
			t.Errorf("profile %q has %d cells, want %d", profile, len(agents), len(ProfileMatrixAgents()))
		}
		for _, agent := range ProfileMatrixAgents() {
			cell, ok := agents[agent]
			if !ok {
				t.Errorf("profile %q missing agent %q", profile, agent)
				continue
			}
			total++
			if !okModel[cell.Model] {
				t.Errorf("%s/%s model %q outside {opus, sonnet}", profile, agent, cell.Model)
			}
			if !okEffort[cell.Effort] {
				t.Errorf("%s/%s effort %q outside {low, medium, high, max}", profile, agent, cell.Effort)
			}
		}
	}
	if total != 33 {
		t.Errorf("matrix cell count = %d, want 33", total)
	}
}

// TestDefaultProfileMatrix_Monotone asserts every agent row is non-increasing
// across high >= medium >= low on a combined (model rank, effort rank) ordering.
// This is the invariant the former super-advisor low>medium cell violated.
func TestDefaultProfileMatrix_Monotone(t *testing.T) {
	modelRank := map[string]int{"sonnet": 0, "opus": 1}
	effortRank := map[string]int{
		EffortLevelLow: 0, EffortLevelMedium: 1,
		EffortLevelHigh: 2, EffortLevelXHigh: 3, EffortLevelMax: 4,
	}
	m := DefaultProfileMatrix()
	rank := func(profile, agent string) int {
		c := m[profile][agent]
		return modelRank[c.Model]*10 + effortRank[c.Effort]
	}
	for _, agent := range ProfileMatrixAgents() {
		hi, med, lo := rank(PerformanceTierHigh, agent), rank(PerformanceTierMedium, agent), rank(PerformanceTierLow, agent)
		if hi < med {
			t.Errorf("%s: high rank %d < medium rank %d — non-monotone", agent, hi, med)
		}
		if med < lo {
			t.Errorf("%s: medium rank %d < low rank %d — non-monotone", agent, med, lo)
		}
	}
}

// TestResolveHarnessAgentModelEffort covers the /moai:harness generation path:
// every purpose class resolves to HarnessAgentModel with the effort of its
// profile-matrix row, an unknown class falls back to the implement class, and a
// config harness_agents cell overrides the derived effort while the model stays
// pinned.
func TestResolveHarnessAgentModelEffort(t *testing.T) {
	// Derived: effort borrowed from the class row, model always pinned.
	want := map[string]string{
		HarnessClassReadOnlyExtract:     EffortLevelLow,    // Explore row
		HarnessClassMechanicalTransform: EffortLevelLow,    // manager-git row
		HarnessClassSynthesize:          EffortLevelLow,    // manager-docs row
		HarnessClassResearch:            EffortLevelHigh,   // plan-auditor row
		HarnessClassVerifyJudge:         EffortLevelHigh,   // sync-auditor row
		HarnessClassImplement:           EffortLevelMax,    // manager-develop row
		HarnessClassDesignArchitecture:  EffortLevelHigh,   // manager-design row
	}
	cfg := config.LLMConfig{Profile: "high"}
	for class, exp := range want {
		got, known := ResolveHarnessAgentModelEffort(cfg, class)
		if !known {
			t.Errorf("%s: should be a known class", class)
		}
		if got.Model != HarnessAgentModel {
			t.Errorf("%s: model = %q, want %q (harness agents are model-uniform)", class, got.Model, HarnessAgentModel)
		}
		if got.Effort != exp {
			t.Errorf("%s: effort = %q, want %q", class, got.Effort, exp)
		}
	}

	// Unknown class → implement fallback, reported as unknown.
	got, known := ResolveHarnessAgentModelEffort(cfg, "not-a-class")
	if known {
		t.Errorf("unknown class should report known=false")
	}
	if got.Effort != EffortLevelMax || got.Model != HarnessAgentModel {
		t.Errorf("unknown class should fall back to implement: got %+v", got)
	}

	// Config override wins on effort; model stays pinned even if config says otherwise.
	override := config.LLMConfig{
		Profile: "high",
		HarnessAgents: map[string]map[string]config.ModelEffort{
			"high": {HarnessClassSynthesize: {Model: "sonnet", Effort: EffortLevelMedium}},
		},
	}
	got, _ = ResolveHarnessAgentModelEffort(override, HarnessClassSynthesize)
	if got.Effort != EffortLevelMedium {
		t.Errorf("config effort should win: got %+v", got)
	}
	if got.Model != HarnessAgentModel {
		t.Errorf("config model must be ignored (pinned): got %+v", got)
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
	// plan-auditor shares spec_auditors but is unaffected → its own medium cell,
	// which the phase-weighted policy sets to high (distinct from the xhigh
	// override above, so the assertion still discriminates).
	got, _ = ResolveAgentModelEffort(cfg, "plan-auditor")
	if (got != config.ModelEffort{Model: "opus", Effort: "high"}) {
		t.Errorf("plan-auditor medium cell should be unaffected: got %+v", got)
	}
}

// TestResolveAgentModelEffort_Inherit covers REQ-MPM-013 / AC-MPM-007: only
// user-added / unknown agents resolve to inherit with hasGroup=false. The
// built-in Explore now has an explicit group (see
// TestResolveAgentModelEffort_ExploreProfileInvariant) and is no longer in this
// inherit set.
func TestResolveAgentModelEffort_Inherit(t *testing.T) {
	cfg := config.LLMConfig{Profile: "max"}
	for _, agent := range []string{"some-user-agent", "another-custom-agent"} {
		got, hasGroup := ResolveAgentModelEffort(cfg, agent)
		if hasGroup {
			t.Errorf("%s should have no group", agent)
		}
		if got.Model != modelInherit {
			t.Errorf("%s should resolve to inherit, got %q", agent, got.Model)
		}
	}
}

// TestResolveAgentModelEffort_ExploreProfileInvariant covers the product
// decision that the built-in Explore agent resolves to sonnet/low with
// hasGroup=true across all three profile columns (profile-invariant, like
// docs/git).
func TestResolveAgentModelEffort_ExploreProfileInvariant(t *testing.T) {
	for _, profile := range []string{"high", "medium", "low"} {
		cfg := config.LLMConfig{Profile: profile}
		got, hasGroup := ResolveAgentModelEffort(cfg, "Explore")
		if !hasGroup {
			t.Errorf("profile %q: Explore should now have a group", profile)
		}
		if (got != config.ModelEffort{Model: "sonnet", Effort: "low"}) {
			t.Errorf("profile %q: Explore got %+v, want sonnet/low", profile, got)
		}
	}
}

// TestResolveAgentModelEffort_ConfigProfilesOverrideDefault covers REQ-MPM-010:
// a config llm.profiles cell overrides the Go default fallback.
func TestResolveAgentModelEffort_ConfigProfilesOverrideDefault(t *testing.T) {
	cfg := config.LLMConfig{
		Profile: "high",
		Profiles: map[string]map[string]config.ModelEffort{
			// Keyed by agent NAME now, not by group.
			"high": {"manager-docs": {Model: "opus", Effort: "high"}},
		},
	}
	got, _ := ResolveAgentModelEffort(cfg, "manager-docs")
	if (got != config.ModelEffort{Model: "opus", Effort: "high"}) {
		t.Errorf("config profiles cell should override Go default: got %+v", got)
	}
	// manager-git absent from config profiles → Go default high cell.
	got, _ = ResolveAgentModelEffort(cfg, "manager-git")
	if (got != config.ModelEffort{Model: "sonnet", Effort: "low"}) {
		t.Errorf("absent cell should fall back to Go default: got %+v", got)
	}
}

// TestResolveAgentModelEffort_StaleGroupKeyedMirror asserts a pre-rename config
// whose profiles mirror is keyed by GROUP name degrades gracefully: the lookup
// misses and the Go default per-agent cell is used instead of erroring.
func TestResolveAgentModelEffort_StaleGroupKeyedMirror(t *testing.T) {
	cfg := config.LLMConfig{
		Profile: "high",
		Profiles: map[string]map[string]config.ModelEffort{
			// Deliberately unequal to the Go default high cell for manager-docs
			// (opus/low) so honoring the stale group key would be observable.
			"high": {GroupDocs: {Model: "sonnet", Effort: "low"}},
		},
	}
	got, mapped := ResolveAgentModelEffort(cfg, "manager-docs")
	if !mapped {
		t.Fatalf("manager-docs should still resolve")
	}
	if (got != config.ModelEffort{Model: "opus", Effort: "low"}) {
		t.Errorf("stale group-keyed cell should be ignored, Go default used: got %+v", got)
	}
}

// TestResolveAgentModelEffort_LegacyAlias covers AC-MPM-002: a legacy config with
// performance_tier and no profile resolves through the alias. performance_tier
// "max" folds to the high column.
func TestResolveAgentModelEffort_LegacyAlias(t *testing.T) {
	cfg := config.LLMConfig{PerformanceTier: "max"} // no profile
	got, _ := ResolveAgentModelEffort(cfg, "manager-develop")
	if (got != config.ModelEffort{Model: "opus", Effort: "max"}) {
		t.Errorf("legacy perf_tier max should resolve to the high column: got %+v", got)
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
