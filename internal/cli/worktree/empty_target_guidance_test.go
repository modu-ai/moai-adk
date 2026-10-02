package worktree

// empty_target_guidance_test.go — SPEC-GITHUB-FLOW-DEFAULT-001 M1b (card t1453).
//
// M1 made an empty integration target a refusal. A refusal that does not say
// WHAT TO SET is itself a defect, so the two worktree-side refusals — `done`
// (ORIGIN_LANDING_UNCONFIRMED) and `sweep` — must name the config key or file
// to fix for every cause that leaves the target empty, and end in the site's
// own escape. `done`'s MERGE_NOT_ON_ORIGIN must say where its landing base came
// from, so a reverted config reads as a reverted config, not an unmerged card.

import (
	"strings"
	"testing"
)

// emptyTargetWants maps each empty-target row of baseRows() to the substrings
// the refusal must carry: the key or file the reader has to fix.
var emptyTargetWants = map[string][]string{
	"git-flow empty develop_branch": {
		"git_strategy.manual.develop_branch is empty",
		".moai/config/sections/git-strategy.yaml",
		"git_strategy.manual.workflow: github-flow",
	},
	"git-flow outside the manual gate": {
		"git-flow resolves a target only when git_strategy.mode is manual",
		"git_strategy.personal.workflow",
	},
	"unknown workflow": {
		`git_strategy.manual.workflow "svn-flow" is not one of github-flow, git-flow, gitlab-flow, release-flow`,
		".moai/config/sections/git-strategy.yaml",
	},
	"no git-strategy.yaml": {
		"cannot read",
		".moai/config/sections/git-strategy.yaml",
		"absent or not valid YAML",
	},
	"unparseable git-strategy.yaml": {
		"cannot read",
		".moai/config/sections/git-strategy.yaml",
		"absent or not valid YAML",
	},
}

func assertContainsAll(t *testing.T, what, got string, wants []string) {
	t.Helper()
	for _, w := range wants {
		if !strings.Contains(got, w) {
			t.Errorf("%s must contain %q, got:\n%s", what, w, got)
		}
	}
}

// TestEmptyTargetRefusalsNameTheFix runs every empty-target cause through the
// two worktree-side refusals and asserts reason, fix, and the site's tail.
func TestEmptyTargetRefusalsNameTheFix(t *testing.T) {
	executed := 0
	for _, row := range baseRows() {
		if row.want != "" {
			continue
		}
		wants, ok := emptyTargetWants[row.name]
		if !ok {
			t.Fatalf("row %q resolves no target but has no expectation: extend emptyTargetWants", row.name)
		}
		t.Run("done/"+row.name, func(t *testing.T) {
			executed++
			f := newLandingFixture(t)
			installBaseRow(t, f.repo, row)
			err := executeDoneForTierGuard(t, "--auto", f.branch)
			if err == nil || !strings.Contains(err.Error(), "ORIGIN_LANDING_UNCONFIRMED") {
				t.Fatalf("done must refuse fail-closed with ORIGIN_LANDING_UNCONFIRMED, got: %v", err)
			}
			assertContainsAll(t, "the done refusal", err.Error(), wants)
			assertTreeSurvives(t, f.tree)
		})
		t.Run("sweep/"+row.name, func(t *testing.T) {
			executed++
			fetched, err := sweepDefaultBaseFor(t, row)
			if err == nil {
				t.Fatalf("sweep must refuse before any fetch; fetched %v", fetched)
			}
			assertContainsAll(t, "the sweep refusal", err.Error(), wants)
			assertContainsAll(t, "the sweep refusal tail", err.Error(), []string{"pass --base origin/<branch>"})
		})
	}
	if executed == 0 {
		t.Fatal("empty sweep: no empty-target row was executed")
	}
}

// TestMergeNotOnOriginNamesItsBase: the refusal says which config value made
// the landing base what it is.
func TestMergeNotOnOriginNamesItsBase(t *testing.T) {
	rows := []struct {
		name   string
		row    baseRow
		target string
		want   []string
	}{
		{"git-flow", gitFlowBaseRow(), "develop", []string{"landing base origin/develop from git_strategy.manual.workflow=git-flow"}},
		{"github-flow", baseRows()[0], "main", []string{"landing base origin/main from git_strategy.manual.workflow=github-flow"}},
	}
	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			f := newLandingFixture(t)
			installBaseRow(t, f.repo, tc.row)
			if tc.target != "develop" {
				// The target branch exists on the remote (the seed commit) but
				// does not carry the card.
				landingGit(t, f.repo, "push", "-q", "origin", "develop:refs/heads/"+tc.target)
			}
			err := executeDoneForTierGuard(t, "--auto", f.branch)
			if err == nil || !strings.Contains(err.Error(), "MERGE_NOT_ON_ORIGIN") {
				t.Fatalf("done must refuse with MERGE_NOT_ON_ORIGIN, got: %v", err)
			}
			assertContainsAll(t, "the MERGE_NOT_ON_ORIGIN refusal", err.Error(), tc.want)
			assertTreeSurvives(t, f.tree)
		})
	}
}
