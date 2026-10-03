# Run-phase verdict — SPEC-HARNESS-RETENTION-HARDEN-001 (card t1432, Tier M)

Author: manager-develop, run phase M4 to M6 (M0 to M3 evidence: `red-baseline.md`, `run-evidence-m1-m3.md`).
Platform of every run: darwin arm64, uid 501, go1.26.8. Linux and Windows runtime: not observed.
Tests ran with the kanban variables unset in the same invocation.

intermediate red commit T = ac40cf3bf

Commits on the card branch (base `1e2151a38`): `ac40cf3bf` (M0, T), `c1cc3fe67` (M1), `f52dd1b1c` (M2), `b3a469eab` (M3), `3badf7875` (M4), `9289a92b6` (M5). This file and its evidence land in the commit that follows `9289a92b6`.

## 1. Claim

1. All 14 acceptance criteria are met on the final tree (`9289a92b6`, clean working tree), with the qualifications under Gaps.
2. Every named test was swept (one `--- PASS` line each) and the whole package passes under `-race`.
3. Every mutant of `plan.md` §F M1, M2 and N2/N1 (15, see the mutation table) fails its named test on the final tree.
4. `internal/lockfile` and `observer.go` are byte-identical to the base, and the four pre-lock function bodies have no commit in `ac40cf3bf..HEAD`.
5. Windows is verified by build and vet only. Windows runtime not observed.

## 2. Evidence

Each row: command, verbatim output, exit code as its own field. Tree SHA for every row: `9289a92b6` unless stated. Long outputs were redirected to files in the session scratch directory and are quoted by their deciding lines.

### 2.1 Pre-flight (tree `b3a469eab`, before M4)

| Command | Output | exit |
|---|---|---|
| `go build ./internal/harness/` | (empty) | 0 |
| `go vet ./internal/harness/` | (empty) | 0 |
| `GOOS=windows go vet ./internal/harness/ ./internal/lockfile/` | (empty) | 0 |
| `go test -count=1 -v ./internal/harness/` | 277 lines `--- PASS`; last line `ok  	github.com/modu-ai/moai-adk/internal/harness	1.067s` | 0 |

### 2.2 M4 (AC-HRH-012)

Mechanism chosen: the blocking read open runs in a goroutine; after a 10 s wait the goroutine is released by opening the FIFO write side with `O_WRONLY|O_NONBLOCK` (that open succeeds only while a reader is blocked in its open), and the test then fails with a message naming the missing archive step. The non-blocking read open was not used: `O_RDONLY|O_NONBLOCK` returns at once, before a pruner has reached its write open, so the reader could see end-of-file early and the pruner would then block on its write open with no reader (the race named in `spec.md` §F). The goroutine form needs no ordering between the two sides.

Mutant mD for this run is built from the CURRENT `retention.go`: line `err := r.prune(retentionDays, now)` in `pruneLocked` replaced by `var err error` (scratch copy `scratchpad/m6_mD.go`, overlay `ov_m6_mD.json`; `diff` showed only `161c161`).

BEFORE (unmodified `drain`, mutant mD, tree `b3a469eab`):

```
command: go test -count=1 -timeout 40s -overlay <ov_m6_mD.json> -run '^TestPruneStamp_StampExistsBeforeTheWork$' ./internal/harness/
panic: test timed out after 40s
	running tests:
		TestPruneStamp_StampExistsBeforeTheWork (40s)
...
syscall.Open({0x698a65224c80?, 0x10045b7c0?}, 0x1000000, 0x0)
...
	.../internal/harness/retention_killed_test.go:43 +0x28
FAIL	github.com/modu-ai/moai-adk/internal/harness	40.515s
FAIL
exit: 1
```

AFTER (fixed `drain`, same mutant, timed from outside with `timeout 30`):

```
command: timeout 30 go test -count=1 -v -timeout 40s -overlay <ov_m6_mD.json> -run '^TestPruneStamp_StampExistsBeforeTheWork$' ./internal/harness/
=== RUN   TestPruneStamp_StampExistsBeforeTheWork
=== PAUSE TestPruneStamp_StampExistsBeforeTheWork
=== CONT  TestPruneStamp_StampExistsBeforeTheWork
    retention_killed_test.go:119: the pruner never reached its archive step: nothing opened the archive FIFO .../archive/2026-08.jsonl.gz for writing within 10s
--- FAIL: TestPruneStamp_StampExistsBeforeTheWork (10.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/harness	10.596s
FAIL
exit: 1
```

Wall time of the whole invocation (build included) was 13.0 to 13.9 s (`time` output), under the 15 s bound; the test itself ran 10.01 s. No `panic: test timed out` in the output.

Unmutated (tree `b3a469eab` plus the uncommitted M4 edit):

```
command: go test -count=1 -v -run '^TestPruneStamp_StampExistsBeforeTheWork$' ./internal/harness/
--- PASS: TestPruneStamp_StampExistsBeforeTheWork (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness	0.554s
exit: 0
```

Whole package after M4: 277 `--- PASS`, 0 `--- FAIL`, `ok ... 1.051s`, exit 0. The M4 commit is `3badf7875`.

### 2.3 M5 (AC-HRH-010)

Counts before the edit (tree `3badf7875`) and after (tree `9289a92b6`):

