# VERIFY

## Input

An exact tree, the SPEC and policy digests, and a check inventory.

## Output

A receipt (planned; not implemented in 2026.11.0-rc.1) with each check marked passed, failed, not_run, blocked, stale, or unknown.

## Procedure

1. Confirm the verification commands and environment the repository actually uses.
2. Before running anything, confirm permissions and the required guards.
3. Run the commands through a trusted runner, and record start, end, exit code, cancellation, and timeout.
4. Bind the task, attempt, tree, SPEC, policy, and tool version to the raw evidence.
5. State skips, coverage gaps, and artifact expiry explicitly.

## Exceptions and Permissions

A command that was not run is not_run. A missing required guard is blocked before anything runs, and exit code 0 alone does not satisfy every AC.

## Completion Check

Bind the output to the task and the exact snapshot. Never report a verification or external effect that was not performed as complete.
