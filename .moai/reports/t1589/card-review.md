# Card review — t1589

Scope: `internal/cli/factory_card.go`, `internal/cli/factory_card_pr.go`, new `internal/cli/factory_card_pr_gitenv_test.go`.

## What changed

- `factoryGitRead` (factory_card_pr.go): `cmd.Env = gitenv.Env()` added. It runs `git -C <dir>`, and the inherited `GIT_DIR` outranked `-C`.
- `factoryBranchOfWorktree` (factory_card.go): `cmd.Env = gitenv.Env()` added. It uses `cmd.Dir`, which `GIT_DIR` also outranks.
- `factoryWorktreeForBranchIn` (factory_card.go): `cmd.Env = gitenv.Env()` added. It lists the worktree of the integration branch; `integration_merge.go:146` and `local_main_resync.go:313` depend on it.
- Imports: `internal/gitenv` added to both files.

## Review lenses

- **Correctness:** the three functions read the repository of their argument. The fix matches the in-repo pattern (`gitenv.Env()` with the comment "an inherited GIT_DIR would outrank -C"). `gitenv.Env()` only removes `RepoScopingVars`; `PATH`, `HOME`, and the rest pass through, so `gh` and `git` still resolve normally.
- **Scope:** three one-line changes plus an import each. No behavior change for callers whose environment has no `GIT_DIR`.
- **Tests:** the new test fails before the fix on all three assertions and passes after, observed (see verdict.md). The 40-test set covering the delivery edge and the integration merge passes.
- **Ordering:** the RED test is its own commit (6b12d5b830) ahead of the fix (4faf5ff818), so the commit graph witnesses the order (verification-claim-integrity §2.3).

## Concerns (not blocking this card)

- The card body's second defect (the strategy branch, git-flow vs github-flow) is out of this change's scope and remains open.
- Two raw `git` call sites outside the delivery edge (factory_card.go:633, :2272) keep the pre-fix shape. Recorded in verdict.md Gaps.
