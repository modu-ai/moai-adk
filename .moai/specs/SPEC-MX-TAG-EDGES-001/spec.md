---
id: SPEC-MX-TAG-EDGES-001
title: "MX tag-kind edges in the code graph + graph-backed fan-in for the MX validator P1 rule (graphify pattern P1-2 + P1-3)"
version: "0.2.0"
status: completed
created: 2026-09-01
updated: 2026-09-02
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/graph, internal/hook/mx, internal/cli"
depends_on: [SPEC-GRAPH-EDGE-CONFIDENCE-001]
lifecycle: spec-anchored
tags: "graph, mx-tags, edges, fan-in, validator, code-call, confidence, edges-jsonl"
era: V3R6
tier: M
---

## §A. Problem Statement

Two coupled gaps, both recorded by the graphify gap analysis
(`.moai/reports/graphify-codegraph-analysis-20260901.md`, recommendations P1-2
and P1-3):

1. **Tag-kind edges (P1-2).** Only `@MX:SPEC` sub-lines become graph edges
   (`mx-spec`, via `mx.Scanner` SpecRef capture). The other five standalone tag
   kinds — ANCHOR, DEBT, WARN, NOTE, TODO, LEGACY — are invisible to
   `edges.jsonl`, so the graph cannot answer tag questions. The stand-in is
   documented at `internal/graph/query.go` `ImportFanIn` (package-import
   ranking "stands in for an @MX:DEBT fan-in query until a tag-kind edge lands
   in edges.jsonl") and repeated in the `moai graph query --fanin` help text.
2. **Grep-based fan-in (P1-3).** The MX validator P1 rule — "fan_in ≥ 3
   callers ⇒ @MX:ANCHOR required" — is computed by a single-traversal
   word-occurrence index (`internal/hook/mx/validator.go` `fanInIndex`,
   threshold 3 at `newValidatorWithConfig`). It counts identifier TOKEN
   occurrences, not callers: comments, strings, dead code, and test files all
   inflate the count. The graph's `code-call` edges (now carrying per-edge
   resolution confidence per SPEC-GRAPH-EDGE-CONFIDENCE-001, already merged in
   this tree) are a strictly better evidence source — but the validator runs
   inside the PostToolUse hook under a 500ms budget
   (`ValidationConfig.PostToolUse.TimeoutMs`, default 500), while a graph
   build/refresh costs more.

Current-tree facts (re-derived on `WT-mx-tag-edges` @ `57d2f3ae3` =
origin/develop `07ee6e74a` + absorbed SPEC-GRAPH-EDGE-CONFIDENCE-001,
2026-09-01 — the analysis report cited main @ `48239c7dc`; drift notes in
plan.md §J):

- `internal/graph/graph.go:50-77` — `Edge{Kind, Source, Target, Line, Grade,
  DisagreesWith, Resolution, Confidence}`; doc kinds `import`/`mx-spec`/
  `spec-depends` (graph.go:37-45); code kinds `code-call`/`code-import`
  (symbol.go:13-18); report kinds `report-milestone`/`milestone-card`.
  `EdgeLess` sorts by (kind, source, target, line); `WriteJSONL` atomic;
  `Build()` aggregates doc layers; `BuildWithCodeLayersMode` adds code layers.
- `internal/graph/graph.go:252-277` — `mxSpecEdges` already runs ONE
  `mx.Scanner.ScanDir` over the project and keeps only `SpecRef`-bearing tags;
  the remaining tags are discarded today.
- `internal/mx/tag.go` — `Tag{Kind, File, Line, Body, Reason, SpecRef, RotRisk,
  AnchorID, ContentHash, CreatedBy, LastSeenAt}`; `TagKind` closed domain of 6
  standalone kinds (NOTE/WARN/ANCHOR/TODO/LEGACY/DEBT). `LastSeenAt` is
  WALL-CLOCK (`time.Now()` at scan time) — must never reach the artifact.
- `internal/graph/symbol/symbol.go:112-206` — `Extract` walks the described
  roots (skipping `testdata/` and dot-dirs), returns calls/imports/`FileDecls
  {File, Names}`. `astx.FuncRange` (StartLine/EndLine) is consumed only
  transiently for `enclosingFunction` and then DROPPED — the seam does not
  retain function ranges today.
- `internal/graph/query.go:16-42` — `ImportFanIn` (import-kind ranking) with
  the stand-in `@MX:NOTE` at line 15; `reader.go` `FindCallers` matches ALL
  edge kinds by target; `BlastRadius` is reverse-BFS with ONLY `mx-spec`
  bidirectional.
