package cli

// factory_aff_s2_t2013_test.go — card t2013, Git-AFF S2: REQ-FCR-026 / AC-FCR-026
// (a)-(f) of SPEC-FACTORY-COMPLETION-RECOVERY-001 (spec.md section D.5).
//
// Under agentic-trunk, `moai factory complete` lands a card through the local
// integration window into the FIXED release branch (release_branch_prefix +
// "main-batch"), after the fast required check (`moai verify run`), never into
// main, and a rerun for the same card HEAD is a no-op. Every case runs real
// git (a temporary repository with a bare origin, real branches and SHAs), the
// real merge step, and the real `moai verify run` code path against a small Go
// module in the card tree — nothing about merge, check, or git is mocked. The
// factory database is a temporary one (fcFixture).
//
// Case (g), the regression set, is the existing tests run unchanged; the last
// test here additionally pins that a project without a provisioned release
// branch keeps the main route.

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorylane"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/verify"
)

const (
	affs2ReleaseBranch = "release/main-batch"
	affs2CardBranch    = "WT-alpha"
	affs2Session       = "sess-lane-1"
)

// affs2Opts selects the fixture variants.
type affs2Opts struct {
	// broken commits a Go file that does not compile into the CARD tree only.
	broken bool
	// unprovisioned leaves the release branch checked out nowhere and gives
	// main to a dedicated worktree (the pre-amendment local-merge shape).
	unprovisioned bool
	// slowTest commits a passing test that sleeps for slowTestPause into the
	// CARD tree, so the fast required check takes longer than that.
	slowTest bool
}

// slowTestPause is how long the slowTest fixture's test sleeps.
const slowTestPause = 3

// affs2Fix is one built fixture.
type affs2Fix struct {
	root, relWT, cardWT string
	relBranch           string // the branch the card is expected to land in
	cardHead            string
	mainBefore          string
	relBefore           string
	remotesBefore       string
	ghCalls             *int
}

func affs2Git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	return fcGit(t, dir, args...)
}

func affs2Write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// affs2NoGH installs a gh stand-in that fails the test: the release route opens
// no pull request, and a route that does must not reach a real gh.
func affs2NoGH(t *testing.T) *int {
	t.Helper()
	calls := 0
	prev := factoryGH
	factoryGH = func(ctx context.Context, dir string, args ...string) ([]byte, error) {
		calls++
		return nil, errors.New("gh is not available in this test")
	}
	t.Cleanup(func() { factoryGH = prev })
	return &calls
}

