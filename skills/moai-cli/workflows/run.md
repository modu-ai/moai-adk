# RUN

## Input

An approved SPEC, a plan, a base, a write scope, and an owner.

## Output

A patch, receipts from actual verification (the internal `run-verify` operation returns them on stdout; storing them is not implemented in 2026.11.0-rc.1), and the remaining gaps.

## Procedure

1. Preflight the project and worktree, the dirty snapshot, and permissions; read the owner and epoch from the claim journal (`claim-show`) before writing, and claim through it when nobody holds the task.
2. Read only the sources the task needs, and choose a change strategy. Gather them with `context-bundle` under an explicit budget and re-check them with `context-verify` before citing; read a `stale` range again. Record context usage from the host's own source (Claude Code statusline `context_window`, Codex `model_context_window` or `/status`); a `bytes_div_4` figure is an estimate.
3. For TDD, confirm why RED fails before moving to GREEN and refactoring. Cover existing behavior with characterization tests. For domain changes, review the bounded context and its terminology together.
4. Change only the allowed scope with the host's native tools, and repeat actual verification.
5. When the base or the SPEC changes, mark existing receipts stale. After a stalled or interrupted attempt, do not re-send anything: read the owner and effects back (`claim-show`, `effects-reconcile`) and resume only on what the records prove.
6. Hand the verified exact candidate to a separate AUDIT.

## Exceptions and Permissions

Never lower acceptance to hide a failure. A provider acknowledgement, an agent's statement that it finished, or a result from another worktree is not evidence of success.

## Completion Check

Bind the output to the task and the exact snapshot. Never report a verification or external effect that was not performed as complete.