| Command (single `grep -c`) | before | after |
|---|---|---|
| `grep -c -F -e "residual window" -e "no cross-process exclusion" -e "burst of hook processes" -e "F5: not reproduced, not measured" -e "lock waiters block with no timeout" -e "5 s hook timeout" -e "appended before the wait" -e "F6: not reproduced, not measured" internal/harness/retention.go` | 0 | 7 (matching lines; `residual window` alone matches 2) |
| `grep -c -F "residual window" ...` | not run | 2 |
| `grep -c -F "no cross-process exclusion" ...` | not run | 1 |
| `grep -c -F "burst of hook processes" ...` | not run | 1 |
| `grep -c -F "F5: not reproduced, not measured" ...` | not run | 1 |
| `grep -c -F "lock waiters block with no timeout" ...` | not run | 1 |
| `grep -c -F "5 s hook timeout" ...` | not run | 1 |
| `grep -c -F "appended before the wait" ...` | not run | 1 |
| `grep -c -F "F6: not reproduced, not measured" ...` | not run | 1 |
| `grep -c -F "events other hooks append in that window are lost" internal/harness/retention.go` | 1 | 0 |
| `grep -c -E 't1[0-9]{3}' internal/harness/retention.go` | 0 | 0 |

Each of these printed its number; the exit code of a `grep -c` that prints 0 is 1, and the tool output showed only stdout, so the exit code was not displayed (the printed counts are the evidence). Non-comment diff: `git diff -U0 internal/harness/retention.go` held 23 lines beginning with `+` or `-` (two of them the `+++`/`---` file headers), and the count of those lines that are neither a header nor a `//` comment line, taken with `grep -c -v -E '^(\+\+\+|---|[+-]//)'`, printed `0`. Whole package after M5: 277 `--- PASS`, `ok ... 1.059s`, exit 0; `go vet` exit 0, `gofmt -l` empty, `GOOS=windows go build` and `go vet` on both packages exit 0. The M5 commit is `9289a92b6`. The reviewer-read part of AC-HRH-010 (the F5 and F6 sentences against REQ-HRH-010 and -011, including the `async` obligation) is not discharged by the sentinels; it is left for the reader.

### 2.4 M6 (a) tests, race, count=20, vet, gofmt, lint (tree `9289a92b6`)

| Command | Output (deciding lines) | exit |
|---|---|---|
| `go test -race -count=1 -v ./internal/harness/` | 277 `--- PASS`, 0 `--- FAIL`, 1 `--- SKIP: TestPruneHelperProcess` (pre-existing helper process); last lines `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/harness	6.794s` | 0 |
| `go vet ./internal/harness/ ./internal/lockfile/` | (empty) | 0 |
| `go test -count=20 -v -run '^(TestPruneStamp_StampExistsBeforeTheWork|TestPruneConcurrentProcessesSingleRewrite)$' ./internal/harness/` | 20 `--- PASS: TestPruneStamp_StampExistsBeforeTheWork`, 20 `--- PASS: TestPruneConcurrentProcessesSingleRewrite`, 0 FAIL or SKIP; `ok  	github.com/modu-ai/moai-adk/internal/harness	11.422s` | 0 |
| `gofmt -l internal/harness/ internal/lockfile/` | (empty) | 0 |
| `golangci-lint run --timeout=3m ./internal/harness/` | `0 issues.` | 0 |

Lint build: `golangci-lint has version v2.1.6 built with go1.26.8`. CI pins v2.1.6; this installed build was not compared byte for byte with the CI-pinned build, so it is stated as the same version string, not as the same build.

Named tests, swept in the race run (each line read from the output file):

```
--- PASS: TestPruneConcurrentProcessesSingleRewrite (4.67s)
--- PASS: TestPruneStateForeignOwnedLeftUntouchedAndWarns (0.00s)
--- PASS: TestOwnerCheckDefault (0.00s)
--- PASS: TestPruneStateUnreplaceableInReadOnlyDirSkips (0.00s)
--- PASS: TestPruneStampWriteFailureSkipsPrune (0.01s)
--- PASS: TestPruneStamp_FreshStampNeedsNoLock (0.01s)
--- PASS: TestHealDoesNotRemoveAFreshStateFile (0.01s)
--- PASS: TestPruneStamp_StateNotOpenableSkipsPruneAndRecordSucceeds (0.01s)
--- PASS: TestPruneStampShorterOverLongerIsExact (0.02s)
--- PASS: TestPruneStateSymlinkReplacedTargetUntouched (0.03s)
--- PASS: TestPruneStateUnwritableFileReplaced (0.04s)
--- PASS: TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval (0.05s)
--- PASS: TestPruneStamp_StampExistsBeforeTheWork (0.02s)
--- PASS: TestPruneNothingStaleLeavesLogUntouched (0.01s)
--- PASS: TestPruneKeepsUnparsedLinesVerbatim (0.00s)
--- PASS: TestPruneCarriesLateEvents (0.31s)
--- PASS: TestPruneTailPartialLineCarriedAndTerminated (0.31s)
```

No `--- SKIP` line belongs to any of them, so no skip was taken on uid 501 for these tests (including `TestOwnerCheckDefault` step (e), which found `/var` as the foreign-owned link in the M2 run).

### 2.5 M6 (b) Windows (tree `9289a92b6`)

| Command | Output | exit |
|---|---|---|
| `GOOS=windows go build ./internal/harness/ ./internal/lockfile/` | (empty) | 0 |
| `GOOS=windows go vet ./internal/harness/ ./internal/lockfile/` | (empty) | 0 |

Verification level: build and vet only. Windows runtime not observed.

### 2.6 M6 (c) boundary (AC-HRH-009, AC-HRH-011)

| Command | Output | exit |
|---|---|---|
| `git diff --quiet 1e2151a38 HEAD -- internal/lockfile internal/harness/observer.go` | (empty) | 0 |
| four-function `git log --format=%h -s -L '/^func (r \*Retention) PruneStaleEntries/,/^}/:internal/harness/retention.go' -L '/^func readStamp(/,/^}/:internal/harness/retention.go' -L '/^func readStampFile(/,/^}/:internal/harness/retention.go' -L '/^func stampIsFresh(/,/^}/:internal/harness/retention.go' ac40cf3bf..HEAD` | (empty) | 0 |
| same, range `1e2151a38..HEAD` | (empty) | 0 |
| positive control, same options, range `fe211e9c9~1..fe211e9c9` | `fe211e9c9` | 0 |
| positive control, `git diff --quiet 6dc31a727~1 6dc31a727 -- internal/harness/observer.go` | (empty) | 1 |
| positive control, out-of-order `-L` pair `readStampFile` then `readStamp` over `1e2151a38..HEAD` | `fatal: -L parameter '^func readStamp(' starting at line 328: regexec() failed to match` | 128 |

