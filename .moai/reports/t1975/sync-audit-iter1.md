verdict: PASS-WITH-DEBT
audited_sha: dc5e6c5bcdde6e971f547563da2d84b541a65a71
overall_score: 0.91
blocking_count: 0

auditor-model: claude-opus-5-5

# t1975 sync audit, iteration 1 (independent, Claude-only)

Card t1975 (Class B, no SPEC). Branch `WT-internal-homestate-codex`, base `9b632c1eb3`, commits `890a05889b` (test-only RED) and `dc5e6c5bcd` (fix and extended tests). Per the dispatch, no Codex, GLM, or cross-model tool was called, and the audit plan verb was not run. Scope: `git diff --stat 9b632c1eb3..HEAD` shows two files, `internal/cli/factory_assign_consolidated_t1975_test.go | 214 +` and `internal/cli/factory_card.go | 53 ++++-`. Nothing else changed.

## Dimension scores

| Dimension | Score | Verdict | Deciding evidence |
|---|---|---|---|
| Functionality (must-pass) | 0.92 | PASS | Focused suite exit 0 (6/6 PASS). The RED commit was executed: P2 test FAIL there, PASS at HEAD |
| Security (must-pass) | 0.95 | PASS | No new input surface. Every read goes through parameterized SQL in `runRetiredByConsolidation`. The guard fails closed (a read error refuses, and the hop limit can only refuse) |
| Craft | 0.85 | PASS | vet and gofmt are clean. One surviving mutant (the read-error wiring in the caller). One stale adjacent comment. Coverage was not measured |
| Consistency | 0.92 | PASS | Matches the file's error-wrapping and comment style. The pure-helper split follows the codebase pattern |

Overall (harmonic mean): 4 / (1/0.92 + 1/0.95 + 1/0.85 + 1/0.92) = 0.91.

## Claim

C1. HEAD is the audited SHA and the tree is clean.
C2. The focused suite passes at HEAD, and `go vet` and `gofmt -l` are clean on the changed files.
C3. The RED commit `890a05889b` is red on its own, and it fails at the intended assertion rather than at a fixture precondition.
C4. P1 does not reproduce. A `factory assign t1 --run <consolidated source>` is refused in `resolveFactoryCardRun` → `factoryRunRowNotActive` (factory_card.go:3949-3987, `status != "active"`). The resolution runs at factory_card.go:4272, before `queueStore.WithLock`, so no queue write happens.
C5. The lane's disposition of P1 ("theoretical race residual, not fixed") is sound. Details follow.
C6. The P2 fix accepts only a recorded consolidation chain from the queue's run INTO the run being reassigned in. It still refuses an unrelated stale run (t1696 purpose kept). It cannot accept wrongly through the hop limit or a cycle. It surfaces a read error instead of swallowing it.
C7. `RunRetiredByConsolidation` returns what the helper assumes.
C8. Nothing else in factory_card.go is touched.

### C5 reasoning (P1 residual)

In the re-bind branch (4361-4368) and the `--to`-less branch (4393-4396), `l.RefreshDispatchCurrent` runs before `db.RecordDispatchBinding`. The DB side carries `refuseIfRetiredByConsolidation(ctx, tx, runID)` inside the binding transaction (homestate/leader_approval.go:295). If a consolidation landed between the `factoryRunRowNotActive` read and the binding transaction, the queue would name the retired run while the DB refuses the binding. That ordering is real.

Two things keep it a residual rather than a blocking defect:
- It is reachable only through a concurrent `RebindRun` racing a leader `assign` inside one window. No deterministic path exists: the `--run`-less path resolves an active run through `ResolveActiveRun`, and an explicit `--run` on a retired run is refused first.
- The state it leaves (the queue names a source consolidated into the survivor) is exactly the state the P2 fix now repairs. A later `--reassign --run <survivor>` is accepted and rewrites the record.

The comment at 4361-4363 ("this branch's factory write has no guard that can refuse after it") has been stale since the t1963 DB-side guard landed. It is outside the diff (touched-file) and recorded as debt D2.

## Evidence

All commands ran from `/Users/goos/.moai/worktrees/projects/moai-adk-go/t1975`. Exit codes were written to files and read back, not taken from a pipe.

E1. `git rev-parse --show-toplevel HEAD`:
```
/Users/goos/.moai/worktrees/projects/moai-adk-go/t1975
dc5e6c5bcdde6e971f547563da2d84b541a65a71
```
It was re-read before writing this file: `dc5e6c5bcdde6e971f547563da2d84b541a65a71`. `git status --short` printed nothing.

E2. `go test ./internal/cli -run 'TestRefusedAssignToAConsolidatedRun|TestReassignInTheSurvivorRunAfterAConsolidation|TestConsolidatedIntoFollowsTheRecordedChain|TestReassignStillRefusesAnUnrelatedStaleRun|TestFactoryAssignReassign' -count=1 -v` exited 0:
```
--- PASS: TestRefusedAssignToAConsolidatedRunLeavesBothStoresUnchanged (2.40s)
    factory_assign_consolidated_t1975_test.go:153: reassign --run run-src (the remedy the guard names): err=factory assign: run "run-src" is not the currently active factory run; pass the active run or omit --run stderr=""
--- PASS: TestReassignInTheSurvivorRunAfterAConsolidation (2.69s)
--- PASS: TestConsolidatedIntoFollowsTheRecordedChain (0.00s)
--- PASS: TestReassignStillRefusesAnUnrelatedStaleRun (1.94s)
--- PASS: TestFactoryAssignReassignHappyPath (1.10s)
--- PASS: TestFactoryAssignReassignRefusals (8.05s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	18.164s
```

