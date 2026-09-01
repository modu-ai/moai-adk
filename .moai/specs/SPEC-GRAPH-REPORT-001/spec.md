---
id: SPEC-GRAPH-REPORT-001
title: "Graph report toolchain: shortest_path MCP query, moai graph report, edges shrink guard, deferred SessionStart edges refresh"
version: "0.1.0"
status: draft
created: 2026-09-02
updated: 2026-09-02
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/graph, internal/cli, internal/hook"
lifecycle: spec-anchored
tags: "graph, mcp, report, shrink-guard, deferred-refresh, edges"
era: V3R6
tier: M
depends_on: [SPEC-MX-TAG-EDGES-001]
---

## §A. Problem Statement

Card t413 adopts the P2 band of the graphify gap analysis (`.moai/reports/graphify-codegraph-analysis-20260901.html`, findings P2-4 through P2-7) into moai-adk-go's existing code-graph subsystem (`internal/graph`, 3 MCP tools, `moai graph build/query/check/stamp`). The analysis confirmed moai's graph already leads on freshness (commit-SHA + content-fingerprint gates, mtime forbidden) and governance edges, and identified four surface gaps graphify validates with mature patterns:

1. **P2-4 — no A→B path query.** `graph_trace_calls` answers "who reaches this symbol" but there is no point-to-point reachability query over the code call graph. graphify exposes `shortest_path` capped at 8 hops; moai's `graph_trace_calls` already enforces `maxTraceDepth = 8` (`internal/graph/codequery.go:21`), so a 4th code-graph MCP tool aligns with the existing cap and registration pattern.
2. **P2-5 — no architecture report.** The graph holds god-node evidence (fan-in), confidence-tagged edges (post SPEC-GRAPH-EDGE-CONFIDENCE-001), and import adjacency, but nothing renders it as a human/auditor-readable artifact. graphify's `GRAPH_REPORT.md` (god nodes, surprising connections, import cycles — all deterministically computed) is the model; moai's translation scores INFERRED edges that cross package boundaries, and the report serves as sync-auditor architecture-review input.
3. **P2-6 — no shrink guard.** `refreshEdgesArtifact` (`internal/cli/graph_refresh_cli.go:53-71`) unconditionally overwrites `edges.jsonl` after a rebuild. A partially-failed rebuild (extraction error, skipped layer) silently shrinks the graph — the exact accident graphify's #1116 shrink guard prevents: refuse the overwrite when removed edges' provenance lies outside the rebuild set.
4. **P2-7 — no background freshness.** Edges staleness is repaired only lazily, at query time, inside the answering call's budget. The SessionStart deferred-scan pattern (`spawnDeferredAdvisoryScans` → durable side effect after the advisory payload ships) is the established house pattern for exactly this shape; extending it to the edges layer moves rebuild cost off the interactive path while preserving the derived-artifacts-never-committed principle.

**Dependency and sequencing**: SPEC-MX-TAG-EDGES-001 (card t412, branch `WT-mx-tag-edges`) is IN-FLIGHT in the same domain — its run-phase M1 (mx-tag edge layer: six tag kinds, range join, single-scan seam) is committed but not yet on develop. Shared files: `internal/cli/mcp_code_tools.go`, `internal/cli/graph.go`, `internal/graph/query.go`, `internal/graph/graph.go` (plus the `reader.go` rebuild path). Every milestone touching those files is sequenced AFTER absorbing origin/develop once t412 lands, with a re-measure step on the absorbed tree. The deferred-refresh milestone (M4) shares no file with t412 and carries no such gate.

## §B. Scope

### §B.1 In Scope

Four milestones (M1 → M3 gated on the t412 absorb; M4 ungated — see §B.2):