- `internal/hook/mx/validator.go` — `fanInIndex` single-traversal identifier
  count per `analyzeFile` call (O(project bytes), .go files, skipping vendor/
  dot-dirs/`*_generated.go`/`mock_*.go`); P1 rule at lines 270-283 (`fanIn >=
  3`, Blocking); `internal/hook/post_tool.go` constructs the validator with a
  500ms default timeout; `internal/hook/session_end.go:296` batch-validates
  under the 4000ms SessionEnd budget.
- `internal/cli/graph.go:96-97,185-190` — `moai graph query --fanin` with the
  "stand-in … carries no tag-kind edges yet" help text.
- `internal/mx/scan_ignore.go:12-13` — documented layering constraint:
  `internal/hook` cannot import `internal/cli` (import cycle).
- Golden: `internal/graph/testdata/edges-doc-golden.jsonl` (8 lines) pins doc
  kinds + `code-import` under a filtered-per-kind comparison; `code-call` lines
  are deliberately NOT pinned (CGO-availability-dependent).

## §B. Goals

1. Every standalone @MX tag occurrence becomes a deterministic edge in
   `edges.jsonl`, giving the graph a tag-question surface (P1-2).
2. The validator P1 fan-in rule gains a graph-backed evidence source —
   distinct callers over confidence-bearing `code-call` edges — as the
   AUTHORITY on batch surfaces, with the hook's instant textual signal
   retained where the 500ms budget governs (P1-3).
3. The `ImportFanIn` stand-in retires: a real `@MX:DEBT` fan-in query mode
   lands on `moai graph query`.
4. Every existing edge kind, the freshness/provenance contract, and the
   PostToolUse cost profile remain byte-for-byte and behaviorally unchanged.

## §C. Requirements (GEARS)

### REQ-MTE-001 — Tag-kind edge kinds (Ubiquitous)

The graph's doc layer shall emit one edge per standalone @MX tag occurrence
found by the scanner, as exactly one edge kind per tag kind from the closed
`TagKind` domain — `mx-note`, `mx-warn`, `mx-anchor`, `mx-todo`, `mx-legacy`,
`mx-debt` — named by lowercasing the tag kind and prefixing `mx-`.

### REQ-MTE-002 — Edge endpoints (Event-driven)

When the builder can join a tag's line to a function/method declaration range
in the same file, the tag edge shall carry `source` = repo-relative file path,
`target` = the enclosing symbol's name, and `line` = the tag's 1-based line;
when no range contains the tag line (file-scope tag, or no range data), the
edge shall carry `target` = the repo-relative file path itself (a self-edge),
so every tag occurrence is represented and none is silently dropped.

### REQ-MTE-003 — Determinism (Event-driven)

When the graph is built twice over the same tree under the same build
configuration, the builder shall produce byte-identical `edges.jsonl`
output including the `mx-*` kinds — no wall-clock field (`Tag.LastSeenAt`,
`CreatedBy`), no map-iteration order, and no scanner-mutable metadata may
reach a serialized edge.

### REQ-MTE-004 — Existing-edge byte identity (Ubiquitous)

The builder shall not change the serialized content of any pre-existing edge
line (`import`, `mx-spec`, `spec-depends`, `report-milestone`,
`milestone-card`, `code-call`, `code-import`); the committed golden
`internal/graph/testdata/edges-doc-golden.jsonl` shall be extended to also pin
the `mx-*` kinds and regenerated in the same change set per its documented
procedure (base SHA named). `code-call` lines remain unpinned
(CGO-availability-dependent), as today.

### REQ-MTE-005 — Legacy artifact compatibility (Event-driven)

When `edges.jsonl` contains no `mx-*` lines (a pre-upgrade artifact), every
graph consumer shall load and serve it without error — absent kinds are
absent edges, not a failure.

### REQ-MTE-006 — Single-scan reuse (Ubiquitous)

The builder shall derive `mx-spec` edges and `mx-*` tag edges from ONE
`mx.Scanner.ScanDir` pass per build — no second project walk is introduced
for the tag layer.

### REQ-MTE-007 — Metadata stays scanner-side (Ubiquitous)

The `mx-*` edge shall carry occurrence + endpoints only (kind, source,
target, line): tag content (`Body`, `Reason`, `Ceiling`/`Upgrade` text),
rot state (`RotRisk`), provenance (`CreatedBy`, `ContentHash`), and wall-clock
(`LastSeenAt`) stay scanner-side, where the mx sidecar remains the single
source of truth answerable via `moai mx query`.

