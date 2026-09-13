# F39 — Current team capability versus retired genealogy

## Claim

The run and orchestration rules now use one resolver that separates explicit
team requests, current runtime probes, and historical `MODE_TEAM_UNAVAILABLE`
genealogy. An environment flag alone no longer proves availability or permits
silent downgrade.

## Evidence

Command:

```text
bash .claude/hooks/tests/test-team-capability-resolver.sh
```

Observed output:

```text
PASS: team mode separates current capability probing from retired genealogy
```

## Baseline-attribution

The contract test ran in `WT-workflow-audit-f39b` after fast-forwarding to the
F38 merge on local `develop`.

## Gaps

No fresh named-teammate probe was run in this card; actual Agent Teams result
return and runtime-version capability remain unmeasured here.

## Residual-risk

The orchestrator must implement the bounded probe and persist its result. Until
then, an explicit team request should remain blocked as indeterminate rather
than being treated as available from configuration text alone.
