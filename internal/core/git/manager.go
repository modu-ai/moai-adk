package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/foundation"
)

// Compile-time interface compliance check.
var _ Repository = (*gitManager)(nil)

// gitManager implements the Repository interface using the system git binary.
type gitManager struct {
	root   string
	logger *slog.Logger
}

// @MX:ANCHOR: [AUTO] Entry point for Git repository management. All Git operations start through this function.
// @MX:REASON: [AUTO] fan_in=15+, starting point for Git operations, called system-wide
// NewRepository opens a Git repository at the given path.
// Returns ErrNotRepository if the path is not inside a Git repository.
func NewRepository(path string) (*gitManager, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path %s: %w", path, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), foundation.DefaultGitTimeout)
	defer cancel()

	// One spawn asks both questions: rev-parse prints one line per flag, in
	// flag order, so --git-dir validates that this is a repository at all and
	// --show-toplevel yields the root. git start-up dominates the cost of
	// either question, so asking them separately doubled the price of opening
	// a repository on the warm path.
	out, err := execGit(ctx, absPath, "rev-parse", "--git-dir", "--show-toplevel")
	if err != nil {
		// The combined form fails for two distinct reasons the caller's error
		// taxonomy separates: not a repository at all, versus a repository
		// with no work tree (a bare repo), where only --show-toplevel failed.
		// Ask the discriminating question here, on the cold path, where a
		// second spawn costs nothing that matters.
		if _, probeErr := execGit(ctx, absPath, "rev-parse", "--git-dir"); probeErr != nil {
			return nil, fmt.Errorf("open repository at %s: %w", absPath, ErrNotRepository)
		}
		return nil, fmt.Errorf("get repository root: %w", err)
	}

	lines := strings.Split(out, "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("get repository root: unexpected rev-parse output %q", out)
	}
	root := strings.TrimSpace(lines[len(lines)-1])

	cleanRoot := filepath.Clean(root)
	logger := slog.Default().With("module", "git")
	logger.Debug("repository opened", "root", cleanRoot)

	return &gitManager{
		root:   cleanRoot,
		logger: logger,
	}, nil
}

// CurrentBranch returns the name of the currently checked-out branch.
func (m *gitManager) CurrentBranch() (string, error) {
	m.logger.Debug("getting current branch")

	ctx, cancel := context.WithTimeout(context.Background(), foundation.DefaultGitTimeout)
	defer cancel()

	out, err := execGit(ctx, m.root, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		return "", fmt.Errorf("current branch: %w", ErrDetachedHEAD)
	}

	m.logger.Debug("current branch retrieved", "branch", out)
	return out, nil
}

// @MX:ANCHOR: [AUTO] Status is the primary working-tree inspection point used by CLI, loop controller, and statusline consumers
// @MX:REASON: fan_in=26 across 7 files, highest among Repository methods; ahead/behind parsing logic and staged/modified/untracked categorization are consumed by multiple callers — behavior changes have wide impact
// Status returns the working tree status.
func (m *gitManager) Status() (*GitStatus, error) {
	m.logger.Debug("getting repository status")

	ctx, cancel := context.WithTimeout(context.Background(), foundation.DefaultGitTimeout)
	defer cancel()

	// --branch adds a leading `## ` header carrying the branch, its upstream,
	// and the ahead/behind counts — the same three facts that otherwise cost a
	// separate `symbolic-ref` and `rev-list` spawn apiece.
	out, err := execGit(ctx, m.root, "status", "--porcelain", "--branch")
	if err != nil {
		return nil, fmt.Errorf("status: %w", err)
	}

	status := &GitStatus{}

	if out != "" {
		lines := strings.SplitSeq(out, "\n")
		for line := range lines {
			// The header, and only the header, starts at column 0 with `## `:
			// an entry line's first two columns are status codes, which are
			// never '#', so a file literally named "## x" renders as "?? ## x".
			if strings.HasPrefix(line, "## ") {
				status.Branch, status.Ahead, status.Behind = parseStatusBranchHeader(line)
				continue
			}
			if len(line) < 3 {
				continue
			}
			x := line[0]
			y := line[1]
			file := line[3:]

			// Handle renamed files: "old -> new"
			if idx := strings.Index(file, " -> "); idx >= 0 {
				file = file[idx+4:]
			}

			switch {
			case x == '?' && y == '?':
				status.Untracked = append(status.Untracked, file)
			default:
				if x != ' ' && x != '?' {
					status.Staged = append(status.Staged, file)
				}
				if y == 'M' || y == 'D' {
					status.Modified = append(status.Modified, file)
				}
			}
		}
	}

	m.logger.Debug("status retrieved",
		"branch", status.Branch,
		"staged", len(status.Staged),
		"modified", len(status.Modified),
		"untracked", len(status.Untracked),
		"ahead", status.Ahead,
		"behind", status.Behind,
	)

	return status, nil
}

