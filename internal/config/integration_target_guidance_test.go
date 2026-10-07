package config

// integration_target_guidance_test.go — SPEC-GITHUB-FLOW-DEFAULT-001 M1b (card
// t1453). EmptyTargetCause / EmptyTargetGuidance / TargetProvenance classify
// the loader state the integration-target reader already holds: every cause
// that leaves the target empty yields a reason naming the key or file to fix,
// the site's tail is appended verbatim, and a resolved target yields nothing.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeGuidanceFixture(t *testing.T, body string, absent bool) string {
	t.Helper()
	root := t.TempDir()
	if absent {
		return root
	}
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestEmptyTargetGuidance(t *testing.T) {
	const tail = "pass --base origin/<branch>"
	manual := func(profile string) string {
		return "git_strategy:\n    mode: manual\n    manual:\n" + profile
	}
	cases := []struct {
		name   string
		body   string
		absent bool
		cause  EmptyTargetCause
		wants  []string
	}{
		{"git-flow develop_branch empty", manual("        workflow: git-flow\n        develop_branch: \"\"\n"), false,
			EmptyTargetDevelopEmpty,
			[]string{"git_strategy.manual.develop_branch is empty; set it (for example `develop`) in .moai/config/sections/git-strategy.yaml, or set git_strategy.manual.workflow: github-flow (target main)"}},
		{"git-flow develop_branch key absent", manual("        workflow: git-flow\n"), false,
			EmptyTargetDevelopEmpty,
			[]string{"git_strategy.manual.develop_branch is empty"}},
		{"file absent", "", true,
			EmptyTargetUnreadable,
			[]string{"cannot read", filepath.Join(".moai", "config", "sections", "git-strategy.yaml"), "(absent or not valid YAML)"}},
		{"file unparseable", "git_strategy: [unterminated\n", false,
			EmptyTargetUnreadable,
			[]string{"cannot read", "(absent or not valid YAML)"}},
		{"git-flow outside manual mode", "git_strategy:\n    mode: personal\n    personal:\n        workflow: git-flow\n        develop_branch: develop\n", false,
			EmptyTargetGitFlowNotManual,
			[]string{"git-flow resolves a target only when git_strategy.mode is manual (mode is personal)", "git_strategy.personal.workflow: github-flow"}},
		{"unknown workflow", manual("        workflow: svn-flow\n"), false,
			EmptyTargetUnknownWorkflow,
			[]string{`git_strategy.manual.workflow "svn-flow" is not one of github-flow, git-flow, gitlab-flow, release-flow`}},
		{"empty workflow value", manual("        workflow: \"\"\n"), false,
			EmptyTargetUnknownWorkflow,
			[]string{`git_strategy.manual.workflow "" is not one of`}},
		{"unknown mode", "git_strategy:\n    mode: teem\n", false,
			EmptyTargetUnknownMode,
			[]string{`git_strategy.mode "teem" is not one of manual, personal, team`}},
		{"gitlab-flow environment empty", manual("        workflow: gitlab-flow\n        environment: \"\"\n"), false,
			EmptyTargetEnvironmentEmpty,
			[]string{"git_strategy.manual.environment is empty"}},
		{"release-flow prefix empty", manual("        workflow: release-flow\n        release_branch_prefix: \"\"\n"), false,
			EmptyTargetReleasePrefixEmpty,
			[]string{"git_strategy.manual.release_branch_prefix is empty"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := writeGuidanceFixture(t, tc.body, tc.absent)
			cfg := LoadGitFlowIntegrationConfig(root)
			if got := cfg.EmptyTargetCause(); got != tc.cause {
				t.Fatalf("EmptyTargetCause = %q, want %q", got, tc.cause)
			}
			msg := cfg.EmptyTargetGuidance(root, tail)
			for _, w := range tc.wants {
				if !strings.Contains(msg, w) {
					t.Errorf("guidance %q does not contain %q", msg, w)
				}
			}
			if !strings.HasSuffix(msg, ", or "+tail) {
				t.Errorf("guidance %q does not end in the site tail %q", msg, tail)
			}
			if bare := cfg.EmptyTargetGuidance(root, ""); bare == "" || strings.Contains(bare, tail) || strings.HasSuffix(bare, ", or ") {
				t.Errorf("an empty tail must append nothing, got %q", bare)
			}
		})
	}
}

func TestEmptyTargetGuidanceResolvedTargetIsSilent(t *testing.T) {
	for name, body := range map[string]string{
		"github-flow": "git_strategy:\n    mode: manual\n    manual:\n        workflow: github-flow\n",
		"git-flow":    "git_strategy:\n    mode: manual\n    manual:\n        workflow: git-flow\n        develop_branch: develop\n",
	} {
		t.Run(name, func(t *testing.T) {
			root := writeGuidanceFixture(t, body, false)
			cfg := LoadGitFlowIntegrationConfig(root)
			if cfg.IntegrationTarget == "" {
				t.Fatal("fixture must resolve a target")
			}
			if cause := cfg.EmptyTargetCause(); cause != "" {
				t.Errorf("EmptyTargetCause = %q, want none", cause)
			}
			if msg := cfg.EmptyTargetGuidance(root, "x"); msg != "" {
				t.Errorf("a resolved target must yield no guidance, got %q", msg)
			}
		})
	}
}

func TestTargetProvenance(t *testing.T) {
	root := writeGuidanceFixture(t, "git_strategy:\n    mode: manual\n    manual:\n        workflow: git-flow\n        develop_branch: develop\n", false)
	if got, want := LoadGitFlowIntegrationConfig(root).TargetProvenance(), "git_strategy.manual.workflow=git-flow"; got != want {
		t.Errorf("TargetProvenance = %q, want %q", got, want)
	}
	if got := LoadGitFlowIntegrationConfig(writeGuidanceFixture(t, "", true)).TargetProvenance(); got != "" {
		t.Errorf("an unreadable file has no provenance, got %q", got)
	}
}
