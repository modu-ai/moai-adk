package status

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// GitInfo represents git repository information
type GitInfo struct {
	Branch     string
	IsClean    bool
	HasChanges bool
}

// GetGitInfo retrieves git information for the current directory
func GetGitInfo() (*GitInfo, error) {
	// Check if we're in a git repository
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("not a git repository")
	}

	if strings.TrimSpace(string(output)) != "true" {
		return nil, fmt.Errorf("not a git repository")
	}

	info := &GitInfo{}

	// Get current branch
	cmd = exec.Command("git", "branch", "--show-current")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error getting git branch: %w", err)
	}
	info.Branch = strings.TrimSpace(string(output))

	// Check if working directory is clean
	cmd = exec.Command("git", "status", "--porcelain")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error checking git status: %w", err)
	}

	info.IsClean = len(strings.TrimSpace(string(output))) == 0
	info.HasChanges = !info.IsClean

	return info, nil
}

// GetRecentCommits retrieves the last N commits
func GetRecentCommits(count int) ([]CommitInfo, error) {
	cmd := exec.Command("git", "log", fmt.Sprintf("-%d", count), "--pretty=format:%h|%s|%ar", "--no-merges")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error getting git log: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	commits := make([]CommitInfo, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		if len(parts) == 3 {
			commits = append(commits, CommitInfo{
				Hash:    parts[0],
				Message: parts[1],
				TimeAgo: parts[2],
			})
		}
	}

	return commits, nil
}

// CommitInfo represents a git commit
type CommitInfo struct {
	Hash    string
	Message string
	TimeAgo string
}

// FormatTimeAgo formats a time duration in a human-readable way
func FormatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	hours := int(duration.Hours())
	days := hours / 24

	if days > 0 {
		return fmt.Sprintf("%dd ago", days)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh ago", hours)
	}
	minutes := int(duration.Minutes())
	if minutes > 0 {
		return fmt.Sprintf("%dm ago", minutes)
	}
	return "just now"
}
