# t1432 M0 — observed RED baseline (SPEC-HARNESS-RETENTION-HARDEN-001)

Measured in this run on the card worktree `.moai/worktrees/t1432`, branch `WT-harness-retention-debt`, HEAD `4934c4362` (production Go code identical to base `1e2151a38`: `git diff --quiet 1e2151a38 HEAD -- internal cmd` exit 0). The three new test files were untracked working-tree files on top of `4934c4362` at measurement time; commit T adds exactly those files plus this report, so T's `internal` tree equals the measured tree. Platform darwin/arm64, uid 501, go1.26.8. Every `go test` was run as `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && <command>` under the lease `go-test-internal-harness`; that prefix is not part of the cited command. Exit codes are the shell's `$?` of the command, recorded as their own field. Judging build: the Go toolchain compiled the tree under test, no project tool (`moai`) was used for any measurement.

Pre-conditions: `go build ./internal/harness/` exit 0; `go vet ./internal/harness/` exit 0; `GOOS=windows go vet ./internal/harness/ ./internal/lockfile/` exit 0 (with the three new files present; `retention_tail_test.go` carries `//go:build !windows`).

## 1. RED at the unmodified base (no overlay)

### AC-HRH-001 — flips at M2

```
command: go test -count=1 -v -run '^TestPruneStateSymlinkReplacedTargetUntouched$' ./internal/harness/
stdout:
=== RUN   TestPruneStateSymlinkReplacedTargetUntouched
=== PAUSE TestPruneStateSymlinkReplacedTargetUntouched
=== CONT  TestPruneStateSymlinkReplacedTargetUntouched
    retention_statepath_test.go:49: victim changed: err=<nil> content="2026-10-02T00:00:00Z"
    retention_statepath_test.go:53: state path is not a regular file: err=<nil>
--- FAIL: TestPruneStateSymlinkReplacedTargetUntouched (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/harness	0.716s
FAIL
exit: 1
tree: 4934c4362
flips at: M2
```

### AC-HRH-002 — flips at M2

```
command: go test -count=1 -v -run '^TestPruneStateUnwritableFileReplaced$' ./internal/harness/
stdout:
=== RUN   TestPruneStateUnwritableFileReplaced
=== PAUSE TestPruneStateUnwritableFileReplaced
=== CONT  TestPruneStateUnwritableFileReplaced
    retention_statepath_test.go:100: first prune returned retention: prune state open failed: open /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestPruneStateUnwritableFileReplaced2595229719/001/usage-log.jsonl.prune-state: permission denied, want nil
    retention_statepath_test.go:103: first stale event still in the log
    retention_statepath_test.go:112: second prune returned retention: prune state open failed: open /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestPruneStateUnwritableFileReplaced2595229719/001/usage-log.jsonl.prune-state: permission denied, want nil
    retention_statepath_test.go:115: second stale event still in the log
    retention_statepath_test.go:120: state file not openable read-write: open /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestPruneStateUnwritableFileReplaced2595229719/001/usage-log.jsonl.prune-state: permission denied
--- FAIL: TestPruneStateUnwritableFileReplaced (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/harness	0.492s
FAIL
exit: 1
tree: 4934c4362
flips at: M2
```

### AC-HRH-003 case b — green at base (pin)

```
command: go test -count=1 -v -run '^TestPruneStateUnreplaceableInReadOnlyDirSkips$' ./internal/harness/
stdout:
=== RUN   TestPruneStateUnreplaceableInReadOnlyDirSkips
--- PASS: TestPruneStateUnreplaceableInReadOnlyDirSkips (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness	0.520s
exit: 0
tree: 4934c4362
flips at: stays green (regression guard; mutation proof at M2)
```

Case a is the existing test, observed green at base:

```
command: go test -count=1 -v -run '^TestPruneStamp_StateNotOpenableSkipsPruneAndRecordSucceeds$' ./internal/harness/
stdout:
=== RUN   TestPruneStamp_StateNotOpenableSkipsPruneAndRecordSucceeds
=== PAUSE TestPruneStamp_StateNotOpenableSkipsPruneAndRecordSucceeds
=== CONT  TestPruneStamp_StateNotOpenableSkipsPruneAndRecordSucceeds
--- PASS: TestPruneStamp_StateNotOpenableSkipsPruneAndRecordSucceeds (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness	0.553s
exit: 0
tree: 4934c4362
```

