package cli

// Audit of a36ea8c368 (final code audit of the integration tree), N61 N62 N65 N67 N68 N64, through the real `moai af`
// verbs, a real ledger and real git. The ledger's idea of "the repository" was a path (the git directory) plus, for a move,
// the presence of the commits its rows name: a clone, a mirror, an unrelated repository or a re-created repository
// standing at the path passed for the repository. The repair is one mechanism, a repository identity marker: a random
// token kept inside the repository's git directory, recorded in the ledger at the first bind and compared by every verb
// and by move-repository. A rename carries it; a clone, a mirror and a repository made afterwards do not have it. What
// each test asserts is the refusal and the unchanged ledger, and each has a control: the same call goes through once the
// one thing it was refused for is put right.

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// afIDMarkerFile is the name of the marker inside the common git directory of the repository.
const afIDMarkerFile = "moai-af-repository-id"

var afIDMarkerShape = regexp.MustCompile(`^[0-9a-f]{32}$`)

// afIDMarkerOf is the marker file of the repository at dir ("" when it has none).
func afIDMarkerOf(t *testing.T, dir string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(afN38GitDir(t, dir), afIDMarkerFile))
	if os.IsNotExist(err) {
		return ""
	}
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(b))
}

// afIDRaw runs one statement on the ledger file straight through the engine, outside the ledger package.
func afIDRaw(t *testing.T, db string, stmts ...string) {
	t.Helper()
	raw, err := sql.Open("sqlite", "file:"+db)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = raw.Close() }()
	for _, q := range stmts {
		if _, err := raw.Exec(q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
}

// afIDLedgerMarker is the marker the ledger file records ("" when it records none or has no such table).
func afIDLedgerMarker(t *testing.T, db string) string {
	t.Helper()
	raw, err := sql.Open("sqlite", "file:"+db+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = raw.Close() }()
	var m sql.NullString
	if err := raw.QueryRow(`SELECT marker FROM repository_marker WHERE id=1`).Scan(&m); err != nil {
		return ""
	}
	return m.String
}

func afIDUserVersion(t *testing.T, db string) int {
	t.Helper()
	raw, err := sql.Open("sqlite", "file:"+db+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = raw.Close() }()
	var v int
	if err := raw.QueryRow(`PRAGMA user_version`).Scan(&v); err != nil {
		t.Fatal(err)
	}
	return v
}

// afIDForget makes the ledger what a build before the marker left behind: bound, with no marker on record.
func afIDForget(t *testing.T, db string) {
	t.Helper()
	afIDRaw(t, db, `DROP TABLE IF EXISTS repository_marker`)
}

// afIDRenamed is afN28Landed with the repository renamed on disk, as afN52Renamed does, but built from a given card.
func afIDRenamed(t *testing.T, base afN3Card) afN52 {
	t.Helper()
	a := afN52{afN3Card: base}
	a.oldID = a.bound()
	a.moved = a.repo + "-moved"
	if err := os.Rename(a.repo, a.moved); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", a.moved)
	return a
}

// afIDFakeGit puts a git in front of the real one in PATH that misbehaves for the commit check (cat-file --batch-check)
// and nothing else: "fail" exits 128, "empty" answers nothing and exits 0, "extra" answers one line more than asked.
func afIDFakeGit(t *testing.T, mode string) {
	t.Helper()
	real, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	body := map[string]string{
		"fail":  `echo 'fatal: injected' >&2; exit 128`,
		"empty": `cat >/dev/null; exit 0`,
		"extra": `cat >/dev/null; echo 'x commit 1'; echo 'y commit 1'; exit 0`,
	}[mode]
	script := "#!/bin/sh\nfor a in \"$@\"; do case \"$a\" in --batch-check) " + body + ";; esac; done\nexec " + real + " \"$@\"\n"
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// afIDMoveResult is what the move verb prints.
type afIDMoveResult struct {
	From, To, Reason string
	CommitsChecked   int    `json:"commits_checked"`
	CommitsUnchecked int    `json:"commits_unchecked"`
	LedgerMovedFrom  string `json:"ledger_moved_from"`
}

func afIDResult(t *testing.T, out string) afIDMoveResult {
	t.Helper()
	var res afIDMoveResult
	if err := json.Unmarshal([]byte(out), &res); err != nil {
		t.Fatalf("stdout %.300q is not the verb's JSON line: %v", out, err)
	}
	return res
}

// ---- the marker itself --------------------------------------------------------------------------------------------------

func TestAFIdentity_TheFirstVerbGivesTheRepositoryAMarkerTheLedgerRecords(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN28Landed(t)
	m := afIDMarkerOf(t, a.repo)
	if !afIDMarkerShape.MatchString(m) {
		t.Fatalf("the marker in the git directory is %q, want 128 bits as 32 lowercase hex digits", m)
	}
	if got := afIDLedgerMarker(t, a.db); got != m {
		t.Errorf("the ledger records %q, the git directory holds %q", got, m)
	}
	if runtime.GOOS != "windows" {
		fi, err := os.Stat(filepath.Join(afN38GitDir(t, a.repo), afIDMarkerFile))
		if err != nil {
			t.Fatal(err)
		}
		if fi.Mode().Perm() != 0o600 {
			t.Errorf("marker mode %v, want 0600", fi.Mode().Perm())
		}
	}
	// a second verb changes nothing: neither the file nor the record
	a.mustRun("reap")
	if afIDMarkerOf(t, a.repo) != m || afIDLedgerMarker(t, a.db) != m {
		t.Error("a second verb changed the marker")
	}
}

// Ledgers that start together on one repository converge on the one marker the repository ends with.
func TestAFIdentity_ConcurrentFirstVerbsOnOneRepositoryConvergeOnOneMarker(t *testing.T) {
	afSkipWithoutSh(t)
	repo, _ := afN38Other(t)
	t.Setenv("CLAUDE_PROJECT_DIR", repo)
	const n = 8
	dbs := make([]string, n)
	for i := range dbs {
		dbs[i] = filepath.Join(t.TempDir(), "af", "ledger.db")
	}
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := range dbs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _, errs[i] = afRunBoth("reap", "--db", dbs[i], "--bind-repo", repo)
		}(i)
	}
	wg.Wait()
	want := afIDMarkerOf(t, repo)
	if !afIDMarkerShape.MatchString(want) {
		t.Fatalf("marker %q", want)
	}
	for i := range dbs {
		if errs[i] != nil {
			t.Errorf("verb %d: %v", i, errs[i])
		}
		if got := afIDLedgerMarker(t, dbs[i]); got != want {
			t.Errorf("ledger %d records %q, the repository holds %q", i, got, want)
		}
	}
}

