# F19 verdict

## Claim

F19 is fixed: plan-auditor retry limits now come from the tier ceiling map, and
one orchestrator-owned `remaining_attempts` budget covers ordinary review and
cross-validation calls.

## Evidence

The workflow no longer repeats a fixed three-iteration directive for Standard
and Thorough modes. The contract test checks the S/M/L map and rejects the old
override.

Command:

```text
.claude/hooks/tests/test-plan-audit-ceiling-contract.sh
PASS: one tier-resolved audit ceiling owns all review attempts
```

## Baseline-attribution

The command ran in `WT-workflow-audit-f19`, based on local `develop` commit
`7bacbd381` before this card's commit.

## Gaps

No runtime scheduler was invoked; the test verifies the SSOT and workflow
references.

## Residual-risk

An implementation must decrement the shared budget on timeout and content
re-review alike; prose alone cannot observe a leaked attempt.