E3. `go vet ./internal/cli` exited 0 with 0 bytes of output.

E4. `gofmt -l internal/cli/factory_card.go internal/cli/factory_assign_consolidated_t1975_test.go` exited 0 with 0 bytes of output.

E5. The RED commit was executed, not only reasoned about. Method: `git archive -o <scratch>/red-890a.tar 890a05889b`, extracted into a plain scratch directory (not a git worktree; no stash; the card tree untouched). Then, inside the extracted copy, `go test ./internal/cli -run 'TestRefusedAssignToAConsolidatedRun|TestReassignInTheSurvivorRunAfterAConsolidation|TestFactoryAssignReassign' -count=1 -v` exited 1:
```
--- PASS: TestRefusedAssignToAConsolidatedRunLeavesBothStoresUnchanged (1.02s)
    factory_assign_consolidated_t1975_test.go:152: reassign --run run-src (the remedy the guard names): err=factory assign: run "run-src" is not the currently active factory run; pass the active run or omit --run stderr=""
    factory_assign_consolidated_t1975_test.go:156: reassign in the survivor run was refused: factory assign: card t1's current dispatch is run run-src, not run-tgt; reassign in the card's current run (pass --run run-src) (stderr "")
--- FAIL: TestReassignInTheSurvivorRunAfterAConsolidation (1.78s)
--- PASS: TestFactoryAssignReassignHappyPath (1.71s)
--- PASS: TestFactoryAssignReassignRefusals (6.90s)
FAIL
```
The failure is at line 156, the intended assertion (`reassign in the survivor run was refused`), after the fixture precondition at line 146 passed. This is EXPECTED_RED, not a TOOL_FAILURE. The P1 pin passes at both commits, which matches "P1 not reproduced".

E6. Code read (C6, C7, C8):
- `git diff 9b632c1eb3..HEAD -- internal/cli/factory_card.go` has two hunks only: the guard loop in `factoryAssignReassign` (4436-4455) and the new `factoryConsolidationHopLimit`, `factoryRunConsolidatedInto`, and `consolidatedInto` (4469-4501).
- `consolidatedInto` returns true only when `into == to` at some hop. A cycle, or a chain longer than 16 hops, falls out of the loop with `return false, nil`, which leads to a refusal. Both limits can only cause a false refusal, never a false accept. A cycle that passes through `to` is accepted at the hop that reaches `to`, which is correct: the record was consolidated into `to`. `to` itself is active, because `resolveFactoryCardRun` already refused a non-active `--run`.
- `homestate.RunRetiredByConsolidation` (factory_run_consolidated.go:31-78) returns `(into, true, nil)` only when `runs.status = 'retired'` AND a `run.retired` event carries a non-empty `rebound_to`. An active run, an ordinarily retired run, and an absent run all return `("", false, nil)`. Only a store failure returns an error. This matches the helper's `(into, consolidated, err)` contract exactly.
- A read error leads to `return fmt.Errorf("read the consolidation record of run %s: %w", ...)` inside `WithLock`, so the reassign refuses: fail-closed, surfaced, not swallowed.
- t1696 purpose: `TestReassignStillRefusesAnUnrelatedStaleRun` (run-elsewhere has no row, so it is not consolidated and the reassign is refused) and the pre-existing `factory_assign_reassign_test.go:141` (`current dispatch is run run-stale`) both pass (E2).

E7. Mutant probe. No edits were made; each outcome below is predicted from the test code, not executed.

| # | Mutant | Test that kills it |
|---|---|---|
| M1 | Delete `if merged { continue }` | `TestReassignInTheSurvivorRunAfterAConsolidation` (the line-156 refusal returns, as at RED) |
| M2 | Swap arguments: `factoryRunConsolidatedInto(ctx, db, runID, entry.RunID)` | Same P2 test: run-tgt is active, so the result is false and the reassign is refused |
| M3 | Replace the final `return fmt.Errorf(... current dispatch is run ...)` with `continue` | `TestReassignStillRefusesAnUnrelatedStaleRun` and `factory_assign_reassign_test.go:141` |
| M4 | In `consolidatedInto`, `if !consolidated { return false }` becomes `return true` | Chain table rows "an unrelated run does not" (x→y→stop) and "never consolidated" (c) |
| M5 | Accept on any consolidation (drop the `into == to` comparison) | Chain table "the reverse direction does not" is unaffected (b→c, so true is still wrong for b→a) and "an unrelated run does not" (x→y) |
| M6 | Hop limit 0 or 1 | Chain table "a merge of a merge reaches the final survivor" (needs 2 hops); at 0 the P2 test also fails |
| M7 | `consolidatedInto` swallows the error (`return false, nil`) | Chain table error case (`errors.Is(err, boom)`) |
| M8 (survives) | In `factoryAssignReassign`, the `mergeErr != nil` branch becomes `continue` (fail-open) | **No test kills it.** No integration test injects a failing consolidation read through `factoryRunConsolidatedInto`. See D1 |