// A linked worktree shares the git directory of its repository, so it shares the marker.
func TestAFIdentity_ALinkedWorktreeHasTheMarkerOfItsRepository(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN28Landed(t)
	m := afIDMarkerOf(t, a.repo)
	wt := filepath.Join(t.TempDir(), "wt")
	afGit(t, a.repo, "worktree", "add", "-q", "--detach", wt, a.c1)
	t.Setenv("CLAUDE_PROJECT_DIR", wt)
	a.mustRun("reap")
	if got := afIDMarkerOf(t, a.repo); got != m {
		t.Errorf("a verb from a linked worktree changed the marker: %q, want %q", got, m)
	}
	if _, err := os.Stat(filepath.Join(wt, ".git", afIDMarkerFile)); err == nil {
		t.Error("a marker was written beside the linked worktree's own git file")
	}
}

// ---- a ledger bound before markers existed adopts the marker of the repository it is bound to -----------------------------

func TestAFIdentity_ALedgerBoundBeforeMarkersAdoptsTheMarkerOfItsRepositoryOnTheNextVerb(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN28Landed(t)
	afIDForget(t, a.db)
	if got := afIDLedgerMarker(t, a.db); got != "" {
		t.Fatalf("the fixture still records %q", got)
	}
	a.mustRun("reap")
	want := afIDMarkerOf(t, a.repo)
	if !afIDMarkerShape.MatchString(want) {
		t.Fatalf("the repository holds %q", want)
	}
	if got := afIDLedgerMarker(t, a.db); got != want {
		t.Errorf("the ledger records %q after a verb, want the repository's %q", got, want)
	}
}

