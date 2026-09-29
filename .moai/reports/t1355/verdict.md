# t1355 — factory-next record-and-claim race: `illegal card transition: leased → leased`

Class B (cause established by orchestrator; lane verified every read, RED-first repro, minimal fix, scoped re-measurement). Tree: WT-factory-next-lease-red @ 6bcc7e4ac (origin/develop tip at dispatch). No push (leader batch owns it).

## CI evidence

Run 36577159420, commit 6bcc7e4ac, ubuntu:

```
factory_classify_test.go:256: lane lane-2 errored: illegal card transition: leased → leased
factory_classify_test.go:269: lanes hold 1 distinct cards, want 2
```

Not deterministic on darwin — local run of the goroutine test passed; the interleaving is pinned by construction in the new repro test instead.

## Root-cause chain (all reads verified in this tree)

1. `factoryNextSelectAndLease` (`internal/cli/factory_card.go:364-372`) snapshots `db.ListCards` into the `recorded` set.
2. Arm (b2) (`factory_card.go:446-454`) selects a queue item with `State == picked && !recorded[id]`. If the queue read sees the item `picked` while the ListCards snapshot predates the other lane's `RecordPicked` commit, the item is chosen as "unrecorded".
3. `factoryNextRecordAndClaim` (`factory_card.go:504-509`, pre-fix) calls `db.RecordPicked(..., homestate.CardFields{}, ...)`; `RecordPicked` (`internal/homestate/card_picked.go:149-151`) returns the EXISTING row unchanged when fields are empty — whatever state it is in. Here: the other lane's row already at `leased` v3.
4. `factoryNextClaim` sees `State != picked`, skips T2, fires T3 `→ leased` with `ExpectedVersion` matching (3==3) → version check passes → edge lookup fails → `illegal card transition: leased → leased`.
5. `factoryNextClaimRefused` (`factory_card.go:546-551`) maps only `ErrStaleVersion`/`ErrLeaseHolder` to a retry; `ErrIllegalTransition` is a hard error → the lane errors instead of re-selecting.

The store's refusal is correct. The defect: the claimer treats RecordPicked's "existing row returned" as "fresh picked row I just created".

## RED (before the fix)

Test `TestFactoryNextRecordAndClaimRaceOnLeasedRow` (`internal/cli/factory_classify_test.go`): queue item t1 at queue state `picked` (parallelizable), row pre-driven picked→assigned→leased (v3) as lane-1 — the state the CI interleaving produced behind lane-2's back — then `factoryNextRecordAndClaim(ctx, db, root, fcRun, "t1", "lane-2")` asserted no-error + raced. No goroutines; the interleaving is pinned by construction.

Command:

```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli/ -run 'TestFactoryNextRecordAndClaimRaceOnLeasedRow' -count=1
```

Verbatim output:

```
--- FAIL: TestFactoryNextRecordAndClaimRaceOnLeasedRow (0.75s)
    factory_classify_test.go:314: RecordAndClaim on an already-leased card: illegal card transition: leased → leased
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.576s
```

Fails for exactly the CI's reason.

## Fix (minimal, `internal/cli/factory_card.go` only)

In `factoryNextRecordAndClaim` only: after the nil-error `RecordPicked` return, if `fresh.State != homestate.CardPicked`, treat it as a race — return `(homestate.Card{}, false, true, nil)` so the selection loop re-selects. Comment names the window (ListCards snapshot vs queue read; RecordPicked returns an existing row as-is when fields are empty — card_picked.go). Both callers (arm b2, arm c) are queue-picked arms where a non-picked row can only mean another lane's progress.

Not touched: `internal/homestate` (the store guard is correct), `factoryNextClaimRefused`, any other code.

## GREEN (after the fix)

Command 1:

```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli/ -run 'FactoryNext' -count=1
```

Verbatim output:

```
ok  	github.com/modu-ai/moai-adk/internal/cli	13.653s
```

Command 2 (slot lease `go-test-internal-cli` acquired 23:44:38Z, released after):

```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli/ -run 'FactoryNext' -race -count=3
```

Verbatim output:

```
ok  	github.com/modu-ai/moai-adk/internal/cli	51.962s
```

Command 3:

```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go vet ./internal/cli/
```

Verbatim output: (empty), exit 0.

## Gaps

- Full `./internal/cli` suite NOT run locally (multi-minute, slot-lease territory per lane protocol) — CI runs it. The `-race -count=3` batch covers the `TestFactoryNextParallelizableConcurrentLeases` goroutine test (the CI-flaking surface) under race.
- The other CI reds on the newer push (bce6d7e08: `TestCensus_SetDifferenceEmptyBothDirections` / `TestResolveStateAnchor_Chain`) are out of scope — different surface, another card.
- `golangci-lint` not run (scoped re-measurement per dispatch: test selector + vet only).

## Residual-risk

- The race window is narrowed to a benign re-select, not eliminated: under genuine three-way contention a re-select loop can still yield a no-card turn, which the caller treats as idle — same behavior as the documented arm-(c) no-card exit, never an error.
- The repro pins the interleaving by construction; it cannot prove the goroutine test can no longer flake on any OS, but the failure mode it exercised now maps to re-select, so the CI symptom (hard lane error) is closed by the same path the version-check races already use.

---

## Sync-gate verdict (sync-auditor, independent) — PASS 96.5/100

- Tree re-verified: `5abf390ae` @ WT-factory-next-lease-red, parent `6bcc7e4ac` confirmed; diff exactly 2 files (factory_card.go +8, factory_classify_test.go +57).
- Scores: Functionality 98 · Security 98 · Craft 92 · Consistency 96. All re-run claims re-observed on this tree (scoped selector `ok 15.235s`; new repro test `-v PASS 1.44s`; `-race -count=3 ok 46.474s`; vet clean; gofmt clean).
- Correctness probes: (a) gate masks no legitimate non-picked return (b2's `!recorded` premise + ListCards full-run coverage; arm c's operator-requeue shape converges to the documented owner re-lease path); (b) duplicate-dispatch guard tests stay GREEN — version-check race path untouched; (c) retry bound sound — worst case is the documented no-card exit, never an error or spin; (d) test pins the interleaving by construction, pre-fix code path provably triggers its Fatalf.
- Mutation probes: gate-removal mutant dies on the new test with the CI-identical message; unconditional-raced mutant dies on the two-card concurrent-lease test.
- Findings: F1 (low, no action) — the gate also absorbs operator-requeue advanced rows, a behavior narrowing aligned with the resumeTarget ownership convention; F2 (info, no action) — `!raced` style note.
- Gaps: RED confirmed analytically from the pre-fix code path + verdict.md's verbatim RED (not re-observed by revert); golangci-lint + full suite + cross-platform build remain CI's verdict (lane-local discipline).

