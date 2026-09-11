# F33 — Shared snapshot direct consumers

## Claim

The shared `moai verify check --key-current` snapshot is now consumed at the
three documented decision points: the Stop hook logs hit/miss/unavailable, the
4-dimension Context schema carries exact snapshot evidence to Judges, and the
fallback `sync-auditor` requires the same query before scoring.

## Evidence

Command:

```text
bash .claude/hooks/tests/test-snapshot-consumer-contract.sh
```

Observed output:

```text
PASS: hook, 4-dimension context, and fallback auditor consume one keyed snapshot
```

## Baseline-attribution

The contract test ran in `WT-workflow-audit-f33b` after fast-forwarding to the
F32 merge on local `develop`.

## Gaps

No live Stop-hook invocation was run with a fake `moai verify` binary, and no
4-dimension agent call was executed to capture a real snapshot output.

## Residual-risk

The Context/Judge runtime still depends on the agent honoring its read-only
query instruction and returning schema-valid evidence. A malformed or omitted
field will fail the typed context call rather than silently proving a snapshot
hit.
