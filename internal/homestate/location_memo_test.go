package homestate

// location_memo_test.go — the repository-location memo (Issue #1948, CLI package
// timeout).
//
// Every card operation resolves where the repository is, and each resolution was
// several git processes: a factory test cell spent about 96 of its 104 git
// processes on `rev-parse` answers about the same directory. The memo remembers an
// answer for as long as the filesystem it was read from looks the same.
//
// Defect answer: these tests catch a memo that goes stale (a repository created,
// replaced or removed under a remembered answer), one that remembers a degraded
// answer produced while git failed, one whose key forgets the repository-locating
// environment, and one that does not spare the git processes it exists to spare.
// The existing location tests (paths_*_test.go) are untouched and still pin what
// the answers are.

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gitenv"
)

// lmGit runs git in dir with a hermetic identity and fails the test on error.
func lmGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(gitenv.Env(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.invalid",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.invalid")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// lmDir is a fresh symlink-resolved temp directory (git spells paths through
// /private on macOS, and the answers are compared as resolved paths).
func lmDir(t *testing.T) string {
	t.Helper()
	d, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// lmRepo is a repository with one commit at <fresh dir>/repo.
func lmRepo(t *testing.T) string {
	t.Helper()
	repo := filepath.Join(lmDir(t), "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	lmGit(t, repo, "init", "-q", "-b", "main")
	lmGit(t, repo, "commit", "-q", "--allow-empty", "-m", "seed")
	return repo
}

// lmStarts counts the git processes a trace2 event file recorded.
func lmStarts(t *testing.T, trace string) int {
	t.Helper()
	b, err := os.ReadFile(trace)
	if err != nil {
		return 0
	}
	return strings.Count(string(b), `"event":"start"`)
}

func TestLocationMemo_RepeatedLookupsSpawnNoGit(t *testing.T) {
	repo := lmRepo(t)
	resetLocationMemo()
	trace := filepath.Join(t.TempDir(), "trace2.json")
	t.Setenv("GIT_TRACE2_EVENT", trace)

	warm := func() (string, string) {
		root := CanonicalProjectRoot(repo)
		dir, err := ProjectDir(repo)
		if err != nil {
			t.Fatal(err)
		}
		ctxRoot, err := CanonicalProjectRootContext(context.Background(), repo)
		if err != nil {
			t.Fatal(err)
		}
		ctxDir, err := ProjectDirContext(context.Background(), repo)
		if err != nil {
			t.Fatal(err)
		}
		if ctxRoot != root || ctxDir != dir {
			t.Fatalf("context variants disagree: %q %q vs %q %q", ctxRoot, ctxDir, root, dir)
		}
		return root, dir
	}
	root, dir := warm()
	spawned := lmStarts(t, trace)
	// Positive control: the instrument sees git at all, so a zero below means something.
	if spawned == 0 {
		t.Fatal("control: the first lookups spawned no git process, so the test measures nothing")
	}
	for i := 0; i < 20; i++ {
		if r, d := warm(); r != root || d != dir {
			t.Fatalf("lookup %d answered %q %q, want %q %q", i, r, d, root, dir)
		}
	}
	if after := lmStarts(t, trace); after != spawned {
		t.Errorf("20 repeated lookups spawned %d more git processes, want 0 (first lookups: %d)", after-spawned, spawned)
	}
	// What was remembered is what a cold lookup answers.
	resetLocationMemo()
	if r, d := warm(); r != root || d != dir {
		t.Errorf("a cold lookup answers %q %q, the remembered answer was %q %q", r, d, root, dir)
	}
}

// outsideRepository reads only the one failure that has a definite answer. Every
// other failure stays a degraded answer that is not remembered.
func TestOutsideRepository_OnlyTheDefiniteAnswerCounts(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"no error", nil, false},
		{"ResolveGitDirs text for a plain directory", errors.New("resolve git dirs: --absolute-git-dir failed: git rev-parse --absolute-git-dir: exited with status 128 (fatal: not a git repository (or any of the parent directories): .git)"), true},
		{"the same diagnostic with another exit status", errors.New("git rev-parse: exited with status 1 (fatal: not a git repository)"), false},
		{"status 128 for another reason (dubious ownership)", errors.New("git rev-parse: exited with status 128 (fatal: detected dubious ownership in repository at '/x')"), false},
		{"status 128, permission denied", errors.New("git rev-parse: exited with status 128 (fatal: cannot change to '/x': Permission denied)"), false},
		{"git missing", errors.New(`exec: "git": executable file not found in $PATH`), false},
		{"a timeout", context.DeadlineExceeded, false},
	}
	for _, c := range cases {
		if got := outsideRepository(c.err); got != c.want {
			t.Errorf("%s: outsideRepository = %v, want %v", c.name, got, c.want)
		}
	}
}

// A directory outside any repository is an answer too: the lookups for it are not
// repeated either (the factory tests resolve temp directories that are not
// repositories as often as ones that are).
func TestLocationMemo_NotARepositoryIsRememberedToo(t *testing.T) {
	plain := filepath.Join(lmDir(t), "plain")
	if err := os.MkdirAll(plain, 0o755); err != nil {
		t.Fatal(err)
	}
	resetLocationMemo()
	trace := filepath.Join(t.TempDir(), "trace2.json")
	t.Setenv("GIT_TRACE2_EVENT", trace)
	first := CanonicalProjectRoot(plain)
	firstGit := isGitDir(plain)
	spawned := lmStarts(t, trace)
	if spawned == 0 {
		t.Fatal("control: the first lookups spawned no git process, so the test measures nothing")
	}
	for i := 0; i < 10; i++ {
		if CanonicalProjectRoot(plain) != first || isGitDir(plain) != firstGit {
			t.Fatal("a repeated lookup answered differently")
		}
	}
	if after := lmStarts(t, trace); after != spawned {
		t.Errorf("10 repeated lookups of a directory outside any repository spawned %d more git processes, want 0 (first: %d)", after-spawned, spawned)
	}
}

// A repository created under a directory that was remembered as "not a
// repository" changes the answer for everything below it.
func TestLocationMemo_RepositoryCreatedUnderARememberedAnswer(t *testing.T) {
	base := lmDir(t)
	sub := filepath.Join(base, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	resetLocationMemo()
	if got := CanonicalProjectRoot(sub); got != sub {
		t.Fatalf("control: a directory outside any repository answers %q, want itself %q", got, sub)
	}
	lmGit(t, base, "init", "-q", "-b", "main")
	if got := CanonicalProjectRoot(sub); got != base {
		t.Errorf("after `git init` in the parent, CanonicalProjectRoot(sub) = %q, want %q (a stale remembered answer)", got, base)
	}
}

// A linked worktree replaced by a standalone repository at the same path.
func TestLocationMemo_LinkedWorktreeReplacedByAStandaloneRepository(t *testing.T) {
	primary := lmRepo(t)
	wt := filepath.Join(lmDir(t), "wt")
	lmGit(t, primary, "worktree", "add", "-q", "-b", "linked", wt)
	resetLocationMemo()
	if got := CanonicalProjectRoot(wt); got != primary {
		t.Fatalf("control: a linked worktree answers %q, want its primary checkout %q", got, primary)
	}
	lmGit(t, primary, "worktree", "remove", "--force", wt)
	if err := os.MkdirAll(wt, 0o755); err != nil {
		t.Fatal(err)
	}
	lmGit(t, wt, "init", "-q", "-b", "main")
	if got := CanonicalProjectRoot(wt); got != wt {
		t.Errorf("after the worktree was replaced by its own repository, CanonicalProjectRoot = %q, want %q (a stale remembered answer)", got, wt)
	}
}

// Audit F19: a linked worktree's `.git` FILE is unchanged when the primary repository drops the
// worktree's administrative directory (`git worktree prune`, or the directory removed by hand). The
// answer then changes (git no longer finds a repository there) while everything the fingerprint reads
// at the worktree itself stays the same.
func TestLocationMemo_LinkedWorktreeAdminDirRemovedUnderARememberedAnswer(t *testing.T) {
	primary := lmRepo(t)
	wt := filepath.Join(lmDir(t), "wt")
	lmGit(t, primary, "worktree", "add", "-q", "-b", "linked", wt)
	resetLocationMemo()
	if got := CanonicalProjectRoot(wt); got != primary {
		t.Fatalf("control: a linked worktree answers %q, want its primary checkout %q", got, primary)
	}
	admin := filepath.Join(primary, ".git", "worktrees")
	entries, err := os.ReadDir(admin)
	if err != nil || len(entries) != 1 {
		t.Fatalf("control: expected one administrative directory under %s, got %v (err %v)", admin, entries, err)
	}
	if err := os.RemoveAll(filepath.Join(admin, entries[0].Name())); err != nil {
		t.Fatal(err)
	}
	if got := CanonicalProjectRoot(wt); got == primary {
		t.Errorf("after the worktree's administrative directory was removed, CanonicalProjectRoot still answers the primary checkout %q (a stale remembered answer)", got)
	}
}

// Audit F19: a variable that changes how git discovers a repository, and that callers do not scrub,
// belongs to the key.
func TestLocationMemo_KeyCarriesTheDiscoveryEnvironment(t *testing.T) {
	t.Setenv("GIT_DISCOVERY_ACROSS_FILESYSTEM", "")
	a := locEnvKey()
	t.Setenv("GIT_DISCOVERY_ACROSS_FILESYSTEM", "1")
	if b := locEnvKey(); a == b {
		t.Error("the key ignores GIT_DISCOVERY_ACROSS_FILESYSTEM")
	}
}

// A repository whose metadata is removed.
func TestLocationMemo_RepositoryRemovedUnderARememberedAnswer(t *testing.T) {
	repo := lmRepo(t)
	sub := filepath.Join(repo, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	resetLocationMemo()
	if got := CanonicalProjectRoot(sub); got != repo {
		t.Fatalf("control: a directory in a repository answers %q, want %q", got, repo)
	}
	if err := os.RemoveAll(filepath.Join(repo, ".git")); err != nil {
		t.Fatal(err)
	}
	if got := CanonicalProjectRoot(sub); got != sub {
		t.Errorf("after the repository was removed, CanonicalProjectRoot(sub) = %q, want %q (a stale remembered answer)", got, sub)
	}
}

// isGitDir answers for a directory that becomes a bare repository.
func TestLocationMemo_DirectoryBecomesAGitDir(t *testing.T) {
	d := filepath.Join(lmDir(t), "bare")
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	resetLocationMemo()
	if isGitDir(d) {
		t.Fatal("control: an empty directory is not a git directory")
	}
	lmGit(t, d, "init", "-q", "--bare")
	if !isGitDir(d) {
		t.Error("after `git init --bare`, isGitDir = false (a stale remembered answer)")
	}
}

// An answer produced while git itself failed is a degraded answer, not a fact
// about the repository, and is never remembered.
func TestLocationMemo_DegradedAnswerIsNeverRemembered(t *testing.T) {
	primary := lmRepo(t)
	wt := filepath.Join(lmDir(t), "wt")
	lmGit(t, primary, "worktree", "add", "-q", "-b", "linked", wt)
	resetLocationMemo()

	t.Run("git unavailable", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir()) // no git on the path
		if got := CanonicalProjectRoot(wt); got != wt {
			t.Fatalf("control: with git unavailable the answer is the degraded %q, got %q", wt, got)
		}
	})
	// git is back: the answer must be the real one, not the degraded one.
	if got := CanonicalProjectRoot(wt); got != primary {
		t.Errorf("after git came back, CanonicalProjectRoot = %q, want %q (the degraded answer was remembered)", got, primary)
	}

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := CanonicalProjectRootContext(ctx, wt); err == nil {
			t.Fatal("a cancelled context must fail")
		}
	})
	if got, err := CanonicalProjectRootContext(context.Background(), wt); err != nil || got != primary {
		t.Errorf("after a cancelled lookup: %q, %v; want %q", got, err, primary)
	}
}

