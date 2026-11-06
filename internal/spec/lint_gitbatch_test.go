package spec

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// lint_gitbatch_test.go — card t2087: the lint run's git cost must not grow
// with the SPEC count. These tests pin two things:
//
//  1. EQUALITY — the batched/indexed lookups return exactly what the per-SPEC
//     lookups (kept as getGitImpliedStatusPerSpec / lookupOwnershipTransitionPerSpec,
//     the oracle) return, for every SPEC of a fixture repository.
//  2. CONSTANCY — the number of git subprocesses a Lint() run spawns is the
//     same for 5 SPECs and for 50 SPECs.
//
// Tests here use os.Chdir (via chdirForTest) and a PATH shim: they MUST NOT
// call t.Parallel().

// gbRepo is a throwaway fixture repository.
type gbRepo struct {
	t   *testing.T
	dir string
}

func newGBRepo(t *testing.T) *gbRepo {
	t.Helper()
	r := &gbRepo{t: t, dir: t.TempDir()}
	r.git("init", "-q", "-b", "main", ".")
	r.git("config", "user.email", "t2087@example.com")
	r.git("config", "user.name", "t2087")
	r.git("config", "commit.gpgsign", "false")
	return r
}

func (r *gbRepo) git(args ...string) {
	r.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = r.dir
	if out, err := cmd.CombinedOutput(); err != nil {
		r.t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

// spec writes .moai/specs/<id>/spec.md with the given status and a body long
// enough that a one-line edit keeps rename similarity high.
func (r *gbRepo) spec(id, status, extra string) {
	r.t.Helper()
	r.specAs(id, id, status, extra)
}

// specAs writes the spec under directory dirID with file content naming
// contentID — a renamed directory keeps its old content, so git still sees a
// rename rather than a delete plus an unrelated add.
func (r *gbRepo) specAs(dirID, contentID, status, extra string) {
	r.t.Helper()
	id := contentID
	dir := filepath.Join(r.dir, ".moai", "specs", dirID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		r.t.Fatalf("mkdir: %v", err)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "---\nid: %s\nstatus: %s\n---\n\n# %s\n\n", id, status, id)
	for i := range 20 {
		fmt.Fprintf(&b, "body line %d of %s\n", i, id)
	}
	if extra != "" {
		b.WriteString(extra + "\n")
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(b.String()), 0o644); err != nil {
		r.t.Fatalf("write: %v", err)
	}
}

func (r *gbRepo) commit(subject, body string) {
	r.t.Helper()
	r.git("add", ".moai")
	args := []string{"commit", "-q", "-m", subject}
	if body != "" {
		args = append(args, "-m", body)
	}
	r.git(args...)
}

func (r *gbRepo) specPath(id string) string {
	return filepath.Join(r.dir, ".moai", "specs", id, "spec.md")
}

// buildFixture lays down the corpus the equality test sweeps. Every convention
// the two lookups care about appears at least once.
func (r *gbRepo) buildFixture() []string {
	ids := []string{
		"SPEC-FX-001", "SPEC-FX-002", "SPEC-FX-HARNESS-001", "SPEC-FX-HARNESS-NAMESPACE-001",
		"SPEC-FX-COMB-A-001", "SPEC-FX-COMB-B-001", "SPEC-FX-NEW-001", "SPEC-FX-REN-NEW-001",
		"SPEC-FX-WIN2-001", "SPEC-FX-WIN3-001", "SPEC-FX-WIN4-001", "SPEC-FX-UNTRACKED-001",
	}

	// plan -> run -> sync lifecycle with Authored-By-Agent trailers
	r.spec("SPEC-FX-001", "draft", "")
	r.commit("feat(SPEC-FX-001): plan-phase artifacts (tier M, 3 artifacts)", "Authored-By-Agent: manager-spec")
	r.spec("SPEC-FX-001", "in-progress", "")
	r.commit("feat(SPEC-FX-001): M1 core", "Authored-By-Agent: manager-develop")

	// no ID in the subject, no trailer; then a chore(spec) sweep commit
	r.spec("SPEC-FX-002", "draft", "")
	r.commit("chore: seed specs", "")
	r.spec("SPEC-FX-002", "draft", "swept")
	r.commit("chore(spec): sweep updated field SPEC-FX-002", "")

	// substring-collision pair: HARNESS-001 is only ever named inside a body
	// and inside the longer NAMESPACE-001 token
	r.spec("SPEC-FX-HARNESS-001", "draft", "")
	r.spec("SPEC-FX-HARNESS-NAMESPACE-001", "draft", "")
	r.commit("feat(SPEC-FX-HARNESS-NAMESPACE-001): namespace plan-phase", "relates to SPEC-FX-HARNESS-001")
	r.spec("SPEC-FX-HARNESS-NAMESPACE-001", "in-progress", "")
	r.commit("fix(SPEC-FX-HARNESS-NAMESPACE-001): M1 namespace", "supersedes SPEC-FX-HARNESS-001\n\nauthored-by-agent: Manager-Develop")

	r.spec("SPEC-FX-001", "implemented", "")
	r.commit("docs(SPEC-FX-001): sync-phase artifacts", "authored-by-agent: Manager-Docs")

	// combined-scope close naming only the scope prefix
	r.spec("SPEC-FX-COMB-A-001", "draft", "")
	r.spec("SPEC-FX-COMB-B-001", "draft", "")
	r.commit("chore(SPEC-FX-COMB): sync 3-phase close (A + B)", "")

	// rename 1: 100% similarity rename (no status delta in the rename commit);
	// the transition lives under the OLD path
	r.spec("SPEC-FX-OLD-001", "draft", "")
	r.commit("feat(SPEC-FX-OLD-001): plan-phase", "Authored-By-Agent: manager-spec")
	r.spec("SPEC-FX-OLD-001", "in-progress", "")
	r.commit("feat(SPEC-FX-OLD-001): M1", "Authored-By-Agent: manager-develop")
	r.git("mv", ".moai/specs/SPEC-FX-OLD-001", ".moai/specs/SPEC-FX-NEW-001")
	r.commit("chore(specs): rename SPEC-FX-OLD-001 to SPEC-FX-NEW-001", "")

	// rename 2: the rename commit itself carries the status delta
	r.spec("SPEC-FX-REN-OLD-001", "draft", "")
	r.commit("feat(SPEC-FX-REN-OLD-001): plan-phase", "Authored-By-Agent: manager-spec")
	r.git("mv", ".moai/specs/SPEC-FX-REN-OLD-001", ".moai/specs/SPEC-FX-REN-NEW-001")
	r.specAs("SPEC-FX-REN-NEW-001", "SPEC-FX-REN-OLD-001", "in-progress", "")
	r.commit("fix(SPEC-FX-REN-NEW-001): M1 rename and advance", "Authored-By-Agent: manager-develop")

	// 50-commit lookback boundary: WIN4 creation is the 51st commit touching
	// its file (out of window), WIN3 the 50th (in window), WIN2 the 56th.
	r.spec("SPEC-FX-WIN4-001", "draft", "")
	r.spec("SPEC-FX-WIN2-001", "draft", "")
	r.commit("chore: seed window specs", "Authored-By-Agent: manager-spec")
	for k := 1; k <= 55; k++ {
		if k == 2 {
			r.spec("SPEC-FX-WIN3-001", "draft", "")
		}
		note := fmt.Sprintf("edit %d", k)
		if k <= 50 {
			r.spec("SPEC-FX-WIN4-001", "draft", note)
		}
		if k >= 2 && k <= 50 {
			r.spec("SPEC-FX-WIN3-001", "draft", note)
		}
		r.spec("SPEC-FX-WIN2-001", "draft", note)
		r.commit(fmt.Sprintf("chore(spec): sweep SPEC-FX-WIN2-001 SPEC-FX-WIN3-001 SPEC-FX-WIN4-001 n=%d", k), "")
	}

	r.spec("SPEC-FX-001", "completed", "")
	r.commit("chore(SPEC-FX-001): sync-phase artifacts + 3-phase close", "Authored-By-Agent: orchestrator-direct")

	// never committed (written after the last commit)
	r.spec("SPEC-FX-UNTRACKED-001", "draft", "")

	return ids
}

// gbErrString renders an error (or its absence) with its sentinel class so a
// message-only match cannot hide a shape change.
func gbErrString(err error) string {
	if err == nil {
		return "<nil>"
	}
	class := "other"
	switch {
	case errors.Is(err, errGitQueryFailed):
		class = "git-query-failed"
	case errors.Is(err, errNoGitHistory):
		class = "no-history"
	case errors.Is(err, errNoClassifiableCommit):
		class = "no-classifiable"
	}
	return class + ": " + err.Error()
}

func gbOwnershipString(rec *ownershipTransitionRecord, err error) string {
	if err != nil {
		return "ERR " + err.Error()
	}
	if rec == nil {
		return "<nil record>"
	}
	return fmt.Sprintf("%+v", *rec)
}

// TestLintGitBatch_EqualsPerSpecOracle is the characterization test: inside a
// Lint() run (per-run cache active) both lookups must return exactly what the
// per-SPEC oracle returns.
func TestLintGitBatch_EqualsPerSpecOracle(t *testing.T) {
	r := newGBRepo(t)
	ids := r.buildFixture()
	sub := filepath.Join(r.dir, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	cwds := map[string]string{"repo-root": r.dir, "subdir": sub}
	for name, cwd := range cwds {
		t.Run(name, func(t *testing.T) {
			chdirForTest(t, cwd)

			// Oracle values, computed with no per-run cache active.
			wantStatus := map[string]string{}
			wantOwn := map[string]string{}
			for _, id := range ids {
				s, err := getGitImpliedStatusPerSpec(id)
				wantStatus[id] = s + " | " + gbErrString(err)
				rec, oerr := lookupOwnershipTransitionPerSpec(r.specPath(id), id)
				wantOwn[id] = gbOwnershipString(rec, oerr)
			}

			startGitQueryCache()
			defer stopGitQueryCache()
			for _, id := range ids {
				s, err := getGitImpliedStatus(id)
				if got := s + " | " + gbErrString(err); got != wantStatus[id] {
					t.Errorf("%s status: batched %q, oracle %q", id, got, wantStatus[id])
				}
				rec, oerr := lookupOwnershipTransitionFromGit(r.specPath(id), id)
				if got := gbOwnershipString(rec, oerr); got != wantOwn[id] {
					t.Errorf("%s ownership: batched %q, oracle %q", id, got, wantOwn[id])
				}
			}

			// Equality is vacuous if the batch silently fell back to the per-SPEC
			// walk: assert the batched read ran and succeeded.
			gitQueryCacheMu.RLock()
			b := gitQueryCacheV.ownershipBatches[".moai/specs"]
			indexSet := gitQueryCacheV.commitIndexSet
			gitQueryCacheMu.RUnlock()
			if b == nil || b.err != nil || len(b.segs) == 0 {
				t.Errorf("ownership batch for .moai/specs not built successfully: %+v", b)
			}
			if !indexSet {
				t.Error("commit index was never built")
			}
		})
	}
}

// TestLintGitBatch_FixtureCoversConventions guards the fixture itself: if the
// oracle stopped distinguishing these cases the equality test would be vacuous.
func TestLintGitBatch_FixtureCoversConventions(t *testing.T) {
	r := newGBRepo(t)
	r.buildFixture()
	chdirForTest(t, r.dir)

	seenStatus := map[string]bool{}
	seenShape := map[string]bool{}
	for _, id := range []string{
		"SPEC-FX-001", "SPEC-FX-002", "SPEC-FX-HARNESS-001", "SPEC-FX-HARNESS-NAMESPACE-001",
		"SPEC-FX-WIN2-001", "SPEC-FX-UNTRACKED-001",
	} {
		s, err := getGitImpliedStatusPerSpec(id)
		seenStatus[s] = true
		seenShape[strings.SplitN(gbErrString(err), ":", 2)[0]] = true
	}
	for _, want := range []string{"completed", "implemented"} {
		if !seenStatus[want] {
			t.Errorf("fixture never yields git-implied status %q (saw %v)", want, seenStatus)
		}
	}
	for _, want := range []string{"<nil>", "no-history", "no-classifiable"} {
		if !seenShape[want] {
			t.Errorf("fixture never yields error shape %q (saw %v)", want, seenShape)
		}
	}

	found := map[string]bool{}
	for _, id := range []string{"SPEC-FX-NEW-001", "SPEC-FX-REN-NEW-001", "SPEC-FX-WIN3-001", "SPEC-FX-WIN4-001", "SPEC-FX-UNTRACKED-001"} {
		rec, err := lookupOwnershipTransitionPerSpec(r.specPath(id), id)
		if err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		found[id] = rec != nil
	}
	if !found["SPEC-FX-NEW-001"] || !found["SPEC-FX-REN-NEW-001"] || !found["SPEC-FX-WIN3-001"] {
		t.Errorf("oracle must find transitions for renamed/in-window specs, got %v", found)
	}
	if found["SPEC-FX-WIN4-001"] || found["SPEC-FX-UNTRACKED-001"] {
		t.Errorf("oracle must find none for out-of-window/untracked specs, got %v", found)
	}
	rec, _ := lookupOwnershipTransitionPerSpec(r.specPath("SPEC-FX-REN-NEW-001"), "SPEC-FX-REN-NEW-001")
	if rec == nil || rec.AuthoredByAgent != "manager-develop" || rec.CurrentStatus != "in-progress" {
		t.Errorf("rename+advance commit not read as expected: %+v", rec)
	}
	if !reflect.DeepEqual(rec.PreviousStatus, "draft") {
		t.Errorf("PreviousStatus = %q, want draft", rec.PreviousStatus)
	}
}

// gbShim installs a `git` wrapper on PATH that appends one line per invocation
// to a log and then execs the real git. It returns the log path.
func gbShim(t *testing.T) string {
	t.Helper()
	// exec.LookPath on Windows resolves an extensionless name only through
	// PATHEXT, so this `#!/bin/sh` shim is never found and the real git.exe
	// runs unobserved. A .cmd shim's argument quoting is unmeasured there.
	if runtime.GOOS == "windows" {
		t.Skip("PATH shim is an extensionless sh script, not resolvable by exec.LookPath on Windows")
	}
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Skipf("git not available: %v", err)
	}
	shimDir := t.TempDir()
	logPath := filepath.Join(t.TempDir(), "git-calls.log")
	script := "#!/bin/sh\necho \"$1\" >> \"" + logPath + "\"\nexec \"" + realGit + "\" \"$@\"\n"
	if err := os.WriteFile(filepath.Join(shimDir, "git"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return logPath
}

func gbCountCalls(t *testing.T, logPath string) int {
	t.Helper()
	data, err := os.ReadFile(logPath)
	if os.IsNotExist(err) {
		return 0
	}
	if err != nil {
		t.Fatal(err)
	}
	return len(strings.Split(strings.TrimRight(string(data), "\n"), "\n"))
}

// lintGitCalls runs one Lint() over n draft SPECs, each with its own plan
// commit, and returns how many git subprocesses the run spawned.
func lintGitCalls(t *testing.T, n int) int {
	t.Helper()
	r := newGBRepo(t)
	var paths []string
	for i := 1; i <= n; i++ {
		id := fmt.Sprintf("SPEC-CNT-%03d", i)
		r.spec(id, "draft", "")
		// unreachableSpecMD-shaped content is not needed: the rules under
		// test read only id/status; reuse the full fixture so Lint parses it.
		if err := os.WriteFile(r.specPath(id), []byte(unreachableSpecMD(id, "draft")), 0o644); err != nil {
			t.Fatal(err)
		}
		r.commit(fmt.Sprintf("feat(%s): plan-phase artifacts (tier S, 2 artifacts)", id), "Authored-By-Agent: manager-spec")
		paths = append(paths, r.specPath(id))
	}
	chdirForTest(t, r.dir)

	logPath := gbShim(t)
	report := runLintOnPaths(t, false, paths...)
	if report == nil {
		t.Fatal("nil report")
	}
	return gbCountCalls(t, logPath)
}

// TestLintGitBatch_ConstantSubprocessCount is the card-t2087 proof: linting 5
// SPECs and linting 50 SPECs spawn the same number of git subprocesses.
func TestLintGitBatch_ConstantSubprocessCount(t *testing.T) {
	small := lintGitCalls(t, 5)
	large := lintGitCalls(t, 50)

	if small == 0 {
		t.Fatal("shim saw no git invocations at all — the counter is not observing the lint run")
	}
	if small != large {
		t.Errorf("git invocations: N=5 -> %d, N=50 -> %d; want equal (count must not grow with the SPEC count)", small, large)
	}
	if large > 12 {
		t.Errorf("git invocations = %d, want a small fixed number (<= 12)", large)
	}
}

// advanceFixture commits a SPEC that goes draft -> in-progress under the wrong
// actor, so the ownership walk has a transition record to find.
func (r *gbRepo) advanceFixture(id string) {
	r.spec(id, "draft", "")
	r.commit("feat("+id+"): plan-phase", "Authored-By-Agent: manager-spec")
	r.spec(id, "in-progress", "")
	r.commit("fix("+id+"): M1", "Authored-By-Agent: manager-docs")
}

// TestLintGitBatch_DiffRelativeFromSubdir pins F2: a user git config
// diff.relative=true makes patch paths relative to the cwd; the batch must still
// return the per-SPEC oracle's record instead of silently dropping it.
func TestLintGitBatch_DiffRelativeFromSubdir(t *testing.T) {
	r := newGBRepo(t)
	const id = "SPEC-REL-001"
	r.advanceFixture(id)
	r.git("config", "diff.relative", "true")
	chdirForTest(t, filepath.Join(r.dir, ".moai", "specs"))

	want, werr := lookupOwnershipTransitionPerSpec(r.specPath(id), id)
	if werr != nil || want == nil {
		t.Fatalf("oracle must find the transition: rec=%v err=%v", want, werr)
	}

	startGitQueryCache()
	defer stopGitQueryCache()
	got, gerr := lookupOwnershipTransitionFromGit(r.specPath(id), id)
	if g, w := gbOwnershipString(got, gerr), gbOwnershipString(want, werr); g != w {
		t.Errorf("batched %q, oracle %q", g, w)
	}
}

// TestLintGitBatch_OldGitRejectingNoRelativeFallsBack pins the F2 safety net: a
// git that does not know --no-relative fails the batch, and the lookup must then
// still return the per-SPEC oracle's record.
func TestLintGitBatch_OldGitRejectingNoRelativeFallsBack(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PATH shim is an extensionless sh script, not resolvable by exec.LookPath on Windows")
	}
	r := newGBRepo(t)
	const id = "SPEC-OLD-001"
	r.advanceFixture(id)
	chdirForTest(t, r.dir)

	want, werr := lookupOwnershipTransitionPerSpec(r.specPath(id), id)
	if werr != nil || want == nil {
		t.Fatalf("oracle must find the transition: rec=%v err=%v", want, werr)
	}

	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Skipf("git not available: %v", err)
	}
	shimDir := t.TempDir()
	script := "#!/bin/sh\nfor a in \"$@\"; do\n  if [ \"$a\" = \"--no-relative\" ]; then echo \"error: unknown option\" >&2; exit 129; fi\ndone\nexec \"" + realGit + "\" \"$@\"\n"
	if err := os.WriteFile(filepath.Join(shimDir, "git"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	startGitQueryCache()
	defer stopGitQueryCache()
	got, gerr := lookupOwnershipTransitionFromGit(r.specPath(id), id)
	if g, w := gbOwnershipString(got, gerr), gbOwnershipString(want, werr); g != w {
		t.Errorf("fallback %q, oracle %q", g, w)
	}
	gitQueryCacheMu.RLock()
	b := gitQueryCacheV.ownershipBatches[".moai/specs"]
	gitQueryCacheMu.RUnlock()
	if b == nil || b.err == nil {
		t.Errorf("batch must have failed to exercise the fallback, got %+v", b)
	}
}

// TestLintGitBatch_QuotedPathHeaderClosesSegment pins F3: a `diff --git "a/..."`
// header (git quotes a path containing a double quote) must close the previous
// file's segment, so its status lines are not credited to the SPEC before it.
func TestLintGitBatch_QuotedPathHeaderClosesSegment(t *testing.T) {
	r := newGBRepo(t)
	const (
		plain  = "SPEC-A-001"
		quoted = `SPEC-Z"Q-001`
	)
	r.spec(plain, "draft", "")
	r.spec(quoted, "draft", "")
	r.commit("feat: plan-phase", "Authored-By-Agent: manager-spec")
	r.spec(plain, "draft", "body edit only")
	r.spec(quoted, "in-progress", "")
	r.commit("fix: M1", "Authored-By-Agent: manager-docs")
	chdirForTest(t, r.dir)

	want, werr := lookupOwnershipTransitionPerSpec(r.specPath(plain), plain)
	if werr != nil {
		t.Fatalf("oracle: %v", werr)
	}
	// The body-only commit carries no status delta for SPEC-A-001, so the answer
	// is its creation commit; a segment merged with the quoted file would answer
	// with the later "fix: M1" commit instead.
	if want == nil || want.CommitSubject != "feat: plan-phase" {
		t.Fatalf("oracle must resolve to the creation commit, got %+v", want)
	}

	startGitQueryCache()
	defer stopGitQueryCache()
	got, gerr := lookupOwnershipTransitionFromGit(r.specPath(plain), plain)
	if g, w := gbOwnershipString(got, gerr), gbOwnershipString(want, werr); g != w {
		t.Errorf("batched %q, oracle %q", g, w)
	}
}
