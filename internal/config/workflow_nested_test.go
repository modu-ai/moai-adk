package config

import (
	"path/filepath"
	"runtime"
	"testing"

	"gopkg.in/yaml.v3"
)

// templateMoaiDir resolves the .moai directory of the embedded template SSOT,
// which contains config/sections/workflow.yaml (the AC-WSE-003 oracle fixture).
func templateMoaiDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// thisFile = .../internal/config/workflow_nested_test.go
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	return filepath.Join(repoRoot, "internal", "template", "templates", ".moai")
}

// TestWorkflowYAMLUnmarshalProductionFixture loads the production workflow.yaml
// (template SSOT) via the standard loader and asserts the AC-WSE-003 values
// (REQ-WSE-003; the workflow.team.* subset was removed with the Agent Teams
// static layer per SPEC-AGENT-TEAM-RETIRE-001 — a residual team: block in the
// yaml is ignored by the non-strict unmarshal until the M3 template cleanup).
func TestWorkflowYAMLUnmarshalProductionFixture(t *testing.T) {
	t.Parallel()

	moaiDir := templateMoaiDir(t)
	cfg, err := NewLoader().Load(moaiDir)
	if err != nil {
		t.Fatalf("Load(%q): %v", moaiDir, err)
	}

	wf := cfg.Workflow

	if got := wf.AutoClear.Enabled; got != true {
		t.Errorf("AutoClear.Enabled: got %v, want true", got)
	}
	if got := wf.AutoClear.TokenThreshold; got != 150000 {
		t.Errorf("AutoClear.TokenThreshold: got %d, want 150000", got)
	}
	if got := wf.TokenBudget.Plan; got != 30000 {
		t.Errorf("TokenBudget.Plan: got %d, want 30000", got)
	}
	if got := wf.TokenBudget.Run; got != 180000 {
		t.Errorf("TokenBudget.Run: got %d, want 180000", got)
	}
	if got := wf.TokenBudget.Sync; got != 40000 {
		t.Errorf("TokenBudget.Sync: got %d, want 40000", got)
	}
	if got := wf.LoopPrevention.MaxIterations; got != 100 {
		t.Errorf("LoopPrevention.MaxIterations: got %d, want 100", got)
	}
	if got := wf.Worktree.SessionNamePattern; got != "moai-{ProjectName}-{SPEC-ID}" {
		t.Errorf("Worktree.SessionNamePattern: got %q, want %q", got, "moai-{ProjectName}-{SPEC-ID}")
	}
	if got := wf.Worktree.TmuxPreferred; got != true {
		t.Errorf("Worktree.TmuxPreferred: got %v, want true", got)
	}
}

// TestWorkflowYAMLUnmarshal_OmittedTokenBudget_PreservesDefaults verifies that
// when a yaml document omits the token_budget block, unmarshalling over the
// populated default struct retains the construction-time defaults rather than
// collapsing to zero-values (Edge-WSE-003).
func TestWorkflowYAMLUnmarshal_OmittedTokenBudget_PreservesDefaults(t *testing.T) {
	t.Parallel()

	// Seed with defaults, then unmarshal a partial yaml that omits token_budget.
	wrapper := &workflowFileWrapper{Workflow: NewDefaultWorkflowConfig()}
	partial := []byte(`
workflow:
  execution_mode: solo
  worktree:
    auto_create: true
`)
	if err := yaml.Unmarshal(partial, wrapper); err != nil {
		t.Fatalf("yaml.Unmarshal: %v", err)
	}
	wf := wrapper.Workflow

	// token_budget omitted → defaults preserved.
	if wf.TokenBudget.Plan != 30000 {
		t.Errorf("TokenBudget.Plan: got %d, want 30000 (default preserved)", wf.TokenBudget.Plan)
	}
	if wf.TokenBudget.Run != 180000 {
		t.Errorf("TokenBudget.Run: got %d, want 180000 (default preserved)", wf.TokenBudget.Run)
	}
	// present keys overrode defaults.
	if wf.ExecutionMode != "solo" {
		t.Errorf("ExecutionMode: got %q, want %q", wf.ExecutionMode, "solo")
	}
	if !wf.Worktree.AutoCreate {
		t.Error("Worktree.AutoCreate: expected partial-yaml override to true")
	}
	// worktree.auto_merge was not present in the partial yaml → default true preserved.
	if !wf.Worktree.AutoMerge {
		t.Error("Worktree.AutoMerge: expected default true to be preserved")
	}
}

// TestWorkflowYAMLUnmarshal_LegacyFlatYamlTypeMismatch_BehaviorDocumented
// documents the behavior when a legacy FLAT yaml key (auto_clear: true scalar)
// is present. Because the FLAT field was renamed to AutoClearLegacy with
// yaml:"-", and the new AutoClear is a nested struct, a scalar bool at the
// auto_clear path is a TYPE MISMATCH and yaml.Unmarshal returns a decode error
// (deterministic per the yaml:"-" tag design) (Edge-WSE-004).
func TestWorkflowYAMLUnmarshal_LegacyFlatYamlTypeMismatch_BehaviorDocumented(t *testing.T) {
	t.Parallel()

	wrapper := &workflowFileWrapper{Workflow: NewDefaultWorkflowConfig()}
	legacy := []byte(`
workflow:
  auto_clear: true
`)
	err := yaml.Unmarshal(legacy, wrapper)
	// Documented behavior: scalar bool against nested struct → decode error.
	if err == nil {
		t.Error("expected yaml decode error for legacy scalar auto_clear against nested struct, got nil")
	}
}

// TestWorkflowConfigInconsistentRoleProfileKeys removed with the Agent Teams
// static layer (SPEC-AGENT-TEAM-RETIRE-001 M1) — its subject types
// (workflow.team role_profiles) no longer exist.