// parseStatusBranchHeader reads the `## ` header line `git status --porcelain
// --branch` prints first, returning the branch name and the ahead/behind
// counts relative to the upstream.
//
// Shapes, verified against real git output rather than assumed:
//
//	## main...origin/main [ahead 1, behind 2]
//	## main...origin/main [ahead 3]
//	## main...origin/main [behind 4]
//	## main...origin/main            in sync
//	## main                          no upstream configured
//	## No commits yet on main        fresh repository
//	## HEAD (no branch)              detached HEAD
//
// Absent counts are 0, which matches what Status reported before the header
// carried them: the separate rev-list spawn failed on a branch with no
// upstream and its error was ignored, leaving both fields at zero.
func parseStatusBranchHeader(line string) (branch string, ahead, behind int) {
	rest := strings.TrimPrefix(line, "## ")

	// A repository with no commits still has a branch in HEAD, and
	// symbolic-ref reports it — keep reporting the same name here. The
	// upstream suffix, if configured, is left for the split below.
	rest = strings.TrimPrefix(rest, "No commits yet on ")

	// Divergence is a bracketed suffix; strip it before the branch/upstream
	// split so that split never sees it.
	if idx := strings.LastIndex(rest, " ["); idx >= 0 && strings.HasSuffix(rest, "]") {
		ahead, behind = parseAheadBehind(rest[idx+2 : len(rest)-1])
		rest = rest[:idx]
	}

	// `branch...upstream` when an upstream is configured, bare `branch`
	// otherwise. A branch name cannot contain "..", so the separator is
	// unambiguous.
	if idx := strings.Index(rest, "..."); idx >= 0 {
		rest = rest[:idx]
	}

	// Whatever is left holding whitespace is not a branch name: detached HEAD
	// renders as `HEAD (no branch)`, and an unrecognised shape is safer
	// reported as absent than as a branch. Callers fall back to
	// CurrentBranch(), which owns the ErrDetachedHEAD contract.
	if strings.ContainsAny(rest, " \t") {
		return "", ahead, behind
	}
	return rest, ahead, behind
}

// parseAheadBehind reads the inside of the header's bracketed divergence
// suffix — "ahead 1", "behind 2", or "ahead 1, behind 2". Unparseable fields
// are skipped and leave their count at zero, matching the prior behaviour of
// ignoring rev-list parse errors.
func parseAheadBehind(s string) (ahead, behind int) {
	for part := range strings.SplitSeq(s, ", ") {
		field, value, ok := strings.Cut(part, " ")
		if !ok {
			continue
		}
		n, err := strconv.Atoi(value)
		if err != nil {
			continue
		}
		switch field {
		case "ahead":
			ahead = n
		case "behind":
			behind = n
		}
	}
	return ahead, behind
}

