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

## Audit (sync-audit, iter 1)
- Verdict file `.moai/reports/t2106/sync-audit.md`: PASS-WITH-DEBT 94 @901273f155; no blocking defects; MINOR F1-F3 (+F4 out of scope) left as debt.
- Codex call: one `codex_audit` call by the auditor, receipt rcpt-b7cae4d26e4f1d2872464a77, pass, 0 findings (Codex read the diff only; it ran no tests).
- Basis for the Codex call: the repository gate `audit.gates.codex: required` and the leader-relayed standing decision d-20261106T064700Z-de2d (max 2 per card), which I read on the decision board. The question I put to this session's user about the call was unanswered ("No response after 60s"), so the user did not approve this call; it rests on the standing decision. Disclosed here.
- Own re-run on the audited HEAD after the audit: target test + `TestAppendMoveListLine_` `-count=3` -> ok (9.2 s).
- Auditor measured old code (movelist.go overlaid with 9cf8a1fc14): target test failed 3 of 10 with `pending,reversed,moved`.

## Landing preparation (after the complete freeze was lifted)
- Lease recovered by the formal path (`moai factory next --card t2106 --run to2bbk` from the primary checkout, then `stage t2106 merge-ready 901273f155` -> v13). No refused-stage reset (8859) was used.
- Local main 04d8c27a7e merged into the card branch (merge commit cd86c39b4c, no conflicts). Re-measure on that tree: `go test ./internal/cli/worktree -run 'TestApply_PendingIntentButNeverMoved_IsSettledThenMoved|TestAppendMoveListLine_' -count=5 -json` -> exit 0; in the JSON 35 pass events at test level, 0 `"Action":"fail"` events.
- First `moai factory complete` was refused: card has no SPEC, so its merge evidence must be `.moai/reports/t2106/verdict.md`. `verdict.md` is a byte copy of the auditor's `sync-audit.md` (same machine lines, `verdict: PASS-WITH-DEBT`, `audited_sha: 901273f1558dcadfb090f564d14e90270ef148d8`); the audit itself was not redone.

## Absorb of the batch main, re-repair and re-audit (2026-11-07, after the batch landed)
- Leader notice d-20261106T230753Z-d852: main b25adf4fca (batch of 38 + repairs). t2106 was left out of the batch because t2079 fixed the same revLines ordering bug differently and the two cancelled each other. Re-lease: `moai factory next --card t2106 --run to2bbk` (CLI; the MCP `factory_next` refused with `stale_server`, so the CLI equivalent was used). Tree entered by absolute path, toplevel verified. No audit was running and the tree was clean (`git status --short` empty) when main was absorbed.
- `git merge --no-ff b25adf4fca` -> merge commit 8bac970c9f, auto-merged without a conflict. On that tree the move-list tests FAILED 4: `TestAppendMoveListLine_EqualToEarlierEntryButNotLatest_IsAppended`, `TestAppendMoveListLine_EqualToLatestEntry_IsNotAppended` (t2079's), `TestAppendMoveListLine_SameTextAfterASettlementOfTheSamePathsIsAppended`, `TestAppendMoveListLine_ANonRecordLineIsComparedWithTheLatestLineOnly` (t2106's). Cause read from `movelist.go`: `revLines` reversed twice (t2079's loop from the end plus t2106's swap block).
- Repair `56017148b4`: the swap block is removed, `latestLineNamingPathsOf` and `namesAny` stay. After it: move-list tests `ok` (17.5 s); the formerly flaky test `-count=25` -> `ok` (111.5 s, run without -v so per-run lines were not read); `gofmt -l internal/cli/worktree` empty.
- Re-audit: one Claude `sync-auditor` at the fixed SHA, no Codex or GLM call (operator rule d-20261106T051044Z-7a4c; the Codex gate is advisory since b25adf4fca). Verdict `.moai/reports/t2106/sync-audit-iter2.md` (copied to `verdict.md`), committed in ea077ac2c5: PASS-WITH-DEBT 95, `audited_sha 56017148b4ba7c06abde8fdcc913cd647f0d572a`, 11 tests run 11 passed, mutant M1 (second swap back) -> 5 red, M2 (main's `rev[0]` check) -> 1 red, M3 (`namesAny` old_path only) survived (optional F5). Verdict file read from disk by this lane before use.
- `moai factory rebind-evidence t2106 --run to2bbk` -> `evidence_sha 56017148b4ba (was 901273f1558d)`.
- decision record: decided_by=claude-lane-19 evidence_refs=.moai/reports/t2106/sync-audit-iter2.md#56017148b4,leader-notice:d-20261106T230753Z-d852 ladder_path=1-disk-evidence (repair kept t2106's helper, dropped only the duplicate reversal)

### Gaps (this round)
- `card-review` (`codex_review` scope card) was not run: the Codex gate is advisory by the leader's notice and the operator's rule allows no Codex tools for this audit; `card-review.md` is therefore absent, reason recorded here.
- The auditor did not measure the merged tree `8bac970c9f` itself (it reproduced the state with mutant M1); the four failures above are this lane's own run.
- Only the move-list test subset plus the 25x run were run, not the whole `internal/cli/worktree` package, `-race`, Windows, or golangci-lint.
- Optional debt carried: F5 (no test distinguishes `new_path` matching), F6 (no test through an actual writer), F1-F4 from iter 1.
