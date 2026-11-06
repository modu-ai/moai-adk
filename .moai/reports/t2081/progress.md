# t2081 progress (lane-19, run to2bbk, Class C)

Operator ask (via leader): run the migration's `git worktree move` for several trees at once ("병렬로 이전"), but first prove concurrent safety (same repository's registration locking; journal and move-list order; the reverse-order guarantee), then bound it with a defaults constant, then give before/after numbers; if the proof fails, do not implement and report why.

## Setup
- `moai factory next --card t2081 --run to2bbk` -> `t2081 stage=- worktree=t2081`, tree .../worktrees/projects/moai-adk-go/t2081, branch WT-git-worktree-move. Base was local main 9b632c1eb3; t2070's commit 2ac39de91c merged first (instruction), later t2070's repair commit 23f0568bc6 merged so both branches carry the same engine code.
- The lease lapsed twice while I worked elsewhere (05:04:24Z); each time I re-leased with `moai factory next --card t2081 --run to2bbk` from the parent checkout.
- Environment of every measurement below: macOS, Apple git 2.54.0 (Apple Git-157), a machine also running other lanes' work (large load noise).

## 1. Proof, part one: git itself (file `internal/cli/worktree/migrate_parallel_t2081_test.go`)

What can run concurrently is only the `git worktree move` call. The shared Go state is not touched by it: the journal rewrite (`migrationJournal.flush`) and `appendMoveListLine` (read the list, compare with the latest line, append) have no lock, so they stay one tree at a time; `hook.RekeyWorktreeEntry` already holds a process mutex and a file lock (read from internal/hook/worktree_registry.go), `rebindCardForPath` and the comparison after the move also stay serial.

Tests:
- `TestGitWorktreeMove_ConcurrentMovesOfDifferentTreesAreConsistent`: 12 registered trees x 8 rounds. Each round moves every tree at once to the other spelling and back, while two `git worktree list` loops and four `git status` loops hammer the same repository and trees. After each round the check compares git's own view with the disk: `worktree list` names exactly the trees where they are, `worktree prune --dry-run -v` is empty, every tree answers `rev-parse --show-toplevel` with its own path, and its `gitdir` back-pointer names that place. Result: PASS, 96 concurrent moves, no failure (`go test ... -run TestGitWorktreeMove_ -v`, 13.9-16.0 s).
- Positive control `TestGitWorktreeMove_TheSameTreeMovedTwiceAtOnceFailsForOne`: two simultaneous moves of ONE tree -> exactly one succeeds, and git and the disk still agree afterwards. PASS.
- Detector control `TestGitWorktreeMove_TheConsistencyCheckSeesADamagedRegistration`: the consistency check reports an error when a tree's back-pointer is rewritten to a place the tree is not. PASS. (Without this the check passing would prove nothing.)
- Environment variables `MOAI_T2081_STRESS_TREES` / `MOAI_T2081_STRESS_ROUNDS` raise the size for a heavier run; the defaults were what ran here.

## 2. Proof, part two: the engine's order and recovery

Design (`runMigrationEngineWith`, `prepareTree` / `finishTree` in migrate.go): up to `config.DefaultWorktreeMigrateMoveConcurrency` (= 4) trees form a batch.
1. Prepare, one tree at a time: manifest, instance token, journal `pending` entry, move-list `pending` intent line.
2. Move: the batch's `git worktree move` calls run together.
3. Finish, one tree at a time in plan order: re-key, card rebind, comparison with the manifest read before the move, the `moved` line, the journal settle — or the one reverse order on any failure.
A batch of one is the old strictly serial order (`moveConcurrency` seam; test pins it).

Tests (same file):
- `TestMigrationEngine_MovesABatchTogetherAndFinishesInPlanOrder`: the move seam holds each call until a second is in flight; peak >= 2 and <= the cap; report rows and move-list `moved` lines in the same order; no journal step left pending.
- `TestMigrationEngine_ConcurrencyOneIsTheOldSerialRun`: a cap of 1 gives peak 1.
- `TestMigrationEngine_StoppedAfterAWholeBatchIsPreparedThenRerunMovesEverything`: the run is stopped (hook `migrationBatchHook`) after the batch's pending lines and before any move: the journal holds one pending step per batch tree, every tree still at its source; the rerun settles each and moves them all (nothing left interrupted).
- `TestMigrationEngine_StoppedAfterABatchMovedThenRollbackRestoresEveryTree`: stopped after all moves returned and before any re-key: the trees stand at their destinations; `moai worktree migrate --rollback <journal>` brings every tree back to its source and git lists it there.
- Tests were written AFTER the implementation here (the design came out of reading the code); to show they can fail, a mutant fixed the batch width at 1: the three batch tests failed (`peak 1`, `the batch held 1 trees; the scenario needs several` x2); reverted, `git diff --stat` showed no change to migrate.go. The t2070 tests (read-ahead bound, manifest equivalence, stale digest still reversed) are unchanged and run in the sweep.

