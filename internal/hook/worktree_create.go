package hook

import (
	"context"
	"log/slog"
)

// worktreeCreateHandler processes WorktreeCreate events.
// Fired when Claude Code creates an isolated git worktree for an agent
// with isolation: worktree in its frontmatter (v2.1.49+).
type worktreeCreateHandler struct{}

// NewWorktreeCreateHandler creates a new WorktreeCreate event handler.
func NewWorktreeCreateHandler() Handler {
	return &worktreeCreateHandler{}
}

// EventType returns EventWorktreeCreate.
func (h *worktreeCreateHandler) EventType() EventType {
	return EventWorktreeCreate
}

// Handle processes a WorktreeCreate event. It logs the worktree creation details
// for session tracking and debugging.
func (h *worktreeCreateHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("worktree created for isolated agent",
		"session_id", input.SessionID,
		"agent_id", input.AgentID,
		"agent_name", input.AgentName,
		"worktree_path", input.WorktreePath,
		"worktree_branch", input.WorktreeBranch,
	)
	return &HookOutput{}, nil
}
