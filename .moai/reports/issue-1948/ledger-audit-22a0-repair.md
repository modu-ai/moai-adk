# Ledger repair after audit FAIL 22a0afd247 (N14, N15, N16, N17)

Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/af-core-claim-runner`, branch `WT-agentic-factory-core-claim-runner`, start HEAD `50e3371e7e`. Code and test HEAD at verification: `1f495197fb`. Scope: `internal/agenticfactory/ledger/**` only. Schema version stays 2 (`SchemaVersion = 2`, no migration change). Local commits, nothing pushed.

## Claim

1. N17: the rework check no longer reads the `actor` column. A caller passing actor `enqueue` (or `claim`, `reap`, `add`) on any move cannot switch it off.
2. N15: the rework mark starts at the FIRST offer of the landed commit, so offering the same commit again after another candidate came and went (the stepping stone) does not wash it out.
3. N14 (decision D5): a gate-only candidate is still accepted by `Enqueue` (no green required) and, once landed, finishes the card without another green; also with the holder gone.
4. N16: the newest run of each kind (green, gate) about the landed commit decides whenever it was recorded; a kind never run is not a failure; at least one passing run must still follow the move back.
5. Earlier permanent tests (done_after_rework, rework_receipt, rework_migration, rework_newest_run, done_commit) stay green; the N2 shapes stay refused.

## Evidence

Commits (RED = test only, fails on its parent tree; GREEN = code only):

| N | RED | GREEN | Tests that fail on RED (observed before committing) |
|---|---|---|---|
| N17 | `93eb8e3c37` | `3f8252a74f` | `TestRework_TheActorOfAMoveDoesNotSwitchTheReworkCheckOff`: 3 subtests (actor `enqueue` on the move back into BUILDING, inside BUILDING, on the hand-off) |
| N15 | `0ed947f5ab` | `dc0fb67602` | `TestRework_OfferingTheLandedCommitAgainDoesNotWashOutTheReworkMark`: 2 subtests (c2 failed / withdrawn, gate only); `TestRework_AFailedCandidateOfferedAgainAfterTheMoveBackNeedsNewEvidence`: 1 subtest |
| N14 | `4a7ee511c3` | `d2daa9fcd5` | `TestDone_AGateOnlyCandidateThatLandedFinishesTheCardWithoutAnotherGreen`: 3 subtests (withdrawn c1 then gate-only c2; holder gone; offered over landed c1 after rework), all `E12 needs merged` |
| N16 | `3a9ef57f32` | `1f495197fb` | `TestMove_DoneReadsTheNewestRunOfEachKindWhateverItsPlaceAroundTheMoveBack`: 12 subtests (3 cases x 2 paths x restart) |

New tests: 6 top-level, 73 subtests (files `rework_actor_test.go`, `rework_stepping_stone_test.go`, `done_gate_only_candidate_test.go`, `rework_newest_kind_test.go`). The guards and controls in those files (N2 shapes `c1 landed, then a new commit measured and never offered` and `through BUILDING`, a green on a third commit after the gate-only landing, c1 measured again, a kind never run) pass on both RED and GREEN by design.

Code change, all in `internal/agenticfactory/ledger/queue.go`:

- `offerOf(card, candidate)`: first transition into MERGING whose reason is Enqueue's reason for that candidate; falls back to the latest entry into MERGING (`to_stage=MERGING AND from_stage<>MERGING`) when no transition names it (legacy ledgers; the v1 migration fixture still passes). No read of `actor`. `reworkMark` and `newestWorkIs` use it; `Enqueue` asks about the landed row's candidate.
- `newestWorkIs`: a newest green about another commit is newer work only if its evidence id is greater than the offer's `evidence_mark`.
- `landingStands`: for each kind the newest result about the candidate (whole history) must not be a failure, plus at least one pass with id greater than the rework mark.
- The schema-2 migration (`migrateEvidenceMark`) computes marks from timestamps and never read `actor`; unchanged.

Mutation check (key condition reverted in `queue.go`, original restored byte for byte afterwards, script in the session scratchpad):

| Mutation | Result |
|---|---|
| N17: offer = latest `actor='enqueue'` transition (the old selection) | `go test` exit 1, 11 `--- FAIL` lines: 8 subtests (actor test 3, stepping stone 4, failed-then-offered 1) plus 3 parent tests |
| N15: first offer -> latest offer (`ORDER BY seq DESC`) | exit 1, 7 `--- FAIL` lines: 5 subtests (stepping stone 4, failed-then-offered 1) plus 2 parent tests |
| N14: `newestWorkIs` returns false when the newest green is about another commit (the old rule) | exit 1, 4 `--- FAIL` lines: 3 subtests plus the parent test |
| N16: per-kind newest run read only after the mark (the old window) | exit 1, 13 `--- FAIL` lines: 12 subtests plus the parent test |

The N15 mutation fails 4 stepping-stone subtests where the RED run failed 2; which extra subtests it fails was not itemized.

Final runs, verbatim:

```
$ go test -race -count=1 ./internal/agenticfactory/ledger/
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/ledger	15.054s
$ go test -race -count=1 -cover ./internal/agenticfactory/core/...
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/core	35.206s	coverage: 96.1% of statements
$ go test -race -count=1 -cover ./internal/agenticfactory/ledger/importer/
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/ledger/importer	1.752s	coverage: 89.6% of statements
$ go test -count=1 -cover -v ./internal/agenticfactory/ledger/     (counted from the saved output)
PASS lines at all levels 327, top-level 134, FAIL 0, SKIP 0; coverage: 85.1% of statements
$ go vet ./internal/agenticfactory/ledger/...      exit=0
$ gofmt -l internal/agenticfactory/ledger/         (no output)
$ golangci-lint run ./internal/agenticfactory/ledger/...    0 issues.
```

Earlier permanent tests in that verbose run (top-level PASS, subtests PASS): `TestMove_DoneNeedsWorkRecordedAfterTheMoveBack` (1, 12), `TestMove_DoneRefusedWhenTheNewestRemeasurementOfTheLandedCommitFailed` (1, 72), `TestOpen_MigratesALedgerWrittenBeforeEvidenceMarks` (1), `TestMove_DoneRefusedForALandedCardBackInBuilding` (1, 2), `TestMove_DoneNeedsTheLandedCandidateToBeTheNewestWork` (1, 5), `TestRework_ANewCommitGoesThroughTheQueueLandsAndFinishes` (1), `TestMove_DoneRefusedWhenTheRowLandedAfterTheCardWentBackForRework` (1, 2).

## Baseline-attribution

Every output above was produced in this run, from source in this tree (`go test` builds the tree; Go go1.26.8, golangci-lint v2.1.6 built with go1.26.8, installed binary, no tree-built lint). Tree HEAD at the test and lint runs: `1f495197fb`, uncommitted: this report and the progress file only. Each RED result was observed on the source tree of its parent commit (test-only commit, source unchanged). The audit's repro test files live in `mergeq` of the integration tree and were NOT run here; their shapes are re-expressed at ledger level (`TestZ2Codex_P2_*` -> `done_gate_only_candidate_test.go`, `TestZ2Done_SteppingStone_*` -> `rework_stepping_stone_test.go`, `TestZ2Done_FailingGreenThenGateAfterMark` -> `rework_newest_kind_test.go`, `TestZ2Done_ActorSpoof` -> `rework_actor_test.go`).

## Gaps

- The mergeq/CLI repros (`TestZ2Codex_P2_*`, `TestZ2Done_*`, `TestA1Done_*`, `TestZ2CLI_ActorSpoof`) were not run: they live in `mergeq/**` and `internal/cli/**`, outside this write scope. Lane-2 runs them after merge.
- N3 (`TestA1Done_WithdrawnLanded_ReclaimedStrands`) is not in this tree and was not observed. No change was aimed at it; a claim that it is unchanged is not made.
- Refused commands: the worktree guard refused one compound command (a `sed -i ... && git add ... && git commit` line) and one `git -C .` form. Both were rerun as single plain commands or through the Edit tool; no check was skipped because of them.
- Baseline total test count before this work was not measured; only the new tests are counted.
- The race run of the whole `agenticfactory` tree and the `GOOS=windows` build were not run (not requested for this lane).

## Residual-risk

- Behavior change by design (N15): a commit that FAILED in the queue, after which the card went back into BUILDING, and is offered again unchanged, now needs a passing re-measurement after the move back before the card can finish (covered by `TestRework_AFailedCandidateOfferedAgainAfterTheMoveBackNeedsNewEvidence`). One extra gate or green run for a flaky-retry flow.
- Loosening by design (N14/D5): an older green about another commit no longer blocks the landing of the commit offered after it. A holder that builds c2 (green) and then deliberately offers an older gated c1 as its last offer can finish on c1; the record shows the offer order, so the choice is visible, but the ledger treats the last offer as the work the card stands on.
- The first-offer lookup matches Enqueue's reason text; a caller can write the same reason on a move in place inside MERGING. That can only place the anchor earlier than the real offer, which makes the check stricter, never looser. A move into MERGING from another stage is impossible through `Move`.
- A landed candidate whose card never went back into BUILDING is not checked for failing newest runs (unchanged: the N16 rule applies only after a move back). Evidence of a failure recorded while the card stays in REVIEWING does not block DONE on the landed row.
- Observation E6 (a trunk commit offered as the candidate) stays out of scope, as D7 states.
