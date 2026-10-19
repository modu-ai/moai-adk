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

## Hold maintained — 2026-10-17T10:37Z (lane-17, run tn1p4e gen 2, fresh session after reboot)

- Leader notice (tn1p4e): run tmvgvn retired 2026-10-17T10:01Z; t1599's factory record is stranded there, so stage/complete/renew fail with `factory card not found`; re-lease via `moai factory next --card t1599` UNLESS the card carries an operator hold or wait condition — in that case keep the condition and do not lease.
- Hold-check result: HOLD EXISTS and is still live (premise re-verified 2026-10-17T10:37Z, this tree):
  - `.moai/state/audit-receipts/rejections/sync-auditor--unknown-spec.json` present — `cause: "verdict line missing"`, `rejected_at: 2026-10-16T21:22:15Z`, `reentry_warned: true` (cat-observed);
  - both receipts still stand (`rcpt-56d38345…` required-backend codex FAIL, `rcpt-16bce58f…` native pass); nothing cleared since the 10-16 decision record above;
  - tree unchanged: `WT-i18n-budget-flake` clean at `f8443930e`; queue state still `picked`; no superseding operator approval found.
- Disposition: NOT re-leased into tn1p4e; the operator-held keep-set condition is kept (leader conditional honored; guard not bypassed). Unblock path: operator clears/approves the stale unknown-spec receipt record (precedent t1705/t1687) or otherwise resolves the no-SPEC verdict-line gap; then re-lease with `moai factory next --card t1599` — stages re-record plan→run→sync, the F1 seam repair (never-firing timer injection per the audit's recommendation) must land, and the sync audit must be re-run before landing.
- Leader notified; this record is the source of truth for the wait state (the message is a nudge only).
- FYI from the same notice, no condition on this card: path-guard block card issued as t1824 (blocked); rationale/recurrence-prevention t1825.

## Delegated decision executed — stale receipt record cleared — 2026-10-17T11:26Z (lane-17, tn1p4e gen 2; earlier draft said 10:44Z — corrected after re-checking the clock, see the wait entry below)

- Authority: leader notice (tn1p4e unattended window) — the operator left decisions with the leader (Jev/Codex advisory); for t1599 the leader authorized the t1705/t1687-precedent remedy: clear the stale unknown-spec receipt record, record target/rationale/precedent here FIRST, then re-lease via `moai factory next --card t1599` and continue with the F1 repair + re-audit. Also permitted by SPEC-CODEX-AUDIT-GATE-AXES-001 plan.md §B.6 (the K4 clearing class: "사람이 기록 파일을 직접 지우는 것도 허용된다").
- Cleared target (the ONLY file removed): `.moai/state/audit-receipts/rejections/sync-auditor--unknown-spec.json`. Both audit receipts are untouched — they are round-1 history and evidence, not the blocking record.
- Preserved content of the removed record (verbatim):

  ```json
  {"agent_type": "sync-auditor", "spec_id": "unknown-spec", "agent_id": "ac0d1c63a037390cb", "cause": "verdict line missing", "rejected_at": "2026-10-16T21:22:15.726712Z", "reentry_warned": true, "tree_root": "/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1599"}
  ```

- Rationale (guard source read: `internal/hook/audit_receipt_guard.go`): `checkAuditReceiptSpawn` denies manager-develop/manager-docs/manager-git while ANY receipt-kind rejection exists for this tree. This record existed ONLY because the round-1 auditor's final message lacked a parseable verdict line (cause "verdict line missing") — the round-1 verdict itself (FAIL, F1: RED-test non-determinism) was never overturned. Round 1 ended FAIL, so the auto-clear path (a proven auditor PASS clears the role's rejections, plan.md §B.6 operator decision K4) could not fire; without the delegated remedy the repair spawn stays blocked indefinitely.
- Recurrence guard for round 2 (recorded so the next auditor does not re-wedge the card): the auditor must end its FINAL message with the verdict line as the LAST non-empty line, format `AUDIT-VERDICT: <PASS|PASS-WITH-DEBT|FAIL> spec=<SPEC-ID> receipts=<rcpt-id>[,...]` (`internal/auditreceipt/store.go:140` verdictLinePattern; plan.md §B.6). A parseable FAIL needs no receipt and records nothing; a PASS must cite a receipt minted during its own instance (`CheckCitedReceiptsSince`).

