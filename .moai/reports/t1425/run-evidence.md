# t1425 run evidence - single-writer retention prune

card: t1425 (class B, urgent, no SPEC). Branch `WT-harness-prune-single-writer`.
RED baseline commit: `785d70fbf` (measured on `be6c80f16`, unmodified `retention.go`; see `red-baseline.md`).
Fix commit: `fe211e9c9` (parent `785d70fbf`). All GREEN figures below were measured on the tree of `fe211e9c9` with a clean `git status` for the measured files, `go1.26.8 darwin/arm64`. The tool measured is the `go test` binary built from this tree each run (no installed `moai` build is involved in any figure).

Every `go test` below was run as `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && timeout 900 go test -timeout 10m ...` under the slot leases `go-test-internal-harness` and `go-test-internal-cli` (acquire exit 0, release exit 0 for both).

## Claim

N=8 concurrent hook-like processes now produce exactly ONE gzip stream per month, zero duplicate archived events, a log holding exactly the kept events, no leftover temp file, and a second wave inside the skip interval changes nothing.

## RED versus GREEN, multi-process guard `TestPruneConcurrentProcessesSingleRewrite`

| | RED (unmodified, 5 runs, `-count=5`) | GREEN (fix, 20 runs, `-count=20`) |
|---|---|---|
| exit code | 1 (5 of 5 FAIL) | 0 (20 of 20 PASS) |
| gzip headers in the archive after wave 1 (expected 2) | 16 in every run (8 processes x 2 months) | 2 in every run |
| readable streams after wave 1 | 0 (both archives corrupt) | 2 (one per month) |
| duplicate archived events | not measurable (archive unreadable) | 0 in every run |
| archive corruption | 2 of 2 files, every run | 0 |
| archive bytes | 709208 | 88651 |
| wave 2 (clock +10 min, one new stale event): log unchanged | `false` in 5 of 5 | `true` in 20 of 20 |
| wave 2: streams / dups | 8/7, 8/7, 4/3, 0/0, 0/0 (June archive partly unreadable) | 2 / 0 in every run |

GREEN summary, verbatim from the `-count=20 -v` run (`grep RESULT ... | sort | uniq -c`):

```
  20 RESULT wave1 streams=2 dups=0 archive_files=2 corruption=0 archive_bytes=88651 gzip_headers=2
  20 RESULT wave2 streams=2 dups=0 log_unchanged=true
```

`--- PASS: TestPruneConcurrentProcessesSingleRewrite` count: 20. Command: `go test -timeout 10m -count=20 -v -run TestPruneConcurrentProcessesSingleRewrite ./internal/harness/` -> exit 0.

Race detector (children are race-instrumented binaries as well): `go test -timeout 10m -race -count=1 -v -run TestPrune ./internal/harness/` -> exit 0, `ok github.com/modu-ai/moai-adk/internal/harness 6.148s`, no `DATA RACE`, wave 1 `streams=2 dups=0 corruption=0`, wave 2 `log_unchanged=true`.

## Package results

`go test -timeout 10m -count=1 ./internal/harness/... ./internal/lockfile/...` -> exit 0; every package `ok` (the harness root package 1.880s; `harness/rosterguard` 32.650s; `internal/lockfile` 0.711s).

`go test -timeout 10m -count=1 -run Harness ./internal/cli/` -> exit 0, `ok github.com/modu-ai/moai-adk/internal/cli 21.954s` (the hook observer tests that go through `RecordEvent` / `RecordExtendedEvent`).

## Guard sensitivity (mutation checks on the fix, then restored byte-for-byte, `cmp` exit 0)

| mutation | result |
|---|---|
| lock acquisition removed (`lockfile.Lock` replaced by `error(nil)`) | multi-process guard FAILS 3 of 3 runs: `wave1 streams=0 corruption=2 archive_bytes=709208 gzip_headers=16` (same signature as the RED) |
| post-lock re-check disabled (`if false && stampIsFresh(...)`) | multi-process guard still PASSES (a waiter then partitions an already-pruned log and finds nothing, so the re-check is an optimisation, not a correctness property in this fixture); the new unit test `TestPruneStamp_RecheckUnderLockSkips` FAILS (`pruneExclusive pruned although a fresh stamp was already on disk`) |

## Static checks

| command | exit | note |
|---|---|---|
| `go vet ./internal/harness/...` | 0 | |
| `golangci-lint run ./internal/harness/` | 0 | `0 issues.`; installed `v2.1.6`, CI pins `@v2.1.6` (`.github/workflows/ci.yml:464`), identical |
| `GOOS=windows GOARCH=amd64 go build ./internal/harness/...` | 0 | |
| `GOOS=windows GOARCH=amd64 go vet ./internal/harness/` | 0 | also compiles the test files for windows (the multi-process test is `!windows`) |
| `gofmt -l internal/harness` | 0, no output | |

## Tests added

`internal/harness/retention_concurrent_test.go` (`!windows`): `TestPruneConcurrentProcessesSingleRewrite`, `TestPruneHelperProcess` (child body, skips in a normal run).
`internal/harness/retention_stamp_test.go`: `TestPruneStamp_PersistedAcrossInstances`, `TestPruneStamp_FutureStampTreatedAsExpired`, `TestPruneStamp_CorruptOrPartialStampIgnored` (subtests `garbage`, `empty`, `truncated`), `TestPruneStamp_StampWrittenWhenNothingToPrune`, `TestPruneStamp_RecheckUnderLockSkips`, `TestPruneStamp_MissingLogCreatesNoState`, `TestPruneStamp_UnwritableStateSkipsPruneAndRecordSucceeds`.

## Gaps

- Not observed: the live 70 MB log, hundreds of simultaneous hook processes, and the real Claude Code hook timeout while a lock waiter blocks behind a long prune. The fixture is 20,000 events and 8 processes.
- Not observed on Windows at runtime: only `go build` / `go vet` with `GOOS=windows` were run. The Windows lock is in-process only, so there the guarantee is the stamp (one rewrite per interval per quiet period), not mutual exclusion.
- `internal/cli` was run only for tests whose names contain `Harness`, not the whole package; `update_preserve_reach_test.go`, `clifix_critical_repro_test.go` and `internal/hook/failure_event_test.go` mention the usage log and were not run.
- The unwritable-state unit test is skipped when run as root or on Windows.

## Residual-risk

- The pruner still reads the whole log and replaces it by rename; an event appended by a non-locking hook between the read and the rename is lost (documented in the `@MX:WARN` / `@MX:REASON` of `PruneStaleEntries`). One such window per hour replaces N windows per hook burst.
- Lock waiters block in `flock`; a long prune delays every concurrent hook by its duration once per interval.
- Pre-existing, out of scope, unchanged: `partitionEvents` drops lines that fail JSON parsing although its comment says they are kept.
- `.moai/harness/usage-log.jsonl.prune-state` is a new runtime file next to the log in user projects; the template ignore file carries no usage-log line, so no template mirror was changed.
