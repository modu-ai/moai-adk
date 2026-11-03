# Final code audit FAIL 0.74 @1438ac5552 — mergeq repairs (lane-22)

Order in the commit graph (verification-claim-integrity.md §2.3): RED `6df59c0851` → GREEN `dfe0d19976` (F1 F2 F3 F6) →
tests `9b32756794` (F5) → merge of lane-18 `e3f385fd08`. Base: `01e37af0a2`.

## Claim

- F1 (BLOCKING): a symbolic trunk is refused with `ErrBadTrunk`, and `update-ref` never follows a symbolic ref.
- F2 (BLOCKING): a lost compare-and-swap sends the row back to waiting under the lock epoch, so the retry's
  verifying write is accepted by the real ledger and the card lands.
- F3 (SHOULD-FIX): Reconcile reads the verifying row's lease again after it holds the landing lock.
- F6: the unused `fakeQueue.inLock` field is gone.
- F5: mergeq statement coverage 80.8% → 86.2% with failure-path tests, not padding.

## Evidence

RED at `6df59c0851` (anchored, `-v`, verbatim lines):

```
    final_audit_fix_test.go:30: a symbolic trunk: <nil>, want ErrBadTrunk
--- FAIL: TestLand_ASymbolicTrunkIsRefusedAndTheCheckedOutBranchStays (0.41s)
    final_audit_fix_test.go:46: update-ref followed the symbolic ref and moved work from 56023397d2b113df9fce90da3d321c8fdf096db2 to 48da9c15894fe63547615d8787a38b8ee03bd36e
--- FAIL: TestGitCLI_UpdateRefDoesNotFollowASymbolicRef (0.23s)
    final_audit_fix_test.go:63: a lost compare-and-swap ended as an error: ledger: invalid argument: a queue row goes from landing to verifying only through [landed failed waiting]
--- FAIL: TestLand_RealLedger_ALostCompareAndSwapIsRetriedAndLands (0.49s)
    final_audit_fix_test.go:93: actions [{CardID:t1 From:verifying To:waiting Why:the candidate is not in the trunk; it is verified again}] err <nil>, want none: the lease is live under the lock
--- FAIL: TestReconcile_RealLedger_TheLeaseIsReadAgainUnderTheLock (0.18s)
```

GREEN at `dfe0d19976` (same command, verbatim):

```
--- PASS: TestLand_ASymbolicTrunkIsRefusedAndTheCheckedOutBranchStays (0.88s)
--- PASS: TestGitCLI_UpdateRefDoesNotFollowASymbolicRef (0.39s)
--- PASS: TestLand_RealLedger_ALostCompareAndSwapIsRetriedAndLands (0.73s)
--- PASS: TestReconcile_RealLedger_TheLeaseIsReadAgainUnderTheLock (0.27s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq	2.546s
```

Coverage (`go test -coverprofile`, this tree): `dfe0d19976` → `coverage: 80.6% of statements`;
`9b32756794` → `coverage: 86.2% of statements` (`go tool cover -func` total `86.2%`). The audit's 80.8% was
measured on `1438ac5552`; this run measured 80.6% before the new tests on the changed tree.

After merging lane-18 `e3f385fd08` (integration `3bd4873bb4`): `go test -count=1 -race -p 2 ./internal/agenticfactory/...`
→ 21 packages `ok` (mergeq 26.1s, ledger 8.2s, ledger/importer 1.6s, core 59.6s, retirement 8.9s); no `FAIL` line.
`go build ./...` and `go vet ./internal/agenticfactory/...` print nothing.

## Baseline-attribution

Each line above was produced in this run by `go test` compiled from the integration worktree at the commit named
beside it. Real git repositories and the real ledger were used for F1–F3; no installed `moai` binary is involved.

## Gaps

- `./internal/web/`, `./internal/homestate/`, `./internal/hook/` and whole `./internal/cli/` (B1) were not re-run
  after these commits (no change in those packages).
- F4 and the rest of F6 other than the unused field belong to other lanes or were not dispatched to this lane.
- The 80.8% (audit) vs 80.6% (this run, before the new tests) difference is not explained: the measured tree differs.
- The symbolic-ref check uses `for-each-ref --format=%(symref)`; it was exercised for a symbolic ref, a plain
  ref and a missing ref, not for a ref that is symbolic to a missing target.

## Residual-risk

A repository whose trunk becomes symbolic after `validTrunk` ran is still protected only by `update-ref --no-deref`
(which replaces or refuses, never follows), not by the earlier refusal. `Reconcile` judges liveness by the caller's
clock against the lease expiry.
