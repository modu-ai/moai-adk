package config

// agentic_trunk_test.go — SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001 M1 (card t1831).
//
// The value-set, disposition, integration-target, local-integration
// predicate and develop-branch-projection tests for the fifth workflow
// value. Reuses the fixture helpers of loader_workflow_disposition_test.go
// (same package, the M2 owner-by-contract file).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAgenticTrunkValueSet pins the value-set and disposition facts
// (REQ-AGT-001, REQ-AGT-002 / AC-AGT-001): agentic-trunk is the fifth and
// last allowed value, classifies valid and not git-flow by exact
// case-sensitive match, and the four existing values classify as before.
func TestAgenticTrunkValueSet(t *testing.T) {
	t.Run("agentic-trunk is the fifth and last allowed value", func(t *testing.T) {
		got := AllowedWorkflows()
		if len(got) != 5 {
			t.Fatalf("AllowedWorkflows() has %d entries, want 5: %v", len(got), got)
		}
		if got[len(got)-1] != WorkflowAgenticTrunk {
			t.Errorf("AllowedWorkflows() last entry = %q, want %q (fifth value appended)", got[len(got)-1], WorkflowAgenticTrunk)
		}
		if got[4] != "agentic-trunk" {
			t.Errorf("AllowedWorkflows()[4] = %q, want the literal agentic-trunk", got[4])
		}
	})

	t.Run("classifies valid and not git-flow", func(t *testing.T) {
		if got := ClassifyWorkflowDisposition(WorkflowAgenticTrunk); got != DispositionValidNonGitFlow {
			t.Errorf("ClassifyWorkflowDisposition(%q) = %q, want %q (the class github-flow has)", WorkflowAgenticTrunk, got, DispositionValidNonGitFlow)
		}
	})

	t.Run("exact case-sensitive match", func(t *testing.T) {
		tests := []struct {
			value string
			want  WorkflowDisposition
		}{
			{value: "Agentic-Trunk", want: DispositionInvalid},  // case differs
			{value: "agentic_trunk", want: DispositionInvalid},  // underscore spelling
			{value: "agentic-trunk ", want: DispositionInvalid}, // trailing space
			{value: "", want: DispositionInvalid},               // empty value present
		}
		for _, tt := range tests {
			if got := ClassifyWorkflowDisposition(tt.value); got != tt.want {
				t.Errorf("ClassifyWorkflowDisposition(%q) = %q, want %q", tt.value, got, tt.want)
			}
		}
	})

	t.Run("the four existing values classify as before", func(t *testing.T) {
		tests := []struct {
			value string
			want  WorkflowDisposition
		}{
			{value: WorkflowGitHubFlow, want: DispositionValidNonGitFlow},
			{value: WorkflowGitFlow, want: DispositionGitFlow},
			{value: WorkflowGitLabFlow, want: DispositionValidNonGitFlow},
			{value: WorkflowReleaseFlow, want: DispositionValidNonGitFlow},
		}
		for _, tt := range tests {
			if got := ClassifyWorkflowDisposition(tt.value); got != tt.want {
				t.Errorf("ClassifyWorkflowDisposition(%q) = %q, want %q (unchanged)", tt.value, got, tt.want)
			}
		}
	})

	t.Run("reader seam classifies the fixture value", func(t *testing.T) {
		root := t.TempDir()
		dispositionWrite(t, root, gitStrategyBody("manual", WorkflowAgenticTrunk, ""))
		got := LoadGitFlowIntegrationConfig(root)
		if got.Disposition != DispositionValidNonGitFlow {
			t.Errorf("Disposition = %q, want %q", got.Disposition, DispositionValidNonGitFlow)
		}
		if got.Workflow != WorkflowAgenticTrunk {
			t.Errorf("Workflow raw = %q, want %q", got.Workflow, WorkflowAgenticTrunk)
		}
		if got.IsGitFlow() {
			t.Error("IsGitFlow() = true on agentic-trunk — it must take the github-flow class")
		}
	})
}

