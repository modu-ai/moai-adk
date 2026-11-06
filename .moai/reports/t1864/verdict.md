verdict: PASS
audited_sha: 9e4de7c132c641627e295285ed1f799b88d0ec8b
required_backend: codex pass
convergence_overall: pass

# t1864 — merge verdict (Class A, no SPEC)

Card: t1864 (run tn1p4e, lane-31). Branch `WT-run-adopt-errcheck`, merge base `bc5817d75`. This file is the SPEC-less card's merge evidence; the audit it cites is `.moai/reports/t1864/sync-audit.md`.

## Claim
The three `fmt.Fprintf(warn, ...)` calls at `internal/cli/factory_run_adopt.go` (lines 79, 91, 103) now discard their results explicitly with `_, _ = fmt.Fprintf(...)`, the package's convention for a best-effort write to a notes writer. `golangci-lint run --enable-only errcheck ./internal/cli/` goes from exit 1 (3 issues) to exit 0. No lint configuration change, no `//nolint`, message text unchanged.

## Evidence (deciding lines)
- RED at the base (`bc5817d75`), same command as the leader's: `exit=1`, `factory_run_adopt.go:79:15`, `:91:14`, `:103:14` "Error return value of `fmt.Fprintf` is not checked (errcheck)", `3 issues: * errcheck: 3` (`.moai/reports/t1864/progress.md`).
- GREEN, lane re-measure at `5976296b9` (this run): same command `exit=0`, `0 issues.`; `go build ./internal/cli/` exit 0; `gofmt -l internal/cli/factory_run_adopt.go` empty; `git diff bc5817d75 5976296b9 -- internal/cli/factory_run_adopt.go` shows 3 insertions and 3 deletions, the three call lines only.
- Scoped tests (implementer run): `go test ./internal/cli -run 'TestAdoptResumedLeaderRun|TestStrictResumeCarrierSeen|TestLeaderRunIDWasMinted' -count=1 -v` -> `ok ... internal/cli 5.960s`, seven tests PASS.
- Sync audit (`.moai/reports/t1864/sync-audit.md`): `verdict: PASS` @026340b5e, required backend codex pass, receipts `rcpt-542faae67b11d116c0e195ad` and `rcpt-8df6302abb885e179d0f2641`, both `codex_verdict: pass`; one MINOR Gap (exit codes of lint/vet runs not captured by the auditor).
- Card review (advisory): `.moai/reports/t1864/card-review.md`, verdict pass, no findings.
- HEAD confirmation after absorbing main (`48bfe97af`, merge commit `ea5dd0be6`), two rounds, recorded as they happened:
  - Round 1, `.moai/reports/t1864/sync-audit-head-confirmation.md`: `verdict: FAIL` @ea5dd0be6; the required codex backend returned a non-pass (`rcpt-86c0804f0d3fad176cfdb209`, `rcpt-e1f7b15447fa56d969e2cde4`, both `codex_verdict: fail`) solely because its sandbox could not load packages to run the acceptance command (`context loading failed: no go files to analyze`, exit 5); it named no code defect.
  - Lane-observed acceptance run outside that sandbox (`/private/tmp/lxq/accept-r2.txt`, machine-local): location `.moai/worktrees/t1864`, head `9e4de7c132c641627e295285ed1f799b88d0ec8b`, clean tree, `golangci-lint run --enable-only errcheck ./internal/cli/` (v2.1.6, go1.26.8), start 2026-10-18T03:01:03Z end 03:01:16Z, `exit: 0`, `0 issues.` (two earlier attempts were blocked by another lane's golangci-lint lock, exit 3, and are not results).
  - Round 2, `.moai/reports/t1864/sync-audit-head-confirmation-r2.md`: `verdict: PASS` @9e4de7c13, `audit_multi` called once with that evidence in its focus, receipt `rcpt-7f463ce25477394c7628b836`, `codex_verdict: pass`, no findings. The auditor's own run of the same command: `exit=0`, `0 issues.`.

## Baseline-attribution
Audited commit `9e4de7c132c641627e295285ed1f799b88d0ec8b` (tree `.moai/worktrees/t1864`; it is `ea5dd0be6` plus evidence files under `.moai/reports/t1864/`). The original sync audit (PASS) read `026340b5e21cb60c7de55dc6e3c7560e4a1d7856`; the card's code is byte-identical in all three commits. Differences after the audited commit are confined to `.moai/reports/t1864/`. The landing re-measure runs against the committed tree and is keyed to it, so its command and result are not appended here (that would change the tree); they are in the lane's landing report to the leader and in the re-measure verb's own record.

## Gaps
- Hosted CI has not run on this change; CI owns the repository-wide verdict.
- The full `internal/cli` test package was not run (scoped to the tests covering `adoptResumedLeaderRun`).
- The `moai` binary behind the audit tools lags the audited commit (reported by the audit as build lag).

## Residual-risk
- The warnings now ignore a failed write to the notes writer; that matches the package convention and the writer is a launch-time notes sink with nowhere to report a failure.