// ---- N61 form 1: the repository is away and the call stands in a clone --------------------------------------------------

func TestAFIdentity_N61AMoveToACloneWhileTheRepositoryIsAwayIsRefused(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN52Renamed(t)
	for name, args := range map[string][]string{
		"clone":  {"clone", "-q", "--no-hardlinks", a.moved},
		"mirror": {"clone", "-q", "--mirror", a.moved},
	} {
		t.Run(name, func(t *testing.T) {
			dst := filepath.Join(t.TempDir(), "copy")
			if out, err := exec.Command("git", append(args, dst)...).CombinedOutput(); err != nil {
				t.Fatalf("%v %s", err, out)
			}
			before := a.state()
			out, err := a.move("--reason", "moved", "--bind-repo", dst)
			a.wantRefused(before, out, err, "identity marker")
			if m := afIDMarkerOf(t, a.moved); strings.Contains(err.Error(), m) {
				t.Errorf("the refusal prints the marker value: %v", err)
			}
		})
	}
	// the control: the repository that was renamed carries its marker with it, and the same call goes through
	out, err := a.move("--reason", "moved")
	a.wantMoved(out, err, a.moved)
}

// ---- N61 form 2: an unrelated repository, no row names a commit ----------------------------------------------------------

func TestAFIdentity_N61AMoveToAnUnrelatedRepositoryWhenNoRowNamesACommitIsRefused(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, c1, st, epoch := afUnqueuedCard(t)
	base := afN3Card{t: t, db: db, repo: repo, c1: c1, st: st, e2: "1"}
	_ = epoch
	if out, errOut, err := afRunBoth("reap", "--db", db, "--bind-repo", repo); err != nil {
		t.Fatalf("reap: %v %s %s", err, out, errOut)
	}
	if rows, err := afRowCommits(context.Background(), st); err != nil || len(rows) != 0 {
		t.Fatalf("the fixture's rows name commits %v %v", rows, err)
	}
	a := afIDRenamed(t, base)
	unrelated := afN52Fresh(t)
	before := a.state()
	out, err := a.move("--reason", "moved", "--bind-repo", unrelated)
	a.wantRefused(before, out, err, "identity marker")
	// the control
	out, err = a.move("--reason", "moved", "--bind-repo", a.moved)
	a.wantMoved(out, err, a.moved)
}

// ---- N61 form 3: the ledger file sits in one repository, the call names another -----------------------------------------

func afIDFileInRepo(t *testing.T) (a afN52, clone string, dbMoved string) {
	t.Helper()
	repo, x := afN49Other(t)
	db := filepath.Join(repo, ".moai", "af", "ledger.db")
	st := afN49Ledger(t, db)
	t.Setenv("CLAUDE_PROJECT_DIR", repo)
	ep := afN49Claim(t, db, repo)
	out, errOut, err := afRunBoth("evidence", "t1", "--db", db, "--epoch", ep, "--kind", "green", "--subject-sha", x, "--result", "pass")
	if err != nil {
		t.Fatalf("evidence: %v %s %s", err, out, errOut)
	}
	_ = st.Close()
	clone = filepath.Join(t.TempDir(), "clone")
	if out, err := exec.Command("git", "clone", "-q", "--no-hardlinks", repo, clone).CombinedOutput(); err != nil {
		t.Fatalf("clone: %v %s", err, out)
	}
	moved := repo + "-moved"
	if err := os.Rename(repo, moved); err != nil {
		t.Fatal(err)
	}
	dbMoved = filepath.Join(moved, ".moai", "af", "ledger.db")
	st2, err := ledger.Open(context.Background(), dbMoved)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st2.Close() })
	t.Setenv("CLAUDE_PROJECT_DIR", moved)
	a = afN52{afN3Card: afN3Card{t: t, db: dbMoved, repo: repo, c1: x, st: st2}, moved: moved, oldID: afN38GitDir(t, moved)}
	a.oldID = a.bound()
	return a, clone, dbMoved
}

