---
id: SPEC-V3R6-GRAPH-FRESHNESS-001
title: "Graph layer freshness: per-layer drift gate, query-time refresh, content-addressed citations, AST symbol layer, MCP code queries"
version: "1.1.0"
status: draft
created: 2026-08-25
updated: 2026-08-25
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/graph"
lifecycle: spec-anchored
tags: "graph, freshness, codemaps, mx-index, astx, mcp, citation"
era: V3R6
tier: L
---

## §A. Problem Statement

moai-adk's graph surfaces (codemaps, @MX index, edges.jsonl) describe a codebase that moves continuously, and nothing measures or signals the gap between the description and the described. Measured in this tree (WT-graph-freshness @ baa100ce5, 2026-08-25):

- `.moai/project/codemaps/` was last regenerated at commit `6da952899` (2026-08-12). Since then, origin/main accumulated **740 commits** with zero notification to any consumer. Measurement provenance: the graft analysis measured 713 at report time on `294b4b6ab` — that figure is no longer reproducible (738 measured at review time); this SPEC's 740 was measured directly at authoring in this tree and is the figure the SPEC stands on.
- `.moai/state/mx-index.json` (primary checkout, 357 KB, 2026-08-20) carries only `schema_version`, `tags`, `scanned_at` — **no commit SHA provenance**. A reader cannot tell which tree or commit the index describes.
- `.moai/state/mx-index.json` and `.moai/project/graph/edges.jsonl` do not exist in a fresh worktree at all (untracked runtime artifacts). Absence is indistinguishable from freshness today.
- Citations in codemaps and reports anchor by line number. The same symbol carried different line numbers across trees (`init.go:773` vs `init.go:898`), producing 2 stale-misjudgment incidents this round.
- The tree-sitter substrate for a code-derived symbol layer already exists (`smacker/go-tree-sitter` in go.mod; 16 `.scm` queries in `internal/navigator/astx/queries/`) but is consumed only by the navigator, and the queries capture declarations only — no call or import edges.
- The 21-tool MCP server exposes SPEC/verification/goal/audit queries but no code-graph query.

The graft analysis (2026-08-24, `.moai/reports/graft/graft-analysis-20260824.md`) concluded: adopt 4 Graft design principles (freshness gate, query-time refresh, content-addressed citations, code symbol layer + code queries) — NOT the tool itself (TS/npm runtime conflicts with the single Go binary and 16-language neutrality).

This round's observed defect family raises the stakes: a cache or index anchored to the wrong tree answers about a different tree (lead worktree 259 commits stale, installed binary 259 commits stale contaminating audits — CR #8/t246 lineage). Freshness machinery that itself anchors to the wrong tree would multiply this defect class instead of closing it.

## §B. Scope

### §B.1 In Scope

Five milestones (execution order M1 → M5; decision-review order in plan.md §F):

- **M1 drift gate** — new `moai graph check` subcommand answering per-layer staleness numerically, exit-code discipline, wiring into `moai gate` and CI.
- **M2 query-time refresh** — `moai graph query` and `moai mx query` refresh their mechanical input layers (changed files only, content-hash cache, no LLM calls, uncommitted edits reflected) before answering; update-cost budget with overrun warning.
- **M3 citation convention switch** — excerpt + content hash becomes the canonical citation form in codemaps, the mx index, and generated reports; line numbers demoted to convenience notation; measured-tree SHA co-stamped.
- **M4 symbol layer** — the astx extractor becomes consumable outside the navigator; code-call and code-import edges are added to edges.jsonl additively; per-language resolution-grade matrix published.
- **M5 MCP code queries** — `graph_find_code`, `graph_trace_calls`, `graph_file_api` (signatures only), depending on M4, file_api first.

### §B.2 Boundary decisions

- M2 refresh applies to the **mechanical layers** (mx-index, edges.jsonl). The curated layer (codemaps) is never auto-rewritten by query paths; its staleness is surfaced by the M1 gate instead. Rationale: regenerating curated prose without an LLM would produce garbage; regenerating a scanner index or a deterministic aggregation is sound.
- Layer set: exactly three layers are gated — codemaps, mx-index, edges.jsonl. The navigator's nav-graph.json keeps its own provenance scheme (`extract_commit_sha` per nav-tokens.md) and is out of this SPEC's gate.

