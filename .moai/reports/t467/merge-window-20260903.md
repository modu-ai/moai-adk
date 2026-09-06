# Card t467 — Merge Window Record (2026-09-03, lane-7)

## Claim

Card t467 (SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001, branch `WT-branchguard-flagclass` @ `eb8fc048e`, 15 commits) merged into **local develop** as merge commit `8cc1b722d`, with the merged tree byte-identical (`git rev-parse HEAD^{tree}`) to a tree on which the re-measurement batch had already been executed.

## Evidence (commands run + verbatim results, this window)

### Window acquisition

```
$ moai integration acquire --name lane-7
release-integration window acquired by aca40488-783b-40c4-ab2c-4ac7de3afbab on WT-branchguard-flagclass
```

Lead call received (lead-1 → lane-7, window order: lane-6 → lane-7). Operator had independently directed the merge; lead folded it into the queue as order 2.

### Absorption (card worktree, branch WT-branchguard-flagclass)

Pre-merge re-read: HEAD `eb8fc048e`, `git status --porcelain | wc -l` → `0`, local develop ref `ce671c160` (matches lead's dispatch).

```
$ git merge develop
Auto-merging CHANGELOG.md
CONFLICT (content): Merge conflict in CHANGELOG.md
Automatic merge failed; fix conflicts and then commit the result.
```

Conflict resolution: both-side added entries under `### Fixed` — kept both (my t467 entry first, then develop-side t463 + t468 entries), removed 3 marker lines; `grep -c '^<<<<<<<\|^=======$\|^>>>>>>>'` → `0`.

```
$ git add CHANGELOG.md && git commit -m "Merge develop into WT-branchguard-flagclass (card t467)"
[WT-branchguard-flagclass f76e2cc54] Merge develop into WT-branchguard-flagclass (card t467)
```

### Re-measurement on the absorbed tree

Tree snapshot: `git rev-parse 'HEAD^{tree}'` → `679b6c5e380a7ffbc11910b59d422e7f257a2b72`

| Check | Command | Result |
|---|---|---|
| gofmt gate (t465) | `gofmt -l . \| wc -l` | `0` |
| Build integrity | `go build ./...` | clean (no output, exit 0) |
| Scoped tests | `go test ./internal/hook/... ./internal/codexadapter/ ./internal/codexwiring/ ./internal/feedback/ ./internal/migration/migrations/ ./internal/permission/` | 16/16 packages `ok` (hook 58.5s, hook/handoff 7.3s, hook/memo 0.9s, hook/memo/taxonomy 6.7s, hook/mx 24.6s, hook/mx/complexity 3.5s, hook/perf 46.1s, hook/quality 27.2s, hook/security 15.4s, hook/testutil 3.2s, hook/trace 4.4s, codexadapter 6.0s, codexwiring 2.0s, feedback 5.5s, migration/migrations 4.9s, permission 1.5s) |
| internal/cli | `go test -timeout 25m ./internal/cli/` | `ok github.com/modu-ai/moai-adk/internal/cli 472.240s` (exit 0; background task bdn6wgjfw on the absorbed tree — same content as the merged tree; raw log: `internal-cli-remeasure-20260903.log`) |

Scope definition: my delta touches exactly one Go package (`internal/hook`: `branch_guard.go`, `branch_guard_flagclass_test.go`) plus docs (doctrine pair, SPEC artifacts, CHANGELOG); reverse-deps of `internal/hook` (measured via `go list -f '{{.ImportPath}} {{join .Imports " "}}' ./...`) are `internal/cli`, `internal/codexadapter`, `internal/codexwiring`, `internal/feedback`, `internal/migration/migrations`, `internal/permission` — all included above.

### Develop merge (integration worktree, branch develop)

Pre-merge probes: HEAD `ce671c160` (unchanged since absorption — no interleaving lane), `git rev-parse -q --verify MERGE_HEAD` → no merge in progress, working tree clean.

```
$ git merge --no-ff WT-branchguard-flagclass -m "Merge card t467 (WT-branchguard-flagclass) into develop: token-level git branch flag-class classifier — every measured mutation form denied, doctrine pair v1.3.3"
 9 files changed, 1881 insertions(+), 38 deletions(-)
```

Results: develop HEAD → `8cc1b722d`; `git rev-parse 'HEAD^{tree}'` → `679b6c5e380a7ffbc11910b59d422e7f257a2b72` **identical to the absorbed-tree snapshot** — the re-measurement batch above is therefore a measurement of this exact merged content (GitFlow §4.1 discipline 5, tree-equality route).

### Window release

```
$ moai integration release
release-integration window released (was aca40488-783b-40c4-ab2c-4ac7de3afbab on WT-branchguard-flagclass)
```

Unpushed state: `git rev-list --count origin/develop..HEAD` (develop @ `8cc1b722d`) → `151`.

### Addendum — internal/cli verdict landed (2026-09-03, post-window)

```
$ go test -timeout 25m ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	472.240s
[exited with code 0]
```

Background task `bdn6wgjfw` completed on the unchanged absorbed tree — HEAD re-read at landing: `f76e2cc54`, only untracked change `.moai/reports/t467/` (the evidence itself). Wall-time 472.240s under the 25m cap; the first attempt's 602.803s timeout (pre-absorption tree, 10m default) is superseded. The two runs are on different trees and different load conditions, so the wall-time delta is not attributed to content. Raw log exported: `internal-cli-remeasure-20260903.log`.

## Baseline-attribution

All figures above were measured in this run, in this window (2026-09-03 ~19:2x–19:4x KST), by lane-7 session `aca40488-783b-40c4-ab2c-4ac7de3afbab`. Trees are named by SHA (`679b6c5e…` absorbed/merged content; `f76e2cc54` absorption commit; `8cc1b722d` develop merge commit). No figure is carried from another package, tree, or point in time.

## Gaps

- ~~`internal/cli` package verdict: not observed at evidence-write time~~ → landed same day (Addendum: `ok`, 472.240s, exit 0); kept struck-through for the record.
- First cli attempt (pre-absorption tree) hit go test's built-in 10m timeout — `FAIL … 602.803s`, goroutine dump at `TestRunDoctor_FixWithFailures` (target_coverage_test.go:1059). Cause undetermined (machine was under multi-lane test load; hang vs slow not separated). Superseded by the `-timeout 25m` rerun (`ok`, 472.240s — Addendum).
- Full-suite CI verdict: not run locally by policy; arrives with the lead's batched develop push.
- Pre-absorption early-signal runs (hook 63.9s pass; 5 light reverse-dep packages pass) are superseded by the absorbed-tree runs above — retained in session record only.

## Residual-risk

- The `internal/cli` PASS above measures the absorbed/merged tree (`679b6c5e…`) as of this window; develop has since moved (17 commits ahead at addendum time), so current-tip health remains CI's full-suite verdict.
- gofmt gate verified `0` at merge time only; later windows' merges can reintroduce drift (each window re-checks per the lead's procedure).
- If develop moved between absorption and merge, tree-equality would have failed; it did not (HEAD verified `ce671c160` immediately pre-merge), so no such risk materialized for this merge.
