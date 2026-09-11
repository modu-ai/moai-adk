# F30 — Read-only status mode mutation boundary

## Claim

`/moai sync status` now has an explicit read-only mode contract: it denies
writers and may report “no changes” only after before/after tree-key equality
and a zero writer-attempt count.

## Evidence

Command:

```text
bash .claude/hooks/tests/test-read-only-status-contract.sh
```

Observed output:

```text
PASS: sync status mode has an explicit writer deny boundary and mutation proof
```

## Baseline-attribution

The contract test ran in the isolated `WT-workflow-audit-f30b` worktree after
it was fast-forwarded to the F29 merge on local `develop`.

## Gaps

No live `/moai sync status` invocation with a write-spy toolchain was run. The
test verifies the rule and workflow text, not runtime enforcement in every
gate implementation.

## Residual-risk

Individual diagnostic tools still need to expose their read-only capability to
the orchestrator. A tool that bypasses the mode guard could mutate despite the
contract; runtime instrumentation remains the follow-up check.
