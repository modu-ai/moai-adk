package hook

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Card t560 (GH #1691) — the remaining leak sites outside internal/hook/quality.
//
// git exports GIT_DIR, and on the commit path GIT_INDEX_FILE, into every hook it
// runs. Those variables OUTRANK the working directory, so `git -C dir` and
// cmd.Dir do NOT confine a child: a command asked about directory X answers
// about repository Y, and every decision downstream is taken on the wrong
// repository's state.
//
// Each test below builds TWO repositories whose observable content differs, and
// asserts the SPECIFIC value belonging to the one the function was given. An
// assertion that merely checked for a non-empty result, or counted entries,
// would pass while reading the leaked repository.
//
// The plumbing helpers are reproduced here rather than shared with
// internal/hook/quality: that package's equivalents live in a _test.go file and
// are not importable.

// ---------------------------------------------------------------------------
// test-local git plumbing (never inherits the leaked environment)
// ---------------------------------------------------------------------------

// gitEnvCleanEnv returns the current environment with every GIT_ variable
// removed, so a fixture-side command reads the repository it names rather than
// the one a leaked GIT_DIR points at.
func gitEnvCleanEnv() []string {
	out := make([]string, 0, len(os.Environ()))
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(name, "GIT_") {
			continue
		}
		out = append(out, kv)
	}
	return out
}