- **M1 — `graph_shortest_path` MCP tool (P2-4)**: a shortest-path A→B query over the code-derived edge layer, hop cap 8 (shared bound with `graph_trace_calls`), registered as the 4th code-graph MCP tool with `project_root` handling, read-only hint annotation, and `toolJSON`/`toolErr` result shaping per the existing three tools' pattern. Deterministic neighbor iteration (total order).
- **M2 — `moai graph report` (P2-5)**: a new `moai graph` subcommand emitting a deterministic markdown report to `.moai/reports/`: god nodes (fan-in ranking over edge layers), surprising connections (INFERRED code-call edges crossing package boundaries, scored up), and import cycles. Labeled fan-in provenance; no MX-validator behavior change.
- **M3 — edges shrink guard (P2-6)**: a graph-package guard that compares a rebuild's edge set against the existing artifact and refuses the overwrite when a removed edge's source lies outside the rebuild's scanned source set (graphify #1116 pattern), enforced on the automatic rebuild-write paths.
- **M4 — deferred SessionStart edges refresh (P2-7)**: extends the `spawnDeferredAdvisoryScans` pattern so a stale edges layer is refreshed in the deferred goroutine after the advisory payload ships — buffered channel, join bound, fail-open, duration measured by the existing `edgesRefreshClock` seam, nothing staged or committed.

### §B.2 Boundary decisions

- **Execution ordering vs milestone numbering**: M4 shares no file with t412 and MAY execute before the absorb gate; M1-M3 each begin only after origin/develop is absorbed post-t412 and the plan's re-measure step (plan.md §F) has run on the absorbed tree.
- **Guard scope**: the shrink guard binds the AUTOMATIC rebuild paths (query-time `refreshEdgesArtifact`, deferred background refresh) and the explicit `moai graph build` alike — a human-invoked build is equally capable of a partial failure, and refusal preserves the existing artifact either way. No override flag: the remedy is fixing the partial failure and rebuilding, not forcing a known-shrunk write.
- **Fan-in provenance is labeled, not switched**: the report's fan-in section states which edge layers it counted. The MX validator's grep-based `fan_inIndex` (`internal/hook/mx/validator.go`) is UNTOUCHED by this SPEC — the grep→graph fan-in migration with parallel observation belongs to SPEC-MX-TAG-EDGES-001's migration gate.
- **nocgo builds**: missing code edges under `CGO_ENABLED=0` are an existing constraint, not a regression (analysis risk 3). The report states the reason in its empty code sections; `graph_shortest_path` surfaces the existing actionable "graph layer absent" error shape.

## §C. History

| Date | Event |
|---|---|
| 2026-09-01 | Graphify gap analysis report authored (read-only exploration + graphify v8 source review); P2 band identified as card t413 |
| 2026-09-02 | Card t413 picked; SPEC-GRAPH-REPORT-001 authored (Tier M) against worktree base 9145806d8 |

## §D. Requirements (GEARS notation)

### M1 — graph_shortest_path MCP tool

#### REQ-GR-001 — Fourth code-graph MCP tool (Ubiquitous)

The code-graph MCP surface shall expose a fourth tool, `graph_shortest_path`, answering an A→B reachability query over the code-derived edge layer, registered alongside `graph_file_api` / `graph_find_code` / `graph_trace_calls` in the MCP server with the same contract: a supplied `project_root` is honored and a bad one REJECTED (never a silent fallback), read-only hint annotation set, results shaped by the package's `toolJSON`/`toolErr`, and every response carrying tree+commit provenance.

#### REQ-GR-002 — Shared hop cap (Ubiquitous)

The shortest-path traversal shall be bounded at 8 hops, enforced by the same constant that bounds `graph_trace_calls` (`maxTraceDepth`) — one bound, two consumers, never a duplicated literal; the MCP tool description restates the cap for human readers exactly as the trace tool's description does.

#### REQ-GR-003 — Structured no-path result (When)

**When** no path from A to B exists within the hop cap, the tool shall return a structured not-found result naming both endpoints, the cap, and the provenance — a valid answer shape, never a transport-level error.

#### REQ-GR-004 — Deterministic traversal (Ubiquitous)

The shortest-path search shall iterate neighbor edges in a total order (node id, then line number), so two runs over the same tree produce byte-identical responses — applying the graphify determinism lesson (edge normalization + total-order re-indexing, analysis §P3-8) even though clustering itself is out of scope.

### M2 — moai graph report

#### REQ-GR-005 — Deterministic report artifact (Ubiquitous)

