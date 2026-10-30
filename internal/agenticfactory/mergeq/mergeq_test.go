package mergeq

// Defect answer: these tests catch a landing that overwrites a trunk that moved while the gates ran, a ref
// update that is not a compare-and-swap, gates run while the landing lock is held, a failed or conflicting
// candidate that still moves the trunk, a lock that leaks on any path, a restart that re-lands or loses a
// card, and a candidate that touches main or a remote. Every test uses a real temporary git repository; the
// ledger is a fake that records the calls it received.

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// --- fixtures -------------------------------------------------------------------------------------------

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_SYSTEM="+os.DevNull,
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.test", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.test")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

// repo makes a repository whose trunk ref (refs/heads/trunk) is NOT checked out anywhere (HEAD is on "work"),
// like a landing target that nobody edits.
func repo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "work")
	write(t, dir, "base.txt", "base\n")
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-q", "-m", "base")
	git(t, dir, "branch", "trunk")
	return dir
}

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// candidate commits file=body on a branch off the trunk and returns that commit.
func candidate(t *testing.T, dir, branch, file, body string) string {
	t.Helper()
	git(t, dir, "checkout", "-q", "-b", branch, "trunk")
	write(t, dir, file, body)
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-q", "-m", "card "+branch)
	sha := git(t, dir, "rev-parse", "HEAD")
	git(t, dir, "checkout", "-q", "work")
	return sha
}

// advanceTrunk lands an unrelated commit directly on the trunk ref and returns the new trunk commit.
func advanceTrunk(t *testing.T, dir, file, body string) string {
	t.Helper()
	sha := candidate(t, dir, "other-"+file, file, body)
	git(t, dir, "update-ref", "refs/heads/trunk", sha, git(t, dir, "rev-parse", "refs/heads/trunk"))
	return sha
}

type fakeQueue struct {
	mu         sync.Mutex
	states     []State
	failReason string
	locks      int
	unlocks    int
	held       bool
	notes      []string
	heldBy     error // when set, AcquireLock returns it
	inLock     func() bool
	verifiedIn []bool // for each state set, was the lock held
}

func (q *fakeQueue) AcquireLock(_ context.Context, name string, _ Holder, _ time.Duration) (int64, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if name != LandingLock {
		return 0, errors.New("unexpected lock " + name)
	}
	if q.heldBy != nil {
		return 0, q.heldBy
	}
	q.locks++
	q.held = true
	return int64(q.locks), nil
}

func (q *fakeQueue) ReleaseLock(_ context.Context, _ string, _ int64) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.unlocks++
	q.held = false
	return nil
}

func (q *fakeQueue) SetState(_ context.Context, _ string, _ int64, st State, _, reason string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.states = append(q.states, st)
	q.verifiedIn = append(q.verifiedIn, q.held)
	if reason != "" {
		q.failReason = reason
	}
	return nil
}

func (q *fakeQueue) AddNote(_ context.Context, kind, card, text string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.notes = append(q.notes, kind+"|"+card+"|"+text)
	return nil
}

func (q *fakeQueue) lockHeld() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.held
}

func newMerger(t *testing.T, dir string, q *fakeQueue, verify Verify) *Merger {
	t.Helper()
	g := GitCLI{Repo: dir, ScratchD: t.TempDir(), Who: Identity{Name: "mergeq", Email: "mergeq@example.test"}}
	return &Merger{
		Git: g, Queue: q, Trunk: "refs/heads/trunk", Verify: verify,
		Holder:  Holder{Session: "s1", PID: os.Getpid()},
		LockTTL: DefaultLockTTL, MaxRetries: DefaultMaxRetries,
		Message: func(card string) string { return "merge(" + card + "): land" },
	}
}

func passing(_ context.Context, _ string) error { return nil }

// --- tests ----------------------------------------------------------------------------------------------

