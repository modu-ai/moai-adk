# Verdict — SPEC-GRAPH-STALE-DISPOSITION-001 (card t493)

- **Date**: 2026-09-07
- **Card**: t493 · branch `WT-graph-mxindex-edges`
- **Run-phase HEAD (this worktree)**: `f59443cd1` (full: `f59443cd13f161fa3c81c6c6562ce48fe7acf96d`)
- **Judging build**: tree-built `./bin/moai` via `make build` from this worktree — version-stamp `Commit=f59443cd1` (matches HEAD; REQ-008 / AC-008). The installed PATH binary (`e79c010b8`, 2026-09-03) predates the t478 restamp-discriminator commits (`2649fe296`, `46f6a3236`) and judged nothing in this SPEC.
- **Coordinate of the dispositioned observation**: develop worktree `.claude/worktrees/develop`, HEAD `df74b3c9d` (2026-09-07 investigation).

## 1. Verdict (REQ-007(a))

**CORRECT REPORTING — no defect.** The `moai graph check` stale verdicts observed in the develop worktree (mx-index `verdict=stale value=361`, edges `verdict=stale value=3`) are the freshness gate reporting history accurately: both artifacts were stamped 2026-08-27 at commit `d2cba5e21` and the observed drift exactly equals real source evolution since those stamps. The repair is a real regeneration, not a threshold change and not a bare re-stamp.

## 2. Evidence chain — re-established by this run phase (REQ-007(b))

The develop-worktree figures were **re-measured read-only from this worktree** (file reads + shared-repo git object access; nothing in the develop worktree was written — REQ-001). All intersection/set operations used `LC_ALL=C` collation on both sides (see §2.4 measurement-method note).

### 2.1 mx-index layer — the 361 reproduces exactly

The develop artifact `.moai/state/mx-index.json` carries provenance `commit_sha=d2cba5e21179bed08333e74888350f878c6031ba`, `tree_root=<develop worktree>`, `file_inventory=2721` entries, mtime **Aug 27 20:19** (never rescanned since).

| Command | Output |
|---|---|
| `git diff --name-only d2cba5e21..df74b3c9d \| sort -u` (LC_ALL=C) | 3346 changed files |
| `comm -12 <(inventory) <(changed)` (LC_ALL=C) | **361** — exact match to the dispatched figure |
| `comm -23 <(inventory) <(git ls-tree -r --name-only df74b3c9d \| sort)` (LC_ALL=C) | **0** — no vanished inventoried files |
| `grep -c '^\.moai/' <inventory>` | **0** — the inventory contains no `.moai/` paths at all |

The 361 is the set of inventoried files genuinely changed by merged commits since the stamp. It contains **zero** `.moai/reports` (or any `.moai/`) entries — report churn cannot contribute. At the current develop HEAD (`33fcb644b`, after the t492 merge landed post-investigation) the same measurement reads **362** / missing 0 — consistent growth from real merges.

### 2.2 edges layer — the moved sets reproduce exactly

`git log --oneline d2cba5e21..df74b3c9d -- <root> | wc -l` per `SourceFingerprintsForEdges` root (`internal/graph/meta.go`):

| Root | Commits since stamp (at `df74b3c9d`) | At current HEAD `33fcb644b` |
|---|---|---|
| `.moai/project/codemaps` | **6** | 6 |
| `.moai/specs` | **836** | 855 |
| `.moai/reports` | **780** | 796 |

All three match the dispatched investigation figures exactly at the investigation coordinate and have grown consistently since. The fourth set (`mx-index`) did not move because the untracked sidecar was never rescanned (mtime unchanged, Aug 27 20:19) — fully consistent with `verdict=stale value=3` naming exactly codemaps/specs/reports.

### 2.3 Regenerated artifact size sanity (card worktree, this phase)

This worktree's own fresh build wrote **205,162 edges** (38 MB, untracked) vs the develop artifact's 27 MB / 183,640 edges at the 8/27 stamp — the tree has grown since, again consistent with real evolution.

### 2.4 Measurement-method note