// TestAgenticTrunkTarget pins the integration-target table row and the
// develop-branch seam projection (REQ-AGT-003, REQ-AGT-004 / AC-AGT-003):
// the target is the fixed main in every mode, independent of the
// develop_branch, environment and release_branch_prefix keys; the legacy
// develop-branch seam projects to main in manual mode only.
func TestAgenticTrunkTarget(t *testing.T) {
	t.Run("manual", func(t *testing.T) {
		root := t.TempDir()
		body := gitStrategyBody("manual", WorkflowAgenticTrunk,
			"        develop_branch: develop\n        environment: staging\n        release_branch_prefix: release/\n")
		dispositionWrite(t, root, body)
		got := LoadGitFlowIntegrationConfig(root)
		if got.IntegrationTarget != "main" {
			t.Errorf("IntegrationTarget = %q, want the fixed main", got.IntegrationTarget)
		}
		if got.DevelopBranch != "main" {
			t.Errorf("DevelopBranch seam = %q, want main (projected)", got.DevelopBranch)
		}
		if !got.UsesLocalIntegration() {
			t.Error("UsesLocalIntegration() = false on manual agentic-trunk")
		}
	})

	t.Run("personal", func(t *testing.T) {
		root := t.TempDir()
		body := gitStrategyBody("personal", WorkflowAgenticTrunk,
			"        develop_branch: develop\n        environment: staging\n")
		dispositionWrite(t, root, body)
		got := LoadGitFlowIntegrationConfig(root)
		if got.IntegrationTarget != "main" {
			t.Errorf("IntegrationTarget = %q, want main (mode-independent)", got.IntegrationTarget)
		}
		if got.DevelopBranch != "" {
			t.Errorf("DevelopBranch seam = %q, want empty in personal mode", got.DevelopBranch)
		}
		if got.UsesLocalIntegration() {
			t.Error("UsesLocalIntegration() = true in personal mode — the seam is manual-only")
		}
	})

	t.Run("team", func(t *testing.T) {
		root := t.TempDir()
		body := gitStrategyBody("team", WorkflowAgenticTrunk,
			"        release_branch_prefix: release/\n")
		dispositionWrite(t, root, body)
		got := LoadGitFlowIntegrationConfig(root)
		if got.IntegrationTarget != "main" {
			t.Errorf("IntegrationTarget = %q, want main (mode-independent)", got.IntegrationTarget)
		}
		if got.DevelopBranch != "" {
			t.Errorf("DevelopBranch seam = %q, want empty in team mode", got.DevelopBranch)
		}
	})

	t.Run("environment and release_branch_prefix change nothing", func(t *testing.T) {
		// The pure table: no key leaks into the agentic-trunk target.
		profile := ModeProfile{
			DevelopBranch:       "develop",
			Environment:         "staging",
			ReleaseBranchPrefix: "release/",
			MainBranch:          "trunk",
		}
		if got := WorkflowIntegrationTarget(WorkflowAgenticTrunk, profile); got != "main" {
			t.Errorf("WorkflowIntegrationTarget(agentic-trunk, all keys set) = %q, want main", got)
		}
	})
}

