# t2104 progress (GitHub #1938)

## Claim
1. `moai factory runs --retire <dead> --rebind-to <live>` already moves an expired `leased` row from a dead run into a live run (the original premise "no formal path" is false).
2. The two-row state (leftover expired row in a dead run plus a live-run row) does NOT corrupt the completion gate or status counting beyond a duplicate display, but it leaves the dead run `active` and unretirable, which makes every run-less verb `AMBIGUOUS_FACTORY` (sync-audit F2, measured).
3. With two dead runs holding the card at different lease expiries, the second rebind was refused (sync-audit F1, reproduced), so the verb did not reach "one row per card".
4. Fix (one code change, RED first): in `RebindRun` the source row's lease columns are ignored in the duplicate comparison, because the lease guard has already proved the source holds no valid lease. Such a row is abandoned with a `card.rebound` pointer instead of refusing.

## Evidence (verbatim, isolated sandboxes, this run)

### F2 measurement (before the change; scratch probe, deleted)
Fixture: `run-src` dead, `t1` leased, `lease=lane-9 expired 2026-09-20T01:00:00Z`; `run-tgt` live, `t1` assigned; `card_dispatch` t1 -> run-tgt.
```
STATUS text err=<nil>
 t1 run=run-src state=leased stage=- version=3 owner=lane-9 lease=lane-9 expired 2026-09-20T01:00:00Z ...
 t1 run=run-tgt state=assigned stage=- version=1 owner=lane-1 lease=none ...
STATUS --run run-tgt err=<nil>
 t1 run=run-tgt state=assigned stage=- version=1 owner=lane-1 lease=none ...
GATE row (RecordedCardRowReadonly): run=run-tgt state=assigned owner=lane-1 linked=true err=<nil>
cards rows for t1 = 2
RETIRE plain err=factory run holds in-progress cards: run run-src holds 1: t1 (leased, lane-9)
src status=active ; rows=run-src:leased ; run-tgt:assigned
ResolveActiveRun err=AMBIGUOUS_FACTORY: run-src (owner dead, ...), run-tgt (owner live, ...)
```
Reading: the completion gate resolves through the `card_dispatch` binding to the live row (no harm); `status` lists the card twice; the dead run cannot be retired and run-less resolution is ambiguous (harm).

### F1 reproduction (before the change)
Two dead runs, `t1` leased in each (expiry 01:00 and 02:00):
```
A err=<nil>   rebound run-a into run-tgt (owner dead) / moved cards: t1
B err=factory run consolidation refused: card t1: run run-tgt already holds a leased row diverging on lease_expires_at (target evidence "", source ""); ...
rows=run-tgt:leased ; run-b:leased  a=retired b=active
```
(With identical expiries the second rebind already abandons with a pointer; only a diverging lease refused.)
Manual way out measured (operator): `decide t1 --choice abandon --run run-b` -> `1 of 1 cards refused`, row becomes `assigned`; second call -> `abandoned`; then `runs --retire run-b` -> `retired run-b (owner dead)`. Without `MOAI_FACTORY_ROLE=operator`: `refused - unknown attribution`.

### RED (tests added first, production code unchanged)
`go test ./internal/cli -run TestFactoryRunsRebind -v -count=1`:
```
--- PASS: TestFactoryRunsRebindMovesExpiredLeasedRow (2.76s)
--- FAIL: TestFactoryRunsRebindAbandonsExpiredRowDuplicatingLiveRow (2.17s)
    an expired lease alone must not block the consolidation: factory run consolidation refused: card t1: run run-tgt already holds a assigned row diverging on lease_holder, lease_expires_at ...
--- FAIL: TestFactoryRunsRebindSecondDeadRunWithDivergingExpiredLease (1.80s)
    rebind of run-b: factory run consolidation refused: card t1: run run-tgt already holds a leased row diverging on lease_expires_at ...
--- PASS: TestFactoryRunsRebindStillRefusesNonLeaseDivergence (1.83s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	11.439s
```

