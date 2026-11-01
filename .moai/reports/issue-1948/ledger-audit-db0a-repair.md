# Ledger repair after audit FAIL db0a69698a (N24, N25, N27, N32, N34 part) and the operator's blockers N3 and N28

Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/af-core-claim-runner`, branch `WT-agentic-factory-core-claim-runner`, start HEAD `850a2fa079`. Code and test HEAD at verification: `bf1da1f4b3`. Scope: `internal/agenticfactory/ledger/**` only (plus this report and the progress file). Schema version stays 2 (`SchemaVersion = 2`; the migration statement changed, the version and the columns did not). Local commits, nothing pushed. Decision D5 (a gate-only candidate is accepted at Enqueue, no green required) stands.

The second part of the dispatch (N3, N28; repair note 1-B withdrew their acceptance) was added after the first part was finished; it is reported here in its own sections and changes some statements the first version of this report made (marked "superseded").

## Claim

1. N24: a version 1 ledger migrated by this code ranks evidence recorded in the same millisecond as a registration (a transition into MERGING) AFTER it; for any other move it still ranks before.
2. N25: `Enqueue` refuses, at registration, a commit that this card already offered and was then taken back for rework, unless the same test `landingStands` applies at DONE holds. After the refusal the way out works: measure again, enqueue, land, finish with no holder.
3. N27: `Enqueue` refuses a candidate when the subject of the card's newest green, gate or review evidence is a different commit. A gate-only candidate that is itself the newest evidence is still taken (D5). Measuring the older commit again opens it.
4. N32: two tests tie the audit's surviving mutations Q1 and Q3 to the suite; both mutations are now caught.
5. N34 part: the `queue.go` comment that said Enqueue would refuse a landed commit a second time was rewritten to say what is true.
6. N3: the holder of a card claimed again (a higher card epoch) takes over the landed row, when it offers the landed commit, or replaces it, when it offers a new commit, once it has measured that commit in its own epoch. Without that measurement the offer is refused and nothing changes. `TestA1Done_WithdrawnLanded_ReclaimedStrands` and `TestZ2Strand_ReworkedThenReclaimed` are re-expressed at ledger level and end in DONE.
7. N28: `landingStands` now refuses a landed candidate (a) whose newest green or newest gate run failed, whether or not the card was ever taken back, and (b) when the card left MERGING after the offer and has recorded evidence about another commit since, unless the candidate was measured again after that evidence with a passing green AND a passing gate. `Enqueue` asks the same for a commit it already saw offered, and refuses a first offer whose newest green failed. The three audit shapes (`FailureRecordedAfterLandingNoRework`, `WithdrawnToReviewingGateOnlyNewCommit`, `ReworkGateOnlyFixThenRemeasureC1`) are refused, each with a way out.
8. The accepted-risk premise of commit `a061ffdf6b` (`accepted_risk_n28_test.go`) is withdrawn: the file was renamed to `done_results_after_landing_test.go` and its assertions turned to REFUSAL. This is a change of premise ordered by the leader (repair note 1-B), not a weakening: every shape it pinned as accepted is now asserted refused, and the file adds the recovery path for each. (Superseded: the first version of this report said N27/N28 shapes were not closed and pinned them; they are closed.)
9. Earlier rework/done tests, N14/N15/N16/N17/N25/N27 tests and D5 stay green; three pre-existing tests had their expectation moved (J1, J7, J8), each in a test commit of its own or inside a RED commit.

## Evidence

Commits (RED = tests only, fails on its parent tree; GREEN = code only):

| N | RED | GREEN | What fails on RED (observed before committing; source unchanged) |
|---|---|---|---|
| N24 | `6f032bd284` | `57f54987d2` | `TestOpen_ASameMillisecondGreenAfterTheOfferIsNewerWork/in_the_millisecond_of_the_offer` (1 subtest + parent). Controls (a millisecond after, a millisecond before, the move-back control) pass on both. |
| N25 | `3f2f6ad3a7` | `f5def92b9b` | 4 top-level, 9 subtests: `TestEnqueue_ACommitOfferedBeforeTheMoveBackNeedsARunRecordedSince` (2), `TestEnqueue_WhatCountsAsARunRecordedSinceTheMoveBack` (2), `TestRework_OfferingTheLandedCommitAgainDoesNotWashOutTheReworkMark` (4), `TestRework_AFailedCandidateOfferedAgainAfterTheMoveBackNeedsNewEvidence` (1) |
| N27 | `be99e031ab` | `a75da5ae11` | `TestEnqueue_TheCandidateMustBeWhatTheCardMeasuredLast`: 4 subtests + parent. The D5 test passes on both. A test-only fixture fix `9890afaaa3` (passing green instead of failing in one D5 case) preceded the N28 GREEN. |
| N32 | `3e7f7c6e33` (tests only; the rules already hold, the tests exist to kill the mutations) | none | passes on HEAD; fails under Q1 and Q3 |
| N3 | `c2d2139176` | `cb2922f689` | `TestEnqueue_AReclaimedCardTakesOverOrReplacesTheLandedRowOnceItMeasuredInItsOwnEpoch`: 4 subtests + parent (take-over and replacement, each for the REVIEWING and the BUILDING reclaim), all `a has landed`. The six refusal subtests pass on both by design. A test-only commit `7823406339` (between RED and GREEN) kept `TestEnqueue_NeverReplacesALandedRow` honest (J7). |
| N28 | `d9e11c3208` | `bf1da1f4b3` | 5 top-level, 7 subtests failing (12 `--- FAIL` lines): `TestMove_DoneNeedsTheLandedCandidateToBeTheNewestWork` (1 subtest, moved premise J8), `TestDone_AFailureRecordedAfterTheLandingRefusesTheCardWhetherOrNotItWasTakenBack` (3: the three failure forms still in MERGING), `TestDone_WithdrawnToReviewingThenAGateOnlyNewCommitIsRefused` (1), `TestDone_ReworkProducesAGateOnlyFixThatIsNeverOffered` (the test itself), `TestEnqueue_AsksForWhatDoneAsksForAfterTheLanding` (2). The "withdrawn to REVIEWING, X failed" subtests pass on both (a failing run already leaves no receipt there). |
| N34 part | in `f5def92b9b` | | comment only |
| lint | | `c8829c13ba` | `refactor`: `namedOfferOf` rewritten without the `switch` the linter flagged; behavior unchanged |
| superseded | `a061ffdf6b` | | the first N28 measurement; replaced by `d9e11c3208` (git rename) |

Code, all in `internal/agenticfactory/ledger`:

- `schema.go` `migrateEvidenceMark`: the correlated subquery is `e.at<transitions.at OR (e.at=transitions.at AND transitions.to_stage<>'MERGING')` (the audit's expression).
- `queue.go`:
  - `namedOfferOf`, `offerMarkSQL`, `noFailingNewestRun`, `measuredInEpoch`, `measuredAfterOtherWork` (new helpers).
  - `landingStands`: (1) newest green and newest gate about the candidate not failures (always); (2) if reworked, a passing run after the mark; (3) `measuredAfterOtherWork`.
  - `Enqueue`: a landed row of a lower card epoch is refused unless the offered commit was measured in this epoch (`measuredInEpoch`), then the same commit takes the row over (`UPDATE merge_queue SET card_epoch`, row stays landed, a decision note) and another commit replaces it (note, then the normal restart). After the row cases and before the row write: the N27 check, the failing-newest-run check, the N25/N28 check (`landingStands` for a commit with a named earlier offer). Every refusal returns before any write.

### Mutation check

`go test -overlay` over a sed-produced copy; the tree is never edited; scripts `scratchpad/mut.sh`, `scratchpad/mut2.sh`; all run at HEAD `bf1da1f4b3` (N24..Q3 re-run after the N3/N28 code); every mutation changed the file.

| Mutation | Result | Tests that failed |
|---|---|---|
| N24: migration back to `e.at<=transitions.at` | exit 1, 2 `--- FAIL` lines (1 subtest, 1 parent) | `TestOpen_ASameMillisecondGreenAfterTheOfferIsNewerWork` |
| N25: `if !stands {` to `if !stands && false {` | exit 1, 9 lines (4 parents) | `TestEnqueue_ACommitOfferedBeforeTheMoveBackNeedsARunRecordedSince`, `TestEnqueue_WhatCountsAsARunRecordedSinceTheMoveBack`, `TestRework_AFailedCandidateOfferedAgainAfterTheMoveBackNeedsNewEvidence`, `TestEnqueue_AsksForWhatDoneAsksForAfterTheLanding` |
| N27: the different-subject case unreachable | exit 1, 5 lines (4 subtests, 1 parent) | `TestEnqueue_TheCandidateMustBeWhatTheCardMeasuredLast` |
| Q1: `id <= mark` to `id < mark` | exit 1, 4 lines (2 parents) | `TestDone_AGreenWithExactlyTheIdOfTheOffersMarkIsOlderWork`, `TestOpen_ASameMillisecondGreenAfterTheOfferIsNewerWork` |
| Q3: latest move back to oldest in `reworkMark` | exit 1, 2 lines (1 parent) | `TestDone_ACardTakenBackTwiceNeedsAMeasurementAfterTheSecondMoveBack` |
| N3a: the "measured in this epoch" refusal switched off | exit 1, 5 lines (2 parents) | `TestEnqueue_AReclaimedCard...`, `TestEnqueue_ADifferentEpochDoesNotReplaceALandedRowWithACommitMeasuredBeforeIt` |
| N3b: `epoch=?` to `epoch<=?` (an earlier holder's run counts) | exit 1, 5 lines (2 parents) | same two |
| N3c: the take-over flag never set (the row is restarted as waiting) | exit 1, 3 lines (1 parent) | `TestEnqueue_AReclaimedCard...` |
| N28a: condition 1 (no failing newest run) switched off | exit 1, 42 lines (3 parents) | `TestDone_AFailureRecordedAfterTheLanding...`, `TestMove_DoneReadsTheNewestRunOfEachKind...`, `TestMove_DoneRefusedWhenTheNewestRemeasurementOfTheLandedCommitFailed` |
| N28b: condition 3 switched off | exit 1, 7 lines (4 parents) | `TestDone_ReworkProduces...`, `TestDone_WithdrawnToReviewing...`, `TestEnqueue_AsksForWhatDoneAsks...`, `TestMove_DoneNeedsTheLandedCandidateToBeTheNewestWork` |
| N28c: condition 3 asks for a gate only (no green) | exit 1, 5 lines (3 parents) | the first three of the row above |
| N28d: the first-offer failing-newest-run refusal at Enqueue switched off | exit 1, 2 lines (1 parent) | `TestEnqueue_AsksForWhatDoneAsksForAfterTheLanding` |

Notes on the table: with the N25 check off, `TestRework_OfferingTheLandedCommitAgainDoesNotWashOutTheReworkMark` still passes because the N27 check refuses the same second offer of c1 (the newest evidence is c2's); a joint mutation of both was not run. With the Q1 mutation the `id == mark` edge is reachable through the verbs only on a migrated ledger (the N24 test's "a millisecond before" case does that and also fails), since N27 refuses an offer made over a newer green; `TestDone_AGreenWithExactlyTheIdOfTheOffersMarkIsOlderWork` sets the mark with a raw `UPDATE` to state the edge directly (J6).

### Audit repro, observed (N24)

Fixtures `scratchpad/v1/*.db` of the audit session; the audit's `ledger_zz_*` files were copied into the package as `zz_*_test.go`, run once, and deleted before the next commit:

```
=== RUN   TestA2Mig_SameMillisecondGreenAfterTheOffer/same_millisecond
    zz_a2_mig_test.go:42: DONE: agenticfactory: transition not allowed from the current state: no S46 edge or named composition moves MERGING to DONE
    zz_a2_mig_test.go:44: stage=MERGING
(one_millisecond_later: same lines)
--- PASS: TestZ2Mig_RealV1 (rework_no_evidence, rework_remeasured, landed_plain), ConcurrentOpens, FailureRollsBack, OldBinaryOnV2, NullMarkWrittenAfterMigration, Fresh
```

Baseline of the audit for the same input: `DONE: <nil> stage=DONE`. The N3 and N28 repro tests of the audits live in `mergeq` / `internal/cli` and were NOT run here; their shapes are re-expressed at ledger level (below). Mapping: `TestA1Done_WithdrawnLanded_ReclaimedStrands` -> `...REVIEWING, c1 measured again takes the landing over`; `TestZ2Strand_ReworkedThenReclaimed` -> the BUILDING variants (take-over and replacement); `TestA2Done_FailureRecordedAfterLandingNoRework` -> `TestDone_AFailureRecordedAfterTheLanding...`; `TestA2Done_WithdrawnToReviewingGateOnlyNewCommit` -> `TestDone_WithdrawnToReviewingThenAGateOnlyNewCommitIsRefused`; `TestA2Done_ReworkGateOnlyFixThenRemeasureC1` -> `TestDone_ReworkProducesAGateOnlyFixThatIsNeverOffered`.

### The shapes of N3 and N28 and their exits (ledger level, all in the final run)

- N3: c1 landed at epoch 1, card withdrawn (REVIEWING) or reworked (BUILDING), holder released, new claim at epoch 2. c1 measured at epoch 2 (green+gate) -> `Enqueue(c1)` ok, row `landed`, epoch 2, stage MERGING, `finish` -> DONE. c2 measured at epoch 2 -> `Enqueue(c2)` ok, row `waiting` c2 epoch 2, the landing of c1 kept as a note, land, `finish` -> DONE. Refused and row unchanged: nothing measured at epoch 2; c2 measured only by the old holder; c1 measured at epoch 2 with a failing green.
- N28 failure after landing: c1 landed, card still in MERGING, green and/or gate of c1 recorded as failing -> `finish` refused (`ErrGuardMissing`); `evidenceOn(c1)` passing -> `finish` ok. Same for the withdrawn-to-REVIEWING card.
- N28 withdrawn to REVIEWING, c2 gated and never offered: DONE refused; after a gate of c1 alone still refused; after green+gate of c1 DONE ok. c2 offered from REVIEWING is refused (`has landed`, the row stays until the card has been in BUILDING); through BUILDING and back to REVIEWING it is accepted gate-only (D5), lands, finishes.
- N28 rework with a gate-only fix: same three steps (c2 gate only -> refused; c1 gate alone -> refused; c1 green+gate -> DONE).
- Control: a gate of c2 recorded while the card stays in MERGING and never leaves does not block finishing on the landed c1.
- Entrance: c1 offered, withdrawn, c2 gated, c1 gated alone -> `Enqueue(c1)` refused (`ErrEvidenceStale`); c1 green+gate -> accepted, lands, finishes. A first offer whose newest green failed -> refused; measured again -> accepted.

### Final runs at HEAD `bf1da1f4b3`, verbatim

```
$ go test -race -count=1 ./internal/agenticfactory/ledger/
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/ledger	28.657s
$ go test -race -count=1 -cover ./internal/agenticfactory/core/...
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/core	42.529s	coverage: 96.1% of statements
$ go test -race -count=1 -cover ./internal/agenticfactory/ledger/importer/
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/ledger/importer	1.539s	coverage: 89.6% of statements
$ go test -count=1 -cover -v ./internal/agenticfactory/ledger/     (counted from the saved output)
exit=0   PASS lines at all levels 388, top-level 148, FAIL 0, SKIP 0; coverage: 84.8% of statements
$ go test -count=1 ./internal/agenticfactory/mergeq/     (this tree's mergeq)
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq	20.342s
$ go vet ./internal/agenticfactory/ledger/...      exit=0
$ gofmt -l internal/agenticfactory/ledger/         (no output, exit 0)
$ golangci-lint run ./internal/agenticfactory/ledger/...    0 issues.   exit=0
```

The tree compiles and is green. (The `posDone redeclared` diagnostic and the one failing `TestOpen_ASameMillisecondGreenAfterTheOfferIsNewerWork` run the lane saw were the intermediate states of this session: a name collision in a new test file fixed before its commit, and the RED run of N24 observed before its GREEN. Neither is in a committed tree.) A second coverage profile of the same tree printed `total: 85.0%`; the figure moves between 84.8% and 85.0% from run to run. Earlier permanent tests in the verbose run all pass; the ones named in the previous report keep their top-level and subtest counts.

## Baseline-attribution

Every output above was produced in this run from source in this tree (`go test` builds the tree; golangci-lint is the installed binary, not built from the tree). Tree HEAD at the final runs: `bf1da1f4b3`; uncommitted: this report and `18-progress.md` only. Each RED result was observed on the source tree of its parent commit (test-only commits; source unchanged between the observation and the commit). The mergeq line was taken at `bf1da1f4b3` in this tree, whose mergeq is not the integration tree's.

## Judgments recorded where the rules pulled apart

- J1: N25 says the N14..N17/N15 tests must stay green and also says `Enqueue` must refuse an unmeasured re-offer; both cannot hold for the stepping-stone tests, which expected the second offer of c1 to be TAKEN and the card refused at DONE. The guarantee they state (no DONE on c1 without a measurement after the move back) now holds one step earlier, so `rework_stepping_stone_test.go` expects `ErrEvidenceStale` at the offer, then measures, offers, lands and finishes. In the N25 RED commit. No test deleted or skipped.
- J2: the N25 check calls `landingStands`, not a bare "passing record after the mark", so a re-offer whose newest run of a kind failed is refused at the entrance too (same trap otherwise).
- J3: the N25/N28 check at Enqueue runs only when `namedOfferOf` finds a transition naming the candidate; the legacy fallback of `offerOf` would make a never-offered c2 inherit c1's offer.
- J4: the N27 check reads kinds green, gate, review; red and merge evidence are about other commits by design.
- J5: the N27 rule is the literal one: after c1 green+gate and a gate-only c2, an offer of c1 is refused until c1 is measured again.
- J6: the Q1 test sets the offer's `evidence_mark` with a raw `UPDATE` on a fresh ledger (after N27 the edge is reachable through the verbs only on a migrated ledger).
- J7: `TestEnqueue_NeverReplacesALandedRow` asserted "not even a higher card epoch replaces a landed row" with the new commit measured IN the new epoch; that is exactly what N3 now allows. The test was renamed `TestEnqueue_ADifferentEpochDoesNotReplaceALandedRowWithACommitMeasuredBeforeIt` and the commit's evidence moved to the old epoch, so it keeps the refusal that still holds; the accepted form is asserted in `enqueue_reclaimed_landed_test.go`. Committed apart (`7823406339`), it passes on the code before and after.
- J8: `TestMove_DoneNeedsTheLandedCandidateToBeTheNewestWork` had a subtest "a gate on another commit alone does not decide" expecting DONE. N28 (operator order) reverses it; it now expects refusal and the recovery by measuring c1 again. In the N28 RED commit.
- J9: the leader's note on N28 reads two ways for `TestA2Done_ReworkGateOnlyFixThenRemeasureC1` (the audit's test ends in DONE after "a gate on c1"). I took the reading that satisfies all three listed pass criteria: that DONE is refused and the recovery is measuring c1 again IN FULL (green and gate) after the other commit's evidence (condition 3). A gate of c1 alone is refused; the tests pin that stricter reading. If the leader meant only "the first DONE with c2 alone", condition 3 can be relaxed to one passing run of either kind by deleting the `for` over kinds; the N28c mutation shows which tests would then change.
- J10: condition 3 applies when the card LEFT MERGING after the offer (to REVIEWING, BUILDING, READY), not only when it re-entered BUILDING. The operator's N28 shapes include a withdrawal to REVIEWING without BUILDING. Evidence about another commit recorded while the card never left MERGING does not trigger it (control test), as the leader's "if there was a move back" asks.
- J11: the failing-newest-run refusal applies to every first offer at Enqueue, not only re-offers. With condition 1 at DONE for every landed candidate, a first offer with a failing green would otherwise land and then be stuck. This changes what a first offer needs (a failing green blocks it); my own earlier N27 D5 test used a failing green and was changed to a passing one in `9890afaaa3`.
- J12: N3 requires "a passing green or gate run recorded under the new card epoch, and no failing newest run" for both the take-over and the replacement. The leader's text says "re-measure"; it does not say both kinds. D5 stays (gate-only accepted), so one passing run of either kind is enough there.

## Gaps

- The mergeq and CLI repros (`TestA1Done_*`, `TestA2*`, `TestZ2Strand_*`, `TestZ2Done_*`, `TestA2Strand_Paths`, `TestA2Done_OlderCandidateOverNewerGreen`, `TestA2Done_FailureRecordedAfterLandingNoRework`, `TestA2Done_ReworkGateOnlyFixThenRemeasureC1`) were not run: they live in `internal/cli/**` and the integration tree. Lane-2 runs them after merge. The ledger-level re-expressions above are not those tests.
- N3 at the verb level: the real `af reclaim`/`af land` path (claim, evidence verb, `enqueue` verb, `reconcile`) was not run; only `Store.Claim`, `AddEvidence`, `Enqueue`, `Move` were. A CLI path that records evidence without the card's current epoch would not pass `measuredInEpoch`.
- The error mapping of the new `Enqueue` refusals (`ErrEvidenceStale`, `ErrInvalid`) through `internal/cli/af.go` was not observed.
- The mutation pairs "N25 and N27 both off" and "condition 2 of `landingStands` off" were not run. Condition 2 (a pass after the mark) is held by earlier tests (`TestMove_DoneRefusedForALandedCardBackInBuilding` and others), which were not re-mutated in this session.
- Coverage of the ledger package: 84.8% in the verbose run, 85.0% in a second profile of the same tree; the baseline of the previous report was 85.1%. The uncovered lines in the new code are the `fmt.Errorf("ledger: %w", err)` returns after a failed query; no fault-injection test was added. The 85% target is therefore not met in the run named above, and the margin is inside run-to-run variation.
- Refused commands: the worktree guard refused (a) a compound command containing `sed -i` and (b) a heredoc-plus-python rewrite of `queue.go`. Both were redone with the Edit tool or as single plain commands; no check was skipped because of them. The mutation scripts run `sed` without `-i` into the scratchpad and were not refused.
- Windows build, Linux run, and the whole `agenticfactory` race run were not requested and not run.
- Baseline test count before this session was not re-measured; the previous report gives 327 PASS lines, this run 388, and the 61 added lines were not itemized per file.
- The N24 tests use a v1 builder written for this change (`v1Millisecond`), not the audit's real v1 files; those were run once from copies and not kept.

## Residual-risk

- Behavior changes by design: (N25) a commit offered, ended and re-offered after the card went back into BUILDING needs a passing run after that move; (N27) a holder that gates c2 after c1 and wants c1 in the trunk must measure c1 again first; (N28) a failing newest green or gate of a landed candidate now blocks DONE even for a card never taken back, and a first offer with a failing green is refused; after other work was recorded since the card left MERGING, finishing on the landed commit takes a full re-measure (green and gate). An honest holder who discards a gate-only experiment pays one extra run.
- N3 trust: the take-over and replacement rest on evidence the new holder recorded in its own epoch; the ledger cannot tell that the run really measured the trunk's content, as everywhere else. A holder that records a passing gate for the landed commit without running it takes over the landing; that is the same trust as any `AddEvidence`.
- Legacy ledgers: if a row's reason does not name its candidate, the Enqueue checks of N25/N28 do not run for it (`namedOfferOf` finds nothing). The DONE-side checks still apply.
- The N27 rule cannot tell whether c2 contains c1; a newest-evidence commit that is a descendant of the candidate still refuses it until it is measured again.
- Condition 3 looks at green, gate and review evidence about other commits after the latest move out of MERGING; a move out and back that crosses nothing (no evidence) leaves it inert, which matches the old behavior for those shapes.
- N26 (`--trunk`, `--repo`) is lane-2's; E6 and N37 stay outside this lane.
