---
id: SPEC-GRAPH-EDGE-CONFIDENCE-001
title: "Per-edge resolution confidence on code-call edges via import-evidence promotion (graphify pattern P1-1)"
version: "1.1.0"
status: in-progress
created: 2026-09-01
updated: 2026-09-01
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/graph"
lifecycle: spec-anchored
tags: "graph, code-call, confidence, edge-resolution, import-evidence, graphify, mcp"
era: V3R6
tier: M
---

## §A. Problem Statement

Every `code-call` edge in `.moai/project/graph/edges.jsonl` is resolved by symbol name
only: `astx.GradeFor` returns `name-based` for all six call-grade languages (go,
python, javascript, typescript, java, rust), and `GradeFull` (scope-aware resolution)
is defined but never assigned — the extractor captures the callee as written, matches
it without scope, and the persisted edge carries no signal of HOW strongly the
relationship is believed. A consumer of `graph_trace_calls` cannot distinguish a call
the extractor is sure about from a name coincidence.

The graphify comparison (`.moai/reports/graphify-codegraph-analysis-20260901.md`,
recommendation P1-1) validates the pragmatic bypass: per-edge resolution confidence
with import-evidence promotion — EXTRACTED (1.0) when evidence supports the
resolution, INFERRED (0.85) when only the name matched. moai already persists the
evidence needed for the promotion join: `code-import` edges record every file's
import set, and the astx `SymbolSet.Functions` ranges already enumerate each file's
declared function names — the join is derivable from data already extracted, with no
new parsing.

This SPEC adds that confidence layer to `code-call` edges only. It does NOT assign
`GradeFull` (the per-language capability matrix stays exactly as it is — the grade
axis and the confidence axis are orthogonal and this SPEC touches only the latter).