### REQ-MTE-008 — Traversal semantics (State-driven)

While `FindCallers` or `BlastRadius` traverses edges, `mx-*` edges shall
propagate reverse-only (source affected by target — the same direction as
`import` and `spec-depends`, NOT bidirectional like `mx-spec`), and a graph
containing zero `mx-*` edges shall return byte-identical traversal results
to the pre-change behavior.

### REQ-MTE-009 — Graph-backed fan-in evidence source (Ubiquitous)

The graph shall expose a pure query — fan-in of a symbol name = the number of
DISTINCT caller FILES, excluding the file that declares the symbol, of
`code-call` edges targeting that name whose resolution confidence is
`extracted` or `intra-package` — suitable as the MX validator P1 rule's
evidence source. The declaring-file exclusion is ANCHOR's semantics: the tag
records an invariant contract owed to EXTERNAL dependents; a same-file call is
not external blast radius.

### REQ-MTE-010 — Two-tier wiring (State-driven)

While the PostToolUse 500ms budget governs, the hook shall keep the existing
single-traversal textual fan-in index; the graph-backed fan-in source shall be
the authority for the validator's P1 rule on the batch surface — the SessionEnd
batch validation (the sole non-PostToolUse validator construction site,
`session_end.go`), selected at validator construction. The complete non-test
construction set is exactly: `post_tool.go` (textual by budget, unchanged) +
`session_end.go` (graph-backed). `moai mx scan` constructs NO validator — its
role is sidecar producer feeding the authority's freshness probes, and this
SPEC shall not add validation behavior to it. The source choice is a
constructor-level seam, with no new config keys in this delivery.

### REQ-MTE-011 — Absent/stale artifact fallback (Event-driven)

When the graph-backed source is selected but `edges.jsonl` is absent (fresh
clone) or its freshness probe fails, the validator shall fall back to the
textual source, shall label which source produced the verdict, and shall not
surface an error — stale-but-labeled and textual-fallback both beat no answer.

### REQ-MTE-012 — Hub exclusion (Ubiquitous)

The graph-backed blocking fan-in count shall exclude callers in test files
(`*_test.go` suffix, `tests/`, `fixtures/`, `testdata/` path components — the
REQ-SPC-004-040 pattern set); vendor and generated files are already excluded
upstream by the extractor walk and the scanner ignore set.

### REQ-MTE-013 — Confidence ruling on the threshold (Ubiquitous)

The blocking `fan_in >= 3` threshold shall count evidence-backed callers only
(REQ-MTE-009's domain); `inferred`-confidence callers shall be counted
separately and carried in the violation's `Reason` text (e.g.
"fan_in(graph)=N evidence-backed (+M inferred-only)") — never added to the
blocking count.

### REQ-MTE-014 — DEBT fan-in query surface (Event-driven)

When `moai graph query --debt-fanin` runs, the CLI shall rank `mx-debt` edge
targets by the REQ-MTE-009 graph fan-in, descending, ties broken by target;
self-edge targets (file-scope DEBT tags) shall rank at fan-in 0 and shall be
listed explicitly rather than omitted; the `ImportFanIn` import-ranking query
shall retain its current behavior, and its stand-in framing (the `@MX:NOTE`
at `query.go` and the `--fanin` help text) shall be rewritten to point at the
new mode.

### REQ-MTE-015 — CGO-disabled behavior (Capability gate)

Where the build is CGO-disabled, `mx-*` edges shall still be emitted from the
scanner layer (line-based, CGO-free) with all targets resolving to the
self-edge form (no range data), the graph-backed fan-in source shall degrade
to the textual fallback with its label, and no new error surface shall
appear.

## §D. Acceptance Criteria — REQ→AC mapping

acceptance.md is the SINGLE SOURCE OF TRUTH for the AC set (Tier M): its §D AC
matrix carries the full Given-When-Then scenarios and verification commands.
This section carries only the traceability mapping.

| REQ | Verified by |
|---|---|
| REQ-MTE-001 | AC-MTE-002 |
| REQ-MTE-002 | AC-MTE-003 |
| REQ-MTE-003 | AC-MTE-001, AC-MTE-007 |
| REQ-MTE-004 | AC-MTE-004 |
| REQ-MTE-005 | AC-MTE-005 |
| REQ-MTE-006 | AC-MTE-006 |
| REQ-MTE-007 | AC-MTE-007 |
| REQ-MTE-008 | AC-MTE-008 |
| REQ-MTE-009 | AC-MTE-009 |
| REQ-MTE-010 | AC-MTE-010 |
| REQ-MTE-011 | AC-MTE-011 |
| REQ-MTE-012 | AC-MTE-012 |
| REQ-MTE-013 | AC-MTE-009 |
| REQ-MTE-014 | AC-MTE-013 |
| REQ-MTE-015 | AC-MTE-014 |

