package worktree

// landing_predicate_test.go — SPEC-GITHUB-FLOW-DEFAULT-001 M2-A (card t1453),
// AC-GFD-002 / REQ-GFD-002 / design D-5: the squash-safe landing predicate.
//
// Every fixture is a REAL scratch repository (t.TempDir) with a LOCAL bare
// remote standing in for origin; the integration target is main (the
// github-flow row of the interpretation table). `gh` is never executed: the
// PR layer reads through the landingGH seam, which every test here replaces
// with a double (and landing_gh_guard_test.go makes the default a failing one).
//
// The predicate is exercised through the two production entry points that run
// all three layers — `moai worktree done --auto` (originLandingRefusal) and the
// `moai worktree sweep` evaluation (sweepEvaluate) — never through an internal
// helper, so a mutant that bypasses the shared predicate cannot hide.

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/internal/session"
)

// gfdFixture is one sandbox: bare origin, primary repo on main, and a card
// worktree on a WT- branch forked from the seed commit.
type gfdFixture struct {
	base, origin, repo, tree string
	branch                   string
}

const gfdSeedLines = 20

func newGFDFixture(t *testing.T) *gfdFixture {
	t.Helper()
	base := t.TempDir()
	f := &gfdFixture{
		base:   base,
		origin: filepath.Join(base, "origin.git"),
		repo:   filepath.Join(base, "repo"),
		tree:   filepath.Join(base, "card-wt"),
		branch: "WT-gfd-card",
	}
	landingGit(t, base, "init", "-q", "--bare", f.origin)
	if err := os.MkdirAll(f.repo, 0o755); err != nil {
		t.Fatal(err)
	}
	landingGit(t, f.repo, "init", "-q", "-b", "main")
	landingGit(t, f.repo, "config", "user.email", "gfd-test@example.com")
	landingGit(t, f.repo, "config", "user.name", "GFD Test")
	landingGit(t, f.repo, "config", "commit.gpgsign", "false")
	landingGit(t, f.repo, "remote", "add", "origin", f.origin)
	var seed strings.Builder
	for i := 1; i <= gfdSeedLines; i++ {
		seed.WriteString("line " + itoa(i) + "\n")
	}
	if err := os.WriteFile(filepath.Join(f.repo, "f.txt"), []byte(seed.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	landingGit(t, f.repo, "add", "f.txt")
	landingGit(t, f.repo, "commit", "-q", "-m", "seed")
	landingGit(t, f.repo, "push", "-q", "-u", "origin", "main")
	landingGit(t, f.repo, "worktree", "add", "-q", "-b", f.branch, f.tree)
	// github-flow: the integration target is main. Untracked, like the
	// M1 fixtures, so every row of a test can replace it.
	installBaseRow(t, f.repo, baseRows()[0])
	withTierTestEnv(t, f.repo)
	return f
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for ; n > 0; n /= 10 {
		b = append([]byte{byte('0' + n%10)}, b...)
	}
	return string(b)
}

// editLine rewrites "line <n>" to "line <n> <suffix>" in dir/f.txt.
func editLine(t *testing.T, dir string, n int, suffix string) {
	t.Helper()
	p := filepath.Join(dir, "f.txt")
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(raw), "\n")
	lines[n-1] = lines[n-1] + " " + suffix
	if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (f *gfdFixture) cardCommit(t *testing.T, line int, suffix string) {
	t.Helper()
	editLine(t, f.tree, line, suffix)
	landingGit(t, f.tree, "commit", "-q", "-a", "-m", "card: line "+itoa(line)+" "+suffix)
}

func (f *gfdFixture) mainCommit(t *testing.T, line int, suffix string) {
	t.Helper()
	editLine(t, f.repo, line, suffix)
	landingGit(t, f.repo, "commit", "-q", "-a", "-m", "main: line "+itoa(line)+" "+suffix)
}

// squash lands the card on main the way a GitHub squash merge does: one new
// commit carrying the cumulative change, no ancestry to the card tip.
func (f *gfdFixture) squash(t *testing.T) string {
	t.Helper()
	landingGit(t, f.repo, "merge", "--squash", f.branch)
	landingGit(t, f.repo, "commit", "-q", "-m", "squash: "+f.branch)
	return landingGit(t, f.repo, "rev-parse", "HEAD")
}

func (f *gfdFixture) pushMain(t *testing.T) {
	t.Helper()
	landingGit(t, f.repo, "push", "-q", "origin", "main")
}

func (f *gfdFixture) cardTip(t *testing.T) string {
	t.Helper()
	return landingGit(t, f.repo, "rev-parse", f.branch)
}

// patchIDs computes patch-ids with plain git, independent of the predicate
// under test, so a fixture's precondition is measured rather than assumed.
func gfdPatchID(t *testing.T, dir, stdin string) []string {
	t.Helper()
	cmd := exec.Command("git", "patch-id", "--stable")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(stdin)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git patch-id: %v", err)
	}
	var ids []string
	for _, ln := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if f := strings.Fields(ln); len(f) > 0 {
			ids = append(ids, f[0])
		}
	}
	return ids
}

