package template

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

func TestValidModelPolicies(t *testing.T) {
	policies := ValidModelPolicies()
	if len(policies) == 0 {
		t.Fatal("ValidModelPolicies() returned empty list")
	}
	if len(policies) != 3 {
		t.Errorf("ValidModelPolicies() returned %d items, want 3", len(policies))
	}

	expected := map[string]bool{"high": true, "medium": true, "low": true}
	for _, p := range policies {
		if !expected[p] {
			t.Errorf("unexpected policy: %q", p)
		}
	}
}

func TestIsValidModelPolicy(t *testing.T) {
	tests := []struct {
		policy string
		valid  bool
	}{
		{"high", true},
		{"medium", true},
		{"low", true},
		{"", false},
		{"ultra", false},
		{"HIGH", false},
		{"Medium", false},
		{"none", false},
	}

	for _, tt := range tests {
		t.Run(tt.policy, func(t *testing.T) {
			got := IsValidModelPolicy(tt.policy)
			if got != tt.valid {
				t.Errorf("IsValidModelPolicy(%q) = %v, want %v", tt.policy, got, tt.valid)
			}
		})
	}
}

// TestModelClaudeOpus48Constant verifies the claude-opus-4-8 model ID constant.
func TestModelClaudeOpus48Constant(t *testing.T) {
	if ModelIDOpus48 == "" {
		t.Error("ModelIDOpus48 constant is empty, want non-empty model ID")
	}
	want := "claude-opus-4-8"
	if ModelIDOpus48 != want {
		t.Errorf("ModelIDOpus48 = %q, want %q", ModelIDOpus48, want)
	}
}

// TestEffortLevelConstants verifies xhigh and max constants exist.
func TestEffortLevelConstants(t *testing.T) {
	if EffortLevelXHigh == "" {
		t.Error("EffortLevelXHigh constant is empty")
	}
	if EffortLevelMax == "" {
		t.Error("EffortLevelMax constant is empty")
	}
	if EffortLevelXHigh != "xhigh" {
		t.Errorf("EffortLevelXHigh = %q, want %q", EffortLevelXHigh, "xhigh")
	}
	if EffortLevelMax != "max" {
		t.Errorf("EffortLevelMax = %q, want %q", EffortLevelMax, "max")
	}
}

func TestNewDeployerWithRenderer(t *testing.T) {
	fsys := testFS()
	r := NewRenderer(fsys)
	d := NewDeployerWithRenderer(fsys, r)
	if d == nil {
		t.Fatal("NewDeployerWithRenderer returned nil")
	}
	// Verify it functions by listing templates
	list := d.ListTemplates()
	if len(list) == 0 {
		t.Error("ListTemplates() returned empty list from DeployerWithRenderer")
	}
}

func TestNewDeployerWithForceUpdate(t *testing.T) {
	fsys := testFS()
	d := NewDeployerWithForceUpdate(fsys, true)
	if d == nil {
		t.Fatal("NewDeployerWithForceUpdate returned nil")
	}
	list := d.ListTemplates()
	if len(list) == 0 {
		t.Error("ListTemplates() returned empty list from DeployerWithForceUpdate")
	}
}

func TestNewDeployerWithRendererAndForceUpdate(t *testing.T) {
	fsys := testFS()
	r := NewRenderer(fsys)
	d := NewDeployerWithRendererAndForceUpdate(fsys, r, true)
	if d == nil {
		t.Fatal("NewDeployerWithRendererAndForceUpdate returned nil")
	}
	list := d.ListTemplates()
	if len(list) == 0 {
		t.Error("ListTemplates() returned empty list from DeployerWithRendererAndForceUpdate")
	}
}