// affs2New builds the AC-FCR-026 common fixture: main held by the primary
// checkout, the fixed release branch held by a dedicated worktree, one
// merge-ready card whose worktree branched from the release branch and whose
// tree is a small Go module, a bare origin, and a valid re-measure record.
func affs2New(t *testing.T, o affs2Opts) affs2Fix {
	t.Helper()
	sdClearLaneEnv(t)
	root, store := fcFixture(t)
	// The merge step reads the real clock for the card lease, so the lease
	// renewals complete performs while the check runs must write real time too
	// (fcFixture pins the factory clock to a past date).
	prevNow := factoryCardNow
	factoryCardNow = time.Now
	t.Cleanup(func() { factoryCardNow = prevNow })
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	fcQueue(t, store, factory.BacklogStatePicked)
	f := affs2Fix{root: root, ghCalls: affs2NoGH(t)}

	affs2Git(t, root, "branch", "-M", "main")
	affs2Write(t, filepath.Join(root, "go.mod"), "module example.com/affs2\n\ngo 1.21\n")
	affs2Write(t, filepath.Join(root, "calc", "calc.go"), "package calc\n\nfunc Add(a, b int) int { return a + b }\n")
	affs2Write(t, filepath.Join(root, "calc", "calc_test.go"), "package calc\n\nimport \"testing\"\n\nfunc TestAdd(t *testing.T) {\n\tif Add(1, 2) != 3 {\n\t\tt.Fatal(\"add\")\n\t}\n}\n")
	affs2Write(t, filepath.Join(root, "base.txt"), "line1\nline2\nline3\n")
	affs2Git(t, root, "add", "go.mod", "calc", "base.txt")
	affs2Git(t, root, "commit", "-q", "-m", "base module")

	bare := filepath.Join(t.TempDir(), "origin.git")
	affs2Git(t, filepath.Dir(bare), "init", "-q", "--bare", "-b", "main", bare)
	affs2Git(t, root, "remote", "add", "origin", bare)
	affs2Git(t, root, "push", "-q", "origin", "main")

	writeAgenticTrunkConfig(t, root) // workflow agentic-trunk, release_branch_prefix: release/
	affs2Git(t, root, "branch", affs2ReleaseBranch)
	f.relBranch = affs2ReleaseBranch
	if o.unprovisioned {
		// main moves to a dedicated worktree (the pre-amendment shape) and the
		// release branch exists but is checked out nowhere.
		affs2Git(t, root, "checkout", "-q", "-b", "wtroot")
		mainWT := filepath.Join(root, ".claude", "worktrees", "main")
		affs2Git(t, root, "worktree", "add", "-q", mainWT, "main")
		f.relBranch = "main"
		f.relWT = mainWT
	} else {
		f.relWT = filepath.Join(root, ".claude", "worktrees", "release-main-batch")
		affs2Git(t, root, "worktree", "add", "-q", f.relWT, affs2ReleaseBranch)
	}

	f.cardWT = filepath.Join(root, ".claude", "worktrees", "t1")
	affs2Git(t, root, "worktree", "add", "-q", "-b", affs2CardBranch, f.cardWT, f.relBranch)
	affs2Write(t, filepath.Join(f.cardWT, "base.txt"), "line1\ncard line2\nline3\n")
	affs2Write(t, filepath.Join(f.cardWT, "calc", "card.go"), "package calc\n\nfunc Sub(a, b int) int { return a - b }\n")
	if o.broken {
		affs2Write(t, filepath.Join(f.cardWT, "calc", "broken.go"), "package calc\n\nfunc Broken() int { return undefinedSymbol }\n")
	}
	if o.slowTest {
		affs2Write(t, filepath.Join(f.cardWT, "calc", "slow_test.go"), "package calc\n\nimport (\n\t\"testing\"\n\t\"time\"\n)\n\nfunc TestSlow(t *testing.T) { time.Sleep("+strconv.Itoa(slowTestPause)+" * time.Second) }\n")
	}
	affs2Git(t, f.cardWT, "add", "base.txt", "calc")
	affs2Git(t, f.cardWT, "commit", "-q", "-m", "card change alpha")
	f.cardHead = affs2Git(t, f.cardWT, "rev-parse", "HEAD")

	sdPlaceMergeReady(t, root, "t1", "lane-1", sdCardTree{wt: f.cardWT, branch: affs2CardBranch})
	if _, err := factory.RunRemeasure(root, f.cardWT, f.relBranch, "true"); err != nil {
		t.Fatalf("place re-measure record: %v", err)
	}
	sdLaneEnv(t, "lane-1", "")
	t.Setenv(config.EnvClaudeCodeSessionID, affs2Session)

	f.mainBefore = affs2Git(t, root, "rev-parse", "main")
	f.relBefore = affs2Git(t, root, "rev-parse", affs2ReleaseBranch)
	f.remotesBefore = affs2Git(t, root, "for-each-ref", "refs/remotes")
	return f
}

// affs2Holds returns the card.hold events the factory database holds.
func affs2Holds(t *testing.T, root string) []map[string]any {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	rows, err := db.DB.Query(`SELECT payload_json FROM events WHERE kind='card.hold' ORDER BY seq`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var holds []map[string]any
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			t.Fatalf("hold payload %q: %v", raw, err)
		}
		holds = append(holds, m)
	}
	return holds
}

// affs2Untouched asserts main and every remote-tracking ref kept the SHA they
// had before the invocation, and that no pull request was requested.
func affs2Untouched(t *testing.T, f affs2Fix, where string) {
	t.Helper()
	if got := affs2Git(t, f.root, "rev-parse", "main"); got != f.mainBefore {
		t.Errorf("%s: main moved %s -> %s", where, f.mainBefore, got)
	}
	if got := affs2Git(t, f.root, "for-each-ref", "refs/remotes"); got != f.remotesBefore {
		t.Errorf("%s: refs/remotes changed:\n%s\n-->\n%s", where, f.remotesBefore, got)
	}
	if *f.ghCalls != 0 {
		t.Errorf("%s: gh was called %d time(s); the release route opens no pull request", where, *f.ghCalls)
	}
}

