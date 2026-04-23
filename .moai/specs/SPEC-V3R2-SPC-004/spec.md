---
id: SPEC-V3R2-SPC-004
title: "@MX anchor resolver (query by SPEC ID, fan_in, danger category)"
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: Wave 4 SPEC Writer
priority: P2 Medium
phase: "v3.0.0 — Phase 7 — Extension"
module: "internal/mx/resolver.go, cmd/moai/mx.go"
dependencies:
  - SPEC-V3R2-SPC-002
related_gap: []
related_problem: []
related_pattern:
  - T-1
related_principle:
  - P2
related_theme: "Layer 2: SPEC & TAG"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "v3r2, mx, resolver, aci, fan-in, anchor"
---

# SPEC-V3R2-SPC-004: @MX anchor resolver

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-23 | Wave 4 SPEC Writer | Initial draft as T-1 ACI command |

---

## 1. Goal (목적)

Expose a structured query API over the @MX TAG sidecar index (SPEC-V3R2-SPC-002) as an ACI-shaped tool: `moai mx query`. Callers can filter tags by SPEC ID, by Kind (NOTE/WARN/ANCHOR/TODO/LEGACY), by fan_in threshold (high-callsite anchors for invariant enforcement), by danger category (WARN-tagged items whose REASON sub-line matches safety-critical patterns), and by file path prefix.

This SPEC is the companion resolver for SPC-002. Together they form the T-1 ACI priority-1 pattern (pattern-library.md §T-1 priority 1 of Top-10): raw `Bash` + `grep` is discouraged for common MX moves; named query commands with structured, paginated, LM-optimized responses replace it.

Output format is JSON (primary) for tool consumption and a human-readable table (secondary) for terminal inspection. This same API is consumed by codemaps generation, evaluator-active scoring hints, and the /moai review phase.

## 2. Scope (범위)

### 2.1 In Scope

- `moai mx query` subcommand with filters: `--spec <SPEC-ID>`, `--kind <note|warn|anchor|todo|legacy>`, `--fan-in-min <N>`, `--danger <category>`, `--file-prefix <path>`, `--since <time>`, `--limit <N>`.
- Go resolver API in `internal/mx/resolver.go`: `Resolver.Resolve(query) []Tag`, `Resolver.ResolveAnchor(anchorID) []Callsite`.
- Fan-in analysis: for an @MX:ANCHOR with a stable ID, count code-reference sites (callsites, imports, ast-grep matches) across the project using LSP `find-references` (SPEC-LSP-CORE-002) where available; fallback to textual search.
- Danger-category mapping: `.moai/config/sections/mx.yaml` `danger_categories:` provides a pattern→category map for WARN REASON text (e.g., "goroutine leak", "unbounded channel", "missing Close", "hardcoded credential" → categories concurrency, resource-leak, cleanup, security).
- JSON output schema: an array of Tag-plus-fanIn records with stable fields for tool consumption.
- Pagination via `--limit` and `--offset`; cursor-based pagination for large result sets.
- SPEC association: a tag is associated with SPEC X if (a) its file path falls under a module listed in SPEC X's `module:` frontmatter OR (b) the tag body explicitly references a SPEC ID (e.g., `@MX:ANCHOR for SPEC-AUTH-001`).

### 2.2 Out of Scope

- Sidecar index maintenance (→ SPEC-V3R2-SPC-002).
- Hook JSON protocol (→ SPEC-V3R2-RT-001).
- Anchor auto-promotion of high-fan-in functions (→ design by SPEC-V3R2-HRN-003 or evaluator tooling).
- Tag authoring (agents add @MX autonomously per mx-tag-protocol.md).
- Multi-project queries; resolver is scoped to the current moai project.

## 3. Environment (환경)

Current moai state (v2.13.2):

- @MX TAGs today are consumed by: `/moai mx` subcommand, agents via on-demand grep, and ad-hoc scripts. No structured resolver API.
- `internal/lsp/` provides `find-references` (via powernap) for 16 languages — reusable for fan-in.
- SPEC frontmatter already declares `module:` (e.g., `internal/hook/`); this establishes the file-path → SPEC mapping when SPEC body does not explicitly reference a tag.
- Performance: today `/moai codemaps` and `/moai mx` both re-scan; a shared resolver pattern eliminates redundant scans once SPC-002 sidecar exists.

References: pattern-library.md §T-1 priority 1 ("strongest single leverage pattern"); master-v3 §7.3 ACI command list includes `moai_locate_mx_anchor`.

## 4. Assumptions (가정)

- SPEC-V3R2-SPC-002 ships the sidecar; SPC-004 reads it.
- LSP `find-references` is available for languages that have running language-servers (Go/Python/TypeScript/etc.); fallback to grep for languages without LSP.
- `danger_categories:` mapping is authored by the user in `mx.yaml`; defaults shipped in template cover common safety concerns.
- Callsite count for ANCHOR fan_in is precise within 10% of ground truth; occasional false positives/negatives acceptable.
- Query response is bounded by `--limit` (default 100) to prevent accidental full-project dumps.

