package worktree

import "time"

// WorktreeInfo represents metadata about a single worktree
type WorktreeInfo struct {
	SpecID       string `json:"spec_id"`
	Path         string `json:"path"`
	Branch       string `json:"branch"`
	CreatedAt    string `json:"created_at"`
	LastAccessed string `json:"last_accessed"`
	Status       string `json:"status"` // "active", "recovered"
}

// NewWorktreeInfo creates a new WorktreeInfo with current timestamps
func NewWorktreeInfo(specID, path, branch string) *WorktreeInfo {
	now := time.Now().UTC().Format(time.RFC3339)
	return &WorktreeInfo{
		SpecID:       specID,
		Path:         path,
		Branch:       branch,
		CreatedAt:    now,
		LastAccessed: now,
		Status:       "active",
	}
}
