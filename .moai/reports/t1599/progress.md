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

## Blocker — 2026-10-16 ~21:45Z (lane-17)

- Sync audit FAIL (F1): the RED test TestBudget_OverrunReportedWhenScanFinishesFirst is probabilistic (pre-fix main.go, GOMAXPROCS=1: 3/5 PASS). Required codex fail rcpt-56d38345ae2e668229271b84; native codex_audit pass rcpt-16bce58f88a1c681e7689d6b. Production fix 0cb764823 judged correct.
- Planned repair (not applied): seam runWithBudgetAfter(root, budget, after func(time.Duration) <-chan time.Time); test injects a never-firing timer so only the done arm can win → deterministic RED on pre-fix code, GREEN after.
- Repair spawn refused by the audit-receipt guard: `AUDIT_RECEIPT_VIOLATION: manager-develop cannot be spawned while an audit PASS stands unproven ... Outstanding: sync-auditor / unknown-spec / verdict line missing` (record .moai/state/audit-receipts/rejections/sync-auditor--unknown-spec.json in this tree). Cause: card has no SPEC, so `spec=t1599` does not match the verdict-line SPEC regex (known gap, cf. t1705/t1687 where the operator approved removing the stale record).
- decision record: decided_by=lane-17(acting leader) evidence_refs=guard-refusal;sync-audit.md ladder_path=keep-set (guard is operator-held) → hold; not bypassed (no direct edit, no alternate agent type, no fabricated SPEC id).