// cumulativeMatches reports whether the card's cumulative diff patch-id equals
// the patch-id of any commit on origin/main since the merge-base.
func (f *gfdFixture) cumulativeMatches(t *testing.T) bool {
	t.Helper()
	mb := landingGit(t, f.repo, "merge-base", f.branch, "origin/main")
	diff := landingGit(t, f.repo, "diff", "--no-renames", mb, f.branch) + "\n"
	card := gfdPatchID(t, f.repo, diff)
	if len(card) != 1 {
		t.Fatalf("card cumulative patch-id: want 1 id, got %v", card)
	}
	log := landingGit(t, f.repo, "log", "-p", "--no-merges", "--no-renames", "--format=commit %H", mb+"..origin/main") + "\n"
	for _, id := range gfdPatchID(t, f.repo, log) {
		if id == card[0] {
			return true
		}
	}
	return false
}

func (f *gfdFixture) isAncestor(t *testing.T) bool {
	t.Helper()
	cmd := exec.Command("git", "merge-base", "--is-ancestor", f.branch, "origin/main")
	cmd.Dir = f.repo
	return cmd.Run() == nil
}

// ghPR is one element of the double's `gh pr list --json` answer.
type ghPR struct {
	Number      int    `json:"number"`
	State       string `json:"state"`
	HeadRefOid  string `json:"headRefOid"`
	MergeCommit *struct {
		Oid string `json:"oid"`
	} `json:"mergeCommit"`
}

func mergedPR(head, mergeOid string) ghPR {
	pr := ghPR{Number: 7, State: "MERGED", HeadRefOid: head}
	pr.MergeCommit = &struct {
		Oid string `json:"oid"`
	}{Oid: mergeOid}
	return pr
}

// ghDouble stands in for the gh CLI behind the landingGH seam.
type ghDouble struct {
	t         *testing.T
	mu        sync.Mutex
	calls     [][]string
	prs       []ghPR
	hang      bool
	failWith  error
	raw       []byte // when set, returned verbatim instead of the JSON of prs
	cancelled bool
}

func (d *ghDouble) run(ctx context.Context, _ string, args ...string) ([]byte, error) {
	d.mu.Lock()
	d.calls = append(d.calls, append([]string(nil), args...))
	d.mu.Unlock()
	if d.hang {
		select {
		case <-ctx.Done():
			d.mu.Lock()
			d.cancelled = true
			d.mu.Unlock()
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
			d.t.Error("gh double was never cancelled: the PR layer did not bound the call with its timeout")
			return nil, errors.New("gh double never cancelled")
		}
	}
	if d.failWith != nil {
		return nil, d.failWith
	}
	if d.raw != nil {
		return d.raw, nil
	}
	return json.Marshal(d.prs)
}

func (d *ghDouble) callCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.calls)
}

// installGH replaces the gh seam with the double and shortens the timeout so a
// hung double is cancelled promptly. Both are restored when the test ends.
func installGH(t *testing.T, d *ghDouble) *ghDouble {
	t.Helper()
	d.t = t
	origGH, origTO := landingGH, landingGHTimeout
	landingGH = d.run
	landingGHTimeout = 150 * time.Millisecond
	t.Cleanup(func() { landingGH, landingGHTimeout = origGH, origTO })
	return d
}

