auditor-model: claude-opus-5-5

verdict: FAIL
audited_sha: 511d8cd9643ad175911538d048237d9c827dc2b8

# t2104 (GitHub #1938) sync audit, iteration 1

Score: 0.78 (Functionality 70 FAIL / Security 95 PASS / Craft 85 PASS / Consistency 90 PASS). Functionality is a must-pass dimension, so the overall result is FAIL.

Scope: commit 511d8cd964 changes three files and no production code. They are `internal/cli/factory_runs_rebind_expired_t2104_test.go` (2 tests), `.moai/docs/factory-expired-lease-dead-run.md` and `.moai/reports/t2104/progress.md`.

## Claim

1. The two pinned tests pass on the audited tree.
2. The doc's mechanism claims match `RebindRun` (`internal/homestate/factory_run_rebind.go`) and `factoryAssignReassign` (`internal/cli/factory_card.go`).
3. The tests are not vacuous. They fail when the expired-lease guard is inverted and when the divergence refusal is removed.
4. The card closes on "document + pin". The documented path does not reach the issue's goal of one row per card in the state the issue itself reports, and the decision depends on an unmeasured "no harm" premise. Iteration 1 therefore fails.

## Evidence

### E1 — the pinned tests on the audited tree
`go test ./internal/cli -run 'TestFactoryRunsRebind(MovesExpired|RefusesCardAlready)' -v -count=1`
```
--- PASS: TestFactoryRunsRebindMovesExpiredLeasedRow (1.69s)
--- PASS: TestFactoryRunsRebindRefusesCardAlreadyAssignedInLiveRun (0.77s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	4.409s
exit=0
```

### E2 — mutation check (scratch copies made with `git archive 511d8cd964`; the worktree was not modified)
Mutant m1, `factory_run_rebind.go:241`: `if holdsValidLease(c, now)` changed to `if !holdsValidLease(c, now)`
```
--- FAIL: TestFactoryRunsRebindMovesExpiredLeasedRow (2.58s)
    ... refused: card t1 of run run-src holds a live lease (lane-9, expires 2026-09-20T01:00:00Z); a leased card is not moved
--- FAIL: TestFactoryRunsRebindRefusesCardAlreadyAssignedInLiveRun (0.75s)
    ... does not name "run-tgt" / does not name "lease_holder"
exit=1
```
Mutant m2: the `divergentColumns` refusal block at lines 278-281 was deleted.
```
--- PASS: TestFactoryRunsRebindMovesExpiredLeasedRow (1.21s)
--- FAIL: TestFactoryRunsRebindRefusesCardAlreadyAssignedInLiveRun (1.25s)
    consolidation accepted a diverging live row:
        rebound run-src into run-tgt (owner dead)
        moved cards: none
        abandoned cards: t1
exit=1
```
Both mutants were killed. In m2, test 2 also exposes the silent-discard outcome: the source row is abandoned.

### E3 — doc claims checked against the source (by reading the code)
- "`assign --reassign` is scoped to the card's current dispatch run": correct. `factoryAssignReassign` (factory_card.go:4436-4439) refuses when the dispatch-current run differs from `--run`, and `ReassignCard` (card_reassign.go:58+) changes only the owner within one run_id. It never moves a row between runs. Also correct, but not stated in the doc: an explicit `--run` whose run row is not `active` is refused earlier by `resolveFactoryCardRun` (factory_card.go:3951-3952). That is the source of the "is not the currently active factory run" message cited in the progress record.
- "Moves intact (state, owner, lease fields)": correct. The re-key is `UPDATE cards SET run_id=?` only (rebind.go:289).
- "An expired lease does not block": correct. `holdsValidLease` returns `!c.LeaseExpired(now)` for lease-holding states (factory_run_lease_rebound.go:200-208).
- "The assign-first row is refused, naming `lease_holder`, `lease_expires_at`; nothing changes": correct. `divergentColumns` compares every non-excluded column. The target's `assigned` row has empty lease fields, and the refusal happens before any write. The test pins this with `lcDump` equality.
- References exist: `.moai/docs/af-flow-landed-card-close.md` is present, and `git cat-file -t f8ba9857bc` prints `commit`.

### E4 — auditor probe of the shape the issue reports: two dead runs each hold t1 `leased` with expired leases
This is a scratch-only test that uses the same helpers. Its output:
```
first rebind (run-a): err=<nil> out=rebound run-a into run-tgt (owner dead)
    moved cards: t1
rows: run-tgt:leased ; run-b:leased
second rebind (run-b): err=factory run consolidation refused: card t1: run run-tgt already holds a leased row diverging on lease_expires_at (target evidence "", source ""); the live row of run run-b stays and the conflict needs a fresh decision
rows: run-tgt:leased ; run-b:leased
run-a=retired run-b=active
```
The documented verb leaves two rows, and a dead run (`run-b`) that cannot be retired, in the exact shape #1938 reports (`tm4exu` and `tm85bw`, both `leased`/expired). The assign-first step plays no part in this.

