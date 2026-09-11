# t661 follow-up — develop CI Lint red (errcheck) fix

## Claim
The develop CI Lint failure `internal/cli/main_test.go:254:16: Error return value of fmt.Fprintf is not checked (errcheck)` (lead-measured: CI run 34522480728 on `f6b121a9c`; previous `ed71054d3` Lint passed) is removed by one line: `_, _ = fmt.Fprintf(homeRedirectStderr, ...)`.

## Evidence
- Premise re-measured: `git fetch origin develop` exit 0; origin/develop = local develop = `f6b121a9c`; line 254 of `internal/cli/main_test.go` was `fmt.Fprintf(homeRedirectStderr,`.
- Absorb: `git merge --no-ff f6b121a9c` exit 0 -> `e793fd1bc` (^1 `ec678fe50`, ^2 `f6b121a9c`), clean.
- Diff: `internal/cli/main_test.go | 2 +-`, only `-fmt.Fprintf(homeRedirectStderr,` / `+_, _ = fmt.Fprintf(homeRedirectStderr,`.
- `gofmt -l internal/cli/main_test.go`: exit 0, 0 bytes output.
- Pre-check before lint: pgrep for go test/cli.test/go vet/go build/golangci-lint exit 1 (none); control pgrep -x zsh 40; load `{ 5.32 5.74 8.13 }`.
- `golangci-lint run ./internal/cli/` (lint-run.txt): exit 0, output `0 issues.`

## Baseline-attribution
Local lint on tree `e793fd1bc` plus the one-line edit (committed next), golangci-lint 2.10.1 (`/opt/homebrew/bin/golangci-lint`, built go1.26.0). Config `.golangci.yml`: errcheck enabled with `check-blank: false` (so `_, _ =` is accepted), no test exclusion (golangci default lints test files), `max-issues-per-linter: 0`.

## Gaps
- The RED was not reproduced locally before the edit (slot approved one lint run); the RED rests on the lead-measured CI run. The config reading above is the control that a local `0 issues.` is not produced by skipping test files.
- CI installs golangci-lint v2.1.6 (per `.golangci.yml` header); the local judge is v2.10.1. The CI Lint job after the lead push is the verdict.

## Residual-risk
- Version skew between local and CI linters could still surface a different finding in CI.

## Integration window
- `moai integration status` free, then `moai integration acquire --name lane-3` exit 0 (since 2026-09-10T23:52:02Z).
- Local develop had moved to `96004e166` (4 commits after `f6b121a9c`); `git fetch origin develop` exit 0, origin/develop `f6b121a9c`, left-right `0 4` (local develop current).
- Re-absorb: `git merge --no-ff 96004e166` exit 0 -> `4147eeaef` (^1 `c5fbe8628`, ^2 `96004e166`), clean.
- Delta vs the linted tree `c5fbe8628`: 1 file, `.moai/reports/t464/reverify-verdict-20260907.md` (no Go file, no go.mod/go.sum, not under an embed path). The lint result carries over; no re-run.
