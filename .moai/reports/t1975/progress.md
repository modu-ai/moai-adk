# t1975 progress (Class B, lane-23, run to2bbk, 2026-11-06)

Card: internal/homestate code findings from a static codex review of tree 25dfc0ed8c (rcpt-989d15d585e46254d84c198c, tests not run there): [P1] the merged-run refusal runs after the queue's current-dispatch write, so a refused `factory assign <card> --run <source>` leaves the queue naming the discarded run; [P2] a run merge moves only the factory DB binding, the queue's Runtime.DispatchCurrent stays on the source, so `factory assign --reassign --run <target>` is refused in factoryAssignReassign. Separate from t1973. Instruction: reproduce first (test first); if not reproduced, evidence and a close proposal.

Tree: /Users/goos/.moai/worktrees/projects/moai-adk-go/t1975, branch WT-internal-homestate-codex, base 9b632c1eb3 (= local main). Lease: `moai factory next --card t1975 --run to2bbk` exit 0.

## Current-code reading (not yet a measurement)
- The cited line moved: on this tree the retired-run refusal is `refuseIfRetiredByConsolidation` called from `upsertDispatchBindingTx` (internal/homestate/leader_approval.go:295), reached through `db.RecordDispatchBinding`.
- internal/cli/factory_card.go `factory assign`: the state-preserving re-bind branch (~4361-4368) and the `--to`-less branch (~4390-4396) call `l.RefreshDispatchCurrent(...)` first, then `db.RecordDispatchBinding(...)`. The comment at ~4361 says that branch's factory write "has no guard that can refuse after it"; the retired-run check is exactly such a guard. The `--to` branch writes the queue in `BeforeCommit`, inside the DB transaction, after the guard, so it is the safe order.
- `RebindRun` leaves the source row as `abandoned` (`abandonWithPointer`, nothing deleted), so after a merge `factory assign <card> --run <source>` takes the re-bind branch (a non-picked, non-legacy row) instead of failing earlier in `RecordPicked` (which refuses a retired run only when it must CREATE a row, card_picked.go:173).
- `RefreshDispatchCurrent` is a no-op unless a dispatch row for the card already exists (internal/factory/todo_dispatch_current.go:43-52), so a fixture must seed one first (`factory.RecordFactoryCardAssignment`, as `seedOlderDispatch` does).

## Reproduction test, attempt 1 (uncommitted): TOOL_FAILURE, not RED
File `internal/cli/factory_assign_consolidated_t1975_test.go`, test `TestRefusedAssignToAConsolidatedRunLeavesBothStoresUnchanged`. `go vet ./internal/cli` exit 0.
Command: `go test ./internal/cli -run 'TestRefusedAssignToAConsolidatedRun' -count=1 -v`, exit 1, verbatim: `factory_assign_consolidated_t1975_test.go:97: fixture: the queue names "" for t1, want run-tgt` / `--- FAIL: TestRefusedAssignToAConsolidatedRunLeavesBothStoresUnchanged (0.72s)`.
Classification (tdd-result-contract): TOOL_FAILURE — the failure is the fixture precondition, before the behaviour under test ran; it does NOT count as EXPECTED_RED. Cause: the fixture called `RefreshDispatchCurrent` on a queue with no dispatch row for t1 (no-op). Fix to apply: seed the row first with `factory.RecordFactoryCardAssignment(root, "run-src", "t1", "lane-1", "")`, then point it at the survivor with `RefreshDispatchCurrent("t1","run-tgt","lane-1")`, and re-run.