### AC-HRH-007 — flips at M1

```
command: go test -count=1 -v -run '^TestPruneCarriesLateEvents$' ./internal/harness/
stdout:
=== RUN   TestPruneCarriesLateEvents
=== PAUSE TestPruneCarriesLateEvents
=== CONT  TestPruneCarriesLateEvents
    retention_tail_test.go:80: late-event count = 0, want 1 (subjects after prune: map[fresh:1])
    retention_tail_test.go:90: want fresh then the late event verbatim, got ["{\"timestamp\":\"2026-10-01T00:00:00Z\",\"event_type\":\"feedback\",\"subject\":\"fresh\",\"context_hash\":\"h\",\"tier_increment\":0,\"schema_version\":\"v2.1\"}"]
--- FAIL: TestPruneCarriesLateEvents (0.31s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/harness	0.712s
FAIL
exit: 1
tree: 4934c4362
flips at: M1
```

### AC-HRH-008 case b — flips at M1

```
command: go test -count=1 -v -run '^TestPruneTailPartialLineCarriedAndTerminated$' ./internal/harness/
stdout:
=== RUN   TestPruneTailPartialLineCarriedAndTerminated
=== PAUSE TestPruneTailPartialLineCarriedAndTerminated
=== CONT  TestPruneTailPartialLineCarriedAndTerminated
    retention_tail_test.go:122: partial fragment occurs 0 times, want 1
--- FAIL: TestPruneTailPartialLineCarriedAndTerminated (0.31s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/harness	0.694s
FAIL
exit: 1
tree: 4934c4362
flips at: M1
```

### AC-HRH-008 case c — flips at M1

```
command: go test -count=1 -v -run '^TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval$' ./internal/harness/
stdout:
=== RUN   TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval
=== PAUSE TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval
=== CONT  TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval
    retention_tail_test.go:164: stale-final count after first prune = 0, want 1 (kept in the log this interval; subjects: map[fresh:1])
--- FAIL: TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval (0.00s)
FAIL
FAIL	github.com/modu-ei/moai-adk/internal/harness	0.510s
FAIL
exit: 1
tree: 4934c4362
flips at: M1
```

### AC-HRH-008 case a — green at base (pin, existing tests)

```
command: go test -count=1 -v -run '^(TestPruneKeepsUnparsedLinesVerbatim|TestPruneNothingStaleLeavesLogUntouched)$' ./internal/harness/
stdout (the run/pause/cont preamble lines are elided; the verdict lines are verbatim):
--- PASS: TestPruneNothingStaleLeavesLogUntouched (0.00s)
--- PASS: TestPruneKeepsUnparsedLinesVerbatim (0.00s)
    --- PASS: TestPruneKeepsUnparsedLinesVerbatim/wrong-field-type (0.00s)
    --- PASS: TestPruneKeepsUnparsedLinesVerbatim/json-array (0.00s)
    --- PASS: TestPruneKeepsUnparsedLinesVerbatim/plain-text (0.01s)
    --- PASS: TestPruneKeepsUnparsedLinesVerbatim/truncated-json (0.01s)
    --- PASS: TestPruneKeepsUnparsedLinesVerbatim/padded-text (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness	0.417s
exit: 0
tree: 4934c4362
```

### AC-HRH-013 — green at base (test-gap criterion; kill is the mutant run in section 2)

```
command: go test -count=1 -v -run '^TestPruneStampShorterOverLongerIsExact$' ./internal/harness/
stdout:
=== RUN   TestPruneStampShorterOverLongerIsExact
=== PAUSE TestPruneStampShorterOverLongerIsExact
=== CONT  TestPruneStampShorterOverLongerIsExact
--- PASS: TestPruneStampShorterOverLongerIsExact (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness	0.520s
exit: 0
tree: 4934c4362
flips at: lands green in M0; fails under mB (section 2)
```

Swept-count reading: every command above printed exactly one top-level `--- PASS` or `--- FAIL` line per named test. Of the seven new tests, five FAIL (AC-001, -002, -007, -008 b, -008 c) and two PASS (AC-003 b, AC-013), as the plan expected. No `[no tests to run]`, no `--- SKIP` and no timeout appeared in this section.