// sweepLanded evaluates the card tree through the sweep's per-tree evaluation
// and returns its landed state and verdict. The fetch is the real one against
// the fixture's local origin, exactly as classifySweepVerdicts runs it.
func (f *gfdFixture) sweepLanded(t *testing.T) (landed, verdict, reason string) {
	t.Helper()
	origCWD := sweepProcessCWDs
	sweepProcessCWDs = func() ([]string, error) { return nil, nil }
	t.Cleanup(func() { sweepProcessCWDs = origCWD })
	fetchErr := sweepFetchBase(f.repo, "origin/main")
	v := sweepEvaluate(git.Worktree{Path: f.tree, Branch: f.branch}, sweepEvalInputs{
		base:     "origin/main",
		fetchErr: fetchErr,
		cwdDirs:  nil,
		locks:    map[string]session.LockInfo{},
		now:      time.Now(),
	})
	return v.Landed, v.Verdict, v.Reason
}

// doneLanded runs `moai worktree done --auto` on the card and reports whether
// disposal proceeded (the tree is gone). A refusal leaves the tree standing.
func (f *gfdFixture) doneLanded(t *testing.T) (landed bool, err error) {
	t.Helper()
	err = executeDoneForTierGuard(t, "--auto", f.branch)
	_, statErr := os.Stat(f.tree)
	gone := os.IsNotExist(statErr)
	if err == nil && !gone {
		t.Fatalf("done returned nil but the tree %s is still there", f.tree)
	}
	return err == nil && gone, err
}

// driftSquash builds the case the plain one-commit squash cannot: main touched
// a line near the card's change BEFORE the squash, so the squash commit's diff
// context differs from the card's cumulative diff and no patch-id matches
// (the layer-2 miss that layer 3 exists for). A later main change to the same
// region follows the squash, as AC-GFD-002 F5 describes.
func (f *gfdFixture) driftSquash(t *testing.T, push bool) (squashSHA string) {
	t.Helper()
	f.cardCommit(t, 10, "card")
	f.mainCommit(t, 13, "other")
	squashSHA = f.squash(t)
	f.mainCommit(t, 10, "later")
	if push {
		f.pushMain(t)
	}
	return squashSHA
}