func affs2WindowReleased(t *testing.T, f affs2Fix, where string) {
	t.Helper()
	if lock := sdWindow(t, f.root); lock.Held() {
		t.Errorf("%s: window still held by %q (branch %q)", where, lock.SessionID, lock.Branch)
	}
}

// (a) A completed card lands as a --no-ff merge into the release branch.
func TestAFFS2CompleteMergesCardIntoReleaseBranch(t *testing.T) {
	f := affs2New(t, affs2Opts{})
	if _, _, err := runFactory(t, "complete", "t1", "--run", fcRun); err != nil {
		t.Fatalf("complete: %v", err)
	}
	c := fcCard(t, f.root, "t1")
	relHead := affs2Git(t, f.relWT, "rev-parse", "HEAD")
	if c.State != homestate.CardMergedLocal || c.MergeSHA != relHead || relHead == f.relBefore {
		t.Fatalf("card = %s merge=%q, release HEAD %s (before %s), want merged-local at the new release HEAD", c.State, c.MergeSHA, relHead, f.relBefore)
	}
	parents := strings.Fields(affs2Git(t, f.relWT, "rev-list", "--parents", "-n", "1", "HEAD"))
	if len(parents) != 3 || parents[1] != f.relBefore || parents[2] != f.cardHead {
		t.Errorf("release HEAD parents = %v, want a --no-ff merge of the card HEAD %s onto %s", parents, f.cardHead, f.relBefore)
	}
	affs2Untouched(t, f, "after the release-route complete")
	affs2WindowReleased(t, f, "after the release-route complete")

	// The fast required check left its record under the CARD tree's key.
	key, err := verify.Key(context.Background(), f.cardWT)
	if err != nil {
		t.Fatal(err)
	}
	store, err := auditreceipt.StoreRoot(f.cardWT)
	if err != nil {
		t.Fatal(err)
	}
	snap, err := verify.Load(store, key)
	if err != nil || snap == nil || len(snap.Checks) == 0 {
		t.Fatalf("no moai verify run record under the card tree key %s (err %v, snapshot %v)", key, err, snap)
	}
	for _, check := range snap.Checks {
		if check.Verdict != "pass" {
			t.Errorf("recorded check %q verdict = %q, want pass", check.Command, check.Verdict)
		}
	}
}

// (b) A failing fast check holds the card and merges nothing.
func TestAFFS2CompleteCheckFailureHoldsAndReleases(t *testing.T) {
	f := affs2New(t, affs2Opts{broken: true})
	_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
	if err == nil {
		t.Fatal("complete succeeded although the card tree does not compile")
	}
	t.Logf("complete refused: %v", err)
	if got := affs2Git(t, f.relWT, "rev-parse", "HEAD"); got != f.relBefore {
		t.Errorf("release branch moved %s -> %s although the check failed", f.relBefore, got)
	}
	if c := fcCard(t, f.root, "t1"); c.State != homestate.CardMergeReady || c.Version != 1 {
		t.Errorf("card = %s v%d, want merge-ready v1 (a hold record, not a transition)", c.State, c.Version)
	}
	holds := affs2Holds(t, f.root)
	if len(holds) != 1 {
		t.Fatalf("hold records = %d, want 1: %v", len(holds), holds)
	}
	h := holds[0]
	if h["reason"] != "check-failed" || h["card_head"] != f.cardHead || h["check_key"] == nil || h["check_key"] == "" {
		t.Errorf("hold = %v, want reason check-failed, card_head %s and a check_key", h, f.cardHead)
	}
	affs2WindowReleased(t, f, "after the failed check")
	affs2Untouched(t, f, "after the failed check")
}

