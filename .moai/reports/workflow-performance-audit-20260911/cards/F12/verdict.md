# F12 verdict

## Claim

F12 is fixed: D7 discovery now reports retired/superseded references as review
candidates and explicitly leaves the blocking decision to the auditor after
reading the surrounding reconciliation text. A keyword miss is not an
automatic BLOCKING verdict.

## Evidence

The fixture creates a retired SPEC reference with an explicit reconciliation,
runs the documented discovery shape, and asserts `REVIEW` plus a candidate but
no `BLOCKING` output.

Command:

```text
.claude/hooks/tests/test-plan-auditor-d7-contract.sh
PASS: D7 emits review candidates and leaves blocking judgment to the auditor
```

## Baseline-attribution

The command ran in `WT-workflow-audit-f12`, based on local `develop` commit
`6b61784ea` before this card's commit.

## Gaps

The fixture does not invoke an LLM auditor; it verifies the mechanical
discovery boundary and the documented decision handoff.

## Residual-risk

The auditor can still misread a candidate. The output now carries the full
paragraph for that judgment, which makes the remaining risk visible rather than
turning it into a false mechanical block.
