auditor-model: claude-sonnet-5-5[1m]

# t1425 sync audit (DELTA, iteration 2) - stamp-before-work repair

verdict: PASS-WITH-DEBT
audited_sha: 990a29b671d2997fb169611ca1edbe21dd5730ef

Card t1425 (Class B urgent repair, no SPEC). Branch `WT-harness-prune-single-writer`, base `c640b0193`. This is the delta audit of `0fa7f49e8..990a29b67` against the first audit `.moai/reports/t1425/sync-audit.md` (PASS-WITH-DEBT 87). Iteration 2. Overall score **90 / 100** (weighted 90.0, harmonic 89.9), judged on the whole delivered state at HEAD.

```
## Evaluation Report
SPEC: none (card t1425)
Overall Verdict: PASS-WITH-DEBT
```

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 92/100 | PASS | F1 closed and measured (dups per kill 15000 -> 0 after the first); one repeat per interval under permanent kills; all card requirements still met |
| Security (25%) | 86/100 | PASS | no new surface; carry-over F4 (state symlink truncated) and F7 (unwritable state file) unchanged, severity unchanged |
| Craft (20%) | 88/100 | PASS | package coverage 87.5 percent; 5 requested mutants + M9' killed; 2 new survivors (mB, mC) and one test hang mode (N1) |
| Consistency (15%) | 94/100 | PASS | gofmt/vet/lint clean, Windows build+vet clean, scope clean, naming and comment style match the file |

Must-pass firewall (Functionality, Security): both above threshold, no dimension at 0. No blocking finding.

## 1. Evidence-bearing report