// (c) A diverged release branch (conflict or the gate that fires before it)
// leaves the release worktree clean and holds the card.
func TestAFFS2CompleteConflictHoldsReleasesAndCleans(t *testing.T) {
	f := affs2New(t, affs2Opts{})
	// The release branch gains a commit editing the SAME line the card edited.
	affs2Write(t, filepath.Join(f.relWT, "base.txt"), "line1\nrelease line2\nline3\n")
	affs2Git(t, f.relWT, "add", "base.txt")
	affs2Git(t, f.relWT, "commit", "-q", "-m", "release edits line2")
	relMoved := affs2Git(t, f.relWT, "rev-parse", "HEAD")

	_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
	if err == nil {
		t.Fatal("complete succeeded although the release branch diverged from the card")
	}
	t.Logf("complete stopped at: %v", err) // the gate or conflict that actually fired
	if got := affs2Git(t, f.relWT, "rev-parse", "HEAD"); got != relMoved {
		t.Errorf("release branch moved %s -> %s", relMoved, got)
	}
	if out := affs2Git(t, f.relWT, "status", "--porcelain"); out != "" {
		t.Errorf("release worktree not clean:\n%s", out)
	}
	merge := exec.Command("git", "-C", f.relWT, "rev-parse", "-q", "--verify", "MERGE_HEAD")
	if out, mergeErr := merge.Output(); mergeErr == nil || strings.TrimSpace(string(out)) != "" {
		t.Errorf("MERGE_HEAD exists in the release worktree: %q (err %v)", out, mergeErr)
	}
	if c := fcCard(t, f.root, "t1"); c.State != homestate.CardMergeReady || c.Version != 1 {
		t.Errorf("card = %s v%d, want merge-ready v1", c.State, c.Version)
	}
	holds := affs2Holds(t, f.root)
	if len(holds) != 1 {
		t.Fatalf("hold records = %d, want 1: %v", len(holds), holds)
	}
	h := holds[0]
	paths, _ := h["conflict_paths"].([]any)
	if h["reason"] == nil || h["reason"] == "" || h["card_head"] != f.cardHead || len(paths) != 1 || paths[0] != "base.txt" {
		t.Errorf("hold = %v, want a reason, card_head %s and conflict_paths [base.txt]", h, f.cardHead)
	}
	affs2WindowReleased(t, f, "after the diverged release branch")
	affs2Untouched(t, f, "after the diverged release branch")
}

// (c, real conflict) The test above stops at a pre-merge gate (the release tip
// moved before complete began), so it never reaches git's own conflict. This one
// does: the release tip is moved by a commit that lands after the merge step's
// last tip re-read and immediately before its `git merge` (the merge-step Git
// seam, installed through factoryCompleteMergeSeamsHook), editing the same line
// the card edited. Git stops the merge at a real conflict, the step aborts it,
// and complete records the stop as a merge-conflict hold naming the diverged path.
func TestAFFS2CompleteRealMergeConflictHoldsAbortsAndReleases(t *testing.T) {
	f := affs2New(t, affs2Opts{})
	real := factorylane.ExecGitRunner{Dir: f.relWT}
	var relMoved, mergeOut string
	var mergeErr error
	factoryCompleteMergeSeamsHook = func(s *factory.MergeStepSeams) {
		s.Git = func(args ...string) (string, error) {
			isMerge := len(args) >= 1 && args[0] == "merge" && !slices.Contains(args, "--abort")
			if isMerge && relMoved == "" {
				affs2Write(t, filepath.Join(f.relWT, "base.txt"), "line1\nrelease line2\nline3\n")
				affs2Git(t, f.relWT, "add", "base.txt")
				affs2Git(t, f.relWT, "commit", "-q", "-m", "release edits line2 after the gates")
				relMoved = affs2Git(t, f.relWT, "rev-parse", "HEAD")
			}
			out, err := real.Git(args...)
			if isMerge {
				mergeOut, mergeErr = out, err
			}
			return out, err
		}
	}
	t.Cleanup(func() { factoryCompleteMergeSeamsHook = nil })

	_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
	if err == nil {
		t.Fatal("complete succeeded although the release branch gained a conflicting commit")
	}
	var coded *exitCodeError
	if !errors.As(err, &coded) || coded.code != factory.MergeExitMergeFailed {
		t.Fatalf("complete error = %v, want the merge-failed exit code %d", err, factory.MergeExitMergeFailed)
	}

	// The stop was git's own conflict, not a gate.
	var exit *factorylane.GitExitError
	if !errors.As(mergeErr, &exit) || exit.ExitCode != 1 || !strings.Contains(mergeOut+exit.Stdout+exit.Stderr, "CONFLICT") {
		t.Fatalf("the merge call = (%q, %v), want git exit 1 reporting a CONFLICT", mergeOut, mergeErr)
	}
	if relMoved == "" || relMoved == f.relBefore {
		t.Fatalf("the release tip never moved before the merge (before %s, moved %s)", f.relBefore, relMoved)
	}

	if got := affs2Git(t, f.relWT, "rev-parse", "HEAD"); got != relMoved {
		t.Errorf("release branch moved %s -> %s although the merge failed", relMoved, got)
	}
	if out := affs2Git(t, f.relWT, "status", "--porcelain"); out != "" {
		t.Errorf("release worktree not clean after the abort:\n%s", out)
	}
	merge := exec.Command("git", "-C", f.relWT, "rev-parse", "-q", "--verify", "MERGE_HEAD")
	if out, mergeHeadErr := merge.Output(); mergeHeadErr == nil || strings.TrimSpace(string(out)) != "" {
		t.Errorf("MERGE_HEAD exists in the release worktree: %q (err %v)", out, mergeHeadErr)
	}
	if c := fcCard(t, f.root, "t1"); c.State != homestate.CardMergeReady || c.Version != 1 {
		t.Errorf("card = %s v%d, want merge-ready v1", c.State, c.Version)
	}
	holds := affs2Holds(t, f.root)
	if len(holds) != 1 {
		t.Fatalf("hold records = %d, want 1: %v", len(holds), holds)
	}
	h := holds[0]
	paths, _ := h["conflict_paths"].([]any)
	if h["reason"] != "merge-conflict" || h["card_head"] != f.cardHead || len(paths) != 1 || paths[0] != "base.txt" {
		t.Errorf("hold = %v, want reason merge-conflict, card_head %s and conflict_paths [base.txt]", h, f.cardHead)
	}
	affs2WindowReleased(t, f, "after the real merge conflict")
	affs2Untouched(t, f, "after the real merge conflict")
}

