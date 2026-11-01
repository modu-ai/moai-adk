package cli

// N37 first bullet (audit db0a69698a, residual classification): the lease of a card records the worktree the holder works in.
// When that worktree is gone (the holder removed it, the machine was cleaned), evidence without --repo looks the commit up
// there and says only that it cannot tell which repository the lease worktree belongs to (the audit saw "names no commit in <lease worktree>"); with --repo the same call goes through.
// The expectation: the refusal says that the lease worktree is gone and that --repo names the repository to look in.

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

func TestAFResidual_N37TheRefusalForALeaseWorktreeThatIsGoneNamesTheWayOn(t *testing.T) {
	afSkipWithoutSh(t)
	gone := filepath.Join(t.TempDir(), "wt")
	if err := os.MkdirAll(gone, 0o755); err != nil {
		t.Fatal(err)
	}
	db, repo, cand, st, epoch := afUnqueuedCardAs(t, ledger.ClaimRequest{SessionID: "s-t1", Worktree: gone, PID: os.Getpid(), PIDStart: "start-t1"})
	afBindFixture(t, st, repo)
	if err := os.RemoveAll(gone); err != nil {
		t.Fatal(err)
	}
	args := []string{"evidence", "t1", "--epoch", strconv.FormatInt(epoch, 10), "--kind", "review", "--subject-sha", cand, "--result", "pass"}
	out, _, err := afRunBoth(append(args, "--db", db)...)
	if err == nil {
		t.Skipf("the call without --repo went through (%.160s): nothing to name", out)
	}
	text := strings.Join(strings.Fields(strings.ReplaceAll(err.Error(), gone, "<lease worktree>")), " ") // not stderr: cobra prints the usage there, flags included
	t.Logf("evidence without --repo, lease worktree gone: %.360s", text)
	if _, _, werr := afRunBoth(append(args, "--db", db, "--repo", repo)...); werr != nil {
		t.Fatalf("setup: the way on (--repo) is refused as well: %v", werr)
	}
	if !strings.Contains(text, "--repo") && !strings.Contains(strings.ToLower(text), "gone") && !strings.Contains(strings.ToLower(text), "no longer exist") {
		t.Errorf("the refusal does not say that the lease worktree is gone or that --repo gets on: %.300s", text)
	}
}