// gitEnvRun runs git in dir with a scrubbed environment plus extraEnv, and
// returns trimmed stdout. extraEnv survives the scrub, which is how a fixture
// pins a committer date the scrub would otherwise strip.
func gitEnvRun(t *testing.T, dir string, extraEnv []string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(gitEnvCleanEnv(), extraEnv...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s in %s: %v\n%s", strings.Join(args, " "), dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// gitEnvNewRepo creates an initialised repository holding a single committed
// file named <marker>.go with the commit subject <marker>-seed and the given
// committer date. The marker differs per repository so every observable below
// — commit subject, committer date, diffed filename — separates host from
// target by value rather than by count.
func gitEnvNewRepo(t *testing.T, dir, marker, committerDate string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	gitEnvRun(t, dir, nil, "init", "-q", "-b", "main")
	gitEnvRun(t, dir, nil, "config", "user.name", "t560")
	gitEnvRun(t, dir, nil, "config", "user.email", "t560@test.invalid")
	if err := os.WriteFile(filepath.Join(dir, marker+".go"), []byte("package "+marker+"\n"), 0o644); err != nil {
		t.Fatalf("write %s.go: %v", marker, err)
	}
	gitEnvRun(t, dir, nil, "add", marker+".go")
	gitEnvRun(t, dir, []string{
		"GIT_COMMITTER_DATE=" + committerDate,
		"GIT_AUTHOR_DATE=" + committerDate,
	}, "commit", "-q", "-m", marker+"-seed")
	return dir
}

// gitEnvLeakPreCommit reproduces the environment git hands a pre-commit hook:
// GIT_DIR names the repository being committed to and GIT_INDEX_FILE the index
// the commit is being assembled in. GIT_WORK_TREE is deliberately NOT set —
// git does not export it to hooks, and a leak model stronger than the real one
// makes children fail outright instead of quietly reading the wrong repository.
func gitEnvLeakPreCommit(t *testing.T, hostGitDir string) {
	t.Helper()
	t.Setenv("GIT_DIR", hostGitDir)
	t.Setenv("GIT_INDEX_FILE", filepath.Join(hostGitDir, "index"))
}

const (
	hostDate   = "2001-01-01T00:00:00+00:00"
	targetDate = "2002-02-02T00:00:00+00:00"
)

// ---------------------------------------------------------------------------
// internal/hook/agentmemory.go — gitOutput(dir, args...) uses `git -C dir`
// ---------------------------------------------------------------------------

// gitOutput must answer about the directory named by its -C argument. It is the
// single git seam of the agent-memory drain: revParseDirs resolves the gitdir
// and common-dir through it, and a leaked GIT_DIR makes the drain classify
// another repository's layout as the one it is draining.
func TestGitOutputReadsGivenRepoNotLeakedGitDir(t *testing.T) {
	requireGit(t)

	base := t.TempDir()
	host := gitEnvNewRepo(t, filepath.Join(base, "host"), "host", hostDate)
	target := gitEnvNewRepo(t, filepath.Join(base, "target"), "target", targetDate)

	// Fixture sanity: each repository reports its own commit subject. Without
	// this the main assertion could not separate "read the wrong repository"
	// from "read nothing at all".
	if got := gitEnvRun(t, host, nil, "log", "-1", "--format=%s"); got != "host-seed" {
		t.Fatalf("fixture broken: host subject %q, want %q", got, "host-seed")
	}
	if got := gitEnvRun(t, target, nil, "log", "-1", "--format=%s"); got != "target-seed" {
		t.Fatalf("fixture broken: target subject %q, want %q", got, "target-seed")
	}

	gitEnvLeakPreCommit(t, filepath.Join(host, ".git"))

	out, err := gitOutput(target, "log", "-1", "--format=%s")
	if err != nil {
		t.Fatalf("gitOutput: %v", err)
	}
	if got := strings.TrimSpace(out); got != "target-seed" {
		t.Errorf("gitOutput(%s) answered about the LEAKED repository: got %q, want %q\nGIT_DIR=%s GIT_INDEX_FILE=%s",
			target, got, "target-seed", os.Getenv("GIT_DIR"), os.Getenv("GIT_INDEX_FILE"))
	}
}

// ---------------------------------------------------------------------------
// internal/hook/navigator_detect.go — changedAtForProject(projectRoot)
// ---------------------------------------------------------------------------

// changedAtForProject stamps the navigator impact record's changed_at with the
// committer date of the given project's HEAD. Under a leak it stamps another
// repository's HEAD date, so two detections that should differ agree — and the
// determinism the stamp is built on asserts about the wrong tree.
//
// The two repositories carry deliberately distinct committer dates: an
// assertion that only checked the result was not the "(no-git)" sentinel would
// pass while reading the host.
func TestChangedAtForProjectReadsGivenRepoNotLeakedGitDir(t *testing.T) {
	requireGit(t)

	base := t.TempDir()
	host := gitEnvNewRepo(t, filepath.Join(base, "host"), "host", hostDate)
	target := gitEnvNewRepo(t, filepath.Join(base, "target"), "target", targetDate)

	wantHost := gitEnvRun(t, host, nil, "log", "-1", "--format=%cI")
	wantTarget := gitEnvRun(t, target, nil, "log", "-1", "--format=%cI")
	if wantHost == wantTarget {
		t.Fatalf("fixture broken: both repositories report committer date %q, so the test cannot separate them", wantHost)
	}

	gitEnvLeakPreCommit(t, filepath.Join(host, ".git"))

	got := changedAtForProject(target)
	if got != wantTarget {
		t.Errorf("changedAtForProject(%s) answered about the LEAKED repository: got %q, want %q (host is %q)\nGIT_DIR=%s",
			target, got, wantTarget, wantHost, os.Getenv("GIT_DIR"))
	}
}

// ---------------------------------------------------------------------------
// internal/hook/session_end.go — getModifiedGoFiles(ctx, projectDir)
// ---------------------------------------------------------------------------

// getModifiedGoFiles reports the .go files modified in the session. It sets
// cmd.Dir and nothing else, so a leaked GIT_DIR makes `git diff --name-only
// HEAD` resolve HEAD in the caller's repository: the returned set names files
// that belong to another tree, and the SessionEnd validator then runs over
// paths that do not exist where it was told to look.
//
// Each repository's tracked .go file is named after the repository, so the
// returned filename — not its count — is what separates a correct answer from
// a leaked one.
func TestGetModifiedGoFilesReadsGivenRepoNotLeakedGitDir(t *testing.T) {
	requireGit(t)

	base := t.TempDir()
	host := gitEnvNewRepo(t, filepath.Join(base, "host"), "host", hostDate)
	target := gitEnvNewRepo(t, filepath.Join(base, "target"), "target", targetDate)

	// Modify the tracked file in each repository so both have something to report.
	for _, r := range []struct{ dir, marker string }{{host, "host"}, {target, "target"}} {
		path := filepath.Join(r.dir, r.marker+".go")
		if err := os.WriteFile(path, []byte("package "+r.marker+"\n\n// modified\n"), 0o644); err != nil {
			t.Fatalf("modify %s: %v", path, err)
		}
	}

	if got := gitEnvRun(t, host, nil, "diff", "--name-only", "HEAD"); got != "host.go" {
		t.Fatalf("fixture broken: host modified set %q, want %q", got, "host.go")
	}
	if got := gitEnvRun(t, target, nil, "diff", "--name-only", "HEAD"); got != "target.go" {
		t.Fatalf("fixture broken: target modified set %q, want %q", got, "target.go")
	}

	gitEnvLeakPreCommit(t, filepath.Join(host, ".git"))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	got := getModifiedGoFiles(ctx, target)
	want := []string{filepath.Join(target, "target.go")}
	if len(got) != len(want) || (len(got) == 1 && got[0] != want[0]) {
		t.Errorf("getModifiedGoFiles(%s) answered about the LEAKED repository: got %v, want %v\nGIT_DIR=%s GIT_INDEX_FILE=%s",
			target, got, want, os.Getenv("GIT_DIR"), os.Getenv("GIT_INDEX_FILE"))
	}
}

// ---------------------------------------------------------------------------
// internal/hook/worktree_create.go — resolveWorktreeRepoRoot(dir)
// ---------------------------------------------------------------------------

// evalSymlinks resolves a path the way git reports one, so a comparison is not
// defeated by /var being a symlink to /private/var on darwin.
func evalSymlinks(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("EvalSymlinks(%s): %v", path, err)
	}
	return resolved
}

// resolveWorktreeRepoRoot gates every worktree-isolated agent spawn: it decides
// which checkout the new worktree is created under. Its own doc comment states
// the case that matters — "the hook may run with cwd deep inside the tree, so
// the root must be resolved rather than assumed" — and that resolution is what
// a leaked GIT_DIR destroys.
//
// The observable needs care. `git rev-parse --show-toplevel` under an explicit
// GIT_DIR does NOT report the leaked repository's checkout: with GIT_DIR set
// and GIT_WORK_TREE unset, git regards the CURRENT directory as the top level.
// Measured on git here, both the plain host/.git and the per-worktree
// host/.git/worktrees/<name> forms return the given directory verbatim.
//
// That is not immunity, it is the defect in a different shape: the answer stops
// being a resolution at all and becomes an echo of the input. So the target is
// a SUBDIRECTORY of its repository. The correct answer is the repository root;
// a leaking implementation returns the subdirectory, and the spawned worktree
// is then created against a path that is not a checkout root.
func TestResolveWorktreeRepoRootReadsGivenRepoNotLeakedGitDir(t *testing.T) {
	requireGit(t)

	base := t.TempDir()
	host := gitEnvNewRepo(t, filepath.Join(base, "host"), "host", hostDate)
	target := gitEnvNewRepo(t, filepath.Join(base, "target"), "target", targetDate)

	targetSub := filepath.Join(target, "internal", "deep")
	if err := os.MkdirAll(targetSub, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", targetSub, err)
	}

	// Fixture sanity: with a scrubbed environment the subdirectory resolves UP
	// to its repository root. Without this the main assertion could not tell
	// "the leak defeated resolution" from "resolution never worked here".
	wantRoot := evalSymlinks(t, target)
	if got := evalSymlinks(t, gitEnvRun(t, targetSub, nil, "rev-parse", "--show-toplevel")); got != wantRoot {
		t.Fatalf("fixture broken: %s resolves to %q, want %q", targetSub, got, wantRoot)
	}
	if evalSymlinks(t, targetSub) == wantRoot {
		t.Fatalf("fixture broken: subdirectory and repository root are the same path %q", wantRoot)
	}

	gitEnvLeakPreCommit(t, filepath.Join(host, ".git"))

	got, err := resolveWorktreeRepoRoot(targetSub)
	if err != nil {
		t.Fatalf("resolveWorktreeRepoRoot(%s): %v", targetSub, err)
	}
	if evalSymlinks(t, got) != wantRoot {
		t.Errorf("resolveWorktreeRepoRoot(%s) did not resolve the repository root under a leaked GIT_DIR: got %q, want %q\nGIT_DIR=%s GIT_INDEX_FILE=%s",
			targetSub, got, wantRoot, os.Getenv("GIT_DIR"), os.Getenv("GIT_INDEX_FILE"))
	}
}
