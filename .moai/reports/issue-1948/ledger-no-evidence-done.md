# Ledger: a landed card back in BUILDING no longer reaches DONE on the landed row alone (issue-1948)

Lane-18, worktree `af-core-claim-runner`, base `96fb6d5c7a`. RED `1dba89b0aa`, GREEN `120a3cf4b0`. Local only, nothing pushed. Scope: the third item of final audit a1ec370195 N2 (the "Not fixed" entry of `ledger-n2-n3-fixes.md`). N3 and N4-N10/F8-F12 are untouched.

## Claim
`factsTx` (`internal/agenticfactory/ledger/queue.go`) withholds `merged`, `closed` and `settled` while the card is in the BUILDING stage. A card whose candidate landed and which was then moved back to `BUILDING/IMPLEMENT` can no longer finish through the `no-handoff-finished` composition on the landed row alone. It leaves BUILDING by the receipt path. A card that stays in REVIEWING and finishes on its landed candidate, and the gate-only no-green card, still finish. `core/move.go` is unchanged.

Why this signal: after the landing the ledger has no record that the rework happened (no new evidence), so "newest work is a different commit" (the earlier N2 fix) has nothing to compare. The position is the one fact the ledger does hold: a landed row of the current epoch plus a BUILDING position can only come from a card taken back out of MERGING/REVIEWING after the landing, because the stage guard (`ledger-queue-stage-guard.md`) keeps a row from landing for a card outside MERGING except in the swap window.

## Evidence
- RED, `go test -count=1 -run '^TestMove_DoneRefusedForALandedCardBackInBuilding$' -v ./internal/agenticfactory/ledger` at `1dba89b0aa`:
```
    done_after_rework_test.go:23: DONE from BUILDING/IMPLEMENT on a landed row alone: <nil>
    --- FAIL: TestMove_DoneRefusedForALandedCardBackInBuilding/BUILDING/IMPLEMENT_with_no_new_evidence (0.01s)
    --- PASS: TestMove_DoneRefusedForALandedCardBackInBuilding/a_card_that_never_left_REVIEWING_still_finishes (0.01s)
FAIL
```
- GREEN, same command after the fix:
```
--- PASS: TestMove_DoneRefusedForALandedCardBackInBuilding (0.01s)
    --- PASS: TestMove_DoneRefusedForALandedCardBackInBuilding/BUILDING/IMPLEMENT_with_no_new_evidence (0.01s)
    --- PASS: TestMove_DoneRefusedForALandedCardBackInBuilding/a_card_that_never_left_REVIEWING_still_finishes (0.01s)
```
- `go test -race -count=1 -cover ./internal/agenticfactory/ledger/... ./internal/agenticfactory/core/...`:
```
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/ledger	6.298s	coverage: 85.7% of statements
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/ledger/importer	1.566s	coverage: 89.6% of statements
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/core	35.119s	coverage: 96.1% of statements
```
- `go test -race -count=1 -cover ./internal/agenticfactory/mergeq/...`: `ok ... mergeq 19.076s coverage: 85.6% of statements`.
- `go test -count=1 -v -run '^TestLand_RealLedger_' ./internal/agenticfactory/mergeq`: 12 `--- PASS` lines, 0 FAIL, including `TestLand_RealLedger_ALandedCardGoesToDoneAndTakesNoSecondCandidate (0.32s)`.
- The audit's repro file (`final-code-audit-a1ec370195/repro/mergeq_zz_audit_a1ec_test.go.txt`) was copied into `internal/agenticfactory/mergeq/` for one run and deleted afterwards (it is not in any commit). One line was added to the temporary copy of `TestA1Done_WithdrawnLanded_ThroughBuilding`: `t.Errorf` when `err2 == nil` or the stage is DONE. `go test -count=1 -v -run '^TestA1Done_' ./internal/agenticfactory/mergeq`:
```
zz_audit_a1ec_test.go:283: REVIEWING -> DONE: <nil> stage=DONE
--- PASS: TestA1Done_WithdrawnLanded_Direct (0.29s)
zz_audit_a1ec_test.go:314: BUILDING/IMPLEMENT -> DONE: agenticfactory: a transition guard has no evidence: no-handoff-finished needs merged, closed, settled stage=BUILDING
--- PASS: TestA1Done_WithdrawnLanded_ThroughBuilding (0.29s)
zz_audit_a1ec_test.go:344: FINDING: work in the trunk, row landed, card REVIEWING: no path to DONE at the new epoch
--- FAIL: TestA1Done_WithdrawnLanded_ReclaimedStrands (0.37s)
--- PASS: TestA1Done_WithdrawnLanded_NewCommitNeverLands (0.49s)
--- PASS: TestA1Done_LandedThenMovedBack_NoWindow (0.39s)
--- PASS: TestA1Done_Controls (0.51s)
```
  `ReclaimedStrands` is N3 (explicitly out of scope); its FINDING line is the same text the earlier report described.