// TestLandingPredicateSquashSafe is AC-GFD-002: nine fixtures, each answered by
// BOTH the `done` landing check and the `sweep` evaluation (REQ-GFD-002 first
// clause — all three layers).
func TestLandingPredicateSquashSafe(t *testing.T) {
	visited := 0
	type cell struct {
		name       string
		build      func(t *testing.T, f *gfdFixture) *ghDouble
		wantLanded bool
		// wantGH is the minimum number of gh calls across done+sweep; -1 means
		// "exactly zero" (layers 1-2 must answer without the network).
		ghCalls func(n int) bool
		why     string
	}
	cells := []cell{
		{name: "F1_one_commit_card_squash_merged", wantLanded: true, why: "layer 2: one commit, cumulative patch-id equals the squash commit's",
			build: func(t *testing.T, f *gfdFixture) *ghDouble {
				f.cardCommit(t, 5, "card")
				f.squash(t)
				f.pushMain(t)
				return &ghDouble{}
			}, ghCalls: func(n int) bool { return n == 0 }},
		{name: "F2_two_commit_card_squash_merged", wantLanded: true, why: "layer 2: git cherry reads two '+', the CUMULATIVE patch-id still matches",
			build: func(t *testing.T, f *gfdFixture) *ghDouble {
				f.cardCommit(t, 5, "card")
				f.cardCommit(t, 15, "card")
				if f.squash(t) == "" {
					t.Fatal("no squash commit")
				}
				f.pushMain(t)
				return &ghDouble{}
			}, ghCalls: func(n int) bool { return n == 0 }},
		{name: "F3_merge_commit_merge", wantLanded: true, why: "layer 1: ancestry",
			build: func(t *testing.T, f *gfdFixture) *ghDouble {
				f.cardCommit(t, 5, "card")
				landingGit(t, f.repo, "merge", "--no-ff", "-q", "-m", "merge card", f.branch)
				f.pushMain(t)
				return &ghDouble{}
			}, ghCalls: func(n int) bool { return n == 0 }},
		{name: "F4_unmerged_card", wantLanded: false, why: "no layer confirms; the PR layer finds no merged PR",
			build: func(t *testing.T, f *gfdFixture) *ghDouble {
				f.cardCommit(t, 5, "card")
				return &ghDouble{prs: []ghPR{}}
			}, ghCalls: func(n int) bool { return n >= 0 }},
		{name: "F5_squash_then_main_changed_same_region", wantLanded: true, why: "layer 3: layer 2 cannot match (squash context drifted), PR MERGED + head == tip + merge commit on origin/main",
			build: func(t *testing.T, f *gfdFixture) *ghDouble {
				sq := f.driftSquash(t, true)
				return &ghDouble{prs: []ghPR{mergedPR(f.cardTip(t), sq)}}
			}, ghCalls: func(n int) bool { return n >= 2 }},
		{name: "F6_fetch_failure_cannot_confirm", wantLanded: false, why: "origin unreachable: a landing nobody can confirm is not a landing, even with a MERGED PR in hand",
			build: func(t *testing.T, f *gfdFixture) *ghDouble {
				f.cardCommit(t, 5, "card")
				sq := f.squash(t)
				f.pushMain(t)
				landingGit(t, f.repo, "remote", "set-url", "origin", filepath.Join(f.base, "missing.git"))
				return &ghDouble{prs: []ghPR{mergedPR(f.cardTip(t), sq)}}
			}, ghCalls: func(n int) bool { return n >= 0 }},
		{name: "F7_merged_pr_but_local_tip_moved_on", wantLanded: false, why: "PR MERGED but the local tip is not the PR head: a later local commit would be deleted",
			build: func(t *testing.T, f *gfdFixture) *ghDouble {
				f.cardCommit(t, 10, "card")
				f.mainCommit(t, 13, "other")
				sq := f.squash(t)
				f.pushMain(t)
				merged := f.cardTip(t)
				f.cardCommit(t, 3, "after-merge") // the unmerged work a "merged PR means landed" mutant deletes
				return &ghDouble{prs: []ghPR{mergedPR(merged, sq)}}
			}, ghCalls: func(n int) bool { return n >= 2 }},
		{name: "F8_gh_exceeds_the_timeout", wantLanded: false, why: "gh hangs: the call is bounded and the timeout reads as 'cannot answer', never as landed",
			build: func(t *testing.T, f *gfdFixture) *ghDouble {
				f.driftSquash(t, true)
				return &ghDouble{hang: true}
			}, ghCalls: func(n int) bool { return n >= 2 }},
		{name: "F9_merge_commit_not_reached_on_local_integration_ref", wantLanded: false, why: "MERGED and head == tip, but the merge commit is not an ancestor of origin/main yet (fetch lag): only the third condition is false",
			build: func(t *testing.T, f *gfdFixture) *ghDouble {
				sq := f.driftSquash(t, false) // squash + later change exist locally, never pushed
				return &ghDouble{prs: []ghPR{mergedPR(f.cardTip(t), sq)}}
			}, ghCalls: func(n int) bool { return n >= 2 }},
	}

	for _, c := range cells {
		t.Run(c.name, func(t *testing.T) {
			visited++
			f := newGFDFixture(t)
			d := installGH(t, c.build(t, f))
			if strings.HasPrefix(c.name, "F5_") || strings.HasPrefix(c.name, "F8_") || strings.HasPrefix(c.name, "F9_") {
				// Fixture precondition, measured: layer 1 and layer 2 cannot confirm.
				if f.isAncestor(t) {
					t.Fatal("fixture invalid: the squash must not leave the tip an ancestor")
				}
				if c.name != "F9_merge_commit_not_reached_on_local_integration_ref" && f.cumulativeMatches(t) {
					t.Fatal("fixture invalid: layer 2 would answer; the cell must force layer 3")
				}
			}

			// sweep first (read-only), then done (removes the tree when landed).
			sweepLanded, verdict, reason := f.sweepLanded(t)
			gotSweep := sweepLanded == staleStateYes && verdict == sweepDispose
			if gotSweep != c.wantLanded {
				t.Errorf("sweep: landed=%q verdict=%q reason=%q, want landed=%v (%s)", sweepLanded, verdict, reason, c.wantLanded, c.why)
			}
			gotDone, doneErr := f.doneLanded(t)
			if gotDone != c.wantLanded {
				t.Errorf("done: landed=%v err=%v, want landed=%v (%s)", gotDone, doneErr, c.wantLanded, c.why)
			}
			if !c.ghCalls(d.callCount()) {
				t.Errorf("gh call count %d violates the cell's expectation (%s)", d.callCount(), c.why)
			}
			if strings.HasPrefix(c.name, "F8_") && !d.cancelled {
				t.Errorf("F8: the gh call was never cancelled by its timeout")
			}
		})
	}
	if visited != 9 {
		t.Fatalf("visited %d fixtures, want 9", visited)
	}
}