func TestAFIdentity_N61ALedgerFileInOneRepositoryIsNotMovedToAnother(t *testing.T) {
	afSkipWithoutSh(t)
	t.Run("a ledger with a marker", func(t *testing.T) {
		a, clone, _ := afIDFileInRepo(t)
		before := a.state()
		out, err := a.move("--reason", "moved", "--bind-repo", clone)
		a.wantRefused(before, out, err, "identity marker")
		out, err = a.move("--reason", "moved")
		a.wantMoved(out, err, a.moved)
	})
	t.Run("a ledger without one asks where the ledger file sits", func(t *testing.T) {
		a, clone, _ := afIDFileInRepo(t)
		afIDForget(t, a.db)
		before := a.state()
		out, err := a.move("--reason", "moved", "--bind-repo", clone)
		a.wantRefused(before, out, err, "sits in")
		out, err = a.move("--reason", "moved", "--bind-repo", a.moved)
		a.wantMoved(out, err, a.moved)
	})
}

// ---- a repository re-created at the same path, and a copy ----------------------------------------------------------------

func TestAFIdentity_ARepositoryRecreatedAtTheSamePathIsNotTheOneTheLedgerBelongsTo(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN28Landed(t)
	if err := os.RemoveAll(a.repo); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(a.repo, 0o755); err != nil {
		t.Fatal(err)
	}
	afGit(t, a.repo, "init", "-q", "-b", "work")
	boundBefore, rowsBefore := a.bound(), afSnapshot(t, a.st)
	for _, args := range [][]string{{"reap"}, {"add", "t9"}, {"land", "t1", "--gate", "true"}} {
		out, err := a.run(args...)
		if err == nil || !strings.Contains(err.Error()+out, "identity marker") {
			t.Errorf("moai af %s: %v %.200s, want a refusal that names the identity marker", args[0], err, out)
		}
	}
	if a.bound() != boundBefore || !reflect.DeepEqual(rowsBefore, afSnapshot(t, a.st)) {
		t.Error("a refused verb changed the ledger")
	}
	if afIDMarkerOf(t, a.repo) != "" {
		t.Error("a refused verb gave the stranger a marker")
	}
}