// TestAgenticTrunkLocalIntegrationPredicate pins the local-integration
// predicate (REQ-AGT-005 / AC-AGT-004): true for manual-mode git-flow and
// manual-mode agentic-trunk, false otherwise; the existing git-flow-only
// predicate keeps its present answers.
func TestAgenticTrunkLocalIntegrationPredicate(t *testing.T) {
	tests := []struct {
		name            string
		body            string // empty body = write no file
		wantLocal       bool
		wantGitFlowOnly bool
	}{
		{name: "git-flow manual", body: gitStrategyBody("manual", WorkflowGitFlow, "        develop_branch: develop\n"), wantLocal: true, wantGitFlowOnly: true},
		{name: "agentic-trunk manual", body: gitStrategyBody("manual", WorkflowAgenticTrunk, ""), wantLocal: true, wantGitFlowOnly: false},
		{name: "agentic-trunk team", body: gitStrategyBody("team", WorkflowAgenticTrunk, ""), wantLocal: false, wantGitFlowOnly: false},
		{name: "github-flow manual", body: gitStrategyBody("manual", WorkflowGitHubFlow, ""), wantLocal: false, wantGitFlowOnly: false},
		{name: "absent file", body: "", wantLocal: false, wantGitFlowOnly: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			dispositionWrite(t, root, tt.body)
			got := LoadGitFlowIntegrationConfig(root)
			if gotLocal := got.UsesLocalIntegration(); gotLocal != tt.wantLocal {
				t.Errorf("UsesLocalIntegration() = %v, want %v", gotLocal, tt.wantLocal)
			}
			if gotFlow := got.IsGitFlow(); gotFlow != tt.wantGitFlowOnly {
				t.Errorf("IsGitFlow() = %v, want %v (existing predicate keeps its answers)", gotFlow, tt.wantGitFlowOnly)
			}
		})
	}
}

// TestAgenticTrunkDevelopBranchProjection pins the develop-branch seam
// projection the three legacy develop_branch consumers sit on
// (REQ-AGT-025 / AC-AGT-025 Judge 1): manual agentic-trunk resolves main
// whether or not the key is present (it is ignored); personal mode has an
// empty seam.
func TestAgenticTrunkDevelopBranchProjection(t *testing.T) {
	t.Run("key absent", func(t *testing.T) {
		root := t.TempDir()
		dispositionWrite(t, root, gitStrategyBody("manual", WorkflowAgenticTrunk, ""))
		got := LoadGitFlowIntegrationConfig(root)
		if got.DevelopBranch != "main" {
			t.Errorf("DevelopBranch = %q, want main with no develop_branch key", got.DevelopBranch)
		}
		if LoadGitFlowDevelopBranch(root) != "main" {
			t.Errorf("LoadGitFlowDevelopBranch = %q, want main", LoadGitFlowDevelopBranch(root))
		}
	})

	t.Run("key present as develop is ignored", func(t *testing.T) {
		root := t.TempDir()
		dispositionWrite(t, root, gitStrategyBody("manual", WorkflowAgenticTrunk, "        develop_branch: develop\n"))
		got := LoadGitFlowIntegrationConfig(root)
		if got.DevelopBranch != "main" {
			t.Errorf("DevelopBranch = %q, want main (the develop_branch key is ignored under agentic-trunk)", got.DevelopBranch)
		}
	})

	t.Run("personal mode empty seam", func(t *testing.T) {
		root := t.TempDir()
		dispositionWrite(t, root, gitStrategyBody("personal", WorkflowAgenticTrunk, ""))
		got := LoadGitFlowIntegrationConfig(root)
		if got.DevelopBranch != "" {
			t.Errorf("DevelopBranch = %q, want empty in personal mode", got.DevelopBranch)
		}
	})

	t.Run("git-flow projection unchanged", func(t *testing.T) {
		root := t.TempDir()
		dispositionWrite(t, root, gitStrategyBody("manual", WorkflowGitFlow, "        develop_branch: develop\n"))
		got := LoadGitFlowIntegrationConfig(root)
		if got.DevelopBranch != "develop" {
			t.Errorf("DevelopBranch = %q, want develop under git-flow (unchanged)", got.DevelopBranch)
		}
	})
}

