# Verdict — t1915 (tuiFake.drive goroutine leak, Class B · P2)

Tree: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1915 — branch WT-managed-codex-tui,
base 8249bff8c3, commit **7e947d975a**, working tree clean after commit.

## Claim

1. The `(*tuiFake).drive` helper leaked one driver goroutine per unreleased caller:
   its cleanup closed only the idle-ticker stop channel, while the goroutine running
   `driveManagedFactorySession` parks in the idle select
   (internal/cli/managed_factory_session.go:349) forever — Run A census goroutine 942371.
2. Of the 10 `f.drive(...)` call sites in internal/cli/managed_codex_tui_test.go, exactly
   one ends without releasing the driver — TestManagedCodexTUIKeepsTerminalClean
   (line ~1213; `sess.stopTUI()` mid-test sets `stopping`, after which the TUI's
   `finish`/`done` can never fire). The other 9 call sites release in their bodies
   (`f.say("bye\n/exit\n")` + `tuiAwait`, or the consecutive-failure ceiling).
3. The repair (test code only; production unchanged): the drive cleanup releases the
   driver the way a real shutdown would and reaps it bounded; a still-parked driver
   fails the test. New test TestTuiFakeDriveTerminatesAfterCleanup pins it.

## Evidence

### RED (before the repair; only the new test + bound constant existed)

Command: `TMPDIR=/tmp go test ./internal/cli -run 'TestTuiFakeDriveTerminatesAfterCleanup' -count=1 -p 1 -v`

(An earlier identical invocation piped through `tail` produced the same FAIL lines but
its exit code was lost to a zsh/bash PIPESTATUS idiom mismatch — Gaps #2.)

Verbatim (second invocation, exit-attributed, TEST_EXIT:1):

```
=== RUN   TestTuiFakeDriveTerminatesAfterCleanup
    managed_codex_tui_test.go:856: the drive goroutine leaked: still parked 2s after cleanup (Run A census, goroutine 942371)
--- FAIL: TestTuiFakeDriveTerminatesAfterCleanup (2.49s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	3.202s
FAIL
```

Exit code observed: 1. (saved: .moai/reports/t1915/red-run.log)

### GREEN (repair applied)

Command: `TMPDIR=/tmp go test ./internal/cli -run 'TestTuiFakeDriveTerminatesAfterCleanup' -count=1 -p 1 -v`

Verbatim (TEST_EXIT:0):

```
=== RUN   TestTuiFakeDriveTerminatesAfterCleanup
--- PASS: TestTuiFakeDriveTerminatesAfterCleanup (0.47s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.336s
```

Exit code observed: 0. (saved: .moai/reports/t1915/green-run.log)

### Existing tuiFake-driven tests, name-pinned

Command: `TMPDIR=/tmp go test ./internal/cli -run 'TestTuiFakeDriveTerminatesAfterCleanup|TestManagedCodexTUIKeepsTerminalClean|TestManagedCodexTUIDefersDeliveryWhileBusy|TestManagedCodexReaderSurvivesOperatorTurns|TestManagedCodexBusyLongTurnKeepsDeferring|TestManagedCodexTUIConsumesResumedThreadReplayFrames|TestManagedCodexTUIExitEndsSession|TestManagedCodexSessionEndStopsTUI' -count=1 -p 1 -v`

Count: 9 top-level tests (8 existing drive-using tests + the new one; 19 rows with
subtests — every `f.drive(...)` call site in the file is inside these 8). One run.

Verbatim tail (TEST_EXIT:0):

```
--- PASS: TestTuiFakeDriveTerminatesAfterCleanup (0.48s)
--- PASS: TestManagedCodexTUIKeepsTerminalClean (0.90s)
--- PASS: TestManagedCodexTUIDefersDeliveryWhileBusy (1.78s)
--- PASS: TestManagedCodexReaderSurvivesOperatorTurns (0.38s)
--- PASS: TestManagedCodexBusyLongTurnKeepsDeferring (1.05s)
--- PASS: TestManagedCodexTUIConsumesResumedThreadReplayFrames (1.32s)
--- PASS: TestManagedCodexTUIExitEndsSession (2.86s)
--- PASS: TestManagedCodexSessionEndStopsTUI (1.11s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	10.712s
```

Every listed row PASS, no FAIL, no SKIP. (saved: .moai/reports/t1915/pinned-run.log)

### Commit

```
[WT-managed-codex-tui 7e947d975a] test(cli): reap the tuiFake drive goroutine in cleanup (leak from Run A census)
 1 file changed, 67 insertions(+), 1 deletion(-)
```

`git status --short` after commit: empty (clean). gofmt -l on the file: empty (exit 0).

## Baseline-attribution

- All three runs: this worktree (HEAD base 8249bff8c3 → commit 7e947d975a), `-count=1`
  so no cached reuse, `-p 1`, TMPDIR=/tmp. Commands and verbatim outputs above are from
  this session's runs against this tree; the RED run predates the repair edit, the GREEN
  and pinned runs postdate it.
- The defect's premise (which callers leak) is grounded in reading every
  `f.drive(...)` call site (lines 1213, 1274, 1311, 1326, 1351, 1587, 1774, 1823, 1865,
  1935) and the production release paths: `managedCodexTUI.finish` (done, once),
  `stop()` setting `stopping` (suppresses finish), `readManagedOperatorInput` (stdin
  EOF does not end the driver — comment at managed_factory_session.go:287-289).

## Gaps

1. The first RED invocation's exit code was not observed (pipeline idiom consumed it);
   the second identical invocation supplied the exit-attributed RED observation. One
   repeat, disclosed here.
2. The pinned run covers the drive-using tests of this file; the rest of
   ./internal/cli was not run (scope-to-the-change discipline; CI runs the full suite).
3. No goroutine-census diff (before/after runtime.NumGoroutine) was taken for the whole
   file run; the witness is per-test (the driver's own result channel), which proves
   termination of the repaired goroutine but not the absence of every other
   goroutine-shape in the package.

## Residual-risk

- The headless-driver shape (no TUI attached) is released by the same cleanup via the
  `/exit` stdin line, but no current call site exercises it — the fallback's headless
  path is untested by this change (all 10 sites attach a TUI).
- The reaper's `bye\n/exit\n` write lands in the pipe buffer when no reader exists
  (EPIPE only if all read ends close; newTUIFake's close cleanup runs last) — if a
  future caller closes stdin before test end, the write errors silently (ignored), and
  the driver release falls to the finish fallback or the bounded failure.

Kill discipline: no pattern kills used; no process was killed in this repair
(release, not force). The `kill -STOP/-CONT` calls seen in existing tests are theirs,
untouched.
