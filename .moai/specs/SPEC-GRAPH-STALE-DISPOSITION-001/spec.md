---
id: SPEC-GRAPH-STALE-DISPOSITION-001
title: "Graph freshness gate stale-verdict disposition — correct-reporting verdict, regeneration demonstration, and mutant RED proof"
version: "0.1.0"
status: completed
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P2
phase: "v3.1.4"
module: "internal/cli/graph"
lifecycle: spec-lite
tags: "graph, mx-index, edges, freshness, disposition, t493"
tier: S
related_specs: [SPEC-GRAPH-GATE-RESTAMP-001, SPEC-V3R6-GRAPH-FRESHNESS-001]
---

# SPEC-GRAPH-STALE-DISPOSITION-001

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-09-07 | 0.1.0 | Initial plan-phase authoring (Tier S, 2 artifacts + progress.md). Formalizes the t493 lane's completed disposition investigation into a SPEC whose run phase writes a tracked verdict report and executes a regeneration demonstration + mutant RED check in the card worktree only. |

## 1. Background and Disposition (investigation record)

### 1.1 The question

`moai graph check` was observed (develop worktree `.claude/worktrees/develop`, HEAD `df74b3c9d`, 2026-09-07) reporting:

- **mx-index layer**: `verdict=stale value=361` — metric `inventory-content-diff`, reason "361 inventoried file(s) changed content", threshold 1.
- **edges layer**: `verdict=stale value=3` — metric `source-fingerprint-mismatch`, reason "source set(s) moved: codemaps, reports, specs".

The question posed to card t493: is this a defect in the gate, or correct reporting?

### 1.2 The verdict — CORRECT REPORTING, no defect

The lane verified the following against the develop worktree this session (2026-09-07). These are investigation evidence; the run phase re-establishes its own baselines and MUST NOT treat these figures as its own measurements.

1. **mx-index artifact**: `.moai/state/mx-index.json` (UNTRACKED runtime artifact) is stamped with provenance `tree_root` = the develop worktree, `commit_sha=d2cba5e21179bed08333e74888350f878c6031ba`, scanned 2026-08-27 20:19 KST, `file_inventory=2721` entries. Since that stamp, develop absorbed dozens of card merges (HEAD now `df74b3c9d`). Verification: the intersection of the inventory keys with `git diff --name-only d2cba5e21..HEAD` is EXACTLY 361 — the drift count equals the set of inventoried files actually changed by merged commits since the stamp. `missing=0` (no vanished files). The 361 grows from real source evolution, NOT from `.moai/reports` churn — the inventory roots are `internal/cmd/pkg` (plus legacy docs-site entries) and contain no `.moai/reports` paths. Threshold 1 ("index older than any source change = RED") is a correct signal.
2. **edges artifact**: `.moai/project/graph/edges.jsonl` (UNTRACKED, 27 MB, 183,640 edges) + `edges.meta.json`, stamped the same moment (2026-08-27T11:19:55Z, commit `d2cba5e21`). Its four source sets are defined in `internal/graph/meta.go` `SourceFingerprintsForEdges`: `codemaps=.moai/project/codemaps/`, `mx-index=.moai/state/mx-index.json`, `specs=.moai/specs/`, `reports=.moai/reports/`. Since the stamp, codemaps changed in 6 commits, specs in 836 commits, reports in 780 commits (`git log d2cba5e21..HEAD` per path) — all three genuinely moved. The mx-index set did NOT move because the untracked sidecar was never rescanned (mtime 8/27 unchanged) — fully consistent.
3. **Exit code note**: `moai graph check` exits 1 when any layer is stale/absent (`internal/cli/graph_check.go:121-131`, `Failed()`/`exitStaleOrAbsent`). An earlier reported "rc=0 despite two stale layers" was a pipe artifact (`$?` captured the downstream `pager`/`head`). The verdict string, not the exit code, is the judgment basis.

### 1.3 Card worktree pre-state (measured 2026-09-07, this worktree)

This card worktree (branch `WT-graph-mxindex-edges`, HEAD `df74b3c9d`) carries **no** `.moai/state/mx-index.json` and **no** `.moai/project/graph/edges.meta.json` — the runtime artifacts exist only in the develop worktree. Consequently the run phase's regeneration here is an **absent→fresh** transition, not the develop worktree's stale→fresh; the mutant RED check (REQ-005 / AC-004) is the demonstration that the gate still catches staleness. `.moai/reports/t493/` also does not exist yet and is created by the run phase.