// TestLandingPredicateLaterChangeObservation records what a main change made
// AFTER the squash does to the cumulative patch-id layer (plan-audit could not
// reproduce the case; research.md §5 and AC-GFD-002 F5 assert a mismatch). It
// is observed here: the squash commit's own patch-id is untouched by a later
// commit, so layer 2 still confirms and the PR layer is never consulted.
func TestLandingPredicateLaterChangeObservation(t *testing.T) {
	f := newGFDFixture(t)
	f.cardCommit(t, 10, "card")
	f.squash(t)
	f.mainCommit(t, 10, "later") // same line, after the squash
	f.pushMain(t)
	d := installGH(t, &ghDouble{prs: []ghPR{}})
	if !f.cumulativeMatches(t) {
		t.Fatal("observation premise: a later same-line change leaves the squash commit's patch-id matching the card's cumulative patch-id")
	}
	landed, verdict, reason := f.sweepLanded(t)
	if landed != staleStateYes || verdict != sweepDispose {
		t.Errorf("sweep: landed=%q verdict=%q reason=%q, want landed via layer 2", landed, verdict, reason)
	}
	if got, err := f.doneLanded(t); !got {
		t.Errorf("done: want landed via layer 2, got err=%v", err)
	}
	if d.callCount() != 0 {
		t.Errorf("layer 2 answered; gh must not be consulted, got %d call(s)", d.callCount())
	}
}

// TestLandingPredicateCommitCap pins the layer-2 comparison cap: the default is
// 500 (design D-5) and a range over the cap is "cannot answer" — it defers to
// the PR layer instead of reading as landed or as not landed.
func TestLandingPredicateCommitCap(t *testing.T) {
	if landingPatchIDCommitCap != 500 {
		t.Fatalf("landingPatchIDCommitCap = %d, want 500 (design D-5)", landingPatchIDCommitCap)
	}
	build := func(t *testing.T) (*gfdFixture, string) {
		f := newGFDFixture(t)
		f.cardCommit(t, 5, "card")
		sq := f.squash(t)
		for i := 0; i < 4; i++ {
			f.mainCommit(t, 18, "pad"+itoa(i))
		}
		f.pushMain(t)
		return f, sq
	}

	t.Run("over_the_cap_layer_2_cannot_answer_and_layer_3_decides", func(t *testing.T) {
		f, sq := build(t)
		orig := landingPatchIDCommitCap
		landingPatchIDCommitCap = 3 // 5 commits since the merge-base
		t.Cleanup(func() { landingPatchIDCommitCap = orig })
		d := installGH(t, &ghDouble{prs: []ghPR{mergedPR(f.cardTip(t), sq)}})
		landed, verdict, reason := f.sweepLanded(t)
		if landed != staleStateYes || verdict != sweepDispose {
			t.Errorf("sweep: landed=%q verdict=%q reason=%q, want landed via layer 3", landed, verdict, reason)
		}
		if d.callCount() < 1 {
			t.Errorf("over the cap the PR layer must be consulted, got %d gh call(s)", d.callCount())
		}
	})

	t.Run("over_the_cap_without_a_merged_pr_preserves", func(t *testing.T) {
		f, _ := build(t)
		orig := landingPatchIDCommitCap
		landingPatchIDCommitCap = 3
		t.Cleanup(func() { landingPatchIDCommitCap = orig })
		installGH(t, &ghDouble{prs: []ghPR{}})
		landed, verdict, reason := f.sweepLanded(t)
		if landed == staleStateYes || verdict == sweepDispose {
			t.Errorf("sweep: landed=%q verdict=%q reason=%q, want preserve (cap exceeded, no PR)", landed, verdict, reason)
		}
	})

	t.Run("within_the_cap_layer_2_answers", func(t *testing.T) {
		f, _ := build(t)
		d := installGH(t, &ghDouble{prs: []ghPR{}})
		landed, verdict, reason := f.sweepLanded(t)
		if landed != staleStateYes || verdict != sweepDispose {
			t.Errorf("sweep: landed=%q verdict=%q reason=%q, want landed via layer 2", landed, verdict, reason)
		}
		if d.callCount() != 0 {
			t.Errorf("layer 2 answered; gh must not be consulted, got %d call(s)", d.callCount())
		}
	})
}

