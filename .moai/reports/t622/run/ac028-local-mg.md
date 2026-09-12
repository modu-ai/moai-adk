## PR Auto-Merge

Execute only with the `--auto-merge` flag (`--merge` is a deprecated alias of `--auto-merge`); without it the PR is not merged. Mode conditions:
- In team mode, `--auto-merge` merges only after all approvals are obtained.
- In personal and manual modes, `--auto-merge` merges without an approval condition (no teammates to approve).

Steps (all modes; CI checks must pass and the PR must have no merge conflicts):
1. Push to remote
2. `gh pr ready`
3. `gh pr checks --watch`
4. `gh pr merge --<merge_method> --delete-branch` using the resolved merge_method
5. Checkout main, pull, delete local branch

