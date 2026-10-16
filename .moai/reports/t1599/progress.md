# t1599 — run evidence (i18n-validator budget overrun flake)

Card class B (cause known). Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1599`, branch `WT-i18n-budget-flake`, base `0bd1da497`.

## Cause

`scripts/i18n-validator/main.go` `runWithBudget` raced the scan goroutine's `done` channel against
`time.After(budget)` in one `select`. When the scan finished before the timer goroutine fired, the
`done` arm won and returned exit 0 even though wall-clock time already exceeded the budget.
`TestBudget_TimeoutExitOnExcess` (1ns budget, 100 files) flaked on loaded runners for this reason
and was skipped on Windows.

## Fix

Record `start := time.Now()`; in the `done` arm, if `time.Since(start) > budget`, return the same
budget-exceeded result as the timeout arm (ExitCode 4, the existing budget message, violations
omitted). Timeout arm unchanged (message hoisted to a shared `budgetMsg` const). Windows `t.Skip`
(and the then-unused `runtime` import) removed from `TestBudget_TimeoutExitOnExcess`.

## RED evidence (commit 45fb4a20a, unmodified main.go)

Deviation: the briefed 200 iterations did NOT go red (`ok ... 1.313s`, `-race -count=1`); 2000 went
red intermittently (1/2000 per failing run); the committed test uses 20000 iterations.

Command: `TMPDIR=/private/tmp/lxv GOTMPDIR=/private/tmp/lxv go test [-race] -count=10 -run 'TestBudget_OverrunReportedWhenScanFinishesFirst' ./scripts/i18n-validator/`

`-race`: 10/10 runs red, e.g.

```
    main_test.go:552: 6/20000 iterations did not report the budget overrun (want exit 4 + 'budget' in stderr)
    main_test.go:552: 15/20000 iterations did not report the budget overrun (want exit 4 + 'budget' in stderr)
FAIL	github.com/modu-ai/moai-adk/scripts/i18n-validator	3.818s
```

No `-race`: 9/10 runs red (1/20000 to 8/20000 per run), `FAIL ... 3.501s`.

## GREEN evidence (commit 0cb764823)

- `go test -race -count=20 -v -run 'TestBudget_OverrunReportedWhenScanFinishesFirst|TestBudget_TimeoutExitOnExcess' ./scripts/i18n-validator/` → 40 `--- PASS` lines; `ok  	github.com/modu-ai/moai-adk/scripts/i18n-validator	9.227s`
- `go vet ./scripts/i18n-validator/` → `vet exit=0`
- `go test -race -count=3 ./scripts/i18n-validator/` → `ok  	github.com/modu-ai/moai-adk/scripts/i18n-validator	26.366s`, exit 0

All with `TMPDIR=/private/tmp/lxv GOTMPDIR=/private/tmp/lxv`, darwin, on this tree.

## Files changed

- `scripts/i18n-validator/main.go`
- `scripts/i18n-validator/main_test.go`

## Gaps

- Windows not run locally; removing the skip rests on the cause analysis (the overrun can no longer be lost on either select arm). CI on the integration branch owns the cross-platform and repository-wide test verdict — PENDING at report time.
- The original loaded-ubuntu-runner flake was not reproduced locally; RED was reproduced with the empty-directory analogue.
- Full repository suite not run (by instruction).

## Residual risk

- The RED test is probabilistic against the old code (roughly 0.005-0.08% of iterations per run); against the fixed code it is deterministic, since a 1ns budget is always exceeded by `time.Since(start)`.
