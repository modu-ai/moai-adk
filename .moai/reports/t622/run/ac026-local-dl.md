#### Step 3.4: Auto-Merge Behavior

Only applies when a PR was created in Step 3.2.

##### Auto-Merge Trigger Conditions

Merging is opt-in. The single criterion is the `--auto-merge` opt-in defined in `manager-git.md` § PR Auto-Merge; worktree context alone never triggers a merge.

Auto-merge trigger conditions:
- `--auto-merge` flag set
- OR `--merge` flag set (deprecated alias of `--auto-merge`, logged as warning)

Mode conditions (same as `manager-git.md` § PR Auto-Merge):
- In team mode, `--auto-merge` merges only after all approvals are obtained.
- In personal and manual modes, `--auto-merge` merges without an approval condition (no teammates to approve).

When auto-merge is triggered:
1. Verify all CI/CD checks pass (gh pr checks)
2. Verify zero merge conflicts (gh pr view --json mergeable)
3. If all checks pass: Execute `gh pr merge --squash --delete-branch`
4. If checks fail: Report error with recovery command, do NOT merge

##### Flag Behavior

- `--auto-merge`: Opt in to merging the PR after sync, under the mode conditions above.
- `--merge`: deprecated alias of `--auto-merge` (logs a warning).
- `--no-merge`: Deprecated no-op kept for compatibility (logs a warning); not merging is already the default.

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

