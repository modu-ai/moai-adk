# RUN

## Input

An approved SPEC, a plan, a base, a write scope, and an owner.

## Output

A patch, receipts from actual verification (the internal `run-verify` operation returns them on stdout; storing them is not implemented in 2026.11.0-rc.1), and the remaining gaps.

## Procedure

1. Preflight the project and worktree, the dirty snapshot, the owner, the epoch, and permissions.
2. Read only the sources the task needs, and choose a change strategy.
3. For TDD, confirm why RED fails before moving to GREEN and refactoring. Cover existing behavior with characterization tests. For domain changes, review the bounded context and its terminology together.
4. Change only the allowed scope with the host's native tools, and repeat actual verification.
5. When the base or the SPEC changes, mark existing receipts stale.
6. Hand the verified exact candidate to a separate AUDIT.

## Exceptions and Permissions

Never lower acceptance to hide a failure. A provider acknowledgement, an agent's statement that it finished, or a result from another worktree is not evidence of success.

## Completion Check

Bind the output to the task and the exact snapshot. Never report a verification or external effect that was not performed as complete.
