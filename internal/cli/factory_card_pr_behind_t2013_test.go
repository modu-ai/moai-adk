package cli

// factory_card_pr_behind_t2013_test.go — card t2013: the PR readiness read
// answers whether the PR head contains the integration target's current head.
// A head that is behind is NOT a refusal of the PR: readiness still succeeds,
// but the output tells the operator to absorb origin/<target> first and says
// "behind N" — absorbing the base is the precondition for re-running CI and for
// the ready transition, and no CI rerun is requested here. A base that could not
// be verified at all (fetch failed, no tracking ref, local fallback) says so and
// refuses readiness with the stale-base reason.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// behindFixture is a card repo on main with a bare origin; advance pushes one
// more commit to origin/main from a separate clone.
func behindFixture(t *testing.T) (root, repo, bare string) {
	t.Helper()
	root, _ = fcFixture(t)
	repo = filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(repo, 0o700); err != nil {
		t.Fatal(err)
	}
	fcGit(t, repo, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(repo, "base.txt"), []byte("base\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fcGit(t, repo, "add", "base.txt")
	fcGit(t, repo, "commit", "-q", "-m", "base")
	bare = filepath.Join(t.TempDir(), "origin.git")
	fcGit(t, filepath.Dir(bare), "init", "-q", "--bare", "-b", "main", bare)
	fcGit(t, repo, "remote", "add", "origin", bare)
	fcGit(t, repo, "push", "-q", "origin", "main")
	return root, repo, bare
}

func behindAdvanceOrigin(t *testing.T, bare string, commits int) {
	t.Helper()
	clone := filepath.Join(t.TempDir(), "clone")
	fcGit(t, filepath.Dir(clone), "clone", "-q", bare, clone)
	for i := 0; i < commits; i++ {
		name := filepath.Join(clone, "fix"+string(rune('a'+i))+".txt")
		if err := os.WriteFile(name, []byte("fix\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		fcGit(t, clone, "add", filepath.Base(name))
		fcGit(t, clone, "commit", "-q", "-m", "common fix")
	}
	fcGit(t, clone, "push", "-q", "origin", "main")
}

func TestFactoryPRReadinessReportsBehindWithoutRefusing(t *testing.T) {
	root, repo, bare := behindFixture(t)
	behindAdvanceOrigin(t, bare, 2)

	var out strings.Builder
	_, err := factoryPRReadiness(&out, root, homestate.Card{CardID: "t9", WorktreePath: repo}, "lane-1", "WT-9", "main")
	if err != nil {
		t.Fatalf("readiness against a behind head: %v, want no refusal of the PR", err)
	}
	for _, want := range []string{"behind 2", "absorb origin/main first"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, out.String())
		}
	}
	if strings.Contains(strings.ToLower(out.String()), "rerun requested") {
		t.Errorf("output claims a CI rerun was requested:\n%s", out.String())
	}
}

func TestFactoryPRReadinessSaysNothingWhenHeadContainsBase(t *testing.T) {
	root, repo, _ := behindFixture(t)

	var out strings.Builder
	if _, err := factoryPRReadiness(&out, root, homestate.Card{CardID: "t9", WorktreePath: repo}, "lane-1", "WT-9", "main"); err != nil {
		t.Fatalf("readiness: %v", err)
	}
	if strings.Contains(out.String(), "behind") {
		t.Errorf("a head that contains origin/main reports behind:\n%s", out.String())
	}
}

func TestFactoryPRReadinessRefusesUnverifiedBaseWithoutTrackingRef(t *testing.T) {
	root, repo, _ := behindFixture(t)
	fcGit(t, repo, "remote", "set-url", "origin", filepath.Join(t.TempDir(), "gone.git"))
	fcGit(t, repo, "update-ref", "-d", "refs/remotes/origin/main")

	var out strings.Builder
	run, err := factoryPRReadiness(&out, root, homestate.Card{CardID: "t9", WorktreePath: repo}, "lane-1", "WT-9", "main")
	if err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("readiness against an unverified base: err = %v, want the stale-base refusal", err)
	}
	if run.Develop != "main" {
		t.Errorf("fallback target = %q, want the kept local fallback main", run.Develop)
	}
	if !strings.Contains(out.String(), "base: unverified (fetch failed)") {
		t.Errorf("output lacks the unverified-base line:\n%s", out.String())
	}
}