An initial intersection computation run without pinned collation returned 181/182 and briefly read as "the dispatched figure does not reproduce". Root cause: `comm` requires byte-identical collation on both inputs; `sort` defaulted to a locale collation that disagrees with the inventory's byte sort. Under `LC_ALL=C` on both sides the dispatched **361 reproduces exactly**. Readers re-running this proof MUST pin `LC_ALL=C` on every `sort`/`comm` invocation.

## 3. Card-worktree regeneration demonstration (AC-001/002/003)

This worktree carried no runtime artifacts (§1.3 of spec.md) — the transition demonstrated is **absent → fresh**, re-establishing that the gate's artifacts are produced by real regeneration and report honestly in each direction.

**AC-001 baseline** — `make build` then `./bin/moai graph check` (direct, no pipe; rc observed directly):

```
codemaps  metric=described-source-diff value=15 threshold=40 verdict=fresh
mx-index  metric=inventory-content-diff value=0 threshold=1 verdict=absent  (mx-index absent (untracked runtime artifact — fresh worktree state))
edges     metric=source-fingerprint-mismatch value=0 threshold=0 verdict=absent  (edges.jsonl absent (untracked derived artifact — fresh worktree state))
citations metric=positive-cited-path-absence value=0 threshold=0 verdict=fresh
graph check: layer mx-index verdict=absent value=0 threshold=1 — mx-index absent (untracked runtime artifact — fresh worktree state)
graph check: layer edges verdict=absent value=0 threshold=0 — edges.jsonl absent (untracked derived artifact — fresh worktree state)
---rc=1
```

**Transition table** (all invocations `./bin/moai`, HEAD `f59443cd1`):

| Layer | Threshold | Before | After `mx scan` + `graph build` | After mutant RED | After restore |
|---|---|---|---|---|---|
| mx-index | 1 | `absent value=0` | **`fresh value=0`** | `stale value=1` | `fresh value=0` |
| edges | 0 | `absent value=0` | **`fresh value=0`** | `fresh value=0` | `fresh value=0` |
| codemaps | 40 | `fresh value=15` | `fresh value=15` | `fresh value=15` | `fresh value=15` |
| citations | 0 | `fresh value=0` | `fresh value=0` | `fresh value=0` | `fresh value=0` |
| exit code | — | 1 (absent) | **0 (fresh)** | 1 (stale) | 0 (fresh) |

