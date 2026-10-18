# t1864 — errcheck: three unchecked fmt.Fprintf results in factory_run_adopt.go (Class A, run)

Card: t1864 (run tn1p4e, lane-31, P1, main-red express). Tree `.moai/worktrees/t1864`, branch `WT-run-adopt-errcheck`, base `bc5817d75` (main after the t1858 landing).

## Claim
`internal/cli/factory_run_adopt.go` lines 79, 91 and 103 call `fmt.Fprintf(warn, ...)` without handling the return values, so `golangci-lint run --enable-only errcheck ./internal/cli/` exits 1 with 3 issues. Source: t1707's `adoptResumedLeaderRun` warnings. CI Lint runs only on origin/main, so the red has not surfaced yet.

## Evidence (this run, this tree `bc5817d75`)
```
golangci-lint --version  -> golangci-lint has version v2.1.6 built with go1.26.8
golangci-lint run --enable-only errcheck ./internal/cli/   (exit=1)
internal/cli/factory_run_adopt.go:79:15: Error return value of `fmt.Fprintf` is not checked (errcheck)
internal/cli/factory_run_adopt.go:91:14: Error return value of `fmt.Fprintf` is not checked (errcheck)
internal/cli/factory_run_adopt.go:103:14: Error return value of `fmt.Fprintf` is not checked (errcheck)
3 issues:
* errcheck: 3
```
The file's and package's convention for a best-effort write to a warning/notes writer is `_, _ = fmt.Fprintf(w, ...)` (for example `internal/cli/factory.go:916`, `internal/cli/factory_adopt.go:95-100`). The warnings go to a launch-time notes writer; a failed write has nowhere to be reported, so ignoring the result is deliberate and matches those sites.

## Repair design
Only the three call sites change: `fmt.Fprintf(warn, ...)` -> `_, _ = fmt.Fprintf(warn, ...)`. No lint configuration change, no `//nolint`.

## Gaps
- The first attempt to rename the branch in this session was refused by the branch guard before any change: `BRANCH_GUARD_VIOLATION: git branch in primary checkout (use a worktree; ...)`. The session's working directory was still the primary checkout at that moment (the `cd <tree> && git branch -m` form is read against the session directory). It was not routed around: once the session's working directory was this worktree the same rename ran there and succeeded.
- No code change yet at the time of writing; verification and the re-run of the same command are recorded below.

## Run evidence (this run, tree 66d752f78 + the 3-line edit)
- RED re-run before editing: `golangci-lint run --enable-only errcheck ./internal/cli/` -> exit=1, 3 issues (lines 79, 91, 103).
- Edit: `fmt.Fprintf(warn, ...)` -> `_, _ = fmt.Fprintf(warn, ...)` at those three sites; message text unchanged.
- GREEN: same command -> exit=0, `0 issues.` (output: /private/tmp/lxq/errcheck-green.txt, machine-local).
- `gofmt -l internal/cli/factory_run_adopt.go` -> empty; `go build ./internal/cli/` -> exit 0.
- `go test ./internal/cli -run 'TestAdoptResumedLeaderRun|TestStrictResumeCarrierSeen|TestLeaderRunIDWasMinted' -count=1 -v` -> 7 PASS, `ok ... 5.960s`.
- `golangci-lint run ./internal/cli/` (default config) -> exit 0, `0 issues.`
- Judging build: golangci-lint v2.1.6 (go1.26.8), installed binary, not built from this tree (it is a third-party tool, no tree lag applies).
- Gaps: full internal/cli suite not run (left to CI by instruction); test/lint output files live in /private/tmp/lxq and are not committed.

## HEAD-confirmation audit after absorbing main: FAIL on the required codex gate (lane-31)

Audit: `.moai/reports/t1864/sync-audit-head-confirmation.md`, `verdict: FAIL` @`ea5dd0be6c030cd11855f60d0904b5ebb2eaed46`, `required_backend: codex fail`, receipts `rcpt-86c0804f0d3fad176cfdb209` and `rcpt-e1f7b15447fa56d969e2cde4` (both `codex_verdict: fail`).
Cause (codex summary, verbatim from the verdict file): `[P2] 병합 트리의 acceptance 명령 통과를 확인할 수 없음 ... golangci-lint run --enable-only errcheck ./internal/cli/를 두 번 실행했지만, 모두 context loading failed: no go files to analyze로 종료됨(exit 5). ... absorb로 인한 코드 결함은 발견하지 못함. ... 권한 확장 실행 요청은 approval request failed로 거부됨.` The finding is that codex's sandbox could not load packages to run the acceptance command; it names no code defect. The auditor's own run of the same command on the merged tree: exit 0, `0 issues.`
Disposition: the required gate is not lowered and `verdict.md` is NOT re-bound to `ea5dd0be6` (that would claim a pass the gate did not give). The audited-line verdict (`sync-audit.md`, PASS @026340b5e) stands for the card's own change. The lane asked the leader for the gate disposition (re-run with the observed acceptance evidence supplied to the codex gate, or the leader's ruling).

## Orchestrator convergence check for the r2 audit (lane-31, 2026-10-18T03:15:44Z)

The r2 verdict (`.moai/reports/t1864/sync-audit-head-confirmation-r2.md`, `audit_multi` receipt `rcpt-7f463ce25477394c7628b836`) states the convergence is not final until the orchestrator writes the digest to a result file and runs `moai verify audit-plan --result-file`. Done by the lane, which is the card's orchestrator:
- Result file (machine-local, `.moai/state/` is not tracked): `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1864/.moai/state/audit-plan-result.json`, content = the digest members copied verbatim from the r2 verdict line `audit_multi result` (per_backend_verdicts codex/required/pass, overall_verdict pass, disagreement_flag null, participant_count 1, fail_open_backends [], audit_receipt, plan_source config; `gate_unmet` absent).
- Command: `moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1864 --result-file /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1864/.moai/state/audit-plan-result.json`, tree HEAD `e2332cd2bab97d0a793b36024f313d34ccb71b73`.
- Result: `exit=0` (start and end 2026-10-18T03:15:44Z); `config_status: "ok"`, `enforced_required: ["codex"]`, and
```
"convergence_check": {
  "ok": true,
  "unmet": []
}
```
- Preserved unchanged: the round-1 FAIL verdict `sync-audit-head-confirmation.md` (HEAD ea5dd0be6, codex fail, receipts `rcpt-86c0804f0d3fad176cfdb209` and `rcpt-e1f7b15447fa56d969e2cde4`).
Gap: the installed `moai` binary behind this verb is the one the audit tools report as lagging the audited commit; the verb reads configuration and a result file, not repository code.