// (d) The same card HEAD completed twice merges once.
func TestAFFS2CompleteSameSHARerunIsNoOp(t *testing.T) {
	f := affs2New(t, affs2Opts{})
	if _, _, err := runFactory(t, "complete", "t1", "--run", fcRun); err != nil {
		t.Fatalf("first complete: %v", err)
	}
	first := fcCard(t, f.root, "t1")
	relAfter := affs2Git(t, f.relWT, "rev-parse", "HEAD")
	countAfter := affs2Git(t, f.relWT, "rev-list", "--count", affs2ReleaseBranch)

	out, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
	if err != nil {
		t.Fatalf("rerun complete: %v (a rerun with the same card HEAD must exit 0)", err)
	}
	if got := affs2Git(t, f.relWT, "rev-parse", "HEAD"); got != relAfter {
		t.Errorf("release branch moved %s -> %s on the rerun", relAfter, got)
	}
	if got := affs2Git(t, f.relWT, "rev-list", "--count", affs2ReleaseBranch); got != countAfter {
		t.Errorf("release commit count %s -> %s on the rerun", countAfter, got)
	}
	cardMerges := 0
	for _, line := range strings.Split(affs2Git(t, f.relWT, "rev-list", "--merges", "--parents", affs2ReleaseBranch), "\n") {
		if fields := strings.Fields(line); len(fields) == 3 && fields[2] == f.cardHead {
			cardMerges++
		}
	}
	if cardMerges != 1 {
		t.Errorf("merge commits carrying the card HEAD = %d, want exactly 1", cardMerges)
	}
	second := fcCard(t, f.root, "t1")
	if second.State != first.State || second.Version != first.Version || second.MergeSHA != first.MergeSHA {
		t.Errorf("card changed on the rerun: %s v%d %s -> %s v%d %s", first.State, first.Version, first.MergeSHA, second.State, second.Version, second.MergeSHA)
	}
	if !strings.Contains(out, first.MergeSHA) {
		t.Errorf("rerun output does not name the existing merge %s:\n%s", first.MergeSHA, out)
	}
	affs2WindowReleased(t, f, "after the rerun")
	affs2Untouched(t, f, "after the rerun")
}