## §E. Constraints

- Determinism is the load-bearing artifact contract: same tree + same build
  configuration ⇒ byte-identical JSONL. `Tag.LastSeenAt` is wall-clock — the
  tag→edge mapping must be a pure function of (File, Kind, Line) plus tree
  content (the range join), never of scan time.
- Layering: `internal/hook/mx` stays graph-agnostic — the fan-in evidence
  source is an interface defined at the consumer (`internal/hook/mx`), the
  edge-backed implementation lives with the data (`internal/graph`), and
  wiring happens at the one construction site that may import
  `internal/graph` and needs it: `internal/hook/session_end.go` (the sole
  non-PostToolUse validator construction site; `moai mx scan` constructs no
  validator and is not touched).
  `internal/hook` cannot import `internal/cli` (documented cycle,
  `scan_ignore.go:12-13`) — this constraint binds the adapter's home.
- The enclosing-symbol join consumes only data `symbol.Extract` already walks:
  retaining `astx.FuncRange` in the seam is retention, not a second parse pass
  (same constraint pattern as SPEC-GRAPH-EDGE-CONFIDENCE-001's declared-names
  retention).
- Hub-exclusion divergence ACCEPTED (minimal change): the graph-backed fan-in
  uses the hard-coded REQ-SPC-004-040 fallback pattern set only and does NOT
  honor a user's mx.yaml `test_paths` globs (which the query-side
  `TextualFanInCounter` accepts). Wiring mx.yaml through the graph source is
  candidate future work; until then the two sources can diverge for a project
  that configures custom test-path globs.
- Target resolution is build-configuration-dependent BY DESIGN: CGO-on yields
  symbol targets, CGO-off yields self-edges. Each configuration is
  deterministic in itself; the golden pins the CGO-available output.
- `edges.jsonl` is untracked/derived: schema "backfill" is a rebuild
  (`moai graph build`), never an in-place migration tool.
- Repo conventions: `t.TempDir()` fixtures; no `t.Setenv` with OTEL vars;
  `fmt.Errorf("...: %w")`; English comments; no new goroutines in the build
  path (the confidence pass precedent — a synchronous post-walk).

## §F. Out of Scope

### Out of Scope — P2 graph tooling

- `shortest_path` MCP tool, `moai graph report`, shrink guard, and
  SessionStart deferred background edges refresh — separate card t413 /
  SPEC-GRAPH-REPORT-001. REQ-MTE-010 deliberately does NOT wire hook-side
  graph refresh: deferred refresh belongs to that card.

### Out of Scope — hook budget path replacement

- Replacing the PostToolUse textual fan-in index with a graph read inside the
  500ms budget. The textual index's O(project-bytes) cost per `analyzeFile`
  call is the status quo; optimizing it is separate work. This SPEC changes
  only WHICH source is the authority on batch surfaces.

### Out of Scope — tag metadata in the graph

- Carrying `Body`, `Reason`, `RotRisk`, `Ceiling`/`Upgrade` content, or any
  tag prose on `mx-*` edges; any new sidecar schema; any change to
  `moai mx query` filters or the mx-index refresh path.

### Out of Scope — analysis and clustering

- Any Leiden/community clustering, hub-exclusion GENERALIZATION beyond the
  REQ-MTE-012 test-pattern set, visualization, Mermaid callflow, LLM
  involvement, `GradeFull` assignment, and non-Go specifier resolution
  (inherited out-of-scope from SPEC-GRAPH-EDGE-CONFIDENCE-001).

## §G. Cross-References

- Evidence input: `.moai/reports/graphify-codegraph-analysis-20260901.md`
  (P1-2 + P1-3 ONLY; P1-1 landed via SPEC-GRAPH-EDGE-CONFIDENCE-001, P2 items
  are card t413).
- Predecessor layer: SPEC-GRAPH-EDGE-CONFIDENCE-001 (Edge.Resolution/
  Confidence fields, confidence tiers, golden mechanism — this SPEC builds on
  and extends the golden; never modifies that SPEC's semantics).
- Sibling card: SPEC-GRAPH-REPORT-001 (t413).
- Doctrine: `.claude/rules/moai/workflow/mx-tag-protocol.md` (tag semantics —
  unchanged by this SPEC; the graph READS tags, never writes them).
