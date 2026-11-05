# Ledger repair after audit 46dbcce0d4 (N41, N40, N39 ledger side, N38 ledger side)

Card issue-1948 (AF-18), SPEC-AGENTIC-FACTORY-001. Tree `.moai/worktrees/af-core-claim-runner`, branch `WT-agentic-factory-core-claim-runner`, start HEAD `396badfc6d4fb588d365a2541e579c8db0823c2f`, code HEAD at the end of the work `71904ebb02d16ce9b880c50863f63ec5c1cda33c` (the report and progress commit follows it). Local only: nothing pushed.

## 1. Claim

1. N41: at Enqueue only a passing green or gate of the candidate measures it again. A review of it (pass or fail) and a failing record of it do not; a review of ANOTHER commit still is the newest evidence. The ledger-level equivalent of `TestA3Done_FailingReviewCountsAsMeasuringAgain` is refused.
2. N40: review joins the newest-result judgment of a candidate (landing stands, Enqueue, measured-in-epoch). A FAILING green, gate or review about another commit recorded after the candidate was offered requires the candidate to be measured again (passing green and passing gate recorded after the failure), whether the failure came before or after the card left MERGING. A passing record about another commit while the card never left MERGING still blocks nothing (D5). merge and red evidence stay out.
3. N39 ledger side: Enqueue takes the optional `OnTrunk(bool)`. A landed row of an earlier card epoch offered again as the same commit is taken over only with `OnTrunk(true)`; with `OnTrunk(false)` or no option the row restarts as `waiting` and a decision note says so.
4. N38 ledger side: the ledger stores the repository identity in a new table `repository` (one row), API `BindRepository` / `Repository`; the first call binds, the same value again is a no-op, a different value is refused with `ErrWrongRepository` (`ledger: ...`). Schema version stays 2.
5. Ledger package coverage stays at or above 85.0% (87.7%).

Commits (RED fails on the parent tree, observed when written; GREEN passes the full package):

| Item | RED | GREEN |
|---|---|---|
| N41 | `6ed5d24ceb` | `3ef799ca90` |
| N40 | `590f3715f1` | `6b494aef42` |
| N39 | `ed8ea412b6` (carries the API stub) | `c054fe62ed` |
| N38 | `b8246743f6` (carries the API stub) | `023f20c7db` |
| pins for two edges | `71904ebb02` (tests only, they pass on the code before and after; added for mutants that survived) | none |

New API signatures (every change):

- `func (s *Store) Enqueue(ctx context.Context, cardID string, epoch int64, candidateSHA string, opts ...EnqueueOption) error` (was without `opts`; callers compile unchanged)
- `type EnqueueOption func(*enqueueOptions)`
- `func OnTrunk(on bool) EnqueueOption`
- `func (s *Store) BindRepository(ctx context.Context, identity string) error`
- `func (s *Store) Repository(ctx context.Context) (string, error)` (`""` while unbound)
- `var ErrWrongRepository = errors.New("ledger: the ledger is bound to another repository")`
- Schema: table `repository(id INTEGER PRIMARY KEY CHECK(id=1), identity TEXT NOT NULL CHECK(identity<>''), bound_at TEXT NOT NULL)` with triggers `repository_no_update` and `repository_no_delete`; created by `CREATE TABLE IF NOT EXISTS` at every Open, so `SchemaVersion` stays 2 and the one-way v1 to v2 migration is untouched.
- Behavior changes without signature change: `noFailingNewestRun` also reads review (Enqueue, `landingStands`, `measuredInEpoch`); Enqueue's "newest evidence" query skips the candidate's own review and failing rows; `measuredAfterOtherWork` adds the failure rule.
- Existing test call sites changed: one, `enqueue_reclaimed_landed_test.go` (take-over now passes `OnTrunk(true)`).

Judgments:

- J1. Without the option the ledger treats the candidate as not on the trunk (restart as waiting). The choice is the safe direction: a caller that forgets the fact cannot finish a card on an unverified landing. It makes any outside caller that relied on the old take-over change behavior until it passes `OnTrunk`.
- J2. A review of the candidate and a failing record of the candidate are skipped in Enqueue's newest-evidence query. A failing run of the candidate is refused anyway by the failing-newest-run check; skipping it only keeps the refusal naming the newer commit.
- J3. The re-measuring bar is the later of "newest failure about another commit after the offer" and "newest evidence about another commit after the latest move out of MERGING". The old rule is the second part.
- J4. The "bound once" rule is also held by triggers, so a holder of the file cannot update or delete the row through plain SQL (I7 style). Someone who drops the trigger can; see Residual-risk.

## 2. Evidence

All commands run from the worktree root on the code HEAD `71904ebb02`.

`go test -race -count=1 -cover ./internal/agenticfactory/ledger/` (verbatim):

```
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/ledger	197.464s	coverage: 87.7% of statements
```

`go vet ./internal/agenticfactory/ledger/...` exit 0, no output. `gofmt -l internal/agenticfactory/ledger/` printed nothing. `golangci-lint run ./internal/agenticfactory/ledger/...` (v2.1.6): `0 issues.`

