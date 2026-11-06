# `moai af land` stage guard and gate hardening (lane-22)

Commit graph (verification-claim-integrity.md §2.3): RED `60140b904e` → GREEN `6a9474afe5` → merge of lane-18
`f052a5af31` (with its RED `108146a57a`) → test `stage_abort_test.go`. Base: `c58e444a87`.

## Claim

1. `af land` refuses a card that is not in MERGING at the epoch its queue row was made at (P2 of lane-19's 34th
   review); a holder who pulled the card back to REVIEWING cannot have the withdrawn work landed.
2. The gate runs under `--gate-timeout` in its own process group that is killed whole, with the repository-scoping
   variables removed, bounded output, and a redacted failure tail; the landed note names the gate.
3. `af land` and `af reconcile` take back a dead holder's landing lock before locking.
4. With lane-18's ledger check merged, a card withdrawn between the gates and the landing makes `Land` stop with
   `ErrStaleRow`: nothing reaches the trunk and the lock is released.

## Evidence

RED at `60140b904e` (anchored, `-v`; reason lines shortened with `...`):

```
    af_land_guard_test.go:58: landing a card that left MERGING: <nil>, want a refusal naming MERGING
--- FAIL: TestAFLand_ACardThatLeftMergingIsNotLanded
    ... --gate sleep 60 & ... --gate-timeout 1s: unknown flag
--- FAIL: TestAFLand_ATimedOutGateIsKilledWithItsChildren
    outcome "failed" reason "gates failed: gate \"test -z \\\"$GIT_DIR\\\"\": exit status 1: ": the gate saw GIT_DIR
--- FAIL: TestAFLand_TheGateRunsWithoutRepositoryScopingVariables
    landed note "<sha> -> <sha> on refs/heads/trunk" does not name the gate "test -f a.txt"
--- FAIL: TestAFLand_ALandedNoteSaysWhichGateVerifiedIt
    the reason carries the secret: "... exit status 1: password=hunter2supersecretvalue"
--- FAIL: TestAFLand_ASecretInAFailedGatesOutputIsRedactedFromTheReason
    outcome "waiting" reason "the landing lock is held by another lander": a dead holder's lock held the queue
--- FAIL: TestAFLand_TheLandingLockOfADeadHolderDoesNotHoldTheQueue
```

GREEN at `6a9474afe5` (same command):

```
--- PASS: TestAFLand_ACardThatLeftMergingIsNotLanded (0.49s)
--- PASS: TestAFLand_ATimedOutGateIsKilledWithItsChildren (1.40s)
--- PASS: TestAFLand_TheGateRunsWithoutRepositoryScopingVariables (0.31s)
--- PASS: TestAFLand_ALandedNoteSaysWhichGateVerifiedIt (0.43s)
--- PASS: TestAFLand_ASecretInAFailedGatesOutputIsRedactedFromTheReason (0.39s)
--- PASS: TestAFLand_TheLandingLockOfADeadHolderDoesNotHoldTheQueue (0.35s)
ok  	github.com/modu-ai/moai-adk/internal/cli	4.337s
```

After merging `f052a5af31`: `TestLand_RealLedger_ACardWithdrawnDuringTheGatesIsNotLanded` →
`--- PASS` (0.57s). It was **not observed red**: the ledger check it depends on was already in the tree when the test
was written. It asserts `ErrStaleRow`, an unmoved trunk, a row that is neither landing nor landed, and a released
lock; without the ledger check the landing would succeed and the assertion would fail (inferred, the check was
not removed to see it).

Whole set at the final tree: `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./internal/cli/` and
`go vet ./internal/cli/ ./internal/agenticfactory/mergeq/` print nothing;
`go test -count=1 -v -run '^(TestAF|TestCardFlow)' ./internal/cli/` → 19 `--- PASS`, `ok 5.3s`;
`go test -count=1 -race -p 2 ./internal/agenticfactory/...` → 21 packages `ok` (mergeq 25.7s, ledger 7.4s,
core 53.9s), no `FAIL` line.

## Baseline-attribution

Each line was produced in this run by `go test` compiled from the integration worktree at the commit named beside
it, on real git repositories and the real ledger. No installed `moai` binary is involved.

## Gaps

- The gates in these tests are small shell commands; no real build or test suite was run as a gate.
- The kill-the-children test checks one grandchild (`sleep 60`) by `kill -0`; the Windows path (no process group,
  no `sh`) is compiled (`GOOS=windows go build`) but not run.
- `--gate-timeout` is exercised with 1s; the default 20m and the 30m verb timeout were not.
- A dead holder's lock is reclaimed through `Reap` with a judge that answers Unknown for every card lease holder;
  the test kills a real process that held the lock, but a lock held by a process without a fingerprint is not tested.
- `core.Redact` masks what it recognises (a `password=` assignment was tested); other secret shapes in a gate's
  output are only as covered as `Redact` itself.
- `internal/web`, `internal/homestate`, `internal/hook`, `internal/template` and whole `internal/cli` (B1) were not
  re-run after these commits.

## Residual-risk

The gate text is recorded in the landed note (after redaction of recognised secrets); a secret the redactor does not
recognise, written into `--gate`, would be kept. `--gate` runs with `sh -c`, so only a trusted caller should pass it.
