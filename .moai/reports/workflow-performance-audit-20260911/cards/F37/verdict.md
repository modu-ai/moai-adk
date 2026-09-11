# F37 — TDD RED semantic classification

## Claim

The TDD workflow now distinguishes `EXPECTED_RED` from regression, tool, and
timeout failures. Only an intended assertion failure with verbatim evidence is
allowed to advance from RED; GREEN and REFACTOR require the regression set to
pass.

## Evidence

Command:

```text
bash .claude/hooks/tests/test-tdd-result-contract.sh
```

Observed output:

```text
PASS: TDD distinguishes expected RED from regression and tool failures
```

## Baseline-attribution

The contract test ran in `WT-workflow-audit-f37b` after fast-forwarding to the
F36 merge on local `develop`.

## Gaps

No intentionally broken test, compile-error fixture, or timeout was executed
through the real manager-develop cycle.

## Residual-risk

The runtime classifier must inspect test identity and output rather than merely
map a non-zero exit code to RED; otherwise the documented safety boundary can
still be bypassed by a caller.
