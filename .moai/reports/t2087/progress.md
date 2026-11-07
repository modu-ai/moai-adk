# t2087 — progress record

## Batch landing record (af-batch-20261107)

Landed in local `main` by the 2026-11-07 batch (merge `b327122293`, batch tip `4fb7c8fdf1`, operator FF). Batch repair commit `c94332235d` registered this card's fixture identity `t2087@example.com` (`internal/spec/lint_gitbatch_test.go`) in `builtinCommitIdentityDenyEmails`; `TestCommitIdentityGuard_BuiltinListCoversFixtureEnumeration` was RED at the card head and GREEN after the repair (observed by the batch auditor). Batch verdict: `.moai/reports/af-batch-20261107/verdict.md` (PASS-WITH-DEBT 0.88, audited_sha 4fb7c8fdf1…).