## §C. History

| Date | Event |
|---|---|
| 2026-08-12 | codemaps last regenerated (`6da952899`) |
| 2026-08-20 | mx-index last scanned (primary checkout) |
| 2026-08-24 | Graft analysis authored at `294b4b6ab` (713-commit drift measured); operator decision: adopt 4 design principles, do not install Graft |
| 2026-08-25 | Card t250 picked; this SPEC authored in WT-graph-freshness @ `baa100ce5` (740-commit drift re-measured; anchors re-verified, see research.md) |

## §D. Requirements (GEARS notation)

### M1 — drift gate

#### REQ-GF-001 — Per-layer numeric staleness report (Ubiquitous)

The `moai graph check` command shall report, for each gated layer (codemaps, mx-index, edges.jsonl), the layer name, the metric kind used, the measured numeric value, the configured threshold, and a verdict (`fresh` | `stale` | `absent`). The report shall be numeric per layer — a bare boolean or prose-only verdict does not satisfy this requirement.

#### REQ-GF-002 — Per-layer staleness metrics with per-layer rationale (Ubiquitous)

The check shall compute each layer's staleness with the layer-specific metric below, chosen because the layer's tracking status and source model force it:

| Layer | Tracked | Metric | Rationale |
|---|---|---|---|
| codemaps | git-tracked | Count of described-source files (trees the codemaps describe, e.g. `internal/`, `cmd/`, `pkg/`) whose working-tree content differs from the content at the layer's stamped generation commit | Tracked in git, so endpoint content comparison between two commits is exact and revert-proof; commit-count alone overstates drift through reverted churn |
| mx-index | untracked runtime state | Count of scanner-read files whose working-tree content differs from the content recorded in the index's stamped scan inventory; index file absent ⇒ verdict `absent` | Untracked ⇒ no git history exists to count; the index is mechanical and cheap to rescan, so any content drift in what it indexed is actionable |
| edges.jsonl | untracked derived artifact | Source-fingerprint mismatch: recompute the fingerprints of all source layers (codemaps set, mx-index, SPEC dir, reports dir) and compare against the fingerprints stamped at build time; artifact absent ⇒ verdict `absent` | A derived artifact goes stale exactly when its sources move; its own history is unobservable (untracked), so source-side fingerprints are the only sound signal |

Comparison anchor under a dirty generation: **Where** the codemaps set was generated on a dirty working tree (provenance `dirty` + content fingerprint per REQ-GF-003), the metric's reference point is the stamped generation-time content fingerprint — the content the artifact was actually generated from — and the working-tree content is compared against that anchor, never against a named commit the generation did not actually see.

The check shall not use filesystem modification time as a staleness signal for any layer: a fresh worktree checkout resets every mtime (observed in this tree — all codemaps files carry the checkout timestamp), which an mtime-based metric would misread as freshly regenerated.

#### REQ-GF-003 — Provenance stamping on all gated artifacts (Ubiquitous)

Every gated artifact shall carry a provenance block recording: the absolute tree root it was generated in, the commit SHA at generation time (or `dirty` with a content fingerprint when the working tree had uncommitted changes to the described sources), and the per-layer source fingerprint REQ-GF-002 relies on. Existing artifacts without a provenance block shall be reported by the check as `absent`-equivalent (cannot be freshness-judged), not silently treated as fresh.

#### REQ-GF-004 — Exit-code discipline (When)

**When** any layer's measured value exceeds its configured threshold or any layer's verdict is `absent`, `moai graph check` shall exit non-zero (exit 1) and name the offending layer, value, and threshold in its output; **When** all layers are within threshold, the command shall exit 0.

#### REQ-GF-005 — `moai gate` integration (Where)

**Where** the graph-freshness step is enabled in gate configuration, `moai gate` shall run the check and fail on its failure; **Where** the step is disabled or in warn-only mode, the gate shall emit an explicit notice carrying the check's verdict, never silence. Default mode for the gate step is warn-only (matching the ast-grep precedent: pre-commit is the wrong place to force a codemaps regeneration).

#### REQ-GF-006 — CI wiring (When)