Current-tree facts (all re-derived on `WT-edge-confidence` @ origin/develop base,
2026-09-01 — the analysis report's citations were measured on main @ `48239c7dc`):

- `internal/graph/graph.go:50` — `Edge{Kind, Source, Target, Line, Grade, DisagreesWith}`.
- `internal/graph/symbol.go:14-18` — `KindCodeCall = "code-call"`, `KindCodeImport = "code-import"`;
  `BuildWithCodeLayersMode` (symbol.go:94) merges doc + code layers, sorted by
  `EdgeLess` (kind, source, target, line).
- `internal/graph/symbol/symbol.go:95` — `Extract` returns `[]CallEdge` (File, Caller,
  Callee, Line, Grade), `[]ImportEdge` (File, Module, Local, Line, Grade), grade matrix.
- `internal/navigator/astx/astx.go:145-172` — grade constants, `callGradeLanguages`
  (6 languages, all `GradeNameBased`), `GradeFor`.
- Consumers: `internal/graph/codequery.go` — `CodeMatch{Symbol, File, Line, Grade, Via}`
  (line 123), `CallTraceEdge{From, To, Line, Grade}` (line 170), `maxTraceDepth = 8`;
  `internal/cli/mcp_code_tools.go` — `handleGraphFindCode` / `handleGraphTraceCalls` /
  `handleGraphFileAPI`; `internal/graph/query.go` — `ImportFanIn`, `UnreferencedSpecs`;
  `internal/graph/reader.go` — `FindCallers`, `BlastRadius`; freshness gate
  `internal/graph/check.go` `checkEdges` (fingerprint over SOURCE files, not artifact
  bytes) wired through `internal/cli/gate.go` graph-freshness step.

## §B. Goals

1. Every persisted `code-call` edge carries a machine-readable resolution confidence:
   `extracted` (1.0), `intra-package` (0.95), or `inferred` (0.85), derived
   deterministically from tree content.
2. The three promotion tiers match the graphify-validated pattern: same-file
   declaration evidence, cross-file import evidence, name-only fallback.
3. Zero disturbance to every other edge flow: `import`, `mx-spec`, `spec-depends`,
   `report-milestone`, `milestone-card`, `code-import` edges serialize exactly as
   before; doc-derived bytes are unchanged.
4. Determinism invariant preserved: same tree → byte-identical `edges.jsonl`.

## §C. Requirements (GEARS)

Scope note on the value domain: the evidence source (graphify) carries a
three-value domain — EXTRACTED / INFERRED / AMBIGUOUS. This SPEC deliberately
reduces it to two promotion classes plus the same-package middle tier
(`extracted` / `intra-package` / `inferred`): AMBIGUOUS is dropped because this
delivery is annotation-only — no consumer acts on the label, so a third
"unsure" tier would add taxonomy without behavior. Revisit when a
confidence-consuming consumer lands.

### REQ-GEC-001 — Confidence on every code-call edge (Ubiquitous)

The code-call edge layer shall carry, on every `code-call` edge persisted to
`edges.jsonl`, a resolution label (`resolution` ∈ {`extracted`, `intra-package`,
`inferred`}) and its numeric confidence (`confidence` ∈ {1.0, 0.95, 0.85}), where
confidence is a pure function of the resolution label defined at exactly one point
in `internal/graph`.

### REQ-GEC-002 — Same-file promotion (Event-driven)

When a `code-call` edge's callee name matches a function or method declared in the
caller's own file, the builder shall mark the edge `extracted` with confidence 1.0.

### REQ-GEC-003 — Import-evidence promotion, Go-module domain (Event-driven)

When a cross-file `code-call` edge's callee name is declared in a module that the
caller file imports — joinable against the already-extracted `code-import` evidence
(the caller file's import set) — and that import specifier is Go-module-shaped
(i.e. the module prefix from `go.mod` was strippable, per
`symbol.localizeModule`'s Local flag), the builder shall mark the edge `extracted`
with confidence 1.0. EFFECTIVE DOMAIN: Go-module imports only. `ImportSite.Module`
is the specifier as written (`astx.go` `ImportSite` doc), and
`localizeModule` resolves declaring directories only for module-prefixed paths —
for the non-Go call-grade languages (python, javascript, typescript, java, rust)
specifier→declaring-directory resolution is not derivable today, so those calls
resolve via REQ-GEC-012 (same directory) or REQ-GEC-004 (fallback), never via
this requirement.

### REQ-GEC-004 — Name-only fallback (Ubiquitous)

The builder shall mark every `code-call` edge that satisfies none of REQ-GEC-002,
REQ-GEC-003, or REQ-GEC-012 as `inferred` with confidence 0.85.

### REQ-GEC-012 — Same-package promotion (Event-driven)

When a cross-file `code-call` edge's callee name is declared in a file whose
directory equals the caller file's directory — in Go, same directory equals same
package by construction — the builder shall mark the edge `intra-package` with
confidence 0.95. The tier is language-neutral (directory equality needs no
import parsing) and sits below import evidence: co-location is a by-construction
guarantee for Go packages, but it is evidence of proximity, not of a declared
dependency.

### REQ-GEC-005 — Determinism (Event-driven)

When the graph is built twice over the same tree, the builder shall produce
byte-identical `edges.jsonl` output (existing contract: sorted by `EdgeLess`, no
timestamps, atomic write — `internal/graph/graph.go` package doc).

### REQ-GEC-006 — Doc-edge additivity (Ubiquitous)

The builder shall not add `resolution` or `confidence` fields to any doc-derived
edge (`import`, `mx-spec`, `spec-depends`, `report-milestone`, `milestone-card`) or
to `code-import` edges; the JSONL serialization of those edges shall remain
byte-identical to the pre-change artifact on the same tree.

### REQ-GEC-007 — Grade axis preservation (Ubiquitous)

The per-language grade matrix shall remain the sole capability axis: this SPEC shall
not change any `Grade` value, `astx.GradeFor`'s mapping, `GradeMatrix`, or
`ValidateGradeMatrix`, and shall not assign `GradeFull` to any language.

### REQ-GEC-008 — MCP consumer exposure (Event-driven)

When a `graph_find_code` or `graph_trace_calls` response includes a match derived
from a `code-call` edge, the handler shall expose that edge's confidence value in
the response.

### REQ-GEC-009 — Legacy artifact compatibility (Event-driven)

When `edges.jsonl` lacks `resolution`/`confidence` fields (a pre-upgrade artifact),
the graph consumers shall load and serve it without error, treating absent
confidence as unknown rather than failing.

### REQ-GEC-010 — CGO-disabled behavior (Capability gate)

Where the build is CGO-disabled, the builder shall behave exactly as today: the astx
extractor reports unsupported, zero code-derived edges are produced, and the
confidence layer is inert — no new error surface, no behavioral change.

### REQ-GEC-011 — Traversal semantics unchanged (State-driven)

While a caller traverses the code-call layer via `FindCallers`, `BlastRadius`,
`TraceCalls`, or `FindCode` name matching, the consumer shall return the same
edge set it returned before this SPEC — confidence annotates edges, it never
filters, ranks, or reorders results in this delivery.

## §D. Acceptance Criteria — REQ→AC mapping

acceptance.md is the SINGLE SOURCE OF TRUTH for the AC set (Tier M): its §D AC
matrix carries the full Given-When-Then scenarios and verification commands.
This section carries only the traceability mapping.

| REQ | Verified by |
|---|---|
| REQ-GEC-001 | AC-GEC-002 |
| REQ-GEC-002 | AC-GEC-003 |
| REQ-GEC-003 | AC-GEC-004 |
| REQ-GEC-004 | AC-GEC-005, AC-GEC-002 (domain closure) |
| REQ-GEC-005 | AC-GEC-001, AC-GEC-011 |
| REQ-GEC-006 | AC-GEC-006 |
| REQ-GEC-007 | AC-GEC-010 |
| REQ-GEC-008 | AC-GEC-007 |
| REQ-GEC-009 | AC-GEC-008 |
| REQ-GEC-010 | AC-GEC-009 |
| REQ-GEC-011 | AC-GEC-010 |
| REQ-GEC-012 | AC-GEC-012 |

## §E. Constraints

- Determinism is the load-bearing contract of the artifact (git-diffable JSONL);
  every derivation input must be tree content — no wall-clock, no map-iteration
  order, no parallelism-dependent ordering in the confidence pass.
- The confidence join consumes only data `symbol.Extract` already walks (call sites,
  import sites, function ranges); adding a second parse pass or a new extractor
  dependency is out of scope.
- The artifact is untracked/derived: schema "backfill" is a rebuild (`moai graph
  build`), never an in-place migration tool.
- Repo conventions: `t.TempDir()` isolation for all test fixtures; no `t.Setenv`
  with OTEL vars; Go error wrapping `fmt.Errorf("...: %w", err)`; English comments.

## §F. Out of Scope

### Out of Scope — tag-kind MX edges and graph-backed fan-in

- `mx-debt`/`mx-anchor` tag-kind edges and replacing the grep-based fan-in
  approximation in the mx validator — separate card t412 / SPEC-MX-TAG-EDGES-001.

### Out of Scope — graph report tooling

- `shortest_path` MCP tool, `moai graph report`, shrink guard, deferred
  SessionStart background refresh — separate card t413 / SPEC-GRAPH-REPORT-001.

### Out of Scope — resolution capability upgrades

- Assigning `GradeFull` / implementing scope-aware resolution in `internal/navigator/astx`
  — this SPEC is explicitly the pragmatic BYPASS of that gap, not its implementation.
- Non-Go specifier→declaring-directory resolution: mapping raw import specifiers
  (python `from x import y`, TS/JS module specifiers, java/rust `use`/`import`
  paths) to the directory that declares a callee name. Not derivable with
  today's `ImportSite.Module` (specifier as written) + `localizeModule`
  (go.mod prefix only); recorded as CANDIDATE FUTURE WORK — a follow-up SPEC
  could add per-language specifier resolution and widen REQ-GEC-003's domain.
- Any Leiden/community clustering, visualization, Mermaid callflow, PreToolUse
  graph induction.

### Out of Scope — confidence-driven behavior

- Filtering, ranking, scoring, or thresholding query results by confidence; any
  consumer that treats `inferred` edges differently from `extracted` ones. This
  delivery annotates only.

## §G. Cross-References

- Evidence input: `.moai/reports/graphify-codegraph-analysis-20260901.md` (P1-1;
  §1.2 extraction/promotion pattern).
- Sibling cards: SPEC-MX-TAG-EDGES-001 (t412), SPEC-GRAPH-REPORT-001 (t413).
- Predecessor layer: SPEC-V3R6-GRAPH-FRESHNESS-001 (edge model, freshness gate,
  provenance sidecar — unchanged by this SPEC).
