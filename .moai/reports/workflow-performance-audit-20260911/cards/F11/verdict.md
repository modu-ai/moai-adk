# F11 verdict

## Claim

F11 is fixed: every cross-build PID is associated with a target, waited on
individually, and contributes to an aggregate failure status. A bare `wait` can
no longer hide a failed target.

## Evidence

The delivery example collects an associative PID map and reports each target;
the regression fixture starts one failing and one successful child and asserts a
matrix failure.

Command:

```text
.claude/hooks/tests/test-cross-build-status-contract.sh
PASS: one failed cross-build target makes the matrix fail
```

## Baseline-attribution

The command ran in `WT-workflow-audit-f11`, based on local `develop` commit
`aace3acd4` before this card's commit.

## Gaps

The fixture exercises status aggregation rather than five real cross-toolchain
builds, which remain environment-dependent.

## Residual-risk

An added target must be inserted into both the PID map and the wait list. The
explicit target list makes omission reviewable but does not prevent a future
copy-and-paste omission without the contract test being updated.