## Reproduction results (this tree, base 9b632c1eb3)
- Fixture fixed (seed the queue's dispatch row first). Tests live in `internal/cli/factory_assign_consolidated_t1975_test.go`.
- **P1 NOT reproduced.** `go test ./internal/cli -run TestRefusedAssignToAConsolidatedRun` -> PASS. `factory assign t1 --run run-src` against a consolidation-retired source is refused at run resolution (`resolveFactoryCardRun` -> `factoryRunRowNotActive`, factory_card.go ~3949-3987) BEFORE the queue lock and before any queue write; verbatim refusal `factory assign: run "run-src" is not the currently active factory run; pass the active run or omit --run`; the queue record stayed on the survivor and the factory DB dump was byte-identical. The reviewer read leader_approval.go:295 in isolation; the earlier active-run check was not in its view. The queue-first order inside the re-bind and `--to`-less branches (factory_card.go ~4361-4396) is real, so a race that retires the run between the active-run check and the write could leave the queue on the retired run; no deterministic reproduction exists without a seam, so it is recorded as a theoretical residual, not fixed (no reproduction -> no repair).
- **P2 REPRODUCED, EXPECTED_RED** (commit 890a05889b, test-only, lands before the fix). `TestReassignInTheSurvivorRunAfterAConsolidation`, fixture precondition passed (queue names run-src before the reassign), then verbatim: `reassign --run run-src (the remedy the guard names): err=factory assign: run "run-src" is not the currently active factory run; pass the active run or omit --run` and `reassign in the survivor run was refused: factory assign: card t1's current dispatch is run run-src, not run-tgt; reassign in the card's current run (pass --run run-src)`. Each refusal points at the other: the leader cannot reassign a consolidated card. (Attempt 1 of the fixture was a TOOL_FAILURE, not RED: `fixture: the queue names "" for t1, want run-tgt`; attempt 2 reached P1 and passed.)

## Fix (commit dc5e6c5bcdde6e971f547563da2d84b541a65a71, 2 files)
- `factoryAssignReassign` (internal/cli/factory_card.go): the stale-run guard skips an entry whose run was consolidated INTO the run being reassigned in (`factoryRunConsolidatedInto` / pure `consolidatedInto`, follows the recorded consolidation chain, 16-hop limit, fail-closed on a failing read). The reassign then rewrites the record to the survivor. Every other mismatch still refuses (t1696 purpose kept).
- Smaller alternative not taken: updating the queue record at merge time from the consolidation caller would not repair cards already stranded; the guard change also covers them. Doing both is a possible follow-up, not needed here.

## Evidence (this run, this tree, after the fix; exit codes read, not inferred)
- `go test ./internal/cli -run 'TestRefusedAssignToAConsolidatedRun|TestReassignInTheSurvivorRunAfterAConsolidation|TestConsolidatedIntoFollowsTheRecordedChain|TestReassignStillRefusesAnUnrelatedStaleRun|TestFactoryAssignReassign' -count=1 -v` exit 0: PASS x6 (P2 test now GREEN, P1 pin PASS, chain table PASS, unrelated stale run still refuses PASS, the two t1696 reassign tests PASS).
- `go test ./internal/cli -run 'FactoryAssign|Reassign|DispatchMirror|Consolidat|CurrentDispatch|LeaderApproval' -count=1` exit 0 (`ok ... 83.017s`).
- `gofmt -l` (both files) empty exit 0; `go vet ./internal/cli` exit 0; `golangci-lint run ./internal/cli/` exit 0, `0 issues.`

## Gaps
- Full `internal/cli` suite and other packages not run (lane-local verification; the suite is the CI/leader's). `internal/homestate` untouched.
- The P1 race window is not exercised (no seam).
- No Windows/Linux run.
- A compound shell command (heredoc append) was refused by the worktree guard once; the same edit was made with the Edit tool; no verification rested on the refused command.

## Factory transitions (2026-11-06, lane-23, run to2bbk)
- `stage t1975 run --run to2bbk` -> `t1975 run v9 lease renewed`; `stage t1975 sync dc5e6c5bcdde6e971f547563da2d84b541a65a71 --run to2bbk` -> `t1975 sync v10 lease renewed`.
- Independent Claude sync-audit (no Codex): `.moai/reports/t1975/sync-audit-iter1.md`, PASS-WITH-DEBT 0.91, blocking 0, audited_sha dc5e6c5bcd; committed with `git add -f` as 8dc57e6862 (evidence-only commit). Optional debt D1 (no failure-injection test for the read-error branch, mutant M8 survives), D2 (stale comment factory_card.go:4361-4363), D3 (P1 queue-first residual).
- LANE ERROR: I entered the sync-audit edge with the verdict file as evidence: `stage t1975 sync-audit dc5e6c5bcd...:.moai/reports/t1975/sync-audit-iter1.md` was refused (`card evidence refused: artifact .moai/reports/t1975/sync-audit-iter1.md is absent at commit dc5e6c5bcdde6e971f547563da2d84b541a65a71.`); I then re-ran it with 8dc57e6862...:<verdict> -> `t1975 sync-audit v11 lease renewed`. That recorded 8dc57e6862 as the evidence commit. The t2073 pattern is code commit as the sync-audit entry evidence (`<code sha>:<code file>`), verdict commit only at merge-ready.
- `stage t1975 merge-ready --run to2bbk` and again with `8dc57e6862...:.moai/reports/t1975/sync-audit-iter1.md`: both refused, verbatim: `card evidence refused: verdict file .../t1975/.moai/reports/t1975/sync-audit-iter1.md audited dc5e6c5bcdde6e971f547563da2d84b541a65a71, not the recorded evidence commit "8dc57e6862313abc9c07c77444d53fbd602b1ff9".`
- Repair attempt `stage t1975 sync-audit dc5e6c5bcd...:internal/cli/factory_card.go --run to2bbk`: refused, verbatim `Factory stage: illegal card transition: sync-audit → sync-audit.` Not worked around (no queue or DB edits, no back-edge tried).
- State now: sync-audit v11, HEAD 8dc57e6862 (+ this progress edit), not pushed. Waiting on the leader for the repair path (a legal edge to re-record the evidence, or a verb).

## Gaps (audit gate)
- The stop hook refused the verdict without a Codex receipt (`AUDIT_RECEIPT_VIOLATION ... no receipt cited`); no Codex call was made for this audit. Operator rule d-20261106T061529Z-5d7e allows up to 2 per card by the auditor itself; not used. The verdict is the auditor's judgment on the merits, not a gate-corroborated PASS.
- `card-review.md` (codex_review scope=card) was not produced: no Codex call approved for the lane's own review in this run; the leader self-review applies.

## Next
Waiting for the leader's repair instruction for the sync-audit evidence record; then merge-ready; stop there (complete is paused until t2098 lands).