## 5. Requirements (EARS)

### 5.1 Ubiquitous

- REQ-SPC-004-001: The system SHALL provide `moai mx query` subcommand supporting filters: `--spec`, `--kind`, `--fan-in-min`, `--danger`, `--file-prefix`, `--since`, `--limit`, `--offset`.
- REQ-SPC-004-002: The system SHALL provide Go API `internal/mx.Resolver.Resolve(query Query) (result []Tag, err error)` returning the same result set as the CLI.
- REQ-SPC-004-003: The resolver SHALL compute fan_in for `@MX:ANCHOR` tags using LSP `find-references` when an LSP server is available for the tag's file language; otherwise it SHALL fall back to ast-grep or textual search.
- REQ-SPC-004-004: The default output format SHALL be JSON (`--format json`); human-readable table format SHALL be available via `--format table`.
- REQ-SPC-004-005: JSON output entries SHALL contain fields `kind`, `file`, `line`, `body`, `reason`, `anchor_id`, `created_by`, `last_seen_at`, `fan_in` (ANCHOR only), `danger_category` (WARN only), `spec_associations` (list of SPEC IDs).
- REQ-SPC-004-006: The resolver SHALL associate a tag with a SPEC when (a) the tag's file path falls under a module listed in the SPEC's `module:` frontmatter, OR (b) the tag body explicitly references `SPEC-[A-Z0-9-]+`.
- REQ-SPC-004-007: The resolver SHALL apply pagination: `--limit` (default 100, max 10000) and `--offset` (default 0).

### 5.2 Event-driven

- REQ-SPC-004-010: WHEN `moai mx query --spec SPEC-V3R2-X-001 --kind anchor` is invoked, the resolver SHALL return only ANCHOR tags associated with SPEC-V3R2-X-001.
- REQ-SPC-004-011: WHEN `moai mx query --fan-in-min 3 --kind anchor` is invoked, the resolver SHALL compute fan_in for every ANCHOR and return only those with fan_in ≥ 3.
- REQ-SPC-004-012: WHEN `moai mx query --danger concurrency` is invoked, the resolver SHALL match WARN tags whose REASON text matches patterns mapped to category "concurrency" in `mx.yaml` `danger_categories:`.
- REQ-SPC-004-013: WHEN the sidecar index is absent or unreadable, the resolver SHALL emit `SidecarUnavailable` error suggesting `/moai mx --full` to rebuild.

### 5.3 State-driven

- REQ-SPC-004-020: WHILE `--fan-in-min` is set AND no LSP server is running for a candidate file's language, the resolver SHALL use the fallback (ast-grep or grep) and annotate the result entry with `fan_in_method: "textual"`.
- REQ-SPC-004-021: WHILE the project contains 10,000+ tags, the resolver SHALL apply `--limit` defaults automatically and emit a `TruncationNotice` in the output header.

### 5.4 Optional

- REQ-SPC-004-030: WHERE `MOAI_MX_QUERY_STRICT=1` is set, the resolver SHALL fail (not fall back to textual search) when LSP is required but unavailable.
- REQ-SPC-004-031: WHERE `moai mx query --format markdown` is used, the output SHALL be a human-readable markdown table suitable for inclusion in reports.

### 5.5 Complex

- REQ-SPC-004-040: WHILE computing fan_in for an ANCHOR AND the anchor_id appears inside a test fixture file (per `.moai/config/sections/mx.yaml` `test_paths:` list), THEN the test fixture references SHALL be excluded from the fan_in count by default; an opt-in flag `--include-tests` overrides.
- REQ-SPC-004-041: IF a query matches zero tags, THEN the resolver SHALL emit an empty JSON array (not null) and exit with status 0; ELSE if the query's filter values are syntactically invalid, exit status is 2 with `InvalidQuery` error.
- REQ-SPC-004-042: WHEN multiple filters are combined (e.g., `--spec X --kind anchor --fan-in-min 3`), THEN filters SHALL be AND-composed; only tags matching all filters are returned.

## 6. Acceptance Criteria