### GREEN (production change applied)
`go test ./internal/cli -run TestFactoryRuns -v -count=1` (11 tests, includes the 4 new and the pre-existing `TestFactoryRunsRetireRebindTo`, `...RetireRefusesInProgressCards`, `...RetireReportsStrandedCardsOfRetiredRun`):
```
--- PASS: TestFactoryRunsRebindMovesExpiredLeasedRow (3.71s)
--- PASS: TestFactoryRunsRebindAbandonsExpiredRowDuplicatingLiveRow (3.92s)
--- PASS: TestFactoryRunsRebindSecondDeadRunWithDivergingExpiredLease (3.81s)
--- PASS: TestFactoryRunsRebindStillRefusesNonLeaseDivergence (2.52s)
--- PASS: TestFactoryRunsRetireRebindTo (21.45s)
ok  	github.com/modu-ai/moai-adk/internal/cli	53.791s
```
`go test ./internal/homestate -run Rebind -count=1` -> `ok ... 137.147s`; `go test ./internal/homestate -run T1973 -count=1` -> `ok ... 244.362s` (all existing rebind and t1973 round tests, including the whole-row transition matrix, stay green).

## Baseline-attribution
Tree: worktree t2104, base HEAD 69b102ccd4 plus this change's working-tree edits (`internal/homestate/factory_run_rebind.go`, `internal/cli/factory_runs_rebind_expired_t2104_test.go`, the doc, this file). Go toolchain: the session's default `go`. No moai binary was used, so tool provenance (2.2) does not apply. Measurements ran in this run against this tree.

## Gaps
- Focused runs only: `internal/cli` `TestFactoryRuns*` and `internal/homestate` `-run Rebind` / `-run T1973`; no whole-suite run, no Windows/CI, no coverage measurement.
- `factory complete` itself was not executed end to end (it needs a git/integration fixture); the completion gate's row resolution was exercised through `RecordedCardRowReadonly`, the function the gate's row resolution uses, and `complete`'s own run resolution through `ResolveActiveRun`.
- The audit's F3 (doc names `resolveFactoryCardRun` and `ReassignCard`) is addressed in the doc; F4 (earlier probe output not kept) is replaced by the verbatim output above; F5 (card_dispatch repoint and `run.retired` events) is covered by pre-existing consolidation tests and was not re-confirmed here, except that the new test asserts one `card.rebound` pointer event.
- Worktree-guard refusals this run: compound commands containing `|` regexes or `sed -i` were refused; they were re-run as plain separate commands with the same effect.
- Codex/GLM/audit_multi tools were not used (forbidden for this task).

## Residual-risk
- Behavior change: an assign-first row used to refuse; it now abandons the source row with a pointer. The abandoned row's lease columns are not copied into the event payload (state, version, stage, evidence_sha are), so an expired lease holder name exists only on the abandoned row itself. It is dead data by the lease guard, and nothing is deleted.
- A source row with a valid (unexpired) lease still refuses the whole operation (unchanged, pinned by pre-existing tests).
- A dead run holding a card whose divergence is on any non-lease column still refuses and needs the operator's `decide` route.

## Repair round 3 (re-audit FAIL 0.85 on 8792a156d4: N1, N2; N3 note; G1 environmental)

### Decision record (N1) — the lane's judgment
Option (b): the relaxation must not apply on the automatic reconcile path (`ResolveActiveRun` -> `ReconcileActiveRuns` -> `consolidateIntoLive`). A silent abandon of a card row as a side effect of resolving the run is not an operator decision, and the change was not documented as reaching that path. Implemented as an opt-in field `RebindRequest.AbandonExpiredLeaseDuplicate`, set true ONLY by `runRebindRetire` (`moai factory runs --retire <dead> --rebind-to <live>`, `internal/cli/factory_handoff_recover.go`). The lease columns are ignored in the divergence comparison only when it is true; false restores the pre-change refusal exactly. Other `RebindRun` callers (join chooser `factory_run_select.go`, leader adoption `factory_run_adopt.go`, `consolidateIntoLive`) leave it false. Alternative (a) was rejected: it widens an unattended write path beyond what card #1938 asks (an operator-verb route).

