auditor-model: claude-sonnet-5-5[1m]

# t1425 sync audit - single-writer harness retention prune

verdict: PASS-WITH-DEBT
audited_sha: 0fa7f49e807887f295026468aab83d80584ba1b1

Card t1425 (Class B urgent repair, no SPEC; card text is the scope). Branch `WT-harness-prune-single-writer`, base `c640b0193`. Iteration 1 of 1 (first audit, no prior verdict). Overall score **87 / 100** (weighted 87.3, harmonic 87.2).

```
## Evaluation Report
SPEC: none (card t1425)
Overall Verdict: PASS-WITH-DEBT
```

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 86/100 | PASS | all six card requirements met (table below); gaps: killed-pruner non-convergence (F1), Windows burst (F5) |
| Security (25%) | 88/100 | PASS | no secret/injection surface; state-file symlink is followed and truncated (F4, confirmed) |
| Craft (20%) | 84/100 | PASS | package coverage 87.3 percent; 6 of 13 mutants survive (error paths, clock freshness, hot path unpinned) |
| Consistency (15%) | 94/100 | PASS | gofmt/vet/lint clean, scope clean, naming and flock idiom match `internal/lockfile` callers |

Must-pass firewall (Functionality, Security): both above threshold. No dimension at 0. Verdict is PASS-WITH-DEBT because there is no blocking finding, but F1 is a major finding that I recommend the leader weigh before landing.

## 1. Evidence-bearing report

### Claim
1. The delivered change makes N concurrent hook processes perform one prune and one log rewrite per interval, with the last-prune time on disk, and the hot path costs one tiny file read.
2. The RED evidence precedes the fix in the commit graph, and the RED is real (reproduced independently).
3. The diff stays inside its declared scope.
4. The tests would catch most wrong implementations, but not all (6 surviving mutants).
5. One scenario the card itself raised - a pruner killed by the 5 s hook timeout - is not made safe by the delivered ordering (F1).

### Evidence

All commands run in this session against worktree `.moai/worktrees/t1425`, HEAD `0fa7f49e807887f295026468aab83d80584ba1b1`, toolchain `go version go1.26.8 darwin/arm64`. Scratch outputs live under the session scratchpad `sync-audit-t1425/` (machine-local, not citable beyond this file; the deciding lines are copied below).

**E1 - tree and diff scope**
```
$ git diff --stat c640b0193 HEAD
 .gitignore                                    |   1 +
 .moai/reports/t1425/decision-records.md       |  21 ++
 .moai/reports/t1425/lane-measurements.md      |  31 ++
 .moai/reports/t1425/red-baseline.md           | 151 +++++++++
 .moai/reports/t1425/run-evidence.md           |  77 +++++
 internal/harness/retention.go                 | 119 +++++++-
 internal/harness/retention_concurrent_test.go | 422 ++++++++++++++++++++++++++
 internal/harness/retention_stamp_test.go      | 234 ++++++++++++++
 8 files changed, 1051 insertions(+), 5 deletions(-)
$ git diff --stat c640b0193 HEAD -- internal/lockfile internal/harness/observer.go internal/cli
(no output)
```
The diff hunks of `retention.go` end at the `// Read log file` comment inside the new `prune`; `partitionEvents`, `archiveEvents`, `appendToGzip`, `overwriteWithEvents` appear in no hunk, so they are byte-unchanged.

**E2 - ignore coverage**
```
$ git check-ignore -v .moai/harness/usage-log.jsonl.prune-state
.gitignore:352:.moai/harness/usage-log.jsonl.prune-state	.moai/harness/usage-log.jsonl.prune-state
$ grep -n 'usage-log\|harness/' internal/template/templates/.gitignore
219:!.agents/skills/moai-harness/
```
The template ignore file carries no `usage-log` line, so there is nothing to mirror (the `usage-log` hits under `internal/template/templates` are in `harness.yaml`, rules and skill docs only).

