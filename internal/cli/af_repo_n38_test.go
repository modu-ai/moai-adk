package cli

// N38 (audit of 46dbcce0d4) through the real `moai af` verbs, a real ledger and real git. A ledger belongs to one
// repository: the first verb that writes it binds the git directory of the repository the verb uses, and a later call
// that would use another repository (the working directory, CLAUDE_PROJECT_DIR, the card's lease worktree) is refused
// with a message naming both, the exit non-zero and the ledger unchanged. A linked worktree of the same repository, or
// another spelling of its path, is the same repository and passes. The audit's reproductions were one-time
// observations; these are the permanent versions, with real assertions.

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// afN38Card is a claimed card with c1 committed. The ledger is bound by a real verb run in the project (measure c1),
// as it is in use: the first writer decides.
func afN38Card(t *testing.T) afN3Card {
	t.Helper()
	db, repo, c1, st, epoch := afUnqueuedCard(t)
	afBindFixture(t, st, repo)
	a := afN3Card{t: t, db: db, repo: repo, c1: c1, st: st, e2: strconv.FormatInt(epoch, 10)}
	a.measure(c1)
	return a
}

// afN38Other is another repository with the dedicated trunk and a commit x on a branch off it.
func afN38Other(t *testing.T) (repo, x string) {
	t.Helper()
	repo = t.TempDir()
	afGit(t, repo, "init", "-q", "-b", "work")
	if err := os.WriteFile(filepath.Join(repo, "other.txt"), []byte("other\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	afGit(t, repo, "add", "-A")
	afGit(t, repo, "commit", "-q", "-m", "other base")
	afGit(t, repo, "branch", "af/trunk")
	afGit(t, repo, "checkout", "-q", "-b", "x", "af/trunk")
	if err := os.WriteFile(filepath.Join(repo, "x.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	afGit(t, repo, "add", "-A")
	afGit(t, repo, "commit", "-q", "-m", "x")
	x = afGit(t, repo, "rev-parse", "HEAD")
	afGit(t, repo, "checkout", "-q", "work")
	return repo, x
}

// afN38GitDir is the repository identity the verbs are expected to compute for a plain repository at dir.
func afN38GitDir(t *testing.T, dir string) string {
	t.Helper()
	p, err := filepath.EvalSymlinks(filepath.Join(dir, ".git"))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// run runs one verb with --db appended (the verbs of this file do not all take a card first, unlike try).
func (a afN3Card) run(args ...string) (string, error) {
	a.t.Helper()
	out, errOut, err := afRunBoth(append(args[:len(args):len(args)], "--db", a.db)...)
	return out + errOut, err
}

func (a afN3Card) mustRun(args ...string) string {
	a.t.Helper()
	out, err := a.run(args...)
	if err != nil {
		a.t.Fatalf("moai af %s: %v\n%.300s", strings.Join(args, " "), err, out)
	}
	return out
}

func (a afN3Card) bound() string {
	a.t.Helper()
	got, err := a.st.Repository(context.Background())
	if err != nil {
		a.t.Fatal(err)
	}
	return got
}

// refused runs a verb that must be refused because of the repository: a non-zero exit whose message names the bound
// repository and the one the call would have used, with nothing in the ledger changed.
func (a afN3Card) refused(what, wantUsed string, args ...string) {
	a.t.Helper()
	before, boundBefore := afSnapshot(a.t, a.st), a.bound()
	out, err := a.run(args...)
	if err == nil {
		a.t.Errorf("%s: accepted, output %.200s", what, out)
		return
	}
	for _, want := range []string{"bound to", boundBefore, wantUsed} {
		if !strings.Contains(err.Error(), want) {
			a.t.Errorf("%s: the refusal does not say %q: %v", what, want, err)
		}
	}
	if after := afSnapshot(a.t, a.st); !reflect.DeepEqual(before, after) {
		a.t.Errorf("%s: the refusal changed the ledger:\nbefore %+v\nafter  %+v", what, before, after)
	}
	if got := a.bound(); got != boundBefore {
		a.t.Errorf("%s: the binding moved %q -> %q", what, boundBefore, got)
	}
}

// afN38EveryVerb is one invocation of every `moai af` verb that opens the ledger, none of which names a repository.
func afN38EveryVerb(a afN3Card) map[string][]string {
	return map[string][]string{
		"claim":     {"claim", "--session", "s-other", "--worktree", a.repo},
		"move":      {"move", "t1", "--epoch", a.e2, "--stage", "REVIEWING"},
		"evidence":  {"evidence", "t1", "--epoch", a.e2, "--kind", "red", "--subject-sha", a.c1, "--result", "fail"},
		"enqueue":   {"enqueue", "t1", "--epoch", a.e2, "--candidate", a.c1},
		"release":   {"release", "t1", "--epoch", a.e2},
		"reap":      {"reap"},
		"land":      {"land", "t1", "--gate", "true"},
		"reconcile": {"reconcile"},
		"resume":    {"resume", "t1", "--epoch", a.e2, "--guard", "reclaim-cap", "--reason", "looked"},
	}
}

func afN38Verbs() []string {
	return []string{"claim", "move", "evidence", "enqueue", "release", "reap", "land", "reconcile", "resume"}
}

func TestAFVerbs_N38TheFirstVerbBindsTheLedgerToTheProjectsGitDirectory(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN38Card(t)
	if got, want := a.bound(), afN38GitDir(t, a.repo); got != want {
		t.Errorf("bound %q, want the project's git directory %q", got, want)
	}
}

// The most ordinary form: everything is measured and queued in the project; only the land verb runs in another clone
// of it, with the ledger named by --db and no flag wrong. c1 exists there.
func TestAFVerbs_N38LandFromASecondCloneOfTheProjectIsRefused(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN38Card(t)
	a.verb("enqueue", "t1", "--epoch", a.e2, "--candidate", a.c1, "--repo", a.repo)
	clone := t.TempDir()
	if out, err := exec.Command("git", "clone", "-q", "--no-hardlinks", a.repo, clone).CombinedOutput(); err != nil {
		t.Fatalf("clone: %v %s", err, out)
	}
	afGit(t, clone, "fetch", "-q", "origin", "refs/heads/af/trunk:refs/heads/af/trunk", "refs/heads/c1:refs/heads/c1")
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	t.Chdir(clone)
	a.refused("land from the clone", afN38GitDir(t, clone), "land", "t1", "--gate", "true")
	if a.inTrunk(a.c1) {
		t.Error("the refused land put c1 in the project's trunk")
	}
	if exec.Command("git", "-C", clone, "merge-base", "--is-ancestor", a.c1, "refs/heads/af/trunk").Run() == nil {
		t.Error("the refused land put c1 in the clone's trunk")
	}
	// the control: from the project the same card lands and finishes
	t.Setenv("CLAUDE_PROJECT_DIR", a.repo)
	t.Chdir(a.repo)
	if out := a.verb("land", "t1", "--gate", "true"); !strings.Contains(out, `"outcome":"landed"`) {
		t.Fatalf("land from the project: %s", out)
	}
	if err := a.tryFinish(); err != nil {
		t.Fatalf("finish: %v", err)
	}
	if !a.inTrunk(a.c1) {
		t.Error("the card is DONE and c1 is not in the project's trunk")
	}
}

// CLAUDE_PROJECT_DIR names another repository: every verb is refused, the --repo check (which compares with that very
// directory) is no help.
func TestAFVerbs_N38EveryVerbIsRefusedWhenTheProjectDirectoryNamesAnotherRepository(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN38Card(t)
	other, _ := afN38Other(t)
	t.Setenv("CLAUDE_PROJECT_DIR", other)
	cmds := afN38EveryVerb(a)
	for _, name := range afN38Verbs() {
		t.Run(name, func(t *testing.T) {
			a := a
			a.t = t
			a.refused(name+" with CLAUDE_PROJECT_DIR in another repository", afN38GitDir(t, other), cmds[name]...)
		})
	}
	// evidence and enqueue with a --repo that agrees with that directory are refused all the same
	a.refused("evidence --repo other", afN38GitDir(t, other), "evidence", "t1", "--epoch", a.e2, "--kind", "red", "--subject-sha", a.c1, "--result", "fail", "--repo", other)
	a.refused("enqueue --repo other", afN38GitDir(t, other), "enqueue", "t1", "--epoch", a.e2, "--candidate", a.c1, "--repo", other)
}

// No environment at all: the working directory is the project directory.
func TestAFVerbs_N38EveryVerbIsRefusedWhenTheWorkingDirectoryIsAnotherRepository(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN38Card(t)
	other, _ := afN38Other(t)
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	t.Chdir(other)
	cmds := afN38EveryVerb(a)
	for _, name := range afN38Verbs() {
		t.Run(name, func(t *testing.T) {
			a := a
			a.t = t
			a.refused(name+" from another repository's working directory", afN38GitDir(t, other), cmds[name]...)
		})
	}
}

// A reconcile from another repository would settle a landing row from that repository's ancestry.
func TestAFVerbs_N38ReconcileFromAnotherRepositoryIsRefused(t *testing.T) {
	afSkipWithoutSh(t)
	ctx := context.Background()
	a := afN38Card(t)
	a.verb("enqueue", "t1", "--epoch", a.e2, "--candidate", a.c1, "--repo", a.repo)
	epoch, _ := strconv.ParseInt(a.e2, 10, 64)
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
	a.refused("reconcile from the clone", afN38GitDir(t, clone), "reconcile")
	rows, qerr := a.st.Queue(ctx)
	if qerr != nil || len(rows) != 1 || rows[0].State != ledger.QueueLanding {
		t.Errorf("the row must still be landing: %+v %v", rows, qerr)
	}
	// the control: from the project the reconcile runs (c1 is not in the project's trunk: the row goes back to waiting)
	t.Setenv("CLAUDE_PROJECT_DIR", a.repo)
	a.mustRun("reconcile")
	if rows, _ = a.st.Queue(ctx); len(rows) != 1 || rows[0].State != ledger.QueueWaiting {
		t.Errorf("reconcile from the project: %+v", rows)
	}
}

// claim takes any path for the worktree. What the lease names is where evidence and enqueue look a commit up when --repo
// is left out; a lease in another repository is refused there.
func TestAFVerbs_N38ALeaseWorktreeInAnotherRepositoryIsRefusedWhereItIsUsed(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN38Card(t)
	other, x := afN38Other(t)
	a.verb("release", "t1", "--epoch", a.e2)
	var claimed struct {
		Epoch int64 `json:"epoch"`
	}
	if err := json.Unmarshal([]byte(a.mustRun("claim", "--session", "s2", "--worktree", other)), &claimed); err != nil || claimed.Epoch == 0 {
		t.Fatalf("claim: %+v %v", claimed, err)
	}
	a.e2 = strconv.FormatInt(claimed.Epoch, 10)
	a.refused("evidence with the lease in another repository", afN38GitDir(t, other),
		"evidence", "t1", "--epoch", a.e2, "--kind", "gate", "--subject-sha", x, "--result", "pass")
	a.refused("enqueue with the lease in another repository", afN38GitDir(t, other),
		"enqueue", "t1", "--epoch", a.e2, "--candidate", x)
	// with the project named by --repo the lookup is in the project, where x does not exist: refused for that, and
	// the lease in the other repository is not what the call used
	if _, err := a.run("evidence", "t1", "--epoch", a.e2, "--kind", "gate", "--subject-sha", x, "--result", "pass", "--repo", a.repo); err == nil {
		t.Error("evidence for a commit of another repository was accepted with --repo naming the project")
	}
}

// The same repository by another route is the same repository: a linked worktree of it, another spelling of its path.
func TestAFVerbs_N38TheSameRepositoryByAnotherRouteIsAccepted(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN38Card(t)
	want := a.bound()
	wt := filepath.Join(t.TempDir(), "wt")
	afGit(t, a.repo, "worktree", "add", "-q", "--detach", wt, a.c1)
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(a.repo, link); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{wt, link, filepath.Join(a.repo, ".git")} {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			t.Setenv("CLAUDE_PROJECT_DIR", dir)
			a := a
			a.t = t
			a.record("red", a.c1, "fail")
			if got := a.bound(); got != want {
				t.Errorf("bound %q -> %q", want, got)
			}
		})
	}
	// a lease in the linked worktree is the project's repository too
	a.verb("release", "t1", "--epoch", a.e2)
	var claimed struct {
		Epoch int64 `json:"epoch"`
	}
	if err := json.Unmarshal([]byte(a.mustRun("claim", "--session", "s3", "--worktree", wt)), &claimed); err != nil || claimed.Epoch == 0 {
		t.Fatalf("claim: %+v %v", claimed, err)
	}
	a.e2 = strconv.FormatInt(claimed.Epoch, 10)
	a.verb("evidence", "t1", "--epoch", a.e2, "--kind", "green", "--subject-sha", a.c1, "--result", "pass")
	a.verb("evidence", "t1", "--epoch", a.e2, "--kind", "gate", "--subject-sha", a.c1, "--result", "pass")
	a.verb("enqueue", "t1", "--epoch", a.e2, "--candidate", a.c1)
	// and the landing from the linked worktree lands in the shared trunk
	t.Setenv("CLAUDE_PROJECT_DIR", wt)
	if out := a.verb("land", "t1", "--gate", "true"); !strings.Contains(out, `"outcome":"landed"`) {
		t.Fatalf("land from the linked worktree: %s", out)
	}
	if !a.inTrunk(a.c1) {
		t.Error("c1 landed from the linked worktree and is not in the shared trunk")
	}
}

// A project directory that is no git repository has no repository to compare: the verbs that run no git carry on.
func TestAFVerbs_N38AVerbThatRunsNoGitCarriesOnInADirectoryThatIsNoRepository(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN38Card(t)
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	if _, err := a.run("reap"); err != nil {
		t.Errorf("reap in a directory that is no repository: %v", err)
	}
	// the verbs that look commits up cannot: they need the repository
	if _, err := a.run("evidence", "t1", "--epoch", a.e2, "--kind", "red", "--subject-sha", a.c1, "--result", "fail"); err == nil {
		t.Error("evidence was accepted with no repository to look the commit up in")
	}
}
