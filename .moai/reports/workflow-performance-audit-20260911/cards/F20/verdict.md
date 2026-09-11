# F20 verdict

## Claim

F20 is fixed: phase jumps use canonical stable IDs with an alias map for legacy
decimal labels; skip/resume validates target existence and DAG cycles before
execution.

## Evidence

Added the stable phase-ID contract and referenced it at run entry. The contract
test verifies canonical IDs, alias handling, and DAG validation language.

Command:

```text
.claude/hooks/tests/test-phase-id-contract.sh
PASS: skip/resume is defined over canonical phase IDs
```

## Baseline-attribution

The command ran in `WT-workflow-audit-f20`, based on local `develop` commit
`c195e774a` before this card's commit.

## Gaps

This card adds the normative contract but not a persisted runtime phase ledger
implementation.

## Residual-risk

Existing state records with only numeric labels need an explicit migration map;
otherwise resume must fail closed rather than guessing.