### 1.4 Disproven hypothesis — threshold mismatch (do not re-raise without new evidence)

The hypothesis that the mx-index threshold of 1 represents a "threshold mismatch with the workflow" was **DISPROVEN** by measurement (§1.2 item 1): the drift count exactly equals the set of genuinely changed inventoried files, the threshold semantics ("any source change since the stamp = stale") behave as designed, and the correct repair is a real regeneration (`moai mx scan` + `moai graph build`), never a threshold change and never a bare meta re-stamp (the t478 restamp-discriminator principle: the cheapest green path is a false green).

## 2. Requirements (GEARS)

### REQ-001 — Isolation (Ubiquitous)

The run phase shall execute every mutation, regeneration, and verification **inside this card worktree only**, and shall not write to any other worktree — the develop worktree included. Refreshing the develop worktree's artifacts is the lead's integration-window concern and shall be carried in the disposition report as a recommendation to the lead, never executed by this SPEC's run phase.

### REQ-002 — No production changes (Ubiquitous)

The run phase shall not modify any production source file, any gate configuration (`gate.yaml`), or any threshold constant; its only tracked write surface shall be the disposition report under `.moai/reports/t493/` plus the SPEC artifacts themselves.

### REQ-003 — Real regeneration (Event-driven)

**When** the run phase regenerates the graph runtime artifacts, it shall do so by full real regeneration (`moai mx scan` then `moai graph build` — recomputing the mx index inventory and all graph edges before stamping) and shall never re-stamp artifact metadata without recomputation.

### REQ-004 — Fresh transition (Event-driven)

**When** the regeneration of REQ-003 completes, `moai graph check` shall report the mx-index layer `verdict=fresh value=0` and the edges layer `verdict=fresh value=0`, with the codemaps and citations layers remaining `fresh`.

### REQ-005 — Mutant RED proof (Event-driven)

**When** one inventoried file is mutated by reverse-applying a real committed change after the fresh transition, a re-run of `moai graph check` shall report the mx-index layer `verdict=stale` with `value>=1` — proving the gate still catches staleness after the artifacts were freshly built.

### REQ-006 — Mutant restoration and clean tree (Event-driven)

**When** the mutated file is restored to its committed state and `moai graph check` is re-run, the mx-index layer shall report `verdict=fresh value=0` again, and `git status --porcelain` on tracked files shall show no modifications other than the SPEC artifacts and the disposition report (all mutant changes reverted; the regenerated runtime artifacts are UNTRACKED and are not committed).

### REQ-007 — Disposition report (Ubiquitous)