func TestAFIdentity_ACopyIsAMoveOnlyOnceTheOriginalIsGone(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN28Landed(t)
	copyDir := a.repo + "-copy"
	if out, err := exec.Command("cp", "-R", a.repo, copyDir).CombinedOutput(); err != nil {
		t.Fatalf("cp: %v %s", err, out)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", copyDir)
	b := afN52{afN3Card: a, moved: copyDir, oldID: a.bound()}
	before := b.state()
	out, err := b.move("--reason", "copied")
	b.wantRefused(before, out, err, "still exists")
	if err := os.RemoveAll(a.repo); err != nil {
		t.Fatal(err)
	}
	out, err = b.move("--reason", "copied")
	b.wantMoved(out, err, copyDir)
}

// ---- N62: a git that cannot answer is a refusal ----------------------------------------------------------------------------

func TestAFIdentity_N62TheFirstBindRefusesWhenGitCannotAnswerTheCommitCheck(t *testing.T) {
	afSkipWithoutSh(t)
	for _, mode := range []string{"fail", "empty", "extra"} {
		t.Run(mode, func(t *testing.T) {
			db, repo, c1, st, epoch := afUnqueuedCard(t)
			if err := st.AddEvidence(context.Background(), "t1", epoch, ledger.Evidence{Kind: ledger.KindGreen, TreeSHA: "t", SubjectSHA: c1, Result: ledger.ResultPass, Command: "x"}); err != nil {
				t.Fatal(err)
			}
			afIDFakeGit(t, mode)
			out, errOut, err := afRunBoth("reap", "--db", db, "--bind-repo", repo)
			if err == nil || !strings.Contains(err.Error(), "could not answer") {
				t.Fatalf("reap with a git that cannot answer: %v %.200s %.200s, want a refusal that says git could not answer", err, out, errOut)
			}
			if got, _ := st.Repository(context.Background()); got != "" {
				t.Errorf("the refused call bound the ledger to %q", got)
			}
		})
	}
}

func TestAFIdentity_N62TheFirstBindRefusesWhenAnObjectIsDamaged(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, c1, st, epoch := afUnqueuedCard(t)
	if err := st.AddEvidence(context.Background(), "t1", epoch, ledger.Evidence{Kind: ledger.KindGreen, TreeSHA: "t", SubjectSHA: c1, Result: ledger.ResultPass, Command: "x"}); err != nil {
		t.Fatal(err)
	}
	obj := filepath.Join(afN38GitDir(t, repo), "objects", c1[:2], c1[2:])
	if err := os.Chmod(obj, 0o644); err != nil {
		t.Skipf("the commit is not a loose object here: %v", err)
	}
	if err := os.WriteFile(obj, []byte("not a zlib stream"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, errOut, err := afRunBoth("reap", "--db", db, "--bind-repo", repo)
	if err == nil {
		t.Fatalf("reap over a damaged commit object was accepted: %.200s %.200s", out, errOut)
	}
	if got, _ := st.Repository(context.Background()); got != "" {
		t.Errorf("the refused call bound the ledger to %q", got)
	}
}

func TestAFIdentity_N62AMoveOfAnUnmarkedLedgerRefusesWhenGitCannotAnswer(t *testing.T) {
	afSkipWithoutSh(t)
	for _, mode := range []string{"fail", "empty", "extra"} {
		t.Run(mode, func(t *testing.T) {
			a := afN52Renamed(t)
			afIDForget(t, a.db)
			path := os.Getenv("PATH")
			afIDFakeGit(t, mode)
			before := a.state()
			out, err := a.move("--reason", "moved", "--bind-repo", a.moved)
			a.wantRefused(before, out, err, "could not answer")
			// the control: the same call with the git that answers
			t.Setenv("PATH", path)
			out, err = a.move("--reason", "moved", "--bind-repo", a.moved)
			a.wantMoved(out, err, a.moved)
		})
	}
}

// A marked ledger needs no commit check, so a git that cannot answer is no reason to refuse its move: the move reports
// the commits it could not confirm.
func TestAFIdentity_N62AMoveOfAMarkedLedgerDoesNotDependOnTheCommitCheck(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN52Renamed(t)
	afIDFakeGit(t, "fail")
	out, err := a.move("--reason", "moved")
	a.wantMoved(out, err, a.moved)
	if res := afIDResult(t, out); res.CommitsChecked != 0 || res.CommitsUnchecked < 1 {
		t.Errorf("the verb printed %+v, want the commits it could not confirm counted as unchecked", res)
	}
}

// ---- N65: a lease the caller recorded is the caller's own word -------------------------------------------------------------

func TestAFIdentity_N65ACloneBoundThroughALeaseNeedsADeclaration(t *testing.T) {
	afSkipWithoutSh(t)
	p, x := afN38Other(t)
	clone := filepath.Join(t.TempDir(), "clone")
	if out, err := exec.Command("git", "clone", "-q", "--no-hardlinks", p, clone).CombinedOutput(); err != nil {
		t.Fatalf("clone: %v %s", err, out)
	}
	db := filepath.Join(t.TempDir(), "central", "ledger.db") // in no repository
	nonrepo := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", nonrepo)
	if out, errOut, err := afRunBoth("add", "t1", "--db", db, "--path", "a.txt", "--class", "A"); err != nil {
		t.Fatalf("add: %v %s %s", err, out, errOut)
	}
	ep := afN49Claim(t, db, clone)
	st, err := ledger.Open(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if got := afN49Bound(t, st); got != "" {
		t.Fatalf("the fixture is bound to %q before the clone speaks", got)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", clone)
	args := []string{"evidence", "t1", "--db", db, "--epoch", ep, "--kind", "green", "--subject-sha", x, "--result", "pass"}
	out, errOut, err := afRunBoth(args...)
	if err == nil || !strings.Contains(err.Error(), "--bind-repo") {
		t.Fatalf("evidence from the clone through its own lease: %v %.200s %.200s, want a refusal that asks for --bind-repo", err, out, errOut)
	}
	if got := afN49Bound(t, st); got != "" {
		t.Errorf("the refused call bound the ledger to %q", got)
	}
	// the control: the caller declares the clone and the first bind goes through
	if out, errOut, err = afRunBoth(append(args, "--bind-repo", clone)...); err != nil {
		t.Fatalf("with the declaration: %v %s %s", err, out, errOut)
	}
	if got, want := afN49Bound(t, st), afN38GitDir(t, clone); got != want {
		t.Errorf("bound %q, want %q", got, want)
	}
}

// ---- N67: a commit that was thrown away does not stop a real move -----------------------------------------------------------

// afIDPruned is a landed card that also recorded a failing run about an attempt whose commit was then thrown away and
// collected, the repository then renamed.
func afIDPruned(t *testing.T) (a afN52, attempt string) {
	t.Helper()
	base := afN28Landed(t)
	tree := afGit(t, base.repo, "rev-parse", "HEAD^{tree}")
	attempt = afGit(t, base.repo, "commit-tree", tree, "-p", base.c1, "-m", "an attempt")
	base.record("green", attempt, "fail")
	afGit(t, base.repo, "reflog", "expire", "--expire=now", "--all")
	afGit(t, base.repo, "gc", "-q", "--prune=now")
	if exec.Command("git", "-C", base.repo, "cat-file", "-e", attempt).Run() == nil {
		t.Fatal("the attempt commit is still in the repository after the collection")
	}
	return afIDRenamed(t, base), attempt
}

func TestAFIdentity_N67AMarkedLedgerMovesPastACommitThatWasPruned(t *testing.T) {
	afSkipWithoutSh(t)
	a, _ := afIDPruned(t)
	out, err := a.move("--reason", "moved")
	a.wantMoved(out, err, a.moved)
	if res := afIDResult(t, out); res.CommitsUnchecked < 1 {
		t.Errorf("the verb printed %+v, want the pruned commit counted as unchecked", res)
	}
}

func TestAFIdentity_N67AnUnmarkedLedgerMovesPastAPrunedEvidenceCommitButNotAMissingCandidate(t *testing.T) {
	afSkipWithoutSh(t)
	a, _ := afIDPruned(t)
	afIDForget(t, a.db)
	out, err := a.move("--reason", "moved", "--bind-repo", a.moved)
	a.wantMoved(out, err, a.moved)
	if res := afIDResult(t, out); res.CommitsUnchecked < 1 || res.CommitsChecked < 1 {
		t.Errorf("the verb printed %+v, want the queue candidate confirmed and the pruned evidence commit unchecked", res)
	}
}

// ---- an unmarked ledger: the move asks for a declaration and says what it could not see ------------------------------------

func TestAFIdentity_AnUnmarkedLedgerMoveNeedsADeclarationNotTheWorkingDirectory(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN52Renamed(t)
	afIDForget(t, a.db)
	before := a.state()
	out, err := a.move("--reason", "moved")
	a.wantRefused(before, out, err, "--bind-repo", "marker")
	out, err = a.move("--reason", "moved", "--bind-repo", a.moved)
	a.wantMoved(out, err, a.moved)
	if got, want := afIDLedgerMarker(t, a.db), afIDMarkerOf(t, a.moved); !afIDMarkerShape.MatchString(want) || got != want {
		t.Errorf("after the move the ledger records %q, the new repository holds %q", got, want)
	}
}

// ---- N68: the first recorded move raises the schema number ------------------------------------------------------------------

func TestAFIdentity_N68ALedgerMovedByThisBuildIsRefusedByABuildThatDoesNotKnowTheMove(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN52Renamed(t)
	const previousBuildSchema = 2
	if v := afIDUserVersion(t, a.db); v != previousBuildSchema {
		t.Fatalf("a ledger that never moved has user_version %d, want %d (readable by the build before)", v, previousBuildSchema)
	}
	out, err := a.move("--reason", "moved")
	a.wantMoved(out, err, a.moved)
	if v := afIDUserVersion(t, a.db); v <= previousBuildSchema {
		t.Errorf("user_version %d after the move: a build that knows schema %d would accept the file and the old path", v, previousBuildSchema)
	}
	// this build keeps working on the file
	a.mustRun("reap")
}

// ---- N64: the default location after the project directory moved ------------------------------------------------------------

type afIDDefault struct {
	old, moved string // the project directory, before and after the rename
	oldLedger  string
	newLedger  string
}

func afIDDefaultMoved(t *testing.T) afIDDefault {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), ".moai"))
	repo, x := afN38Other(t)
	t.Setenv("CLAUDE_PROJECT_DIR", repo)
	d := afIDDefault{old: repo, moved: repo + "-moved"}
	var err error
	if d.oldLedger, err = afLedgerPath(""); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"add", "t1", "--path", "a.txt", "--class", "A"},
		{"claim", "--session", "s1", "--worktree", repo},
		{"evidence", "t1", "--epoch", "1", "--kind", "green", "--subject-sha", x, "--result", "pass"},
	} {
		if out, errOut, err := afRunBoth(args...); err != nil {
			t.Fatalf("moai af %s: %v %s %s", args[0], err, out, errOut)
		}
	}
	if _, err := os.Stat(d.oldLedger); err != nil {
		t.Fatalf("no ledger at the default location %s: %v", d.oldLedger, err)
	}
	if err := os.Rename(repo, d.moved); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", d.moved)
	if d.newLedger, err = afLedgerPath(""); err != nil {
		t.Fatal(err)
	}
	if d.newLedger == d.oldLedger {
		t.Fatalf("the default location did not change with the project directory: %s", d.newLedger)
	}
	return d
}

