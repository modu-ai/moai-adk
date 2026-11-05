# `moai af land` — lane-19 35th review repairs (lane-22)

Commit graph (verification-claim-integrity.md §2.3): RED `244e20ee50` → GREEN `ac4da7b2be` (items 2, 3) → merge of
lane-18 `822635ac8f` (with its RED `bfd98db940`) → RED `e56ad66322` → GREEN (lock reclaim through `ReapLocks`, item 4).
Base: `5d90e66fe7`.

## Claim

- (2) A gate that exits 0 while a background child holds its output pipes lands the card, and the gate's process
  group is killed once the gate returns, whatever it returned.
- (3) The failure output is masked before it is cut to its tail; the help text says the gate receives the caller's
  full environment except the repository-scoping variables.
- (4) `af land` / `af reconcile` take back a dead holder's landing lock with `Store.ReapLocks(afAlive)`: no card
  lease is reclaimed and no note about another card is written.
- (1) A landing row whose card left MERGING after the lander wrote landing is settled landed by `af reconcile` when
  its candidate is in the trunk (lane-18's rule is in the tree; the path is covered by a test).

## Evidence

RED at `244e20ee50` (anchored, `-v`):

```
    outcome "failed" reason "... exec: WaitDelay expired before I/O complete: ": a gate that exited 0 failed the card
--- FAIL: TestAFLand_AGateThatLeavesABackgroundChildLandsAndTheChildIsKilled (5.62s)
    the tail kept the end of the secret: "... exit status 1: rsecretvalue\nyyyy..."
--- FAIL: TestAFLand_ASecretCutByTheTailLimitIsStillRedacted
--- FAIL: TestAFLand_HelpSaysTheGateReceivesTheCallersEnvironment
```

GREEN at `ac4da7b2be`: `--- PASS` for the three (6.08s, 0.91s, 0.00s) and for the 8 older `TestAFLand_*`; `ok 16.9s`.

RED at `e56ad66322`:

```
    the landing wrote a note about another card's lease: {Seq:1 Kind:stalled CardID:t2 Text:lease epoch 1 expired; the holder's liveness is unknown, so it was not reclaimed ...}
--- FAIL: TestAFLand_ReclaimingALockLeavesOtherCardsLeasesAndNotesAlone
```

GREEN after the switch to `ReapLocks`: both `TestAFLand_ReclaimingALockLeavesOtherCardsLeasesAndNotesAlone` and
`TestAFReconcile_ALandingRowOfAWithdrawnCardIsSettledLanded` `--- PASS`; `-run '^(TestAFLand_|TestAFReconcile_)'`
→ 14 `--- PASS`, `ok 15.3s`. `-run '^(TestAF|TestCardFlow)'` → 25 `--- PASS`, `ok 14.8s`.
`go test -count=1 -race -p 2 ./internal/agenticfactory/...` → 21 packages `ok` (mergeq 59.1s, ledger 10.1s,
core 99.5s), no `FAIL` line. `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./internal/cli/` and
`go vet ./internal/cli/ ./internal/agenticfactory/...` print nothing.

## Baseline-attribution

Each line was produced in this run by `go test` compiled from the integration worktree at the commit named beside
it; real git repositories and the real ledger were used. No installed `moai` binary is involved.

## Gaps

- `TestAFReconcile_ALandingRowOfAWithdrawnCardIsSettledLanded` passed on its first run: lane-18's landed rule and
  Reconcile's ancestry settlement were already in the tree. It is a regression guard for that path, not observed red.
  The help test also first passed against a loose assertion and was tightened before its RED was recorded.
- The `Land`-through-gate path for a card withdrawn after the lander wrote landing (as opposed to the reconcile path)
  has no test of its own.
- Native Windows is compiled, not run (no process group, no `sh`).
- `internal/web`, `internal/homestate`, `internal/hook`, `internal/template` and the whole `internal/cli` package were
  not re-run after these commits.

## Residual-risk

A secret the redactor does not recognise, written into `--gate` or printed by the gate, would be kept in the reason
and the landed note. The gate runs with `sh -c` and the caller's environment.
