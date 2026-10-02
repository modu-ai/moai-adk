# t1425 observed RED - concurrent retention pruning on UNMODIFIED code

Measured on HEAD `be6c80f16` (branch `WT-harness-prune-single-writer`, base `c640b0193`), worktree `.moai/worktrees/t1425`, `go1.26.8 darwin/arm64`, machine under heavy load. `internal/harness/retention.go` is byte-identical to HEAD at measurement time (`git status --short` listed only the two new untracked test files). The test files are NOT in this commit; they stay uncommitted until the fix commit.

Slot lease taken first: `moai slot acquire --resource go-test-internal-harness --max-duration 25m` -> exit 0.

## Claim

Unmodified `PruneStaleEntries` lets N concurrent hook processes all prune: every process appends its own gzip streams to the monthly archive and rewrites the log. Under 8 concurrent processes the archive ends up corrupt (interleaved appends from several processes), not merely duplicated.

## Evidence 1 - multi-process guard, 5 runs

Command (kanban env scrubbed in the same invocation):

```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && timeout 900 go test -timeout 10m -count=5 -v -run TestPruneConcurrentProcessesSingleRewrite ./internal/harness/
```

Exit code: 1 (all 5 runs FAIL). Verbatim output:

```
=== RUN   TestPruneConcurrentProcessesSingleRewrite
    retention_concurrent_test.go:336: RESULT wave1 streams=0 dups=0 archive_files=0 corruption=2 archive_bytes=709208 gzip_headers=16
    retention_concurrent_test.go:339: archive corruption: [2026-07.jsonl.gz: stream 1: flate: corrupt input before offset 1 2026-08.jsonl.gz: stream 1: flate: corrupt input before offset 1]
    retention_concurrent_test.go:342: archive files: want 2 (one per month), got []
    retention_concurrent_test.go:354: stale event stale-02731 archived 0 times, want 1
    retention_concurrent_test.go:412: RESULT wave2 streams=8 dups=7 log_unchanged=false
    retention_concurrent_test.go:414: second wave changed the archive stream count: 0 -> 8
    retention_concurrent_test.go:417: second wave rewrote the log inside the skip interval
--- FAIL: TestPruneConcurrentProcessesSingleRewrite (1.33s)
=== RUN   TestPruneConcurrentProcessesSingleRewrite
    retention_concurrent_test.go:336: RESULT wave1 streams=0 dups=0 archive_files=0 corruption=2 archive_bytes=709208 gzip_headers=16
    retention_concurrent_test.go:339: archive corruption: [2026-07.jsonl.gz: stream 1: flate: corrupt input before offset 1 2026-08.jsonl.gz: stream 1: flate: corrupt input before offset 1]
    retention_concurrent_test.go:342: archive files: want 2 (one per month), got []
    retention_concurrent_test.go:354: stale event stale-03023 archived 0 times, want 1
    retention_concurrent_test.go:412: RESULT wave2 streams=8 dups=7 log_unchanged=false
    retention_concurrent_test.go:414: second wave changed the archive stream count: 0 -> 8
    retention_concurrent_test.go:417: second wave rewrote the log inside the skip interval
--- FAIL: TestPruneConcurrentProcessesSingleRewrite (1.52s)
=== RUN   TestPruneConcurrentProcessesSingleRewrite
    retention_concurrent_test.go:336: RESULT wave1 streams=0 dups=0 archive_files=0 corruption=2 archive_bytes=709208 gzip_headers=16
    retention_concurrent_test.go:339: archive corruption: [2026-07.jsonl.gz: stream 1: flate: corrupt input before offset 1 2026-08.jsonl.gz: stream 1: flate: corrupt input before offset 1]
    retention_concurrent_test.go:342: archive files: want 2 (one per month), got []
    retention_concurrent_test.go:354: stale event stale-05517 archived 0 times, want 1
    retention_concurrent_test.go:412: RESULT wave2 streams=4 dups=3 log_unchanged=false
    retention_concurrent_test.go:414: second wave changed the archive stream count: 0 -> 4
    retention_concurrent_test.go:417: second wave rewrote the log inside the skip interval
--- FAIL: TestPruneConcurrentProcessesSingleRewrite (1.10s)
=== RUN   TestPruneConcurrentProcessesSingleRewrite
    retention_concurrent_test.go:336: RESULT wave1 streams=0 dups=0 archive_files=0 corruption=2 archive_bytes=709208 gzip_headers=16
    retention_concurrent_test.go:339: archive corruption: [2026-07.jsonl.gz: stream 1: gzip: invalid checksum 2026-08.jsonl.gz: stream 1: flate: corrupt input before offset 112131]
    retention_concurrent_test.go:342: archive files: want 2 (one per month), got []
    retention_concurrent_test.go:354: stale event stale-03000 archived 0 times, want 1
    retention_concurrent_test.go:412: RESULT wave2 streams=0 dups=0 log_unchanged=false
    retention_concurrent_test.go:417: second wave rewrote the log inside the skip interval
--- FAIL: TestPruneConcurrentProcessesSingleRewrite (1.20s)
=== RUN   TestPruneConcurrentProcessesSingleRewrite
    retention_concurrent_test.go:336: RESULT wave1 streams=0 dups=0 archive_files=0 corruption=2 archive_bytes=709208 gzip_headers=16
    retention_concurrent_test.go:339: archive corruption: [2026-07.jsonl.gz: stream 1: flate: corrupt input before offset 1 2026-08.jsonl.gz: stream 1: flate: corrupt input before offset 1]
    retention_concurrent_test.go:342: archive files: want 2 (one per month), got []
    retention_concurrent_test.go:354: stale event stale-05462 archived 0 times, want 1
    retention_concurrent_test.go:412: RESULT wave2 streams=0 dups=0 log_unchanged=false
    retention_concurrent_test.go:417: second wave rewrote the log inside the skip interval
--- FAIL: TestPruneConcurrentProcessesSingleRewrite (0.94s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/harness	6.775s
FAIL
```