### E5 — static checks
`gofmt -l internal/cli/factory_runs_rebind_expired_t2104_test.go` printed nothing (exit 0). `go vet ./internal/cli` exited 0 with no output.

### E6 — issue text
`gh issue view 1938`: "Expected: A documented leader verb that retires or hands over an expired `leased` row from a dead run ... so the card keeps one row." The "Not verified" list includes: "Whether the three-row state causes harm later (for example at `factory complete`, or in status counts)."

## Findings (structured defect list)

- **F1 [Medium] [blocking] [in-diff, reproduced] `.moai/docs/factory-expired-lease-dead-run.md:12-22`**. The doc's measured table covers only one dead run holding the card. In the state #1938 actually reports, two dead runs hold the card. The documented verb then moves one row, refuses the second consolidation on `lease_expires_at` (E4), and leaves two rows plus an active dead run that cannot be retired. The doc presents the verb as the answer to #1938 but does not cover this outcome, and an operator who follows it on the reported state ends up with an undocumented refusal.
  **Required fix:** add the multi-dead-run row to the doc's table. Pin it with a third test that mirrors E4 (second rebind refused, rows unchanged), and state what the operator does next. Alternatively, record a decision that this case needs a code path (a follow-up card) and stop calling the card's goal met.
- **F2 [Medium] [blocking] [in-diff: progress.md "Gaps" and "Residual-risk"]** The "document + pin, no code change" decision depends on the premise that the two-row state does no harm, and that premise was never measured. The progress record lists it honestly as a Gap. Under surface 4 of verification-claim-integrity, however, a recommendation still cannot rest on an unverified premise. E4 shows that the two-row residue is reachable through the documented path itself, not only through the assign-first misstep, so the question carries the decision's weight and is not hypothetical.
  **Required fix:** measure the two-row state with one isolated-sandbox probe or test. Cover (a) `moai factory status` row and count output and (b) `moai factory complete` (or its row-resolution step) for a card with a live-run row plus a leftover `leased` row in another run. Record the verbatim output. If harm shows, the "if none" branch does not apply and a code-change follow-up is needed.
- **F3 [Low] [optional] [in-diff] `.moai/docs/factory-expired-lease-dead-run.md:3-6`**. The doc attributes the scoping to `factoryAssignReassign` only. The `--run <non-active run>` refusal that the operator actually sees comes from `resolveFactoryCardRun` (factory_card.go:3951). A dead-owner run whose row is still `active` passes that check, and it is then `ReassignCard` that keeps the move within one run. The doc could name both.
- **F4 [Low] [optional] [in-diff] `progress.md:7-10`**. The "earlier probe" results (`assign --reassign --run <dead>` refused; `status lists both`) come from a deleted probe. Their verbatim output was not kept, so this audit could not re-witness them.
- **F5 [Info] [optional] [in-diff] test fixture**. The leases expire on fixed past dates (`2026-09-20`), which is safe against the wall clock. The tests do not check that `card_dispatch` is repointed or that `run.retired` events are written. Pre-existing consolidation tests cover those, but this audit did not confirm that.

## Baseline-attribution

Tree: worktree t2104, HEAD `511d8cd9643ad175911538d048237d9c827dc2b8` (`git rev-parse HEAD`, clean status at session start). E1 and E5 ran against this tree. E2 and E4 ran on scratch copies extracted from `git archive 511d8cd964` with a single change each, shown by `diff` before the runs. The Go toolchain is the session's default `go`. No moai binary was used, so §2.2 tool provenance does not apply.

## Gaps

- I did not call `mcp__moai__codex_audit`, `audit_multi`, `glm_audit` or `claude_audit`, because the task explicitly forbids them. The SubagentStop hook demanded a codex receipt (`AUDIT_RECEIPT_VIOLATION`), and I declined it for that reason. No receipt exists.
- I did not run `moai verify audit-plan`. The cross-model leg is forbidden for this audit, so its result would not change what I do.
- I did not run the full `internal/cli` suite, did not run on Windows or in CI, and did not measure coverage, since no production code changed.
- E4 assumes both dead runs are `status=active`. In #1938, `next --card t1` reported ambiguity only between `tm85bw` and `tnhicy`, which suggests `tm4exu` may already have been non-active. In that case the rebind leaves a `leased` row in a retired run rather than refusing. The leftover-row question of F2 applies either way.
- My worktree-guard refusals: two compound `git archive | tar` commands were refused. I re-ran them as split plain commands with the same effect. A cleanup `rm` of the scratch copies was denied, so the scratch copies remain in the session scratchpad.

## Residual-risk

- Without the F2 measurement, a leftover `leased` row in another run may still confuse `factory complete` row resolution, status counts, or the lease ambiguity detection that #1938 saw first (`AMBIGUOUS_FACTORY`).
- The tests pin the CLI path only. A future change to `cardPairExcludedColumns` that excluded the lease columns would turn the assign-first refusal into a silent abandon. Test 2 would catch that, as m2 shows.

## Iteration history

- Iteration 1 (this file): FAIL 0.78 at 511d8cd964. Blocking findings: F1 and F2.