// TestLandingPredicateGHReadsTheCardBranch pins the PR lookup shape: by the
// card branch as head, merged PRs only, bounded (no unbounded list).
func TestLandingPredicateGHReadsTheCardBranch(t *testing.T) {
	f := newGFDFixture(t)
	f.driftSquash(t, true)
	d := installGH(t, &ghDouble{prs: []ghPR{}})
	f.sweepLanded(t)
	if d.callCount() == 0 {
		t.Fatal("the PR layer made no gh call for a card layers 1-2 cannot confirm")
	}
	call := strings.Join(d.calls[0], " ")
	for _, want := range []string{"pr", "list", "--head " + f.branch, "--state merged", "headRefOid", "mergeCommit"} {
		if !strings.Contains(call, want) {
			t.Errorf("gh call %q lacks %q", call, want)
		}
	}
}

// TestLandingPredicateGHFailureIsFailClosed: any gh failure — a non-zero exit,
// malformed JSON — reads as "cannot answer", never as landed.
func TestLandingPredicateGHFailureIsFailClosed(t *testing.T) {
	cases := map[string]*ghDouble{
		"non_zero_exit":  {failWith: errors.New("gh: exit status 1 (not logged in)")},
		"malformed_json": {raw: []byte("{not json")},
	}
	for name, d := range cases {
		t.Run(name, func(t *testing.T) {
			f := newGFDFixture(t)
			f.driftSquash(t, true)
			installGH(t, d)
			landed, verdict, reason := f.sweepLanded(t)
			if landed == staleStateYes || verdict == sweepDispose {
				t.Errorf("sweep: landed=%q verdict=%q reason=%q, want preserve on a gh failure", landed, verdict, reason)
			}
		})
	}
}

