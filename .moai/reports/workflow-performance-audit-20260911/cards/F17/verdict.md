# F17 verdict

## Claim

F17 is fixed: the sync AC source is resolved by tier. Tier S reads inline ACs
from `spec.md`; Tier M/L read `acceptance.md`; missing or empty sources block
instead of becoming a false zero-count pass.

## Evidence

The manager-docs contract now defines the resolver and evidence fields. The
fixture verifies S, M, and missing-source behavior.

Command:

```text
.claude/hooks/tests/test-tier-ac-source-contract.sh
PASS: tier S resolves inline AC and M/L require acceptance.md
```

## Baseline-attribution

The command ran in `WT-workflow-audit-f17`, based on local `develop` commit
`160659e3c` before this card's commit.

## Gaps

The current resolver is documented shell logic; no production CLI call path was
changed in this card.

## Residual-risk

An absent `tier:` still defaults to L for backward compatibility. Callers must
resolve that default before selecting the AC source and record it in evidence.
