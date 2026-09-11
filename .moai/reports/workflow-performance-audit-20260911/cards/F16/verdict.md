# F16 verdict

## Claim

F16 is fixed: clarity interview output records a provisional tier before
research, explicit user tiers are reused, and final tier questioning is
conditional on missing or changed evidence. Research can promote a tier but is
not required merely to discover an already explicit Tier S.

## Evidence

The interview and spec-assembly instructions were aligned and the contract test
checks the pre-research handoff and conditional final question.

Command:

```text
.claude/hooks/tests/test-tier-before-research-contract.sh
PASS: tier routing is resolved before optional research and finalized once
```

## Baseline-attribution

The command ran in `WT-workflow-audit-f16`, based on local `develop` commit
`a1fa573e3` before this card's commit.

## Gaps

The tier classifier remains an orchestration rule rather than a measured cost
model. No user waiting-time benchmark was run.

## Residual-risk

An overly optimistic provisional tier can still be promoted only after research
findings arrive; the final tier gate must remain mandatory before artifacts are
written.