- AC-SPC-004-01: Given a sidecar with 20 tags across 2 SPECs, When `moai mx query --spec SPEC-X-001 --kind anchor` runs, Then only ANCHOR tags associated with SPEC-X-001 appear in output. (maps REQ-SPC-004-001, REQ-SPC-004-006, REQ-SPC-004-010)
- AC-SPC-004-02: Given 5 ANCHOR tags with fan_in values 1, 2, 3, 5, 10, When `moai mx query --fan-in-min 3 --kind anchor` runs, Then exactly the last three are returned. (maps REQ-SPC-004-011)
- AC-SPC-004-03: Given a WARN tag with REASON "goroutine leak on panic" and `mx.yaml` maps "goroutine leak" → concurrency, When `moai mx query --danger concurrency` runs, Then that WARN tag appears in output. (maps REQ-SPC-004-012)
- AC-SPC-004-04: Given sidecar is missing, When `moai mx query` runs, Then stderr contains `SidecarUnavailable` and suggests `/moai mx --full`. (maps REQ-SPC-004-013)
- AC-SPC-004-05: Given `--format json` (default), When the query runs, Then stdout is a valid JSON array with entries matching the REQ-SPC-004-005 schema. (maps REQ-SPC-004-004, REQ-SPC-004-005)
- AC-SPC-004-06: Given `--format table`, When invoked, Then stdout is human-readable columnar output. (maps REQ-SPC-004-004)
- AC-SPC-004-07: Given no LSP is running for Python, When `moai mx query --fan-in-min 2 --file-prefix internal/py/` runs (Python-only), Then results are returned with `fan_in_method: "textual"` annotation. (maps REQ-SPC-004-020)
- AC-SPC-004-08: Given 10,000 tags in sidecar and no explicit `--limit`, When the query runs, Then response contains at most 100 entries and output header contains `TruncationNotice`. (maps REQ-SPC-004-021, REQ-SPC-004-007)
- AC-SPC-004-09: Given `MOAI_MX_QUERY_STRICT=1` AND no LSP for target language, When `moai mx query --fan-in-min 3` runs, Then it exits non-zero with `LSPRequired` error. (maps REQ-SPC-004-030)
- AC-SPC-004-10: Given `--format markdown`, When invoked, Then output is a markdown table. (maps REQ-SPC-004-031)
- AC-SPC-004-11: Given an ANCHOR inside `tests/fixtures/mock_handler_test.go`, When fan_in is computed with no `--include-tests`, Then references from other test fixtures are excluded. (maps REQ-SPC-004-040)
- AC-SPC-004-12: Given a query matching zero tags, When it runs, Then stdout is `[]` and exit status is 0. (maps REQ-SPC-004-041)
- AC-SPC-004-13: Given invalid filter `--kind nonexistent`, When it runs, Then exit status is 2 and stderr contains `InvalidQuery`. (maps REQ-SPC-004-041)
- AC-SPC-004-14: Given combined filters `--spec X --kind anchor --fan-in-min 3`, When the query runs, Then only anchors that (a) are associated with SPEC X AND (b) have fan_in ≥ 3 are returned. (maps REQ-SPC-004-042)
- AC-SPC-004-15: Given an @MX body "ANCHOR for SPEC-AUTH-001 handler", When the resolver associates it, Then the result's `spec_associations` contains "SPEC-AUTH-001". (maps REQ-SPC-004-006)

## 7. Constraints (제약)

- Resolver performance: <100ms for a sidecar with 1,000 tags and no fan_in computation; <2s with fan_in computation via LSP for 50 anchors.
- CLI response size cap: default 10MB for JSON output (truncation warning beyond).
- No mutations of sidecar from resolver path; read-only.
- LSP dependency via powernap is optional; textual fallback always available.
- 16-language neutrality: danger_categories default patterns cover generic safety concerns; per-language extensions via `mx.yaml`.

## 8. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Fan_in false positives (textual matches in strings/comments) | MEDIUM | MEDIUM | Prefer LSP; annotate `fan_in_method: textual` when fallback used |
| `danger_categories:` patterns over-match | MEDIUM | LOW | User-customizable; defaults conservative; `--danger` optional |
| SPEC association heuristic misses tags in unrelated files | LOW | LOW | Secondary matching via explicit body text; `/moai mx query --spec none` lists unassociated tags |
| Output size blows up on `--fan-in-min 0` + no `--limit` | LOW | LOW | Default `--limit 100`; truncation notice |
| Resolver usage bypasses SPC-002 freshness | LOW | MEDIUM | Sidecar age check: warn if sidecar >24h old |

## 9. Dependencies

### 9.1 Blocked by

- SPEC-V3R2-SPC-002 (resolver reads the sidecar index).

### 9.2 Blocks

- Codemaps generation tools (can query resolver for per-module anchor counts).
- Evaluator-active scoring (may use danger_category distribution as a harness signal per SPEC-V3R2-HRN-003).

### 9.3 Related

- SPEC-V3R2-RT-001 (hook JSON protocol — unrelated to resolver, but SPC-002 depends on it).
- SPEC-LSP-CORE-002 (LSP client for fan-in via `find-references`).
- pattern-library.md §T-1 priority 1.

## 10. Traceability

- Theme: Layer 2 SPEC & TAG (master-v3 §4).
- Principles: P2 Interface Design Over Tool Count / ACI (design-principles.md §P2 — named query command, LM-optimized JSON response).
- Patterns: T-1 ACI (pattern-library.md §T-1 priority 1 — "6 commands" including `moai_locate_mx_anchor`).
- Wave 1 sources: R1 §11 SWE-agent ACI (8× improvement), R2 §8 SWE-agent deep analysis.
- Wave 2 sources: pattern-library.md §T-1 priority 1 ("strongest single leverage pattern"); master-v3 §5.3 ACI command list.