### N3 note
An unparseable `lease_expires_at` counts as expired (`holdsValidLease` -> `LeaseExpired` contract), so under the opt-in such a row is abandoned like an expired one. The write path always records RFC3339Nano; only corrupted data reaches this. Stated in `.moai/docs/factory-expired-lease-dead-run.md`.

### RED (N1) — new test against the 8792a156d4 production code (field absent, relaxation unconditional)
`go test ./internal/cli -run 'TestFactoryAutomaticReconcileKeepsRefusing|TestFactoryRunsRebindValidLeaseSource' -count=1`:
```
--- FAIL: TestFactoryAutomaticReconcileKeepsRefusingExpiredLeaseDuplicate (0.89s)
    factory_runs_rebind_expired_t2104_test.go:37: source row state = "abandoned", want leased (not silently abandoned)
    factory_runs_rebind_expired_t2104_test.go:40: source status = "retired", want active (not silently retired)
FAIL	github.com/modu-ai/moai-adk/internal/cli	3.249s
```
(the N2 test passed there, as expected: the guard was present.)

### GREEN (N1 fix)
`go test ./internal/cli -run 'TestFactoryAutomaticReconcileKeepsRefusing|TestFactoryRunsRebind|TestFactoryRuns' -count=1 -v`:
```
--- PASS: TestFactoryAutomaticReconcileKeepsRefusingExpiredLeaseDuplicate (1.26s)
--- PASS: TestFactoryRunsRebindValidLeaseSourceWithDuplicateTargetRefuses (1.00s)
--- PASS: TestFactoryRunsRebindMovesExpiredLeasedRow (0.65s)
--- PASS: TestFactoryRunsRebindAbandonsExpiredRowDuplicatingLiveRow (0.84s)
--- PASS: TestFactoryRunsRebindSecondDeadRunWithDivergingExpiredLease (0.99s)
--- PASS: TestFactoryRunsRebindStillRefusesNonLeaseDivergence (0.49s)
--- PASS: TestFactoryRunsRetireRebindTo (5.64s)
ok  	github.com/modu-ai/moai-adk/internal/cli	17.016s
```
`go test ./internal/homestate -run Rebind -count=1` -> `ok ... 32.788s`. `go test ./internal/homestate -run T1973 -count=1` -> `ok ... 33.091s`.

### N2 mutation (guard order)
Temporary mutation via `go test -overlay` (nothing in the tree was edited): the `if holdsValidLease(c, now) {...}` block removed from `RebindRun`. Verbatim:
```
--- FAIL: TestFactoryRunsRebindValidLeaseSourceWithDuplicateTargetRefuses (0.64s)
    factory_runs_rebind_expired_t2104_test.go:61: consolidation accepted a source row with a valid lease:
        rebound run-src into run-tgt (owner dead)
        moved cards: none
        abandoned cards: t1
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.497s
```
Unmutated: PASS (above).

### Baseline-attribution (round 3)
Tree: worktree t2104, HEAD d112459e63 plus this round's edits; `go` session default; a dedicated GOCACHE under the scratchpad (deleted after). No moai binary used (no 2.2 provenance).

### Gaps (round 3)
- G1 (codex gate unmet by a stale MCP server) is environmental and was not touched; no codex/audit tool was used.
- Focused runs only; no whole-suite, CI, Windows, or coverage runs. The join chooser and leader adoption callers keep the old (refusing) behavior and have no new test of their own beyond the existing suites.
- The RED run exercised the automatic path through `factorymsg.ResolveActiveRun`; the other run-less verbs (E7 list) share that function and were not driven individually.
- One earlier edit of the production file silently missed its second hunk; the first GREEN run failed on it and was fixed before the results above.

### Residual-risk (round 3)
- Under the opt-in, a lane whose lease expired but who is still alive on another run can have its source row abandoned by the operator verb; the abandoned row keeps its lease columns and nothing is deleted.
- The explicit verb's relaxation is only as strong as the valid-lease guard; N2's test now pins its order.
