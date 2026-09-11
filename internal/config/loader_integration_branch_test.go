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
