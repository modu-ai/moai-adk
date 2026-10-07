package worktree

// base_defaults_test.go — the git-strategy.yaml row fixtures the landing-base
// tests share, and the git-flow compatibility row of the landed-ref chain
// (SPEC-GITHUB-FLOW-CI-RESIDUE-001 REQ-GFC-001).
//
// The interpretation-table tests this file used to carry
// (TestBaseDefaultsGitFlowUnchanged / TestBaseDefaultsFollowIntegrationTarget,
// card t1453) pinned the contract the chain swap REPLACED: since
// SPEC-GITHUB-FLOW-CI-RESIDUE-001 the landing surfaces resolve their base
// through factory.LandedRefForWithLevel (worktree_base_branch, then
// refs/remotes/origin/HEAD, then the compiled-in default), and the chain
// behavior lives in landing_base_chain_test.go. What stays here is the
// shared row infrastructure — baseRows()/installBaseRow() are consumed by
// landing_predicate_test.go, sweep_test.go and the chain tests — plus the
// compat row: a git-flow project that names its base through
// worktree_base_branch keeps origin/develop (design D-1 compatibility).

import (
	"os"
	"path/filepath"
	"testing"
)

// baseRow is one row of a git-strategy.yaml fixture table (the field set the
// chain tests and the landing-predicate fixtures share).
type baseRow struct {
	name   string
	body   string
	absent bool
	want   string
}

func baseRows() []baseRow {
	manual := func(profile string) string {
		return "git_strategy:\n    mode: manual\n    manual:\n" + profile
	}
	return []baseRow{
		{name: "github-flow", body: manual("        workflow: github-flow\n"), want: "main"},
		{name: "git-flow develop", body: manual("        workflow: git-flow\n        develop_branch: develop\n"), want: "develop"},
		{name: "git-flow custom develop branch", body: manual("        workflow: git-flow\n        develop_branch: staging\n"), want: "staging"},
		{name: "git-flow empty develop_branch", body: manual("        workflow: git-flow\n        develop_branch: \"\"\n"), want: ""},
		{name: "git-flow outside the manual gate", body: "git_strategy:\n    mode: personal\n    personal:\n        workflow: git-flow\n        develop_branch: develop\n", want: ""},
		{name: "gitlab-flow environment", body: manual("        workflow: gitlab-flow\n        environment: production\n"), want: "production"},
		{name: "release-flow prefix", body: manual("        workflow: release-flow\n        release_branch_prefix: release/\n"), want: "release/"},
		{name: "unknown workflow", body: manual("        workflow: svn-flow\n"), want: ""},
		{name: "no git-strategy.yaml", absent: true, want: ""},
		{name: "unparseable git-strategy.yaml", body: "git_strategy: [unterminated\n", want: ""},
	}
}

func gitFlowBaseRow() baseRow { return baseRows()[1] }

// gitFlowCompatRow is the git-flow shape that keeps origin/develop under the
// landed-ref chain: worktree_base_branch IS the develop branch, so level 1
// answers it (the done_landing fixtures install it to keep exercising the
// develop-based machine check).
func gitFlowCompatRow() baseRow {
	return baseRow{
		name: "git-flow with worktree_base_branch develop",
		body: "git_strategy:\n    mode: manual\n    worktree_base_branch: develop\n    manual:\n        workflow: git-flow\n        develop_branch: develop\n",
		want: "develop",
	}
}

// installBaseRow writes the row's git-strategy.yaml under root.
func installBaseRow(t *testing.T, root string, row baseRow) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if row.absent {
		// An absent row also clears a configuration a fixture seeded earlier.
		if err := os.Remove(filepath.Join(dir, "git-strategy.yaml")); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(row.body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// sweepDefaultBaseFor runs the sweep WITHOUT --base against a mock environment
// whose project root carries the row's configuration, and returns the bases the
// fetch seam observed plus the run error.
func sweepDefaultBaseFor(t *testing.T, row baseRow) (fetched []string, err error) {
	t.Helper()
	m := sweepMockEnv(t, nil)
	root := t.TempDir()
	installBaseRow(t, root, row)
	sweepConfigRoot = func() string { return root }
	_, err = runSweepCmd(t, map[string]string{})
	return m.fetchedBases, err
}
