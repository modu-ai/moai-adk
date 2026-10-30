# AC-AFL-017 and AC-AFL-013 — `moai af` verbs and the card flow (lane-22)

Tree: integration worktree `WT-agentic-factory-integration`. Commit order (the commit graph is the witness,
`verification-claim-integrity.md` §2.3): RED `5a9370f894` (AC-AFL-017) → RED `95e100e754` (AC-AFL-013) → GREEN
`b58b1566b6` (both). Base: `03490b9b8f`.

## Claim

1. `moai af` is registered, visible, and offers `claim`, `move`, `evidence`, `enqueue`, `release`, each with a run
   function and a one-line description (AC-AFL-017, `TestAFCommand_VerbsRegistered`).
2. A class A card goes claim → red/green/gate evidence → enqueue → merge-queue landing → MERGING(merge seen) →
   DONE through the `moai af` verbs on a real ledger and a real git repository; the card holds no lease afterwards
   and a `landed` note exists (AC-AFL-013, `TestCardFlow_ClaimToDone_ThroughMergeQueue`).

## Evidence

RED, `TestAFCommand_VerbsRegistered` at `5a9370f894` (verbatim):

```
$ go test -count=1 -v -run TestAFCommand_VerbsRegistered ./internal/cli/
=== RUN   TestAFCommand_VerbsRegistered
    af_test.go:31: `moai af` is not registered on the root command
--- FAIL: TestAFCommand_VerbsRegistered (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.524s
```

RED, `TestCardFlow_ClaimToDone_ThroughMergeQueue` at `95e100e754` (a skeleton `newAFCmd` without verbs, not
registered; verbatim):

```
$ go test -count=1 -v -run TestCardFlow_ClaimToDone_ThroughMergeQueue ./internal/cli/
=== RUN   TestCardFlow_ClaimToDone_ThroughMergeQueue
    af_flow_test.go:85: moai af claim --db .../af/ledger.db --session s-t1 --worktree ...: unknown flag: --db
        stderr: Error: unknown flag: --db
--- FAIL: TestCardFlow_ClaimToDone_ThroughMergeQueue (0.18s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.362s
```

GREEN at `b58b1566b6` (verbatim):

```
$ go test -count=1 -v -run TestCardFlow_ClaimToDone_ThroughMergeQueue ./internal/cli/
=== RUN   TestCardFlow_ClaimToDone_ThroughMergeQueue
--- PASS: TestCardFlow_ClaimToDone_ThroughMergeQueue (0.40s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.587s

$ go test -count=1 -v -run TestAFCommand_VerbsRegistered ./internal/cli/
=== RUN   TestAFCommand_VerbsRegistered
--- PASS: TestAFCommand_VerbsRegistered (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.232s
```

Neighbours at `b58b1566b6`: `go test -count=1 -p 2 -run Root ./internal/cli/` ok 75.1s; `-run Trivial` ok 1.0s;
`-run AgenticFactory` ok 2.0s; `go vet ./internal/cli/` and `go build ./...` print nothing (exit 0).

## Baseline-attribution

Every output above was produced in this run by the Go toolchain in the integration worktree, against the commit
named next to it. The judging build is `go test` compiled from that tree (no installed `moai` binary is involved).

## Gaps

- The verb names (`claim`, `move`, `evidence`, `enqueue`, `release`) are this lane's choice for the names OD-33
  leaves open; no operator decision pins them.
- `TestCardFlow_ClaimToDone_ThroughMergeQueue` reaches DONE through `move` and the merge queue; it does not pass
  through REVIEWING (`spec.risk` empty, diff under 400 lines) and does not assert that REVIEWING was skipped by a
  position history; the card spec simply has no risk.
- The 3-iteration fix-limit park (REQ-AFL-013) and the `fail_reason` return path are not exercised by this test.
- The `claim` verb binds the session's hook heartbeat under the project directory; the test isolates that with
  `CLAUDE_PROJECT_DIR`, so the real binding file is not covered.
- Whole `internal/cli` (B1) was not run.

## Residual-risk

The lease records the parent process of the `moai af claim` call, not the command itself; if a caller invokes the
verb from a short-lived wrapper, the lease holder reads as dead sooner than the session. Verb grammar may change when
OD-33/OD-29 close.
