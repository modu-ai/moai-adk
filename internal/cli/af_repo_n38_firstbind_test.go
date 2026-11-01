package cli

// N38, the first bind (leader's re-measurement of the audit of 46dbcce0d4). A ledger that is not bound yet used to take
// the repository of whichever call came first, so a caller whose project directory was another repository bound the
// ledger to it although the ledger already knew where its card's commits live. Before it binds, a verb now asks what the
// ledger already knows: the repository of every lease worktree it records, and the repository the ledger file itself sits
// in. A repository that disagrees with any of them is refused, with a message naming both, the exit non-zero and the
// ledger unchanged. What the ledger cannot know (no lease with a worktree that is a repository, a file outside any
// repository) it asks the caller to declare the repository with --bind-repo (N51); TestAFVerbs_N38FirstBindWithNothingKnownNeedsADeclaration
// pins that.
//
// The fixtures here leave the ledger UNBOUND (the card is claimed straight through the store, evidence and queue rows are
// written through the store), unlike af_repo_n38_test.go whose card is first measured through a verb run in the project.

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// afN38UnboundCard is a claimed card with c1 committed and nothing recorded: the ledger has no binding.
func afN38UnboundCard(t *testing.T) afN3Card {
	t.Helper()
	db, repo, c1, st, epoch := afUnqueuedCard(t)
	a := afN3Card{t: t, db: db, repo: repo, c1: c1, st: st, e2: strconv.FormatInt(epoch, 10)}
	if got := a.bound(); got != "" {
		t.Fatalf("the fixture must leave the ledger unbound, it is bound to %q", got)
	}
	return a
}

// storeMeasure records the passing green and gate runs of sha straight in the store, which binds nothing.
func (a afN3Card) storeMeasure(sha string) {
	a.t.Helper()
	epoch, _ := strconv.ParseInt(a.e2, 10, 64)
	for _, kind := range []string{"green", "gate"} {
		if err := a.st.AddEvidence(context.Background(), "t1", epoch, ledger.Evidence{Kind: kind, SubjectSHA: sha, Result: "pass"}); err != nil {
			a.t.Fatal(err)
		}
	}
}

// refusedAtFirstBind runs a verb that must be refused because the ledger already knows another repository than the one
// the call would bind: a non-zero exit whose message names both, the ledger unchanged and still unbound.
func (a afN3Card) refusedAtFirstBind(what, known, used string, args ...string) {
	a.t.Helper()
	before := afSnapshot(a.t, a.st)
	out, err := a.run(args...)
	if err == nil {
		a.t.Errorf("%s: accepted, output %.200s", what, out)
		return
	}
	for _, want := range []string{known, used} {
		if !strings.Contains(err.Error(), want) {
			a.t.Errorf("%s: the refusal does not name %q: %v", what, want, err)
		}
	}
	if after := afSnapshot(a.t, a.st); !reflect.DeepEqual(before, after) {
		a.t.Errorf("%s: the refusal changed the ledger:\nbefore %+v\nafter  %+v", what, before, after)
	}
	if got := a.bound(); got != "" {
		a.t.Errorf("%s: the refused call bound the ledger to %q", what, got)
	}
}

// The audit's reproduction (TestA3Repo_WorkingDirectoryNamesAnotherRepository): the card's lease worktree is the project,
// and the first call is `evidence ... --repo other` with the working directory in `other` and no CLAUDE_PROJECT_DIR. The
// ledger used to bind `other` and the sequence ended with the card DONE and the commit absent from the project.
func TestAFVerbs_N38FirstBindWorkingDirectoryNamesAnotherRepository(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN38UnboundCard(t)
	other, x := afN38Other(t)
	project, otherID := afN38GitDir(t, a.repo), afN38GitDir(t, other)
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	t.Chdir(other)
	sequence := []struct {
		name string
		args []string
	}{
		{"evidence green", []string{"evidence", "t1", "--epoch", a.e2, "--kind", "green", "--subject-sha", x, "--result", "pass", "--repo", other}},
		{"evidence gate", []string{"evidence", "t1", "--epoch", a.e2, "--kind", "gate", "--subject-sha", x, "--result", "pass", "--repo", other}},
		{"enqueue", []string{"enqueue", "t1", "--epoch", a.e2, "--candidate", x, "--repo", other}},
		{"land", []string{"land", "t1", "--gate", "true"}},
		{"finish: MERGING", []string{"move", "t1", "--epoch", a.e2, "--stage", "MERGING", "--remote-integrated"}},
		{"finish: DONE", []string{"move", "t1", "--epoch", a.e2, "--stage", "DONE"}},
	}
	for _, s := range sequence {
		a.refusedAtFirstBind(s.name, project, otherID, s.args...)
	}
	if got := a.stage(); got == "DONE" {
		t.Errorf("the card is %s", got)
	}
	// the project repository is never required to hold a commit it lacks: nothing was written, nothing landed
	if exec.Command("git", "-C", a.repo, "cat-file", "-e", x).Run() == nil {
		t.Error("the project repository holds the other repository's commit")
	}
	if rows, err := a.st.Queue(context.Background()); err != nil || len(rows) != 0 {
		t.Errorf("the queue must be empty: %+v %v", rows, err)
	}
	// the control: from the project the same card is measured, bound to the project and carried on
	t.Setenv("CLAUDE_PROJECT_DIR", a.repo)
	t.Chdir(a.repo)
	a.measureDeclared(a.c1) // N65: the first bind of this ledger declares its repository, the lease alone is the caller's word
	if got := a.bound(); got != project {
		t.Errorf("bound %q from the project, want %q", got, project)
	}
}

