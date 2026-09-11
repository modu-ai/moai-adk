# F35 — Mutation-aware graceful exit

## Claim

Sync aborts now require before/after tree-key evidence. “No changes” is
reserved for an unchanged tree; a writer or unverified rollback produces a
partial-change report with applied/unapplied operations and retry information.

## Evidence

Command:

```text
bash .claude/hooks/tests/test-graceful-exit-mutation-contract.sh
```

Observed output:

```text
PASS: graceful exits distinguish no mutation from partial or rolled-back changes
```

## Baseline-attribution

The contract test ran in `WT-workflow-audit-f35b` after fast-forwarding to the
F34 merge on local `develop`.

## Gaps

No injected writer failure or rollback fault was executed, so runtime path-list
accuracy and rollback restoration are not measured here.

## Residual-risk

The orchestrator must populate the keys and mutation list on every abort path;
the prose contract alone cannot detect a caller that exits before taking the
post-mutation snapshot.