The empty `-L` outputs are read as a pass because the range `1e2151a38..HEAD` holds commits that touch the file (listing below) and the controls fired.

### 2.7 M6 (e) M0 Exit ordering check (card branch, tree `9289a92b6`)

| Command | Output | exit |
|---|---|---|
| `git log --diff-filter=A --format=%h -- internal/harness/retention_statepath_test.go` | `ac40cf3bf` | 0 |
| `git log --reverse --format=%h 1e2151a38..HEAD -- internal/harness/retention.go` | `c1cc3fe67` `f52dd1b1c` `b3a469eab` `9289a92b6` (non-empty) | 0 |
| `git merge-base --is-ancestor ac40cf3bf c1cc3fe67` | (empty) | 0 |
| `git merge-base --is-ancestor ac40cf3bf f52dd1b1c` | (empty) | 0 |
| `git merge-base --is-ancestor ac40cf3bf b3a469eab` | (empty) | 0 |
| `git merge-base --is-ancestor ac40cf3bf 9289a92b6` | (empty) | 0 |
| control, mis-ordered: `git merge-base --is-ancestor c1cc3fe67 ac40cf3bf` | (empty) | 1 |
| control, bounded form lists a hash on a range with a change: `git log --reverse --format=%h fe211e9c9~1..fe211e9c9 -- internal/harness/retention.go` | `fe211e9c9` | 0 |