`go build ./internal/cli/ ./internal/agenticfactory/...` printed nothing and exited 0; `go vet ./internal/agenticfactory/mergeq/` exited 0 (the mergeq tests in this tree call `Enqueue` with four arguments and still compile). `internal/cli` tests were not compiled or run.

RED observations on the parent tree (lines the failing runs printed when the tests were written):

- N41: `enqueue_measuring_again_test.go:46: a FAILING review of c1 counted as measuring c1 again: <nil>` and `:56: a review of c1 counted as measuring c1 again: <nil>`.
- N40: `the card finished although the review of its landed commit failed: <nil>`, `DONE with a failing review of the landed commit: <nil>`, `the card finished although gate of c2 failed after c1 was offered: <nil>` (gate and review, in MERGING and withdrawn), `a measurement of c1 older than the second failure was taken for enough: <nil>`: 8 failing subtests under 2 top-level tests.
- N39: with the stub that ignores the option, four subtests fail with `the row must wait like a new offer: {... State:landed ...}`.
- N38: against the `ErrNotImplemented` stub every `TestRepository_*` test fails (11 failing lines: 8 top-level tests and 3 subtests).

Mutation checks through `go test -count=1 -overlay=<json> ./internal/agenticfactory/ledger/` with a mutated copy of the file under a scratch directory (the tree was never edited; `git status --short` after the run shows only the progress file). The mutant list is the semantic equivalent of the audit's `N27-green-gate-only` and `R3-no-review` (their sed patterns match the pre-repair text, which no longer exists) plus mutants of the new rules. "FAIL lines" are top-level `--- FAIL:` lines of the package run.

| Mutant | What it does | FAIL lines | First failing test |
|---|---|---|---|
| N27-green-gate-only | the Enqueue newest-evidence query reads green and gate only | 1 | `TestEnqueue_AReviewOfAnotherCommitIsStillTheNewestEvidence` |
| R3-no-review | rule 3's "other work" query drops review | 1 | `TestDone_AReviewOfAnotherCommitAfterTheWithdrawalNeedsTheCandidateMeasuredAgain` |
| N41-review-counts | the query has no skip clause (old behavior) | 2 | `TestEnqueue_AReviewOrAFailureOfTheCandidateIsNotMeasuringItAgain` |
| N41-failure-counts | a failing run of the candidate counts as newest | 1 | `TestEnqueue_AFailingRunOfTheCandidateDoesNotHideTheNewerCommit` |
| N40-newest-run-no-review | `noFailingNewestRun` ignores review | 1 | `TestDone_AFailingReviewOfTheLandedCommitRefusesTheCard` |
| N40-failed-no-review | the failure rule ignores review | 1 | `TestDone_AFailureAboutAnotherCommitAfterTheOfferNeedsTheCandidateMeasuredAgain` |
| N40-failed-off | the failure rule is not part of the bar | 1 | same |
| N40-failed-from-zero | failures before the offer count | 1 | `TestDone_AFailureAboutAnotherCommitBeforeTheOfferBlocksNoGateOnlyCandidate` |
| N39-takeover-always | the old behavior | 1 | `TestEnqueue_ATakeOverAsksTheCallerWhetherTheCandidateIsOnTheTrunk` |
| N39-takeover-inverted | `OnTrunk` inverted | 2 | `TestEnqueue_AReclaimedCardTakesOverOrReplacesTheLandedRowOnceItMeasuredInItsOwnEpoch` |
| N38-no-compare | a different identity is accepted | 2 | `TestRepository_ADifferentIdentityIsRefusedAndNothingChanges` |
| N38-overwrite | `DO UPDATE` instead of `DO NOTHING` | 3 | `TestRepository_TheSameIdentityAgainIsANoOp` |
| N38-no-update-trigger | update trigger removed | 1 | `TestRepository_TheBindingCannotBeUpdatedOrDeletedByAnyoneHoldingTheFile` |
| N38-no-delete-trigger | delete trigger removed | 1 | same |
| N38-blank-allowed | whitespace identity accepted | 1 | `TestRepository_ABlankIdentityIsInvalid` |

All 15 mutants are caught. Two mutants (N41-failure-counts, N40-failed-from-zero) first survived the RED/GREEN tests; the pins in `71904ebb02` close them. A first run of `N38-blank-allowed` did not compile (unused import); the mutant was fixed and rerun, the table shows the rerun.

Ledger-level mapping of the audit's acceptance tests (CLI names are the leader's list):