func TestDeployWithForceUpdate(t *testing.T) {
	root, mgr := setupDeployProject(t)
	fsys := testFS()

	// First deploy normally
	d := NewDeployer(fsys)
	if err := d.Deploy(context.Background(), root, mgr, nil); err != nil {
		t.Fatalf("initial Deploy error: %v", err)
	}

	// Modify a deployed file to simulate user changes
	claudeMDPath := filepath.Join(root, "CLAUDE.md")
	if err := os.WriteFile(claudeMDPath, []byte("user modified content"), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	// Deploy with forceUpdate=true should overwrite
	fd := NewDeployerWithForceUpdate(fsys, true)
	if err := fd.Deploy(context.Background(), root, mgr, nil); err != nil {
		t.Fatalf("force Deploy error: %v", err)
	}

	content, err := os.ReadFile(claudeMDPath)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if string(content) == "user modified content" {
		t.Error("forceUpdate did not overwrite user-modified file")
	}
}

func TestDeployWithTemplateRendering(t *testing.T) {
	tmplFS := fstest.MapFS{
		"config.yaml.tmpl": &fstest.MapFile{
			Data: []byte("project: {{.ProjectName}}\nversion: {{.Version}}\n"),
		},
	}

	root, mgr := setupDeployProject(t)
	r := NewRenderer(tmplFS)
	d := NewDeployerWithRenderer(tmplFS, r)

	ctx := NewTemplateContext(
		WithProject("test-project", root),
		WithVersion("1.0.0"),
	)

	if err := d.Deploy(context.Background(), root, mgr, ctx); err != nil {
		t.Fatalf("Deploy error: %v", err)
	}

	// Verify the rendered file (without .tmpl suffix)
	content, err := os.ReadFile(filepath.Join(root, "config.yaml"))
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if !containsString(string(content), "project: test-project") {
		t.Errorf("rendered content missing project name: %s", content)
	}
	if !containsString(string(content), "version: 1.0.0") {
		t.Errorf("rendered content missing version: %s", content)
	}
}

func TestDeployShellScriptPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix file permissions not supported on Windows")
	}

	fsys := fstest.MapFS{
		"scripts/run.sh": &fstest.MapFile{
			Data: []byte("#!/bin/bash\necho hello"),
		},
		"docs/readme.md": &fstest.MapFile{
			Data: []byte("# Readme"),
		},
	}

	root, mgr := setupDeployProject(t)
	d := NewDeployer(fsys)

	if err := d.Deploy(context.Background(), root, mgr, nil); err != nil {
		t.Fatalf("Deploy error: %v", err)
	}

	// Shell scripts should have executable permissions
	info, err := os.Stat(filepath.Join(root, "scripts", "run.sh"))
	if err != nil {
		t.Fatalf("Stat error: %v", err)
	}
	perm := info.Mode().Perm()
	if perm&0o100 == 0 {
		t.Errorf("shell script should be executable, got permissions: %o", perm)
	}

	// Non-shell files should NOT be executable
	info2, err := os.Stat(filepath.Join(root, "docs", "readme.md"))
	if err != nil {
		t.Fatalf("Stat error: %v", err)
	}
	perm2 := info2.Mode().Perm()
	if perm2&0o100 != 0 {
		t.Errorf("non-shell file should not be executable, got permissions: %o", perm2)
	}
}

