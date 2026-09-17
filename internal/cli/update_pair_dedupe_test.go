package cli

// update_pair_dedupe_test.go — SPEC-INIT-UPDATE-CONSISTENCY-001 REQ-ICU-003
// (F12): the update outcome's "Updated N" count must not charge both members
// of a `.sh`/`.sh.tmpl` deployment pair as two files. The counting source is
// managedRedeployCount; the ListTemplates stripped-target contract and the
// t694 display render are untouched.

import "testing"

// TestManagedRedeployCount_PairCountsOnce pins the pair dedupe: a synthetic
// template list carrying the 4 hook-wrapper `.sh`/`.sh.tmpl` pairs (plus a
// standalone `.sh` and a non-managed file) yields one count per deployed
// target — 4 pairs → 4, not 8.
func TestManagedRedeployCount_PairCountsOnce(t *testing.T) {
	t.Parallel()
	files := []string{
		// 4 real hook-wrapper pairs: both list entries, one rendered target.
		".claude/hooks/moai/handle-agent-hook.sh",
		".claude/hooks/moai/handle-agent-hook.sh.tmpl",
		".claude/hooks/moai/handle-session-start.sh",
		".claude/hooks/moai/handle-session-start.sh.tmpl",
		".claude/hooks/moai/handle-subagent-start.sh",
		".claude/hooks/moai/handle-subagent-start.sh.tmpl",
		".claude/hooks/moai/handle-post-tool.sh",
		".claude/hooks/moai/handle-post-tool.sh.tmpl",
		// Standalone managed file with no .tmpl sibling — counts 1.
		".moai/config/sections/workflow.yaml",
		// Non-managed target — never counted.
		"README.md",
	}
	// 4 pairs + 1 standalone = 5 managed deployed targets.
	const wantManaged = 5

	got, restoredSet := managedRedeployCount(files)
	if got != wantManaged {
		t.Fatalf("managedRedeployed = %d, want %d — a .sh/.sh.tmpl pair must count once (REQ-ICU-003)", got, wantManaged)
	}
	// The rendered-target set must carry both the stripped pair target and the
	// standalone entry (removal accounting depends on it).
	for _, want := range []string{
		".claude/hooks/moai/handle-agent-hook.sh",
		".moai/config/sections/workflow.yaml",
	} {
		if !restoredSet[want] {
			t.Errorf("restoredSet missing %q", want)
		}
	}
}

// TestManagedRedeployCount_TmplOnlyEntry pins the reverse order: the .tmpl
// sibling may precede its rendered target in the list; the pair still counts
// once.
func TestManagedRedeployCount_TmplOnlyEntry(t *testing.T) {
	t.Parallel()
	files := []string{
		".claude/hooks/moai/handle-agent-hook.sh.tmpl",
		".claude/hooks/moai/handle-agent-hook.sh",
	}
	got, _ := managedRedeployCount(files)
	if got != 1 {
		t.Fatalf("managedRedeployed = %d, want 1 (order-independent pair dedupe, REQ-ICU-003)", got)
	}
}
