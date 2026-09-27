package gitio

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// newRepo creates a throwaway git repository with two commits on one branch.
func newRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	run("init", "-b", "main")
	run("config", "user.name", "t")
	run("config", "user.email", "t@example.com")
	write(t, dir, "a.txt", "a\n")
	run("add", "a.txt")
	run("commit", "-m", "one")
	write(t, dir, "b.txt", "b\n")
	run("add", "b.txt")
	run("commit", "-m", "two")
	return dir
}

func environ() []string {
	var out []string
	for _, e := range os.Environ() {
		switch {
		case strings.HasPrefix(e, "GIT_DIR="),
			strings.HasPrefix(e, "GIT_WORK_TREE="),
			strings.HasPrefix(e, "GIT_INDEX_FILE="):
			continue
		}
		out = append(out, e)
	}
	return out
}

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// TestHeadAndAncestry pins Head, ResolveRef, and IsAncestor over a real repo.
func TestHeadAndAncestry(t *testing.T) {
	dir := newRepo(t)
	head, err := Head(dir)
	if err != nil {
		t.Fatalf("head: %v", err)
	}
	if len(head) != 40 {
		t.Fatalf("head = %q", head)
	}
	one, err := ResolveRef(dir, "HEAD~1")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	anc, err := IsAncestor(dir, one, head)
	if err != nil || !anc {
		t.Fatalf("IsAncestor(one, head) = %v %v, want true", anc, err)
	}
	anc, err = IsAncestor(dir, head, one)
	if err != nil || anc {
		t.Fatalf("IsAncestor(head, one) = %v %v, want false", anc, err)
	}
	anc, err = IsAncestor(dir, head, head)
	if err != nil || !anc {
		t.Fatalf("IsAncestor(head, head) = %v %v, want true", anc, err)
	}
}

// TestNonMergeCommits pins the range listing shape: oldest first, paths
// attached, empty range empty.
func TestNonMergeCommits(t *testing.T) {
	dir := newRepo(t)
	one, _ := ResolveRef(dir, "HEAD~1")
	head, _ := Head(dir)
	commits, err := NonMergeCommits(dir, one, head)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(commits) != 1 || len(commits[0].Paths) != 1 || commits[0].Paths[0] != "b.txt" {
		t.Fatalf("commits = %+v", commits)
	}
	empty, err := NonMergeCommits(dir, head, head)
	if err != nil || len(empty) != 0 {
		t.Fatalf("empty range = %v %v", empty, err)
	}
}

// TestWorktreeListAndCommonDir pins the linked-worktree listing and the
// common dir (the primary checkout resolution of §C.3).
func TestWorktreeListAndCommonDir(t *testing.T) {
	dir := newRepo(t)
	link := t.TempDir() + "/w1"
	cmd := exec.Command("git", "worktree", "add", link, "-b", "wt-one")
	cmd.Dir = dir
	cmd.Env = environ()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("worktree add: %v: %s", err, out)
	}
	wts, err := WorktreeList(dir)
	if err != nil {
		t.Fatalf("worktree list: %v", err)
	}
	if len(wts) != 2 {
		t.Fatalf("worktrees = %+v, want 2", wts)
	}
	common, err := CommonDir(dir)
	if err != nil {
		t.Fatalf("common dir: %v", err)
	}
	// The primary checkout is the parent of the common dir (§C.3): for a
	// plain repo that is the repo root itself.
	if got := filepath.Dir(common); got != dir {
		t.Fatalf("primary checkout = %q, want the repo %q (common dir %q)", got, dir, common)
	}
}

// TestBlobAndPathsUnder pins the tree-read helpers.
func TestBlobAndPathsUnder(t *testing.T) {
	dir := newRepo(t)
	head, err := Head(dir)
	if err != nil {
		t.Fatalf("head: %v", err)
	}
	data, err := Blob(dir, head, "a.txt")
	if err != nil || string(data) != "a\n" {
		t.Fatalf("blob = %q err %v", data, err)
	}
	if _, err := Blob(dir, head, "missing.txt"); err == nil {
		t.Fatalf("missing blob succeeded")
	}
	paths, err := PathsUnder(dir, head, "")
	if err != nil || len(paths) != 2 {
		t.Fatalf("paths = %v err %v, want two files", paths, err)
	}
	if _, err := PathsUnder(dir, head, "no-such-dir"); err != nil {
		t.Fatalf("paths under empty dir: %v", err)
	}
}

// TestCurrentAndUpstreamBranch pins the branch readers (upstream empty when
// unconfigured).
func TestCurrentAndUpstreamBranch(t *testing.T) {
	dir := newRepo(t)
	branch, err := CurrentBranch(dir)
	if err != nil || branch != "main" {
		t.Fatalf("branch = %q err %v, want main", branch, err)
	}
	up, err := UpstreamBranch(dir, "main")
	if err != nil || up != "" {
		t.Fatalf("upstream = %q err %v, want empty", up, err)
	}
}

// TestDiffAndChangedFileCount pins the diff helpers.
func TestDiffAndChangedFileCount(t *testing.T) {
	dir := newRepo(t)
	one, _ := ResolveRef(dir, "HEAD~1")
	head, _ := Head(dir)
	diff, err := Diff(dir, one, head)
	if err != nil || !strings.Contains(diff, "b.txt") {
		t.Fatalf("diff = %q err %v", diff, err)
	}
	n, err := ChangedFileCount(dir, one, head)
	if err != nil || n != 1 {
		t.Fatalf("changed files = %d err %v, want 1", n, err)
	}
	if n, _ := ChangedFileCount(dir, head, head); n != 0 {
		t.Fatalf("identical revs changed = %d, want 0", n)
	}
	if _, err := Run(dir, "rev-parse", "--verify", "definitely-not-a-ref"); err == nil {
		t.Fatalf("bad rev succeeded")
	}
	if !isSHA(head) || isSHA("not-a-sha") {
		t.Fatalf("isSHA discrimination")
	}
}

// TestTimeoutIsError pins the bounded-git-work property: a timeout surfaces
// as an error, which the readiness evaluator maps to push_check_undetermined.
func TestTimeoutIsError(t *testing.T) {
	saved := timeout()
	defer func() { timeout = func() time.Duration { return saved } }()
	timeout = func() time.Duration { return time.Nanosecond }
	_, err := Run(t.TempDir(), "status")
	if err == nil {
		t.Fatalf("expected a timeout error, got nil")
	}
	var cmdErr *exec.ExitError
	if errors.As(err, &cmdErr) {
		// a killed command is acceptable; a nil error is not
		t.Logf("timeout surfaced as exec error: %v", err)
	}
}
