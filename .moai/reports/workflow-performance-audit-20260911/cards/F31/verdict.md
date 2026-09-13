# F31 — Mode-aware sync scan and coverage scope

## Claim

Sync now selects and records a verification scope before diagnostics: `auto`
stays on changed packages and their impact closure, while `force` and
`project` explicitly own full-repository coverage. Pre-existing coverage debt
is reported separately.

## Evidence

Command:

```text
bash .claude/hooks/tests/test-sync-coverage-scope-contract.sh
```

Observed output:

```text
PASS: sync diagnostic and coverage scope is explicit for every mode
```

## Baseline-attribution

The contract test ran in `WT-workflow-audit-f31b` after fast-forwarding to the
F30 merge on local `develop`.

## Gaps

No repository with a multi-package impact graph was run through both auto and
force modes, so the measured command fan-out and coverage-time reduction are
not established here.

## Residual-risk

The runtime selector must still compute the import/dependency closure
correctly. A bad impact graph could under-scan an auto change; the report must
surface the selected package list so a reviewer can challenge it.
