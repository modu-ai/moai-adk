# F32 — Duplicate verification command suppression

## Claim

Run and sync phases now share a verification plan keyed by tree, command,
toolchain, environment, and scope. Exact COMPLETE results are reused; writers
invalidate the key and reruns must record their reason.

## Evidence

Command:

```text
bash .claude/hooks/tests/test-verification-plan-contract.sh
```

Observed output:

```text
PASS: verification commands have one attributable owner and exact-key reuse
```

## Baseline-attribution

The contract test ran in `WT-workflow-audit-f32b` after fast-forwarding to the
F31 merge on local `develop`.

## Gaps

No live command ledger was populated across a run→sync handoff, so duplicate
process count and wall-clock savings are not measured here.

## Residual-risk

The orchestrator and tools must actually register keys and honor COMPLETE
reuse. A caller that omits environment or scope could still create a false
match; the contract makes that omission a non-reusable result.
