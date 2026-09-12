#### Step 3.4: Auto-Merge Behavior

Only applies when a PR was created in Step 3.2.

##### Auto-Merge Trigger Conditions

Auto-merge trigger conditions:
- `is_worktree_context == true` AND `--no-merge` flag NOT set
- OR `--merge` flag explicitly set (deprecated, logged as warning)

When auto-merge is triggered:
1. Verify all CI/CD checks pass (gh pr checks)
2. Verify zero merge conflicts (gh pr view --json mergeable)
3. If all checks pass: Execute `gh pr merge --squash --delete-branch`
4. If checks fail: Report error with recovery command, do NOT merge

##### Flag Behavior

- `--no-merge`: Skip auto-merge even in worktree context. PR is created but not merged.
- `--merge`: Deprecated. Logs warning: "The --merge flag is deprecated. Auto-merge is now the default for worktree contexts."

##### Auto-Merge Execution

1. Check CI/CD status via `gh pr checks --watch` (wait for completion)
2. Check merge conflicts via `gh pr view --json mergeable`
3. If passing and mergeable: Execute `gh pr merge --squash --delete-branch`
4. Checkout target branch, fetch latest
5. Verify local is synchronized with remote

##### Auto-Merge Failures

- If CI/CD fails: Report failure, display error details, do NOT merge
- If merge conflicts: Report conflicts, provide manual resolution guidance, do NOT merge
- If approvals missing (Team mode): Report pending approvals, do NOT merge

##### Post-Merge Automatic Cleanup

Condition: Auto-merge succeeded AND `workflow.worktree.auto_cleanup == true`

Steps:
1. Detect worktree path for current SPEC-ID from registry
2. Execute cleanup equivalent to `moai worktree done SPEC-{ID} --auto --delete-branch`:
   - Remove worktree directory
   - Remove feature branch (already deleted by --delete-branch in merge)
   - Update worktree registry
3. Log cleanup result

Error handling:
- Cleanup failure does NOT block or affect merge result
- On failure: Log warning with manual cleanup command
- Message: "Worktree cleanup warning: {error}. Manual: `moai worktree done SPEC-{ID}`"