// Log returns the most recent n commits from HEAD, newest first.
func (m *gitManager) Log(n int) ([]Commit, error) {
	m.logger.Debug("getting commit log", "count", n)

	ctx, cancel := context.WithTimeout(context.Background(), foundation.DefaultGitTimeout)
	defer cancel()

	// Use unit separator (\x1f) as field delimiter.
	out, err := execGit(ctx, m.root, "log",
		fmt.Sprintf("-%d", n),
		"--format=%H\x1f%an\x1f%aI\x1f%s",
	)
	if err != nil {
		return nil, fmt.Errorf("log: %w", err)
	}

	if out == "" {
		return nil, nil
	}

	var commits []Commit
	lines := strings.SplitSeq(out, "\n")
	for line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\x1f", 4)
		if len(parts) < 4 {
			continue
		}

		date, err := time.Parse(time.RFC3339, parts[2])
		if err != nil {
			date = time.Time{}
		}

		commits = append(commits, Commit{
			Hash:    parts[0],
			Author:  parts[1],
			Date:    date,
			Message: parts[3],
		})
	}

	m.logger.Debug("commit log retrieved", "count", len(commits))
	return commits, nil
}

// Diff returns the unified diff between two references.
func (m *gitManager) Diff(ref1, ref2 string) (string, error) {
	m.logger.Debug("getting diff", "ref1", ref1, "ref2", ref2)

	ctx, cancel := context.WithTimeout(context.Background(), foundation.DefaultGitTimeout)
	defer cancel()

	out, err := execGit(ctx, m.root, "diff", ref1, ref2)
	if err != nil {
		return "", fmt.Errorf("diff %s %s: %w", ref1, ref2, err)
	}

	m.logger.Debug("diff retrieved", "bytes", len(out))
	return out, nil
}

// IsClean returns true if the working tree has no uncommitted changes.
func (m *gitManager) IsClean() (bool, error) {
	m.logger.Debug("checking working tree cleanness")

	status, err := m.Status()
	if err != nil {
		return false, fmt.Errorf("is clean: %w", err)
	}

	clean := len(status.Staged) == 0 && len(status.Modified) == 0 && len(status.Untracked) == 0
	m.logger.Debug("cleanness check complete", "clean", clean)
	return clean, nil
}

// Root returns the absolute path to the repository root directory.
func (m *gitManager) Root() string {
	return m.root
}

// @MX:ANCHOR: [AUTO] execGit is the core git command executor used by all Repository methods
// @MX:REASON: [AUTO] fan_in=5, called from branch.go, conflict.go, manager.go, worktree.go, event.go
// execGit executes a git command in the given directory and returns stdout.
// It sets GIT_TERMINAL_PROMPT=0 and LC_ALL=C for consistent behavior.
func execGit(ctx context.Context, dir string, args ...string) (string, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return "", fmt.Errorf("system git lookup: %w", ErrSystemGitNotFound)
	}

	cmd := exec.CommandContext(ctx, gitPath, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"LC_ALL=C",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if runErr := cmd.Run(); runErr != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		op := firstArg(args)
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			// Non-zero git exit. Deliberately NOT %w-chained through
			// *exec.ExitError: its ExitCode() method structurally satisfies
			// the CLI's ExitCoder interface (cmd/moai/main.go and
			// internal/cli/fang.go match `ExitCode() int` with errors.As),
			// which silences the error text and passes git's raw exit code
			// through — the silent `moai worktree done` failure measured
			// 2026-08-15 (rc=128, 0 lines of stderr). CommandError keeps the
			// stderr text printable and the exit status readable as data.
			if ctxErr := ctx.Err(); ctxErr != nil {
				// Deadline kills arrive as a killed process with empty
				// stderr; name the deadline so a retry (or the raw command,
				// which has no deadline) is recognisable as the remedy.
				stderrStr = strings.TrimSpace(stderrStr + fmt.Sprintf(" (git killed: %v)", ctxErr))
			}
			return "", &CommandError{Op: op, Stderr: stderrStr, ExitStatus: exitErr.ExitCode()}
		}
		return "", fmt.Errorf("git %s: %w", op, runErr)
	}

	return strings.TrimRight(stdout.String(), "\n\r"), nil
}

