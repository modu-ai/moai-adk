package cli

// factory_card_t1930_test.go — card t1930: the github-flow delivery edge's
// squash-merged regression (defect ②), the re-lease stage/evidence carry-over
// (defect ③), the rebind-evidence verb (defect ④), and the guard row for the
// already-landed origin-ref probing (defect ①, landed in 5a9d34fbb3 — green
// on arrival by design, like the factorylane guard row of the same name).
//
// Every fixture follows factory_card_pr_test.go / factory_card_carry_test.go:
// throwaway repositories under t.TempDir(), a bare origin beside them, `gh`
// doubled through factoryGH, and an isolated SQLite factory store per root.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// ghfSquashOnOrigin simulates what GitHub did to a merged pull request and
// what main did afterwards: one squash commit whose tree is the card
// branch's, then a later landing that edits the same line — so the card
// branch is NOT conflict-free against the evolved origin/main even though
// its pull request is already MERGED. It returns the squash commit (the
// mergeCommit `gh pr view` reports).
func ghfSquashOnOrigin(t *testing.T, f ghfFixture) string {
	t.Helper()
	parent := t.TempDir()
	fcGit(t, parent, "clone", "-q", f.bare, "clone")
	clone := filepath.Join(parent, "clone")
	fcGit(t, clone, "checkout", "-q", "main")
	fcGit(t, clone, "checkout", f.tip, "--", ".")
	fcGit(t, clone, "commit", "-q", "-m", "squash of the card (#7)")
	squash := fcGit(t, clone, "rev-parse", "HEAD")
	if err := os.WriteFile(filepath.Join(clone, "base.txt"), []byte("base\npost-merge edit\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fcGit(t, clone, "commit", "-q", "-am", "a later card edits the same line")
	fcGit(t, clone, "push", "-q", "origin", "main")
	return squash
}

// TestFactoryCompleteRecordsMergedPROnASquashMergedCard — card t1930 defect ②:
// a merge-ready card whose pull request ALREADY reads MERGED (a squash merge
// landed it on origin/main) must reach the merged-pr record through the
// observation path, not the readiness triple — after the squash the card
// branch's commits are not in main's history, so the merge-tree probe reports
// the collision and the deliver path refused a pull request GitHub has
// already merged (reproduced on t1832 PR #1895 and t1894 PR #1897).
func TestFactoryCompleteRecordsMergedPROnASquashMergedCard(t *testing.T) {
	f := ghfNew(t, ghfOpts{syncStatus: "complete"})
	// The pull request's history: the branch was pushed, the PR opened, and
	// GitHub squashed it onto origin/main; main then evolved past the squash.
	fcGit(t, f.root, "push", "-q", "origin", ghfBranch)
	squash := ghfSquashOnOrigin(t, f)
	d := newGHDouble(t, f)
	d.created = true
	d.state, d.mergeOid = "MERGED", squash

	out, err := ghfComplete(t)
	if err != nil {
		t.Fatalf("complete on a squash-merged pull request: %v\nout: %s", err, out)
	}
	c := fcCard(t, f.root, "t1")
	if c.State != homestate.CardMergedPR {
		t.Fatalf("card state = %s, want merged-pr (the merged pull request must be observed, not re-delivered)", c.State)
	}
	if c.MergeSHA != squash {
		t.Errorf("merge = %q, want the squash commit %s", c.MergeSHA, squash)
	}
	if !strings.Contains(out, "merged-pr") {
		t.Errorf("output does not report merged-pr:\n%s", out)
	}
	if d.count("pr", "create") != 0 || d.count("pr", "merge") != 0 {
		t.Errorf("an already-merged pull request was re-opened or re-requested: create=%d merge=%d", d.count("pr", "create"), d.count("pr", "merge"))
	}
}

// TestFactoryCompleteStillRefusesTheUnmergedConflictingCard pins the boundary
// of the defect-② repair: a conflicting card whose pull request does NOT read
// MERGED keeps the readiness refusal verbatim — the merged-PR observation
// branch diverts nothing when the pull request is open or absent.
func TestFactoryCompleteStillRefusesTheUnmergedConflictingCard(t *testing.T) {
	f := ghfNew(t, ghfOpts{syncStatus: "complete", mainMoves: "conflict"})
	d := newGHDouble(t, f)
	_, err := ghfComplete(t)
	if err == nil || !strings.Contains(err.Error(), "conflict-free") {
		t.Fatalf("err = %v, want the merge-readiness refusal naming conflict-free", err)
	}
	if d.count("pr", "create") != 0 || f.remoteBranchTip(t) != "" {
		t.Error("a failing triple still pushed or opened a pull request")
	}
	if c := fcCard(t, f.root, "t1"); c.State != homestate.CardMergeReady {
		t.Errorf("card state = %s, want merge-ready", c.State)
	}
}

// TestFactoryPRReadinessProbesTheOriginRef — card t1930 defect ① guard row
// (the repair itself landed in 5a9d34fbb3, never reimplemented here): the
// readiness triple runs against the integration target's REMOTE ref — the ref
// the pull request merges into — not the possibly stale local branch, and the
// local ref is the fallback only when no fetched origin ref exists.
func TestFactoryPRReadinessProbesTheOriginRef(t *testing.T) {
	newRepo := func(t *testing.T) (string, string, string) {
		t.Helper()
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
		return root, repo, bare
	}
	t.Run("origin_ref_preferred_when_present", func(t *testing.T) {
		root, repo, bare := newRepo(t)
		// origin/main advances past the local branch (the local main is stale).
		parent := t.TempDir()
		fcGit(t, parent, "clone", "-q", bare, "clone")
		clone := filepath.Join(parent, "clone")
		if err := os.WriteFile(filepath.Join(clone, "other.txt"), []byte("other\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		fcGit(t, clone, "add", "other.txt")
		fcGit(t, clone, "commit", "-q", "-m", "origin moves on")
		fcGit(t, clone, "push", "-q", "origin", "main")
		var out strings.Builder
		run, err := factoryPRReadiness(&out, root, homestate.Card{CardID: "t9", WorktreePath: repo}, "lane-1", "WT-9", "main")
		if err != nil {
			t.Fatalf("readiness: %v", err)
		}
		if run.Develop != "origin/main" {
			t.Fatalf("triple target = %q, want origin/main (the fetched remote ref, never the stale local branch)", run.Develop)
		}
	})
	t.Run("local_ref_used_when_no_origin_ref_exists", func(t *testing.T) {
		root, repo, _ := newRepo(t)
		// The fetch fails (the remote is unreachable) and no remote-tracking
		// ref exists: the triple falls back to the local integration branch.
		fcGit(t, repo, "remote", "set-url", "origin", filepath.Join(t.TempDir(), "gone.git"))
		fcGit(t, repo, "update-ref", "-d", "refs/remotes/origin/main")
		var out strings.Builder
		run, err := factoryPRReadiness(&out, root, homestate.Card{CardID: "t9", WorktreePath: repo}, "lane-1", "WT-9", "main")
		if err != nil {
			t.Fatalf("readiness: %v", err)
		}
		if run.Develop != "main" {
			t.Fatalf("triple target = %q, want the local fallback main", run.Develop)
		}
		if !strings.Contains(out.String(), "could not fetch origin/main") {
			t.Errorf("output does not note the failed fetch:\n%s", out.String())
		}
	})
}

// TestFactoryNextCarriesStageAndEvidenceAcrossRunReplacement — card t1930
// defect ③: a card re-leased under a NEW run keeps the stage and evidence_sha
// its most recent prior factory row recorded. The fresh row used to be born
// with both empty, orphaning the card's progress (reproduced: t1828 re-batched
// read stage "-", t1832's re-lease reset the row to stage="" v4 spec_id="").
func TestFactoryNextCarriesStageAndEvidenceAcrossRunReplacement(t *testing.T) {
	root, store := carryFixture(t)
	fcQueue(t, store, factory.BacklogStateQueued)
	landing := carryLanding(root, "t1")
	if err := os.MkdirAll(sdWorktreesDir(root), 0o755); err != nil {
		t.Fatal(err)
	}
	fcGit(t, root, "worktree", "add", "-q", landing, "-b", "WT-previous-tree")
	evidence := strings.Repeat("ab", 20)
	fcPlace(t, root, homestate.Card{
		RunID: carryPrevRun, CardID: "t1", State: homestate.CardMergeReady,
		Stage: homestate.CardMergeReady, EvidenceSHA: evidence, WorktreePath: landing,
	})
	sdRegisterLane(t, root, "lane-1")
	sdLaneEnv(t, "lane-1", "")
	t.Chdir(root)

	out, stderr, err := runFactory(t, "next", "--run", carryNewRun)
	if err != nil {
		t.Fatalf("next over the card's own previous tree: %v\nstdout: %s\nstderr: %s", err, out, stderr)
	}
	c := carryCard(t, root, carryNewRun, "t1")
	if c.Stage != homestate.CardMergeReady {
		t.Errorf("re-leased stage = %q, want %q (the re-lease must carry the prior row's stage, not reset it)", c.Stage, homestate.CardMergeReady)
	}
	if c.EvidenceSHA != evidence {
		t.Errorf("re-leased evidence_sha = %q, want the prior row's %q", c.EvidenceSHA, evidence)
	}
}

// TestFactoryNextKeepsFreshCardsStageless — the defect-③ boundary: a card no
// previous run ever recorded gets no carry-over; its fresh row stays empty
// exactly as before.
func TestFactoryNextKeepsFreshCardsStageless(t *testing.T) {
	root, store := carryFixture(t)
	fcQueue(t, store, factory.BacklogStateQueued)
	sdRegisterLane(t, root, "lane-1")
	sdLaneEnv(t, "lane-1", "")
	t.Chdir(root)

	if _, _, err := runFactory(t, "next", "--run", carryNewRun); err != nil {
		t.Fatalf("next on a fresh card: %v", err)
	}
	c := carryCard(t, root, carryNewRun, "t1")
	if c.Stage != "" || c.EvidenceSHA != "" {
		t.Errorf("a never-before-recorded card carries stage=%q evidence=%q; both must stay empty", c.Stage, c.EvidenceSHA)
	}
}
