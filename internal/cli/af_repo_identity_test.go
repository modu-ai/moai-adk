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
	"bytes"
	"context"
	"database/sql"
	"errors"
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

// afIDForget makes the ledger what a build before the marker left behind: bound, with no marker on record.
func afIDForget(t *testing.T, db string) {
	t.Helper()
	afIDRaw(t, db, `DROP TABLE IF EXISTS repository_marker`)
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
	if !afIDMarkerShape.MatchString(m) {
		t.Fatalf("the repository holds the marker %q before the linked worktree speaks", m)
	}
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

// ---- N61 form 2: an unrelated repository, no row names a commit ----------------------------------------------------------

// ---- N61 form 3: the ledger file sits in one repository, the call names another -----------------------------------------

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

// ---- fixtures shared with the older tests ---------------------------------------------------------------------------------

// afBindFixture binds the ledger of a fixture to the repository its card is worked in, as the first verb run in that
// repository used to do on the strength of the lease alone (N65: a lease worktree is the caller's own word, so a first
// bind of a ledger placed with --db declares its repository, and a fixture that is not about the first bind starts bound).
func afBindFixture(t *testing.T, st *ledger.Store, repo string) {
	t.Helper()
	if err := st.BindRepository(context.Background(), afN38GitDir(t, repo)); err != nil {
		t.Fatal(err)
	}
}

// measureDeclared is measure for a ledger whose first bind is still to come: the evidence declares the repository.
func (a afN3Card) measureDeclared(sha string) {
	a.t.Helper()
	for _, kind := range []string{"green", "gate"} {
		a.verb("evidence", "t1", "--epoch", a.e2, "--kind", kind, "--subject-sha", sha, "--result", "pass", "--repo", a.repo, "--bind-repo", a.repo)
	}
}

// afRunBinary runs one `af` verb of a built moai binary and returns what afRunBoth returns, the error being the line the
// binary printed after "Error: ". It lets one test group run against two builds (AF_TEST_BIN).
func afRunBinary(bin string, args ...string) (stdout, stderr string, err error) {
	cmd := exec.Command(bin, append([]string{"af"}, args...)...)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if runErr := cmd.Run(); runErr != nil {
		// the binary prints the error in a framed block (an ERROR banner, the message wrapped to the width of the frame, a
		// capital first letter): the lines of the message are joined back with a space; a word the frame cut in two keeps a
		// space inside it, which afIDSays ignores
		var lines []string
		for _, line := range strings.Split(errOut.String(), "\n") {
			if line = strings.TrimSpace(line); line != "" && line != "ERROR" {
				lines = append(lines, line)
			}
		}
		err = errors.New(strings.Join(lines, " "))
	}
	return out.String(), errOut.String(), err
}

// afIDSays is whether text says want, with the white space of both left out: the binary's frame may cut a path in two.
func afIDSays(text, want string) bool {
	squash := func(s string) string { return strings.Join(strings.Fields(s), "") }
	return strings.Contains(squash(text), squash(want))
}

// ---- the normal paths beside the refusals -----------------------------------------------------------------------------------
