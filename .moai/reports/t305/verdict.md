# t305 — statusline warm path: serial git spawns 7 → 2

Card: t305 · Branch: `WT-statusline-git-spawn` · Base: local `develop` 2660bcd09

## Claim

One `moai statusline` render spawned **7** git child processes, serially. It now spawns **2**.
The rendered line is byte-identical, and so are `moai todo pr --help` and `moai todo done --help`.

The card, following t215, expected 5. The measurement found 7 — t215's profile covered the
builder+collector path only, and never saw the other two.

## Evidence

Spawns are counted with a PATH shim that logs every `git` invocation
(`.moai/reports/t305/shim/git`). The count is exact and load-independent, which is what makes it
evidence rather than a timing impression. Both binaries were run against the same tree, in the
same window, by the lane independently of the implementing agent:

```
--- base (7) ---                                     --- after (2) ---
1 -C <wt> rev-parse --path-format=absolute \
          --git-dir --git-common-dir
2 -C <wt> rev-parse --path-format=absolute \
          --git-dir --git-common-dir
3 rev-parse --git-dir                                1 rev-parse --git-dir --show-toplevel
4 rev-parse --show-toplevel                          2 status --porcelain --branch
5 symbolic-ref --short HEAD
6 status --porcelain
7 rev-list --count --left-right @{upstream}...HEAD

diff render-base render-after  →  RENDER IDENTICAL
```

`lane-verify-spawns-base.txt`, `lane-verify-spawns-after.txt`, `lane-verify-render-after.txt`.

### The two spawns t215 never saw

Spawns 1 and 2 fired at **cobra command-tree construction time**. `newTodoPRCmd()` and
`newTodoDoneCmd()` each resolved the landed ref eagerly, to interpolate it into help text. The
tree is built at process start, so every `moai <anything>` invocation — a statusline render
included, once per render — paid two `git rev-parse` spawns to build help for a command it was
not running. The resolution also took the ADOPTING entry point
(`kanban.ResolveTodoQueueRootAdopting`), which on the home-fallback branch performs filesystem
adoption; a pure render path had a write reachable from it.

Resolution is now lazy and memoized. Measured: `moai todo pr --help` costs 1 spawn (was 2) and
still prints the resolved `origin/main` in both the body and the `--require-landed` flag
description — `lane-verify-help-spawns-after.txt`, and both help surfaces diff clean against the
baseline binary.

### The remaining three

`NewRepository` asked `rev-parse --git-dir` purely to validate and threw the answer away, then
asked `--show-toplevel`. One spawn now asks both; the cold path spawns a second time only to keep
`ErrNotRepository` and `get repository root` distinguishable, which costs nothing on the warm path.

`Status()` now passes `--branch`, so the `## ` header carries the branch, the upstream, and the
divergence — retiring both the `symbolic-ref` and the `rev-list` spawn. The statusline collector
takes the branch from that header and falls back to `CurrentBranch()` only when it is empty
(detached HEAD, or a failed status), so `CurrentBranch()` and its `ErrDetachedHEAD` contract are
untouched.

The header-skip hazard is real and covered: `#` is neither `' '` nor `'?'`, so an unskipped
`## ` line would have been counted as a staged file. `TestStatusSkipsBranchHeaderEntry` fails
without the skip. Seven header shapes are driven against real git in `t.TempDir()` fixture
repositories rather than assumed — `TestStatusBranchHeaderShapes`.

## Baseline-attribution

Measured by the lane in this run, in this worktree, at `2660bcd09` + working changes:

    go build ./...                                        → rc 0
    go vet ./internal/core/git/... ./internal/statusline/... ./internal/cli/...  → rc 0
    go test ./internal/core/git/ -count=1                 → ok  67.997s
    go test ./internal/statusline/ -count=1               → ok  11.668s
    go test ./internal/cli/ -count=1                      → ok 443.744s
    go test ./internal/kanban/ -count=1                   → ok 137.774s

Consumer scope, measured rather than assumed — every package that reads `Repository.Status` or
constructs a `GitStatus`:

    grep -rln '\.Status()\|GitStatus{' --include='*.go' internal/ pkg/ cmd/ | grep -v _test.go
      → internal/cli, internal/core/git, internal/statusline

All three are in the suite above, so "other packages were not run" is not an open gap here.

Latency, interleaved arms on one machine state, measured twice by different actors at different
loads:

    implementing agent (load 6.68):  base 323.23ms → after 151.72ms median  (−53.1%)
    lane, independent  (load 12.88): base 334.80ms → after 153.60ms median  (−54.1%)

The absolute figures are load-bound and are the weaker evidence; the ratio holding across a 2x
load difference is what makes them worth reporting at all. The spawn count is the strong evidence.

## Gaps

- **CI has not run.** No push, per lane discipline. The full suite and the darwin/windows matrix
  are unobserved; the develop push is where that verdict comes from.
- **Ahead/behind end-to-end through the statusline is unobserved.** This worktree's branch has no
  upstream, so the render exercised 0/0 only. Covered one layer down at `Status()` by
  `TestStatusAheadBehindFromHeader` against real fixture repositories (in-sync 0/0, diverged 1/1).
- **Only git 2.50.1 (Apple Git-155) was exercised.**
- The non-`--branch` `status --porcelain` output shape is no longer produced by this code path, so
  any consumer that parsed `Status()`'s raw output rather than its struct would be affected. None
  exists — `Status()` returns a struct and its raw output never escapes the method.

## Residual-risk

- **Older git emits `## Initial commit on <branch>` where 2.50.1 emits `## No commits yet on
  <branch>`.** The parser strips only the modern prefix, so on an older git a fresh repository's
  header contains a space and yields an empty branch — which falls back to `CurrentBranch()` and
  produces the right name at the cost of one extra spawn. It degrades to the old behaviour rather
  than to a wrong answer, which is why it is left unhandled rather than guessed at.
- The lazy landed-ref resolution memoizes per process. A long-lived process that changed
  repositories mid-life would keep the first answer; no such process exists (`moai` is
  process-per-invocation), and the previous code had the same property one construction earlier.
- `TestResolveBacklogCounts_LatencyBudget` failed once during the implementing agent's
  three-package concurrent run (p95 98ms against a 25ms ceiling) and passed on every isolated
  re-run, with and without these changes. It measures SQLite-vs-JSON backlog reads and touches no
  git path. Recorded as a load-flake, not attributed to this change.