// (e) A merge made but never recorded is adopted, not repeated.
func TestAFFS2CompleteAdoptsMergeAfterCrash(t *testing.T) {
	f := affs2New(t, affs2Opts{})
	// The merge verb merged the card (the crash happened before complete
	// recorded it): the merge step, then the re-measure of the merged tree.
	sdHoldWindow(t, f.root, affs2Session, "lane-1", f.relBranch, factory.BranchSourceConfig, f.relWT, "t1")
	mergeSHA, err := factory.RunMergeStep(factory.MergeStepInput{
		Root: f.root, IntegrationWorktree: f.relWT, IntegrationBranch: f.relBranch,
		CardID: "t1", CallerSessionID: affs2Session,
	}, factory.MergeStepSeams{ReadCard: func(id string) (factory.MergeCardState, error) {
		return integrationReadMergeCardForRun(context.Background(), f.root, fcRun, id, "lane-1")
	}})
	if err != nil {
		t.Fatalf("the merge verb's step: %v", err)
	}
	if _, err := factory.RunRemeasure(f.root, f.relWT, f.relBranch, "true"); err != nil {
		t.Fatalf("re-measure the merged tree: %v", err)
	}
	if c := fcCard(t, f.root, "t1"); c.State != homestate.CardMergeReady {
		t.Fatalf("card = %s before complete, want merge-ready (the crash window)", c.State)
	}

	if _, _, err := runFactory(t, "complete", "t1", "--run", fcRun); err != nil {
		t.Fatalf("complete after the crash: %v", err)
	}
	c := fcCard(t, f.root, "t1")
	if c.State != homestate.CardMergedLocal || c.MergeSHA != mergeSHA {
		t.Errorf("card = %s merge=%q, want merged-local adopting %s", c.State, c.MergeSHA, mergeSHA)
	}
	if got := affs2Git(t, f.relWT, "rev-parse", "HEAD"); got != mergeSHA {
		t.Errorf("release HEAD = %s, want the adopted merge %s (no new commit)", got, mergeSHA)
	}
	affs2WindowReleased(t, f, "after the adoption")
	affs2Untouched(t, f, "after the adoption")
}

// (f) A second completer never merges while another session holds the window.
func TestAFFS2CompleteSecondCompleterBlocked(t *testing.T) {
	f := affs2New(t, affs2Opts{})
	sdHoldWindow(t, f.root, "sess-other", "lane-2", f.relBranch, factory.BranchSourceConfig, f.relWT, "t9")
	windowBefore := sdWindow(t, f.root)
	cardBefore := fcCard(t, f.root, "t1")

	_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
	if err == nil || !strings.Contains(err.Error(), "lane-2") {
		t.Fatalf("complete = %v, want a refusal naming the holder lane-2", err)
	}
	if got := affs2Git(t, f.relWT, "rev-parse", "HEAD"); got != f.relBefore {
		t.Errorf("release branch moved %s -> %s while another session held the window", f.relBefore, got)
	}
	sdCardUnchanged(t, "second completer", f.root, "t1", cardBefore)
	after := sdWindow(t, f.root)
	if !after.Held() || after.SessionID != windowBefore.SessionID || after.Branch != windowBefore.Branch || after.Card != windowBefore.Card {
		t.Errorf("window record changed: %+v -> %+v", windowBefore, after)
	}
	if holds := affs2Holds(t, f.root); len(holds) != 0 {
		t.Errorf("a refused second completer wrote hold records: %v", holds)
	}
	affs2Untouched(t, f, "after the blocked second completer")
}

// (g) Where the release branch is not provisioned (checked out nowhere), the
// existing local-merge path is taken unchanged: the card lands in the worktree
// that holds main and the release branch does not move.
func TestAFFS2CompleteWithoutProvisionedReleaseBranchKeepsMainRoute(t *testing.T) {
	f := affs2New(t, affs2Opts{unprovisioned: true})
	if _, _, err := runFactory(t, "complete", "t1", "--run", fcRun); err != nil {
		t.Fatalf("complete: %v", err)
	}
	c := fcCard(t, f.root, "t1")
	if c.State != homestate.CardMergedLocal || c.MergeSHA == "" {
		t.Fatalf("card = %s merge=%q, want merged-local through the existing main route", c.State, c.MergeSHA)
	}
	if got := affs2Git(t, f.relWT, "rev-parse", "HEAD"); got != c.MergeSHA {
		t.Errorf("main HEAD = %s, want the merge %s", got, c.MergeSHA)
	}
	if got := affs2Git(t, f.root, "rev-parse", affs2ReleaseBranch); got != f.relBefore {
		t.Errorf("the unprovisioned release branch moved %s -> %s", f.relBefore, got)
	}
	if holds := affs2Holds(t, f.root); len(holds) != 0 {
		t.Errorf("the existing route wrote hold records: %v", holds)
	}
}
