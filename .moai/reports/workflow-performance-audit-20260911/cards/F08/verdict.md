# F08 verdict

## Claim

F08 is fixed: Tier no longer grants a phase agent a direct push. The delivery
policy makes `manager-git` the owner of every push, with PR as the normal route
and an explicitly configured `WT-*` local integration route as the only no-PR
exception.

## Evidence

The canonical delivery rule, spec route text, manager-git frontmatter, and sync
delivery instructions were aligned. The static contract test rejects the old
Tier S/M direct-push wording.

Command:

```text
.claude/hooks/tests/test-delivery-policy-contract.sh
PASS: all tiers share manager-git delivery ownership
```

## Baseline-attribution

The command ran in `WT-workflow-audit-f08`, based on local `develop` commit
`be4b25b60` before this card's commit.

## Gaps

Remote branch-protection settings were not queried in this local card. The
policy intentionally fails closed when route metadata is absent or conflicting.

## Residual-risk

Older archived workflow examples may still mention direct push. They are not
active route owners, but should be migrated if they become executable again.
