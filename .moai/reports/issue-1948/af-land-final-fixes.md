# Final audit of 5d90e66fe7 — `af land` / mergeq repairs (lane-22)

Commit graph (verification-claim-integrity.md §2.3): RED `9c36ed1fb0` → GREEN `93f5fc3a2a` → merge of lane-18
`5190695ad3` (with its RED `7ad6ed93d3`) → this evidence. Base: `59f603bb8f`.

## Claim

- F1 (BLOCKING): `af land` refuses a card whose position carries a park (an operator hold in MERGING).
- F5: a gate that exits 0 and prints 9 MB lands; a failing gate keeps its last line (TailWriter keeps the last
  64 KiB, never fails). The git calls keep their refusing limit.
- F4: `leaveLanding` no longer attempts a failed write the ledger refuses; when the waiting write fails the row stays
  landing, the error says so, and the next Land lands.
- F7: mergeq statement coverage 83.6% (audit) → 85.8%.
- lane-19: help says a background process the gate left is killed and does not fail an exit-0 gate, and that Git
  Bash has no process-group kill; the refusal of a withdrawn card points at `moai af reconcile`.
- lane-18 `5190695ad3` merged: a parked card cannot advance its row; a withdrawn waiting row is queued again.

## Evidence

RED at `9c36ed1fb0` (anchored, `-v`; reason lines shortened with `...`):

```
    af_land_final_test.go:32: landing a parked card: <nil>, want a refusal saying it is parked
--- FAIL: TestAFLand_ACardUnderAParkIsNotLanded (0.61s)
    af_land_final_test.go:45: outcome "failed" reason "gates failed: gate ... exit 0: mergeq: git output exceeded its limit: xxxx..."
--- FAIL: TestAFLand_ALoudPassingGateStillLands (6.03s)
    af_land_final_test.go:58: the gate's last line is not in the reason (2166 bytes)
--- FAIL: TestAFLand_ALoudFailingGateKeepsItsLastLine (6.31s)
    the help of `af land` lacks "background process the gate left"; ... does not say that Git Bash has no process-group kill
--- FAIL: TestAFLand_HelpAndRefusalPointTheOperatorOnward (0.00s)
    the refusal of a withdrawn card: af land: t1 is not in MERGING ... want it to point at `moai af reconcile`
--- FAIL: TestAFLand_TheRefusalOfAWithdrawnCardNamesReconcile (0.35s)
    leave_landing_test.go:65: 1 failed writes were attempted: that fallback cannot succeed (the ledger wants a card-kind epoch there)
--- FAIL: TestLand_RealLedger_AFailedWaitingWriteLeavesTheRowLandingForTheNextLand (1.15s)
```

GREEN at `93f5fc3a2a` (same commands):

```
--- PASS: TestAFLand_ACardUnderAParkIsNotLanded (0.33s)
--- PASS: TestAFLand_ALoudPassingGateStillLands (0.93s)
--- PASS: TestAFLand_ALoudFailingGateKeepsItsLastLine (1.00s)
--- PASS: TestAFLand_HelpAndRefusalPointTheOperatorOnward (0.00s)
--- PASS: TestAFLand_TheRefusalOfAWithdrawnCardNamesReconcile (0.33s)
ok  	github.com/modu-ai/moai-adk/internal/cli	3.741s
--- PASS: TestLand_RealLedger_AFailedWaitingWriteLeavesTheRowLandingForTheNextLand (2.83s)
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq	3.148s
```

At the final tree (after merging `5190695ad3`):

```
$ go test -count=1 -cover -p 2 ./internal/agenticfactory/mergeq/
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq	44.837s	coverage: 85.8% of statements
$ go test -count=1 -race -p 2 -coverprofile=... ./internal/agenticfactory/mergeq/ ./internal/agenticfactory/ledger/...
ok  	.../mergeq	38.458s	coverage: 85.8% of statements
ok  	.../ledger	11.019s	coverage: 85.8% of statements
ok  	.../ledger/importer	1.798s	coverage: 89.6% of statements
```

