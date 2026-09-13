## PR Auto-Merge (Team Mode)

Execute only with `--auto-merge` flag AND all approvals obtained:
1. Push to remote
2. `gh pr ready`
3. `gh pr checks --watch`
4. `gh pr merge --<merge_method> --delete-branch` using the resolved merge_method
5. Checkout main, pull, delete local branch