## Per-run table (wave 1 = 8 processes race on the first prune; expected after a correct fix: 2 archive files, 1 stream each, 0 duplicates, 0 corruption)

Fixture per run: 20,000 events, 10,000 stale (5,000 in 2026-07, 5,000 in 2026-08) + 10,000 kept; fixed clock 2026-10-02T00:00:00Z, retention 30 days.

| run | gzip headers in archive (wave 1; expected 2) | readable streams (wave 1) | duplicates measurable | archive corruption | archive bytes |
|---|---|---|---|---|---|
| 1 | 16 | 0 | no, archive unreadable | 2 of 2 files | 709208 |
| 2 | 16 | 0 | no, archive unreadable | 2 of 2 files | 709208 |
| 3 | 16 | 0 | no, archive unreadable | 2 of 2 files | 709208 |
| 4 | 16 | 0 | no, archive unreadable | 2 of 2 files | 709208 |
| 5 | 16 | 0 | no, archive unreadable | 2 of 2 files | 709208 |

Reading of the table: 16 gzip header magics = 8 processes x 2 months, i.e. every one of the 8 processes appended its own stream to each monthly archive. Because the processes write concurrently with `O_APPEND` in several chunks each, the streams interleave and neither archive can be decompressed ("flate: corrupt input before offset 1" / "gzip: invalid checksum"). The duplicate count therefore cannot be read directly in wave 1; the proxy is the header count (16 vs 2) and the archive size.

Wave 2 (8 fresh processes, clock +10 minutes, one new stale event appended; a correct fix must leave the log and archive untouched): the log was rewritten in 5 of 5 runs (`log_unchanged=false`), and where the June archive stayed readable it shows the same single event archived 8 times (run 1 and 2: `streams=8 dups=7`) or 4 times (run 3: `streams=4 dups=3`); in runs 4 and 5 the June archive was itself unreadable (`streams=0`).

## Evidence 2 - stamp unit tests on unmodified code (expected-fail run)

```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && timeout 900 go test -timeout 10m -count=1 -v -run TestPruneStamp_ ./internal/harness/
```

Exit code: 1. Verbatim output:

```
=== RUN   TestPruneStamp_PersistedAcrossInstances
=== PAUSE TestPruneStamp_PersistedAcrossInstances
=== RUN   TestPruneStamp_FutureStampTreatedAsExpired
=== PAUSE TestPruneStamp_FutureStampTreatedAsExpired
=== RUN   TestPruneStamp_CorruptOrPartialStampIgnored
=== PAUSE TestPruneStamp_CorruptOrPartialStampIgnored
=== RUN   TestPruneStamp_StampWrittenWhenNothingToPrune
=== PAUSE TestPruneStamp_StampWrittenWhenNothingToPrune
=== RUN   TestPruneStamp_MissingLogCreatesNoState
=== PAUSE TestPruneStamp_MissingLogCreatesNoState
=== RUN   TestPruneStamp_UnwritableStateSkipsPruneAndRecordSucceeds
--- PASS: TestPruneStamp_UnwritableStateSkipsPruneAndRecordSucceeds (0.00s)
=== CONT  TestPruneStamp_PersistedAcrossInstances
=== CONT  TestPruneStamp_StampWrittenWhenNothingToPrune
=== CONT  TestPruneStamp_CorruptOrPartialStampIgnored
=== CONT  TestPruneStamp_MissingLogCreatesNoState
=== CONT  TestPruneStamp_FutureStampTreatedAsExpired
=== RUN   TestPruneStamp_CorruptOrPartialStampIgnored/garbage
=== PAUSE TestPruneStamp_CorruptOrPartialStampIgnored/garbage
=== RUN   TestPruneStamp_CorruptOrPartialStampIgnored/empty
=== PAUSE TestPruneStamp_CorruptOrPartialStampIgnored/empty
=== RUN   TestPruneStamp_CorruptOrPartialStampIgnored/truncated
=== PAUSE TestPruneStamp_CorruptOrPartialStampIgnored/truncated
=== CONT  TestPruneStamp_CorruptOrPartialStampIgnored/garbage
=== CONT  TestPruneStamp_CorruptOrPartialStampIgnored/truncated
=== CONT  TestPruneStamp_CorruptOrPartialStampIgnored/empty
--- PASS: TestPruneStamp_MissingLogCreatesNoState (0.00s)
=== NAME  TestPruneStamp_StampWrittenWhenNothingToPrune
    retention_stamp_test.go:159: no stamp after a no-op prune: stat /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestPruneStamp_StampWrittenWhenNothingToPrune2931505435/001/usage-log.jsonl.prune-state: no such file or directory
--- FAIL: TestPruneStamp_StampWrittenWhenNothingToPrune (0.01s)
=== NAME  TestPruneStamp_PersistedAcrossInstances
    retention_stamp_test.go:69: stamp file not written: stat /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestPruneStamp_PersistedAcrossInstances2726203965/001/usage-log.jsonl.prune-state: no such file or directory
--- FAIL: TestPruneStamp_PersistedAcrossInstances (0.01s)
--- PASS: TestPruneStamp_FutureStampTreatedAsExpired (0.01s)
--- PASS: TestPruneStamp_CorruptOrPartialStampIgnored (0.00s)
    --- PASS: TestPruneStamp_CorruptOrPartialStampIgnored/garbage (0.01s)
    --- PASS: TestPruneStamp_CorruptOrPartialStampIgnored/empty (0.01s)
    --- PASS: TestPruneStamp_CorruptOrPartialStampIgnored/truncated (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/harness	0.577s
FAIL
```

Of the six unit tests, 2 fail as designed (`PersistedAcrossInstances`, `StampWrittenWhenNothingToPrune`). The other four pass on the old code because they pin behaviour that must be PRESERVED (future/corrupt/partial stamp handling is vacuous when no stamp exists; a missing log creates nothing; an unwritable directory already errors out because `CreateTemp` fails).

## Baseline-attribution

Both commands were run in this session, against this tree (HEAD `be6c80f16`, `retention.go` unmodified). The binary under test is the freshly compiled `go test` binary of this tree, so no installed-build lag applies (the tool measured is the Go test binary built from the tree, not an installed `moai`).

## Gaps

- Duplicate-event counts for wave 1 are not directly observed (archive unreadable); the header count and corruption are the observed proxies.
- The live 70 MB log and the 35 GB temp-file incident were not reproduced (the live `.moai/harness` data is out of bounds); the fixture is 20,000 events.

## Residual-risk

The RED shows the race at 8 processes with a 4 MB fixture; the production storm used a ~70 MB log and 1,000+ processes, so the real overlap window is larger than measured here.