Crash model, stated: a kill during a batch can now leave up to 4 trees with a pending journal step and intent line at once (before: at most 1). `--rollback` and the rerun settle steps one by one from the journal, which the two stop tests exercise for a whole batch. A kill DURING the concurrent moves (some renamed, some not) was not injected as such; each tree's state there is one of the two states the existing per-tree tests (`ForwardMoveInterrupted*`) already cover, and the stop tests cover the all-moved and none-moved ends.

## 3. Numbers (probe `probe-end-to-end-width.go.txt`: 100 extra trees, real lsof, 104 moved / 2 skipped each time)

Measured in the same session, alternating serial (`MOAI_T2081_WIDTH=1`) and the default cap of 4; the machine was loaded by other work, so absolute values drift between pairs and only pairs are comparable:

| pair | cap | plan | whole apply | move phase | per moved tree |
|---|---|---|---|---|---|
| 1 | 1 (serial) | 1.43 s | 27.94 s | 26.51 s | 255 ms |
| 1 | 4 | 1.28 s | 21.44 s | 20.16 s | 194 ms |
| 2 | 1 (serial) | 3.38 s | 50.95 s | 47.57 s | 457 ms |
| 2 | 4 | 1.79 s | 31.90 s | 30.10 s | 289 ms |

Whole apply 23% (pair 1) and 37% (pair 2) shorter with the cap of 4; the move phase 24% and 37% shorter per tree. The gain is the `git worktree move` call only; the rest of a tree's work (manifest after, re-key, rebind, settle) is still one tree at a time. Against the t2070 numbers from before this session (13-15 s whole apply at lighter load) these runs are slower in absolute terms because the machine was busier, which is why the comparison is within pairs.

## Verification
- `gofmt -l internal/cli/worktree internal/config`: empty. `GOOS=windows go vet ./internal/cli/worktree ./internal/config`: no output. `GOOS=windows go test -c -o /dev/null ./internal/cli/worktree`: no output.
- Migration-surface sweep `go test ./internal/cli/worktree -run "M6|Migrat|PlanMigration|Journal|Rollback|Interrupted|Manifest|Progress|Classify|Verify_|Plan_|Apply_|MovedTo_|UpdateConsolidation|GitWorktreeMove|TestProcessCWDIndex|TestTreeProgress" -count=1` (432 s, no `-v`): exactly one `--- FAIL`, `TestApply_PendingIntentButNeverMoved_IsSettledThenMoved` (move-list results = pending,reversed,moved, want pending,reversed,pending,moved), the test t2070 showed to fail on the pre-change tree too. Not the full package.

## Gaps
- Only macOS with Apple git 2.54.0 was run. Linux and Windows git were not, and git's locking of `.git/worktrees` could differ there; the cap of 4 applies on every platform. The proof tests above run in CI on the OS matrix and would show a difference there; until then it is unobserved.
- The cap of 4 is a chosen value, not a measured optimum; the proof tried every tree of a round (12) moving at once.
- A kill in the middle of the concurrent moves was not injected; see the crash model above. The stop tests cover the states just before and just after the concurrent step.
- The gain was measured on tiny fixture trees under variable machine load, not on mo.ai.kr's trees. On large real trees the manifest and the rest of the serial part dominate more, so the share the concurrent move saves is expected to be smaller (not measured).
- `/verify` (the `/moai gate`) was not invoked before the commits: it runs tests beyond the targeted set and the card's constraint forbids the full suite while the leader's regression runs. The scoped equivalents above stand in.
- golangci-lint was not run.
- The flaky test above is unfixed and unexplained here (the t2070 audit's reading of `revLines` is a hypothesis, unverified).
- Guard refusals: a shell heredoc used to append the new tests to the test file was refused ("... too complex to verify that it stays inside the worktree. Refusing to run it ... Split it into plain, separate commands"); I made the same edit with the Edit tool instead and checked it by running the tests. A later notice (`WorktreeGuardRefusal ... without the cross-tree redirect`) arrived after a successful plain command; I did not find a command it referred to.
- Lease: expired twice (see Setup), re-leased through the normal verb each time.

## Options (not done)
1. Raise or lower `DefaultWorktreeMigrateMoveConcurrency` after a run on large real trees.
2. Also overlap the after-the-move manifest of finished trees with later batches (saves more, but a mismatch would then be found after more trees have moved, which changes the reverse-order guarantee).
3. Make the move-list append and the journal flush safe for concurrent callers (locks) so rekey/rebind could overlap too; not needed for the current gain and a larger change to files `--rollback` recovers from.