// The same with CLAUDE_PROJECT_DIR naming the other repository (audit TestA3Repo_EnvNamesAnotherRepository), every verb.
func TestAFVerbs_N38FirstBindEnvNamesAnotherRepositoryRefusesEveryVerb(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN38UnboundCard(t)
	other, _ := afN38Other(t)
	project, otherID := afN38GitDir(t, a.repo), afN38GitDir(t, other)
	t.Setenv("CLAUDE_PROJECT_DIR", other)
	cmds := afN38EveryVerb(a)
	for _, name := range afN38Verbs() {
		t.Run(name, func(t *testing.T) {
			a := a
			a.t = t
			a.refusedAtFirstBind(name+" with CLAUDE_PROJECT_DIR in another repository", project, otherID, cmds[name]...)
		})
	}
	a.refusedAtFirstBind("evidence --repo other", project, otherID, "evidence", "t1", "--epoch", a.e2, "--kind", "red", "--subject-sha", a.c1, "--result", "fail", "--repo", other)
}

// Every verb from the working directory of another repository (no environment).
func TestAFVerbs_N38FirstBindWorkingDirectoryRefusesEveryVerb(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN38UnboundCard(t)
	other, _ := afN38Other(t)
	project, otherID := afN38GitDir(t, a.repo), afN38GitDir(t, other)
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	t.Chdir(other)
	cmds := afN38EveryVerb(a)
	for _, name := range afN38Verbs() {
		t.Run(name, func(t *testing.T) {
			a := a
			a.t = t
			a.refusedAtFirstBind(name+" from another repository's working directory", project, otherID, cmds[name]...)
		})
	}
}

// audit TestA3Repo_LandFromASecondCloneOfTheProject on an unbound ledger: the card is measured and queued without a verb,
// and the first verb is the land, run in a second clone of the project.
func TestAFVerbs_N38FirstBindLandFromASecondCloneOfTheProjectIsRefused(t *testing.T) {
	afSkipWithoutSh(t)
	ctx := context.Background()
	a := afN38UnboundCard(t)
	a.storeMeasure(a.c1)
	epoch, _ := strconv.ParseInt(a.e2, 10, 64)
	if err := a.st.Enqueue(ctx, "t1", epoch, a.c1); err != nil {
		t.Fatal(err)
	}
	clone := t.TempDir()
	if out, err := exec.Command("git", "clone", "-q", "--no-hardlinks", a.repo, clone).CombinedOutput(); err != nil {
		t.Fatalf("clone: %v %s", err, out)
	}
	afGit(t, clone, "fetch", "-q", "origin", "refs/heads/af/trunk:refs/heads/af/trunk", "refs/heads/c1:refs/heads/c1")
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	t.Chdir(clone)
	a.refusedAtFirstBind("land from the clone", afN38GitDir(t, a.repo), afN38GitDir(t, clone), "land", "t1", "--gate", "true")
	if a.inTrunk(a.c1) {
		t.Error("the refused land put c1 in the project's trunk")
	}
	if exec.Command("git", "-C", clone, "merge-base", "--is-ancestor", a.c1, "refs/heads/af/trunk").Run() == nil {
		t.Error("the refused land put c1 in the clone's trunk")
	}
}