func afIDExists(path string) bool { _, err := os.Lstat(path); return err == nil }

func TestAFIdentity_N64TheDefaultLedgerOfAMovedProjectIsMovedWithFrom(t *testing.T) {
	afSkipWithoutSh(t)
	d := afIDDefaultMoved(t)

	// without --from there is nothing to find at the new key; the refusal says how to name the old one
	out, errOut, err := afRunBoth("move-repository", "--reason", "moved")
	if err == nil || !strings.Contains(err.Error(), "--from") {
		t.Fatalf("without --from: %v %.200s %.200s, want a refusal that names --from", err, out, errOut)
	}
	if afIDExists(d.newLedger) || !afIDExists(d.oldLedger) {
		t.Fatalf("the refused call touched the ledgers: new exists=%v old exists=%v", afIDExists(d.newLedger), afIDExists(d.oldLedger))
	}

	out, errOut, err = afRunBoth("move-repository", "--reason", "moved", "--from", d.old)
	if err != nil {
		t.Fatalf("with --from: %v %.300s %.300s", err, out, errOut)
	}
	if res := afIDResult(t, out); res.LedgerMovedFrom != d.oldLedger {
		t.Errorf("the verb printed %+v, want ledger_moved_from %q", res, d.oldLedger)
	}
	for _, suffix := range []string{"", "-wal", "-shm"} {
		if afIDExists(d.oldLedger + suffix) {
			t.Errorf("%s is still at the old key", d.oldLedger+suffix)
		}
	}
	for _, suffix := range []string{"-wal", "-shm"} {
		if afIDExists(d.newLedger + suffix) {
			t.Errorf("%s was left beside the moved ledger (the old ledger was closed, so its log is empty)", d.newLedger+suffix)
		}
	}
	st, err := ledger.Open(context.Background(), d.newLedger)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()
	if got, want := afN49Bound(t, st), afN38GitDir(t, d.moved); got != want {
		t.Errorf("bound %q, want %q", got, want)
	}
	if moves, err := st.RepositoryMoves(context.Background()); err != nil || len(moves) != 1 {
		t.Errorf("moves %+v %v, want the one just recorded", moves, err)
	}
	if cards, err := st.Cards(context.Background()); err != nil || len(cards) != 1 || cards[0].ID != "t1" {
		t.Errorf("cards %+v %v: the card of the old ledger must be in the moved one", cards, err)
	}
	// every other verb works from the new place with no --db
	if out, errOut, err := afRunBoth("reap"); err != nil {
		t.Errorf("reap with no --db after the move: %v %s %s", err, out, errOut)
	}
}

