## Synchronization

Pre-flight status reads (`git fetch`, `git status`, `git rev-list --count --left-right`, `gh pr checks --json`) are independent and read-only: issue them as ONE single-turn multi-Bash batch per `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution (grouping rationale and batch-safety taxonomy: `.claude/rules/moai/workflow/verification-batch-pattern.md`).

- Checkpoint before remote operations
- Verify branch and check uncommitted changes
- `git fetch origin` → `git pull origin [branch]`
- Conflict detection with resolution guidance
- Feature branch rebase on latest main after PR merges

## PR Auto-Merge (Team Mode)