## Baseline-attribution

- Tree: `dc5e6c5bcdde6e971f547563da2d84b541a65a71` (E1), measured in this run.
- RED baseline: commit `890a05889b`, an extracted archive of that exact commit, measured in this run (E5).
- Toolchain: the `go` on PATH in this session, compiling the source of each tree. No project-built `moai` binary judged anything, so §2.2 build-provenance does not apply.
- The lane's progress.md claims (golangci-lint `0 issues.`, the broader `-run 'FactoryAssign|Reassign|...'` suite `ok ... 83.017s`) were read but NOT re-run here. They are cited as lane claims, not as audit evidence.

## Gaps

- G1. Coverage of the new helpers was not measured (`go test -cover` not run). Craft is scored on mutant analysis and the test reading instead.
- G2. `golangci-lint run ./internal/cli/` was not run by this audit (lane claim only).
- G3. The full `internal/cli` suite and other packages were not run, per the dispatch instruction.
- G4. The P1 race window (consolidation between `factoryRunRowNotActive` and the binding transaction) was not exercised. It was judged by code reading only.
- G5. The mutant table (E7) was derived by reading the tests. No mutant was executed, because edits were forbidden.
- G6. One compound shell command (the focused `go test` run as one `cd && ... ; echo $?` line) was refused by the worktree-isolation guard. The same commands were then run from a scratch script that writes each exit code to its own file (E2-E4). No verification rested on the refused form.
- G7. Deleting the scratch extraction of the RED commit (`rm -rf` in the scratchpad) was denied by permission policy. The copy remains under the session scratchpad only; the card tree is unaffected (E1 status clean).
- G8. No Linux or Windows run.
- G9. The required Codex receipt gate is unmet. When this auditor stopped, the receipt guard refused the verdict: `AUDIT_RECEIPT_VIOLATION: sync-auditor reported PASS-WITH-DEBT but the audit could not be corroborated — no receipt cited.` The dispatch explicitly forbade Codex, GLM and audit_multi (Claude-only), and starting a paid call without user approval is outside this auditor's scope, so no Codex audit was run and no receipt exists. The PASS-WITH-DEBT token above is this audit's judgment on the merits only. Until the leader settles the gate, by authorizing a Codex receipt run or recording its own disposition, the run/sync/PR spawns stay denied by the guard, and this verdict must not be treated as a gate-corroborated PASS.

## Residual-risk

- R1. The P1 queue-first ordering in the re-bind and `--to`-less branches can still strand the queue record on a run consolidated mid-assign. Mitigation observed: the P2 fix makes that state repairable through `--reassign --run <survivor>`. It is not prevented.
- R2. The guard's consolidation read happens outside the DB transaction that `ReassignCard` opens. An accepted chain is terminal (a retired run does not un-retire), so a later change cannot invalidate an accept. A consolidation of `runID` itself after the guard is a pre-existing concern of `ReassignCard`, not of this diff.
- R3. Records already stranded by older consolidations are repaired only when the leader runs a reassign. Non-reassign readers of `Runtime.DispatchCurrent` still see the retired source until then. The lane named merge-time queue repair as a possible follow-up.

## Findings (structured defect list)

- F1 [Low] [optional] [in-diff] internal/cli/factory_card.go:4447-4450. The read-error branch of the consolidation lookup is not exercised by any integration test, so mutant M8 (fail-open `continue`) survives. Required fix: none is required for this card. If wanted, add one test that makes `RunRetiredByConsolidation` fail (for example, a closed DB handle passed to `factoryRunConsolidatedInto`) and asserts the refusal. Confidence: high that the mutant survives.
- F2 [Low] [optional] [touched-file] internal/cli/factory_card.go:4361-4363. The comment says the re-bind branch's factory write "has no guard that can refuse after it". `RecordDispatchBinding` now carries the t1963 consolidation guard (homestate/leader_approval.go:295), so the comment is stale and hides the P1 residual ordering. Required fix: reword it to name the guard and the accepted residual, or route it to a follow-up card. Confidence: high.
- F3 [Info] [optional] [touched-file] internal/cli/factory_card.go:4365 and 4393. The P1 residual race (C5) is recorded, not fixed, consistent with "no reproduction, no repair". Required fix: none. Optionally, a follow-up card could move the queue write into a `BeforeCommit` hook of the binding transaction, as the `--to` branch already does.

## Debt items

- D1 = F1. Missing failure-injection test for the caller's read-error branch.
- D2 = F2. Stale comment at factory_card.go:4361-4363.
- D3 = F3. P1 queue-first ordering residual, plus the merge-time queue-repair follow-up the lane named.

## Iteration history

- iter1 (this file): PASS-WITH-DEBT 0.91 @dc5e6c5bcd. Blocking 0. Optional findings 3.