// CommandError reports a git subprocess that exited non-zero. It exists so
// the failure carries git's stderr text WITHOUT exposing *exec.ExitError in
// the wrap chain (see execGit for why that matters). Callers that need the
// exit status read ExitStatus; errors.As for *exec.ExitError deliberately
// finds nothing here.
//
// @MX:NOTE: t41 — git subprocess failures stay printable and exit non-zero
// through the CLI boundary instead of passing git's raw code silently.
type CommandError struct {
	Op         string // git subcommand (first argv element)
	Stderr     string // trimmed stderr output, verbatim
	ExitStatus int
}

func (e *CommandError) Error() string {
	if e.Stderr == "" {
		return fmt.Sprintf("git %s: exited with status %d (no stderr)", e.Op, e.ExitStatus)
	}
	return fmt.Sprintf("git %s: %s", e.Op, e.Stderr)
}

// gitResult holds the result of a git subprocess whose exit code is itself a
// verdict (SPEC-WORKTREE-SQUASH-MERGE-001 REQ-WSM-007). err is non-nil ONLY for
// infrastructural failures (git binary missing, context cancellation); a
// non-zero exit code is returned in exitCode for the caller to interpret per
// the per-probe verdict table. stdout and stderr are untrimmed so the caller
// can apply the trim policy each probe requires.
type gitResult struct {
	stdout   string
	stderr   string
	exitCode int
}

// execGitExit runs git with an optional per-call environment overlay and
// returns the structured result. It is the per-call environment hook the
// synthetic-commit probe (REQ-WSM-008) uses to pin author/committer identity
// for deterministic object creation, and the exit-code-aware path the merge
// predicate uses for probes whose non-zero exits are verdicts rather than
// failures (git diff --quiet rc 1; git merge-base rc 1 on unrelated histories).
//
// The construction stays exec.CommandContext(ctx, gitPath, args...) — direct
// argv, no shell — so the no-shell invariant on the git path is preserved
// (plan.md §D). extraEnv is appended after GIT_TERMINAL_PROMPT=0 / LC_ALL=C;
// passing nil yields identical behavior to execGit apart from the structured
// return.
func execGitExit(ctx context.Context, dir string, extraEnv []string, args ...string) (gitResult, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return gitResult{}, fmt.Errorf("system git lookup: %w", ErrSystemGitNotFound)
	}

	cmd := exec.CommandContext(ctx, gitPath, args...)
	cmd.Dir = dir
	env := append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "LC_ALL=C")
	env = append(env, extraEnv...)
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if runErr := cmd.Run(); runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			// Non-zero exit: a verdict for some probes, a failure for others.
			// Surface it to the caller via exitCode without wrapping as error.
			return gitResult{
				stdout:   stdout.String(),
				stderr:   stderr.String(),
				exitCode: exitErr.ExitCode(),
			}, nil
		}
		// Infrastructural failure (context cancellation, git binary gone mid-run).
		return gitResult{}, fmt.Errorf("git %s: %w", firstArg(args), runErr)
	}
	return gitResult{
		stdout:   stdout.String(),
		stderr:   stderr.String(),
		exitCode: 0,
	}, nil
}

// firstArg returns args[0] or "git" when args is empty, for error messages.
func firstArg(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return "git"
}

// currentBranch is a package-level helper to get the current branch name.
func currentBranch(ctx context.Context, dir string) (string, error) {
	out, err := execGit(ctx, dir, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		return "", ErrDetachedHEAD
	}
	return out, nil
}

// isWorkingTreeClean is a package-level helper to check working tree cleanliness.
func isWorkingTreeClean(ctx context.Context, dir string) (bool, error) {
	out, err := execGit(ctx, dir, "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("check working tree: %w", err)
	}
	return out == "", nil
}

// branchExists checks whether a local branch exists.
func branchExists(ctx context.Context, dir, name string) bool {
	_, err := execGit(ctx, dir, "rev-parse", "--verify", "refs/heads/"+name)
	return err == nil
}
