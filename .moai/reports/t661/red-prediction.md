# t661 — RED control prediction (pinned before any mutant is injected)

Target test: `go test ./internal/cli/ -count=1 -v -run '^TestUserHomeDirFnSandboxesRealHome$'`
(does not touch `~/.claude`; it only resolves paths). Base: this branch with the C wrapper committed.

## Mutant M1 — remove the TestMain call
Delete both `restoreUserHomeDir := sandboxUserHomeDir()` and `restoreUserHomeDir()` in `TestMain`
(deleting only one breaks compilation — not a valid RED).

Prediction: subtest `real_home_redirects_to_sandbox` FAILS with
`userHomeDirFn() = "<real home>", which is the real home. TestMain must call sandboxUserHomeDir() ...`;
subtest `overridden_home_passes_through` PASSES; package exit 1; no
`moai-cli-test: userHomeDirFn redirected` line in the output.

## Mutant M2 — unconditional redirect
In `homeRedirectingFn`, change the pass-through condition to `if err != nil {` (so every successful
result is replaced by the sandbox).

Prediction: subtest `real_home_redirects_to_sandbox` PASSES; subtest `overridden_home_passes_through`
FAILS with `userHomeDirFn() = "<sandbox>" with HOME="<tmp>"; want the overridden HOME returned
unchanged ...`; package exit 1.

Falsifier: either mutant leaving the guard fully green means that branch of the guard is vacuous.
Each mutant is reverted and the revert proven with an empty `git diff --stat -- internal/cli/main_test.go`
before the next step.
