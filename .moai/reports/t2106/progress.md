# t2106 progress (lane-19, run to2bbk, Class B, RC-blocking)

Card: `TestApply_PendingIntentButNeverMoved_IsSettledThenMoved` is unstable (RC42 4/10, RC43 8/10 failures, independent of -race). Establish the cause, reproduce with counts, RED, repair, and check the effect on every writer of the move list.

## Setup
- `moai factory next --card t2106 --run to2bbk` -> `t2106 stage=- worktree=t2106`, tree /Users/goos/.moai/worktrees/projects/moai-adk-go/t2106, branch WT-testapply-pendingintentb, base local main 9cf8a1fc14.

## Cause (read from code, then confirmed by the failing sequence)
`appendMoveListLine` (internal/cli/worktree/movelist.go) documents its duplicate rule as "no append when the line equals the latest entry naming either of its two paths". It compared the new line with `revLines(data)[0]`, and `revLines` is documented "last first" but returned the lines in FILE order, so it compared with the FIRST line of the file. Every move-list line carries `"ts"` in whole seconds (`m6MoveLineTriggerReason`), so a new line equals the file's first line only when both were produced within the same wall-clock second. In `TestApply_PendingIntentButNeverMoved_IsSettledThenMoved` the file's first line is the seeded `pending` and the engine's new `pending` for the same tree is dropped as a "duplicate" exactly when the two fall in the same second: the recorded results read `pending,reversed,moved` instead of `pending,reversed,pending,moved`. That is why it fails some of the time and not always. The same defect ran the other way too: a true duplicate of the latest record was appended again whenever another tree's record was first in the file.
`revLines` has one caller (`appendMoveListLine`); `appendMoveListLine` is called from migrate.go and tests, and its signature is unchanged.

## Reproduction counts (the test alone, `-count=10` unless noted, without -race)
- pre-change tree of card t2070 (earlier this session): 4 of 10 failed.
- this tree, base main 9cf8a1fc14, before the repair: 1 of 10 failed (variance is large; the leader's RC42 4/10 and RC43 8/10 are the same test).
- after the repair: `-count=25` -> 25 of 25 passed (104.8 s).
A clean 25 is consistent with the repair; it is not proof by itself (at a 1-in-10 failure rate 25 clean runs happen about 7% of the time). The deterministic proof is the unit tests below, which do not depend on the clock.

## RED -> GREEN (`internal/cli/worktree/movelist_dedupe_t2106_test.go`)
RED before the repair (3 failing, 2 passing controls):
- `SameTextAfterASettlementOfTheSamePathsIsAppended`: pending, reversed, then the same pending text -> must stay 3 records; got 2.
- `DuplicateOfTheLatestRecordIsDroppedWhateverComesFirst`: another tree's record first, then mine; appending mine again -> must stay 2; got 3.
- `ANonRecordLineIsComparedWithTheLatestLineOnly`: a non-record line is judged against the latest line only; failed in both directions.
- Controls that passed before and after: an identical re-run right after adds nothing; another tree's record in between does not change the comparison.
GREEN after: `revLines` really returns last first, and `latestLineNamingPathsOf` finds the latest record that names either path (old_path or new_path in either position) and the line is skipped only when it equals that record's text; a line that is not a record, or has no paths, is judged against the latest line only.

## Effect on every writer
All move-list writers go through `appendMoveListLine` (migrate.go: intent `pending`, `moved`, `reversed`, `refused`, `restored`, and the settle of an interrupted step). The migration-related test sweep (`-run "M6|Migrat|PlanMigration|Journal|Rollback|Interrupted|Manifest|Progress|Classify|Verify_|Plan_|Apply_|MovedTo_|UpdateConsolidation|MoveList|AppendMoveListLine" -count=1`, 504.6 s, no -v so the number of tests that passed was not printed) -> ok. The one intended change in behaviour: an identical line is dropped only when it is the latest record for those paths (before: only when it equalled the first line of the file, by accident of timing), so a re-applied identical step after a different record for the same paths is now recorded, as the function's own comment says.

## Verification
- `gofmt -l internal/cli/worktree`: empty. `go vet ./internal/cli/worktree` and `GOOS=windows go vet ./internal/cli/worktree`: no output.

## Gaps
- The cause was established by reading the code and by matching the failing result sequence and the second-resolution `ts`; no mutation run was done on the old code beyond the RED unit tests above.
- `-race` was not run for this card's tests (they touch no goroutines).
- No audit yet (one Claude `sync-auditor` is the operator's rule; if the repository's required Codex gate blocks a PASS, the standing decisions allow it and I will ask this session's user first - external paid call).
- Windows: vet only; the tests were not run there.
- The full `internal/cli/worktree` package and `/verify` (the gate) were not run (the leader's constraint).
- golangci-lint not run.
