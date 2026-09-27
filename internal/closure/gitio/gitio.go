// Package gitio is the only subprocess user of internal/closure: git
// commands behind plain functions with a bounded timeout (spec.md §E: git
// work is bounded by a timeout, and a timeout is an error, which the push
// evaluator maps to push_check_undetermined). It never reads MoAI
// configuration; callers pass the directory every command runs in.
package gitio

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// DefaultTimeout bounds every git call. A hook invocation carries a 10s
// budget; the guard reserves a slice of it for the whole evaluation.
const DefaultTimeout = 5 * time.Second

// CommitPaths is one non-merge commit and the repo-relative paths it changed.
type CommitPaths struct {
	SHA   string
	Paths []string
}

// Worktree is one entry of `git worktree list --porcelain`.
type Worktree struct {
	Path   string
	Head   string
	Branch string // refs/heads/..., empty for a detached worktree
	Bare   bool
}

// timeout returns the effective per-call timeout.
var timeout = func() time.Duration { return DefaultTimeout }

// Run executes one git command in dir and returns its trimmed stdout. Any
// non-zero exit, missing binary, or timeout is an error naming the command.
func Run(dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout())
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gitio: git %s in %s: %w: %s",
			strings.Join(args, " "), dir, err, strings.TrimSpace(errb.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

// Head returns the commit HEAD points at.
func Head(dir string) (string, error) {
	return Run(dir, "rev-parse", "HEAD")
}

// ResolveRef resolves any rev to its commit SHA.
func ResolveRef(dir, ref string) (string, error) {
	return Run(dir, "rev-parse", "--verify", ref+"^{commit}")
}

// IsAncestor reports whether a is b or an ancestor of b.
func IsAncestor(dir, a, b string) (bool, error) {
	if a == b {
		return true, nil
	}
	_, err := Run(dir, "merge-base", "--is-ancestor", a, b)
	if err == nil {
		return true, nil
	}
	// merge-base --is-ancestor exits 1 for "not an ancestor"; anything else
	// is a real error. Run wraps the exec error with %w, so errors.As sees
	// through the wrapper.
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}
	return false, err
}

// NonMergeCommits lists the non-merge commits in from..to (from exclusive,
// to inclusive), oldest first, with each commit's changed paths.
func NonMergeCommits(dir, from, to string) ([]CommitPaths, error) {
	out, err := Run(dir, "log", "--no-merges", "--format=%H", "--name-only", from+".."+to)
	if err != nil {
		return nil, err
	}
	var commits []CommitPaths
	var cur *CommitPaths
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if isSHA(line) {
			commits = append(commits, CommitPaths{SHA: line})
			cur = &commits[len(commits)-1]
			continue
		}
		if cur != nil {
			cur.Paths = append(cur.Paths, filepath.ToSlash(line))
		}
	}
	return commits, nil
}

func isSHA(s string) bool {
	if len(s) < 7 || len(s) > 40 {
		return false
	}
	for _, r := range s {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

// WorktreeList returns the linked worktrees of the repository dir belongs to.
func WorktreeList(dir string) ([]Worktree, error) {
	out, err := Run(dir, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	var list []Worktree
	var cur *Worktree
	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			list = append(list, Worktree{Path: strings.TrimPrefix(line, "worktree ")})
			cur = &list[len(list)-1]
		case strings.HasPrefix(line, "HEAD "):
			if cur != nil {
				cur.Head = strings.TrimPrefix(line, "HEAD ")
			}
		case strings.HasPrefix(line, "branch "):
			if cur != nil {
				cur.Branch = strings.TrimPrefix(line, "branch ")
			}
		case line == "bare":
			if cur != nil {
				cur.Bare = true
			}
		case line == "detached":
			if cur != nil {
				cur.Branch = ""
			}
		}
	}
	return list, nil
}

// CommonDir returns the repository's common directory (the primary
// checkout's .git directory in a linked-worktree layout).
func CommonDir(dir string) (string, error) {
	out, err := Run(dir, "rev-parse", "--git-common-dir")
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(out) {
		out = filepath.Join(dir, out)
	}
	return filepath.Clean(out), nil
}

// CurrentBranch returns the checked-out branch of dir.
func CurrentBranch(dir string) (string, error) {
	out, err := Run(dir, "branch", "--show-current")
	if err != nil {
		return "", err
	}
	return out, nil
}

// UpstreamBranch returns the upstream branch of the named local branch in
// <remote>/<branch> form, or "" when none is configured.
func UpstreamBranch(dir, branch string) (string, error) {
	out, err := Run(dir, "rev-parse", "--abbrev-ref", branch+"@{upstream}")
	if err != nil {
		// No upstream configured is the common case; treat it as empty.
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", nil
		}
		return "", err
	}
	return out, nil
}

// ChangedFileCount returns the number of files changed between two revs.
func ChangedFileCount(dir, from, to string) (int, error) {
	out, err := Run(dir, "diff", "--name-only", from, to)
	if err != nil {
		return 0, err
	}
	if out == "" {
		return 0, nil
	}
	return len(strings.Split(out, "\n")), nil
}

// DiffSHA returns the SHA-256 of `git diff <from> <to>` output — no, of its
// bytes hashed by the caller; this helper returns the diff itself.
func Diff(dir, from, to string) (string, error) {
	return Run(dir, "diff", from, to)
}

// Blob returns one file's bytes at a rev (`git show <rev>:<path>`). A path
// absent from the rev is an error the caller maps to "absent".
func Blob(dir, rev, path string) ([]byte, error) {
	out, err := RunRaw(dir, "show", rev+":"+path)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PathsUnder lists the repo-relative file paths under prefix (a directory)
// at a rev.
func PathsUnder(dir, rev, prefix string) ([]string, error) {
	args := []string{"ls-tree", "-r", "--name-only", rev}
	if prefix != "" {
		args = append(args, "--", prefix)
	}
	out, err := Run(dir, args...)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

// RunRaw executes one git command and returns its raw stdout bytes.
func RunRaw(dir string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout())
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("gitio: git %s in %s: %w: %s",
			strings.Join(args, " "), dir, err, strings.TrimSpace(errb.String()))
	}
	return out.Bytes(), nil
}