| AC | test | exit | result at base | tree | flips at |
|---|---|---|---|---|---|
| AC-HRH-001 | TestPruneStateSymlinkReplacedTargetUntouched | 1 | FAIL (assertion: victim changed, path still a link) | 4934c4362 | M2 |
| AC-HRH-002 | TestPruneStateUnwritableFileReplaced | 1 | FAIL (assertion: permission denied, events not archived) | 4934c4362 | M2 |
| AC-HRH-003 b | TestPruneStateUnreplaceableInReadOnlyDirSkips | 0 | PASS (pin) | 4934c4362 | stays green |
| AC-HRH-007 | TestPruneCarriesLateEvents | 1 | FAIL (late-event count = 0) | 4934c4362 | M1 |
| AC-HRH-008 b | TestPruneTailPartialLineCarriedAndTerminated | 1 | FAIL (fragment occurs 0 times) | 4934c4362 | M1 |
| AC-HRH-008 c | TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval | 1 | FAIL (stale-final archived at once) | 4934c4362 | M1 |
| AC-HRH-013 | TestPruneStampShorterOverLongerIsExact | 0 | PASS (test-gap criterion) | 4934c4362 | M0 (kill under mB) |

## 2. Mutant re-observation (overlays name committed copies in `.moai/reports/t1432/red-now-drafts/`; production code is the base)

The overlay JSON files in that directory already name this worktree's committed sources (verified by reading them); no regeneration was needed. `diff internal/harness/retention.go <mutant>` observed: mB deletes lines 210-212 (the `Truncate(0)` block); mC replaces lines 142-144 by `_ = writeStamp(sf, now)`; mD replaces line 146 `err = r.prune(retentionDays, now)` by `err = nil`.

The plan-phase overlay `overlay-mB-newtest.json` injects the draft `zz_stampbytes_test.go`, whose test name now collides with the committed `retention_stampbytes_test.go` (a redeclaration compile error), so it was NOT used. The kill was observed with `overlay-mB.json` against the tree-resident test instead. For the whole-package survivor runs the seven new tests are excluded with `-skip` (five of them are RED at base by design and would hide the survivor result); this is the in-tree equivalent of the plan-phase whole-package runs E-006 and E-008, which ran before the new tests existed.

### AC-HRH-012 — mD, killed-pruner test hangs to the alarm (RED; flips at M4)

```
command: go test -count=1 -v -timeout 40s -overlay .moai/reports/t1432/red-now-drafts/overlay-mD.json -run '^TestPruneStamp_StampExistsBeforeTheWork$' ./internal/harness/
stdout (redirected to the session scratch file; first 12 lines and the deciding lines verbatim):
=== RUN   TestPruneStamp_StampExistsBeforeTheWork
=== PAUSE TestPruneStamp_StampExistsBeforeTheWork
=== CONT  TestPruneStamp_StampExistsBeforeTheWork
panic: test timed out after 40s
	running tests:
		TestPruneStamp_StampExistsBeforeTheWork (40s)

goroutine 83 [running]:
testing.(*M).startAlarm.func1()
...
syscall.Open({0x329148d5f480?, 0x1007977c0?}, 0x1000000, 0x0)
...
	.../internal/harness/retention_killed_test.go:43 +0x28
...
	.../internal/harness/retention_killed_test.go:85 +0x62c
FAIL	github.com/modu-ai/moai-adk/internal/harness	40.554s
FAIL
exit: 1
tree: 4934c4362 (overlay: retention_mD.go)
```

### AC-HRH-013 — mB survivor, then kill

```
command: go test -count=1 -v -overlay .moai/reports/t1432/red-now-drafts/overlay-mB.json -skip '^(TestPruneStateSymlinkReplacedTargetUntouched|TestPruneStateUnwritableFileReplaced|TestPruneStateUnreplaceableInReadOnlyDirSkips|TestPruneCarriesLateEvents|TestPruneTailPartialLineCarriedAndTerminated|TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval|TestPruneStampShorterOverLongerIsExact)$' ./internal/harness/
stdout (redirected to a scratch file, read after): 266 top-level `--- PASS` lines, 0 `--- FAIL`, 1 `--- SKIP` (TestPruneHelperProcess, a pre-existing helper); last lines:
--- PASS: TestGoMeasurer_BuildError_FailsClosed (0.11s)
PASS
ok  	github.com/modu-ei/moai-adk/internal/harness	1.091s
exit: 0   (the existing suite is blind to mB: SURVIVOR)
tree: 4934c4362 (overlay: retention_mB.go)
```

