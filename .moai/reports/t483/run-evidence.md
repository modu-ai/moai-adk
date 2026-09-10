# t483 Run-phase Evidence — deferred edges refresh removal (t448 Option-1)

Commit: `7a663ea5f` (branch `WT-t448-residue`, on baseline `a825183dd`)

## Claim

The SessionStart deferred edges-refresh path is fully removed and zero references remain; the preserved consumer-pays path (`edgesRefreshNeeded` / `refreshEdgesArtifact`, the `graph.go` query-path consumer, `internal/graph` entirely) is untouched; all four verifications pass on the touched packages.

## Evidence

Anchor gate (start of run):

```
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t483
$ git branch --show-current
WT-t448-residue
```

POST-REMOVAL SWEEP — expect 0 hits:

```
$ grep -rn "DeferredEdgesRefresh\|deferredEdgesRefresh\|runDeferredEdgesRefresh\|edgesStale" internal/ cmd/ --include="*.go"
(no output)
grep exit: 1
```

1. gofmt:

```
$ gofmt -l internal/hook internal/cli
(no output)
gofmt exit: 0
```

2. go vet:

```
$ go vet ./internal/hook/... ./internal/cli/...
VET_OK exit=0
```

3. Tests:

```
$ go test ./internal/hook/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/hook	40.315s

$ go test ./internal/cli/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	552.371s
```

4. Build:

```
$ go build ./...
BUILD_OK exit=0
```

Commit:

```
$ git log -1 --format="%h %s"
7a663ea5f refactor(t483): remove deferred edges refresh from hook path — consumers pay at query time (t448 Option-1)
6 files changed, 31 insertions(+), 449 deletions(-)
 delete mode 100644 internal/hook/session_start_deferred_edges_test.go
```

Files changed:

- `internal/hook/session_start.go` — removed `DeferredEdgesRefresh` type, `deferredEdgesRefresh` field, `WithDeferredEdgesRefresh` option, the `edgesStale` probe block in `Handle`, both dispatch sites, `runDeferredEdgesRefresh` (+ its `@MX:NOTE`), the `edgesStale` param of `spawnDeferredAdvisoryScans`, and the now-unused `internal/graph` import
- `internal/cli/deps.go` — removed the `WithDeferredEdgesRefresh(deferredEdgesRefresh)` wiring + comment
- `internal/cli/graph_refresh_cli.go` — removed the `deferredEdgesRefresh` wrapper (+ now-unused `os` import); updated the `refreshEdgesArtifact` doc comment (deferred-path fail-safe shape no longer exists)
- `internal/cli/graph_deferred_refresh_test.go` — deleted the two deferred-wiring tests (fresh-no-rewrite: gated skip no longer exists; budget-overrun-warns: warning lived in the removed wrapper); the stale-rebuild test repointed to `refreshEdgesArtifact` with identical assertions (artifact written, staleness predicate false, nothing staged), renamed `TestRefreshEdgesArtifact_StaleRefreshesAndStagesNothing`
- `internal/cli/graph_shrink_test.go` — `TestShrinkGuard_DeferredPathInheritsRefusal` repointed to `refreshEdgesArtifact` with identical assertions, renamed `TestShrinkGuard_RefreshEdgesArtifactInheritsRefusal`
- `internal/hook/session_start_deferred_edges_test.go` — deleted (exercised only the removed path)

## Baseline-attribution

All commands above were run in this session, in this worktree (`.claude/worktrees/t483`, branch `WT-t448-residue`), against HEAD `a825183dd` before the commit and against `7a663ea5f`'s identical tree after it (the commit staged exactly the six files above; `git status --short` was re-read immediately before staging and showed only those files plus the untracked reports dir).

Post-commit re-measurement of the preserved symbol (run AFTER the commit, on the committed tree):

```
$ grep -rn "EdgesSourcesMoved" internal/ cmd/ --include="*.go"
internal/graph/check_test.go, internal/graph/fanin_edge.go, internal/graph/meta.go,
internal/cli/graph_refresh_test.go, internal/cli/graph_refresh_cli.go
```

(full verbatim output: 23 lines — `EdgesSourcesMoved`/`EdgesSourcesMovedFor` definitions and tests in `internal/graph/` (`meta.go:154`, `meta.go:163`, `fanin_edge.go:93`, `check_test.go`), the query-path consumer `edgesRefreshNeeded` at `internal/cli/graph_refresh_cli.go:44`, and its tests in `internal/cli/graph_refresh_test.go`. Zero hits in `internal/hook/`.)

## Gaps

- Did NOT run the full `go test ./...` suite locally (CLAUDE.local.md §6 prohibition — full-suite verdict is CI's). Only `./internal/hook/` and `./internal/cli/` were exercised; other packages depending on `internal/hook`/`internal/cli` signatures were covered by `go build ./...` (compilation) but not by tests.
- Did NOT run `go vet`/tests with `-race`.
- Did NOT verify behavior of a live SessionStart hook end-to-end (no runtime invocation of the hook binary; verification is static + package tests).
- Did NOT remove the no-arg `graph.EdgesSourcesMoved()`'s now-zero caller count — declared a known follow-up, out of scope.
- Did NOT check non-Go references to the removed symbols (docs, templates, scripts) — the mission's sweep scope was `internal/ cmd/ --include="*.go"`.

## Residual-risk

- Any consumer outside `internal/`+`cmd/` Go code (e.g. a template hook script calling a removed CLI surface) would not be caught by this sweep — none was known to exist; the removed symbols were package-private or DI-only.
- The `TestDeferredEdgesRefresh_FreshNoRewrite` property (fresh tree → zero writes) is now untested anywhere: its subject was the deferred gating itself, which is removed. The query path's freshness gating (`edgesRefreshNeeded` in `graph.go`) retains its own tests in `graph_refresh_test.go`.
- If another branch concurrently adds a caller of the removed symbols, the compile fails there — no cross-branch guard exists in this worktree.