func TestDeployExistingUserFile(t *testing.T) {
	fsys := testFS()
	root, mgr := setupDeployProject(t)

	// Pre-create a file that is NOT tracked in manifest
	claudeDir := filepath.Join(root, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}
	settingsPath := filepath.Join(root, ".claude", "settings.json")
	if err := os.WriteFile(settingsPath, []byte(`{"user": true}`), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	d := NewDeployer(fsys)
	if err := d.Deploy(context.Background(), root, mgr, nil); err != nil {
		t.Fatalf("Deploy error: %v", err)
	}

	// The pre-existing file should be preserved (not overwritten)
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if string(content) != `{"user": true}` {
		t.Errorf("existing user file was overwritten: %s", content)
	}

	// It should be tracked as user_created in manifest
	entry, found := mgr.GetEntry(".claude/settings.json")
	if !found {
		t.Error("expected manifest entry for user file")
	} else if entry.Provenance != manifest.UserCreated {
		t.Errorf("provenance = %v, want UserCreated", entry.Provenance)
	}
}

// === SPEC-MODEL-TIER-PLANTYPE-001 M2 ===
//
// Unified plan_type × tier profile fidelity + ApplyTierProfile replace-both.
// The expected matrix is re-declared test-side (test-side EXPECTED values are
// the point of a fidelity test — plan.md §G); production carries exactly one
// structure. Values copied verbatim from spec.md §B.6 (api / Plan A rev2) and
// §B.7 (subscription / Plan B).

type tierProfileExpectCell struct {
	model  string
	effort string
}

// expectedTierProfiles mirrors spec.md §B.6/§B.7: plan → agent → [max, medium, low].
var expectedTierProfiles = map[string]map[string][3]tierProfileExpectCell{
	"api": {
		"manager-spec":    {{"fable", "high"}, {"fable", "high"}, {"opus", "high"}},
		"plan-auditor":    {{"fable", "high"}, {"fable", "high"}, {"opus", "high"}},
		"sync-auditor":    {{"fable", "high"}, {"opus", "high"}, {"opus", "medium"}},
		"manager-design":  {{"fable", "high"}, {"fable", "high"}, {"opus", "high"}},
		"super-advisor":   {{"fable", "xhigh"}, {"fable", "high"}, {"opus", "high"}},
		"manager-develop": {{"fable", "high"}, {"opus", "high"}, {"opus", "medium"}},
		"builder-harness": {{"opus", "high"}, {"opus", "medium"}, {"opus", "medium"}},
		"e2e-specialist":  {{"opus", "high"}, {"opus", "medium"}, {"opus", "medium"}},
		"manager-docs":    {{"sonnet", "medium"}, {"sonnet", "low"}, {"sonnet", "low"}},
		"manager-git":     {{"sonnet", "low"}, {"sonnet", "low"}, {"sonnet", "low"}},
		"Explore":         {{"inherit", "medium"}, {"inherit", "low"}, {"inherit", "low"}},
	},
	"subscription": {
		"manager-spec":    {{"opus", "high"}, {"opus", "high"}, {"opus", "medium"}},
		"plan-auditor":    {{"opus", "high"}, {"opus", "medium"}, {"sonnet", "high"}},
		"sync-auditor":    {{"opus", "high"}, {"opus", "medium"}, {"sonnet", "high"}},
		"manager-design":  {{"opus", "high"}, {"opus", "medium"}, {"sonnet", "high"}},
		"super-advisor":   {{"opus", "xhigh"}, {"opus", "high"}, {"opus", "medium"}},
		"manager-develop": {{"sonnet", "high"}, {"sonnet", "high"}, {"sonnet", "high"}},
		"builder-harness": {{"sonnet", "high"}, {"sonnet", "medium"}, {"sonnet", "medium"}},
		"e2e-specialist":  {{"sonnet", "high"}, {"sonnet", "medium"}, {"sonnet", "medium"}},
		"manager-docs":    {{"sonnet", "low"}, {"sonnet", "low"}, {"sonnet", "low"}},
		"manager-git":     {{"sonnet", "low"}, {"sonnet", "low"}, {"sonnet", "low"}},
		"Explore":         {{"inherit", "medium"}, {"inherit", "low"}, {"inherit", "low"}},
	},
}

var tierColumnOrder = [3]string{"max", "medium", "low"}

// TestTierProfileMatrixFidelity (AC-MTP-006, REQ-MTP-006..009) asserts every
// (plan × tier × agent) cell of GetTierProfileEntry equals the spec.md §B.6/§B.7
// {model, effort} pair. 2 plans × 11 agents × 3 tiers = 66 asserted cells.
func TestTierProfileMatrixFidelity(t *testing.T) {
	asserted := 0
	for plan, agents := range expectedTierProfiles {
		for agent, cols := range agents {
			for i, tier := range tierColumnOrder {
				want := cols[i]
				got, ok := GetTierProfileEntry(plan, agent, tier)
				if !ok {
					t.Errorf("GetTierProfileEntry(%q, %q, %q): ok=false, want a row", plan, agent, tier)
					continue
				}
				if got.Model != want.model || got.Effort != want.effort {
					t.Errorf("GetTierProfileEntry(%q, %q, %q) = {%q, %q}, want {%q, %q}",
						plan, agent, tier, got.Model, got.Effort, want.model, want.effort)
				}
				asserted++
			}
		}
	}
	if asserted != 66 {
		t.Errorf("asserted %d cells, want 66 (2 plans × 11 agents × 3 tiers)", asserted)
	}
}

// TestTierProfiles_AllElevenAgents (REQ-MTP-009) asserts both plans carry explicit
// rows for all 11 retained agents — the auditors, manager-design, and
// super-advisor moved off implicit inherit; e2e-specialist added by
// SPEC-E2E-REVIVAL-001; Explore is retained as an explicit
// inherit row for display/derivation surfaces.
func TestTierProfiles_AllElevenAgents(t *testing.T) {
	agents := []string{
		"manager-spec", "plan-auditor", "sync-auditor", "manager-design",
		"super-advisor", "manager-develop", "builder-harness", "e2e-specialist",
		"manager-docs", "manager-git", "Explore",
	}
	for _, plan := range []string{"api", "subscription"} {
		for _, a := range agents {
			if _, ok := GetTierProfileEntry(plan, a, "medium"); !ok {
				t.Errorf("plan %q missing explicit row for agent %q", plan, a)
			}
		}
	}
}

// TestMapModelPolicyToTier (AC-MTP-012, REQ-MTP-012) asserts the legacy
// ModelPolicy→tier mapping is tier-only (high→max, medium→medium, low→low) and
// returns a tier string with NO plan value in its signature.
func TestMapModelPolicyToTier(t *testing.T) {
	tests := []struct {
		policy ModelPolicy
		want   string
	}{
		{ModelPolicyHigh, "max"},
		{ModelPolicyMedium, "medium"},
		{ModelPolicyLow, "low"},
		{ModelPolicy(""), "medium"},      // empty → D6 default-when-absent
		{ModelPolicy("bogus"), "medium"}, // unknown → D6 default-when-absent
	}
	for _, tt := range tests {
		if got := MapModelPolicyToTier(tt.policy); got != tt.want {
			t.Errorf("MapModelPolicyToTier(%q) = %q, want %q", tt.policy, got, tt.want)
		}
	}
}

// TestMapModelPolicyToEffort asserts the runtime-LAUNCH effort projection of the
// legacy {high,medium,low} ModelPolicy vocabulary: high→high, medium→medium,
// low→low on the EFFORT axis. This is DISTINCT from MapModelPolicyToTier, which
// projects high→max on the TIER axis. Empty/unknown → "" (no override), so an
// absent model_policy preserves today's launch behavior byte-identically.
func TestMapModelPolicyToEffort(t *testing.T) {
	tests := []struct {
		policy ModelPolicy
		want   string
	}{
		{ModelPolicyHigh, "high"},
		{ModelPolicyMedium, "medium"},
		{ModelPolicyLow, "low"},
		{ModelPolicy(""), ""},    // empty → no override (byte-identical to today)
		{ModelPolicy("xyz"), ""}, // unknown → no override
	}
	for _, tt := range tests {
		if got := MapModelPolicyToEffort(tt.policy); got != tt.want {
			t.Errorf("MapModelPolicyToEffort(%q) = %q, want %q", tt.policy, got, tt.want)
		}
	}
}

// TestNormalizeToTier asserts the call-site resolver accepts BOTH the canonical
// performance-tier vocabulary ({max, medium, low}) and the legacy ModelPolicy
// vocabulary ({high, medium, low}), defaulting to medium (D6).
func TestNormalizeToTier(t *testing.T) {
	cases := map[string]string{
		"max": "max", "medium": "medium", "low": "low",
		"high": "max", "": "medium", "bogus": "medium",
	}
	for in, want := range cases {
		if got := NormalizeToTier(in); got != want {
			t.Errorf("NormalizeToTier(%q) = %q, want %q", in, got, want)
		}
	}
}

// writeTierAgentFixture creates a shipped-style agent .md file carrying both
// model: and effort: frontmatter under .claude/agents/moai/.
func writeTierAgentFixture(t *testing.T, root, name, model, effort string) {
	t.Helper()
	dir := filepath.Join(root, ".claude", "agents", "moai")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	content := "---\nname: " + name + "\nmodel: " + model + "\neffort: " + effort + "\n---\n# " + name + "\n"
	if err := os.WriteFile(filepath.Join(dir, name+".md"), []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile %s: %v", name, err)
	}
}

// newTierManifest sets up a loaded manifest manager rooted at root.
func newTierManifest(t *testing.T, root string) manifest.Manager {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("MkdirAll .moai: %v", err)
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("manifest Load: %v", err)
	}
	return mgr
}

