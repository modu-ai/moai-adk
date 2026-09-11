## Synchronization

Pre-flight status reads are read-only, but `git rev-list --count --left-right` reads the remote-tracking refs that `git fetch` updates, so the two are not independent: run `git fetch` first and wait until it completes, then run `git rev-list --count --left-right` — never in the same batch as the fetch. Reads that do not consume the fetch result (`git status`, `gh pr checks --json`) may run in parallel with the fetch as one single-turn multi-Bash batch per `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution (grouping rationale and batch-safety taxonomy: `.claude/rules/moai/workflow/verification-batch-pattern.md`).

- Checkpoint before remote operations
- Verify branch and check uncommitted changes
- `git fetch origin` → `git pull origin [branch]`
- Conflict detection with resolution guidance
- Feature branch rebase on latest main after PR merges

## PR Auto-Merge