// The key carries the environment that decides which repository git acts on.
// The callers scrub it before spawning git, so today it cannot change an answer;
// the key keeps the dependence explicit so that a caller that stops scrubbing
// cannot be answered from another environment's entry.
func TestLocationMemo_KeyCarriesTheRepositoryLocatingEnvironment(t *testing.T) {
	base := locEnvKey()
	t.Setenv("GIT_DIR", "/one")
	one := locEnvKey()
	t.Setenv("GIT_DIR", "/two")
	two := locEnvKey()
	if one == base || two == one {
		t.Errorf("GIT_DIR must change the key: %q %q %q", base, one, two)
	}
	for _, name := range gitenv.RepoScopingVars {
		t.Setenv(name, "x-"+name)
	}
	if all := locEnvKey(); all == two {
		t.Error("the other repository-locating variables must change the key too")
	}
	before := locEnvKey()
	t.Setenv("GIT_AUTHOR_NAME", "someone")
	t.Setenv("GIT_EDITOR", "true")
	if locEnvKey() != before {
		t.Error("identity and behaviour variables do not decide which repository is acted on and must not change the key")
	}
}

// Concurrent lookups of one directory, with resets in between, under -race.
func TestLocationMemo_ConcurrentLookupsAreSafe(t *testing.T) {
	repo := lmRepo(t)
	want := CanonicalProjectRoot(repo)
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 40; i++ {
				if i%13 == 0 {
					resetLocationMemo()
				}
				if got := CanonicalProjectRoot(repo); got != want {
					t.Errorf("answered %q, want %q", got, want)
					return
				}
				_ = isGitDir(repo)
			}
		}()
	}
	wg.Wait()
}
