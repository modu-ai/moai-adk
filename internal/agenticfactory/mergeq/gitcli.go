package mergeq

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/gitenv"
)

// gitTimeout bounds every git call of the merge queue; a hung git must not hold the landing lock.
const gitTimeout = 2 * time.Minute

// gitWaitDelay is how long to wait for git's output pipes after its context ended.
const gitWaitDelay = 5 * time.Second

// maxGitOutput bounds what one git call may return, so a runaway command cannot fill memory.
const maxGitOutput = 8 << 20

// ErrMerge marks a merge of the candidate into the trunk that git refused (a conflict); the candidate is
// not mergeable as it stands and the card goes back to be fixed.
var ErrMerge = errors.New("mergeq: the candidate does not merge into the trunk")

// ErrRefMoved marks a conditional ref update that found the ref no longer at the value it was read at.
var ErrRefMoved = errors.New("mergeq: the trunk moved since it was read")

// Identity is the author the merge commit is made with. It is explicit: the queue never reads the
// operator's global git configuration for it.
type Identity struct{ Name, Email string }

// Candidate is a merge result built in a throwaway worktree: Dir is where the gates run, Result is the
// merge commit they ran against.
type Candidate struct {
	Dir    string
	Result string
}

// Git is what the merge queue needs from git. GitCLI is the production implementation.
type Git interface {
	RevParse(ctx context.Context, ref string) (string, error)
	IsAncestor(ctx context.Context, ancestor, descendant string) (bool, error)
	// PrepareCandidate checks out trunkSHA detached in a fresh worktree and merges candidateSHA into it
	// (no fast-forward), so the gates see the real merge result. Nothing it does touches the trunk ref.
	PrepareCandidate(ctx context.Context, trunkSHA, candidateSHA, message string) (Candidate, error)
	Discard(ctx context.Context, c Candidate) error
	// CheckedOutAt returns the worktree that has ref checked out, or "" when none does.
	CheckedOutAt(ctx context.Context, ref string) (string, error)
	// UpdateRef moves ref from oldSHA to newSHA only if it still points at oldSHA (`git update-ref ref new old`).
	UpdateRef(ctx context.Context, ref, newSHA, oldSHA string) error
}

// GitCLI runs the git command line against one repository. Its environment is scrubbed of the variables
// that choose a repository, and it never prompts.
type GitCLI struct {
	Repo     string // the repository (any worktree of it)
	ScratchD string // where candidate worktrees are made; must be outside the repository's worktrees
	Who      Identity
}

func (g GitCLI) run(ctx context.Context, dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, gitTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
	cmd.WaitDelay = gitWaitDelay // a descendant that keeps the pipes open must not outlive the timeout
	cmd.Env = append(gitenv.Env(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_AUTHOR_NAME="+g.Who.Name, "GIT_AUTHOR_EMAIL="+g.Who.Email,
		"GIT_COMMITTER_NAME="+g.Who.Name, "GIT_COMMITTER_EMAIL="+g.Who.Email,
	)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &limited{b: &out}, &limited{b: &errb}
	if err := cmd.Run(); err != nil {
		return strings.TrimSpace(out.String()), fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(errb.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

type limited struct{ b *bytes.Buffer }

func (l *limited) Write(p []byte) (int, error) {
	if l.b.Len()+len(p) > maxGitOutput {
		return 0, errors.New("mergeq: git output exceeded its limit")
	}
	return l.b.Write(p)
}

// RevParse resolves ref to the commit it names.
func (g GitCLI) RevParse(ctx context.Context, ref string) (string, error) {
	return g.run(ctx, g.Repo, "rev-parse", "--verify", "--quiet", ref+"^{commit}")
}

// IsAncestor reports whether ancestor is reachable from descendant (equal commits count).
func (g GitCLI) IsAncestor(ctx context.Context, ancestor, descendant string) (bool, error) {
	_, err := g.run(ctx, g.Repo, "merge-base", "--is-ancestor", ancestor, descendant)
	if err == nil {
		return true, nil
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) && ee.ExitCode() == 1 {
		return false, nil
	}
	return false, err
}

// PrepareCandidate builds the merge of candidateSHA into trunkSHA in a throwaway worktree.
func (g GitCLI) PrepareCandidate(ctx context.Context, trunkSHA, candidateSHA, message string) (Candidate, error) {
	if g.ScratchD == "" {
		return Candidate{}, errors.New("mergeq: no scratch directory is configured for candidate worktrees")
	}
	if err := os.MkdirAll(g.ScratchD, 0o755); err != nil {
		return Candidate{}, err
	}
	dir, err := os.MkdirTemp(g.ScratchD, "cand-")
	if err != nil {
		return Candidate{}, err
	}
	if _, err := g.run(ctx, g.Repo, "worktree", "add", "--detach", dir, trunkSHA); err != nil {
		_ = os.RemoveAll(dir)
		return Candidate{}, err
	}
	c := Candidate{Dir: dir}
	if _, err := g.run(ctx, dir, "merge", "--no-ff", "--no-edit", "-m", message, candidateSHA); err != nil {
		_ = g.Discard(ctx, c)
		return Candidate{}, fmt.Errorf("%w: %v", ErrMerge, err)
	}
	res, err := g.run(ctx, dir, "rev-parse", "HEAD")
	if err != nil {
		_ = g.Discard(ctx, c)
		return Candidate{}, err
	}
	c.Result = res
	return c, nil
}

// Discard removes a candidate worktree. A tree that is already gone is not an error.
func (g GitCLI) Discard(ctx context.Context, c Candidate) error {
	if c.Dir == "" {
		return nil
	}
	// The tree is throwaway and may hold build output, so it is removed as a directory and git is told to forget
	// it; no force flag is involved (a repository-wide scan keeps force flags out of production literals).
	if err := os.RemoveAll(c.Dir); err != nil {
		return err
	}
	_, err := g.run(ctx, g.Repo, "worktree", "prune")
	return err
}

// CheckedOutAt returns the directory of a worktree that has ref checked out, or "" when none does. It reads
// `git worktree list --porcelain` for the repository, so a bare repository and every linked tree are covered.
func (g GitCLI) CheckedOutAt(ctx context.Context, ref string) (string, error) {
	out, err := g.run(ctx, g.Repo, "worktree", "list", "--porcelain")
	if err != nil {
		return "", err
	}
	var dir string
	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			dir = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "branch "):
			if strings.TrimPrefix(line, "branch ") == ref {
				return dir, nil
			}
		}
	}
	return "", nil
}

// UpdateRef is the compare-and-swap that lands a card: it succeeds only if ref still names oldSHA.
func (g GitCLI) UpdateRef(ctx context.Context, ref, newSHA, oldSHA string) error {
	if _, err := g.run(ctx, g.Repo, "update-ref", ref, newSHA, oldSHA); err != nil {
		// Only a value mismatch means the trunk moved; an invalid name, a missing object or a lock failure is a
		// different problem and must not be retried as if another lander had won.
		if strings.Contains(err.Error(), "but expected") {
			return fmt.Errorf("%w: %v", ErrRefMoved, err)
		}
		return err
	}
	return nil
}
