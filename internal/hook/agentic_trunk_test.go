package hook

// agentic_trunk_test.go — SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001 M2 (card
// t1831). AC-AGT-025 Judge 2: the closure push guard's integrationBranchFor
// resolves main through the loader projection, in the tree and in a
// config-orphaned worktree that falls back to the primary checkout's yaml.

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/closure/closuretest"
)

// writeAgenticTrunkConfig writes the fixture's git-strategy.yaml naming
// manual agentic-trunk with no develop_branch key — the state REQ-AGT-018
// produces at the flip.
func writeAgenticTrunkConfig(t *testing.T, f *closuretest.Fixture) {
	t.Helper()
	f.Write(f.Root+"/.moai/config/sections/git-strategy.yaml",
		"git_strategy:\n  mode: manual\n  manual:\n    workflow: agentic-trunk\n")
}

func TestAgenticTrunkIntegrationBranchFor(t *testing.T) {
	t.Run("the primary tree resolves main", func(t *testing.T) {
		f := closuretest.New(t)
		writeAgenticTrunkConfig(t, f)
		if got := integrationBranchFor(f.Root); got != "main" {
			t.Errorf("integrationBranchFor(primary) = %q, want main (the projection)", got)
		}
	})

	t.Run("a config-orphaned worktree falls back to the primary checkout", func(t *testing.T) {
		f := closuretest.New(t)
		writeAgenticTrunkConfig(t, f)
		// The card worktree carries no .moai/config of its own, so
		// integrationBranchFor takes the value from the primary checkout
		// through the git common directory.
		if got := integrationBranchFor(f.CardDir); got != "main" {
			t.Errorf("integrationBranchFor(worktree) = %q, want main through the primary's yaml", got)
		}
	})

	t.Run("git-flow answers unchanged", func(t *testing.T) {
		f := closuretest.New(t)
		writeGitFlowConfig(t, f)
		if got := integrationBranchFor(f.Root); got != "develop" {
			t.Errorf("integrationBranchFor(primary) = %q, want develop under git-flow (unchanged)", got)
		}
	})
}