// audit TestA3Repo_ReconcileFromAnotherRepository on an unbound ledger.
func TestAFVerbs_N38FirstBindReconcileFromAnotherRepositoryIsRefused(t *testing.T) {
	afSkipWithoutSh(t)
	ctx := context.Background()
	a := afN38UnboundCard(t)
	a.storeMeasure(a.c1)
	epoch, _ := strconv.ParseInt(a.e2, 10, 64)
	if err := a.st.Enqueue(ctx, "t1", epoch, a.c1); err != nil {
		t.Fatal(err)
	}
	if err := a.st.QueueUpdate(ctx, "t1", ledger.CardEpoch(epoch, a.c1), ledger.QueueVerifying, "", ""); err != nil {
		t.Fatal(err)
	}
	lock, err := a.st.AcquireLock(ctx, ledger.LockMerge, ledger.LockHolder{SessionID: "dead", PID: os.Getpid(), PIDStart: "start-dead"}, 60e9)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.st.QueueUpdate(ctx, "t1", ledger.LockEpoch(lock, a.c1), ledger.QueueLanding, "", ""); err != nil {
		t.Fatal(err)
	}
	if err := a.st.ReleaseLock(ctx, ledger.LockMerge, lock); err != nil {
		t.Fatal(err)
	}
	clone := t.TempDir()
	if out, cerr := exec.Command("git", "clone", "-q", "--no-hardlinks", a.repo, clone).CombinedOutput(); cerr != nil {
		t.Fatalf("clone: %v %s", cerr, out)
	}
	afGit(t, clone, "fetch", "-q", "origin", "refs/heads/c1:refs/heads/af/trunk") // in the clone c1 IS the trunk
	t.Setenv("CLAUDE_PROJECT_DIR", clone)
	a.refusedAtFirstBind("reconcile from the clone", afN38GitDir(t, a.repo), afN38GitDir(t, clone), "reconcile")
	if rows, qerr := a.st.Queue(ctx); qerr != nil || len(rows) != 1 || rows[0].State != ledger.QueueLanding {
		t.Errorf("the row must still be landing: %+v %v", rows, qerr)
	}
}

// audit TestA3Repo_LeaseWorktreeNamesAnotherRepository: the lease's worktree is another repository (claim takes any
// path), the project is the project. The ledger knows `other` from the lease, so a call that would bind the project is
// the one that disagrees; every verb is refused and none binds.
func TestAFVerbs_N38FirstBindLeaseWorktreeNamesAnotherRepositoryRefusesEveryVerb(t *testing.T) {
	afSkipWithoutSh(t)
	ctx := context.Background()
	a := afN38UnboundCard(t)
	other, x := afN38Other(t)
	epoch, _ := strconv.ParseInt(a.e2, 10, 64)
	if err := a.st.Release(ctx, "t1", epoch); err != nil {
		t.Fatal(err)
	}
	c, err := a.st.Claim(ctx, ledger.ClaimRequest{SessionID: "s2", Worktree: other, PID: os.Getpid(), PIDStart: "start-s2"})
	if err != nil {
		t.Fatal(err)
	}
	a.e2 = strconv.FormatInt(c.Epoch, 10)
	project, otherID := afN38GitDir(t, a.repo), afN38GitDir(t, other)
	cmds := afN38EveryVerb(a)
	for _, name := range afN38Verbs() {
		t.Run(name, func(t *testing.T) {
			a := a
			a.t = t
			a.refusedAtFirstBind(name+" with a lease in another repository", otherID, project, cmds[name]...)
		})
	}
	a.refusedAtFirstBind("evidence about the other repository's commit, no --repo", otherID, project,
		"evidence", "t1", "--epoch", a.e2, "--kind", "gate", "--subject-sha", x, "--result", "pass")
	a.refusedAtFirstBind("enqueue about the other repository's commit, no --repo", otherID, project,
		"enqueue", "t1", "--epoch", a.e2, "--candidate", x)
}

// The ledger file sits inside repository A while the caller's project is repository B: the first bind is refused, and the
// ledger stays unbound. Pointed at A (or a linked worktree of A) the same verb binds A.
func TestAFVerbs_N38FirstBindLedgerFileInsideAnotherRepositoryIsRefused(t *testing.T) {
	ctx := context.Background()
	newRepo := func() string {
		dir := t.TempDir()
		afGit(t, dir, "init", "-q", "-b", "work")
		if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("f\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		afGit(t, dir, "add", "-A")
		afGit(t, dir, "commit", "-q", "-m", "base")
		return dir
	}
	repoA, repoB := newRepo(), newRepo()
	idA, idB := afN38GitDir(t, repoA), afN38GitDir(t, repoB)
	db := filepath.Join(repoA, "af", "ledger.db")
	boundOf := func(db string) string {
		st, err := ledger.Open(ctx, db)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = st.Close() }()
		got, err := st.Repository(ctx)
		if err != nil {
			t.Fatal(err)
		}
		return got
	}
	t.Setenv("CLAUDE_PROJECT_DIR", repoB)
	_, errOut, err := afRunBoth("reap", "--db", db)
	if err == nil {
		t.Fatalf("reap bound the ledger of repository A to the project repository B (stderr %.200s)", errOut)
	}
	for _, want := range []string{idA, idB} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %q: %v", want, err)
		}
	}
	if got := boundOf(db); got != "" {
		t.Fatalf("the refused call bound the ledger to %q", got)
	}
	// the control: the project is the repository the file sits in
	t.Setenv("CLAUDE_PROJECT_DIR", repoA)
	if _, _, err := afRunBoth("reap", "--db", db); err != nil {
		t.Fatalf("reap from repository A: %v", err)
	}
	if got := boundOf(db); got != idA {
		t.Errorf("bound %q, want %q", got, idA)
	}
	// and a linked worktree of A is A
	wt := filepath.Join(t.TempDir(), "wt")
	afGit(t, repoA, "worktree", "add", "-q", "--detach", wt)
	db2 := filepath.Join(repoA, "af2", "ledger.db")
	t.Setenv("CLAUDE_PROJECT_DIR", wt)
	if _, _, err := afRunBoth("reap", "--db", db2); err != nil {
		t.Fatalf("reap from a linked worktree of repository A: %v", err)
	}
	if got := boundOf(db2); got != idA {
		t.Errorf("bound %q from the linked worktree, want %q", got, idA)
	}
}