**When** CI runs on a pull request or push to main, a job shall build the binary from the PR head, bootstrap the untracked mechanical layers to that head (`moai mx scan`, then `moai graph build`), then run `moai graph check`, and fail the build when the check exits non-zero. The bootstrap scopes the CI signal to codemaps drift — the tracked, curated layer: the mechanical layers are refreshed to head by the job itself, so an `absent` verdict for an untracked layer cannot fire in CI. CI-side thresholds may be configured independently of local defaults.

### M2 — query-time refresh

#### REQ-GF-007 — Refresh before answering (When)

**When** `moai graph query` or `moai mx query` runs and a mechanical input layer it reads (edges sources, mx-index) is stale by REQ-GF-002's metric, the command shall refresh that layer before answering, re-reading only files whose content hash changed, making no LLM calls and no network requests, and reflecting uncommitted working-tree edits in the refreshed result.

#### REQ-GF-008 — Per-tree cache anchoring (Ubiquitous)

The refresh cache shall be keyed by the absolute tree root and content hashes, so no two trees share cache entries; every query answer shall name the tree root and commit (or `dirty` + fingerprint) it was computed from. The cache shall never be keyed by repository identity alone.

#### REQ-GF-009 — Update-cost budget (When)

**When** a refresh's measured cost exceeds the configured update-cost budget, the command shall emit a warning naming the measured cost and the budget, and shall still answer. Budget defaults are hypotheses until measured in this repository and shall be calibrated from local measurement, never from foreign-repository figures.

### M3 — citation convention switch

#### REQ-GF-010 — Excerpt + content hash as citation canon (Ubiquitous)

Codemaps and generated reports produced or regenerated under this SPEC shall cite code by excerpt (verbatim source snippet) plus content hash of the cited region; line numbers may appear only as convenience notation alongside the canon, never as the sole anchor.

#### REQ-GF-011 — mx-index content-hash anchoring (Ubiquitous)

The mx index shall anchor tag locations by file path plus content hash, with line numbers carried as convenience data, so tag lookups survive line drift.

#### REQ-GF-012 — Two-tree resolution guarantee (When)

**When** the same citation (excerpt + content hash) is resolved in two trees whose cited files differ only by line drift (e.g. lines inserted above), the citation shall resolve to the same target in both trees.

### M4 — symbol layer

#### REQ-GF-013 — astx decoupled from navigator (Ubiquitous)

The astx extraction package shall be importable by non-navigator consumers (e.g. the graph builder) without pulling navigator-tier dependencies into the consumer.

#### REQ-GF-014 — Code-derived edge layers added additively (Ubiquitous)

`moai graph build` shall add code-call and code-import edge layers derived from AST extraction to edges.jsonl, alongside the existing doc-derived layers (import-from-codemaps, mx-spec, spec-depends, report-milestone, milestone-card). The code-derived layers shall not replace, drop, or rewrite any doc-derived edge.

#### REQ-GF-015 — Layer disagreement exposed (When)

**When** a doc-derived edge and a code-derived edge disagree about the same relationship, the artifact and query surface shall expose both edges plus an explicit disagreement marker; the system shall not silently select one layer's claim.

#### REQ-GF-016 — Resolution-grade matrix, no empty cells (Ubiquitous)

The build shall publish a per-language resolution-grade matrix with a grade (`full` scope-aware resolution | `name-based` resolution | `none` extraction unavailable) for every supported language; **When** a matrix cell lacks a grade, the build shall report it as a defect verdict rather than omitting the cell.

### M5 — MCP code queries

#### REQ-GF-017 — `graph_file_api` tool (Ubiquitous)

The MCP server shall expose `graph_file_api`, answering a file with its exported signatures only (no full source bodies), implemented first among the three code-query tools.

#### REQ-GF-018 — `graph_find_code` and `graph_trace_calls` tools (Ubiquitous)

The MCP server shall expose `graph_find_code` (symbol/text search over the code-derived layer) and `graph_trace_calls` (caller/callee traversal over code-call edges), both answering from the M4 layer.

#### REQ-GF-019 — Signature-level output discipline with provenance (Ubiquitous)

All three code-query tools shall return signature-level output and shall name the tree root and commit each answer was computed from.

#### REQ-GF-020 — Baseline measured before M5 work (When)