func TestAFIdentity_N64AFromThatWouldOverwriteALedgerAtTheNewKeyIsRefused(t *testing.T) {
	afSkipWithoutSh(t)
	d := afIDDefaultMoved(t)
	// a verb ran at the new place first, and made a ledger of its own there
	if out, errOut, err := afRunBoth("reap"); err != nil {
		t.Fatalf("reap at the new place: %v %s %s", err, out, errOut)
	}
	oldBytes, newBytes := afLedgerBytes(t, d.oldLedger), afLedgerBytes(t, d.newLedger)
	out, errOut, err := afRunBoth("move-repository", "--reason", "moved", "--from", d.old)
	if err == nil {
		t.Fatalf("--from over an existing ledger was accepted: %s", out)
	}
	for _, want := range []string{d.oldLedger, d.newLedger, "nothing was changed"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not say %q: %v", want, err)
		}
	}
	if string(afLedgerBytes(t, d.oldLedger)) != string(oldBytes) || string(afLedgerBytes(t, d.newLedger)) != string(newBytes) {
		t.Errorf("the refused call changed a ledger file %s %s", out, errOut)
	}
}

func TestAFIdentity_N64AFreshLedgerAtTheNewKeyIsNotCalledTheRepositoryThatDidNotMove(t *testing.T) {
	afSkipWithoutSh(t)
	d := afIDDefaultMoved(t)
	if out, errOut, err := afRunBoth("reap"); err != nil {
		t.Fatalf("reap at the new place: %v %s %s", err, out, errOut)
	}
	_, _, err := afRunBoth("move-repository", "--reason", "moved")
	if err == nil {
		t.Fatal("move-repository over a fresh ledger was accepted")
	}
	for _, want := range []string{"fresh", "--from"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not say %q: %v", want, err)
		}
	}
	if strings.Contains(err.Error(), "so it did not move") {
		t.Errorf("the refusal still says the repository did not move: %v", err)
	}
	_ = d
}

