# F38 — Clear versus warm context policy

## Claim

Run now chooses `warm` or `clear` using plan/tree identity, context size,
prefix changes, audit independence, and model/effort switches. Both paths
carry a handoff ledger; a bare `/clear` no longer implies continuity.

## Evidence

Command:

```text
bash .claude/hooks/tests/test-context-clear-policy.sh
```

Observed output:

```text
PASS: clear versus warm context decisions require an identity-bearing handoff
```

## Baseline-attribution

The contract test ran in `WT-workflow-audit-f38b` after fast-forwarding to the
F37 merge on local `develop`.

## Gaps

No real context-window measurement or cache hit-rate comparison was available
in this run; the test validates the decision contract and handoff fields.

## Residual-risk

The runtime orchestrator must calculate context size/prefix validity and write
the ledger before clearing. If it records an incorrect identity, a warm handoff
could still reuse stale instructions.