// TestAgenticTrunkGuidance pins the diagnostic surfaces (REQ-AGT-009 /
// AC-AGT-008): the invalid-value allowed-set listing carries all five
// values, and every git-flow fix suggestion that names
// `github-flow (target main)` also names `agentic-trunk (target main)`.
func TestAgenticTrunkGuidance(t *testing.T) {
	t.Run("invalid-value listing carries all five values", func(t *testing.T) {
		root := t.TempDir()
		dispositionWrite(t, root, gitStrategyBody("manual", "svn-flow", ""))
		got := LoadGitFlowIntegrationConfig(root)
		msg := got.EmptyTargetGuidance(root, "")
		if want := strings.Join(AllowedWorkflows(), ", "); !strings.Contains(msg, want) {
			t.Errorf("invalid-value guidance %q does not list the allowed set %q", msg, want)
		}
	})

	t.Run("git-flow outside manual mode fix names both flows", func(t *testing.T) {
		root := t.TempDir()
		dispositionWrite(t, root, "git_strategy:\n    mode: personal\n    personal:\n        workflow: git-flow\n        develop_branch: develop\n")
		got := LoadGitFlowIntegrationConfig(root)
		msg := got.EmptyTargetGuidance(root, "")
		if !strings.Contains(msg, "github-flow (target main)") {
			t.Errorf("fix no longer names github-flow (target main): %q", msg)
		}
		if !strings.Contains(msg, "agentic-trunk (target main)") {
			t.Errorf("fix does not name agentic-trunk (target main): %q", msg)
		}
	})

	t.Run("empty develop_branch fix names both flows", func(t *testing.T) {
		root := t.TempDir()
		dispositionWrite(t, root, gitStrategyBody("manual", WorkflowGitFlow, "        develop_branch: \"\"\n"))
		got := LoadGitFlowIntegrationConfig(root)
		msg := got.EmptyTargetGuidance(root, "")
		if !strings.Contains(msg, "github-flow (target main)") {
			t.Errorf("fix no longer names github-flow (target main): %q", msg)
		}
		if !strings.Contains(msg, "agentic-trunk (target main)") {
			t.Errorf("fix does not name agentic-trunk (target main): %q", msg)
		}
	})
}

// TestAgenticTrunkDistributedDefaultsUnchanged pins the shipped defaults
// (REQ-AGT-015 / AC-AGT-014 Judge A): the built-in default configuration and
// the template yaml carry github-flow for all three profiles, and neither
// carries a workflow: agentic-trunk line — nothing deployed to user projects
// changes a default.
func TestAgenticTrunkDistributedDefaultsUnchanged(t *testing.T) {
	t.Run("built-in defaults are github-flow in all three profiles", func(t *testing.T) {
		gs := NewDefaultGitStrategyConfig()
		for name, got := range map[string]string{
			"manual":   gs.Manual.Workflow,
			"personal": gs.Personal.Workflow,
			"team":     gs.Team.Workflow,
		} {
			if got != WorkflowGitHubFlow {
				t.Errorf("built-in default %s.workflow = %q, want %q", name, got, WorkflowGitHubFlow)
			}
		}
	})

	t.Run("template yaml is github-flow in all three profiles", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join("..", "template", "templates", ".moai", "config", "sections", "git-strategy.yaml.tmpl"))
		if err != nil {
			t.Fatalf("read the template git-strategy.yaml.tmpl: %v", err)
		}
		body := string(data)
		if got := strings.Count(body, "workflow: github-flow"); got < 3 {
			t.Errorf("template carries %d github-flow lines, want at least 3", got)
		}
		if strings.Contains(body, "workflow: agentic-trunk") {
			t.Error("template yaml carries workflow: agentic-trunk — the shipped default must never change")
		}
	})

	t.Run("built-in defaults name no agentic-trunk anywhere", func(t *testing.T) {
		gs := NewDefaultGitStrategyConfig()
		if gs.Manual.Workflow == WorkflowAgenticTrunk || gs.Personal.Workflow == WorkflowAgenticTrunk || gs.Team.Workflow == WorkflowAgenticTrunk {
			t.Error("a built-in default profile carries agentic-trunk")
		}
	})
}