func TestLand_CleanCandidateMovesTheTrunkToTheMergeResult(t *testing.T) {
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	before := git(t, dir, "rev-parse", "refs/heads/trunk")
	q := &fakeQueue{}
	m := newMerger(t, dir, q, passing)

	res, err := m.Land(context.Background(), "t1", cand)
	if err != nil {
		t.Fatal(err)
	}
	after := git(t, dir, "rev-parse", "refs/heads/trunk")
	if res.Outcome != OutcomeLanded || res.TrunkAfter != after || res.TrunkBefore != before {
		t.Fatalf("result %+v, trunk %s -> %s", res, before, after)
	}
	if after == before || git(t, dir, "show", "refs/heads/trunk:a.txt") != "a" {
		t.Errorf("the trunk did not receive the candidate's change")
	}
	if got := git(t, dir, "log", "-1", "--format=%s", "refs/heads/trunk"); got != "merge(t1): land" {
		t.Errorf("merge commit subject = %q, want it to carry the card id", got)
	}
	want := []State{StateVerifying, StateLanding, StateLanded}
	if strings.Join(statesToStrings(q.states), ",") != strings.Join(statesToStrings(want), ",") {
		t.Errorf("states %v, want %v", q.states, want)
	}
	if q.locks != 1 || q.unlocks != 1 || q.held {
		t.Errorf("lock acquired %d released %d held %v, want 1, 1, false", q.locks, q.unlocks, q.held)
	}
	landed := 0
	for _, n := range q.notes {
		if strings.HasPrefix(n, "landed|t1|") {
			landed++
		}
	}
	if landed != 1 {
		t.Errorf("landed notes = %d (%v), want exactly 1", landed, q.notes)
	}
	// the candidate worktree is gone
	if out := git(t, dir, "worktree", "list", "--porcelain"); strings.Count(out, "worktree ") != 1 {
		t.Errorf("a candidate worktree was left behind:\n%s", out)
	}
}

func statesToStrings(s []State) []string {
	out := make([]string, len(s))
	for i, v := range s {
		out[i] = string(v)
	}
	return out
}

func TestLand_GatesRunOutsideTheLock(t *testing.T) {
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	q := &fakeQueue{}
	sawLock := false
	m := newMerger(t, dir, q, func(context.Context, string) error {
		sawLock = q.lockHeld()
		return nil
	})
	if _, err := m.Land(context.Background(), "t1", cand); err != nil {
		t.Fatal(err)
	}
	if sawLock {
		t.Error("the gates ran while the landing lock was held")
	}
	if q.locks != 1 {
		t.Fatalf("control: the lock was taken %d times, want 1 (so the check above could have fired)", q.locks)
	}
}

func TestLand_FailedGatesLeaveTheTrunkAndTakeNoLock(t *testing.T) {
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	before := git(t, dir, "rev-parse", "refs/heads/trunk")
	q := &fakeQueue{}
	m := newMerger(t, dir, q, func(context.Context, string) error { return errors.New("go test failed") })
	res, err := m.Land(context.Background(), "t1", cand)
	if err != nil {
		t.Fatal(err)
	}
	if res.Outcome != OutcomeFailed || !strings.Contains(res.Reason, "go test failed") {
		t.Errorf("result %+v", res)
	}
	if git(t, dir, "rev-parse", "refs/heads/trunk") != before {
		t.Error("the trunk moved for a candidate whose gates failed")
	}
	if q.locks != 0 {
		t.Errorf("the lock was taken %d times for a failed candidate", q.locks)
	}
	if last := q.states[len(q.states)-1]; last != StateFailed || !strings.Contains(q.failReason, "go test failed") {
		t.Errorf("final state %v reason %q", last, q.failReason)
	}
}

func TestLand_AConflictingCandidateFailsAndTheTrunkStays(t *testing.T) {
	dir := repo(t)
	// both sides change base.txt differently
	git(t, dir, "checkout", "-q", "-b", "c1", "trunk")
	write(t, dir, "base.txt", "candidate\n")
	git(t, dir, "commit", "-q", "-am", "c1")
	cand := git(t, dir, "rev-parse", "HEAD")
	git(t, dir, "checkout", "-q", "work")
	git(t, dir, "checkout", "-q", "-b", "o1", "trunk")
	write(t, dir, "base.txt", "trunk side\n")
	git(t, dir, "commit", "-q", "-am", "o1")
	other := git(t, dir, "rev-parse", "HEAD")
	git(t, dir, "checkout", "-q", "work")
	git(t, dir, "update-ref", "refs/heads/trunk", other, git(t, dir, "rev-parse", "refs/heads/trunk"))

	q := &fakeQueue{}
	calls := 0
	m := newMerger(t, dir, q, func(context.Context, string) error { calls++; return nil })
	res, err := m.Land(context.Background(), "t1", cand)
	if err != nil {
		t.Fatal(err)
	}
	if res.Outcome != OutcomeFailed || !strings.Contains(strings.ToLower(res.Reason), "merge") {
		t.Errorf("result %+v", res)
	}
	if calls != 0 {
		t.Errorf("the gates ran %d times on a candidate that does not merge", calls)
	}
	if git(t, dir, "rev-parse", "refs/heads/trunk") != other {
		t.Error("the trunk moved for a conflicting candidate")
	}
}