// Patch-id's default whitespace folding is unsafe for a deletion predicate:
// whitespace inside a literal is data, even when Git considers the patches equal.
func TestLandingPredicatePreservesWhitespaceDistinctContent(t *testing.T) {
	for _, tc := range []struct {
		name, card, remote string
	}{
		{"string_spaces", "package fixture\nconst value = \"a  b\"\n", "package fixture\nconst value = \"a b\"\n"},
		{"configured_textconv", "package fixture\nconst value = \"a  b\"\n", "package fixture\nconst value = \"a b\"\n"},
		{"string_tab", "package fixture\nconst value = `a\tb`\n", "package fixture\nconst value = `a b`\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newGFDFixture(t)
			if tc.name == "configured_textconv" {
				if _, err := exec.LookPath("sed"); err != nil {
					t.Skip("textconv fixture needs sed")
				}
				landingGit(t, f.repo, "config", "diff.landing-fold.textconv", "sed -e 's/ //g'")
				if err := os.WriteFile(filepath.Join(f.repo, ".git", "info", "attributes"), []byte("value.go diff=landing-fold\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			for _, change := range []struct{ dir, content string }{{f.tree, tc.card}, {f.repo, tc.remote}} {
				if err := os.WriteFile(filepath.Join(change.dir, "value.go"), []byte(change.content), 0o644); err != nil {
					t.Fatal(err)
				}
				landingGit(t, change.dir, "add", "value.go")
				landingGit(t, change.dir, "commit", "-q", "-m", "add distinct literal")
			}
			f.pushMain(t)
			installGH(t, &ghDouble{})
			if !f.cumulativeMatches(t) {
				t.Fatal("fixture must collide under whitespace-folding patch-id")
			}
			if landed, err := LandedByPatchID(f.repo, f.branch, "origin/main"); err != nil || landed {
				t.Errorf("distinct literal must not be landed: landed=%v err=%v", landed, err)
			}
			if landed, verdict, reason := f.sweepLanded(t); landed == staleStateYes || verdict == sweepDispose {
				t.Errorf("sweep must preserve distinct content: landed=%q verdict=%q reason=%q", landed, verdict, reason)
			}
			if gone, err := f.doneLanded(t); gone || err == nil {
				t.Errorf("done must refuse distinct content: gone=%v err=%v", gone, err)
			}
			if raw, err := os.ReadFile(filepath.Join(f.tree, "value.go")); err != nil || string(raw) != tc.card {
				t.Errorf("card bytes must survive: content=%q err=%v", raw, err)
			}
		})
	}
}

// Relocating a byte-identical patch must remain squash-safe even when another
// integration commit changes its hunk line numbers.
func TestLandingPredicateRecognizesRelocatedSquash(t *testing.T) {
	f := newGFDFixture(t)
	f.cardCommit(t, 15, "card")
	p := filepath.Join(f.repo, "f.txt")
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, append([]byte("unrelated prefix\n"), raw...), 0o644); err != nil {
		t.Fatal(err)
	}
	landingGit(t, f.repo, "commit", "-q", "-a", "-m", "move card hunk down")
	f.squash(t)
	f.pushMain(t)
	if landed, err := LandedByPatchID(f.repo, f.branch, "origin/main"); err != nil || !landed {
		t.Fatalf("byte-identical relocated squash must be landed: landed=%v err=%v", landed, err)
	}
}

// Tool failures and incomplete merge metadata must never authorize deleting a
// worktree, including when a Git version cannot compute verbatim patch IDs.
func TestLandingPredicatePreservesUnconfirmableState(t *testing.T) {
	f := newGFDFixture(t)
	f.mainCommit(t, 5, "unrelated")
	f.pushMain(t)
	if landed, err := LandedByPatchID(f.repo, f.branch, "origin/main"); landed || err == nil {
		t.Fatalf("empty card must remain unconfirmed: landed=%v err=%v", landed, err)
	}
	t.Run("missing_card_tip", func(t *testing.T) {
		if landed, err := LandedByPatchID(f.repo, "missing-card-tip", "origin/main"); landed || err == nil {
			t.Fatalf("missing tip must remain unconfirmed: landed=%v err=%v", landed, err)
		}
		if landed, err := landedByMergedPR(f.repo, "missing-card-tip", "origin/main"); landed || err == nil {
			t.Fatalf("PR layer must reject an unresolved local tip: landed=%v err=%v", landed, err)
		}
	})
	t.Run("git_unavailable", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		if _, err := landingPatchIDs(f.repo, "unreadable patch"); err == nil {
			t.Fatal("missing Git must fail patch-ID computation")
		}
	})
	for _, tc := range []struct {
		name    string
		pr      ghPR
		wantErr bool
	}{
		{"missing_merge_commit", ghPR{State: "MERGED", HeadRefOid: f.cardTip(t)}, false},
		{"unknown_merge_commit", mergedPR(f.cardTip(t), strings.Repeat("f", 40)), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			installGH(t, &ghDouble{prs: []ghPR{tc.pr}})
			if landed, err := landedByMergedPR(f.repo, f.branch, "origin/main"); landed || (err != nil) != tc.wantErr {
				t.Fatalf("unconfirmable PR must preserve the tree: landed=%v err=%v", landed, err)
			}
		})
	}
}