### Claim
1. F1 is closed: a pruner killed after the archive append and before the rename no longer causes later hooks to re-archive the same events inside the interval (measured with a real SIGKILLed child process, old ordering vs new).
2. The five mutants the run agent names (M4, M8', M9, M10, M11) are killed by the new tests, and so is a second M9 variant (stale clock used for the stamp comparison only).
3. F2 and F3 of the first audit are fixed.
4. The reorder introduces no new blocking defect; the failure paths I could construct behave as designed.
5. The new tests are CI-safe except for one failure-mode hang (N1) and two untested branches (N2).

### Evidence

All commands ran in this session against worktree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1425`, HEAD `990a29b671d2997fb169611ca1edbe21dd5730ef`, `go1.26.8 darwin/arm64`. Scratch outputs are under the session scratchpad `delta-audit-t1425/` (machine-local; the deciding lines are copied below). Kanban env scrubbed in the same invocation as every `go test`. Slot `go-test-internal-harness` acquired (exit 0) before the first test and released (exit 0, `free`) before this file was written; it was re-acquired once for the hourly-kill probe and released again.

**E1 - the diff** (`git diff 0fa7f49e8 HEAD -- internal/harness/retention.go`, read in full): 17 insertions / 16 deletions. The single behavioural change in `pruneExclusive`: after the re-check under the lock, `if err := writeStamp(sf, now); err != nil { return fmt.Errorf("retention: prune state write failed: %w", err) }` then `return r.prune(retentionDays, now)`. The old `pruneErr/stampErr` block is gone. Comments and `@MX` text updated. Nothing else in `retention.go` changed (`prune`, `partitionEvents`, `archiveEvents`, `appendToGzip`, `overwriteWithEvents`, `writeStamp`, `readStamp` are byte-identical).

**E2 - F1 reproduction, real killed process, old ordering vs new.** Scratch test `zz_audit_kill_test.go` added through `-overlay` only (not in the tree). The parent builds a 41,127,780-byte log (15,000 stale + 105,000 fresh events) and re-executes the test binary as a child; the child prunes with an injected clock and SIGKILLs itself as soon as its own archive append has grown and been stable for 30 ms, i.e. after `archiveEvents` and before the rename. "Old" is `retention.go` rebuilt from the HEAD copy with the pre-diff ordering (prune, then stamp), overlaid; "new" is HEAD unmodified. Children 1..5 start at `base+k minutes` (inside one interval).
```
OLD (stamp after the work), command: go test -timeout 8m -count=1 -v -run '^TestAuditKillParent$' -overlay=kill_old.json ./internal/harness/   exit 0
AUDIT-K log bytes=41127780 stale=15000 fresh=105000
child 1: killed=true archive_events=15000 unique=15000 dups=0     stamp="" log_still_has_stale=true
child 2: killed=true archive_events=30000 unique=15000 dups=15000 stamp="" log_still_has_stale=true
child 3: killed=true archive_events=45000 unique=15000 dups=30000 stamp="" log_still_has_stale=true
child 4: killed=true archive_events=60000 unique=15000 dups=45000 stamp="" log_still_has_stale=true
child 5: killed=true archive_events=75000 unique=15000 dups=60000 stamp="" log_still_has_stale=true

NEW (HEAD), same command with -overlay=kill_new.json   exit 0
child 1: killed=true  archive_events=15000 unique=15000 dups=0 stamp="2026-10-02T00:01:00Z" log_still_has_stale=true
child 2: killed=false archive_events=15000 unique=15000 dups=0 stamp="2026-10-02T00:01:00Z" log_still_has_stale=true
child 3: killed=false archive_events=15000 unique=15000 dups=0 ...
child 4: killed=false archive_events=15000 unique=15000 dups=0 ...
child 5: killed=false archive_events=15000 unique=15000 dups=0 ...
```
Old: 15,000 duplicate events per further kill, no stamp, unbounded. New: the first kill leaves the stamp; children 2-5 skip at once (exit 0, not killed, archive unchanged).

**E3 - permanent slowness (every prune killed), new code, children at +1, +2, +62, +63, +124 minutes** (`go test ... -overlay=kill_new.json`, exit 0):
```
child 1:   killed=true  archive_events=15000 dups=0     stamp="2026-10-02T00:01:00Z"
child 2:   killed=false archive_events=15000 dups=0
child 62:  killed=true  archive_events=30000 dups=15000 stamp="2026-10-02T01:02:00Z"
child 63:  killed=false archive_events=30000 dups=15000
child 124: killed=true  archive_events=45000 dups=30000 stamp="2026-10-02T02:04:00Z"
```
Under a machine so slow that every prune is killed, the repeat is one duplicate batch per interval (hourly) and the log never shrinks; the repeat no longer scales with the number of hooks. This is the design's stated bound ("at most once per interval").

**E4 - mutants, independent re-run.** Method: `go test -timeout 8m -count=1 -overlay=<mutant>.json ./internal/harness/` replacing only `retention.go` with a scratch copy of HEAD plus one change; the worktree was never edited. Verbatim red lines (filtered with grep):

| Mutant | Change | Killing test (first red) | Verbatim red | exit |
|---|---|---|---|---|
| M4 stamp only on success | `if pruneErr := r.prune(...); pruneErr != nil { return pruneErr }` then `writeStamp` | `TestPruneStamp_FailedPruneStillStampsAndIsNotRetried` (also `TestPruneStamp_StampExistsBeforeTheWork`) | `retention_stamp_contract_test.go:59: a failed prune left no fresh stamp on disk: ""`; `retention_killed_test.go:70: no stamp on disk while the pruner was still working: a killed pruner would be repeated`; `FAIL ... 3.229s` | 1 |
| M8' old ordering (prune, then stamp, always) | pre-diff block restored | `TestPruneStamp_StampExistsBeforeTheWork` | `retention_killed_test.go:70: no stamp on disk while the pruner was still working: a killed pruner would be repeated` (fails after the 2 s bound; `FAIL ... 3.173s`) | 1 |
| M9 re-check under the lock uses the pre-lock clock (struct field `preNow`, no second clock read) | `now := r.preNow` | `TestPruneStamp_RecheckUsesFreshClock` | `retention_stamp_contract_test.go:180: expected exactly two clock readings (pre-lock, post-lock), got 1` AND `:183: pruned although the post-lock reading shows a fresh stamp` | 1 |
| **M9b** (new variant): clock read twice, FIRST reading used for the stamp comparison, second for the prune | `stampIsFresh(readStamp(sf), cmpNow)` with `cmpNow = r.preNow` | `TestPruneStamp_RecheckUsesFreshClock` (the prune assertion, not the call count) | `retention_stamp_contract_test.go:183: pruned although the post-lock reading shows a fresh stamp`; `FAIL ... 2.311s` | 1 |
| M10 lock-free fast path disabled | `if false && stampIsFresh(readStampFile(...), now)` | `TestPruneStamp_FreshStampNeedsNoLock` (also `TestPruneStamp_StampExistsBeforeTheWork`) | `:142: PruneStaleEntries behind a held lock with a fresh stamp did not return within 2s`; `retention_killed_test.go:75: second hook during a stuck prune did not return within 2s` | 1 |
| M11 prune unlocked when the state file cannot be opened | `return r.prune(retentionDays, r.nowFn())` on open error | `TestPruneStamp_StateNotOpenableSkipsPruneAndRecordSucceeds` | `retention_stamp_contract_test.go:95: expected an error when the state file cannot be opened` | 1 |
| mA stamp written with the wrong time (`now-2h`) | `writeStamp(sf, now.Add(-2*time.Hour))` | `TestPruneConcurrentProcessesSingleRewrite`, `TestPruneStamp_FailedPruneStillStampsAndIsNotRetried`, `TestPruneStamp_PersistedAcrossInstances`, `TestPruneStamp_StampExistsBeforeTheWork` | `a fresh instance inside the skip interval pruned: stamp not respected`; `second wave changed the archive stream count: 2 -> 3` | 1 |
| **mB SURVIVES** stamp not truncated before the write | `Truncate(0)` removed from `writeStamp` | none | `ok  	github.com/modu-ai/moai-adk/internal/harness	1.812s` | 0 |
| **mC SURVIVES** stamp write error ignored (`_ = writeStamp(...)`) | the "skip prune when the stamp cannot be recorded" branch removed | none | `ok  	github.com/modu-ai/moai-adk/internal/harness	3.087s` | 0 |
| mD pruner stamps then does nothing (hang probe, `-timeout 40s`) | `return nil` instead of `r.prune` | `TestPruneStamp_StampExistsBeforeTheWork` HANGS (see N1) | `panic: test timed out after 40s ... TestPruneStamp_StampExistsBeforeTheWork (40s)`, goroutine in `syscall.Open` | 1 |

M9 note: the run agent edited `t.Fatalf` to `t.Errorf` after running M9 and did not re-run it. I re-ran it after that edit: the mutant is killed, and both assertions fire (the call count AND the prune assertion), so the Errorf change did not weaken the kill. M9b shows the prune assertion alone (not the count) kills the variant that reads the clock twice.

The M9 mutant implemented as a signature change did not compile against the existing direct caller in `retention_stamp_test.go:177`; that is a build failure, not a kill, so I rebuilt it as a per-instance struct field (no package-level state, so no cross-test leak).

**E5 - new tests, repeat and race.**
```
go test -timeout 8m -count=20 -v -run 'TestPruneStamp_' ./internal/harness/   exit 0   ok 1.107s
  240 "--- PASS: TestPruneStamp_" lines (12 tests x 20); the five new tests 20 PASS each; 0 FAIL, 0 DATA RACE
go test -timeout 8m -count=20 -v -run 'TestPruneConcurrentProcessesSingleRewrite' ./internal/harness/   exit 0   ok 11.058s
  20 PASS; 20x RESULT wave1 streams=2 dups=0 archive_files=2 corruption=0 archive_bytes=88651 gzip_headers=2
            20x RESULT wave2 streams=2 dups=0 log_unchanged=true
go test -timeout 8m -race -count=1 -v -run 'TestPrune' ./internal/harness/   exit 0   ok 6.338s
  RESULT wave1 streams=2 dups=0 archive_files=2 corruption=0 archive_bytes=88651 gzip_headers=2; wave2 log_unchanged=true; no DATA RACE
go test -timeout 8m -count=1 ./internal/harness/... ./internal/lockfile/...   exit 0
  ok internal/harness 1.241s ... ok internal/harness/rosterguard 26.558s ... ok internal/lockfile 0.943s (all 16 packages ok)
go test -count=1 -coverprofile=... ./internal/harness/   ok 1.391s coverage: 87.5% of statements
  PruneStaleEntries 100.0%  pruneExclusive 87.5%  readStamp 100.0%  readStampFile 100.0%  stampIsFresh 100.0%  writeStamp 75.0%  prune 84.6%
```
These numbers match the run-evidence iteration-2 section (guard 20/20 with the identical RESULT lines; race ok 8.210 s there, 6.338 s here; coverage differs only by the new tests: 87.3 -> 87.5 percent).

**E6 - static and build checks**, each a separate plain command:
```
go vet ./internal/harness/...                                  exit 0, no output
golangci-lint --version -> golangci-lint has version v2.1.6 built with go1.26.8 ...
grep -n golangci-lint@ .github/workflows/ci.yml -> 464: run: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6
golangci-lint run ./internal/harness/...                       exit 0, "0 issues."
gofmt -l internal/harness                                      exit 0, no output
GOOS=windows GOARCH=amd64 go build ./internal/harness/...      exit 0
GOOS=windows GOARCH=amd64 go vet ./internal/harness/           exit 0
```
Installed lint build equals the CI pin. The `!windows` constraint on `retention_killed_test.go` is correct: it is the only file using `syscall.Mkfifo`, and the Windows vet (which compiles the other test files) passes.

**E7 - scope.**
```
git diff --stat c640b0193 HEAD   -> 11 files, 1631 insertions(+), 5 deletions(-):
  .gitignore (+1); .moai/reports/t1425/{decision-records,lane-measurements,red-baseline,run-evidence,sync-audit}.md;
  internal/harness/{retention.go, retention_concurrent_test.go, retention_killed_test.go, retention_stamp_contract_test.go, retention_stamp_test.go}
git diff --stat c640b0193 HEAD -- internal/lockfile internal/harness/observer.go internal/cli   -> (no output), exit 0
git status --short -> (empty) before and after all runs; HEAD 990a29b671d2997fb169611ca1edbe21dd5730ef throughout
```
The scope is still `internal/harness/` (retention.go + three new test files + the earlier retention_concurrent/stamp tests), the one `.gitignore` line, and `.moai/reports/t1425/`.

**E8 - CI platform matrix** (`.github/workflows/ci.yml`, `release-pr-multi-os.yml`): PR `test` job is `os: [ubuntu-latest]` with `go test -json -coverprofile ... ./...` (no `-race`); the PR `race` job runs ubuntu `go test -json -race -count=1 -timeout 20m ./...`; the release-PR job `Release Verify` runs `[ubuntu-latest, macos-latest, windows-latest]` with `go test -json -race -timeout 25m ./...`. So the new tests run on Linux on every PR (with and without `-race`) and on macOS+Windows at release time; the FIFO test runs on Linux and macOS, the other four on all three. I ran darwin only.

**E9 - hygiene read of `retention_killed_test.go`.** Cleanup: `defer drain()` guarded by `sync.Once`; every wait is bounded (2 s stamp poll, `waitOrFail` 2 s, 5 s finish) except the FIFO read-open inside `drain` (N1). No package-level mutable state in any of the three new test files or in `retention.go` (only constants). `t.Parallel` is safe: each test uses its own `t.TempDir()`. `syscall.Mkfifo` is POSIX and present on Linux and darwin; the test skips if it fails. `-count=20` and `-race` verified above. `TestPruneStamp_RecheckUsesFreshClock` couples to exactly two `nowFn` readings in `PruneStaleEntries` + `pruneExclusive` (intentional and documented in the test; a future third read fails it).

### Baseline-attribution
- Tree measured: `.moai/worktrees/t1425`, HEAD `990a29b671d2997fb169611ca1edbe21dd5730ef`; worktree clean at the start, at the end of every run batch, and before this file was written (the only new file is this verdict, under the ignored `.moai/reports/`).
- Judging build: every test figure comes from a `go test` binary compiled from this tree, or from this tree with ONE file replaced via `-overlay` (marked); `go1.26.8 darwin/arm64`; `golangci-lint v2.1.6` (equals the CI pin at `ci.yml:464`). The installed `moai` binary was used only for `moai slot` leases and corroborates nothing here.
- The kill experiments use a synthetic 41 MB log and a self-SIGKILL from a watcher goroutine; they measure the stamp/duplicate behaviour, not the production log or the Claude Code hook timeout. Timings are one-shot on a loaded machine.
- The old ordering ("OLD", M8') is rebuilt by hand from the HEAD copy to match the pre-diff text exactly (the diff in E1 is the specification); see Gaps for why it was not extracted from git.

### Gaps
- Refused commands: the worktree guard refused `git show 0fa7f49e8:internal/harness/retention.go > <scratch>/old.go` (also with `git -C <own path>`, and with `| cat >`) as "too complex to verify". I used `git show ... | sed -n` (read-only) to read the old text and rebuilt the old ordering by editing the HEAD copy; the OLD column of E2 and M8' therefore rest on a faithful hand reconstruction, not on the git blob. No verdict claim rests on a refused command's inferred output.
- Not observed: a real Claude Code hook-timeout kill (I used a watcher-goroutine SIGKILL); Linux (CI's platform) at runtime, only darwin; Windows at runtime (`GOOS=windows` build and vet only), so the Windows behaviour of the in-process-only lock and of a stamp written by a concurrent process is reasoned from the code, not measured; the 65 MB production log; `go test ./...` and the whole `internal/cli` suite (the first audit ran the `Harness`-named ones, ok 30.4 s; this change does not touch them: `git diff --stat ... internal/cli` empty).
- Not re-run in this iteration: M1 (no lock), M2 (no re-check under the lock), M3, M5, M7a/b, M12 from the first audit. The tests that killed them are unchanged and the multi-process guard (which kills M1/M3) passed 20/20 at HEAD, but I did not re-mutate them against the reordered code.
- Not examined: the truncate-then-write window of `writeStamp` as a real kill (a kill between `Truncate(0)` and `WriteAt` leaves an empty stamp that parses as "no stamp"); analysis only, see F-D1.
- F4 symlink, F5 Windows in-process lock, F6 no lock timeout, F7 unwritable state file, orphan `usage-log-*.tmp` sweep: deliberately untouched and not re-graded.

### Residual-risk
- Under permanent slowness every prune is killed: one duplicate archive batch per interval and the log never shrinks (E3). Bounded, not eliminated.
- A pruner killed between the stamp and the archive defers retention by one hour. Against a 30-day window this is harmless; no events are lost, only left in the log one hour longer.
- A kill mid-rewrite still leaves an orphan `usage-log-*.tmp` the size of the kept log; nothing sweeps it (named in the `@MX:NOTE`; the first audit's optional sweep recommendation was not adopted).
- Events appended by non-locking hooks between the pruner's read and the rename are lost (pre-existing, disclosed by the `@MX:WARN`).
- Windows: the in-process lock does not stop a burst of hooks that all read "no stamp" before the first stamp lands; stamp-before-work shrinks that window from the whole prune to the stamp check-to-write interval (strictly better than before).
- A partial multi-month archive failure (month 1 appended, month 2 fails) leaves duplicates on the next hourly retry. Pre-existing (the old code also stamped after a failure) and not made worse.

## 2. Disposition of the first audit's findings

| Finding | Disposition | Evidence |
|---|---|---|
| F1 killed pruner re-archives every later hook | **FIXED** (stamp part). Orphan-tmp sweep part of F1's "consider" was not adopted; disclosed in the `@MX:NOTE` and run evidence, optional | E2: OLD dups 0/15000/30000/45000/60000 with `stamp=""`; NEW dups 0 for all five, stamp present after the first kill, children 2-5 not killed; E3: one batch per interval under permanent kills |
| F2 no test pins stamp-after-failed-prune | **FIXED** | M4 killed by `TestPruneStamp_FailedPruneStillStampsAndIsNotRetried` (`a failed prune left no fresh stamp on disk: ""`). The test uses an un-creatable archive directory instead of a >64 KiB log line; the claim (stamp survives a failed prune, a second instance inside the interval does not retry, one past the interval does) is the same |
| F3 M11 / M9 / M10 survivors | **FIXED** | M11 killed by `TestPruneStamp_StateNotOpenableSkipsPruneAndRecordSucceeds` (state path is a directory, log directory stays writable, so the old "read-only directory fails anyway" masking is gone); M9 and the extra M9b killed by `TestPruneStamp_RecheckUsesFreshClock`; M10 killed by `TestPruneStamp_FreshStampNeedsNoLock` |

## 3. New findings

Severity scale: blocker / major / minor / info. Class: blocking (correctness or a stated requirement) / optional.

- **N1 [minor] [optional]** `internal/harness/retention_killed_test.go:39-50` (`drain`) - the cleanup opens the FIFO with a blocking `O_RDONLY`, which waits for a writer. If the pruner never reaches its archive step (it returned an error early, or a regression makes it skip the work after stamping), `drain()` blocks forever and the test only ends when `go test -timeout` fires: observed with mutant mD, `panic: test timed out after 40s ... TestPruneStamp_StampExistsBeforeTheWork (40s)`, goroutine parked in `syscall.Open`. The default package timeout is 10 minutes, and CI uses `-timeout 20m`/`25m` for `./...`. The test therefore turns an assertion failure into a long hang rather than a fast FAIL. Required fix (not applied): open the FIFO read side with `O_RDONLY|O_NONBLOCK` (succeeds without a writer and still completes a blocked writer's open), or bound `drain` with a timer that unblocks by opening the FIFO for write. A missing pruner is still detected; this only changes how a failure ends.
- **N2 [minor] [optional]** untested branches in `writeStamp` and `pruneExclusive` (`writeStamp 75.0%`, `pruneExclusive 87.5%`): mutant mB (drop `Truncate(0)`) and mC (ignore the stamp write error, keep pruning) both survive. mB matters in principle because `time.RFC3339Nano` trims trailing zeros, so a shorter stamp written over a longer one without truncation leaves trailing bytes and parses as "no stamp" (a rewrite storm); HEAD is correct because the truncate is present, so this is a test gap, not a defect. Required fix if wanted: one test that writes a long stamp (nine nanosecond digits), then a shorter one, and asserts `stampIsFresh` on the file content; the mC branch needs a write-failing file (for example `/dev/full` on Linux) and may not be worth a test.
- **N3 [info]** `retention_stamp_contract_test.go` `TestPruneStamp_RecheckUsesFreshClock` asserts `calls == 2` with `t.Errorf` and then the prune assertion with `t.Fatalf`; this is the order that lets both messages show (observed under M9). The coupling to exactly two clock reads is intentional.
- **N4 [info]** Reorder effect on the unchanged slow path: after a failed prune `r.lastPruneAt` stays unset in-process (unchanged from before), so the next call in the same process takes the on-disk fast path (one small file read) instead of the in-memory one; negligible. If `writeStamp` fails after a successful open (disk full), every hook now skips the prune after taking the lock (previously every hook pruned): retention is off while the disk is full, but the storm cannot start; the prune would fail on a full disk anyway.
- **F-D1 [info, analysis only]** The truncate-then-write window in `writeStamp`: a kill between `Truncate(0)` and `WriteAt` leaves an empty stamp, which parses as "no stamp"; the next hook then wins the lock and prunes once. This cannot produce a repeating storm (the next pruner writes a stamp before its work), so it costs at most one extra prune in a microsecond window. Not measured (Gaps).

Carry-over F4-F7 and the orphan-tmp sweep: the reorder does not change their severity. F5's Windows window is smaller than before (see Residual-risk); F7 is unchanged (the state file is still opened `O_RDWR|O_CREATE` before any stamp work, and still fails permanently when unwritable).

## 4. Verdict

**PASS-WITH-DEBT, 90/100.** F1 is closed with direct evidence from a real killed process: the repeat per further kill falls from 15,000 events (growing without bound) to zero, and to one batch per interval under permanent kills. The five named mutants plus a second M9 variant are killed independently by the new tests, including the M9 case the run agent had not re-run. Static, Windows, race and repeat checks are clean and the scope is unchanged. Remaining debt is all optional: one failure-mode hang in the new FIFO test (N1), two untested branches (N2), and the previously disclosed carry-over (F4-F7, orphan `usage-log-*.tmp`). No blocking finding.

Worktree state at the end: `git status --short` empty, HEAD `990a29b671d2997fb169611ca1edbe21dd5730ef`; slot `go-test-internal-harness` released and reported `free`.
