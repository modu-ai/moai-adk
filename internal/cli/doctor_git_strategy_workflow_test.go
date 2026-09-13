package cli

// doctor_git_strategy_workflow_test.go — SPEC-GITSTRAT-WORKFLOW-READER-001
// M3 (card t656). The doctor check is the interpretation table's production
// consumer (REQ-GWS-009): every state names what was read — no silent pass.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// writeGitStrategyWorkflowFixture writes a minimal manual-mode
// git-strategy.yaml naming the given workflow value.
func writeGitStrategyWorkflowFixture(t *testing.T, root, workflow string) {
	t.Helper()
	writeGitStrategyBody(t, root,
		"git_strategy:\n    mode: manual\n    manual:\n        workflow: "+workflow+"\n        develop_branch: develop\n")
}

// TestDoctorGitStrategyWorkflow pins the three reportable states over real
// fixtures (AC-GWS-013): invalid → WARN naming the offending value and all
// 4 allowed entries; git-flow → OK naming the flow and resolved target;
// github-flow (valid non-git-flow) → OK naming the flow, "main", and the
// standing-branch interpretation.
func TestDoctorGitStrategyWorkflow(t *testing.T) {
	t.Run("invalid warns naming value and allowed set", func(t *testing.T) {
		root := t.TempDir()
		writeGitStrategyWorkflowFixture(t, root, "git-flwo")

		check := checkGitStrategyWorkflow(root, false)
		if check.Status != uikit.CheckWarn {
			t.Errorf("status = %v, want WARN on an invalid workflow value", check.Status)
		}
		if !strings.Contains(check.Message, "git-flwo") {
			t.Errorf("message %q does not name the offending value", check.Message)
		}
		for _, allowed := range []string{"github-flow", "git-flow", "gitlab-flow", "release-flow"} {
			if !strings.Contains(check.Message, allowed) {
				t.Errorf("message %q does not name allowed entry %q", check.Message, allowed)
			}
		}
		if !strings.Contains(check.Message, "git-strategy.yaml") {
			t.Errorf("message %q does not name the repair surface", check.Message)
		}
	})

	t.Run("git-flow reports the resolved target", func(t *testing.T) {
		root := t.TempDir()
		writeGitStrategyWorkflowFixture(t, root, "git-flow")

		check := checkGitStrategyWorkflow(root, false)
		if check.Status != uikit.CheckOK {
			t.Errorf("status = %v, want OK on git-flow", check.Status)
		}
		if !strings.Contains(check.Message, "git-flow") {
			t.Errorf("message %q does not name the flow", check.Message)
		}
		if !strings.Contains(check.Message, "develop") {
			t.Errorf("message %q does not name the resolved target", check.Message)
		}
	})

	t.Run("github-flow reports flow, main, and standing branches", func(t *testing.T) {
		root := t.TempDir()
		writeGitStrategyWorkflowFixture(t, root, "github-flow")

		check := checkGitStrategyWorkflow(root, false)
		if check.Status != uikit.CheckOK {
			t.Errorf("status = %v, want OK on github-flow", check.Status)
		}
		if !strings.Contains(check.Message, "github-flow") {
			t.Errorf("message %q does not name the flow", check.Message)
		}
		if !strings.Contains(check.Message, "main") {
			t.Errorf("message %q does not name the resolved target main", check.Message)
		}
		if !strings.Contains(strings.ToLower(check.Message), "standing") {
			t.Errorf("message %q does not carry the standing-branch interpretation", check.Message)
		}
	})

	t.Run("unreadable file names what was read, not a silent pass", func(t *testing.T) {
		root := t.TempDir() // no git-strategy.yaml at all

		check := checkGitStrategyWorkflow(root, false)
		if !strings.Contains(check.Message, "workflow") {
			t.Errorf("message %q does not name what was read", check.Message)
		}
	})

	// AC-GWS-007: an empty environment key under gitlab-flow resolves the
	// empty caller-fallback neutral; the check still reports OK with the
	// fallback named.
	t.Run("gitlab-flow empty environment names the fallback", func(t *testing.T) {
		root := t.TempDir()
		writeGitStrategyBody(t, root,
			"git_strategy:\n    mode: manual\n    manual:\n        workflow: gitlab-flow\n        environment: \"\"\n")

		check := checkGitStrategyWorkflow(root, false)
		if check.Status != uikit.CheckOK {
			t.Errorf("status = %v, want OK on gitlab-flow with an empty environment key", check.Status)
		}
		if !strings.Contains(check.Message, "gitlab-flow") {
			t.Errorf("message %q does not name the flow", check.Message)
		}
	})

	// The remaining flow arm: release-flow resolves release_branch_prefix.
	t.Run("release-flow resolves the prefix", func(t *testing.T) {
		root := t.TempDir()
		writeGitStrategyBody(t, root,
			"git_strategy:\n    mode: manual\n    manual:\n        workflow: release-flow\n        release_branch_prefix: release/\n")

		check := checkGitStrategyWorkflow(root, false)
		if check.Status != uikit.CheckOK {
			t.Errorf("status = %v, want OK on release-flow", check.Status)
		}
		if !strings.Contains(check.Message, "release-flow") || !strings.Contains(check.Message, "release/") {
			t.Errorf("message %q does not name the flow and resolved target", check.Message)
		}
	})

	// Verbose mode carries Detail on every state, and the R2 discretion note
	// flags a shipped environment-label default as a non-branch target.
	t.Run("verbose detail names what was read", func(t *testing.T) {
		root := t.TempDir()
		writeGitStrategyWorkflowFixture(t, root, "git-flow")
		if check := checkGitStrategyWorkflow(root, true); check.Detail == "" {
			t.Error("verbose git-flow check carries no Detail")
		}

		root = t.TempDir()
		writeGitStrategyBody(t, root,
			"git_strategy:\n    mode: manual\n    manual:\n        workflow: gitlab-flow\n        environment: local\n")
		check := checkGitStrategyWorkflow(root, true)
		if !strings.Contains(check.Detail, "local") || !strings.Contains(check.Detail, "not a branch name") {
			t.Errorf("verbose gitlab-flow Detail %q does not flag the shipped label default", check.Detail)
		}

		root = t.TempDir()
		writeGitStrategyWorkflowFixture(t, root, "git-flwo")
		if check := checkGitStrategyWorkflow(root, true); !strings.Contains(check.Detail, "trunk-based") {
			t.Errorf("verbose invalid Detail %q does not note the trunk-based exclusion", check.Detail)
		}

		root = t.TempDir() // unreadable: verbose Detail names the setup step
		if check := checkGitStrategyWorkflow(root, true); check.Detail == "" {
			t.Error("verbose unknown-state check carries no Detail")
		}
	})
}
