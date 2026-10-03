# Run evidence: F2 and F3 test reinforcement (SPEC-HARNESS-RETENTION-HARDEN-001, card t1432)

Scope: two new tests, no production or SPEC change. Tree HEAD at the start of the run: `9d9bc1a0b`, branch `WT-harness-retention-debt`, clean status.

## Claim

1. `TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal` (`internal/harness/retention_owner_test.go`) pins the wiring from `healStateEntry` into `removeStateEntryIfUnchanged` (sync-audit F2, mutant L4). It passes on the current code and fails on mutant L4.
2. `TestPruneTailAlreadyTerminatedGetsNoExtraNewline` (`internal/harness/retention_tail_test.go`) pins that an already newline-terminated tail gains no extra newline, by raw-byte comparison (sync-audit F3, mutant U7). It passes on the current code and fails on mutant U7.
3. Each new test is race-clean over 15 runs and the whole package stays green.

## Evidence

Unmutated tree, the two new tests (`go test -count=1 -v -run '<two names>' ./internal/harness/`):

```
=== RUN   TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal
=== PAUSE TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal
=== RUN   TestPruneTailAlreadyTerminatedGetsNoExtraNewline
=== PAUSE TestPruneTailAlreadyTerminatedGetsNoExtraNewline
=== CONT  TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal
=== CONT  TestPruneTailAlreadyTerminatedGetsNoExtraNewline
--- PASS: TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal (0.01s)
--- PASS: TestPruneTailAlreadyTerminatedGetsNoExtraNewline (0.32s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness	1.117s
exit=0
```

Mutant L4 (`healStateEntry` calls `os.Remove(statePath)` instead of `removeStateEntryIfUnchanged`), built as a scratch copy of the current `retention.go` outside the tree and run through `go test -overlay` with a literal-path overlay JSON:

```
=== NAME  TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal
    retention_owner_test.go:227: the fresh state file was removed or replaced by the heal: err=<nil>
--- FAIL: TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal (0.01s)
--- PASS: TestPruneTailAlreadyTerminatedGetsNoExtraNewline (0.31s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/harness	1.092s
FAIL
exit=1
```

Mutant U7 (the tail terminator is appended unconditionally), same overlay method:

```
--- PASS: TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal (0.00s)
=== NAME  TestPruneTailAlreadyTerminatedGetsNoExtraNewline
    retention_tail_test.go:133: pruned log bytes = "{...\"subject\":\"fresh\"...}\n{...\"subject\":\"late-event\"...}\n\n", want "{...\"subject\":\"fresh\"...}\n{...\"subject\":\"late-event\"...}\n" (the kept line plus the late bytes, no extra blank line)
--- FAIL: TestPruneTailAlreadyTerminatedGetsNoExtraNewline (0.33s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/harness	1.128s
FAIL
exit=1
```

(The U7 assertion line is abbreviated with `...` in this file only; the run printed the full JSON. The pruned log ends with `\n\n`, the wanted bytes with a single `\n`.)

Whole package, unmutated: `go test -count=1 ./internal/harness/`

```
ok  	github.com/modu-ai/moai-adk/internal/harness	2.885s
exit=0
```

Race, the two new tests: `go test -race -count=15 -run '<two names>' ./internal/harness/`

```
ok  	github.com/modu-ai/moai-adk/internal/harness	9.221s
exit=0
```

Static checks: `go vet ./internal/harness/ ./internal/lockfile/` exit 0; `gofmt -l internal/harness/` printed nothing, exit 0; `GOOS=windows go build ./internal/harness/ ./internal/lockfile/` exit 0; `GOOS=windows go vet ./internal/harness/ ./internal/lockfile/` exit 0.

Every test run was wrapped as `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ...` in one invocation, under the `go-test-internal-harness` slot lease (acquired before, released after).

## Baseline-attribution

All figures above were measured in this run against the worktree `.moai/worktrees/t1432` at HEAD `9d9bc1a0b` plus the two uncommitted test edits. The mutants were scratch copies of the `retention.go` at that HEAD, applied through `go test -overlay`; the tree never held a mutant. Judging build: the Go toolchain compiled the tree directly, no installed project binary judged it. The L4 and U7 identifiers and survival come from `.moai/reports/t1432/sync-audit.md` (F2, F3); the kill results here are fresh measurements, not carried over.

## Gaps

- Linux is unobserved: only macOS (darwin) was run. On Linux the L4 test compares file identity (`os.SameFile`) and the new file created after the mutant's removal could in principle reuse the removed inode; the test also compares content (the stamp written by the mutant differs from the fresh bytes) but that path was not exercised on Linux.
- Windows is unobserved at runtime: both new tests carry `//go:build !windows` through their host files, so only compile/vet of the package for Windows was observed, not a test run.
- Only mutants L4 and U7 were re-run against the new tests; the other survivors of the sync audit (F4 to F6 and the unlisted ones) were not re-measured.
- The new tests were not run in the CI environment or under load beyond the `-race -count=15` run.
- The L4 failure message reads `err=<nil>` because the identity check fails without a Lstat error; the message names the symptom but not the identity mismatch.

## Residual-risk

- The F2 test interposes through the `ownerCheck` function field, which runs before the removal inside the heal. A refactor that moved the owner check after the removal would make the test pass vacuously or fail for a different reason; the test guards against that only through its "owner check never ran" fatal.
- The F3 test depends on `blockedPruner` timing (a 300 ms settle delay), the same dependency the existing tail tests carry.
- The heal check-then-remove window itself (recorded in the SPEC as residual risk) remains: the test pins the conditional removal, not an atomic one.
