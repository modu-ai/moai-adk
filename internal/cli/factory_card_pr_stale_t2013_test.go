package cli

// factory_card_pr_stale_t2013_test.go — card t2013: the pull-request edge's
// readiness check refuses a base it could not refresh. When the fetch of the
// integration target fails while a remote-tracking ref from an earlier fetch
// still exists, that ref is stale: judging the triple against it would report
// readiness against a base that may have moved, so the result is marked stale
// and the check refuses. A fetch that succeeds is unchanged (the t1930 guard
// row), as is the no-tracking-ref fallback.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestFactoryPRReadinessRefusesStaleBaseWhenFetchFails(t *testing.T) {
	root, _ := fcFixture(t)
	repo := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(repo, 0o700); err != nil {
		t.Fatal(err)
	}
	fcGit(t, repo, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(repo, "base.txt"), []byte("base\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fcGit(t, repo, "add", "base.txt")
	fcGit(t, repo, "commit", "-q", "-m", "base")
	bare := filepath.Join(t.TempDir(), "origin.git")
	fcGit(t, filepath.Dir(bare), "init", "-q", "--bare", "-b", "main", bare)
	fcGit(t, repo, "remote", "add", "origin", bare)
	fcGit(t, repo, "push", "-q", "origin", "main")
	// The earlier fetch left refs/remotes/origin/main behind; the remote is now unreachable.
	fcGit(t, repo, "remote", "set-url", "origin", filepath.Join(t.TempDir(), "gone.git"))
	if got := fcGit(t, repo, "rev-parse", "--verify", "--quiet", "refs/remotes/origin/main"); strings.TrimSpace(got) == "" {
		t.Fatal("fixture: the stale remote-tracking ref is missing")
	}

	var out strings.Builder
	run, err := factoryPRReadiness(&out, root, homestate.Card{CardID: "t9", WorktreePath: repo}, "lane-1", "WT-9", "main")
	if err == nil {
		t.Fatalf("readiness against an unrefreshable base returned no error (run %+v), want a stale refusal", run)
	}
	if !strings.Contains(err.Error(), "stale") {
		t.Errorf("error = %q, want it to name the stale base", err.Error())
	}
	if run.AllPassed {
		t.Errorf("a refused readiness result carries AllPassed = true")
	}
}