- Runner local path and the rest: `go test -p 4 -count=1 ./internal/agenticfactory/...` -> 21 packages `ok` (runner `ok 11.416s`, mergeq `ok 21.099s`, ledger `ok 1.846s`).
- `golangci-lint run ./internal/agenticfactory/ledger/... ./internal/agenticfactory/core/... ./internal/agenticfactory/mergeq/...` -> `0 issues.` `gofmt -l internal/agenticfactory` -> no output.
- Mutation: with the condition `&& card.Position.Stage != core.StageBuilding` removed from `queue.go`, the same test fails again (`DONE from BUILDING/IMPLEMENT on a landed row alone: <nil>`, `--- FAIL ... /BUILDING/IMPLEMENT_with_no_new_evidence`); restored, the test and the earlier `TestMove_DoneNeedsTheLandedCandidate...` / `TestMove_DoneStillWorks...` tests pass (`git diff --stat`: queue.go 4 insertions, 1 deletion).

## Baseline-attribution
All commands ran in this run, in this worktree (`WT-agentic-factory-core-claim-runner`), against HEAD `96fb6d5c7a` plus the working-tree change that became GREEN `120a3cf4b0`. Tools: `go`/`golangci-lint`/`gofmt` as installed on this machine, not built from the tree (they are the language toolchain, not the project's own tooling). Coverage figures are from the commands above, not carried over.

## Gaps
- No permanent mergeq-level assertion was added: the asserting version of `TestA1Done_WithdrawnLanded_ThroughBuilding` exists only in the temporary copy that was deleted. The persistent assertion is the ledger test.
- Not run: the full `./internal/cli` suite and any independent audit (excluded by the task). The earlier report's `go test ./internal/cli -run 'TestAf|...'` was not repeated.
- The three other known failing tests (Clean, LaneChoiceMergeOffer, AutoPickMirrorParity) were not run in isolation and were not touched.
- The audit's own verdict and repro were read from the primary checkout's `.moai/reports/issue-1948/final-code-audit-a1ec370195/` (they are not in this worktree); a first Bash call combining a path variable with `sed` and one `python3` heredoc/redirect compound were refused by the worktree guard and re-run as plain commands (no other fallback).

## Residual-risk
- The rule is positional, not evidential: a card in BUILDING with a landed row never finishes, even if the holder really did rebuild the same landed commit and recorded fresh evidence there. Such a card must go BUILDING -> REVIEWING (receipt path) and finish from REVIEWING. I did not exercise that receipt path end to end in this task; the control only covers a card that never left REVIEWING.
- A holder that moves a landed card BUILDING -> REVIEWING with the old c1 receipt and then finishes still reaches DONE on c1 (the REVIEWING path of the earlier N2 note: nothing says its work moved on). Closing that needs a signal the ledger does not hold.
- A BUILDING/SPEC card with a landed row is now also refused DONE; no existing test or runner path produces that combination (all of `./internal/agenticfactory/...` passes), but a future flow that does would need the receipt path.
