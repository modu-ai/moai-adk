# `moai af land` / `moai af reconcile` and the landing-stuck repair (lane-22)

Commit graph (verification-claim-integrity.md §2.3): RED `e1df123331` → GREEN `0ca5a52757` (mergeq, P2-1) →
GREEN `e18ebe1e43` (cli verbs). Base: `5fbde3f463`.

## Claim

1. The binary has a non-test importer of `agenticfactory/mergeq`: `git grep -n 'agenticfactory/mergeq"' -- ':!*_test.go'`
   prints `internal/cli/af_land.go:15`. `moai af land <card> --gate <cmd>` lands an enqueued card; `moai af
   reconcile` settles the rows a restart left in verifying or landing.
2. A row cannot be left stuck in landing: a swap that fails for another reason goes back to waiting; a swap that
   took effect but reported an error settles as landed; a landing row left by a dead lander whose candidate is not
   in the trunk is reset under the lock before the new verifying write.
3. The card flow test lands through `moai af land`, not through a direct `Merger` call.

## Evidence

RED at `e1df123331` (anchored, `-v`; verbatim lines, long paths shortened to `...`):

```
cli:
    af_flow_test.go:109: moai af land t1 --db .../ledger.db --repo ... --trunk refs/heads/trunk --gate true: unknown command "land" for "af"
--- FAIL: TestCardFlow_ClaimToDone_ThroughMergeQueue
    af_land_test.go:86: moai af land t1 ... --gate false: unknown command "land" for "af"
--- FAIL: TestAFLand_AGateThatFailsFailsTheCardAndLeavesTheTrunk
    af_land_test.go:135: moai af reconcile ...: unknown command "reconcile" for "af"
--- FAIL: TestAFReconcile_SettlesALandingRowWhoseCandidateIsInTheTrunk
    af_test.go:48: no `moai af land` verb for the land step
    af_test.go:48: no `moai af reconcile` verb for the reconcile step
--- FAIL: TestAFCommand_VerbsRegistered
mergeq:
    landing_stuck_test.go:36: the row is landing after the failed swap, want waiting (not stuck in landing)
--- FAIL: TestLand_RealLedger_ASwapThatFailsForAnotherReasonLeavesTheRowRetryable
    landing_stuck_test.go:71: result {...} err git was killed after it wrote the ref, want landed: the trunk did move
--- FAIL: TestLand_RealLedger_ASwapThatTookEffectButReportedAnErrorSettlesAsLanded
    landing_stuck_test.go:90: result {...} err ledger: invalid argument: a queue row goes from landing to verifying only through [landed failed waiting], want landed: ...
--- FAIL: TestLand_RealLedger_ALandingRowWhoseCandidateIsNotInTheTrunkIsResetBeforeVerifying
```

`TestAFLand_RefusesACardThatIsNotQueued` passed in the first RED run (the unknown command is also an error); the
assertion was then tightened to require "not in the queue" before the RED commit. It was not observed red in its
tightened form.

GREEN, mergeq at `0ca5a52757`:

```
--- PASS: TestLand_RealLedger_ASwapThatFailsForAnotherReasonLeavesTheRowRetryable (0.78s)
--- PASS: TestLand_RealLedger_ASwapThatTookEffectButReportedAnErrorSettlesAsLanded (0.40s)
--- PASS: TestLand_RealLedger_ALandingRowWhoseCandidateIsNotInTheTrunkIsResetBeforeVerifying (0.72s)
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq	2.195s
```

GREEN, cli at `e18ebe1e43`:

```
--- PASS: TestCardFlow_ClaimToDone_ThroughMergeQueue (0.43s)
--- PASS: TestAFLand_AGateThatFailsFailsTheCardAndLeavesTheTrunk (0.35s)
--- PASS: TestAFLand_RefusesACardThatIsNotQueued (0.17s)
--- PASS: TestAFReconcile_SettlesALandingRowWhoseCandidateIsInTheTrunk (0.21s)
--- PASS: TestAFCommand_VerbsRegistered (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	2.286s
```

At `e18ebe1e43`: `go build ./...` prints nothing; `go vet ./internal/cli/` prints nothing; `go test -count=1 -v -run '^TestAF'
./internal/cli/` → 13 `--- PASS` lines, `ok`; `go test -count=1 -race -p 2 ./internal/agenticfactory/...` → 21 packages `ok`
(mergeq 27.6s, ledger 8.2s, core 61.0s), no `FAIL` line.

## Baseline-attribution

Every line was produced in this run by `go test` compiled from the integration worktree at the commit named beside
it, on real git repositories and the real ledger. No installed `moai` binary is involved.

## Gaps

- The gate in the CLI tests is `true` or `false`: no build or test runs in the candidate worktree. A real gate
  command was not exercised.
- `af land` was run with a trunk that no worktree has checked out; the default `refs/heads/af/trunk` was not created
  or landed on in a test, and `--repo` defaulting to the project directory was not exercised.
- The "git killed after it wrote the ref" case is simulated by a Git wrapper that calls the real update-ref and then
  returns an error; no real killed git process was observed.
- A row left landing whose trunk cannot be read is left landing on purpose (nothing can be judged); that branch has
  no test.
- Lane-19's P2-1 also mentions SetState(waiting) failing after a lost swap; the `failed`-fallback of `leaveLanding`
  is covered only by reading, not by a test that makes the waiting write fail.
- `internal/web`, `internal/homestate`, `internal/hook`, `internal/template` and whole `internal/cli` (B1) were not
  re-run after these commits.

## Residual-risk

`--gate` is run with `sh -c`; its text comes from the caller and is not sanitised, so only a trusted caller should
pass it. The lock holder is the `af land` process, which lives only as long as the command: a killed `af land` leaves
a landing row that the next `af land` or `af reconcile` settles.
