# AUDIT

## Input

The approved contract, the exact diff, runner receipts (planned; not implemented in 2026.11.0-rc.1), and findings.

## Output

Findings from an independent context and an evidence-bound verdict.

## Procedure

1. Give only the necessary review package to a host-native context kept separate from the implementation.
2. Cross-check sources, digests, test traces, unrun checks, and the allowed scope directly.
3. Record functional, security, consistency, and maintainability findings, with the conditions for fixing each.
4. Have a deterministic admission step confirm the must-pass conditions and their evidence.
5. If the reviewer changes code, verify the result again as a new candidate.

## Exceptions and Permissions

Self-assessment in the same context and a judge's recommendation do not replace the required independent audit. If a reviewer cannot be created or the evidence is insufficient, the result is blocked or unknown.

## Completion Check

Bind the output to the task and the exact snapshot. Never report a verification or external effect that was not performed as complete.
