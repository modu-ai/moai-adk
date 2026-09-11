# F10 verdict

## Claim

F10 is fixed: the local lint mirror now distinguishes tool absence (SKIP), a
successful installed linter (PASS), and an installed linter returning non-zero
(FAIL with the original status).

## Evidence

The delivery example uses `command -v` followed by an explicit `if/elif/else`;
the fixture supplies a fake linter that exits 7 and asserts that 7 is returned.

Command:

```text
.claude/hooks/tests/test-lint-status-contract.sh
PASS: installed lint failure preserves exit status 7
```

## Baseline-attribution

The command ran in `WT-workflow-audit-f10`, based on local `develop` commit
`02383ac75` before this card's commit.

## Gaps

The fixture does not install or run the repository's real golangci-lint binary;
it verifies status propagation at the shell boundary.

## Residual-risk

Other language examples can reproduce the same short-circuit mistake. Apply the
same status-preserving pattern to every local-CI mirror branch as it is edited.