// assertFrontmatter reads the agent file and asserts its model:/effort: lines
// equal the wanted values.
func assertFrontmatter(t *testing.T, root, name, wantModel, wantEffort string) {
	t.Helper()
	p := filepath.Join(root, ".claude", "agents", "moai", name+".md")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("ReadFile %s: %v", name, err)
	}
	s := string(b)
	if !containsString(s, "model: "+wantModel+"\n") {
		t.Errorf("%s: want model: %q, got:\n%s", name, wantModel, s)
	}
	if !containsString(s, "effort: "+wantEffort+"\n") {
		t.Errorf("%s: want effort: %q, got:\n%s", name, wantEffort, s)
	}
}

// TestApplyTierProfile_ReplaceBoth_API (AC-MTP-008/011, REQ-MTP-010/011) — under
// (api, medium) the pass REPLACES both existing model: and effort: lines with the
// profile values: manager-develop (seeded sonnet/xhigh) → opus/high; plan-auditor
// (seeded inherit/max, i.e. moved off inherit) → fable/high.
func TestApplyTierProfile_ReplaceBoth_API(t *testing.T) {
	root := t.TempDir()
	writeTierAgentFixture(t, root, "manager-develop", "sonnet", "xhigh")
	writeTierAgentFixture(t, root, "plan-auditor", "inherit", "max")
	mgr := newTierManifest(t, root)

	if err := ApplyTierProfile(root, "api", "medium", mgr); err != nil {
		t.Fatalf("ApplyTierProfile error: %v", err)
	}
	assertFrontmatter(t, root, "manager-develop", "opus", "high")
	assertFrontmatter(t, root, "plan-auditor", "fable", "high")

	// Replace-both: the seeded values must be gone.
	b, _ := os.ReadFile(filepath.Join(root, ".claude", "agents", "moai", "manager-develop.md"))
	if containsString(string(b), "model: sonnet\n") || containsString(string(b), "effort: xhigh\n") {
		t.Errorf("manager-develop retained a seeded value (replace-both failed):\n%s", b)
	}
}

