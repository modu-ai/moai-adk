package cli

// factory_card_pr_test.go — SPEC-GITHUB-FLOW-DEFAULT-001 M2-B (card t1453):
// the github-flow delivery edge of `moai factory complete` (AC-GFD-004), the
// merge-readiness check that runs before the pull request (AC-GFD-005), and
// the integration window that is no prerequisite of it (AC-GFD-006).
//
// Every fixture is a throwaway repository under t.TempDir() with a bare origin
// beside it; `gh` is a double (factoryGH — the test binary's default is a
// failing one, factory_card_pr_guard_test.go), so nothing here reaches the real
// gh CLI or a network. The git side — push, fetch, merge-tree — is real git on
// those local repositories.

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

const (
	ghfBranch = "WT-ghf-card"
	ghfSpec   = "SPEC-GHF-001"
	ghfLane   = "lane-1"
	ghfPRURL  = "https://github.example/org/repo/pull/7"
)

// ghfOpts shapes one card fixture.
type ghfOpts struct {
	syncStatus string // §E.4 sync_status value; "" = no §E.4 section at all
	mainMoves  string // "" | "conflict" | "unrelated": origin/main advances after the card forked
	mergeCfg   string // merge_method line value; "" omits the key
	noOrigin   bool   // the repository has no remote
	state      string // card state to place; default merge-ready
}

type ghfFixture struct {
	root, bare, wt string
	tip            string
	mainBefore     string
}

func ghfWriteConfig(t *testing.T, root, mergeMethod string) {
	t.Helper()
	body := "git_strategy:\n    mode: manual\n    manual:\n        workflow: github-flow\n"
	if mergeMethod != "" {
		body += "        merge_method: " + mergeMethod + "\n"
	}
	writeGitStrategyBody(t, root, body)
}

