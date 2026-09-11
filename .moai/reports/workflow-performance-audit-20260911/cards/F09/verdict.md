# F09 verdict

## Claim

F09 is fixed: plan, run, and sync execute in launcher-created worktrees, while
the shared primary checkout is never switched or reset. The branch guard now
teaches the same launcher path instead of a bare `git worktree add` example.

## Evidence

The active SPEC workflow table and phase preconditions use `moai cc -w`; the
destructive main-checkout cleanup commands were removed. The contract test
rejects those commands and verifies the launcher references.

Command:

```text
.claude/hooks/tests/test-shared-checkout-contract.sh
PASS: plan/run/sync use launcher worktrees and non-destructive integration
```

## Baseline-attribution

The command ran in `WT-workflow-audit-f09`, based on local `develop` commit
`bf6da779b` before this card's commit.

## Gaps

The test is a document contract check; it does not attempt to mutate the
shared primary checkout. The launcher itself was exercised while creating this
card worktree and returned the environment's Claude weekly-limit message after
creating the worktree.

## Residual-risk

Archived examples outside the active rule can still contain direct checkout
commands. They must not be copied into an active route without passing this
contract test.