- `TestA3Done_FailingReviewCountsAsMeasuringAgain` is `TestEnqueue_AReviewOrAFailureOfTheCandidateIsNotMeasuringItAgain`.
- `TestA3Risk_ReviewFailsAfterLanding` and the review case of `TestA3Done_OtherKindsFailAfterLanding` are `TestDone_AFailingReviewOfTheLandedCommitRefusesTheCard`; merge and red stay excluded in `TestDone_AFailingMergeOrRedRecordIsNotTheLandedCommitsResult`.
- `TestA3Done_FailureAboutANewerCommitAfterLanding` and `...BeforeTheWithdrawal` are `TestDone_AFailureAboutAnotherCommitAfterTheOfferNeedsTheCandidateMeasuredAgain`; D5 is pinned by `TestDone_APassingRecordAboutAnotherCommitBeforeTheWithdrawalBlocksNothing`.
- `TestA3Codex_P1_*` (ledger side) are `TestEnqueue_ATakeOverAsksTheCallerWhetherTheCandidateIsOnTheTrunk` and `TestEnqueue_OnTrunkChangesNothingOutsideTheTakeOverOfALandedRow`.
- N38 ledger side: `TestRepository_*` in `repository_identity_test.go`, including v1 and v2-without-table files opening into the new schema, a race of two first writes (exactly one winner), and a broken file returning a `ledger:` error.

## 3. Baseline-attribution

- Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/af-core-claim-runner`, code HEAD `71904ebb02d16ce9b880c50863f63ec5c1cda33c`, start HEAD `396badfc6d4fb588d365a2541e579c8db0823c2f`.
- Before the work, in the same tree: `go test -race -count=1 -cover ./internal/agenticfactory/ledger/` printed `ok ... 29.127s coverage: 87.8% of statements`. After: 87.7% (table above). No carried-over number.
- Judging build: the tests and vet were built from this tree by `go test`/`go vet` of the toolchain in the shell; `golangci-lint` is the installed v2.1.6 binary of the shell path (its own commit is not a repository build, so it is not compared with this HEAD).
- The audit's own repros and logs were read from `.moai/reports/issue-1948/final-code-audit-46dbcce0d4/repro/`; none of their CLI tests was copied or run here.

## 4. Gaps

- mergeq repro and CLI repros were not run; `./internal/cli/...` was not run at all (only `go build` of it, plus `go vet` of mergeq). The N38 verb side, the N39 verb side, N42, N43, N45 are lane-2's and are untested here.
- Outside callers that need updating by lane-2: `internal/cli/af.go:227` (`st.Enqueue(ctx, args[0], f.epoch, candidate)`) must ask the bound repository whether the candidate is on `refs/heads/af/trunk` and pass `ledger.OnTrunk(...)`; every verb must compare the repository it uses with `Repository()` and call `BindRepository` at first write. Other callers found by grep in this tree (compile verified, behavior unchanged for them): `internal/cli/af_land_test.go:64`, `internal/agenticfactory/mergeq/real_ledger_test.go:46`, `real_ledger_settle_test.go:106`, `withdrawn_requeue_test.go:38`. The CLI take-over tests named `TestAFVerbs_N3*` in the audit mutant log are not present in this tree; in the tree that has them they will see the restart-as-waiting behavior until `af.go` passes `OnTrunk`.
- The audit's own sed mutants were not applied literally (their patterns match text that no longer exists); semantic equivalents were run (section 2).
- The mutation set covers the new and changed rules only, not the whole package.
- The full-package race run took 197.464s against 29.127s for the baseline run before the work; the cause of the difference was not investigated (the test count grew, other load on the machine was not measured). One run, no repeat.
- A Bash call that fed a python text containing a git word was refused by the worktree guard ("cannot be shown to stay inside the worktree"). The edit was redone with the Edit tool; no result depends on it.
- The task said the tree was clean, but `.moai/reports/agentic-factory/18-progress.md` carried two uncommitted lines from lane-1 (ACK 7 and a hand-off line, written before this work). They are committed together with this task's result line, nothing else was added to them.
- The `git add -f` of this report was needed because `.moai/reports/*` is ignored.
- Not observed: behavior of the ledger under a real two-process race for `BindRepository` (the test races two goroutines on one Store; SQLite serializes writers, the multi-process case is by design, not measured).

## 5. Residual-risk

- Wrong repository bound at first write stays bound: whoever calls first decides, there is no verification at bind time and no rebind verb. Fixing it needs a deliberate operator act on the file.
- Moving the repository on disk (a different git common directory path) is refused, because the identity is compared as a plain string; two spellings of one path (symlink, case, trailing slash) are two identities unless the caller normalises. The ledger does not.
- A holder of the file who drops the `repository_no_update` trigger can rewrite the identity; the guard stops accidents and plain SQL, not a deliberate edit.
- An outside caller that does not pass `OnTrunk(true)` never takes over a landed row; the row then needs a real landing, which the merge side must be able to do for a commit that is already an ancestor of the trunk. That behavior is lane-2's to verify.
- The failure rule is deliberately strict: a failing experiment on another commit after the offer, even one the holder meant to drop, now costs one passing green and gate of the candidate. A review that failed on a non-risk card also blocks Enqueue until a passing review is recorded.
- The ledger cannot tell whether commit c2 contains c1; every judgment here is about equality of commit ids, as before.
- Mutation coverage and the 87.7% figure are from one run each on this tree; a later merge with lane-2 changes needs a fresh measurement.