**E3 - commit-graph ordering (item 3)**
```
$ git log --format='%h %ad | %s' --date=iso c640b0193..HEAD
0fa7f49e8 2026-10-02 16:16:14 +0900 | docs(t1425): record the lane's own prune-duration and guard measurements
9fbf6d0a8 2026-10-02 16:13:11 +0900 | docs(t1425): record the run evidence ...
fe211e9c9 2026-10-02 16:12:32 +0900 | fix(harness): single-writer retention prune gated by an on-disk stamp (t1425)
785d70fbf 2026-10-02 16:05:45 +0900 | docs(t1425): record the observed RED for concurrent retention pruning
be6c80f16 2026-10-02 16:01:14 +0900 | docs(t1425): record the design decisions for the harness retention storm repair
$ git show --stat 785d70fbf   ->  only .moai/reports/t1425/red-baseline.md (151 insertions)
$ git show --stat fe211e9c9   ->  .gitignore, retention.go, retention_concurrent_test.go, retention_stamp_test.go
```
`785d70fbf` is the parent of `fe211e9c9`; it holds no test file and no production file (the tests entered with the fix, as `785d70fbf`'s message says). The baseline artifact is its own commit ahead of the change (`verification-claim-integrity.md` 2.3 satisfied for the artifact; the RED run itself was against untracked test files, which the commit message discloses).

**E4 - RED re-derived on UNMODIFIED `c640b0193:internal/harness/retention.go`** (overlay, worktree untouched; `retention_stamp_test.go` masked in the overlay because it needs the new `pruneExclusive`). Command: `go test -overlay=OLD.overlay.json -count=5 -v -run TestPruneConcurrentProcessesSingleRewrite ./internal/harness/` -> exit 1, `FAIL` 5 of 5. Verbatim RESULT lines (filtered):
```
RESULT wave1 streams=0 dups=0 archive_files=0 corruption=2 archive_bytes=709208 gzip_headers=16
archive corruption: [2026-07.jsonl.gz: stream 1: flate: corrupt input before offset 1 2026-08.jsonl.gz: stream 1: flate: corrupt input before offset 1]
RESULT wave2 streams=0 dups=0 log_unchanged=false
RESULT wave1 streams=2 dups=0 archive_files=2 corruption=2 archive_bytes=709208 gzip_headers=16
archive corruption: [2026-07.jsonl.gz: stream 2: gzip: invalid checksum 2026-08.jsonl.gz: stream 2: flate: corrupt input before offset 1]
RESULT wave2 streams=10 dups=7 log_unchanged=false
RESULT wave1 streams=1 dups=0 archive_files=1 corruption=2 archive_bytes=709208 gzip_headers=16
RESULT wave2 streams=5 dups=3 log_unchanged=false
RESULT wave1 streams=0 ... corruption=2 archive_bytes=709208 gzip_headers=16   (runs 4, 5 similar; wave2 streams=8/dups=7 and streams=2/dups=1)
```
Confirmed independently: old code leaves 16 gzip headers (8 processes x 2 months) in every run, corrupts both monthly archives in 5 of 5 runs, and rewrites the log again in the second wave in 5 of 5 runs. The run agent's corruption claim holds. The numbers in `red-baseline.md` (709208 bytes, 16 headers, corruption=2, `log_unchanged=false` 5 of 5, wave-2 streams 8/8/4/0/0 with dups 7/7/3) match the shape I observed; the individual wave-2 stream counts differ run to run because they depend on interleaving, which is expected for a race.

**E5 - GREEN on HEAD, flakiness (item 2), `-count=20`** (`go test -timeout 8m -count=20 -v -run TestPruneConcurrentProcessesSingleRewrite ./internal/harness/`, exit 0, slot `go-test-internal-harness` held):
```
$ grep -c '^--- PASS' count20.out
20
  20 RESULT wave1 streams=2 dups=0 archive_files=2 corruption=0 archive_bytes=88651 gzip_headers=2
  20 RESULT wave2 streams=2 dups=0 log_unchanged=true
ok  	github.com/modu-ai/moai-adk/internal/harness	7.203s
```
0 of 20 flaked on a loaded machine. Package run: `go test -count=1 ./internal/harness/` -> `ok 1.364s`.

**E6 - mutation table (item 2).** Method: `go test -overlay=<mutant>.overlay.json -timeout 4m -count=1 ./internal/harness/` replacing only `retention.go`; the worktree was never edited (`git status --short` empty before and after).

| Mutant | Change | Result | Killed by (verbatim) |
|---|---|---|---|
| M1 no-lock | `lockfile.Lock(sf)` -> `error(nil)` | KILLED (exit 1) | `--- FAIL: TestPruneConcurrentProcessesSingleRewrite`; `RESULT wave1 streams=0 dups=0 archive_files=0 corruption=2 archive_bytes=709208 gzip_headers=16` (same signature as the RED) |
| M2 no re-check under lock | `if false && stampIsFresh(readStamp(sf), now)` | KILLED | `--- FAIL: TestPruneStamp_RecheckUnderLockSkips` (the multi-process test does NOT catch it) |
| M3 never write the stamp | `stampErr := writeStamp(...)` -> `var stampErr error` | KILLED | `--- FAIL: TestPruneConcurrentProcessesSingleRewrite` (`wave2 streams=3 log_unchanged=false`), `--- FAIL: TestPruneStamp_PersistedAcrossInstances` |
| M4 stamp only on success | write the stamp only if `pruneErr == nil` | **SURVIVES** | `ok github.com/modu-ai/moai-adk/internal/harness 1.216s` |
| M5 future stamp counts as fresh | drop `age >= 0` | KILLED | `--- FAIL: TestPruneStamp_FutureStampTreatedAsExpired` |
| M6 skip in-memory fast path | `if false && !r.lastPruneAt.IsZero()...` | SURVIVES (equivalent: the disk stamp subsumes it) | `ok ... 1.062s` |
| M7a mtime instead of injected clock (naive) | `mtimeFresh(path, now)` | KILLED | 6 FAILs incl. `TestPruneStaleEntriesRemovesOldEvents`, `TestPruneSkipsIfRecentlyPruned`, `TestIntegration_RetentionWithObserver`, `TestPruneStamp_PersistedAcrossInstances` |
| M7b mtime, non-negative age | `age >= 0 && age < 1h` on mtime | KILLED | `--- FAIL: TestPruneConcurrentProcessesSingleRewrite` (`wave2 streams=3 log_unchanged=false`), `TestPruneStamp_RecheckUnderLockSkips`, `TestPruneStamp_PersistedAcrossInstances` |
| M8 stamp BEFORE the prune | stamp then `r.prune(...)` | SURVIVES (ordering unpinned) | `ok ... 1.602s` |
| M9 re-check under lock with the PRE-lock clock | `now := preLock` | **SURVIVES** | `ok ... 1.244s` (tests use a fixed clock, so a stale pre-lock `now` is indistinguishable) |
| M10 drop the lock-free fast path | `if false && stampIsFresh(readStampFile(statePath), now)` | **SURVIVES** | `ok ... 1.423s` (hot-path cost is unpinned) |
| M11 prune UNLOCKED when the state file cannot be opened | `return r.prune(...)` on open error | **SURVIVES** | `ok ... 1.228s` (the unwritable test makes the dir read-only, so the prune fails anyway) |
| M12 stamp with wall clock instead of `nowFn` | `time.Now()` in `writeStamp` | KILLED | `TestPruneConcurrentProcessesSingleRewrite`, `TestPruneStamp_PersistedAcrossInstances` |

7 killed (M1, M2, M3, M5, M7a, M7b, M12), 6 survive (M4, M6, M8, M9, M10, M11). M6 is an equivalent mutant (no defect). M8 is an ordering choice, not a regression (see F1). M4, M9, M10, M11 are genuine test gaps (F2, F3). Coverage of the changed code (`go test -coverprofile`, `go tool cover -func`): `PruneStaleEntries 100.0%`, `pruneExclusive 85.0%`, `readStampFile 100.0%`, `stampIsFresh 100.0%`, `writeStamp 75.0%`, `prune 76.9%`; package `coverage: 87.3% of statements` (>= 85).

**E7 - the killed-pruner scenario, measured** (scratch tests added through the overlay, never in the tree; synthetic log of 200,000 events / 55.6 MB, 25,000 stale = 12.5 percent, comparable to the lane's +6 day row):
```
AUDIT-E partition=664.578042ms archive=59.117583ms rewrite=854.54275ms (kill window archive+rewrite=913.660333ms of total 1.578238375s)
AUDIT-E one full PruneStaleEntries end to end=1.115554834s
AUDIT-E fast path (fresh stamp, 65MB log never read) mean=15.212µs over 2000 calls
AUDIT-C after pruner #1: streams=2 dups=100 corruption=0
AUDIT-C2 kill #1: archive streams=1 dups=0 (log still holds stale events, no stamp: true)
AUDIT-C2 kill #2: archive streams=2 dups=100 ...
AUDIT-C2 kill #3: archive streams=3 dups=200 ...
AUDIT-C2 kill #5: archive streams=5 dups=400 (log still holds stale events, no stamp: true)
```
The window between "archive appended" and "rename done" is about 58 percent of the whole prune (0.91 s of 1.58 s here), and the lane measured 1.79 s for 12.5 percent stale on the live-log copy, i.e. about 36 percent of the 5 s hook timeout on an idle machine. A pruner killed inside that window writes no stamp, so every later hook finds "no stamp", wins the lock and archives the same events again (100 stale events -> 100 duplicates per kill; the production batch would repeat ~1 day of events or more per kill) and, if killed mid-rewrite, leaves a `usage-log-*.tmp` the size of the kept log. Nothing sweeps orphan tmp files.

**E8 - other scratch probes (same overlay method)**
```
AUDIT-A symlink: prune err=<nil> victim now="2026-10-02T00:00:00Z"
```
A symlink planted at `<log>.prune-state` is followed (`O_RDWR|O_CREATE`), then truncated and overwritten with the stamp: the victim's 56 bytes became the 20-byte timestamp.
```
AUDIT-B readonly-state call 0: err=retention: prune state open failed: ... permission denied log_has_stale=true   (calls 1 and 2 identical)
```
A non-writable existing state file (for instance left by another user) disables pruning permanently and silently (the observer discards the error). Cheap each time (no log read), but retention never runs.
```
AUDIT-D call 0 (+0s): err=retention: 이벤트 분류 실패: 파일 스캔: bufio.Scanner: token too long
AUDIT-D call 1 (+1m0s): err=<nil>
AUDIT-D call 2 (+30m0s): err=<nil>
AUDIT-D call 3 (+1h1m0s): err=retention: ... token too long
```
Stamp-after-failure works as designed: a persistently failing prune (a log line over 64 KiB, a pre-existing `bufio.Scanner` limit) is retried hourly, not by every hook. No delivered test pins this (M4 survives).
```
AUDIT-F waiter still blocked after 701.558625ms (no timeout, no try-lock)
AUDIT-F waiter finished after release: err=<nil> total=706.955959ms
```
A waiter blocks until the holder releases; the kernel releases an flock when a process dies, so a waiter killed at 5 s costs nothing and the next waiter proceeds.

**E9 - static and build checks (item 5)**
```
$ go vet ./internal/harness/...                       -> exit 0
$ golangci-lint --version                             -> golangci-lint has version v2.1.6 built with go1.26.8
$ grep -n golangci-lint .github/workflows/ci.yml      -> 464: run: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6
$ golangci-lint run ./internal/harness/...            -> exit 0, "0 issues."
$ GOOS=windows GOARCH=amd64 go build ./internal/harness/...  -> exit 0
$ GOOS=windows GOARCH=amd64 go vet ./internal/harness/       -> exit 0
$ gofmt -l internal/harness                           -> exit 0, no output
```
The repo `.golangci.yml` enables only `errcheck, govet, ineffassign, staticcheck, unused` (`default: none`). The IDE hints the lane saw (`scannererr` near line 218-226, `rangeint` at 102/125/290/307 of `retention_concurrent_test.go`) belong to gopls analyzers that are not in that set; v2.1.6 with the repo config reports none of them, CI runs the same command (`golangci-lint run --timeout=5m`), so they do not matter for CI. They are style hints only (`for i := 0; i < n; i++` could be `for i := range n`; `logSubjects` does not read `sc.Err()`).

**E10 - process evidence (item 7)**
- Line facts re-read at the base: `observer.go` calls `PruneStaleEntries` at lines 93 and 141 (`git show c640b0193:internal/harness/observer.go | grep -n`), `hook.go` builds `NewRetention` at 822, 919, 1056, 1220 (grep), `internal/lockfile` exposes only blocking `Lock`/`Unlock` (Windows = in-process mutex per path). The only non-test caller of `PruneStaleEntries` is the observer, so the stamp suppresses no explicit user-initiated prune command.
- 65,506,619 bytes is a `stat` figure the lane took on 2026-10-02; I cannot re-measure that instant (the live log has grown, the lane's later copy read 65,804,810 bytes, consistent with growth). Not re-observed by me.
- Jev answers: `scratchpad/t1425/jev-answer.json` -> `model jev-1.13.0`; `where_prune_runs=gate_in_hook` (0.96), `lock_behavior=block_then_recheck` (1.00). Both are design choices (where to prune, whether a waiter blocks), not completion/merge/queue judgments. Observation: the state text told Jev "Planned fix already decided"; the `lock_behavior` criteria describe `block_then_recheck` with its benefit ("No change to internal/lockfile") and `skip_if_held` with its cost ("needs a new non-blocking try-lock"), and the seven facts omit the 5 s hook timeout, which is the fact that makes a blocking waiter's cost visible (F6).
- Trailers: `Authored-By-Agent: manager-develop` is present (consumer predicate, `git log --grep`) on `785d70fbf`, `fe211e9c9`, `9fbf6d0a8`; absent on the lane-authored docs commits `be6c80f16` and `0fa7f49e8`. `%(trailers:)` prints empty for all five because `🗿 MoAI` follows the trailer (known git-parser limitation, adjudicated earlier; the consumer reads the body with its own regex). The run role was a `general-purpose` spawn playing manager-develop (documented worktree-isolation workaround), so the trailer names a role identity, not the literal agent type; this card has no SPEC directory, so no phase-owned artifact exists for the ownership matrix to arbitrate.

**E11 - adjacent tests (item 8, run by me)**
```
internal/hook  -run 'TestPostToolFailure|TestEvidenceWriter'                          -> ok 0.740s (slot go-test-internal-hook)
internal/cli   -run 'TestTierPromotionHighWater_Repro|TestHarnessMutePreserve_Repro|TestCleanReinstall_PreservesUserArea|TestBackupSubsystem_DestructiveSurfaces|TestMigrateLegacyMemoryDir_PreservesUserArea'  -> ok 0.863s
internal/cli   -run Harness                                                            -> ok 30.432s (slot go-test-internal-cli; re-run because it is cheap and the lane's figure was 21.954s)
```
One extra run with the pattern `Preserve` also matched two `internal/cli` todo tests that failed with `moai add: refused - lane boundary: a lane session cannot mutate the queue`. That is the lane-session guard reacting to the factory env that my `unset` line did not clear (`MOAI_FACTORY_*` is not in the scrub list); it concerns the todo queue, not retention, and I excluded those two tests from the pattern. Not a finding against this card.

### Baseline-attribution
- Tree measured: worktree `.moai/worktrees/t1425`, `HEAD = 0fa7f49e807887f295026468aab83d80584ba1b1`, `git status --short` empty at the start and at the end (checked after the last run).
- Build that judged: every test figure is the `go test` binary compiled from this tree (or from the tree with ONE file replaced via `-overlay`, as marked); `go1.26.8 darwin/arm64`; `golangci-lint v2.1.6` (installed build, equal to the CI pin at `.github/workflows/ci.yml:464`). The installed `moai` binary was used only for `moai slot` leases and is not corroboration of anything here.
- The overlay mutants and the scratch probes were measured against the same tree; the old-code RED was measured by overlaying `git show c640b0193:internal/harness/retention.go`.
- Synthetic fixtures (55.6 MB, 200,000 events) are not the live log; timings are from a loaded machine and are one-shot per case (single measurement; not repeated).

### Gaps
- Not observed: the live 70 MB log with hundreds of hooks; a prune on a machine under the incident load; Claude Code actually killing a pruner or a waiter at the 5 s hook timeout (I simulated "killed after archive" by calling the internal steps, I did not kill a real process mid-rename).
- Not observed: Windows at runtime (only `GOOS=windows go build` and `go vet`); the multi-process test is `!windows`.
- Not re-measured: the 65,506,619-byte `stat` and the original machine-load numbers (they describe a past instant).
- Not run: `go test ./...`, the whole `internal/cli` suite, `internal/harness/rosterguard`, the race-detector run (the lane reported `-race` clean; I did not repeat it).
- Not examined: the live primary checkout's `.moai/harness/` contents (out of scope for a worktree-isolated audit); the existing 1,051 orphan `usage-log-*.tmp` files / 35 GB from the incident, which need an operator cleanup the card did not ask this change to perform.
- Refusals recorded: the worktree guard refused (a) `bash <path>/run-mutants.sh` ("construct too complex to verify"), (b) one `awk` program, (c) a `for ...; do git log ... done` loop. I did what the guard asked instead: each mutant became its own plain `go test -overlay=...` command and the per-commit trailer checks became `git log --grep` / `--format` calls. No verdict claim rests on a refused command's inferred output.
- Not independently verified: Jev's confidence values beyond reading `jev-answer.json`; the claim that the 35 GB / load-840 incident ran on this exact code path (I read the code path, I did not reproduce the incident).

### Residual-risk
- Killed or slow pruner re-archives the same events and can leave orphan tmp files (F1). Bounded by lock serialization, not eliminated.
- Event loss window: events appended by non-locking hooks between the pruner's read and rename are lost; in steady state events cross the 30-day cutoff continuously, so the whole-log rewrite recurs about hourly (about 1.1 to 1.6 s window per hour here). Disclosed by the `@MX:WARN` and the run evidence.
- Windows: no cross-process lock, and the stamp is written only AFTER the prune, so a burst of hooks at a stamp expiry still launches N concurrent prunes once per interval (F5).
- Pre-existing, unchanged: `partitionEvents` silently drops unparsable lines and fails on any line over 64 KiB; `overwriteWithEvents` creates the replacement log with `CreateTemp` mode 0600.

## 2. Acceptance table against the card

| # | Card requirement | Verdict | Evidence |
|---|---|---|---|
| A1 | Last prune time on disk | PASS | `<log>.prune-state` holds an RFC3339Nano stamp from `nowFn`; `TestPruneStamp_PersistedAcrossInstances` green; M3/M7a/M12 killed |
| A2 | File lock, single pruner per interval | PASS on Unix (partial on Windows) | `lockfile.Lock` blocking flock + re-check; E5 wave1 `streams=2 dups=0 corruption=0` x20; M1 killed; Windows lock is in-process only (disclosed; F5 refines the claim) |
| A3 | Hot-path cost kept low (timeout 5, async) | PASS (unpinned) | fast path reads <= 128 bytes, no lock, no log read: mean 15.2 µs over 2000 calls (E7); M10 survives, no regression guard |
| A4 | Load-regression guard N concurrent hooks -> ONE rewrite | PASS | `TestPruneConcurrentProcessesSingleRewrite`, 8 re-executed binaries, wave1 + wave2; 20/20 green, M1 and M3 killed |
| A5 | Observed RED first | PASS | `785d70fbf` precedes `fe211e9c9`, holds no test or production file; RED re-derived on old code: 5/5 FAIL, `gzip_headers=16`, `corruption=2` (E3, E4) |
| A6 | After the repair one rewrite | PASS | E5: `wave1 streams=2 dups=0 gzip_headers=2`, `wave2 log_unchanged=true` in 20/20 |
| A7 | Deadline: develop landing before 2026-10-04T09:18Z | N/A to the audit | landing is the leader's act |

## 3. Findings

Severity scale: blocker / major / minor. Class: blocking (correctness or a stated requirement) / optional (hardening, discretionary).

- **F1 [major] [optional, recommended before landing]** `internal/harness/retention.go:128-135` (`pruneExclusive`) - the stamp is written AFTER `r.prune`. A pruner that is killed or crashes after `archiveEvents` and before the rename leaves no stamp and an unchanged log, so the next hook re-archives the same events (measured: 100 stale events -> 100 duplicates per kill, `AUDIT-C2` kill #5 `dups=400`) and a kill during the rewrite leaves a `usage-log-*.tmp` of up to the kept-log size; no code sweeps orphan tmp files. The kill window is about 58 percent of a prune (0.91 s of 1.58 s), and a prune that takes longer than the remaining hook budget (5 s) is killed every time, so the repeat does not converge while the machine stays slow. The lane disclosed the re-archive but did not quantify it, give a mitigation, or mention the tmp leak. Required fix (not applied; findings only): write the attempt stamp BEFORE the prune (M8 shows every delivered test still passes with that ordering), so a killed pruner costs one repeat per interval, and consider removing `usage-log-*.tmp` older than the interval at the start of the slow path. Trade-off to record: a pruner killed before the archive then defers retention by one hour, which is harmless against a 30-day window. Why not blocking: the card's stated requirement (one prune and one rewrite per interval) holds in every measured non-kill run, and the delivered ordering is a defensible choice that the lane recorded in `lane-measurements.md`.
- **F2 [minor] [optional]** `retention_stamp_test.go` - no test pins the stamp-after-failed-prune behaviour that the commit message and the code comment claim (M4 survives; `AUDIT-D` shows it works today). A regression would let a persistently failing prune (for example a log line over 64 KiB) be retried by every hook, which is the original storm. Required fix: one unit test with a log line longer than 64 KiB asserting the second call within the interval returns nil and does not read the log again.
- **F3 [minor] [optional]** `retention_stamp_test.go:205` (`TestPruneStamp_UnwritableStateSkipsPruneAndRecordSucceeds`) - the test makes the whole directory read-only, so even a prune that ignored the lock (M11) fails and the test passes. The decision "never prune without the lock" is therefore not pinned. Required fix: make the state path unopenable while the directory stays writable (state file mode 0444, or the state path being a directory), assert the stale event is still in the log. Also: no test uses a clock that advances, so the fresh-clock re-check (M9) and the lock-free fast path (M10) have no regression guard.
- **F4 [minor] [optional]** `retention.go:99` (`os.OpenFile(statePath, O_RDWR|O_CREATE, 0o644)`) - confirmed by `AUDIT-A`: a symlink at `<log>.prune-state` is followed and the target is truncated and overwritten with the stamp. The existing code already follows symlinks when it appends to the log and the archive, and the state path sits in the project's own `.moai/harness/`, so the exposure is a hostile repository that ships the link; truncation is, however, more destructive than the existing appends. Required fix if wanted: open with `O_NOFOLLOW` on Unix (build-tagged) or `Lstat` the state path and refuse a non-regular file.
- **F5 [minor] [optional]** Windows - `lockfile.Lock` is an in-process mutex, and the stamp is written only after the prune, so N hook processes that start inside the same expired interval all read "no stamp" and all prune concurrently, once per interval. `run-evidence.md` says the stamp alone gives "one rewrite per interval per quiet period"; the `@MX:REASON` says "the stamp alone bounds the rewrites". Both are true only for sequential hooks, not for a burst at expiry. Disclosure refinement, not a code change; F1's stamp-before-work reduces the window to the time between the stamp check and the stamp write.
- **F6 [minor] [optional]** Lock waiters have no timeout and no try-lock (`AUDIT-F`: blocked until release). Every hook arriving during a prune, once per hour, waits for the whole prune and is killed at 5 s if the prune is slower. Jev chose `block_then_recheck` (1.00) without being told about the 5 s hook timeout, and the criteria text favoured it; the choice is defensible (the kernel releases the flock on death, `async: true` hooks do not block the agent) but a non-blocking skip would cap hook latency at zero and needs a Unix `LOCK_NB` helper. Accepted residual; the record should say the timeout fact was missing from the question.
- **F7 [minor] [optional]** `retention.go:99` - a state file that exists but is not writable by the current user (left by another user, or created under `sudo`) disables retention silently and permanently (`AUDIT-B`, three calls identical; the observer discards the error with `_ =`). Pre-fix this state could not occur. Not disclosed by the delivery. Required fix if wanted: fall back to a read-only stamp check plus a logged warning, or log the failure once to `.moai/logs/`.
- **F8 [info]** `NewRetention` 66.7 percent and `overwriteWithEvents` 55.6 percent coverage are pre-existing gaps in unchanged code; `pruneExclusive` 85.0 percent (the `writeStamp` and prune-error branches are the uncovered ones, consistent with F2).
- **F9 [info]** The IDE diagnostics in `retention_concurrent_test.go` (`scannererr`, four `rangeint`) are gopls hints outside the CI linter set; v2.1.6 with the repo config reports `0 issues.`.

## 4. Honest-gap judgment (item 8)

| Gap | Cheap read-only check | Done? |
|---|---|---|
| Live 70 MB log, hundreds of hooks | none that is safe from a worktree; a synthetic 55.6 MB log stood in (E7) | partial |
| Prune on a loaded machine | the audit ran on the loaded machine; one-shot timings only | partial |
| Lock waiter killed at 5 s | kernel releases flock on death (documented behaviour); `AUDIT-F` shows the waiter blocks and resumes cleanly | reasoned + probed |
| Windows runtime | not available here; `GOOS=windows` build and vet pass | no |
| `internal/cli` tests mentioning the usage log | named tests run by me (E11), all green | done |
| `internal/hook/failure_event_test.go` | `TestPostToolFailure*`, `TestEvidenceWriter*` run, green | done |
| `update_preserve_reach_test.go` | its three tests run, green | done |

## 5. Verdict

**PASS-WITH-DEBT, 87/100.** The core repair is real and correct: the on-disk stamp, the flock with a double check, the lock-free fast path and the multi-process guard all work, the RED is genuine and reproduced independently on the old code (archive corruption confirmed), the commit graph orders baseline before fix, and the scope is clean. The debt: F1 (killed-pruner repeat and orphan tmp, recommended fix is a two-line reordering the existing tests already tolerate), and a set of test-gap and hardening items (F2-F7). No finding is blocking, so the verdict is not FAIL; the leader should decide whether F1 goes in before the 2026-10-04T09:18Z landing deadline, because the incident that created this card is the very condition (slow machine, 5 s kills) under which F1 bites.

Worktree state at the end: `git status --short` empty, HEAD unchanged (`0fa7f49e807887f295026468aab83d80584ba1b1`); all three slot leases (`go-test-internal-harness`, `go-test-internal-hook`, `go-test-internal-cli`) released and reported `free`.
