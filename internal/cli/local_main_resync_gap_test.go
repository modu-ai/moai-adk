package cli

// local_main_resync_gap_test.go — card t1619, codex review P1 (SPEC-LOCAL-MAIN-FLOW-001).
// The re-sync re-checks window ownership and then runs the fast-forward it guards. Both
// must sit in one mutation section (the RunMergeStep shape). A rival's --force takeover
// that lands between the two would leave the previous owner moving primary's main under
// the new owner's merge. The rival here takes the window from inside the gap.

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

func TestLocalMainResyncRivalTakeoverBeforeFastForwardCannotLand(t *testing.T) {
	root := lmfResyncFixture(t, "")
	base := lmfHead(t, root)
	target := lmfOriginCommit(t, root, base, "origin-ahead", map[string]string{"ahead.txt": "ahead\n"}, false)

	var rivalErr error
	prev := localMainResyncBeforeFastForward
	localMainResyncBeforeFastForward = func(repoRoot string) {
		_, rivalErr = factory.AcquireIntegrationLock(repoRoot, factory.IntegrationLock{
			SessionID:    lmfRival,
			SessionName:  "lane-rival",
			PID:          os.Getpid(),
			PIDSource:    factory.PIDSourceSessionOwner,
			Branch:       localMainResyncBranch,
			BranchSource: factory.BranchSourceConfig,
			Worktree:     repoRoot,
		}, true)
	}
	t.Cleanup(func() { localMainResyncBeforeFastForward = prev })

	_, err := runLocalMainResync(root)
	if rivalErr == nil {
		t.Fatalf("a rival --force takeover landed between the ownership re-check and the fast-forward (re-sync err = %v): the move is not serialized with the window's mutation lock", err)
	}
	if err != nil {
		t.Fatalf("with the rival's takeover refused, the re-sync must complete: %v", err)
	}
	if head := lmfHead(t, root); head != target {
		t.Fatalf("local main must move to origin/main %s, got HEAD %s", target, head)
	}
}