**When** M5 implementation begins, a recorded baseline of Grep/Read tool-call counts for a fixed task set shall already exist (measured before any M5 implementation commit); M5's completion judgment shall compare against that baseline, and a baseline produced after implementation began shall not satisfy this requirement.

## §E. Constraints

- **No new dependencies.** The symbol layer uses the already-present `smacker/go-tree-sitter` and the `.scm` query mechanism. Graft is NOT installed; its TS/npm runtime conflicts with the single Go binary and 16-language template neutrality.
- **No LLM calls** in any freshness check, refresh, or citation path added by this SPEC. These are mechanical surfaces; cost figures from the graft analysis (`~3ms`, token-reduction percentages) are foreign-repository measurements and shall not be copied into moai-adk docs or code as promises.
- **Cross-platform**: all path handling works on linux/darwin/windows; tests use `t.TempDir()` isolation.
- **Hardcoding prevention**: thresholds and budgets live in config (defaults in `internal/config/defaults.go` / gate config), not inline constants at use sites; env-var names via `internal/config/envkeys.go`.
- **Template neutrality**: any file mirrored into `internal/template/templates/` carries no SPEC IDs, dates, or internal commit SHAs (C1-C8 catalogue).
- **Exit-code discipline**: `0` pass, `1` check failure, `2` system error — per `internal/cli` conventions.
- **16-language equality**: the resolution-grade matrix and symbol layer treat all 16 supported languages equally; no language is privileged beyond its measured grade.

## §F. Exclusions (What NOT to Build)

### Out of Scope — Graft installation

- Installing, vendoring, or shelling out to the Graft tool, its npm/TS runtime, or its MCP server. Design principles only are adopted.
- Porting Graft's hook set (session-start/prompt/post-tool-use/stop hooks for graph refresh).

### Out of Scope — LLM concept-node layer

- Graft's layer 2 (LLM-generated concept nodes with Summary/Crux/Links/Notes). moai-adk already carries a human-authored semantic layer (SPEC, @MX, codemaps); a second machine semantic layer has no disagreement detector and is not adopted.

### Out of Scope — benchmark importation

- Copying Graft's self-reported benchmark figures (−42% tokens, 54%→66% SWE-bench) into moai-adk documentation, README, or release notes as performance claims.

### Out of Scope — LSP-backed resolution

- LSP opt-in language servers for scope-aware call resolution. The matrix grades `full` vs `name-based` resolution from AST extraction only.

### Out of Scope — doc-derived layer replacement

- Removing, rewriting, or "correcting" doc-derived edges when code-derived edges disagree (additive-only; REQ-GF-015 governs disagreement).

### Out of Scope — nav-graph.json gating

- Extending the freshness gate to `.moai/project/navigator/nav-graph.json`; it carries its own provenance scheme.

## §G. Success Criteria

Mirrored from the graft analysis §성공 판정 (operator-mandated verdicts):

| Milestone | Success verdict |
|---|---|
| M1 | The stale gap is answered numerically; deliberately aging a layer turns the gate actually red (non-zero exit observed, not asserted) |
| M2 | Uncommitted edits are reflected in query answers; refresh cost within budget on this repository's measurement |
| M3 | One citation resolves to the same target in two trees despite line drift |
| M4 | A real call edge absent from docs appears in blast-radius answers; language grade matrix published with no empty cells |
| M5 | Grep/Read call-count reduction on a fixed task set vs a baseline measured before work began |

## §H. Cross-References

- Evidence base: `.moai/reports/graft/graft-analysis-20260824.md` (operator-mandated; measured at `294b4b6ab`) — the report file is untracked and lives only in the primary checkout's `.moai/reports/graft/`; it is not present in worktrees
- Anchor re-verification in this tree: `research.md` §Anchor Re-Verification
- Related SPECs: `SPEC-V3R6-DOCS-CODEMAPS-V3-001` (codemaps v3 SSOT), `SPEC-DWF-CODEMAPS-PILOT-001` (codemaps pilot) — neither covers freshness, symbol layer, or code queries
- Edge layers today: `internal/cli/graph.go` `moai graph build` (5 doc-derived layers)
- Provenance precedent: nav-graph `extract_commit_sha` (`.claude/rules/moai/workflow/nav-tokens.md`)
- MCP wiring: `internal/cli/mcp_server.go` (`add()` registration + tool catalog)
