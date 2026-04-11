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

// @MX:ANCHOR: [AUTO] Hook event type selector for the dispatcher routing table
// @MX:REASON: fan_in=40, every hook handler exposes this method; the dispatcher calls it on all registered handlers to route events
// EventType returns EventWorktreeRemove.
func (h *worktreeRemoveHandler) EventType() EventType {
	return EventWorktreeRemove
}

// @MX:ANCHOR: [AUTO] Worktree lifecycle hook handler called by the central hook dispatcher
// @MX:REASON: fan_in=51, the dispatcher invokes Handle on all registered handlers; signature changes affect every lifecycle handler
// Handle processes a WorktreeRemove event. It logs the worktree removal details
// and removes the entry from the worktree registry.
func (h *worktreeRemoveHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("worktree removed after isolated agent termination",
		"session_id", input.SessionID,
		"agent_id", input.AgentID,
		"agent_name", input.AgentName,
		"worktree_path", input.WorktreePath,
		"worktree_branch", input.WorktreeBranch,
	)

	// Remove the registry entry for the cleaned-up worktree.
	if input.CWD != "" && input.WorktreePath != "" {
		unregisterWorktree(input.CWD, input.WorktreePath)
	}

	return &HookOutput{}, nil
}