## Lease attempt refused — serial slot held — 2026-10-17T11:28Z (lane-17, tn1p4e)

- Precondition completed first (entry above): the stale unknown-spec rejection record was cleared at ~11:26Z (rejections/ verified empty after the rm; both round-1 receipts preserved). The clearing entry's heading originally read "10:44Z" — written before the clock was re-checked; actual execution time is ~11:26Z. Corrected in the heading.
- `moai factory next --card t1599` (MCP, project_root=primary checkout) → refused verbatim: `factory next: refused serial-slot: another serial card is in flight and holds the serial slot`. t1599 is mode=serial; the tn1p4e serial slot is held by t1697 (lane-13, live lease until 2026-10-17T11:30:57Z at status time 11:28:48Z); t1825 (serial, lane-1) shows its lease expired 10:57:43Z; t1615 and t1536 also wait on the same slot per their wait records. The refusal consumed nothing — t1599 still has no tn1p4e record (verified: `moai factory status` shows t1599 only under tmvgvn).
- Disposition: disk wait + retry (leader unattended rule 2; bare next remains forbidden). Recheck cron c6cc46bd fires every ~5 min and retries ONLY the named-card lease `factory next --card t1599`; on success it deletes itself and continues the pipeline: `factory stage t1599 run` (class B) → manager-develop F1 seam repair (effort medium) → sync re-audit (verdict-line format enforced in the spawn prompt) → `moai factory complete t1599` landing. No push/PR/tag/deploy; done + tree preservation after the merge per the unattended rules. Lease renewal runs in-turn between work steps.
- Known gap carried forward (recorded, not bypassed): the re-audit's auditor cannot emit a parseable verdict line for a spec-less card (`specIDBody` requires SPEC-form), so its stop may re-create an unknown-spec rejection. The landing path needs no gated Agent spawns after the audit, so the landing is not wedged; any such record is left for the operator's stale-record sweep and noted here when it forms.

## Wait handed back to leader — 2026-10-17T12:0xZ (lane-17, tn1p4e)

- 5 retry attempts over ~25 min all refused `serial-slot` (holders observed: t1697 → t1668, live lease). Leader then assigned lane-17 a parallelizable card (t1839) and directed: t1599 stays parked, retry cron deleted (c6cc46bd removed this turn), and the LEADER will re-assign t1599 when the serial slot frees — lane-17 no longer self-retries.
- State on park: stale rejection record already cleared (entry above), tree clean at f8443930e, no tn1p4e record for t1599. Resume path unchanged: leader assigns → `moai factory next --card t1599` → stage run → F1 seam repair (never-firing timer injection) → re-audit (verdict-line last line) → `moai factory complete`. This progress record is the source of truth for the wait.

## F1 repair (lane-27, tn1p4e)

Scope: `scripts/i18n-validator/main.go` and `main_test.go` only. Production done-arm check (`time.Since(start) > budget`) and the timeout arm are unchanged; the only production change is the `runWithBudgetAfter` timer seam (`runWithBudget` is a one-line wrapper passing `time.After`). No skip, no timeout widening.

Commit graph (ordering witness, verification-claim-integrity.md §2.3): `1d9100f5b` RED test (package does not compile there: `undefined: runWithBudgetAfter`, from `go vet` at that tree) precedes `54121066e` GREEN seam. Judging toolchain: `go` as resolved on PATH, no moai tool build involved.

### RED — new test against pre-fix select logic (overlay, outside the repo)

Overlay file `/private/tmp/t1599-red/main_prefix.go` = `git show 45fb4a20a:scripts/i18n-validator/main.go` (pre-fix body, no done-arm elapsed-time check) plus the seam signature only. Overlay map: `/private/tmp/t1599-red/overlay.json`, replacing the repo `main.go`. Test tree: `1d9100f5b`.

Command:
`go test -overlay /private/tmp/t1599-red/overlay.json -count=20 -run TestBudget_OverrunReportedWhenScanFinishesFirst ./scripts/i18n-validator/` (TMPDIR/GOTMPDIR=/private/tmp/t1599-tmp) — exit=1.