// TestApplyTierProfile_ReplaceBoth_Subscription (AC-MTP-009, REQ-MTP-010/011) —
// under (subscription, max): manager-develop → sonnet/high; manager-spec →
// opus/high; manager-docs → sonnet/low.
func TestApplyTierProfile_ReplaceBoth_Subscription(t *testing.T) {
	root := t.TempDir()
	writeTierAgentFixture(t, root, "manager-develop", "opus", "low")
	writeTierAgentFixture(t, root, "manager-spec", "sonnet", "low")
	writeTierAgentFixture(t, root, "manager-docs", "opus", "high")
	mgr := newTierManifest(t, root)

	if err := ApplyTierProfile(root, "subscription", "max", mgr); err != nil {
		t.Fatalf("ApplyTierProfile error: %v", err)
	}
	assertFrontmatter(t, root, "manager-develop", "sonnet", "high")
	assertFrontmatter(t, root, "manager-spec", "opus", "high")
	assertFrontmatter(t, root, "manager-docs", "sonnet", "low")
}

// TestApplyTierProfile_PlanBranchSelection asserts the api vs subscription branch
// yields different {model, effort} for the same agent+tier (manager-develop,
// medium: api → opus/high, subscription → sonnet/high).
func TestApplyTierProfile_PlanBranchSelection(t *testing.T) {
	rootAPI := t.TempDir()
	writeTierAgentFixture(t, rootAPI, "manager-develop", "sonnet", "low")
	if err := ApplyTierProfile(rootAPI, "api", "medium", newTierManifest(t, rootAPI)); err != nil {
		t.Fatalf("api ApplyTierProfile: %v", err)
	}
	assertFrontmatter(t, rootAPI, "manager-develop", "opus", "high")

	rootSub := t.TempDir()
	writeTierAgentFixture(t, rootSub, "manager-develop", "opus", "low")
	if err := ApplyTierProfile(rootSub, "subscription", "medium", newTierManifest(t, rootSub)); err != nil {
		t.Fatalf("subscription ApplyTierProfile: %v", err)
	}
	assertFrontmatter(t, rootSub, "manager-develop", "sonnet", "high")
}