The run phase shall write a tracked disposition report at `.moai/reports/t493/verdict.md` recording: (a) the verdict "correct reporting — no defect"; (b) the evidence chain of §1.2 re-established against the run-phase's own measurements; (c) the disproven threshold-mismatch hypothesis with its disproving measurement; (d) the exit-code-vs-verdict-string clarification; and (e) a recommendation to the lead to refresh the develop worktree's `.moai/state/mx-index.json` and `.moai/project/graph/edges.*` artifacts during the integration window (lead-owned, out of this card's scope).

### REQ-008 — Judgment basis and build provenance (Ubiquitous)

The run phase shall base every layer verdict claim in the report on the verdict string of the graph check output, and shall record exit codes only as annotated secondary observations with the pipe-artifact caveat of §1.2 item 3. Every cited measurement shall additionally be produced by a binary built from this card tree (`make build`, then invoked by path `./bin/moai ...`) — never by the installed `moai` from PATH — and the evidence shall name the invocation path used, per verification-claim-integrity.md §2.2 (tool-provenance attribution): the installed build (v3.2.0-rc.0, build `e79c010b8`, 2026-09-03) predates the two commits that landed the t478 restamp-discriminator predicate (`2649fe296`, `46f6a3236` — `internal/graph/check.go` + `internal/cli/graph_check.go`), so a PATH-invoked build cannot judge this tree's freshness behavior.

## 3. Acceptance Criteria (Given-When-Then)

### AC-001 — Baseline recorded

**Given** the card worktree at its run-phase HEAD **when** the run phase runs `moai graph check` before any regeneration and records the output in the disposition report **then** the report states the observed per-layer baseline (expected on this worktree: mx-index and edges layers `absent` per §1.3; if any layer unexpectedly reads `stale`, its value and reason are recorded verbatim) and explicitly labels it a this-worktree measurement, distinct from the develop-worktree investigation figures.

### AC-002 — Regeneration produces fresh artifacts

**Given** the baseline of AC-001 **when** `moai mx scan` and `moai graph build` complete **then** `.moai/state/mx-index.json` and `.moai/project/graph/edges.meta.json` exist in the card worktree, their stamped `commit_sha` equals the run-phase HEAD, and a `moai graph check` run reports mx-index `verdict=fresh value=0`, edges `verdict=fresh value=0`, codemaps `fresh`, citations `fresh`.

### AC-003 — Edges regeneration is real

**Given** the freshly built artifacts of AC-002 **when** the report documents the build **then** the report cites the regenerated edge count (observed in the build output) and confirms it was recomputed by `moai graph build`, not carried over — the report shall contain no bare re-stamp of metadata.

### AC-004 — Mutant is caught RED

**Given** the fresh state of AC-002 **when** one inventoried file is mutated by reverse-applying a real committed change (`git show <sha> -- <file> | git apply -R -`), and `moai graph check` is re-run **then** the mx-index layer reports `verdict=stale` with `value>=1`, and the observed output is recorded verbatim in the report.

### AC-005 — Restoration returns to fresh, tree clean

**Given** the RED state of AC-004 **when** the mutated file is restored (`git checkout -- <file>`) and `moai graph check` is re-run **then** the mx-index layer reports `verdict=fresh value=0`, and `git status --porcelain` shows no tracked-file modifications beyond the SPEC artifacts and `.moai/reports/t493/` (all mutants reverted; regenerated runtime artifacts remain UNTRACKED and uncommitted).

### AC-006 — Report completeness

**Given** the completed run phase **when** the disposition report is read **then** it contains all five elements of REQ-007, names every command it ran with its observed output, and carries a Card Cross-Check reference to card t493.

### AC-007 — No out-of-scope writes

**Given** the completed run phase **when** the branch diff and untracked inventory are inspected **then** no file under `internal/`, `pkg/`, `cmd/`, or `gate.yaml` differs from the run-phase-start tree, and no artifact of any other worktree was touched.

### AC-008 — Tree-built binary provenance

**Given** the installed `moai` binary predates the t478 restamp-discriminator commits `2649fe296`/`46f6a3236` **when** the run phase executes any graph operation cited as evidence (baseline check, `mx scan`, `graph build`, every re-check, the mutant RED and restore checks) **then** each invocation used the card-tree-built binary at `./bin/moai` after a `make build` from this tree, and every evidence row in the disposition report names the invocation path (not a bare `moai` from PATH); no cited measurement may carry the installed build's provenance.

## 4. Out of Scope

The following are explicitly excluded from this SPEC:

### Out of Scope — Gate design and threshold changes

- Any modification to `gate.yaml`, any freshness threshold constant, or any gate verdict logic. The threshold-mismatch hypothesis was disproven (§1.4); re-raising it requires new measurement evidence, not this SPEC.
- Any change to `internal/cli/graph_check.go`, `internal/graph/meta.go`, or any other production source.

### Out of Scope — Other worktrees

- Refreshing the develop worktree's `.moai/state/mx-index.json` or `.moai/project/graph/edges*` artifacts — this is the lead's integration-window concern and shall be carried only as a recommendation in the disposition report (REQ-007(e)).

### Out of Scope — Rescanning cadence and CI wiring

- Any change to when or how often the graph artifacts are rebuilt (the freshness-cadence domain is owned by SPEC-GRAPH-FRESHNESS-CADENCE-001 / SPEC-V3R6-GRAPH-FRESHNESS-001/002).
- Committing the regenerated runtime artifacts — `.moai/state/mx-index.json` and `.moai/project/graph/edges*` are UNTRACKED by design and remain so.

## 5. References

- `internal/cli/graph_check.go:121-131` — stale/absent exit semantics (`Failed()` / `exitStaleOrAbsent`).
- `internal/graph/meta.go` `SourceFingerprintsForEdges` — the four edges source sets and their roots.
- `.claude/rules/moai/core/verification-claim-integrity.md` §2.2 — the tool-provenance rule grounding the tree-built-binary requirement (REQ-008 / AC-008); installed build `e79c010b8` (2026-09-03) predates `2649fe296` + `46f6a3236`.
- SPEC-GRAPH-GATE-RESTAMP-001 — the restamp-vs-regeneration discriminator principle (t478 lineage).
- SPEC-V3R6-GRAPH-FRESHNESS-001/002, SPEC-GRAPH-FRESHNESS-CADENCE-001 — freshness-cadence domain owners.
- Card t493 dispatch evidence — the lane's 2026-09-07 measurements recorded in §1.2.
