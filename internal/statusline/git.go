package statusline

import (
	"context"
	"log/slog"

	gitpkg "github.com/modu-ai/moai-adk/internal/core/git"
)

// gitCollector adapts a git.Repository to the GitDataProvider interface.
type gitCollector struct {
	repo gitpkg.Repository
}

// NewGitCollector creates a GitDataProvider that collects status from a
// git.Repository. If repo is nil, CollectGitStatus returns empty data
// with Available=false (non-error behavior for non-git directories).
func NewGitCollector(repo gitpkg.Repository) GitDataProvider {
	return &gitCollector{repo: repo}
}

// CollectGitStatus retrieves the current branch and working tree status.
// Individual failures are handled gracefully: a branch error results in
// empty branch name, a status error results in zero counts. This method
// never returns an error that should stop the statusline pipeline.
func (c *gitCollector) CollectGitStatus(ctx context.Context) (*GitStatusData, error) {
	if c.repo == nil {
		return &GitStatusData{Available: false}, nil
	}

	data := &GitStatusData{Available: true}

	branch, err := c.repo.CurrentBranch()
	if err != nil {
		slog.Debug("git branch collection failed", "error", err)
		// Continue with empty branch; non-fatal
	}
	data.Branch = branch

	status, err := c.repo.Status()
	if err != nil {
		slog.Debug("git status collection failed", "error", err)
		// Return partial data with branch only
		return data, nil
	}

	data.Modified = len(status.Modified)
	data.Staged = len(status.Staged)
	data.Untracked = len(status.Untracked)
	data.Ahead = status.Ahead
	data.Behind = status.Behind

	return data, nil
}