// TestApplyTierProfile_UnknownAgentByteIdentical (AC-MTP-013, REQ-MTP-013) — an
// agent with no profile row is left byte-identical.
func TestApplyTierProfile_UnknownAgentByteIdentical(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".claude", "agents", "moai")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	original := "---\nname: custom-agent\nmodel: opus\neffort: max\n---\n# custom\n"
	if err := os.WriteFile(filepath.Join(dir, "custom-agent.md"), []byte(original), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := ApplyTierProfile(root, "api", "medium", newTierManifest(t, root)); err != nil {
		t.Fatalf("ApplyTierProfile error: %v", err)
	}
	got, _ := os.ReadFile(filepath.Join(dir, "custom-agent.md"))
	if string(got) != original {
		t.Errorf("unknown agent modified; want byte-identical:\ngot:\n%s\nwant:\n%s", got, original)
	}
}

// TestApplyTierProfile_InheritAgentSkipped — an inherit-model agent (Explore) is
// skipped byte-identically (inherit is never written as a model: value).
func TestApplyTierProfile_InheritAgentSkipped(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".claude", "agents", "moai")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	original := "---\nname: Explore\nmodel: inherit\neffort: high\n---\n# Explore\n"
	if err := os.WriteFile(filepath.Join(dir, "Explore.md"), []byte(original), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := ApplyTierProfile(root, "api", "medium", newTierManifest(t, root)); err != nil {
		t.Fatalf("ApplyTierProfile error: %v", err)
	}
	got, _ := os.ReadFile(filepath.Join(dir, "Explore.md"))
	if string(got) != original {
		t.Errorf("inherit agent modified; want skipped byte-identical:\ngot:\n%s\nwant:\n%s", got, original)
	}
}

// TestResolveProjectPlanType asserts the effective plan-type read from llm.yaml:
// absent/commented → subscription default; explicit api → api.
func TestResolveProjectPlanType(t *testing.T) {
	root := t.TempDir()
	if got := ResolveProjectPlanType(root); got != "subscription" {
		t.Errorf("absent llm.yaml: got %q, want subscription", got)
	}

	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte("llm:\n  plan_type: \"api\"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if got := ResolveProjectPlanType(root); got != "api" {
		t.Errorf("explicit api: got %q, want api", got)
	}

	if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte("llm:\n  # plan_type: api\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if got := ResolveProjectPlanType(root); got != "subscription" {
		t.Errorf("commented plan_type: got %q, want subscription", got)
	}
}

// TestGetTierProfileEntry_Fallbacks (REQ-MTP-002 BC + plan.md D6) — an
// out-of-set plan type falls back to the subscription profile, and an
// out-of-set tier falls back to the medium column.
func TestGetTierProfileEntry_Fallbacks(t *testing.T) {
	// Unknown plan → subscription branch (BC default). subscription/medium
	// manager-develop = sonnet/high.
	got, ok := GetTierProfileEntry("enterprise", "manager-develop", "medium")
	if !ok {
		t.Fatal("unknown plan: ok=false, want subscription fallback row")
	}
	if got.Model != "sonnet" || got.Effort != "high" {
		t.Errorf("unknown plan fallback = {%q, %q}, want subscription {sonnet, high}", got.Model, got.Effort)
	}

	// Unknown tier → medium column (D6). api/<medium> manager-develop = opus/high.
	got, ok = GetTierProfileEntry("api", "manager-develop", "bogus-tier")
	if !ok {
		t.Fatal("unknown tier: ok=false, want a row")
	}
	if got.Model != "opus" || got.Effort != "high" {
		t.Errorf("unknown tier fallback = {%q, %q}, want medium column {opus, high}", got.Model, got.Effort)
	}
}