The exit codes were read by running each command in its own invocation followed by `; echo "exit=$?"` in the same shell call (each command is a single `git` invocation, so `$?` is that command's status).

### 2.7a Mutation table (final tree `9289a92b6`)

Each mutant was generated by one script (`scratchpad/gen_mutants.py`) from the CURRENT `retention.go` or `retention_owner_unix.go`, asserting that each replacement matched exactly once; each is a scratch copy under `scratchpad/m6/` and an overlay JSON, applied with `go test -count=1 -v -overlay <json> -run '<test>' ./internal/harness/`; none was ever in the tree (`git status --short` empty afterwards). Every output was read for a `--- FAIL` line and checked against a build error (a grep for `build failed`, `cannot use`, `undefined:`, `declared and not used` over all 15 outputs matched nothing). All 15 exit codes are 1.

| # | Mutant | Criterion | Killing test | exit | Deciding failing output |
|---|---|---|---|---|---|
| 1 | tail copy without terminator (the `if tail[len(tail)-1] != '\n'` block removed) | AC-HRH-008 | `TestPruneTailPartialLineCarriedAndTerminated`, `TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval` | 1 | `replacement log does not end with a newline`; `the next append is not on its own line and parseable`; `rewritten log does not end with a newline` |
| 2 | classify the final unterminated line (the guard in `scanTerminatedLines` removed) | AC-HRH-008 | `TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval` | 1 | `stale-final count after first prune = 0, want 1 (kept in the log this interval; subjects: map[fresh:1])` |
| 3 | drop the tail copy (the `appendLogTail` call removed) | AC-HRH-007, -008 | `TestPruneCarriesLateEvents`, `TestPruneTailPartialLineCarriedAndTerminated`, `TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval` | 1 | `late-event count = 0, want 1`; `partial fragment occurs 0 times, want 1` |
| 4 | owner result ignored (`if false && !r.ownerCheck(...)`) | AC-HRH-004 | `TestPruneStateForeignOwnedLeftUntouchedAndWarns` | 1 | `want an error naming <state path>, got <nil>`; `state-path entry changed`; `want exactly one warning line ... got ""` |
| 5 | removal before the ownership check | AC-HRH-004 | same | 1 | `state-path entry changed: err=lstat ...: no such file or directory` |
| 6 | warning omitted | AC-HRH-004 | same | 1 | `want exactly one warning line naming <state path>, got ""` (both subtests) |
| 7 | owner check always "owned" | AC-HRH-005 | `TestOwnerCheckDefault` | 1 | `/ must not count as owned for a non-root user`; `/var is a symbolic link owned by another user and must not count as owned` |
| 8 | owner read follows the link (`os.Stat`) | AC-HRH-005 | `TestOwnerCheckDefault` | 1 | `a user-owned link to / must count as owned: the link's own owner decides, not its target's` |
| 9 | every symbolic link treated as owned | AC-HRH-005 | `TestOwnerCheckDefault` | 1 | `/var is a symbolic link owned by another user and must not count as owned` |
| 10 | unconditional removal in `removeStateEntryIfUnchanged` | AC-HRH-006 | `TestHealDoesNotRemoveAFreshStateFile` | 1 | `removal of a changed entry: removed=true err=<nil>, want false and nil`; `the fresh state file was removed or replaced` |
| 11 | a heal that removes a directory | AC-HRH-003 (a) | `TestPruneStamp_StateNotOpenableSkipsPruneAndRecordSucceeds` | 1 | `expected an error when the state file cannot be opened` |
| 12 | a heal that ignores the removal failure | AC-HRH-003 (b) | `TestPruneStateUnreplaceableInReadOnlyDirSkips` | 1 | `error does not wrap a permission error: retention: prune state entry ... changed on every inspection; prune skipped` |
| 13 | mB: the `Truncate(0)` block removed | AC-HRH-013 | `TestPruneStampShorterOverLongerIsExact` | 1 | `state file = "2026-10-02T00:00:00Z123456789Z", want exactly "2026-10-02T00:00:00Z"` |
| 14 | mC: stamp-write error ignored (`_ = writeStamp(sf, now)`) | AC-HRH-014 | `TestPruneStampWriteFailureSkipsPrune` | 1 | `pruneLocked returned nil, want the stamp-write error`; `log changed although the stamp could not be written` |
| 15 | mD: the prune call replaced by `var err error` | AC-HRH-012 | `TestPruneStamp_StampExistsBeforeTheWork` | 1 | `the pruner never reached its archive step: ...` after 10.01 s, no panic |

Count: the task text said 14 mutants; `plan.md` §F M1 step "Mutation runs" lists 3, M2 step 6 lists 9, and mB, mC, mD are 3 more, which is 15. All 15 were run. A 14-count was not reproduced from the plan text.

### 2.8 M6 (f) project tool and `moai spec lint`

| Command | Output | exit |
|---|---|---|
| `go build -o <scratch>/moai-tree ./cmd/moai` (tree `9289a92b6`, clean status) | (empty) | 0 |
| `<scratch>/moai-tree version` | `v3.1.3   none   built unknown` (plain `go build`, no ldflags) | 0 |
| `<scratch>/moai-tree spec lint SPEC-HARNESS-RETENTION-HARDEN-001` | `✓ No findings — all SPEC documents are valid` | 0 |

Judging build, stated next to the tree HEAD (`verification-claim-integrity.md` §2.2): the judging binary was built by `go build ./cmd/moai` from the working tree at HEAD `9289a92b6` with an empty `git status --short` before and after. Its embedded VCS stamp is not usable as the tree identity: `go version -m` printed `vcs.revision=c8f245c2c9a58083518f0b7cbebff0542848e033`, `vcs.time=2026-09-29T06:19:47Z`, `vcs.modified=true`, and `git merge-base --is-ancestor c8f245c2c HEAD` exits 1 (so that commit is not in this branch's history; it appears to be the primary checkout's HEAD, stamped because this is a linked worktree). The identity claim therefore rests on how the binary was built (from this tree's sources), not on the stamp. The installed `moai` is `v3.2.0-rc.26`, `archive/t1401-293-g45600e4ee`; `git merge-base --is-ancestor 45600e4ee HEAD` exits 0 (a strict ancestor), so the installed build lags the tree and was not used for this measurement.

## 3. Per-AC table

Adoption = how the criterion was adopted (RB: observed RED then flipped; RG: mutation proof). "Flip commit" = the commit that makes the named command pass on the card branch.

| AC | Test or check | Result (final tree) | Adoption | Flip commit |
|---|---|---|---|---|
| AC-HRH-001 | `TestPruneStateSymlinkReplacedTargetUntouched` | PASS | RB, RED observed at M0 (`red-baseline.md`) | `f52dd1b1c` (M2) |
| AC-HRH-002 | `TestPruneStateUnwritableFileReplaced` | PASS | RB, RED at M0 | `f52dd1b1c` (M2) |
| AC-HRH-003 | `TestPruneStamp_StateNotOpenableSkipsPruneAndRecordSucceeds` (a), `TestPruneStateUnreplaceableInReadOnlyDirSkips` (b) | PASS, PASS | RG, green at base; mutants 11 and 12 killed | stays green; case (b) strengthened in `f52dd1b1c` |
| AC-HRH-004 | `TestPruneStateForeignOwnedLeftUntouchedAndWarns` | PASS | RG, mutants 4 to 6 killed (no RED obtainable, G-1) | `f52dd1b1c` (M2) |
| AC-HRH-005 | `TestOwnerCheckDefault` | PASS (step (e) not skipped) | RG, mutants 7 to 9 killed (G-1) | `f52dd1b1c` (M2) |
| AC-HRH-006 | `TestHealDoesNotRemoveAFreshStateFile` | PASS | RG, mutant 10 killed (G-1) | `f52dd1b1c` (M2) |
| AC-HRH-007 | `TestPruneCarriesLateEvents` | PASS | RB, RED at M0 | `c1cc3fe67` (M1) |
| AC-HRH-008 | `TestPruneKeepsUnparsedLinesVerbatim`, `TestPruneNothingStaleLeavesLogUntouched`, `TestPruneTailPartialLineCarriedAndTerminated`, `TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval` | 4 PASS | cases b, c RB (RED at M0); case a RG pin; mutants 1 to 3 killed | `c1cc3fe67` (M1) |
| AC-HRH-009 | `git diff --quiet` and four-function `-L` form, section 2.6 | PASS (exit 0, empty) | RG, positive controls fired | holds at every card commit |
| AC-HRH-010 | eight sentinels, old sentence, card-id guard, section 2.3 | PASS (counts as shown) | RB, baseline counts 0 and 1 | `9289a92b6` (M5) |
| AC-HRH-011 | `GOOS=windows go build` and `go vet`, section 2.5 | PASS (build and vet only) | RG, green at base; controls E-017 and E-018 are plan-phase entries, not re-run here | holds at every card commit |
| AC-HRH-012 | `TestPruneStamp_StampExistsBeforeTheWork` under mD | fails in 10.01 s (13 s wall); unmutated PASS | RB, RED at plan phase (E-009) and re-observed in section 2.2 | `3badf7875` (M4) |
| AC-HRH-013 | `TestPruneStampShorterOverLongerIsExact` | PASS; fails under mB (mutant 13) | RB (survivor RED), test landed at M0 | `ac40cf3bf` (M0), re-run at M3 |
| AC-HRH-014 | `TestPruneStampWriteFailureSkipsPrune` | PASS; fails under mC (mutant 14) | RB (survivor RED), kill observed at M3 | `b3a469eab` (M3) |

## 4. Baseline-attribution

Every row in section 2 was measured in this run (session 783a9ff3, manager-develop), on the card worktree `.moai/worktrees/t1432`, branch `WT-harness-retention-debt`, at the tree SHA stated per section (`b3a469eab` for 2.1 and the BEFORE half of 2.2; the M4 edit uncommitted for the AFTER half of 2.2; `9289a92b6` for 2.3 onward; the M5 non-comment diff was measured on the uncommitted M5 edit against `3badf7875`). Toolchain: go1.26.8, darwin arm64, uid 501; `golangci-lint` v2.1.6. Figures from `run-evidence-m1-m3.md` are not reused as measurements here; where a section depends on M0 to M3 (the RED baseline, M2 mutant history) it cites those files as prior evidence, not as re-measurement, except the 15 mutants which were all re-run.

## 5. Gaps

- **Windows runtime not observed.** Windows evidence is `GOOS=windows go build` and `go vet` (which compiles test files) only. The Windows owner-check twin (every entry "not owned") and the in-process Windows lock are unexecuted.
- **Linux is unobserved.** Every run was on darwin. The read-only-handle truncate seam (AC-HRH-014), the FIFO tests, the stderr capture, the `O_EXCL` create and the owner stat were first exercised on darwin only; CI on `ubuntu-latest` exercises them first. Whether a Linux CI runner finds a foreign-owned symbolic link among `/var`, `/tmp`, `/etc`, `/bin` for AC-HRH-005 (e) is not observed.
- **The heal burst and the residual loss window are not measured.** The gap between the last re-inspection and the removal (D4.c), the burst of N hook processes meeting the same fault, and the residual window between the final tail reading and the rename (`spec.md` §F) are disclosed, not measured. Whether a concurrent `O_APPEND` write can be seen half-complete at the size reading was not observed (the boundary rule is tested deterministically).
- **The FIFO-at-state-path hang is a recorded pre-existing condition, not repaired** (`spec.md` §A FIFO row, §E, §F; ledger E-025). No test here uses a FIFO at the state path; `PruneStaleEntries` still blocks in the lock-free pre-check on one.
- **Time-of-check gap** between the inspection and the open of a regular file is closed for writes by the post-open identity check and is not pinned by a test (`spec.md` §F).
- **AC-HRH-011 controls** (E-017, E-018 untagged and tagged FIFO test files) were plan-phase observations and were not re-run in this run. Only the positive control of AC-HRH-009 (section 2.6) was re-run.
- **Reviewer-read parts**: the F5 and F6 sentences against REQ-HRH-010 and REQ-HRH-011 (the `async` obligation included), and the phrase naming the missing archive step in AC-HRH-012 (audit debt O9: judged by reading; the committed message reads `the pruner never reached its archive step`).
- **The M4 fix on a slow runner**: `archiveStepWait` is 10 s. A runner so slow that a healthy pruner needs more than 10 s to reach its archive step would see a false red; not observed, and the direction is a false red, not a missed defect. The 10 s wait is a test constant, not a production value.
- **Judging-build stamp**: the tree-built binary's embedded `vcs.revision` is not the tree's HEAD (section 2.8), so the second coordinate of §2.2 rests on the build procedure, not on the stamp.
- **Refused or non-displayed commands.** No command was refused by the worktree guard in this run. Two things were worked around, not refused: (1) the shell tool did not display the exit code of a no-match or counting `grep`; those commands are reported by their printed counts, and exit codes for single `git`/`go` commands were taken by `; echo "exit=$?"` in the same invocation. (2) The `Grep` tool is not available to this agent (`No such tool available: Grep`), so `grep` through the shell was used instead. The `time` wrapper printed the command line with the overlay path removed (the shell's `time` display); the result lines in the output file are intact.
- **Count discrepancy**: the task said 14 mutants; the plan lists 15 (section 2.7a).
- **Lint**: the installed `golangci-lint` is v2.1.6, the version CI pins; its build was not compared with CI's.
- **Coverage** was not re-measured in this run. The t1425 delta audit's 87.5 percent is a different tree and is not carried over.
- **Audit debt carried** (accepted, listed in `.moai/reports/t1432/decision-records.md` and `plan-audit-iter3.md`): N1 (DoD-1 wording), N2 (stale "untracked" note in G-7 and `plan.md` B5), N3 (a statement that CI never compiles Windows tests; `ci.yml` runs `GOOS=windows go vet ./...`), N4 (AC-HRH-005 (e) skip granularity), N5 (no end-to-end case for a user-owned link to a root-owned target), N6 (mutant copies in `red-now-drafts/` are base-bound; discharged here by regenerating every mutant from the current files, none from the stale copies), O4 (AC-HRH-002 identity assertion could be falsely red on a file system that reuses a freed inode number: inferred, APFS observed not to reuse, Linux unobserved), O8 (AC-HRH-007 compares `late-event` by substring, so a re-encoding mutant passes), O9 (above). The SPEC body was not edited.

## 6. Residual-risk

- A pruner that is killed or slow in the residual window loses an event appended in that window; the tail carry narrows the window to the gap between the final tail reading and the rename and does not close it.
- A foreign-owned or Windows state-path entry keeps retention off, with one warning per prune attempt and no rate limit; the observer still discards the returned error by design, and whether a hook's stderr reaches a user was not measured.
- A burst of hooks meeting a faulty owned entry at the same instant can still produce more than one pruner within the gap between the last re-inspection and the removal.
- The FIFO hang remains: a FIFO (or a link to one) at the state path blocks the hook until its 5 s timeout.
- Linux behaviour of the seams listed under Gaps could differ from darwin and would show first in CI.
- The sentinel greps prove presence of phrases, not that the sentences state the obligations.

## 7. Deviations between the SPEC text and what was built

1. **Strengthened test (M2).** `acceptance.md` AC-HRH-003 case (b) and `plan.md` M2 step 6 say the mutant "a heal that ignores the removal failure" is killed by the case-(b) test, whose stated assertions are a non-nil error and byte-identical state. With the removal error ignored the loop still ends in an error after three inspections and leaves everything untouched, so those assertions let the mutant survive. M2 added one assertion to the committed `TestPruneStateUnreplaceableInReadOnlyDirSkips`: the returned error must satisfy `errors.Is(err, fs.ErrPermission)`. The unmodified code passes it and mutant 12 fails it. The strengthened test was not re-run against base code (READ only: the base error wraps the open's permission error).
2. **M1 gate (M1).** `plan.md` M1 implies a green package at the M1 commit; the two M0 REDs (AC-HRH-001, -002) belong to M2, so the M1 gate was the package minus those two named tests (`run-evidence-m1-m3.md`).
3. **M4 mechanism.** The SPEC criterion binds the outcome (fail within 15 s naming the missing step), not the mechanism; the goroutine-and-release form was chosen over the non-blocking read open (section 2.2). The fixed test waits 10 s before failing, so the elapsed test time under mD is about 10 s.
4. **Mutant count** 14 in the task text versus 15 in the plan (section 2.7a).

## 8. For the sync phase (`progress.md`) — not edited here

- Run-phase section (§E.2): commits `ac40cf3bf` (T), `c1cc3fe67`, `f52dd1b1c`, `b3a469eab`, `3badf7875`, `9289a92b6`, and the verdict commit that follows; evidence paths `.moai/reports/t1432/red-baseline.md`, `run-evidence-m1-m3.md`, `verdict.md`.
- The M0 Exit ordering check evaluated on the card branch (section 2.7) with its controls.
- Linux CI is the first run of: the read-only-handle truncate seam, the FIFO tests, `O_EXCL` create, the owner stat on Linux and the AC-HRH-005 (e) candidate list.
- The four deviations of section 7; the carried debt N1 to N6, O4, O8, O9; the G-3 and G-7 items now resolved or still open (overlay JSON files in `red-now-drafts/` still name the plan-phase scratch copies; mutants for this run live in the session scratch directory and are not part of the change).
- The `moai spec lint` result with its judging-build coordinates (section 2.8).
- The single allowed test-only field (`ownerCheck`) was spent on the owner lookup; the operator-held options D1-C, D2-B, D2-C and D3-A remain unselected and untouched; `internal/lockfile` and `observer.go` are byte-identical to `1e2151a38`.
- Whether `.moai/reports/t1432/*` is pushed stays with the leader (B8).

---

# Amendment 0.4.1 run phase — the heal lock (M7 to M10)

Run-phase takeover from lane-12 (stopped at its usage limit), in an isolated agent worktree
(branch `worktree-agent-a3712f21dbd667f41`) that merged `WT-harness-retention-debt` @ `09a701ad2`
into local `develop` @ `02f939b14` (merge `f040ffcd3`; one CHANGELOG conflict resolved by keeping
both sides). The original tree `.moai/worktrees/t1432` was not modified. Platform of every
measurement: darwin arm64, uid 501, go1.26.8; host load average 70 to 88 during M10 (other
sessions' test runs). Nothing pushed, nothing merged into develop.

## A1. Commits

| Milestone | SHA | Content |
|---|---|---|
| merge | `f040ffcd3` | take over `WT-harness-retention-debt` |
| M7 — intermediate red commit T2 | `46485ec48` | `retention_heallock_test.go` (new), B11 guard in the swap test, `red-baseline-amend.md` |
| M8 | `84ea58d48` | `retention.go`, `retention_heal_unix.go`, `retention_heal_windows.go`, `.gitignore` |
| M9 | `aa42a4398` | N1, N4 (`retention_owner_test.go`), N2 (`retention_tail_test.go`), D1 bound (`retention_heallock_test.go`) |
| M10 | the commit carrying this section | verdict, progress, CHANGELOG wording, load-measurement draft |

## A2. RED (M7, T2 `46485ec48`)

Full record: `.moai/reports/t1432/red-baseline-amend.md`. AC-HRH-006 (b), (b2), AC-HRH-015 and the
four AC-HRH-016 cases FAIL against the unmodified `retention.go`; (c) both variants, (d),
AC-HRH-003 (b) and (c) and the guarded swap test PASS. Each matches its ledger cell (E-036 to
E-038, E-047a, E-039, E-048a).

## A3. GREEN (M8, tree `84ea58d48`)

`go test -count=1 -race -v -timeout 300s -run "TestPrune[HCS]" ./internal/harness/` (50 `=== RUN`),
`ok github.com/modu-ai/moai-adk/internal/harness 11.697s`:

```
--- PASS: TestPruneHealLockHeldPastTheBoundFailsClosed (2.02s)
--- PASS: TestPruneHealLockHostileEntryFailsClosed (0.06s)
    --- PASS: TestPruneHealLockHostileEntryFailsClosed/symlink (0.02s)
    --- PASS: TestPruneHealLockHostileEntryFailsClosed/directory (0.01s)
    --- PASS: TestPruneHealLockHostileEntryFailsClosed/fifo (0.02s)
    --- PASS: TestPruneHealLockHostileEntryFailsClosed/not-owned (0.01s)
--- PASS: TestPruneCommonPathCreatesNoHealLock (0.05s)
--- PASS: TestPruneStateRemovalFailureInReadOnlyDirSkips (0.01s)
--- PASS: TestPruneStateUnreplaceableInReadOnlyDirSkips (0.01s)
--- PASS: TestPruneCommonPathIgnoresAHeldHealLock (0.05s)
--- PASS: TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal (0.03s)
--- PASS: TestPruneHealWaitsForASharedHolder (0.31s)
--- PASS: TestPruneHealSerializesOnTheHealLock (0.33s)
--- PASS: TestPruneHealLockIsFreeWhilePrunerIsInItsArchiveStep (0.33s)
```

Sentinels after M8: `grep -c -F "heal lock gives no exclusion on Windows" internal/harness/retention.go` → `1`;
`grep -c -F "heal window remains" internal/harness/retention.go` → `0`;
`git check-ignore -v .moai/harness/usage-log.jsonl.prune-heal` → `.gitignore:353:.moai/harness/usage-log.jsonl.prune-heal`, exit 0.

## A4. Mutation table (regenerated from the files at `84ea58d48`; M9 rows from `aa42a4398`)

Each mutant is a scratch copy of the then-current file injected with `go test -overlay`
(session scratch directory, not committed). Run: `-run "TestPrune[HCS]"` (M9 rows: the three M9 tests).

| Mutant | Killed by (verbatim) |
|---|---|
| heal lock requested in shared mode (`LOCK_SH`) | only `TestPruneHealWaitsForASharedHolder`: `retention_heallock_test.go:426: the pruner returned (<nil>) while another descriptor held a shared heal lock: it did not request an exclusive lock` |
| lock never taken (flock call replaced by a nil error) | `TestPruneHealSerializesOnTheHealLock` (`:116 ... it did not wait`), `TestPruneHealWaitsForASharedHolder` (`:426`), `TestPruneHealLockHeldPastTheBoundFailsClosed` (`:201 want an error naming <tmp>/.../usage-log.jsonl.prune-heal, got <nil>`) |
| heal-lock error formatted with `%v` (no wrap) | `TestPruneStateUnreplaceableInReadOnlyDirSkips` (AC-HRH-003 b): `retention_statepath_test.go:177: error does not wrap a permission error: retention: prune heal lock <tmp>/.../usage-log.jsonl.prune-heal cannot be created: open ...` |
| heal ignores the removal failure | `TestPruneStateRemovalFailureInReadOnlyDirSkips` (AC-HRH-003 c): `:498: the removal failure's cause is not wrapped (a heal that ignores the failure ends differently): retention: prune state entry <tmp>/... changed on every inspection...` |
| heal lock taken on the common path when its file exists | `TestPruneCommonPathIgnoresAHeldHealLock`: `:394: the common-path prune took 2.009280042s while the heal lock was held: it waited for a lock it must not take`; also AC-HRH-015/016 (duplicate warning absent), (b2) and AC-HRH-003 (c) |
| heal lock never released (`defer release()` dropped) | `TestPruneStateRemovalFailureInReadOnlyDirSkips`: `:522: the heal lock is still held after the prune returned: resource temporarily unavailable`. `TestPruneHealLockIsFreeWhilePrunerIsInItsArchiveStep` (d) did NOT fail on this mutant in this run (the unreferenced descriptor can be closed by the garbage collector's finalizer, releasing the lock); not relied on |
| blind 2 s sleep before a single try (audit debt D1) | survived at `84ea58d48` (`ok ... 7.937s`); after M9 killed by `:157: the pruner returned 1.699337708s after the heal lock was released, want under 1s: it did not poll for the lock` and `:452: ... 1.701022042s after the shared heal lock was released ...` |
| N1: removal decided by modification time alone | `TestHealRemovalIsNotDecidedByModTimeAlone`: `retention_owner_test.go:332: removal of a changed entry with an equal modification time: removed=true err=<nil>, want false and nil` |
| N2: empty tail writes a newline | `TestPruneWithoutLateEventAddsNoBlankLine`: `retention_tail_test.go:240: replacement log = "{...\"subject\":\"fresh\"...}\n\n": want exact...` |
| N4: file arm removes directly, bypassing the conditional removal | `TestPruneHealFileArmKeepsAFreshStateFileSwappedInDuringTheHeal`: `retention_owner_test.go:392: the fresh state file was removed or replaced by the heal's file arm: err=<nil>` |

Not built in this run (reasoned only, as in the delta audit): open following links, owner check
skipped, hostile entry removed and recreated, unbounded wait, heal proceeding after the bound.
Limit stated by the SPEC and unchanged: a lock released BEFORE the removal is not killed by any
deterministic test; the race probe (A5) is its only guard.

## A5. Race probe (Definition of Done 10, final tree `aa42a4398`)

Overlays regenerated with this worktree's paths (session scratch). Each `go test -count=1 -v -timeout 300s -overlay <ov> -run <name> ./internal/harness/`:

```
zz_heal_race_probe_test.go:97: PROBE heal-window: trials=20000 heal_removed_the_swapped_in_fresh_entry=0 path_holds_F=20000 path_holds_other=0 heal_errors=0
--- PASS: TestZZProbeHealWindow (49.70s)
zz_heal_race_probe_nolock_test.go:75: PROBE heal-window-NOLOCK: trials=2346 heal_removed_the_swapped_in_fresh_entry=5 path_holds_F=2341 path_holds_other=0 heal_errors=0
zz_audit_race_test.go:67: PROBE remove-window: trials=2474 helper_removed_the_swapped_in_fresh_entry=5 helper_first=2467 swap_first=2
```

The probe ran the full 20000 trials (no early stop, not on the deadline): 0 removals. The no-lock
control (the draft with the swapper's two lock lines removed; scratch copy, not committed) and the
unchanged bare-helper audit probe both reach 5 removals, so the probe can fail and the lock is what
makes it pass. Pre-fix references 5/2135 and 5/1270 (lane-12 and the ledger).

## A6. Verification batch (M10, tree `aa42a4398` plus the uncommitted M10 text edits; no Go file differs)

| Command | Result |
|---|---|
| `go test -count=1 -race -timeout 30m ./internal/harness/...` | 15 packages `ok` (`internal/harness 11.270s`; `rosterguard 1226.148s` under host load), exit 0 |
| `go test -race -count=5 -v -run "TestPruneHeal\|TestPruneCommonPath\|TestPruneStateRemovalFailureInReadOnlyDirSkips\|TestPruneStateUnreplaceableInReadOnlyDirSkips\|TestHealRemovalIsNotDecidedByModTimeAlone\|TestPruneWithoutLateEventAddsNoBlankLine" ./internal/harness/` | 65 `--- PASS` (13 tests × 5, AC-HRH-015 included), 0 FAIL, 0 SKIP, `ok ... 14.111s` |
| `go vet ./internal/harness/ ./internal/lockfile/` | exit 0 |
| `gofmt -l internal/harness/` | empty |
| `GOOS=windows go build ./...` | exit 0 |
| `GOOS=windows go vet ./internal/harness/... ./internal/lockfile/` | exit 0 |
| `golangci-lint run ./internal/harness/` (v2.1.6) | `0 issues.`, exit 0 |
| `git diff --quiet 7639c04c1 HEAD -- internal/lockfile internal/harness/observer.go` | exit 0 |
| AC-HRH-009 `git log --format=%h -s -L ... 7639c04c1..HEAD` (and `1e2151a38..HEAD`) | empty |
| DoD 9: `git log --diff-filter=A --format=%h -- internal/harness/retention_heallock_test.go` | `46485ec48` |
| DoD 9: `git log --reverse --format=%h 7639c04c1..HEAD -- internal/harness/retention.go internal/harness/retention_heal_unix.go internal/harness/retention_heal_windows.go` | `84ea58d48` (non-empty) |
| DoD 9: `git merge-base --is-ancestor 46485ec48 84ea58d48` | exit 0 |
| DoD 9 control: `git merge-base --is-ancestor 84ea58d48 46485ec48` | exit 1 |

DoD 9 was evaluated on this takeover branch, which carries the card branch plus local develop; the
leader re-evaluates it on whatever branch is merged.

## A7. Load measurement (Definition of Done 15) — REPORTED TO THE LEADER: every observation exceeds 5 s

Instrument: `.moai/reports/t1432/amend-drafts/zz_load_measure_test.go` (committed, injected with
`-overlay`). Fixture: synthetic log 65,800,120 bytes, 1 in 8 events older than the cut, a user-owned
symlink at the state path, a test-owned descriptor holding the heal lock and releasing it 1.8 s
after the call starts. Load: 16 CPU-busy goroutines in the test process, context deadline, stopped
by `t.Cleanup`; whole run under `timeout`. Every observation returned `err=<nil>`, shrank the log to
57,574,933 bytes (valid).

Run 1 (`timeout 180 go test ... -timeout 150s`, 120 s load deadline):

```
MEASURE no-load #1: wall=14.38954125s  #2: wall=9.024902s  #3: wall=7.2170115s
MEASURE load    #1: wall=17.848337292s #2: wall=17.711502833s #3: wall=23.42840225s
```

Run 2 (after the first exceeded 5 s: prune-only decomposition added; `timeout 330 ... -timeout 300s`, 200 s load deadline):

```
MEASURE no-load prune-only #1: wall=13.962545334s #2: wall=10.149629083s #3: wall=11.715824875s
MEASURE no-load            #1: wall=14.557523958s #2: wall=24.747596583s #3: wall=33.613738625s
MEASURE load               #1: wall=34.269506375s #2: wall=27.446193s    #3: wall=18.909516417s
MEASURE load prune-only    #1: wall=13.1276555s   #2: wall=7.739591583s  #3: wall=9.635313292s
```

Reading: on this host the prune of a 65.8 MB log ALONE took 7.7 to 14.0 s, already above the 5 s
hook timeout before any heal-lock wait; the heal-lock wait adds the holder's ~1.8 s on top. The
"no-load" control is not unloaded: the host load average was 70 to 88 from other sessions
throughout (`uptime` 12:43 and 12:48), so neither arm isolates this process, and the run-to-run
spread (7 s to 34 s) is dominated by that external load. The t1425 lane's 1.79 s prune for a log of
this size was not reproduced here (not re-measured on an idle machine). Consequence for the 2 s
bound: the bound itself is not what crosses 5 s on this host; a large log's prune does, with or
without the heal lock. Not accepted silently: this is the leader's decision.

Stated unmeasured: real hook processes, an idle machine, a larger log, a holder stalling past
1.8 s, the waiter-behind-another-prune case (same magnitude, not separately timed), Linux, Windows,
a cold page cache, production frequency.

## A8. Gaps

- Windows runtime not observed (`GOOS=windows` build and vet only). Linux unobserved; CI on ubuntu-latest is the first run of the heal-lock tests (flock semantics, O_NOFOLLOW/O_NONBLOCK, FIFO create).
- Skipped tests: none skipped in the heal-lock group on darwin uid 501 (0 SKIP in the count=5 run); `TestPruneStateRemovalFailureInReadOnlyDirSkips` and `TestPruneHealFileArmKeepsAFreshStateFileSwappedInDuringTheHeal` skip as uid 0 (not exercised).
- The kanban environment variables were not scrubbed for the test runs: the worktree-isolation guard refused every compound form (`unset ... && go test`, `env -u ... go test`) and admitted only the plain command; `grep -rl MOAI_KANBAN internal/harness/` printed nothing.
- Audit debt D2 (warning line on a create failure unpinned) not taken; D3, D4, D6 (SPEC text) not touched because the SPEC body is frozen for the plan-audit hash.
- `moai spec lint` not run in this takeover.
- Mutant copies, overlays and the no-lock probe variant live in the session scratch directory and are not committed (the committed drafts under `amend-drafts/` are the originals).
- The never-released mutant was not killed by AC-HRH-006 (d) in this run (finalizer-dependent); it is killed by AC-HRH-003 (c).

## A9. Residual-risk

- The load measurement above: a large log's prune exceeds the 5 s hook timeout on a loaded host independent of the heal lock.
- A process that does not honour the heal lock (an older installed binary during the upgrade window) still races the removal, as the no-lock control shows.
- On Windows no heal runs (the owner check refuses every entry); the heal lock gives no exclusion there.