```
command: go test -count=1 -v -overlay .moai/reports/t1432/red-now-drafts/overlay-mB.json -run '^TestPruneStampShorterOverLongerIsExact$' ./internal/harness/
stdout:
=== RUN   TestPruneStampShorterOverLongerIsExact
=== PAUSE TestPruneStampShorterOverLongerIsExact
=== CONT  TestPruneStampShorterOverLongerIsExact
    retention_stampbytes_test.go:35: state file = "2026-10-02T00:00:00Z123456789Z", want exactly "2026-10-02T00:00:00Z"
    retention_stampbytes_test.go:38: state file does not parse as a fresh stamp: "2026-10-02T00:00:00Z123456789Z"
--- FAIL: TestPruneStampShorterOverLongerIsExact (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/harness	0.577s
FAIL
exit: 1   (the new test kills mB)
tree: 4934c4362 (overlay: retention_mB.go)
```

### AC-HRH-014 — mC survivor (kill needs the extracted locked phase, M3; Gap G-2 stays open)

```
command: go test -count=1 -v -overlay .moai/reports/t1432/red-now-drafts/overlay-mC.json -skip '^(TestPruneStateSymlinkReplacedTargetUntouched|TestPruneStateUnwritableFileReplaced|TestPruneStateUnreplaceableInReadOnlyDirSkips|TestPruneCarriesLateEvents|TestPruneTailPartialLineCarriedAndTerminated|TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval|TestPruneStampShorterOverLongerIsExact)$' ./internal/harness/
stdout (redirected to a scratch file, read after): 266 top-level `--- PASS` lines, 0 `--- FAIL`, 1 `--- SKIP` (TestPruneHelperProcess); last lines:
    --- PASS: TestPruneSweepsOldOrphanTmp/log-with-stale-event (0.03s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness	0.895s
exit: 0   (SURVIVOR)
tree: 4934c4362 (overlay: retention_mC.go)
```

## 3. AC-HRH-010 baseline greps and the card-id guard

The shell tool did not display the grep exit code on its own; each command was followed by `echo "exit=$?"` and the value below is that echo. For `grep -c`, exit 0 means at least one line matched and exit 1 means none, so the exit code is derivable from the count (count 0 gives exit 1, count at least 1 gives exit 0) and the two agree on every line below.

```
command: grep -c -F -e "residual window" -e "no cross-process exclusion" -e "burst of hook processes" -e "F5: not reproduced, not measured" -e "lock waiters block with no timeout" -e "5 s hook timeout" -e "appended before the wait" -e "F6: not reproduced, not measured" internal/harness/retention.go
stdout: 0
exit: 1
tree: 4934c4362   (E-010 form; flips at M5)
```

```
command: grep -c -F "events other hooks append in that window are lost" internal/harness/retention.go
stdout: 1
exit: 0
tree: 4934c4362   (E-011 form; the old sentence is removed at M5)
```

```
command: grep -c -E 't1[0-9]{3}' internal/harness/retention.go
stdout: 0
exit: 1
tree: 4934c4362   (card-id guard count at base; must stay 0)
```

## Gaps

- Linux and Windows runtime behaviour is not observed (G-5 stands): all runs darwin/arm64, uid 501. The Windows evidence is `GOOS=windows go vet` only.
- AC-HRH-004, -005, -006 and -014 have no RED-now by design (G-1, G-2); they are not part of M0.
- The exit-code derivation for the three greps is stated above; the tool itself did not print the exit code of a bare grep.
- The mutant runs use overlay files inside the worktree's `.moai/reports/t1432/red-now-drafts/` directory (they replace `retention.go` only virtually; no production file was edited), not a location outside the tree.
- The plan-phase kill command E-007 (`overlay-mB-newtest.json`) cannot be reused as is once the committed test exists (duplicate test name); replaced by the in-tree kill above.
- The two whole-package survivor runs exclude the seven new tests by `-skip`; the other 266 top-level tests passed.
- The FIFO hang at the state path (E-025) is not re-measured here (not a criterion).

## Residual-risk

- The 300 ms bounded delay in the tail tests (G-4) can only make a test pass vacuously on a slow runner, never fail falsely; the RED observed here ran with the delay satisfied.
- AC-HRH-002's identity assertion depends on holding the original inode open; on a filesystem that reuses a freed inode the held handle keeps this decisive, but the assertion has not been run on Linux.