func TestAFIdentity_N64FromIsRefusedWithADbAndWithoutALedger(t *testing.T) {
	afSkipWithoutSh(t)
	d := afIDDefaultMoved(t)
	_, _, err := afRunBoth("move-repository", "--reason", "moved", "--from", d.old, "--db", filepath.Join(t.TempDir(), "x.db"))
	if err == nil || !strings.Contains(err.Error(), "--db") {
		t.Errorf("--from with --db: %v, want a refusal that names --db", err)
	}
	_, _, err = afRunBoth("move-repository", "--reason", "moved", "--from", d.old+"-nowhere")
	if err == nil || !strings.Contains(err.Error(), "no ledger") {
		t.Errorf("--from naming a project with no ledger: %v, want a refusal that says there is no ledger", err)
	}
	if !afIDExists(d.oldLedger) || afIDExists(d.newLedger) {
		t.Error("a refused call touched the ledgers")
	}
}

// ---- the help ---------------------------------------------------------------------------------------------------------------

func TestAFIdentity_TheMoveHelpNamesFromAndTheMarker(t *testing.T) {
	flat := afHelpOf(t, "move-repository")
	for _, want := range []string{"--from", "identity marker"} {
		if !strings.Contains(flat, want) {
			t.Errorf("`moai af move-repository --help` does not say %q:\n%s", want, flat)
		}
	}
}
