##### Worktree Context Detection

Detect if the current session is running within a MoAI worktree:
- Check if current git directory path contains `/.moai/worktrees/` component
- OR check if `.moai/worktrees/registry.json` has an active entry for current SPEC-ID
- Store result as `is_worktree_context` boolean for use in Phase 13

This affects auto-merge behavior: worktree contexts default to auto-merge.

