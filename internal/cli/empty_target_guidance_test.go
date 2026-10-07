package cli

// empty_target_guidance_test.go — SPEC-GITHUB-FLOW-DEFAULT-001 M1b (card t1453).
//
// The cli-side base-branch consumers refuse an empty integration target rather
// than substituting a branch (M1). A refusal that does not say WHAT TO SET is
// itself a defect. Every empty-target cause of the interpretation table is run
// through each cli-side refusal site and the message must name the config key
// or file to fix plus the site's own escape:
//
//   - goal approve        — none (the contract has no override flag)
//   - factory merge ready — "pass --develop <branch>", never the merge-SOURCE flag --branch
//   - factory complete    — "re-acquire with --branch <integration-target>"
//   - codex card base     — the reason string recorded in the scope basis
//
// The worktree-side sites (done, sweep, MERGE_NOT_ON_ORIGIN) are covered in
// internal/cli/worktree/empty_target_guidance_test.go.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// emptyTargetWants maps each empty-target row of interpretationRows() to the
// substrings the refusal must carry: the key or file the reader has to fix.
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

// emptyTargetRows returns the rows that resolve no target, each with its
// expectation; it fails the test when a row has none (a new table row must be
// given an expectation, never silently skipped).
func emptyTargetRows(t *testing.T) []targetRow {
	t.Helper()
	var rows []targetRow
	for _, r := range interpretationRows() {
		if r.want != "" {
			continue
		}
		if _, ok := emptyTargetWants[r.name]; !ok {
			t.Fatalf("row %q resolves no target but has no expectation: extend emptyTargetWants", r.name)
		}
		rows = append(rows, r)
	}
	if len(rows) == 0 {
		t.Fatal("empty sweep: no empty-target row")
	}
	return rows
}

func assertAllContained(t *testing.T, what, got string, wants []string) {
	t.Helper()
	for _, w := range wants {
		if !strings.Contains(got, w) {
			t.Errorf("%s must contain %q, got:\n%s", what, w, got)
		}
	}
}

func TestEmptyTargetGoalApproveNamesTheFix(t *testing.T) {
	for _, row := range emptyTargetRows(t) {
		t.Run(row.name, func(t *testing.T) {
			_, _, err := approveMissionWithRow(t, row)
			if err == nil {
				t.Fatal("goal approve must refuse when no target resolves")
			}
			assertAllContained(t, "the goal approve refusal", err.Error(), emptyTargetWants[row.name])
		})
	}
}

func TestEmptyTargetMergeReadyNamesTheFixAndTheTargetFlag(t *testing.T) {
	for _, row := range emptyTargetRows(t) {
		t.Run(row.name, func(t *testing.T) {
			run, err := runMergeReadyWithRow(t, row)
			if err == nil {
				t.Fatal("factory merge ready must refuse when no target resolves")
			}
			if run != nil {
				t.Fatalf("a refused resolution must record no run, got %+v", run)
			}
			msg := err.Error()
			assertAllContained(t, "the merge ready refusal", msg, emptyTargetWants[row.name])
			assertAllContained(t, "the merge ready refusal tail", msg, []string{"pass --develop <branch>"})
			if strings.Contains(msg, "--branch") {
				t.Errorf("--branch is the merge SOURCE flag and must not be offered as the target fix, got:\n%s", msg)
			}
		})
	}
}

func TestEmptyTargetFactoryCompleteNamesTheFix(t *testing.T) {
	for _, row := range emptyTargetRows(t) {
		t.Run(row.name, func(t *testing.T) {
			// A window the lane itself holds, recorded against the caller's own
			// tree: acquire fell back because no target resolved.
			root, _, cards := sdMergeFixture(t, false, false, false, 1)
			if err := os.Remove(filepath.Join(root, ".moai", "config", "sections", "git-strategy.yaml")); err != nil && !os.IsNotExist(err) {
				t.Fatal(err)
			}
			installTargetRow(t, root, row)
			sdPlaceMergeReady(t, root, "t1", "lane-1", cards[0])
			sdHoldWindow(t, root, "sess-lane-1", "lane-1", cards[0].branch, baseSourceCallerWire, cards[0].wt, "t1")

			sdLaneEnv(t, "lane-1", "")
			t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
			_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
			if err == nil {
				t.Fatal("factory complete must refuse a caller-source window")
			}
			msg := err.Error()
			assertAllContained(t, "the factory complete refusal", msg, emptyTargetWants[row.name])
			assertAllContained(t, "the factory complete refusal tail", msg, []string{"re-acquire with --branch <integration-target>"})
			if c := fcCard(t, root, "t1"); c.State == homestate.CardMergedLocal {
				t.Error("card reached merged-local on the refusal")
			}
		})
	}
}

func TestEmptyTargetCardMergeBaseNamesTheFix(t *testing.T) {
	for _, row := range emptyTargetRows(t) {
		t.Run(row.name, func(t *testing.T) {
			dir := t.TempDir()
			installTargetRow(t, dir, row)
			_, err := cardMergeBase(dir)
			if err == nil {
				t.Fatal("the card merge base must refuse when no target resolves")
			}
			assertAllContained(t, "the card merge base refusal", err.Error(), emptyTargetWants[row.name])
		})
	}
}