`go test -count=1 -v -run '^(TestAF|TestCardFlow)' ./internal/cli/` → `ok 15.96s`, no `--- FAIL` (30 `--- PASS` before the
merge). Before the merge, `go test -count=1 -race -p 2 ./internal/agenticfactory/...` → 21 packages `ok`, no `FAIL` line.
`go build ./...`, `GOOS=windows GOARCH=amd64 go build ./internal/cli/` and `go vet ./internal/cli/ ./internal/agenticfactory/...`
print nothing.

## Baseline-attribution

Each line was produced in this run by `go test` compiled from the integration worktree at the commit named beside it
(coverage lines at the final tree); real git repositories and the real ledger were used for F1–F4. The auditor's
reproduction files were copied in temporarily to read their results and removed before any commit.

## Gaps

- The `-race` run over `internal/agenticfactory/...` was taken before the merge of `5190695ad3`; after it, `-race` was
  run for mergeq and ledger/... only.
- The Git Bash sentence in the help is this lane's reading of lane-19's point 2 (its text was not available here);
  the process-group limit it states is a code fact, not an observed Git Bash run.
- F8–F12 are recorded as debt by the leader and not touched.
- Native Windows is compiled, not run. `internal/web`, `internal/homestate`, `internal/hook`, `internal/template` and
  the whole `internal/cli` package were not re-run.

## Addendum — lane-19 37th review (non-blocking 1–4)

Commit graph: RED `080a9f29b6` → GREEN `9b8c929043`. Base: `8c58e97f30`.

RED at `080a9f29b6` (anchored, `-v`; long reasons shortened with `...`):

```
    the reported reason keeps the control character U+001B: "gates failed: gate ...: exit status 1: before\n\x1b[31mred..."
    the ledger's fail reason keeps the control character U+001B: ...
--- FAIL: TestAFLand_ControlCharactersInAFailedGatesOutputAreNotKept (0.33s)
    the help of `af land` lacks "WSL (or Git Bash, without process-group kill)":
    the help of `af land` lacks "prints without end is not an error: it runs until --gate-timeout":
--- FAIL: TestAFLand_HelpSaysWhatWindowsAndAnEndlessGateGet (0.00s)
    withdrawn_requeue_test.go:35: the row is verifying, want waiting (a verifying row cannot be queued again)
--- FAIL: TestLand_RealLedger_AWithdrawnCardsRowGoesBackToWaitingAndTheCardCanBeQueuedAgain (0.34s)
```

GREEN at `9b8c929043` (same commands):

```
--- PASS: TestAFLand_ControlCharactersInAFailedGatesOutputAreNotKept (0.34s)
--- PASS: TestAFLand_HelpSaysWhatWindowsAndAnEndlessGateGet (0.00s)
--- PASS: TestLand_RealLedger_AWithdrawnCardsRowGoesBackToWaitingAndTheCardCanBeQueuedAgain (0.43s)
--- PASS: TestLand_RealLedger_ACardWithdrawnDuringTheGatesIsNotLanded (0.61s)
```

At the final tree: `go test -count=1 -cover -p 2 ./internal/agenticfactory/mergeq/` →
`ok ... 30.620s coverage: 85.6% of statements`; `-run '^(TestAF|TestCardFlow)' ./internal/cli/` → `ok 14.4s`, no
`--- FAIL`; `go test -count=1 -race -p 2 ./internal/agenticfactory/...` → 21 packages `ok`; `go build ./...`,
`GOOS=windows GOARCH=amd64 go build ./internal/cli/`, `go vet ./internal/cli/ ./internal/agenticfactory/...` print nothing.

Gaps of the addendum: the control-character test covers escape, CR, bell, backspace, DEL and one C1 byte pair, not
every code point; the help sentences are asserted as text, not rendered by a terminal. (4) states that an endless
gate runs until `--gate-timeout` — the 20m default was not run.

## Residual-risk

A park that the ledger lets a card carry in a stage other than MERGING is refused by the stage check; a park resolved
between the check and the landing is covered by the ledger's own refusal at verifying and landing (lane-18). The
gate runs with `sh -c` and the caller's environment.