// What the ledger cannot know: no lease worktree that is a repository, no commit in its rows and a ledger file outside any
// repository. It was a residual risk (the first call decided); since the clone form of N51 that first call has to declare
// the repository with --bind-repo, and without the declaration it is refused with the ledger unbound. A lease whose
// worktree is gone, or is no repository, is not a clue and not an error.
func TestAFVerbs_N38FirstBindWithNothingKnownNeedsADeclaration(t *testing.T) {
	ctx := context.Background()
	cases := map[string]func(t *testing.T, a afN3Card){
		"no lease at all": func(t *testing.T, a afN3Card) {
			epoch, _ := strconv.ParseInt(a.e2, 10, 64)
			if err := a.st.Release(ctx, "t1", epoch); err != nil {
				t.Fatal(err)
			}
		},
		"lease worktree gone": func(t *testing.T, a afN3Card) {
			afN38Reclaim(t, a, filepath.Join(t.TempDir(), "gone"))
		},
		"lease worktree is no repository": func(t *testing.T, a afN3Card) {
			afN38Reclaim(t, a, t.TempDir())
		},
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			a := afN38UnboundCard(t)
			setup(t, a)
			other, _ := afN38Other(t)
			t.Setenv("CLAUDE_PROJECT_DIR", other)
			a.refusedAtFirstBind("reap with nothing known and no declaration", "--bind-repo", afN38GitDir(t, other), "reap")
			if _, err := a.run("reap", "--bind-repo", other); err != nil {
				t.Fatalf("reap with nothing known and the repository declared: %v", err)
			}
			if got, want := a.bound(), afN38GitDir(t, other); got != want {
				t.Errorf("bound %q, want %q (the declared repository)", got, want)
			}
		})
	}
}

// afN38Reclaim lets go of the card and claims it again with the given worktree.
func afN38Reclaim(t *testing.T, a afN3Card, worktree string) {
	t.Helper()
	ctx := context.Background()
	epoch, _ := strconv.ParseInt(a.e2, 10, 64)
	if err := a.st.Release(ctx, "t1", epoch); err != nil {
		t.Fatal(err)
	}
	if _, err := a.st.Claim(ctx, ledger.ClaimRequest{SessionID: "s2", Worktree: worktree, PID: os.Getpid(), PIDStart: "start-s2"}); err != nil {
		t.Fatal(err)
	}
}

// A linked worktree of the project is the project: a lease in one, a project directory in one, in any pairing.
func TestAFVerbs_N38FirstBindLinkedWorktreeOfTheProjectPasses(t *testing.T) {
	for _, spelling := range []string{"project", "linked_worktree", "symlink"} {
		t.Run(spelling, func(t *testing.T) {
			a := afN38UnboundCard(t) // a fresh unbound ledger for each first bind
			wt := filepath.Join(t.TempDir(), "wt")
			afGit(t, a.repo, "worktree", "add", "-q", "--detach", wt, a.c1)
			link := filepath.Join(t.TempDir(), "link")
			if err := os.Symlink(a.repo, link); err != nil {
				t.Fatal(err)
			}
			afN38Reclaim(t, a, wt) // the lease is in the linked worktree...
			project := map[string]string{"project": a.repo, "linked_worktree": wt, "symlink": link}[spelling]
			t.Setenv("CLAUDE_PROJECT_DIR", project) // ...and the first call comes from each spelling of the project
			// N65: a lease alone no longer says where the ledger belongs, so the first bind declares the project (its
			// agreement with the lease in the linked worktree is still compared, and a disagreement is still refused)
			if _, err := a.run("reap", "--bind-repo", project); err != nil {
				t.Fatalf("reap with the project directory %s: %v", project, err)
			}
			if got, want := a.bound(), afN38GitDir(t, a.repo); got != want {
				t.Errorf("bound %q, want the project's git directory %q", got, want)
			}
		})
	}
}
