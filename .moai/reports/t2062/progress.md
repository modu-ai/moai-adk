# t2062 — progress (Class A/B, codemaps freshness; leader ruling (c): described-source-diff only)

Card: `[GRAPH] codemaps described-source-diff >= 40`. Batch PR #1969 was blocked by the graph-freshness job (codemaps value=115, threshold=40). Leader ruling (c): t2062 alone, described-source-diff only; t2066 (fold-guard leak) and t1693 stay untouched and t2066 follows after t2062 lands; a fold leak caused by the regeneration means stop and report. No `moai factory complete`, no push (80bd), no Codex approval.

## Cause (measured)
`moai graph check` on a tree built from main `cf4ad8510d` (binary stamped `v3.2.1-t2062 cf4ad8510`, built with `go build` into scratch; the installed f9d886695 is not used as evidence) reported `codemaps metric=described-source-diff value=115 threshold=40 verdict=stale`, `measured from: 67b4734d6 (working-tree-differs-from-stamp)`. The metric is the count of described-worthy Go files (non-test, non-testdata, under internal/cmd/pkg) that differ between the stamp commit and the working tree. `git diff --name-status 67b4734d6b HEAD` filtered the same way gave exactly 115 files: 18 added, 97 modified (the stamp `provenance.json` named t2061's PR-head `67b4734d6b` and main moved on).

## Decisions
| # | Decision | Basis |
|---|---|---|
| 1 | Incremental refresh of three documents (`modules.md`, `data-flow.md`, `entry-points.md`) through the codemaps skill's phases, NOT `--force` full regeneration. | `fold-judgments.txt` holds 82 judgments (53 fold, 29 omission); a full rewrite of ~500 KB is what produced the fold leak t2066 repairs, and the delta is 115 files in 36 package dirs with one new package. Skill step "Existing codemaps (if --force not set) for incremental updates" supports this. I first invoked `--force`, then chose incremental after reading the delta; the choice is mine, flagged in the report to the leader. |
| 2 | Facts came from one read-only Explore agent over the 115-file list (Phase 1); where a sentence rested on its inference I re-read the source (`verify_clear_rejection.go` Short text, `ClearRejectionOnApproval`, `approvals.jsonl`, `jev_cost.go`, `--base-sha`). One sentence was reworded after that check. | Explore report plus source greps |
| 3 | Four sentences that this window made false were corrected: ledger "depends on core only" (now `core` and `sqliteuri`); Jev ledger "no read path" in `modules.md` (two places), `data-flow.md`, `entry-points.md` (now `moai jev cost` reads it). | grep of the old wording, edits |
| 4 | Stamp at `git merge-base HEAD main` = `cf4ad8510d8ba9aabe6a804cadf043611390b31b`, never a branch-local HEAD (squash-merge orphaning, SPEC-STAMP-REACHABILITY-001). Local main is this repository's landing surface. | `moai graph stamp codemaps --help` |

## Evidence (this run, tree `b101ede4c9` after the stamp commit)
- Before: `.moai/reports/t2062/graph-check-before.txt` — `codemaps ... value=115 threshold=40 verdict=stale`.
- After: `.moai/reports/t2062/graph-check-after.txt` — `codemaps  metric=described-source-diff value=0 threshold=40 verdict=fresh`; `citations  metric=positive-cited-path-absence value=0 threshold=0 verdict=fresh`. `mx-index` and `edges` read `absent` (untracked runtime artifacts, fresh worktree state); the command exits 1 for that reason, not for codemaps.
- Phase 4 fold and omission check (the skill's awk over fold-judgments.txt and the six top-level documents): `COLLECTED: fold=53 omission=29`, no `UNCOVERED`, no `FOLD-PROSE`, no `GAP`.
- Fold guard tests, test files untouched, run singly: baseline on cf4ad8510d before any edit `TestCodemapsFoldPreservationGuard` ok 1.704s and `TestCodemapsFoldGuardFixtures` ok 1.366s; after the doc edits `--- PASS: TestCodemapsFoldPreservationGuard (0.02s)` and `--- PASS: TestCodemapsFoldGuardFixtures (1.29s)`.
- Package count: `go list ./internal/... ./cmd/... ./pkg/...` = 199 (198 before + `internal/sqliteuri`).
- Commits: docs refresh (three documents), then the stamp (`provenance.json` only, `commit_sha` cf4ad8510d8b...).

## Baseline-attribution
Tree `WT-codemaps-described-sourc-2`, base main `cf4ad8510d`; binary built from that tree by `go build -ldflags "-X .../pkg/version.Commit=cf4ad8510d ..."` (not `make build`, which can rewrite tracked catalog hashes); judging build commit equals the tree's base, not a strict ancestor of HEAD.

## Gaps
- Accuracy of the doc text rests on one Explore agent's reading plus my spot-checks of six facts; I did not re-verify every sentence against source (freshness and accuracy are separate axes — the gate is the stamp, not an accuracy proof). The `citations` row proves only that no positively-cited path is absent.
- I did not touch lines the Explore report marked as stale-but-outside-this-window: package-count statements in `dependencies.md`/`overview.md` (176/older) were already stale before this window and the metric does not read them.
- The t1693 extra item (removing a positive citation of the deleted `mcp_factory_msg.go`) is not in this card; the citations row is fresh with value 0 anyway.
- `mx-index` and `edges` layers absent in this tree (untracked artifacts); CI computes them in its own tree.
- Tool refusals: the worktree guard refused (a) a compound build-and-version command, (b) a `for` loop with a computed `sed` line number, (c) `go test -run 'A|B'` with an alternation; each was re-run as plain separate commands. Codex `codex_review` was not run (no approval from the leader); card-review is a self-review.
- `go test ./internal/graph/` as a whole package was not run; only the two named guard tests.
- Windows not exercised.

## Residual-risk
- The stamp makes the freshness metric read 0 while documents may still carry descriptions older than this window; the 36 package directories in the delta were covered only where the Explore report named a docs-impact.
- If main moves again before the batch PR lands, `described-source-diff` rises again by the changed described-worthy files (t2066 and others).