Count of failing runs: `grep -c -- "--- FAIL: TestBudget_OverrunReportedWhenScanFinishesFirst" red20.txt` → `20` (20/20). Verbatim lines (identical in all 20):

```
--- FAIL: TestBudget_OverrunReportedWhenScanFinishesFirst (0.00s)
    main_test.go:543: expected exit code 4 when the scan finishes after the budget, got 0
    main_test.go:546: expected 'budget' in stderr, got: ""
FAIL
FAIL	github.com/modu-ai/moai-adk/scripts/i18n-validator	0.235s
```

With `GOMAXPROCS=1`, `-count=1` (same overlay): the same three failure lines, then `FAIL	github.com/modu-ai/moai-adk/scripts/i18n-validator	0.082s`. (The old loop test passed 3/5 under GOMAXPROCS=1 per the audit.)

### GREEN — repo tree at `54121066e`

- `go test -race -count=20 -run 'TestBudget_OverrunReportedWhenScanFinishesFirst|TestBudget_TimeoutExitOnExcess' ./scripts/i18n-validator/` → `ok  	github.com/modu-ai/moai-adk/scripts/i18n-validator	1.561s`
- Same with `GOMAXPROCS=1` → `ok  	github.com/modu-ai/moai-adk/scripts/i18n-validator	1.826s`
- `go vet ./scripts/i18n-validator/` → no output, `vet_exit=0`
- `gofmt -l scripts/i18n-validator/` → no output, `gofmt_exit=0`
- `go test -race -count=1 -cover ./scripts/i18n-validator/` → `ok  	github.com/modu-ai/moai-adk/scripts/i18n-validator	15.127s	coverage: 86.0% of statements`

Files changed: `scripts/i18n-validator/main.go` (seam), `scripts/i18n-validator/main_test.go` (deterministic test; `TestBudget_TimeoutExitOnExcess` untouched).

### Gaps

- Windows not run locally. Full repo suite not run (CI owns the repository-wide verdict; PENDING at report time).
- The RED overlay was a hand-built pre-fix copy of `main.go` (from `45fb4a20a`), not a checkout of that commit; the new test file in the overlay run is the committed `1d9100f5b` test.
- The `-run` GREEN selector was checked by the `ok` line only (no `-v` count of swept tests was taken); the two named tests exist in `main_test.go` (`grep -n "TestBudget_"` earlier in this run listed both).
- No push, PR, or merge performed.

### Residual-risk

- `TestBudget_TimeoutExitOnExcess` still depends on a real 1ns timer race; it passes under either arm winning (both yield exit 4), so it was left alone as instructed.
- A nil-channel timer plus a 1ns budget relies on any scan taking longer than 1ns of wall clock; true on any real clock but not enforced by the type system.
- Sync re-audit is still owed on this repair.

## Sync re-audit round 2 PASS — 2026-10-18 (lane-27, tn1p4e)

- Sync audit round 2: PASS, audited_sha 04d51d9815c56161e5b2ff995fae6dd2b47ba59a, required backend codex=pass rcpt-767efceacf7485a235b800a1 (+ native codex_audit pass rcpt-cca34120a81a8e012aaa14f2). Verdict file: .moai/reports/t1599/sync-audit.md (committed 78ed98f94). Auditor judging build ee981a932.
- Lane-observed: tree HEAD 78ed98f94 clean; `go test -race -count=20 -run 'TestBudget_OverrunReportedWhenScanFinishesFirst|TestBudget_TimeoutExitOnExcess' ./scripts/i18n-validator/` -> `ok  	github.com/modu-ai/moai-adk/scripts/i18n-validator	1.353s` (run on 04d51d981, same code as HEAD).
- card-review: .moai/reports/t1599/card-review.md refreshed (codex_review scope=card, pass, 0 findings).
- Gaps: Windows and the full repo suite not run locally (CI on the integration branch owns them); base 0bd1da497 predates the main advance, so landing re-measures after absorbing main.
- Next: stage merge-ready, then `factory complete` in the leader's slot order.