The `moai graph report` subcommand shall emit a markdown report to a fixed path under `.moai/reports/` containing three deterministically ordered sections: (1) god nodes — packages ranked by distinct fan-in over the edge layers, highest first, ties broken by node id; (2) surprising connections — INFERRED code-call edges whose endpoints lie in different packages, ranked above same-confidence intra-package edges; (3) import cycles — every cycle detected over import edges, each rendered with its member nodes in a canonical rotation. Two runs over the same tree shall produce byte-identical report files.

#### REQ-GR-006 — Empty-section tolerance (When)

**When** a section has no findings — no import cycles, or a code layer absent under a nocgo build — the report shall still be emitted with that section present but empty and the reason stated (e.g. "code layer absent: CGO disabled or no extraction"), never a missing file, never an error exit.

#### REQ-GR-007 — Fan-in provenance labeling (Ubiquitous)

The report's fan-in section shall name the edge kinds it counted, and this SPEC shall not modify the MX validator's grep-based fan-in computation: during the grep→graph migration the two operate in parallel, and a validator-verdict change is t412's migration gate to pass, not this SPEC's side effect.

### M3 — edges shrink guard

#### REQ-GR-008 — Refuse unexplained shrink (When)

**When** an edges rebuild produces a set smaller than the existing artifact and at least one removed edge's source file lies outside the rebuild's scanned source set, the rebuild path shall refuse the overwrite: the existing `edges.jsonl` and its meta sidecar remain byte-identical, and the refusal names the removed edges and their unscanned source files (the graphify #1116 pattern) so the partial failure is diagnosable rather than silent.

#### REQ-GR-009 — Guard on every automatic write path (Ubiquitous)

The shrink guard shall evaluate on every automatic rebuild-write path — the query-time refresh and the deferred background refresh — and its refusal shall be fail-safe: the caller answers from the existing artifact (query path) or skips the refresh (deferred path) with a stated warning, never writes a shrunk graph, and never loses the prior artifact.

### M4 — deferred SessionStart edges refresh

#### REQ-GR-010 — Deferred refresh after advisories (When)

**When** a session starts and the derived edges layer is stale (its source fingerprints moved or the mx-index layer drifted, per the existing refresh-needed predicate), the SessionStart deferred goroutine shall refresh the edges layer after the advisory payload has been sent, following the `spawnDeferredAdvisoryScans` contract: buffered result channel that never blocks the session, bounded join, best-effort and fail-open.

#### REQ-GR-011 — Derived artifacts never committed (Ubiquitous)

The deferred refresh shall perform no git operation: `edges.jsonl` remains an untracked derived artifact, nothing is staged or committed, and a refresh failure leaves the prior artifact intact with the failure logged, never blocking session start.

#### REQ-GR-012 — Refresh cost observability (When)

**When** the deferred refresh runs, its duration shall be measured through the existing refresh-duration seam and evaluated against the configured update budget, producing the same warning-only overrun signal as the query-time refresh — the observation channel for the full-rebuild cost trend (analysis risk 2) that later decides whether incremental rebuild is ever needed.

## §E. Constraints

- **No new dependencies** — no graph library, no clustering dependency; the path search and cycle detection are standard-library BFS/DFS over the existing edge slice.
- **Shared-file sequencing gate (HARD)**: M1-M3 touch files shared with in-flight SPEC-MX-TAG-EDGES-001 (`internal/cli/mcp_code_tools.go`, `internal/cli/graph.go`, `internal/graph/query.go`, `internal/graph/graph.go`, `reader.go` rebuild path). Each gated milestone begins only after `git merge origin/develop` absorbs t412's landed work and the plan.md §F re-measure step has run on the absorbed tree.
- **Verification discipline**: targeted `go test ./internal/<pkg>/...` for affected packages only — NEVER `go test ./...` locally (CLAUDE.local.md §4/§6); CI is the full-suite judge. `GOOS=windows GOARCH=amd64 go build ./...` for cross-platform parity on touched packages.
- **Test isolation**: new tests use `t.TempDir()`; no OTEL env vars in parallel tests; no goroutine leaks in SessionStart tests (join via the `completed` seam per the existing goleak-hygiene pattern).
- **Determinism everywhere**: every new output surface (tool response, report file) is a pure function of the tree content — no wall-clock timestamps in bodies, stable sorts only.
- **Behavioral conservatism**: the guard refuses rather than repairs; the deferred refresh reuses `refreshEdgesArtifact` rather than forking a second rebuild path; no refactor of the existing rebuild pipeline beyond the guard's insertion point.

## §F. Exclusions (What NOT to Build)

### Out of Scope — clustering (P3-8)

- Leiden/Louvain community detection, any graspologic-equivalent dependency, community metadata on edges, and hub-exclusion percentile logic. Go package boundaries already cluster strongly; the determinism lessons (edge normalization, total-order re-indexing) are applied to shortest_path and the report instead (REQ-GR-004, REQ-GR-005).

### Out of Scope — incremental rebuild

- File-level manifests, extraction caching, per-file incremental invalidation. The edge count is currently six digits; the full-rebuild cost trend is observed first (REQ-GR-012's budget signal) before any incremental decision (analysis risk 2).

### Out of Scope — MX validator fan-in switch (P1-3)

- Replacing the grep-based `fan_inIndex` in `internal/hook/mx/validator.go` with edge aggregation. That switch is SPEC-MX-TAG-EDGES-001's migration obligation (parallel operation, observe, then switch). This SPEC only labels the report's own fan-in provenance (REQ-GR-007).

### Out of Scope — visualization and PreToolUse graph nudging (P3-9, P3-10)

- Mermaid/vis-network report rendering, interactive graph artifacts, and any PreToolUse hook steering agents toward graph queries before grep. Demand confirmed first.

### Out of Scope — graphify mechanisms moai declined (borrow-refusal list)

- LLM semantic layers and AMBIGUOUS edges, multimodal document extraction, Python runtime shell-outs, a monolithic `graph.json` artifact, the PR dashboard, and work-memory reflection — each conflicts with the deterministic zero-cost single-binary philosophy or duplicates an existing moai subsystem.

## §G. Success Criteria

| Milestone | Success verdict |
|---|---|
| M1 | `graph_shortest_path` appears in the MCP tool registry tests; hop-cap, no-path, determinism, and project-root-rejection tests green; targeted internal/graph + internal/cli tests pass |
| M2 | `moai graph report` emits the three-section report on a fixture (god nodes ranked, boundary-INFERRED edge scored up, cycle listed); empty-section and nocgo paths emit with stated reasons; two runs byte-identical |
| M3 | Injected partial-failure rebuild is refused with prior artifact byte-identical (SHA-compared) and a named-edges refusal message; legitimate rebuilds overwrite normally; guard tests green on both automatic paths |
| M4 | Synchronous-deferred-scans test shows edges refreshed after session start when stale and skipped when fresh; no goroutine leak; no git staging; budget-overrun warning observed with injected duration |
| Cross-cutting | Absorbed-tree re-measure (plan.md §F) recorded for M1-M3; `go vet` + `golangci-lint` clean on touched packages; windows cross-build green |

## §H. Cross-References

- Source analysis: `.moai/reports/graphify-codegraph-analysis-20260901.html` (P2-4..P2-7, risks 1-4, determinism lessons §P3-8)
- In-flight dependency: `.moai/specs/SPEC-MX-TAG-EDGES-001/` (branch `WT-mx-tag-edges`, card t412 — landed on develop before M1-M3 begin)
- Predecessor pattern: `.moai/specs/SPEC-V3R6-GRAPH-FRESHNESS-001/` + `-002/` (MCP tool registration, refresh path, duration seam, query-time staleness predicate this SPEC extends)
- Landed confidence layer: `.moai/specs/SPEC-GRAPH-EDGE-CONFIDENCE-001/` (Edge.Resolution/Confidence consumed by the report's INFERRED scoring)
- Key code anchors: `internal/graph/codequery.go:21` (`maxTraceDepth`), `internal/cli/graph_refresh_cli.go:53` (`refreshEdgesArtifact` — guard insertion point), `internal/hook/session_start.go:680` (`spawnDeferredAdvisoryScans` — deferred pattern), `internal/cli/mcp_server.go:486-509` (tool registration pattern)