// TestApplyTierProfile_InsertsEffortWhenAbsent — an agent file carrying model:
// but NO effort: line gets effort inserted (the insertEffortInFrontmatter path).
func TestApplyTierProfile_InsertsEffortWhenAbsent(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".claude", "agents", "moai")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// manager-git has model: but no effort: line.
	content := "---\nname: manager-git\nmodel: opus\n---\n# manager-git\n"
	if err := os.WriteFile(filepath.Join(dir, "manager-git.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := ApplyTierProfile(root, "api", "max", newTierManifest(t, root)); err != nil {
		t.Fatalf("ApplyTierProfile error: %v", err)
	}
	// api/max manager-git = sonnet/low.
	assertFrontmatter(t, root, "manager-git", "sonnet", "low")
}

// TestApplyTierProfile_NoAgentsDir — an absent agents directory is a graceful
// no-op (no error).
func TestApplyTierProfile_NoAgentsDir(t *testing.T) {
	root := t.TempDir()
	if err := ApplyTierProfile(root, "api", "medium", newTierManifest(t, root)); err != nil {
		t.Fatalf("ApplyTierProfile on missing agents dir returned error: %v", err)
	}
}

// TestApplyTierProfile_AlreadyAtTarget — an agent already at the profile
// {model, effort} is left byte-identical (no rewrite).
func TestApplyTierProfile_AlreadyAtTarget(t *testing.T) {
	root := t.TempDir()
	// api/max manager-git = sonnet/low — seed the fixture already at that value.
	writeTierAgentFixture(t, root, "manager-git", "sonnet", "low")
	p := filepath.Join(root, ".claude", "agents", "moai", "manager-git.md")
	before, _ := os.ReadFile(p)
	if err := ApplyTierProfile(root, "api", "max", newTierManifest(t, root)); err != nil {
		t.Fatalf("ApplyTierProfile error: %v", err)
	}
	after, _ := os.ReadFile(p)
	if string(before) != string(after) {
		t.Errorf("already-at-target agent was rewritten:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// containsString checks if s contains substr.
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// writeLLMSection writes a minimal llm.yaml under root's sections dir with the
// given body (already indented under `llm:`).
func writeLLMSection(t *testing.T, root, body string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile llm.yaml: %v", err)
	}
}

func readLLMSection(t *testing.T, root string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "llm.yaml"))
	if err != nil {
		t.Fatalf("ReadFile llm.yaml: %v", err)
	}
	return string(b)
}

