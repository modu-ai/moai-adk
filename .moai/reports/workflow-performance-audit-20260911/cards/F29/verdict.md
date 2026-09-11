# F29 — Pre-spawn fetch/rev-list ordering

## Claim

The pre-spawn sync contract now makes the `fetch origin/main` → `rev-list`
dependency explicit and preserves session discovery as an independent lane.

## Evidence

Command:

```text
bash .claude/hooks/tests/test-pre-spawn-fetch-order.sh
```

Observed output:

```text
PASS: pre-spawn fetch completion gates the divergence read while session discovery remains independent
```

The contract also records the fetch exit status and requires the fetched-ref
baseline to be attributable beside the divergence output.

## Baseline-attribution

The test ran in the isolated `WT-workflow-audit-f29b` worktree after it was
fast-forwarded to local `develop` at the F28 merge baseline.

## Gaps

No delayed-network fixture was run; the verification is a contract test of the
ordering and failure boundary, not a live fetch race measurement.

## Residual-risk

The launcher/orchestrator must still implement the two lanes with an actual
join. This rule prevents the known stale-ref ordering, but cannot prove every
caller obeys it unless the caller is exercised.
