package hook

import (
	"context"
	"log/slog"
)

// worktreeRemoveHandler processes WorktreeRemove events.
// Fired when Claude Code removes an isolated git worktree after an agent
// with isolation: worktree terminates (v2.1.49+).
type worktreeRemoveHandler struct{}

// NewWorktreeRemoveHandler creates a new WorktreeRemove event handler.
func NewWorktreeRemoveHandler() Handler {
	return &worktreeRemoveHandler{}
}

// EventType returns EventWorktreeRemove.
func (h *worktreeRemoveHandler) EventType() EventType {
	return EventWorktreeRemove
}

// Handle processes a WorktreeRemove event. It logs the worktree removal details
// for session tracking and cleanup verification.
func (h *worktreeRemoveHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("worktree removed after isolated agent termination",
		"session_id", input.SessionID,
		"agent_id", input.AgentID,
		"agent_name", input.AgentName,
		"worktree_path", input.WorktreePath,
		"worktree_branch", input.WorktreeBranch,
	)
	return &HookOutput{}, nil
}
