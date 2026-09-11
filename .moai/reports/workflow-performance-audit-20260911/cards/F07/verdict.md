# F07 verdict

## Claim

F07 is fixed: sync documentation now classifies SPEC divergence and hands body
changes to `manager-spec`. `manager-docs` owns only approved frontmatter
transitions and sync-facing documents, consistent with its agent contract.

## Evidence

The workflow text has an explicit body-handoff step and the ownership test
checks both sides of the boundary.

Command:

```text
.claude/hooks/tests/test-spec-ownership-contract.sh
PASS: SPEC body ownership routes divergence to manager-spec
```

## Baseline-attribution

The command ran in `WT-workflow-audit-f07`, based on local `develop` commit
`a1fae2935` before this card's commit.

## Gaps

The test checks routing text, not a live multi-agent delegation. The orchestrator
still must consume the structured blocker and re-delegate it.

## Residual-risk

An unreviewed workflow paragraph could reintroduce the old owner. Keep this
contract test in the workflow test set and run it whenever sync ownership text
changes.
