package cli

// doctor_git_strategy_target_test.go — SPEC-GITHUB-FLOW-DEFAULT-001 M1b (card
// t1453). After M1 the re-pointed consumers REFUSE an empty integration target
// instead of falling back to the caller's branch, so the Git Strategy Workflow
// check may no longer report `ok` for it. It also warns when the integration
// target differs from worktree_base_branch — the signature of a primary config
// revert — but only when both are non-empty.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

func TestDoctorGitStrategyEmptyTargetWarns(t *testing.T) {
	cases := []struct {
		name string
		body string
		keys []string
	}{
		{"git-flow empty develop_branch",
			"git_strategy:\n    mode: manual\n    manual:\n        workflow: git-flow\n        develop_branch: \"\"\n",
			[]string{"git_strategy.manual.develop_branch"}},
		{"git-flow key absent",
			"git_strategy:\n    mode: manual\n    manual:\n        workflow: git-flow\n",
			[]string{"git_strategy.manual.develop_branch"}},
		{"git-flow outside manual mode",
			"git_strategy:\n    mode: personal\n    personal:\n        workflow: git-flow\n        develop_branch: develop\n",
			[]string{"git_strategy.mode"}},
		{"gitlab-flow empty environment",
			"git_strategy:\n    mode: manual\n    manual:\n        workflow: gitlab-flow\n        environment: \"\"\n",
			[]string{"git_strategy.manual.environment"}},
		{"release-flow empty prefix",
			"git_strategy:\n    mode: manual\n    manual:\n        workflow: release-flow\n        release_branch_prefix: \"\"\n",
			[]string{"git_strategy.manual.release_branch_prefix"}},
	}
	executed := 0
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			executed++
			root := t.TempDir()
			writeGitStrategyBody(t, root, tc.body)
			check := checkGitStrategyWorkflow(root, false)
			if check.Status != uikit.CheckWarn {
				t.Errorf("status = %v, want WARN for an empty integration target; message %q", check.Status, check.Message)
			}
			for _, k := range tc.keys {
				if !strings.Contains(check.Message, k) {
					t.Errorf("message %q does not name %q", check.Message, k)
				}
			}
			if strings.Contains(check.Message, "fall back to the caller's branch") {
				t.Errorf("message %q still claims a caller fallback the consumers no longer have", check.Message)
			}
		})
	}
	if executed == 0 {
		t.Fatal("empty sweep: no case executed")
	}
}

func TestDoctorGitStrategyTargetVersusWorktreeBase(t *testing.T) {
	const gitFlow = "git_strategy:\n    mode: manual\n    manual:\n        workflow: git-flow\n        develop_branch: develop\n"
	const githubFlow = "git_strategy:\n    mode: manual\n    manual:\n        workflow: github-flow\n"

	t.Run("target differs from worktree_base_branch warns naming both", func(t *testing.T) {
		root := t.TempDir()
		writeGitStrategyBody(t, root, githubFlow+"    worktree_base_branch: develop\n")
		check := checkGitStrategyWorkflow(root, false)
		if check.Status != uikit.CheckWarn {
			t.Fatalf("status = %v, want WARN on a target/worktree_base_branch mismatch; message %q", check.Status, check.Message)
		}
		for _, want := range []string{"worktree_base_branch", "develop", "main"} {
			if !strings.Contains(check.Message, want) {
				t.Errorf("message %q does not name %q", check.Message, want)
			}
		}
	})

	t.Run("matching target and worktree_base_branch is the unchanged healthy output", func(t *testing.T) {
		plain := t.TempDir()
		writeGitStrategyBody(t, plain, gitFlow)
		want := checkGitStrategyWorkflow(plain, false)

		root := t.TempDir()
		writeGitStrategyBody(t, root, gitFlow+"    worktree_base_branch: develop\n")
		got := checkGitStrategyWorkflow(root, false)
		if got.Status != uikit.CheckOK || got.Message != want.Message {
			t.Errorf("healthy output changed: status %v message %q, want OK %q", got.Status, got.Message, want.Message)
		}
	})

	t.Run("an unset worktree_base_branch never warns", func(t *testing.T) {
		root := t.TempDir()
		writeGitStrategyBody(t, root, githubFlow)
		if check := checkGitStrategyWorkflow(root, false); check.Status != uikit.CheckOK {
			t.Errorf("status = %v, want OK with no worktree_base_branch; message %q", check.Status, check.Message)
		}
	})

	t.Run("an empty target is the empty-target warning, not a mismatch", func(t *testing.T) {
		root := t.TempDir()
		writeGitStrategyBody(t, root,
			"git_strategy:\n    mode: manual\n    manual:\n        workflow: git-flow\n        develop_branch: \"\"\n    worktree_base_branch: develop\n")
		check := checkGitStrategyWorkflow(root, false)
		if check.Status != uikit.CheckWarn || !strings.Contains(check.Message, "git_strategy.manual.develop_branch") {
			t.Errorf("status %v message %q, want the develop_branch warning", check.Status, check.Message)
		}
	})
}
