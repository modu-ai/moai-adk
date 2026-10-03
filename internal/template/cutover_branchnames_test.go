// cutover_branchnames_test.go: the cutover scripts carry no default branch name
// (SPEC-GITHUB-FLOW-DEFAULT-001, card t1453, leader decision D17 option 2).
//
// The retiring integration branch and the target base branch are supplied by the
// caller (the runbook, or a test fixture). A script run without them must stop
// with a usage error (exit 2) that names the missing argument, before it reads a
// repository, calls a reader or creates a directory. A script that fell back to a
// built-in name would pass every other test (the fixtures supply the names) and
// would put the literal back into the tree, so this is the guard that keeps the
// absence of a default observable.
package template_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// cvoWantUsage asserts a usage error (exit 2) whose output names the missing flag
// and says it is required. "is required" separates a missing value from an unknown
// flag, whose message also names the flag.
func cvoWantUsage(t *testing.T, res rlsResult, flag string) {
	t.Helper()
	if res.exit != 2 {
		t.Errorf("exit code = %d, want 2 (usage error naming %s)\n%s", res.exit, flag, rlsNorm(res.out))
	}
	rlsMustContain(t, rlsNorm(res.out), flag+" is required")
}

// TestCutoverScriptsRequireBranchNames: no script has a default branch name.
func TestCutoverScriptsRequireBranchNames(t *testing.T) {
	rlsRequireTools(t)

	t.Run("precheck_without_a_retiring_branch_is_a_usage_error", func(t *testing.T) {
		poison := cvoNewPoison(t)
		// A directory that is not a repository: the name check must come first, so the
		// message names the flag and not "not a git repository".
		res := rlsRun(t, t.TempDir(), poison.env(), "bash", cvoScript(t, cvoPrecheckRel), "--exclude-card", "t1453")
		poison.assertUntouched(t)
		cvoWantUsage(t, res, "--retiring-branch")
	})

	t.Run("precheck_with_an_empty_retiring_branch_is_a_usage_error", func(t *testing.T) {
		poison := cvoNewPoison(t)
		res := rlsRun(t, t.TempDir(), poison.env(), "bash", cvoScript(t, cvoPrecheckRel), "--retiring-branch", "")
		poison.assertUntouched(t)
		cvoWantUsage(t, res, "--retiring-branch")
	})

	t.Run("rehearsal_names_each_missing_branch_argument", func(t *testing.T) {
		src, _, _ := cvoSource(t)
		work := filepath.Join(t.TempDir(), "m6")

		res := cvoRehearse(t, "--source", src, "--workdir", work)
		cvoWantUsage(t, res, "--retiring-branch")

		res = cvoRehearse(t, "--source", src, "--workdir", work, "--retiring-branch", "develop")
		cvoWantUsage(t, res, "--target-branch")

		res = cvoRehearse(t, "--source", src, "--workdir", work, "--target-branch", "main")
		cvoWantUsage(t, res, "--retiring-branch")

		// A refused run leaves nothing behind.
		if _, err := os.Stat(work); err == nil {
			t.Errorf("a usage error created the scratch directory %s", work)
		}
	})

	t.Run("rehearsal_remote_check_needs_no_branch_names", func(t *testing.T) {
		base := t.TempDir()
		bare := filepath.Join(base, "bare.git")
		rlsGit(t, base, "init", "-q", "--bare", bare)
		repo := filepath.Join(base, "r")
		rlsGit(t, base, "init", "-q", repo)
		rlsGit(t, repo, "remote", "add", "origin", bare)
		if res := cvoRehearse(t, "--check-remotes-only", repo); res.exit != 0 {
			t.Errorf("the remote check reads no branch name; exit = %d\n%s", res.exit, rlsNorm(res.out))
		}
	})

	t.Run("protection_compare_baseline_names_the_missing_retiring_branch", func(t *testing.T) {
		p := cvoNewProtection(t)
		cvoWantUsage(t, p.runRaw(), "--retiring-branch")
		cvoWantUsage(t, p.runRaw("--expect", "baseline"), "--retiring-branch")
		if calls := p.calls(); len(calls) != 0 {
			t.Errorf("a usage error must precede every gh read, got %d call(s): %v", len(calls), calls)
		}
	})

	t.Run("protection_compare_reads_the_branch_it_was_given", func(t *testing.T) {
		p := cvoNewProtection(t)
		// The result is not judged here (the stub answers only the fixture's branch);
		// the logged endpoint is what proves the supplied name reached the read.
		_ = p.runRaw("--retiring-branch", "alpha")
		var sawAlpha bool
		for _, c := range p.calls() {
			if strings.Contains(c, "branches/alpha/protection") {
				sawAlpha = true
			}
			if strings.Contains(c, "branches/"+cvoRetiringBranch+"/protection") {
				t.Errorf("the read used a name that was not supplied: %q", c)
			}
		}
		if !sawAlpha {
			t.Errorf("no read of branches/alpha/protection; calls: %v", p.calls())
		}
	})

	t.Run("protection_compare_post_cutover_reads_no_retiring_branch", func(t *testing.T) {
		p := cvoNewProtection(t)
		p.edit("main-protection.json", `, "Release PR Multi-OS Gate"`, ``)
		res := p.runRaw("--expect", "post-cutover")
		if res.exit != 0 {
			t.Errorf("post-cutover judges no retiring-branch state and needs no name; exit = %d\n%s", res.exit, rlsNorm(res.out))
		}
	})
}