func TestLand_ATrunkThatMovesDuringTheGatesIsRebuiltNotOverwritten(t *testing.T) {
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	q := &fakeQueue{}
	runs := 0
	var moved string
	m := newMerger(t, dir, q, func(context.Context, string) error {
		runs++
		if runs == 1 {
			moved = advanceTrunk(t, dir, "b.txt", "b\n") // someone else lands while the gates run
		}
		return nil
	})
	res, err := m.Land(context.Background(), "t1", cand)
	if err != nil {
		t.Fatal(err)
	}
	if res.Outcome != OutcomeLanded || res.Verified != 2 || runs != 2 {
		t.Fatalf("result %+v, gate runs %d, want landed after 2 verifications", res, runs)
	}
	if git(t, dir, "show", "refs/heads/trunk:a.txt") != "a" || git(t, dir, "show", "refs/heads/trunk:b.txt") != "b" {
		t.Error("the landed trunk lost one side: a candidate overwrote the trunk that moved")
	}
	if ok, err := (GitCLI{Repo: dir}).IsAncestor(context.Background(), moved, "refs/heads/trunk"); err != nil || !ok {
		t.Errorf("the commit that landed meanwhile is not an ancestor of the new trunk (%v, %v)", ok, err)
	}
	if res.TrunkBefore != moved {
		t.Errorf("TrunkBefore = %s, want the moved trunk %s", res.TrunkBefore, moved)
	}
}

// movingGit moves the trunk once, right before the first UpdateRef, to prove the update is a CAS.
type movingGit struct {
	Git
	t     *testing.T
	dir   string
	moved bool
}

func (g *movingGit) UpdateRef(ctx context.Context, ref, newSHA, oldSHA string) error {
	if !g.moved {
		g.moved = true
		advanceTrunk(g.t, g.dir, "late.txt", "late\n")
	}
	return g.Git.UpdateRef(ctx, ref, newSHA, oldSHA)
}

func TestLand_ATrunkThatMovesAfterTheRecheckIsCaughtByTheCompareAndSwap(t *testing.T) {
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	q := &fakeQueue{}
	m := newMerger(t, dir, q, passing)
	m.Git = &movingGit{Git: m.Git, t: t, dir: dir}
	res, err := m.Land(context.Background(), "t1", cand)
	if err != nil {
		t.Fatal(err)
	}
	if res.Outcome != OutcomeLanded {
		t.Fatalf("result %+v", res)
	}
	if git(t, dir, "show", "refs/heads/trunk:late.txt") != "late" || git(t, dir, "show", "refs/heads/trunk:a.txt") != "a" {
		t.Error("the compare-and-swap did not protect the commit that landed in the window")
	}
	if q.locks != q.unlocks || q.held {
		t.Errorf("the lock leaked: acquired %d released %d held %v", q.locks, q.unlocks, q.held)
	}
}

func TestLand_ATrunkThatNeverStopsMovingEndsAsFailedAfterTheRetryCap(t *testing.T) {
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	q := &fakeQueue{}
	n := 0
	m := newMerger(t, dir, q, func(context.Context, string) error {
		n++
		advanceTrunk(t, dir, "x"+string(rune('a'+n))+".txt", "x\n")
		return nil
	})
	res, err := m.Land(context.Background(), "t1", cand)
	if err != nil {
		t.Fatal(err)
	}
	if res.Outcome != OutcomeFailed || !strings.Contains(res.Reason, "moved") {
		t.Errorf("result %+v", res)
	}
	if n != DefaultMaxRetries {
		t.Errorf("the gates ran %d times, want the cap %d", n, DefaultMaxRetries)
	}
	if q.locks != q.unlocks || q.held {
		t.Errorf("the lock leaked: acquired %d released %d held %v", q.locks, q.unlocks, q.held)
	}
}

func TestLand_ALockHeldByAnotherLanderIsWaitingNotFailure(t *testing.T) {
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	before := git(t, dir, "rev-parse", "refs/heads/trunk")
	q := &fakeQueue{heldBy: ErrLockHeld}
	m := newMerger(t, dir, q, passing)
	res, err := m.Land(context.Background(), "t1", cand)
	if err != nil {
		t.Fatal(err)
	}
	if res.Outcome != OutcomeWaiting {
		t.Errorf("result %+v, want waiting", res)
	}
	if git(t, dir, "rev-parse", "refs/heads/trunk") != before {
		t.Error("the trunk moved without the lock")
	}
	for _, s := range q.states {
		if s == StateFailed {
			t.Error("a held lock was recorded as a failure")
		}
	}
}