**AC-002 stamp check**: post-build provenance of `.moai/state/mx-index.json` and `.moai/project/graph/edges.meta.json` both carry `commit_sha=f59443cd13f161fa3c81c6c6562ce48fe7acf96d` (equal to run-phase HEAD) and `tree_root=<this worktree>`; fresh inventory = **2997** entries (vs the develop artifact's 2721 at its 8/27 stamp).

**AC-003 real build**: `./bin/moai graph build` output:

```
OK: wrote 205162 edges to <worktree>/.moai/project/graph/edges.jsonl
  mx-spec: 108
  spec-depends: 165
  code-call: 189574
  code-import: 14287
```

Edges were recomputed (205,162 written by `graph build`); no metadata was re-stamped without recomputation (REQ-003 / t478 restamp-discriminator principle respected).

## 4. Mutant RED proof (AC-004/005)

The gate still catches staleness after a fresh build. Mutant file: `internal/cli/todo.go` (present in this tree's fresh inventory; verified via `python3` key scan before mutation). Last-touching commit: `1d6a902a6` (t472).

**RED** — `git show 1d6a902a6 -- internal/cli/todo.go | git apply -R -` (applied: 2 insertions, 9 deletions reversed), then `./bin/moai graph check`:

```
codemaps  metric=described-source-diff value=15 threshold=40 verdict=fresh
mx-index  metric=inventory-content-diff value=1 threshold=1 verdict=stale  (1 inventoried file(s) changed content)
edges     metric=source-fingerprint-mismatch value=0 threshold=0 verdict=fresh
citations metric=positive-cited-path-absence value=0 threshold=0 verdict=fresh
graph check: layer mx-index verdict=stale value=1 threshold=1 — 1 inventoried file(s) changed content
---rc=1
```

**Restore** — `git checkout -- internal/cli/todo.go`; `git status --porcelain` filtered to tracked files: **0 lines**; re-check: mx-index `verdict=fresh value=0`, full-suite rc=0 (final column of the §3 table).

## 5. Disproven threshold-mismatch hypothesis (REQ-007(c))

The hypothesis that mx-index `threshold=1` mismatches the workflow is **disproven by measurement**, on two independent legs re-established this phase:

1. The drift count equals exactly the set of inventoried files changed by real merges since the stamp (361 at the observation coordinate; 0 vanished; §2.1). Threshold 1 simply says "any source change since the stamp = stale", which is the designed semantics.
2. The inventory contains **zero** `.moai/` paths (§2.1) — report churn is not and cannot be in the inventory, so the "records churn inflates the count" theory has no mechanism.
3. The card-worktree mutant (§4) shows a single changed file is caught at exactly value=1 — the threshold is neither asleep nor over-firing.

Re-raising the hypothesis requires new measurement evidence, per spec.md §1.4. The correct repair for a stale index is `moai mx scan` + `moai graph build`, never a threshold edit and never a bare meta re-stamp (the cheapest green is a false green).

## 6. Exit code vs verdict string (REQ-007(d))

`moai graph check` exits **1** when any layer is stale or absent and **0** only when all layers are fresh — `internal/cli/graph_check.go:121-131` (`res.Failed()` → `exitCodeError{code: exitStaleOrAbsent}`), source-verified in this tree. This phase's direct measurements corroborate both directions: rc=1 on baseline (absent layers, unpiped) and on the mutant (stale), rc=0 on fresh. The earlier reported "rc=0 despite two stale layers" was a pipe artifact (`$?` captured the downstream `head`). **The verdict string, not the exit code, is the judgment basis**; exit codes are recorded here only as annotated secondary observations.

## 7. Recommendation to the lead (REQ-007(e)) — lead-owned, out of this card's scope

The develop worktree's own runtime artifacts still hold the 8/27 stamps (`.moai/state/mx-index.json`, `.moai/project/graph/edges.jsonl` + `edges.meta.json`, mtime Aug 27 20:19, stamp `d2cba5e21`). The gate will keep reporting `stale 361→362+ / 3` there, correctly, until they are refreshed. **Refreshing them is a lead-window action** — this lane is forbidden from writing another session's worktree (REQ-001). Recommended, during an integration window, in the develop worktree:

```
moai mx scan
moai graph build
moai graph check   # expect mx-index fresh value=0, edges fresh value=0, rc=0
```

Real regeneration only — no bare re-stamp (SPEC-GRAPH-GATE-RESTAMP-001).

## 8. Scope and hygiene (AC-006/007)

- No file under `internal/`, `pkg/`, `cmd/`, or `gate.yaml` differs from the run-phase-start tree: the sole mutant (`internal/cli/todo.go`) was restored via `git checkout --` and post-restore `git status --porcelain` shows **zero** tracked modifications (verified in §4).
- The regenerated runtime artifacts (`bin/`, `.moai/state/mx-index.json`, `.moai/project/graph/*`) are gitignored/untracked by design and are NOT committed.
- Every cited graph measurement was produced by `./bin/moai` built from this tree via `make build` (REQ-008 / AC-008); the develop-worktree re-measurements (§2) are file reads + git object reads executed from this worktree — no external build judged any tree.

## 9. AC Matrix

| AC | Status | Basis |
|---|---|---|
| AC-001 baseline recorded | PASS | §3 baseline output (absent/absent, verbatim, this-worktree) |
| AC-002 fresh artifacts | PASS | §3 transition table + provenance SHA = HEAD |
| AC-003 real edges build | PASS | §3 `graph build` output, 205,162 edges recomputed |
| AC-004 mutant caught RED | PASS | §4 `stale value=1 threshold=1`, verbatim |
| AC-005 restore → fresh, tree clean | PASS | §4 + `git status --porcelain` tracked = 0 |
| AC-006 report completeness | PASS | REQ-007(a)-(e) in §§1, 2, 5, 6, 7; every command named with output |
| AC-007 no out-of-scope writes | PASS | §8; no production path modified |
| AC-008 tree-built binary provenance | PASS | Header + §8; all `./bin/moai`, HEAD-matching build |

## Card Cross-Check

| Milestone | Deliverable | Card |
|---|---|---|
| M1 | Disposition verdict + regeneration demo + mutant RED (this report) | **t493** |