// TestApplyPlanType (REQ-MTP-016/018) — persisting the plan type rewrites the
// plan_type: line in llm.yaml, preserving surrounding indentation and other
// lines byte-for-byte.
func TestApplyPlanType(t *testing.T) {
	t.Run("replaces quoted subscription with api", func(t *testing.T) {
		root := t.TempDir()
		writeLLMSection(t, root, "llm:\n  plan_type: \"subscription\"\n  performance_tier: \"medium\"\n")
		if err := ApplyPlanType(root, "api"); err != nil {
			t.Fatalf("ApplyPlanType: %v", err)
		}
		got := readLLMSection(t, root)
		if !containsString(got, "plan_type: api") {
			t.Errorf("plan_type not rewritten to api:\n%s", got)
		}
		if ResolveProjectPlanType(root) != "api" {
			t.Errorf("ResolveProjectPlanType after apply = %q, want api", ResolveProjectPlanType(root))
		}
		// The performance_tier line must be untouched.
		if !containsString(got, "performance_tier: \"medium\"") {
			t.Errorf("performance_tier line was altered:\n%s", got)
		}
	})

	t.Run("replaces empty value", func(t *testing.T) {
		root := t.TempDir()
		writeLLMSection(t, root, "llm:\n    plan_type: \"\"\n")
		if err := ApplyPlanType(root, "subscription"); err != nil {
			t.Fatalf("ApplyPlanType: %v", err)
		}
		if ResolveProjectPlanType(root) != "subscription" {
			t.Errorf("after apply = %q, want subscription", ResolveProjectPlanType(root))
		}
		// 4-space indentation must be preserved.
		if !containsString(readLLMSection(t, root), "    plan_type: subscription") {
			t.Errorf("indentation not preserved:\n%s", readLLMSection(t, root))
		}
	})

	t.Run("absent file is graceful no-op", func(t *testing.T) {
		root := t.TempDir()
		if err := ApplyPlanType(root, "api"); err != nil {
			t.Errorf("absent llm.yaml should be a no-op, got err: %v", err)
		}
	})

	t.Run("already at target does not rewrite", func(t *testing.T) {
		root := t.TempDir()
		// Unquoted value already equal to the target — the regex replacement
		// yields byte-identical content, so ApplyPlanType returns without writing.
		writeLLMSection(t, root, "llm:\n  plan_type: api\n")
		before := readLLMSection(t, root)
		if err := ApplyPlanType(root, "api"); err != nil {
			t.Fatalf("ApplyPlanType: %v", err)
		}
		if got := readLLMSection(t, root); got != before {
			t.Errorf("already-at-target llm.yaml was rewritten:\nbefore:\n%s\nafter:\n%s", before, got)
		}
	})
}

// TestResolveProjectPerformanceTier (REQ-MTP-018) — the update path reads the
// persisted performance_tier from llm.yaml; absent/empty → medium (D6 default).
func TestResolveProjectPerformanceTier(t *testing.T) {
	root := t.TempDir()
	if got := ResolveProjectPerformanceTier(root); got != "medium" {
		t.Errorf("absent llm.yaml: got %q, want medium", got)
	}

	writeLLMSection(t, root, "llm:\n  performance_tier: \"max\"\n")
	if got := ResolveProjectPerformanceTier(root); got != "max" {
		t.Errorf("explicit max: got %q, want max", got)
	}

	writeLLMSection(t, root, "llm:\n  performance_tier: \"\"\n")
	if got := ResolveProjectPerformanceTier(root); got != "medium" {
		t.Errorf("empty performance_tier: got %q, want medium", got)
	}

	writeLLMSection(t, root, "llm:\n  performance_tier: low\n")
	if got := ResolveProjectPerformanceTier(root); got != "low" {
		t.Errorf("unquoted low: got %q, want low", got)
	}
}

// TestTierProfileAgents covers the model-policy preview accessor (M4): it returns
// exactly the 11 retained agents in stable display order, includes Explore for the
// display/derivation surface, every listed agent resolves in both plan profiles,
// and the returned slice is a defensive copy the caller cannot use to mutate the
// package-level order.
func TestTierProfileAgents(t *testing.T) {
	agents := TierProfileAgents()
	if len(agents) != 11 {
		t.Fatalf("TierProfileAgents() returned %d agents, want 11", len(agents))
	}
	foundExplore := false
	for _, a := range agents {
		if a == "Explore" {
			foundExplore = true
		}
		if _, ok := GetTierProfileEntry("api", a, PerformanceTierMax); !ok {
			t.Errorf("api profile has no row for %q", a)
		}
		if _, ok := GetTierProfileEntry("subscription", a, PerformanceTierMax); !ok {
			t.Errorf("subscription profile has no row for %q", a)
		}
	}
	if !foundExplore {
		t.Error("TierProfileAgents() must include Explore for the display surface")
	}
	// Defensive copy: mutating the returned slice must not affect a fresh call.
	agents[0] = "MUTATED"
	if TierProfileAgents()[0] == "MUTATED" {
		t.Error("TierProfileAgents() must return a defensive copy, not the package slice")
	}
}
