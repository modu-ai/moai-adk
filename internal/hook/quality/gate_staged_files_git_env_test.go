package quality

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Card t560 (GH #1691, follow-up to t516) — the sibling gap.
//
// t516 closed the leak at runStep (gate.go), where a step's child inherited
// the caller's git environment. It did not close the SIBLING call in the same
// file and the same process: stagedFiles builds its own exec.Cmd, sets cmd.Dir,
// and leaves cmd.Env nil.
//
// The consequence differs from the runStep one and is quieter. stagedFiles only
// reads, so it corrupts nothing — it answers about the WRONG repository. Under
// a pre-commit hook the leaked GIT_DIR and GIT_INDEX_FILE name the repository
// being committed to, so `git diff --cached` resolves there rather than in the
// directory the gate was asked about. The gate then decides which steps to run
// from another repository's staged set: steps that should run are skipped, and
// steps that should not are run, with no error on either path.
//
// This is the one site in the sweep whose reachability from a git hook is
// established rather than assumed: the pre-commit hook shells out to
// `moai gate`, and stagedFiles runs inside that process.

// newRepoWithStaged creates a repository holding one staged (uncommitted) file
// with the given name, and returns the repository path.
func newRepoWithStaged(t *testing.T, dir, staged string) string {
	t.Helper()
	repo := newRepo(t, dir)
	if err := os.WriteFile(filepath.Join(repo, staged), []byte("x\n"), 0o644); err != nil {
		t.Fatalf("write %s: %v", staged, err)
	}
	gitIn(t, repo, "add", staged)
	return repo
}

// stagedFiles must answer about the directory it was given. A leaked GIT_DIR /
// GIT_INDEX_FILE outranks cmd.Dir, so without a scrubbed environment the answer
// comes from the caller's repository instead.
//
// The two repositories hold DIFFERENT staged filenames on purpose: an assertion
// that merely counted entries, or checked for a non-empty result, would pass
// while reading the wrong repository. The filename is what separates them.
func TestStagedFiles_ReadsGivenRepoNotLeakedGitDir(t *testing.T) {
	requireGit(t)

	base := t.TempDir()
	host := newRepoWithStaged(t, filepath.Join(base, "host"), "host-staged.txt")
	target := newRepoWithStaged(t, filepath.Join(base, "target"), "target-staged.txt")

	// Sanity: before any leak, each repository reports its own staged file.
	// Without this the test could not tell "read the wrong repo" from "read
	// nothing at all" — both would fail the main assertion identically.
	if got := gitIn(t, target, "diff", "--cached", "--name-only"); got != "target-staged.txt" {
		t.Fatalf("fixture broken: target staged set is %q, want %q", got, "target-staged.txt")
	}
	if got := gitIn(t, host, "diff", "--cached", "--name-only"); got != "host-staged.txt" {
		t.Fatalf("fixture broken: host staged set is %q, want %q", got, "host-staged.txt")
	}

	leakPreCommitGitEnv(t, host)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	got, err := stagedFiles(ctx, target)
	if err != nil {
		t.Fatalf("stagedFiles: %v", err)
	}

	joined := strings.Join(got, ",")
	if joined != "target-staged.txt" {
		t.Errorf("stagedFiles(%s) answered about the LEAKED repository: got %q, want %q\n"+
			"GIT_DIR=%s GIT_INDEX_FILE=%s",
			target, joined, "target-staged.txt",
			os.Getenv("GIT_DIR"), os.Getenv("GIT_INDEX_FILE"))
	}
}

// The same leak, with nothing staged in the given repository. This separates a
// real fix from one that merely returns something plausible: the correct answer
// here is EMPTY, and a leaking implementation returns the host's staged file —
// so a caller reading "no staged files" versus "one staged file" takes opposite
// branches. The sibling test above cannot catch this shape, because there both
// repositories are non-empty.
func TestStagedFiles_EmptyGivenRepoIsNotFilledFromLeakedGitDir(t *testing.T) {
	requireGit(t)

	base := t.TempDir()
	host := newRepoWithStaged(t, filepath.Join(base, "host"), "host-staged.txt")
	target := newRepo(t, filepath.Join(base, "target")) // committed seed, nothing staged

	if got := gitIn(t, target, "diff", "--cached", "--name-only"); got != "" {
		t.Fatalf("fixture broken: target should have nothing staged, got %q", got)
	}

	leakPreCommitGitEnv(t, host)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	got, err := stagedFiles(ctx, target)
	if err != nil {
		t.Fatalf("stagedFiles: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("stagedFiles(%s) reported %v from the LEAKED repository; the given repository has nothing staged",
			target, got)
	}
}
