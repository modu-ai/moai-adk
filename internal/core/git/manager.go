package git

import (
	"bytes"
	"context"
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

// @MX:ANCHOR: [AUTO] Git 리포지토리 관리의 진입점입니다. 모든 Git 작업이 이 함수를 통해 시작됩니다.
// @MX:REASON: [AUTO] fan_in=15+, Git 작업의 시작점이며 시스템 전방위에서 호출됩니다
// NewRepository opens a Git repository at the given path.
// Returns ErrNotRepository if the path is not inside a Git repository.
func NewRepository(path string) (*gitManager, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path %s: %w", path, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), foundation.DefaultGitTimeout)
	defer cancel()

	// Verify the path is a git repository.
	_, err = execGit(ctx, absPath, "rev-parse", "--git-dir")
	if err != nil {
		return nil, fmt.Errorf("open repository at %s: %w", absPath, ErrNotRepository)
	}

	// Get the repository root (toplevel).
	root, err := execGit(ctx, absPath, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("get repository root: %w", err)
	}

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

// Status returns the working tree status.
func (m *gitManager) Status() (*GitStatus, error) {
	m.logger.Debug("getting repository status")

	ctx, cancel := context.WithTimeout(context.Background(), foundation.DefaultGitTimeout)
	defer cancel()

	out, err := execGit(ctx, m.root, "status", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("status: %w", err)
	}

	status := &GitStatus{}

	if out != "" {
		lines := strings.SplitSeq(out, "\n")
		for line := range lines {
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

	// Get ahead/behind count relative to upstream.
	// Errors are ignored (e.g., no upstream configured).
	aheadBehind, err := execGit(ctx, m.root, "rev-list", "--count", "--left-right", "@{upstream}...HEAD")
	if err == nil {
		parts := strings.Split(aheadBehind, "\t")
		if len(parts) == 2 {
			// Parsing errors are intentionally ignored; invalid values default to 0.
			behind, parseErr := strconv.Atoi(parts[0])
			if parseErr == nil {
				status.Behind = behind
			} else {
				m.logger.Debug("failed to parse behind count",
					"value", parts[0],
					"error", parseErr)
			}
			ahead, parseErr := strconv.Atoi(parts[1])
			if parseErr == nil {
				status.Ahead = ahead
			} else {
				m.logger.Debug("failed to parse ahead count",
					"value", parts[1],
					"error", parseErr)
			}
		} else {
			m.logger.Debug("unexpected ahead/behind format",
				"output", aheadBehind,
				"parts", len(parts))
		}
	}

	m.logger.Debug("status retrieved",
		"staged", len(status.Staged),
		"modified", len(status.Modified),
		"untracked", len(status.Untracked),
		"ahead", status.Ahead,
		"behind", status.Behind,
	)

	return status, nil
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

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if len(args) > 0 {
			return "", fmt.Errorf("git %s: %s: %w", args[0], stderrStr, err)
		}
		return "", fmt.Errorf("git: %s: %w", stderrStr, err)
	}

	return strings.TrimRight(stdout.String(), "\n\r"), nil
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
