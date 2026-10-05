package config

// loader_integration_branch_test.go — card t449.
//
// `moai integration acquire` must default its recorded branch to the project's
// configured git-flow develop branch, never to a hardcoded name and never to
// the caller's checked-out branch. These tests pin the gate: the key flows
// through ONLY when the active mode profile's workflow is git-flow, and every
// other shape — other workflow, other mode, missing key, missing file — yields
// the neutral empty string.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGitFlowDevelopBranch(t *testing.T) {
	write := func(t *testing.T, root, body string) {
		t.Helper()
		if body == "" {
			return // no file at all: the missing-file case
		}
		dir := filepath.Join(root, ".moai", "config", "sections")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	const gitFlowBody = "git_strategy:\n" +
		"    mode: manual\n" +
		"    manual:\n" +
		"        workflow: git-flow\n" +
		"        develop_branch: fixture-integration\n"

	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "git-flow develop branch flows through",
			body: gitFlowBody,
			want: "fixture-integration",
		},
		{
			name: "develop_branch under github-flow is not an integration branch",
			body: "git_strategy:\n" +
				"    mode: manual\n" +
				"    manual:\n" +
				"        workflow: github-flow\n" +
				"        develop_branch: fixture-integration\n",
			want: "",
		},
		{
			name: "a git-flow-shaped personal profile does not qualify",
			body: "git_strategy:\n" +
				"    mode: personal\n" +
				"    personal:\n" +
				"        workflow: git-flow\n" +
				"        develop_branch: fixture-integration\n",
			want: "",
		},
		{
			name: "empty develop_branch yields the neutral value",
			body: "git_strategy:\n" +
				"    mode: manual\n" +
				"    manual:\n" +
				"        workflow: git-flow\n" +
				"        develop_branch: \"\"\n",
			want: "",
		},
		{
			name: "missing mode has no active profile",
			body: "git_strategy:\n" +
				"    manual:\n" +
				"        workflow: git-flow\n" +
				"        develop_branch: fixture-integration\n",
			want: "",
		},
		{
			name: "missing file yields the neutral value",
			body: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			write(t, root, tt.body)
			if got := LoadGitFlowDevelopBranch(root); got != tt.want {
				t.Errorf("LoadGitFlowDevelopBranch() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestLoadGitFlowIntegrationConfig_SeparatesGitFlowFromBranch — card t637.
//
// LoadGitFlowDevelopBranch answers "" for three cases the acquire warning must
// tell apart: not git-flow, git-flow with an empty develop_branch, and no
// readable file. The seam reports "is this project git-flow" separately from
// the branch value, so an empty value under git-flow is distinguishable from
// an empty value everywhere else.
func TestLoadGitFlowIntegrationConfig_SeparatesGitFlowFromBranch(t *testing.T) {
	write := func(t *testing.T, root, body string) {
		t.Helper()
		if body == "" {
			return // no file at all
		}
		dir := filepath.Join(root, ".moai", "config", "sections")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name        string
		body        string
		wantGitFlow bool
		wantBranch  string
	}{
		{
			name:        "a manual git-flow with develop set",
			body:        "git_strategy:\n    mode: manual\n    manual:\n        workflow: git-flow\n        develop_branch: fixture-integration\n",
			wantGitFlow: true,
			wantBranch:  "fixture-integration",
		},
		{
			name:        "b manual git-flow with develop empty",
			body:        "git_strategy:\n    mode: manual\n    manual:\n        workflow: git-flow\n        develop_branch: \"\"\n",
			wantGitFlow: true,
			wantBranch:  "",
		},
		{
			name:        "c manual github-flow",
			body:        "git_strategy:\n    mode: manual\n    manual:\n        workflow: github-flow\n        develop_branch: fixture-integration\n",
			wantGitFlow: false,
			wantBranch:  "",
		},
		{
			name:        "d personal git-flow",
			body:        "git_strategy:\n    mode: personal\n    personal:\n        workflow: git-flow\n        develop_branch: fixture-integration\n",
			wantGitFlow: false,
			wantBranch:  "",
		},
		{
			name:        "e no file",
			body:        "",
			wantGitFlow: false,
			wantBranch:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			write(t, root, tt.body)
			got := LoadGitFlowIntegrationConfig(root)
			if got.IsGitFlow() != tt.wantGitFlow {
				t.Errorf("IsGitFlow() = %v, want %v (config %+v)", got.IsGitFlow(), tt.wantGitFlow, got)
			}
			if got.DevelopBranch != tt.wantBranch {
				t.Errorf("DevelopBranch = %q, want %q", got.DevelopBranch, tt.wantBranch)
			}
			// The seam must never disagree with the pre-existing reader.
			if legacy := LoadGitFlowDevelopBranch(root); legacy != got.DevelopBranch {
				t.Errorf("LoadGitFlowDevelopBranch() = %q, seam DevelopBranch = %q; the two readers disagree", legacy, got.DevelopBranch)
			}
		})
	}
}

// --- Characterization suite (SPEC-GITSTRAT-WORKFLOW-READER-001 M1, card t656) ---
//
// These tests pin the CURRENT behavior of the workflow reader before any
// production extension lands (REQ-GWS-007). They are written and committed
// against the unmodified tree; every later milestone must keep them passing
// byte-unmodified (AC-GWS-010 diffs this file against the M1 commit SHA).

// characterizeWrite is the fixture writer shared by the characterization
// tests. An empty body means "write no file at all" (the missing-file case).
func characterizeWrite(t *testing.T, root, body string) {
	t.Helper()
	if body == "" {
		return
	}
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestCharacterize_LoadGitFlowIntegrationConfig_FailurePaths pins the zero /
// false halves the reader yields on every failure path: missing file,
// unparseable YAML, non-manual mode, non-git-flow workflow, no active
// profile, and an empty or whitespace-only develop_branch.
func TestCharacterize_LoadGitFlowIntegrationConfig_FailurePaths(t *testing.T) {
	tests := []struct {
		name            string
		body            string
		wantManual      bool
		wantGitFlow     bool
		wantBranch      string
		wantZeroAllHalf bool // Manual, GitFlowWorkflow and DevelopBranch all zero
	}{
		{
			name:            "missing file yields the zero value",
			body:            "",
			wantZeroAllHalf: true,
		},
		{
			name:            "unparseable yaml yields the zero value",
			body:            "git_strategy: [broken\n    mode: !!@%",
			wantZeroAllHalf: true,
		},
		{
			name:        "team mode with a git-flow workflow does not qualify",
			body:        "git_strategy:\n    mode: team\n    team:\n        workflow: git-flow\n        develop_branch: fixture-integration\n",
			wantManual:  false,
			wantGitFlow: true,
			wantBranch:  "",
		},
		{
			name:        "github-flow workflow is not git-flow",
			body:        "git_strategy:\n    mode: manual\n    manual:\n        workflow: github-flow\n        develop_branch: fixture-integration\n",
			wantManual:  true,
			wantGitFlow: false,
			wantBranch:  "",
		},
		{
			name:        "gitlab-flow workflow is not git-flow",
			body:        "git_strategy:\n    mode: manual\n    manual:\n        workflow: gitlab-flow\n",
			wantManual:  true,
			wantGitFlow: false,
			wantBranch:  "",
		},
		{
			name:        "missing mode has no active profile",
			body:        "git_strategy:\n    manual:\n        workflow: git-flow\n        develop_branch: fixture-integration\n",
			wantManual:  false,
			wantGitFlow: false,
			wantBranch:  "",
		},
		{
			name:        "empty develop_branch keeps the git-flow halves",
			body:        "git_strategy:\n    mode: manual\n    manual:\n        workflow: git-flow\n        develop_branch: \"\"\n",
			wantManual:  true,
			wantGitFlow: true,
			wantBranch:  "",
		},
		{
			name:        "whitespace-only develop_branch trims to empty",
			body:        "git_strategy:\n    mode: manual\n    manual:\n        workflow: git-flow\n        develop_branch: \"   \"\n",
			wantManual:  true,
			wantGitFlow: true,
			wantBranch:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			characterizeWrite(t, root, tt.body)
			got := LoadGitFlowIntegrationConfig(root)
			if got.Manual != tt.wantManual {
				t.Errorf("Manual = %v, want %v", got.Manual, tt.wantManual)
			}
			if got.GitFlowWorkflow != tt.wantGitFlow {
				t.Errorf("GitFlowWorkflow = %v, want %v", got.GitFlowWorkflow, tt.wantGitFlow)
			}
			if got.DevelopBranch != tt.wantBranch {
				t.Errorf("DevelopBranch = %q, want %q", got.DevelopBranch, tt.wantBranch)
			}
			if tt.wantZeroAllHalf && (got.Manual || got.GitFlowWorkflow || got.DevelopBranch != "") {
				t.Errorf("expected the all-zero failure value, got %+v", got)
			}
		})
	}
}

// TestCharacterize_LoadGitFlowIntegrationConfig_SuccessPath pins the one
// success shape: manual mode + git-flow workflow + a develop branch flows
// through as the full struct.
func TestCharacterize_LoadGitFlowIntegrationConfig_SuccessPath(t *testing.T) {
	root := t.TempDir()
	characterizeWrite(t, root,
		"git_strategy:\n    mode: manual\n    manual:\n        workflow: git-flow\n        develop_branch: fixture-integration\n")
	got := LoadGitFlowIntegrationConfig(root)
	if !got.Manual || !got.GitFlowWorkflow {
		t.Fatalf("expected both halves true, got %+v", got)
	}
	if got.DevelopBranch != "fixture-integration" {
		t.Errorf("DevelopBranch = %q, want fixture-integration", got.DevelopBranch)
	}
	if !got.IsGitFlow() {
		t.Error("IsGitFlow() = false on the manual git-flow success shape")
	}
}

// TestCharacterize_IsGitFlow_TruthTable pins the AND semantics of the
// predicate across all four halves combinations.
func TestCharacterize_IsGitFlow_TruthTable(t *testing.T) {
	tests := []struct {
		name            string
		manual          bool
		gitFlowWorkflow bool
		want            bool
	}{
		{name: "neither half", manual: false, gitFlowWorkflow: false, want: false},
		{name: "manual only", manual: true, gitFlowWorkflow: false, want: false},
		{name: "workflow only", manual: false, gitFlowWorkflow: true, want: false},
		{name: "both halves", manual: true, gitFlowWorkflow: true, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GitFlowIntegrationConfig{Manual: tt.manual, GitFlowWorkflow: tt.gitFlowWorkflow}
			if got.IsGitFlow() != tt.want {
				t.Errorf("IsGitFlow() with (Manual=%v, GitFlowWorkflow=%v) = %v, want %v",
					tt.manual, tt.gitFlowWorkflow, got.IsGitFlow(), tt.want)
			}
		})
	}
}

// TestCharacterize_LoadGitFlowDevelopBranch_Trims pins the trim contract: a
// develop_branch padded with whitespace is returned trimmed, and the
// delegation to the seam is exact.
func TestCharacterize_LoadGitFlowDevelopBranch_Trims(t *testing.T) {
	root := t.TempDir()
	characterizeWrite(t, root,
		"git_strategy:\n    mode: manual\n    manual:\n        workflow: git-flow\n        develop_branch: \"  padded-integration  \\t\"\n")
	want := "padded-integration"
	if got := LoadGitFlowDevelopBranch(root); got != want {
		t.Errorf("LoadGitFlowDevelopBranch() = %q, want %q (trimmed)", got, want)
	}
	seam := LoadGitFlowIntegrationConfig(root)
	if seam.DevelopBranch != want {
		t.Errorf("seam DevelopBranch = %q, want %q; the two readers disagree", seam.DevelopBranch, want)
	}
}
