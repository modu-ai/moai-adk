# M4 — Checkpoint four-stage API call counts (Issue #1948)

Claim: per-stage remote calls of a checkpoint recovery, counted against fakes.

Command: `go test -race -count=1 -v -run 'TestMeasureCheckpoint_' ./internal/agenticfactory/e2e/`
Tree: integration worktree, HEAD at the commit that adds `measure_checkpoint_test.go`
(includes lane-18 `e6499b731a` checkpoint model and lane-20 `91f6690d74`, `CountingRemote`).

## Observed (verbatim from the run)

| Stage | What runs | Remote calls observed |
|---|---|---|
| 1 commit | `Saver.Save` (claim store document) | 0 (local store) |
| 2 remote ACK | `Record.ConfirmRemote` | head=1, contains=1 (git lookups; 0 API calls) |
| 3 manifest | `Saver.Verify` (reread of the stored document) | 0 (local store) |
| 4 summary link, first publish | `PublishSummaryLink` | Find 1, Create 1, Update 0, total 2; 1 comment |
| 4 replay, same checkpoint | `PublishSummaryLink` | Find 1, Create 0, total 1; still 1 comment |
| 4 not durable yet | `PublishSummaryLink` | total 0 (skipped, nothing sent) |
| 4 create response lost, then retry | two publishes | after first: Find 2, Create 1; after retry: Find 3, Create 1; 1 comment |
| 4 comment failure (HTTP 502) | `PublishSummaryLink` | Find 2, Create 1; status `failed`; 0 comments; recovery verdict unchanged |

## Gaps (not observed)
- Elapsed time: NOT_RUN (fakes make it meaningless).
- Tokens: NOT_MEASURED; no estimate offered.
- Git process count and storage size of a real push: NOT_RUN. The remote here is a fake `Head`/`Contains`.
- Real GitHub / real git remote: none touched.

## Residual risk
- The counts depend on lane-20's marker lookup order; a change there moves Find counts, not Create counts.
- The 502 case counts the Create attempts the ledger made inside one flush; a production runner with its own retry policy can add calls.
