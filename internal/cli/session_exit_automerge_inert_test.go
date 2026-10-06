package cli

// session_exit_automerge_inert_test.go — SPEC-GITHUB-FLOW-DEFAULT-001 M1
// (card t1453), AC-GFD-001 guard row.
//
// The session-exit auto-merge reads the git-flow develop branch on purpose
// (design D-1): it is a git-flow-only path that must stay dormant under every
// other workflow. Moving that read onto the integration-target interpreter
// would turn a local `git merge` into main on under github-flow. This guard
// runs the real loader against real configuration files — not a stubbed
// GitFlowIntegrationConfig — and asserts the dormant flows execute no git
// command and write no window record, with the git-flow control proving the
// same harness reaches the merge when the configuration allows it.

import (
	"bytes"
	"strings"
	"testing"
)

// TestSessionExitAutoMergeInertUnderGitHubFlow: with auto_merge on and a clean
// exit, the session-exit path stays inert under github-flow (and the other
// non-git-flow workflows), and runs the merge under git-flow.
func TestSessionExitAutoMergeInertUnderGitHubFlow(t *testing.T) {
	inert := []struct {
		name string
		body string
	}{
		{"github-flow", "git_strategy:\n    mode: manual\n    manual:\n        workflow: github-flow\n"},
		{"gitlab-flow", "git_strategy:\n    mode: manual\n    manual:\n        workflow: gitlab-flow\n        environment: production\n"},
		{"release-flow", "git_strategy:\n    mode: manual\n    manual:\n        workflow: release-flow\n        release_branch_prefix: release/\n"},
	}
	for _, tc := range inert {
		t.Run(tc.name+" runs no merge and writes no window record", func(t *testing.T) {
			root := t.TempDir()
			writeGitStrategyBody(t, root, tc.body)

			rec := &amRecorder{}
			seams := baseAMSeams(rec)
			seams.lockRoot = func() string { return root }
			seams.gitFlow = nil // the REAL loader reads the fixture above
			swapAutoMergeSeams(t, seams)
			swapCleanStatusSeams(t)
			out := &bytes.Buffer{}

			sessionExitAutoMerge(amCfg(true, false), amSessionWt, true, out)

			if len(rec.mergeBranches) != 0 {
				t.Errorf("a non-git-flow workflow must not run `git merge`, merged: %v", rec.mergeBranches)
			}
			if len(rec.gitLog) != 0 {
				t.Errorf("no git invocation may run under a non-git-flow workflow, got: %v", rec.gitLog)
			}
			if len(rec.lockLog) != 0 {
				t.Errorf("no integration-window record may be written, got: %v", rec.lockLog)
			}
			if !strings.Contains(out.String(), "no integration target") {
				t.Errorf("the skip must be announced as a notice, got: %q", out.String())
			}
		})
	}

	t.Run("control: git-flow with develop reaches the merge", func(t *testing.T) {
		root := t.TempDir()
		writeGitStrategyFixture(t, root, "git-flow", "develop")

		rec := &amRecorder{}
		seams := baseAMSeams(rec)
		seams.lockRoot = func() string { return root }
		seams.gitFlow = nil
		swapAutoMergeSeams(t, seams)
		swapCleanStatusSeams(t)
		out := &bytes.Buffer{}

		sessionExitAutoMerge(amCfg(true, false), amSessionWt, true, out)

		if len(rec.mergeBranches) != 1 {
			t.Fatalf("the git-flow control must merge once (an inert result above would otherwise prove nothing), merged: %v; notice: %q",
				rec.mergeBranches, out.String())
		}
	})
}
