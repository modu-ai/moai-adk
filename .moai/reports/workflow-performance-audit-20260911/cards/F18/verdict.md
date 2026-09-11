# F18 verdict

## Claim

F18 is fixed: the CHANGELOG AC grammar includes an optional lowercase suffix,
so `AC-SYN-001a` and `AC-SYN-001b` are counted as two distinct criteria.

## Evidence

The documented parser already carries `[a-z]?`; the regression fixture extracts
two suffixed identifiers and asserts a distinct count of two.

Command:

```text
.claude/hooks/tests/test-ac-suffix-contract.sh
PASS: AC suffixes a/b remain distinct criteria (count=2)
```

## Baseline-attribution

The command ran in `WT-workflow-audit-f18`, based on local `develop` commit
`8ab854d70` before this card's commit.

## Gaps

The fixture validates identifier extraction, not a full CHANGELOG emission with
all reserved-token states.

## Residual-risk

Any future parser replacement must preserve suffix, domain-less, and
digit-bearing identifier cases together; the current fixture should be extended
if the grammar changes.
