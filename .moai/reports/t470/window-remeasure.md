# t470 — window re-measurement (merged tree)

Window holder: lane-4 (session f9efb761-ced8-4706-93b7-ee3da31f7b53, pid 3441)
Absorb merge commit: e22e6caa1 ("Merge branch 'develop' into WT-queue-upgrade-proof")
Local develop tip absorbed: 79c7a0e2f
Post-absorb divergence: `git rev-list --count --left-right develop...HEAD` -> `0	15`
Working tree: `git status --porcelain | wc -l` -> 0

## Claim

The card's deliverables and its queue-storage axis still hold on the merged tree,
and the live-queue isolation control still holds there.

## Evidence (commands run in this tree, verbatim rc read without a pipe)

| # | Command | rc | Result | Log |
|---|---------|----|--------|-----|
| 1 | `go test ./internal/kanban/...` | 0 | `ok ... 139.169s` | `window-kanban.log` |
| 2 | `gofmt -l .` | 0 | 0 lines | `window-gofmt.txt` |
| 3 | `go test ./internal/cli/ -run 'TestTodoComposedUpgrade' -v` | 0 | 2 PASS / 0 FAIL, 1.399s | `window-cli-composed.log` |
| 4 | `go test ./internal/cli/ -run 'TestTodo'` | 0 | `ok ... 63.748s` | `window-cli-todo.log` |
| 5 | live-queue control (below) | 0 | before == after | `window-live-{before,after}.txt`, `window-ctrl-run.log` |

### Live-queue isolation control (re-established, not carried over)

The judgment's `live_queue_ctrl_before/after` assert that the primary checkout's live
`backlog.db` is byte-identical ACROSS this card's test run -- not that its digest holds a
fixed historical value. The historical value (`ecefa722b...`) is stale by construction:
other lanes have mutated the live queue since. The control was therefore re-established on
the merged tree rather than compared against the old digest.

- before: `13be53dffc4ba2af8ddecd0d2cdbd0ac0233d9b7ff4572b843b33948eb8a0092`, mtime 1788443858, size 372736
- run:    `go test ./internal/cli/ -run 'TestTodoComposedUpgrade'` -> rc=0
- after:  identical (`diff before after` -> rc=0, all three fields)

## Baseline-attribution

Every figure above was measured in this run, on tree `e22e6caa1`, in worktree
`.claude/worktrees/t470`. No value is carried over from the pre-absorb tree.

## Gaps (explicitly NOT observed)

- `./internal/cli/` full package was NOT run in this window (~1002s per the lead's dispatch);
  scope was restricted to the `TestTodo` selector plus the card's own two tests.
- `internal/template/agentemit` drift was NOT re-measured; it is inherited and not attributed
  to this card (zero production changes here).
- `go vet` / `golangci-lint` full-tree were NOT run in this window.
- CI verdict belongs to `origin/develop` after the lead's batch push; this lane does not push.

## Residual-risk

- Three lanes were active during the window. The live-queue control is a controlled-window
  observation of this card's own run; a foreign mutation landing between the two measurements
  would have shown as a diff, and did not -- but the window is narrow, not exclusive.
- The `TestTodo` selector covers the todo-axis surface; a regression reachable only from a
  non-`TestTodo` name in `internal/cli` would not have been seen here.