func TestLand_ACandidateAlreadyInTheTrunkLandsWithoutAnotherMerge(t *testing.T) {
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	git(t, dir, "update-ref", "refs/heads/trunk", cand, git(t, dir, "rev-parse", "refs/heads/trunk"))
	before := git(t, dir, "rev-parse", "refs/heads/trunk")
	q := &fakeQueue{}
	runs := 0
	m := newMerger(t, dir, q, func(context.Context, string) error { runs++; return nil })
	res, err := m.Land(context.Background(), "t1", cand)
	if err != nil {
		t.Fatal(err)
	}
	if res.Outcome != OutcomeLanded || git(t, dir, "rev-parse", "refs/heads/trunk") != before || runs != 0 {
		t.Errorf("result %+v, trunk changed %v, gate runs %d: an already-landed candidate was merged again",
			res, git(t, dir, "rev-parse", "refs/heads/trunk") != before, runs)
	}
}

func TestLand_NeverTouchesMainOrAnyOtherRef(t *testing.T) {
	dir := repo(t)
	git(t, dir, "branch", "main", "trunk")
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	refsBefore := git(t, dir, "for-each-ref", "--format=%(refname) %(objectname)", "refs/heads/main", "refs/heads/work")
	m := newMerger(t, dir, &fakeQueue{}, passing)
	if _, err := m.Land(context.Background(), "t1", cand); err != nil {
		t.Fatal(err)
	}
	if got := git(t, dir, "for-each-ref", "--format=%(refname) %(objectname)", "refs/heads/main", "refs/heads/work"); got != refsBefore {
		t.Errorf("a ref other than the trunk changed:\nbefore %s\nafter  %s", refsBefore, got)
	}
	if out := git(t, dir, "remote"); out != "" {
		t.Errorf("a remote appeared: %q", out)
	}
}

func TestReconcile_UsesGitAncestryNotTheRecordedState(t *testing.T) {
	dir := repo(t)
	landed := candidate(t, dir, "c1", "a.txt", "a\n")
	notLanded := candidate(t, dir, "c2", "b.txt", "b\n")
	git(t, dir, "update-ref", "refs/heads/trunk", landed, git(t, dir, "rev-parse", "refs/heads/trunk"))
	m := newMerger(t, dir, &fakeQueue{}, passing)

	acts, err := m.Reconcile(context.Background(), []Item{
		{CardID: "t1", CandidateSHA: landed, State: StateLanding},    // crashed after update-ref
		{CardID: "t2", CandidateSHA: notLanded, State: StateLanding}, // crashed before update-ref
		{CardID: "t3", CandidateSHA: notLanded, State: StateVerifying},
		{CardID: "t4", CandidateSHA: notLanded, State: StateWaiting}, // untouched
		{CardID: "t5", CandidateSHA: landed, State: StateLanded},     // already settled
	})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]State{}
	for _, a := range acts {
		got[a.CardID] = a.To
	}
	want := map[string]State{"t1": StateLanded, "t2": StateWaiting, "t3": StateWaiting}
	for id, st := range want {
		if got[id] != st {
			t.Errorf("%s -> %q, want %q (all actions: %+v)", id, got[id], st, acts)
		}
	}
	if _, touched := got["t4"]; touched {
		t.Error("a waiting row was changed")
	}
	if _, touched := got["t5"]; touched {
		t.Error("a landed row was changed")
	}
}

func TestReconcile_RecordsOneLandedNoteAndIsIdempotent(t *testing.T) {
	dir := repo(t)
	landed := candidate(t, dir, "c1", "a.txt", "a\n")
	git(t, dir, "update-ref", "refs/heads/trunk", landed, git(t, dir, "rev-parse", "refs/heads/trunk"))
	q := &fakeQueue{}
	m := newMerger(t, dir, q, passing)
	items := []Item{{CardID: "t1", CandidateSHA: landed, State: StateLanding}}
	if _, err := m.Reconcile(context.Background(), items); err != nil {
		t.Fatal(err)
	}
	// a second pass over the same (now settled) row changes nothing more
	items[0].State = StateLanded
	if acts, err := m.Reconcile(context.Background(), items); err != nil || len(acts) != 0 {
		t.Errorf("second pass: actions %+v err %v, want none", acts, err)
	}
	n := 0
	for _, note := range q.notes {
		if strings.HasPrefix(note, "landed|t1|") {
			n++
		}
	}
	if n != 1 {
		t.Errorf("landed notes = %d, want 1", n)
	}
}
