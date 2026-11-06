# t2068 — TestManagedCardChildDeliversInboxThroughLoop "no such table: peers"

## Cause (read from CI log + code)
- CI (job 127665171114, ubuntu cli-3, race): `managed_card_child_test.go:237 managed owner sessions ended before binding; stdout="" stderr=""` joined with `SQL logic error: no such table: peers (1)`.
- `factorymsg.OpenExistingWithContext` (store.go) stats `broker.db`, then runs `SELECT count(*) FROM peers`. `openResolvedWithContext` creates the file with its first connection (pragmas) and only afterwards runs the `schema` statement. A file that exists with no schema committed yet makes the existing-open read fail with "no such table: peers". Under the race detector the window is wider.
- Attribution: the hook path (`factoryHookOpenInbox = OpenExistingWithContext`) is the consumer of that function. That the CI failure went through exactly this path is inferred from the identical error string, not observed (no CI stack trace).

## Reproduction
- Local, target test, race: `go test ./internal/cli -run TestManagedCardChildDeliversInboxThroughLoop -race -count=20` → `ok ... 75.018s` (pre-fix). The CI failure did NOT reproduce locally at the cli level.
- Deterministic reproduction of the window one level down: new `TestOpenExistingToleratesABrokerFileWithoutSchemaYet` (empty broker.db). RED commit 66337060f6: `SQL logic error: no such table: peers (1)` (same text as CI).

## Repair (minimal)
- `internal/factorymsg/store.go`: in `OpenExistingWithContext`, a "no such table" from the peers read falls back to `openResolvedWithContext` (CREATE IF NOT EXISTS, serialized by the immediate lock). 1 file, +9 lines; plus the RED test file.

## Verification (this tree, this run)
- new test: PASS.
- `go test ./internal/factorymsg -race -count=1` → ok 128.704s
- `go test ./internal/hook -run Factory -race -count=1` → ok 72.530s
- `go test ./internal/cli -run TestManagedCardChild -race -count=5` → ok 63.910s
- gofmt -l internal/factorymsg: no output; go vet ./internal/factorymsg: no output.

## Gaps
- The CI failure itself was not reproduced at the cli level; the cause is established at the factorymsg level and attributed by identical error text.
- `internal/cli` full suite not run (leader RC41 constraint). card-review not yet run.
- Editor diagnostics (gopls, workspace) reported undefined names in store.go; compile and tests succeeded, treated as workspace noise.

## Residual risk
- Other readers that stat the file then read tables could share the window; only OpenExisting was changed.