// ghfNew builds a github-flow project with one card on its own worktree branch.
func ghfNew(t *testing.T, o ghfOpts) ghfFixture {
	t.Helper()
	// The fixture seeds the queue with `todo add`, which the lane guard
	// refuses when the test PROCESS carries lane variables — a lane session
	// running the suite locally (measured on card t1542). Clear before any
	// seeding; ghfNew re-arms the lane env itself for the complete verb.
	sdClearLaneEnv(t)
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	ghfWriteConfig(t, root, o.mergeCfg)
	fcGit(t, root, "branch", "-M", "main")
	if err := os.WriteFile(filepath.Join(root, "base.txt"), []byte("base\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fcGit(t, root, "add", "base.txt")
	fcGit(t, root, "commit", "-q", "-m", "base")
	f := ghfFixture{root: root}
	if !o.noOrigin {
		f.bare = filepath.Join(t.TempDir(), "origin.git")
		fcGit(t, filepath.Dir(f.bare), "init", "-q", "--bare", "-b", "main", f.bare)
		fcGit(t, root, "remote", "add", "origin", f.bare)
		fcGit(t, root, "push", "-q", "origin", "main")
	}
	f.mainBefore = fcGit(t, root, "rev-parse", "main")

	f.wt = filepath.Join(root, ".claude", "worktrees", "t1")
	fcGit(t, root, "worktree", "add", "-q", "-b", ghfBranch, f.wt)
	if err := os.WriteFile(filepath.Join(f.wt, "base.txt"), []byte("base\ncard edit\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(f.wt, "card.txt"), []byte("card\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if o.syncStatus != "" {
		spec := filepath.Join(f.wt, ".moai", "specs", ghfSpec)
		if err := os.MkdirAll(spec, 0o700); err != nil {
			t.Fatal(err)
		}
		progress := "## §E.2 Run-phase Evidence\n\n...\n\n## §E.4 Sync-phase Audit-Ready Signal\n\nsync_status: " + o.syncStatus + "\n"
		if err := os.WriteFile(filepath.Join(spec, "progress.md"), []byte(progress), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	fcGit(t, f.wt, "add", "-A")
	fcGit(t, f.wt, "commit", "-q", "-m", "card change")
	f.tip = fcGit(t, f.wt, "rev-parse", "HEAD")

	switch o.mainMoves {
	case "conflict":
		if err := os.WriteFile(filepath.Join(root, "base.txt"), []byte("base\nmain edit\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		fcGit(t, root, "commit", "-q", "-am", "main edits the same line")
	case "unrelated":
		if err := os.WriteFile(filepath.Join(root, "other.txt"), []byte("other\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		fcGit(t, root, "add", "other.txt")
		fcGit(t, root, "commit", "-q", "-m", "main adds an unrelated file")
	}
	if o.mainMoves != "" && !o.noOrigin {
		fcGit(t, root, "push", "-q", "origin", "main")
		f.mainBefore = fcGit(t, root, "rev-parse", "main")
	}

	state := o.state
	if state == "" {
		state = homestate.CardMergeReady
	}
	fcPlace(t, root, homestate.Card{
		CardID: "t1", State: state, Stage: homestate.CardMergeReady, OwnerLabel: ghfLane,
		LeaseHolder: ghfLane, LeaseExpiresAt: "2026-10-01T00:00:00Z", WorktreePath: f.wt,
	})
	db := fcOpen(t, root)
	if _, err := db.DB.Exec(`UPDATE cards SET spec_id=? WHERE card_id='t1'`, ghfSpec); err != nil {
		t.Fatalf("set spec id: %v", err)
	}
	_ = db.Close()
	sdLaneEnv(t, ghfLane, "")
	t.Setenv(config.EnvClaudeCodeSessionID, "")
	return f
}

// ghfSquashOnMain simulates the merge the PR's auto-merge performs: one squash
// commit of the card branch on origin/main. It returns that commit.
func (f ghfFixture) squashOnMain(t *testing.T) string {
	t.Helper()
	fcGit(t, f.root, "merge", "--squash", ghfBranch)
	fcGit(t, f.root, "commit", "-q", "-m", "squash of the card (#7)")
	fcGit(t, f.root, "push", "-q", "origin", "main")
	return fcGit(t, f.root, "rev-parse", "HEAD")
}

func (f ghfFixture) remoteBranchTip(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("git", "--git-dir", f.bare, "rev-parse", "--verify", "-q", "refs/heads/"+ghfBranch).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// ghDouble stands in for `gh`: a stateful pull request the verbs create, view
// and merge. It records every call, and every time it runs it reads the
// integration window so a test can say "never held, at any moment".
type ghDouble struct {
	t         *testing.T
	f         ghfFixture
	calls     [][]string
	created   bool
	state     string // OPEN | MERGED | CLOSED
	mergeOid  string
	headOid   string // the head the PR reports; default the card tip
	viewErr   error
	viewJunk  string
	mergeErr  error
	createErr error
	hang      bool // block until the call's context ends

	windowSeenHeld bool
	pushedAtCreate bool
}

func newGHDouble(t *testing.T, f ghfFixture) *ghDouble {
	t.Helper()
	d := &ghDouble{t: t, f: f, state: "OPEN", headOid: f.tip}
	prev := factoryGH
	factoryGH = d.run
	t.Cleanup(func() { factoryGH = prev })
	return d
}

func (d *ghDouble) count(sub ...string) int {
	n := 0
	for _, c := range d.calls {
		if len(c) >= len(sub) && strings.Join(c[:len(sub)], " ") == strings.Join(sub, " ") {
			n++
		}
	}
	return n
}

func (d *ghDouble) last(sub ...string) []string {
	for i := len(d.calls) - 1; i >= 0; i-- {
		c := d.calls[i]
		if len(c) >= len(sub) && strings.Join(c[:len(sub)], " ") == strings.Join(sub, " ") {
			return c
		}
	}
	return nil
}

func ghFlag(args []string, name string) string {
	for i, a := range args {
		if a == name && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func (d *ghDouble) run(ctx context.Context, _ string, args ...string) ([]byte, error) {
	d.calls = append(d.calls, append([]string(nil), args...))
	if lock, err := factory.ReadIntegrationLock(d.f.root); err == nil && lock.Held() {
		d.windowSeenHeld = true
	}
	if d.hang {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	if len(args) < 2 || args[0] != "pr" {
		return nil, fmt.Errorf("gh double: unexpected call %v", args)
	}
	switch args[1] {
	case "view":
		if d.viewErr != nil {
			return nil, d.viewErr
		}
		if d.viewJunk != "" {
			return []byte(d.viewJunk), nil
		}
		if !d.created && d.state == "OPEN" {
			return nil, errors.New(`no pull requests found for branch "` + ghfBranch + `"`)
		}
		merge := "null"
		if d.state == "MERGED" {
			merge = fmt.Sprintf(`{"oid":%q}`, d.mergeOid)
		}
		return []byte(fmt.Sprintf(`{"number":7,"url":%q,"state":%q,"baseRefName":"main","headRefName":%q,"headRefOid":%q,"mergeCommit":%s}`,
			ghfPRURL, d.state, ghfBranch, d.headOid, merge)), nil
	case "create":
		if d.createErr != nil {
			return nil, d.createErr
		}
		d.pushedAtCreate = d.f.remoteBranchTip(d.t) == d.f.tip
		d.created = true
		return []byte(ghfPRURL + "\n"), nil
	case "merge":
		return nil, d.mergeErr
	}
	return nil, fmt.Errorf("gh double: unexpected pr verb %v", args)
}

func ghfComplete(t *testing.T) (string, error) {
	t.Helper()
	out, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
	return out, err
}

func ghfNoWindow(t *testing.T, f ghfFixture, where string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(f.root, ".moai", "state", factory.IntegrationLockFileName)); err == nil {
		t.Errorf("%s: an integration window record exists", where)
	}
	if lock, err := factory.ReadIntegrationLock(f.root); err != nil || lock.Held() {
		t.Errorf("%s: the integration window is held (%+v, err %v)", where, lock, err)
	}
}

// AC-GFD-004 — the delivery edge: push, a PR against the integration target,
// auto-merge requested, `pr-open` recorded — and `merged-pr` only after the PR
// is observed MERGED. The homestate half (the predecessor pass and the `done`
// path of the new state) is the same-named test in internal/homestate.
func TestFactoryCompleteGitHubFlowPR(t *testing.T) {
	t.Run("pushes_the_branch_opens_a_pr_against_main_and_requests_auto_merge", func(t *testing.T) {
		f := ghfNew(t, ghfOpts{syncStatus: "complete"})
		d := newGHDouble(t, f)
		out, err := ghfComplete(t)
		if err != nil {
			t.Fatalf("complete: %v", err)
		}
		if got := f.remoteBranchTip(t); got != f.tip {
			t.Fatalf("origin %s = %q, want the card tip %s", ghfBranch, got, f.tip)
		}
		if !d.pushedAtCreate {
			t.Error("the PR was opened before the branch was on origin")
		}
		create := d.last("pr", "create")
		if create == nil {
			t.Fatalf("no `gh pr create` call; calls: %v", d.calls)
		}
		if ghFlag(create, "--base") != "main" || ghFlag(create, "--head") != ghfBranch {
			t.Errorf("pr create base/head = %q/%q, want main/%s", ghFlag(create, "--base"), ghFlag(create, "--head"), ghfBranch)
		}
		if title := ghFlag(create, "--title"); !strings.Contains(title, "t1") {
			t.Errorf("PR title %q does not carry the card id (the traceability carrier)", title)
		}
		body := ghFlag(create, "--body")
		if !strings.Contains(body, ".moai/reports/t1/") {
			t.Errorf("PR body does not state the evidence path:\n%s", body)
		}
		if !strings.HasSuffix(strings.TrimRight(body, "\n"), "\n🗿 MoAI") {
			t.Errorf("PR body does not end with the line 🗿 MoAI:\n%s", body)
		}
		merge := d.last("pr", "merge")
		if merge == nil || strings.Join(merge, " ") != "pr merge 7 --auto --squash --match-head-commit "+f.tip {
			t.Errorf("auto-merge request = %v, want `pr merge 7 --auto --squash --match-head-commit %s` (the readiness-judged tip, card-review r6)", merge, f.tip)
		}
		c := fcCard(t, f.root, "t1")
		if c.State != homestate.CardPROpen {
			t.Fatalf("card = %s, want pr-open", c.State)
		}
		if c.LeaseHolder != "" {
			t.Errorf("pr-open card still holds the lease for %q", c.LeaseHolder)
		}
		if !strings.Contains(out, ghfPRURL) || strings.Contains(out, "integration window") {
			t.Errorf("output = %q, want the PR URL and no window line", out)
		}
	})

	t.Run("not_recorded_merged_before_the_pr_is_observed_merged", func(t *testing.T) {
		f := ghfNew(t, ghfOpts{syncStatus: "complete"})
		d := newGHDouble(t, f)
		if _, err := ghfComplete(t); err != nil {
			t.Fatalf("first complete: %v", err)
		}
		if c := fcCard(t, f.root, "t1"); c.State != homestate.CardPROpen {
			t.Fatalf("after opening the PR the card is %s, want pr-open (never merged-pr yet)", c.State)
		}
		before := fcCard(t, f.root, "t1")
		// The PR is still open: nothing is recorded, and that is not a failure.
		out, err := ghfComplete(t)
		if err != nil {
			t.Fatalf("complete while the PR is open: %v", err)
		}
		if c := fcCard(t, f.root, "t1"); c.State != homestate.CardPROpen || c.Version != before.Version {
			t.Fatalf("an open PR moved the card: %s v%d → %s v%d", before.State, before.Version, c.State, c.Version)
		}
		if !strings.Contains(out, "open") {
			t.Errorf("output %q does not say the PR is still open", out)
		}
		if d.count("pr", "create") != 1 || d.count("pr", "merge") != 1 {
			t.Errorf("re-running complete opened or merge-requested again: create=%d merge=%d", d.count("pr", "create"), d.count("pr", "merge"))
		}
		// The PR merges: GitHub put one squash commit on main.
		d.state, d.mergeOid = "MERGED", f.squashOnMain(t)
		out, err = ghfComplete(t)
		if err != nil {
			t.Fatalf("complete after the merge: %v", err)
		}
		c := fcCard(t, f.root, "t1")
		if c.State != homestate.CardMergedPR || c.MergeSHA != d.mergeOid {
			t.Fatalf("card = %s merge=%s, want merged-pr with the squash commit %s", c.State, c.MergeSHA, d.mergeOid)
		}
		if !strings.Contains(out, "merged-pr") {
			t.Errorf("output %q does not report merged-pr", out)
		}
	})

	t.Run("gh_failures_are_cannot_confirm_and_record_nothing", func(t *testing.T) {
		cases := []struct {
			name  string
			setup func(t *testing.T, d *ghDouble, f ghfFixture)
		}{
			{"gh_exits_non_zero", func(_ *testing.T, d *ghDouble, _ ghfFixture) { d.viewErr = errors.New("gh: exit status 1") }},
			{"gh_times_out", func(_ *testing.T, d *ghDouble, _ ghfFixture) { d.hang = true; factoryGHTimeout = 50 * time.Millisecond }},
			{"unreadable_answer", func(_ *testing.T, d *ghDouble, _ ghfFixture) { d.viewJunk = "not json" }},
			{"merged_from_another_head", func(t *testing.T, d *ghDouble, f ghfFixture) {
				d.state, d.mergeOid, d.headOid = "MERGED", f.squashOnMain(t), strings.Repeat("a", 40)
			}},
			{"merged_without_a_merge_commit", func(_ *testing.T, d *ghDouble, _ ghfFixture) { d.state, d.mergeOid = "MERGED", "" }},
			{"merge_commit_not_on_origin", func(_ *testing.T, d *ghDouble, _ ghfFixture) { d.state, d.mergeOid = "MERGED", strings.Repeat("b", 40) }},
			{"closed_without_merging", func(_ *testing.T, d *ghDouble, _ ghfFixture) { d.state = "CLOSED" }},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				prev := factoryGHTimeout
				t.Cleanup(func() { factoryGHTimeout = prev })
				f := ghfNew(t, ghfOpts{syncStatus: "complete"})
				d := newGHDouble(t, f)
				if _, err := ghfComplete(t); err != nil {
					t.Fatalf("opening the PR: %v", err)
				}
				before := fcCard(t, f.root, "t1")
				events := fcEventCount(t, f.root, "card.transition")
				tc.setup(t, d, f)
				if _, err := ghfComplete(t); err == nil {
					t.Fatal("complete answered success without being able to confirm the merge")
				}
				after := fcCard(t, f.root, "t1")
				if after.State != homestate.CardPROpen || after.Version != before.Version || after.MergeSHA != "" {
					t.Errorf("a failed confirmation changed the record: %s v%d merge=%q", after.State, after.Version, after.MergeSHA)
				}
				if got := fcEventCount(t, f.root, "card.transition"); got != events {
					t.Errorf("a failed confirmation wrote %d event(s)", got-events)
				}
			})
		}
	})

	t.Run("no_git_merge_and_no_window_lock_file", func(t *testing.T) {
		f := ghfNew(t, ghfOpts{syncStatus: "complete"})
		newGHDouble(t, f)
		if _, err := ghfComplete(t); err != nil {
			t.Fatalf("complete: %v", err)
		}
		if merges := fcGit(t, f.root, "rev-list", "--merges", "--all"); merges != "" {
			t.Errorf("a merge commit exists after complete: %s", merges)
		}
		if got := fcGit(t, f.root, "rev-parse", "main"); got != f.mainBefore {
			t.Errorf("local main moved: %s → %s", f.mainBefore, got)
		}
		ghfNoWindow(t, f, "after complete")
	})
}

// complete is re-runnable from the states it leaves a card in.
func TestFactoryCompleteGitHubFlowResumesFromMerging(t *testing.T) {
	f := ghfNew(t, ghfOpts{syncStatus: "complete", state: homestate.CardMerging})
	d := newGHDouble(t, f)
	// A previous run opened the PR but died before recording pr-open.
	d.created = true
	if _, err := ghfComplete(t); err != nil {
		t.Fatalf("complete from merging: %v", err)
	}
	if c := fcCard(t, f.root, "t1"); c.State != homestate.CardPROpen {
		t.Fatalf("card = %s, want pr-open", c.State)
	}
	if d.count("pr", "create") != 0 {
		t.Error("an existing PR was opened a second time")
	}
}

// The merge method rides the configuration: git_strategy.manual.merge_method.
func TestFactoryCompleteGitHubFlowMergeMethod(t *testing.T) {
	for cfg, want := range map[string]string{
		"squash": "--squash", "merge": "--merge", "rebase": "--rebase", "": "--squash", "bogus": "--squash",
	} {
		t.Run("merge_method_"+cfg, func(t *testing.T) {
			f := ghfNew(t, ghfOpts{syncStatus: "complete", mergeCfg: cfg})
			d := newGHDouble(t, f)
			if _, err := ghfComplete(t); err != nil {
				t.Fatalf("complete: %v", err)
			}
			if got := strings.Join(d.last("pr", "merge"), " "); got != "pr merge 7 --auto "+want+" --match-head-commit "+f.tip {
				t.Errorf("merge request = %q, want %q", got, "pr merge 7 --auto "+want+" --match-head-commit "+f.tip)
			}
		})
	}
}

// AC-GFD-005 — the merge-readiness triple runs against the integration target
// BEFORE the PR exists; a failing triple opens nothing.
func TestMergeReadinessBeforePR(t *testing.T) {
	visited := 0
	cases := []struct {
		name    string
		o       ghfOpts
		failing string // "" = passes
	}{
		{"conflicting_card", ghfOpts{syncStatus: "complete", mainMoves: "conflict"}, "conflict-free"},
		{"conflict_free_card", ghfOpts{syncStatus: "complete"}, ""},
		{"no_sync_audit_pass_record", ghfOpts{syncStatus: ""}, "sync-audit"},
		{"broken_tree_identity", ghfOpts{syncStatus: "complete", mainMoves: "unrelated"}, "tree-identity"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			visited++
			f := ghfNew(t, tc.o)
			d := newGHDouble(t, f)
			out, err := ghfComplete(t)
			c := fcCard(t, f.root, "t1")
			if tc.failing == "" {
				if err != nil {
					t.Fatalf("a clean card was refused: %v", err)
				}
				if d.count("pr", "create") != 1 || c.State != homestate.CardPROpen {
					t.Fatalf("clean card: create calls=%d state=%s", d.count("pr", "create"), c.State)
				}
				return
			}
			if err == nil {
				t.Fatalf("complete passed a card failing %s", tc.failing)
			}
			if !strings.Contains(err.Error()+out, tc.failing) {
				t.Errorf("the refusal does not name %s:\nerr: %v\nout: %s", tc.failing, err, out)
			}
			if d.count("pr", "create") != 0 || len(d.calls) != 0 {
				t.Errorf("a failing triple still called gh: %v", d.calls)
			}
			if f.remoteBranchTip(t) != "" {
				t.Error("a failing triple still pushed the card branch")
			}
			if c.State != homestate.CardMergeReady {
				t.Errorf("a failing triple moved the card to %s", c.State)
			}
		})
	}
	if visited != 4 {
		t.Fatalf("visited %d fixtures, want 4", visited)
	}
}

// Pre-flight refusals fire before any record changes.
func TestFactoryCompleteGitHubFlowRefusals(t *testing.T) {
	t.Run("no_origin_remote", func(t *testing.T) {
		f := ghfNew(t, ghfOpts{syncStatus: "complete", noOrigin: true})
		d := newGHDouble(t, f)
		_, err := ghfComplete(t)
		if err == nil || !strings.Contains(err.Error(), "origin") {
			t.Fatalf("err = %v, want a refusal naming the missing origin remote", err)
		}
		if c := fcCard(t, f.root, "t1"); c.State != homestate.CardMergeReady || len(d.calls) != 0 {
			t.Errorf("card = %s, gh calls = %v: the refusal must precede every change", c.State, d.calls)
		}
	})
	t.Run("card_not_at_merge_ready", func(t *testing.T) {
		f := ghfNew(t, ghfOpts{syncStatus: "complete", state: homestate.CardRun})
		d := newGHDouble(t, f)
		_, err := ghfComplete(t)
		if err == nil || !strings.Contains(err.Error(), homestate.CardRun) {
			t.Fatalf("err = %v, want a refusal naming the card's state", err)
		}
		if len(d.calls) != 0 {
			t.Errorf("gh called: %v", d.calls)
		}
	})
	t.Run("codex_lane_refused_before_anything", func(t *testing.T) {
		f := ghfNew(t, ghfOpts{syncStatus: "complete"})
		sdLaneEnv(t, ghfLane, factory.BackendGPT)
		d := newGHDouble(t, f)
		if _, err := ghfComplete(t); err == nil {
			t.Fatal("a Codex lane completed a card")
		}
		if c := fcCard(t, f.root, "t1"); c.State != homestate.CardMergeReady || len(d.calls) != 0 {
			t.Errorf("card = %s, gh calls = %v: the Codex refusal must precede every change", c.State, d.calls)
		}
	})
}

// AC-GFD-006 — under github-flow the integration window is no prerequisite and
// nothing here takes it; a window held by another lane neither blocks complete
// nor is disturbed by it.
func TestFactoryCompleteNoWindowGitHubFlow(t *testing.T) {
	t.Run("never_held_at_any_moment", func(t *testing.T) {
		f := ghfNew(t, ghfOpts{syncStatus: "complete"})
		d := newGHDouble(t, f)
		if _, err := ghfComplete(t); err != nil {
			t.Fatalf("complete without a session id or a window: %v", err)
		}
		d.state, d.mergeOid = "MERGED", f.squashOnMain(t)
		if _, err := ghfComplete(t); err != nil {
			t.Fatalf("observing the merge: %v", err)
		}
		if d.windowSeenHeld {
			t.Error("the window was held while a gh call was in flight")
		}
		ghfNoWindow(t, f, "after both completes")
	})
	t.Run("another_lanes_window_is_neither_waited_for_nor_disturbed", func(t *testing.T) {
		f := ghfNew(t, ghfOpts{syncStatus: "complete"})
		newGHDouble(t, f)
		sdHoldWindow(t, f.root, "sess-lane-2", "lane-2", "main", factory.BranchSourceConfig, "", "t2")
		if _, err := ghfComplete(t); err != nil {
			t.Fatalf("complete with a foreign window held: %v", err)
		}
		lock := sdWindow(t, f.root)
		if !lock.Held() || lock.SessionID != "sess-lane-2" {
			t.Errorf("the foreign window changed: held=%v by %q", lock.Held(), lock.SessionID)
		}
	})
}

// AC-GFD-006 — git-flow keeps the window: complete takes it, merges --no-ff in
// the integration worktree, and never calls gh.
func TestFactoryCompleteWindowGitFlowUnchanged(t *testing.T) {
	root, integWT, cards := sdMergeFixture(t, true, true, false, 1)
	sdPlaceMergeReady(t, root, "t1", "lane-1", cards[0])
	before := factoryGHUnexpectedCalls.Load()
	sdLaneEnv(t, "lane-1", "")
	t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
	if _, _, err := runFactory(t, "complete", "t1", "--run", fcRun); err != nil {
		t.Fatalf("complete: %v", err)
	}
	c := fcCard(t, root, "t1")
	if c.State != homestate.CardMergedLocal || c.MergeSHA == "" {
		t.Fatalf("card = %s merge=%q, want merged-local", c.State, c.MergeSHA)
	}
	if got := fcGit(t, integWT, "rev-list", "--parents", "-n", "1", "HEAD"); len(strings.Fields(got)) != 3 {
		t.Errorf("integration HEAD is not a two-parent merge: %s", got)
	}
	if lock := sdWindow(t, root); !lock.Held() || lock.SessionID != "sess-lane-1" {
		t.Errorf("window after complete = held=%v by %q, want held by sess-lane-1", lock.Held(), lock.SessionID)
	}
	if got := factoryGHUnexpectedCalls.Load(); got != before {
		t.Errorf("git-flow complete called gh %d time(s)", got-before)
	}
}

// The stage verb is not a back door: the PR states need evidence it cannot carry.
func TestFactoryStageVerbCannotMintPRStates(t *testing.T) {
	f := ghfNew(t, ghfOpts{syncStatus: "complete", state: homestate.CardMerging})
	// The branch IS pushed: the only thing the stage verb cannot supply is the PR identity.
	fcGit(t, f.wt, "push", "-q", "origin", ghfBranch)
	ctx := context.Background()
	if _, err := factoryStageCard(ctx, f.root, "t1", homestate.CardPROpen, "", fcRun, ghfLane); err == nil {
		t.Fatal("`factory stage … pr-open` recorded a PR state without a PR identity")
	}
	if c := fcCard(t, f.root, "t1"); c.State != homestate.CardMerging {
		t.Fatalf("card = %s, want merging unchanged", c.State)
	}
}

// The operator's push gate closes a PR-merged card (there is nothing to push).
func TestFactoryDecidePushGateClosesAMergedPRCard(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardMergedPR})
	db := fcOpen(t, root)
	got, err := decideOne(context.Background(), db, fcRun, "t1", "push", "", "")
	if err != nil {
		t.Fatalf("decide --gate push on a merged-pr card: %v", err)
	}
	if got.State != homestate.CardDone {
		t.Fatalf("card = %s, want done", got.State)
	}
}

// `factory merge ready` is the pre-PR check's entry too; under github-flow it
// takes no window.
func TestFactoryMergeReadyGitHubFlowTakesNoWindow(t *testing.T) {
	_, lockRoot, specID := mergeReadyFixture(t, "complete", "lane-9", "sess-lane-9")
	ghfWriteConfig(t, lockRoot, "")
	out, err := runFactoryMerge(t, "merge", "ready", "--card", "t9001", "--spec", specID, "--branch", "WT-card", "--develop", "develop", "--session", "sess-lane-9", "--json")
	if err != nil {
		t.Fatalf("merge ready: %v", err)
	}
	if !strings.Contains(out, `"verdict":"cleared"`) {
		t.Fatalf("ready did not clear: %s", out)
	}
	if lock := readLockRoot(t, lockRoot); lock.Held() {
		t.Errorf("merge ready took the window under github-flow: %+v", lock)
	}
}

// The two cli-side state consumers: a PR-delivered card has finished the lane's
// implementation work (the serial slot re-admits selection), and both of its
// states sit past merge-ready (a Codex lane never selects such a card).
func TestFactoryGitHubFlowStatesAreConsumedAsPostMergeReady(t *testing.T) {
	for _, s := range []string{homestate.CardPROpen, homestate.CardMergedPR} {
		if !factorySerialSlotFree(s) {
			t.Errorf("factorySerialSlotFree(%s) = false: an open PR must not hold the serial slot", s)
		}
		if !cardStageAtOrAfterMergeReady(s) {
			t.Errorf("cardStageAtOrAfterMergeReady(%s) = false", s)
		}
	}
	if factorySerialSlotFree(homestate.CardRun) {
		t.Error("factorySerialSlotFree(run) = true: the slot must stay held while a card is being worked")
	}
}

// TestFactoryCompleteGitHubFlowClosesTheQueueCard — the runtime completion's
// archive authority (card t1542, the audit P1): a card the record moves to
// merged-pr has mechanically answered the landing question, so the delivery
// edge closes the lane's own queue card (archive + landing verdict) and
// records the runtime completion row — the lane lands its own card without a
// leader `todo done`. A re-run reconciles: already closed reads as closed.
func TestFactoryCompleteGitHubFlowClosesTheQueueCard(t *testing.T) {
	f := ghfNew(t, ghfOpts{syncStatus: "complete"})
	d := newGHDouble(t, f)
	if _, err := ghfComplete(t); err != nil {
		t.Fatalf("first complete: %v", err)
	}
	d.state, d.mergeOid = "MERGED", f.squashOnMain(t)
	if _, err := ghfComplete(t); err != nil {
		t.Fatalf("complete after the merge: %v", err)
	}
	if c := fcCard(t, f.root, "t1"); c.State != homestate.CardMergedPR {
		t.Fatalf("card = %s, want merged-pr", c.State)
	}
	rec, err := factory.NewBacklogStore(todoBacklogPath(f.root)).LoadPure()
	if err != nil {
		t.Fatalf("load queue: %v", err)
	}
	if len(rec.Items) != 0 {
		t.Fatalf("the queue still holds %d live card(s) after merged-pr", len(rec.Items))
	}
	if len(rec.Archived) != 1 || rec.Archived[0].Item.ID != "t1" {
		t.Fatalf("archive holds %d entries, want exactly t1", len(rec.Archived))
	}
	if v := rec.Archived[0].LandingVerdict; v == nil || v.Verdict != factory.LandingLanded || v.Ref != "main" {
		t.Fatalf("landing verdict = %+v, want landed against main", v)
	}
	found := false
	for _, a := range rec.Runtime.Assignments {
		if a.CardID == "t1" && a.EventKind == "card.completed" {
			found = true
		}
	}
	if !found {
		t.Fatalf("no runtime completion row for t1: %+v", rec.Runtime.Assignments)
	}
	// Reconciliation: a re-run of complete stays green and changes nothing.
	if _, err := ghfComplete(t); err != nil {
		t.Fatalf("reconciling re-run: %v", err)
	}
	rec, err = factory.NewBacklogStore(todoBacklogPath(f.root)).LoadPure()
	if err != nil {
		t.Fatalf("reload queue: %v", err)
	}
	if len(rec.Items) != 0 || len(rec.Archived) != 1 {
		t.Fatalf("reconcile moved the queue: live=%d archived=%d", len(rec.Items), len(rec.Archived))
	}
}

// TestFactoryCompleteRefusesExpiredLeaseOnDelivery — card-review r5: the
// deliver half pushes, opens the PR, and asks for auto-merge — EXTERNAL
// mutations a retry must not run on a lapsed lease. An expired-lease card
// at merge-ready is refused before anything reaches the remote; the
// observation half (pr-open, lease released) is unaffected by design.
func TestFactoryCompleteRefusesExpiredLeaseOnDelivery(t *testing.T) {
	f := ghfNew(t, ghfOpts{syncStatus: "complete"})
	d := newGHDouble(t, f)
	// fcFixture pins factoryCardNow to 2026-09-26; the fixture lease
	// (2026-10-01) is therefore live — expire it in place.
	db := fcOpen(t, f.root)
	if _, err := db.DB.Exec(`UPDATE cards SET lease_expires_at='2026-09-01T00:00:00Z' WHERE card_id='t1'`); err != nil {
		t.Fatalf("expire the lease: %v", err)
	}
	_ = db.Close()
	sdLaneEnv(t, ghfLane, "")
	if _, err := ghfComplete(t); err == nil {
		t.Fatal("complete delivered to the remote on an expired lease")
	}
	if d.count("pr", "create") != 0 || d.count("pr", "merge") != 0 {
		t.Errorf("the refused run still called gh: create=%d merge=%d", d.count("pr", "create"), d.count("pr", "merge"))
	}
	if got := f.remoteBranchTip(t); got != "" {
		t.Errorf("the refused run pushed %s to origin (tip %q)", ghfBranch, got)
	}
	if c := fcCard(t, f.root, "t1"); c.State != homestate.CardMergeReady {
		t.Fatalf("the refused run moved the card to %s", c.State)
	}
}

// TestFactoryCompleteRefusesAnotherLanesCard — card-review r1: a lane
// completes only its own card. lane-2 running complete on lane-1's card is
// refused at entry, and lane-1's queue card stays live.
func TestFactoryCompleteRefusesAnotherLanesCard(t *testing.T) {
	f := ghfNew(t, ghfOpts{syncStatus: "complete"})
	newGHDouble(t, f)
	sdLaneEnv(t, "lane-2", "")
	if _, err := ghfComplete(t); err == nil {
		t.Fatal("lane-2 completed lane-1's card")
	}
	rec, err := factory.NewBacklogStore(todoBacklogPath(f.root)).LoadPure()
	if err != nil {
		t.Fatalf("load queue: %v", err)
	}
	if len(rec.Items) != 1 {
		t.Fatalf("lane-1's queue card did not stay live: live=%d archived=%d", len(rec.Items), len(rec.Archived))
	}
	if c := fcCard(t, f.root, "t1"); c.State == homestate.CardPROpen || c.State == homestate.CardMergedPR {
		t.Fatalf("lane-2's complete moved the card to %s", c.State)
	}
}

// TestFactoryCloseLaneCardUsesTheCallerRoot — card-review r1: the close
// path archives the queue of the ROOT THE CALLER NAMED, not the queue this
// process's working directory resolves to (the MCP surface passes a
// project_root that differs from the server cwd).
func TestFactoryCloseLaneCardUsesTheCallerRoot(t *testing.T) {
	// Seeding runs `todo add`, refused when the test process carries lane
	// variables (a lane session running the suite locally — card t1542).
	sdClearLaneEnv(t)
	// Project A owns the card and receives the close; the process cwd will
	// name project B, whose same-id card must stay live.
	rootA, storeA := todoFixture(t)
	if _, _, err := runTodo(t, "add", "project A card"); err != nil {
		t.Fatalf("seed A: %v", err)
	}
	rootB := t.TempDir()
	initGitRepo(t, rootB)
	storeB := factory.NewBacklogStore(todoBacklogPath(rootB))
	t.Setenv("CLAUDE_PROJECT_DIR", rootB)
	if _, _, err := runTodo(t, "add", "project B card"); err != nil {
		t.Fatalf("seed B: %v", err)
	}

	var out strings.Builder
	card := homestate.Card{CardID: "t1", OwnerLabel: "lane-1", LeaseHolder: "lane-1"}
	factoryCloseLaneCard(&out, rootA, fcRun, card, "lane-1", "main")

	recA, err := storeA.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(recA.Items) != 0 || len(recA.Archived) != 1 || recA.Archived[0].Item.ID != "t1" {
		t.Fatalf("project A queue after close: live=%d archived=%d, want t1 archived", len(recA.Items), len(recA.Archived))
	}
	recB, err := storeB.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(recB.Items) != 1 || len(recB.Archived) != 0 {
		t.Fatalf("project B queue changed: live=%d archived=%d, want untouched", len(recB.Items), len(recB.Archived))
	}
}
