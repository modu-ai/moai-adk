---
id: SPEC-NAVIGATOR-SYNC-001
title: "Navigator Sync (BAS M0) — SSOT binding-token trio + graph-join schema layer uniting the 3 Navigator chains"
version: "0.1.0"
status: in-progress
created: 2026-08-05
updated: 2026-08-05
author: manager-spec
priority: P1
phase: "v3.3 target"
module: navigator-sync
lifecycle: spec-anchored
tier: L
era: V3R6
tags: "navigator, sync, graph-join, ssot, binding-token, blueprint, bas-epic"
related_specs: [SPEC-PROJECT-NAVIGATOR-001, SPEC-PROJECT-NAVIGATOR-002, SPEC-PROJECT-NAVIGATOR-003, SPEC-LSEL-LOCAL-EVOLUTION-001]
---

# SPEC-NAVIGATOR-SYNC-001 — Navigator Sync (BAS Epic M0)

## HISTORY

- 2026-08-05 (initial draft) — BAS (Blueprint-Anchored Synchronization) Epic M0, plan-phase. The design is authored in `.moai/reports/navigator-redesign-bas-20260805.{md,html}` (session 3db3f943). M0 is the Epic's critical-path entry: the SSOT binding-token trio + the graph-join schema layer that unifies the three existing one-shot Navigator chains (001 regen, 002 audit, 003 enrich — all `status: completed`) into a single addressable graph. M1–M5 are follow-on SPECs, out of scope here. Boundary decision: this is a NEW Epic (`SPEC-NAVIGATOR-SYNC-NNN`), not a continuation of `SPEC-PROJECT-NAVIGATOR-NNN` — the predecessors are `completed` and referenced via `related_specs` (non-blocking), not `depends_on`. Design-report precision correction applied: the lean twin claims `@MX:SPEC` is "absorbed" (implying dormant), but `internal/mx/scanner.go:139-155` + `spec_association.go:17-20` already consume it as a third SPEC-association source — M0 therefore BRIDGES (not absorbs) the existing mx-scanner association into the Navigator graph, and does NOT modify `internal/mx/`.

## §A. User Story

**As a** Claude Code session (and the human orchestrator) orienting in the moai-adk-go repo,
**I want** a single addressable graph that joins design decisions, SPEC IDs, and code symbols via stable binding tokens,
**so that** a code edit to a symbol, a design-decision change, and a SPEC status update are each traceable to the other two surfaces WITHOUT regenerating the whole Navigator from scratch.

The integration layer sits ON TOP of the three existing Navigator chains (001/002/003). It consumes their outputs. It does not rewrite them. The binding-token trio is the join key that makes the graph addressable.

## §B. Scope

### §B.1 In Scope — M0 only

1. **SSOT binding-token trio** — three addressable token families joined into one graph:
   - `@NAV:DEC-<id>` — design-decision token, authored into `.moai/project/{product,structure,tech}.md` and ADRs (NEW token; M0 adds the scanner).
   - `@MX:SPEC:<id>` — code→SPEC back-pointer (EXISTING token; already consumed by `internal/mx/scanner.go` + `spec_association.go` as a third SPEC-association source; M0 reuses that association output — it does NOT re-scan code for `@MX:SPEC`).
   - `@NAV:SYM:<symbol>` — code-symbol token, authored into code comments or design docs to point at a named symbol (NEW token; M0 adds the scanner).
2. **Graph-join schema layer** — a single artifact `nav-graph.json` under `.moai/project/navigator/` whose nodes are entities (Decision / Spec / Symbol) and whose edges are the three token families. Produced by joining:
   - 001's `capability-map.md` (header-driven),
   - 003's `capability-symbols.json` (symbol enriched rows),
   - 002's `audit-report.json` (drift findings),
   - the mx-scanner's `spec_association` output (already aggregates `@MX:SPEC` + path-based + body-based sources).
3. **Provenance + atomic-write + fail-open capability gate** — inherited from 003's `navigator_enrich.go:128` `atomicWrite` + 003's `enrich.go:21` `Provenance` model + 003's REQ-NT-002 capability gate.
4. **Template-First documentation** — the two new tokens (`@NAV:DEC`, `@NAV:SYM`) are author-facing and MUST be documented under `internal/template/templates/` first (CLAUDE.local.md §2). M0 documents the tokens; it does not require template users to author them (adoption is gradual).

### §B.2 Carry-forward non-overlap invariants (load-bearing)

The three predecessor SPECs codified non-overlap contracts that keep each chain scoped. M0 integrates ABOVE the three chains; it does NOT relax these contracts:

- 001 REQ-PN-016 — Navigator ↔ LSEL non-overlap.
- 002 REQ-NA-011 — Audit ↔ LSEL + Audit ↔ 003 non-overlap.
- 003 REQ-NT-018 / REQ-NT-019 — Enricher ↔ LSEL + Enricher ↔ 002 audit non-overlap.

M0 carries these forward as REQ-NS-013 / REQ-NS-015 / REQ-NS-016, extended to the new integration-layer write surface.

## §C. Requirements (GEARS notation)

### §C.1 SSOT binding-token trio

#### REQ-NS-001 (Ubiquitous — token trio recognition)

The BAS integration layer SHALL recognize exactly three SSOT binding-token families: `@NAV:DEC-<id>` (design decision), `@MX:SPEC:<id>` (code→SPEC, existing), and `@NAV:SYM:<symbol>` (code symbol), each emitted as an addressable edge in the unified graph.

#### REQ-NS-002 (Ubiquitous — per-token binding record)

For every recognized token occurrence, the integration layer SHALL emit a binding record carrying five fields: `token_family` ∈ `{NAV:DEC, MX:SPEC, NAV:SYM}`, `identifier` (the `<id>` or `<symbol>`), `source_path` (absolute), `line_number` (1-indexed), and `commit_sha` (the git baseline provenance).

#### REQ-NS-003 (Ubiquitous — `@NAV:DEC` scanner, design docs)

The integration layer SHALL scan `.moai/project/{product,structure,tech}.md` and `.moai/docs/**/*.md` (read-only) for `@NAV:DEC-<id>` occurrences and emit one binding record per occurrence. The `<id>` grammar SHALL match `[A-Z][A-Z0-9-]*` (uppercase, consistent with SPEC-ID domain tokens).

#### REQ-NS-004 (Ubiquitous — `@NAV:SYM` scanner, code + design)

The integration layer SHALL scan the project source tree (Go `*.go` files excluding `*_test.go` and vendored paths, plus `.moai/project/**/*.md` and `.moai/docs/**/*.md`) for `@NAV:SYM:<symbol>` occurrences and emit one binding record per occurrence. The `<symbol>` grammar SHALL match `[A-Za-z_][A-Za-z0-9_.]*` (identifier-shaped, language-neutral).

#### REQ-NS-005 (Ubiquitous — `@MX:SPEC` consumption via existing mx-scanner, no re-scan)

The integration layer SHALL obtain `@MX:SPEC:<id>` associations by CONSUMING the existing `internal/mx/spec_association.go` `SpecAssociator` output (which already aggregates path-based + body-based + sub-line `@MX:SPEC` sources); it SHALL NOT re-scan the code tree for `@MX:SPEC` tokens and SHALL NOT modify `internal/mx/`.

### §C.2 Graph-join schema layer

#### REQ-NS-006 (Ubiquitous — single graph artifact)

The integration layer SHALL emit a single graph artifact at `.moai/project/navigator/nav-graph.json` whose top-level shape is `{ "provenance": {...}, "nodes": [...], "edges": [...] }`, produced by joining the three chain outputs with the binding-record sets from REQ-NS-003/004/005.

#### REQ-NS-007 (Ubiquitous — node set)

The graph's node set SHALL contain exactly three entity types: `decision` (one per unique `@NAV:DEC-<id>`), `spec` (one per SPEC ID surfaced by the mx-scanner association OR the capability-map), and `symbol` (one per unique symbol surfaced by 003's `capability-symbols.json` PrimarySymbols ∪ `@NAV:SYM:<symbol>` occurrences). Each node carries `entity_type`, `identifier`, and `display_name`.

#### REQ-NS-008 (Ubiquitous — edge set)

The graph's edge set SHALL contain edges typed by the three token families: `dec-edge` (Decision↔Spec or Decision↔Symbol, sourced from `@NAV:DEC`), `spec-edge` (Code↔Spec, sourced from the mx-scanner association), and `sym-edge` (Symbol↔Symbol or Symbol↔Doc, sourced from `@NAV:SYM`). Each edge carries `edge_type`, `source_node`, `target_node`, `source_path`, and `line_number`.

#### REQ-NS-009 (Ubiquitous — Provenance block, git baseline)

The `nav-graph.json` artifact SHALL carry a `provenance` block with `extract_commit_sha` (`git rev-parse HEAD`) and `captured_at` (committer date of that SHA, `git log -1 --format=%cI`), using NO wall-clock timestamp, so two runs on the same HEAD produce byte-identical output. This carries forward 003's `enrich.go:21` `Provenance` contract.

#### REQ-NS-010 (Ubiquitous — atomic write with barrier test hook)

The integration layer SHALL write `nav-graph.json` via the atomic-rename pattern (write to `nav-graph.json.tmp`, then `os.Rename`) and SHALL honor the `NAVIGATOR_PRE_RENAME_BARRIER` environment variable as a synchronized test hook for the concurrency fixture, mirroring 003's `navigator_enrich.go:128` `atomicWrite`.

#### REQ-NS-011 (Capability gate — fail-open when inputs absent)

**Where** the project lacks `capability-map.md` (001's output), the integration layer SHALL fail-open: exit code 0, emit an info log to `.moai/logs/navigator-sync.log`, and write NO `nav-graph.json` output. This carries forward 003's REQ-NT-002 capability-gate contract.

### §C.3 Non-overlap invariants (integration layer surface)

#### REQ-NS-012 (Ubiquitous — consumer-only on the 3 existing chains)

The integration layer SHALL be strictly a CONSUMER of the three existing Navigator chains' outputs (`capability-map.md`, `capability-symbols.{md,json}`, `audit-report.{md,json}`); it SHALL NOT modify the chains' Go code, their output schemas, or their individual entry-point commands (`/moai project`, `/moai codemaps` enrich step, `navigator-audit.sh`).

#### REQ-NS-013 (Ubiquitous — LSEL non-overlap, carries forward REQ-PN-016 / REQ-NT-018)

The integration layer SHALL NOT read or write any LSEL surface (`.moai/lessons-inbox.jsonl`, `.moai/state/lsel/`, `memory/feedback_*.md`, `hns-lsel-*` skills, `.moai/state/lsel/proposals/`), carrying 001's REQ-PN-016 and 003's REQ-NT-018 forward to the integration-layer boundary.

#### REQ-NS-014 (Ubiquitous — write-surface isolation)

The integration layer's write surface SHALL be limited to exactly one path: `.moai/project/navigator/nav-graph.json` (plus its `.tmp` transient). It SHALL NOT write to `.moai/specs/`, `.moai/project/navigator/capability-map.md`, `.moai/project/codemaps/capability-symbols.{md,json}`, or `.moai/project/navigator/audit-report.{md,json}`.

#### REQ-NS-015 (Ubiquitous — 002 audit non-overlap, carries forward REQ-NT-019)

The integration layer SHALL NOT write to 002's audit output surface (`audit-report.{md,json}`); it READS `audit-report.json` (when present) as one of its join inputs, and the read is advisory-only (a missing or malformed audit-report does not block graph emission).

#### REQ-NS-016 (Ubiquitous — 003 enrich non-overlap)

The integration layer SHALL NOT write to 003's enriched-symbol output surface (`capability-symbols.{md,json}`); it READS `capability-symbols.json` (when present) as the symbol-node source-of-truth, and a missing or malformed file degrades the symbol node set to "symbols referenced only via `@NAV:SYM` tokens" (graceful degradation, not failure).

### §C.4 Verification + template-first

#### REQ-NS-017 (Event-detected — malformed-token diagnostic)

**When** the integration layer encounters a malformed token (`@NAV:DEC-` with empty id, `@NAV:SYM:` with empty symbol, or an `<id>` / `<symbol>` violating the grammars of REQ-NS-003 / REQ-NS-004), the scanner SHALL emit a diagnostic warning line to `.moai/logs/navigator-sync.log`, skip the record, and continue processing — the scanner SHALL NOT abort the graph build on a malformed token.

#### REQ-NS-018 (Capability gate — template-first documentation)

**Where** this SPEC ships author-facing tokens (`@NAV:DEC`, `@NAV:SYM`) that distributed template users would author into their own projects, the documentation for those tokens SHALL be authored under `internal/template/templates/.claude/rules/moai/workflow/` first, regenerated via `make build`, and verified by the `internal/template/internal_content_leak_test.go` neutrality guard (CLAUDE.local.md §2 + §25).

## §D. Constraints

- **Language**: Go (the integration layer is a Go binary/library under `internal/navigator/`).
- **Idempotence**: re-running on the same HEAD MUST produce byte-identical `nav-graph.json` (REQ-NS-009 + sorted map iteration).
- **Concurrency**: atomic-rename + barrier test hook (REQ-NS-010).
- **Determinism**: scan order is filepath-sorted; map iteration is sorted on output; no wall-clock.
- **Fail-open**: the integration layer never aborts `/moai project` or any caller; all errors are logged and the exit code is 0 (carries forward 003's fail-open contract).
- **Provenance alignment with `verification-claim-integrity.md`**: every claim the integration layer makes about a binding is traceable to a git baseline (REQ-NS-009). The `nav-graph.json` artifact is the Evidence; the binding record is the Claim.
- **Non-overlap**: REQ-NS-013 / REQ-NS-015 / REQ-NS-016 are inviolable.
- **No `@MX:SPEC` re-scan**: REQ-NS-005 forbids duplicating the mx-scanner's job.
- **Template-First**: REQ-NS-018 binds any documentation edit to the template-first cycle.

## §E. Out of Scope

### Out of Scope — BAS Epic M1–M5 (follow-on SPECs)

- `SPEC-NAVIGATOR-SYNC-002` (M1) — PostToolUse Detect hook (changed-path → affected-rows mapping via audit-matching reuse).
- `SPEC-NAVIGATOR-SYNC-003` (M2) — Route audit advisory → work-item promotion.
- `SPEC-NAVIGATOR-SYNC-004` (M3) — Fix AI-drafted incremental refresh (the most complex milestone).
- `SPEC-NAVIGATOR-SYNC-005` (M4) — 4-tier map (Tier 0 Contract / Tier 1 Blueprint / Tier 2 ADR / Tier 3 Symbol extension to SCIP).
- `SPEC-NAVIGATOR-SYNC-006` (M5) — Brownfield reverse-extraction (tessl `document --code` pattern).

### Out of Scope — reverse-direction writes

- M0 is read-only on the design docs and source tree; it does NOT write back `@NAV:DEC` / `@NAV:SYM` tokens into files. Reverse-direction writes (the "Fix" element of the Falconer trio) are M3.
- M0 does NOT propagate SPEC status changes from `spec.md` frontmatter back into design docs. That is M2 (Route) + M3 (Fix).

### Out of Scope — modification of the 3 existing chains or the mx-scanner

- M0 does NOT modify `internal/navigator/astx/`, `internal/cli/navigator_enrich.go`, `scripts/navigator-audit.sh`, `scripts/navigator-regen.sh`, or `internal/mx/`. These are CONSUMED (REQ-NS-005 / REQ-NS-012).

### Out of Scope — LSEL surfaces

- M0 strictly avoids LSEL surfaces (REQ-NS-013). The "harness self-evolution" concern is owned by SPEC-LSEL-LOCAL-EVOLUTION-001.

### Out of Scope — a new Go CLI subcommand visible at the top level

- M0's integration layer is invoked from the existing `/moai project` surface as a sibling step (the same way 003's `navigator-enrich` is wired). No new top-level `moai` subcommand; the join step is a Hidden cobra command (mirroring 003's `navigator-enrich` Hidden subcommand at `navigator_enrich.go:54`).

## §F. AC Matrix (summary — full detail in acceptance.md)

| AC-ID  | REQ-ID        | Summary                                                                                           |
|--------|---------------|---------------------------------------------------------------------------------------------------|
| AC-001 | REQ-NS-001    | `nav-graph.json` contains edges of exactly 3 families                                            |
| AC-002 | REQ-NS-002    | Each binding record carries the 5 required fields                                                 |
| AC-003 | REQ-NS-003    | `@NAV:DEC-` scanner emits records from design docs                                                |
| AC-004 | REQ-NS-004    | `@NAV:SYM:` scanner emits records from code + design                                              |
| AC-005 | REQ-NS-005    | `@MX:SPEC` associations sourced from `internal/mx.SpecAssociator`; no `internal/mx/` modification |
| AC-006 | REQ-NS-006    | `nav-graph.json` top-level shape is `{provenance, nodes, edges}`                                  |
| AC-007 | REQ-NS-007    | Node set has exactly 3 entity types                                                               |
| AC-008 | REQ-NS-008    | Edge set has exactly 3 edge types                                                                 |
| AC-009 | REQ-NS-009    | Provenance block has no wall-clock; byte-identical on re-run                                       |
| AC-010 | REQ-NS-010    | Atomic-rename + NAVIGATOR_PRE_RENAME_BARRIER test hook works                                      |
| AC-011 | REQ-NS-011    | Fail-open: no capability-map.md → exit 0, no output                                               |
| AC-012 | REQ-NS-012    | Consumer-only: 3 chain entry-points unchanged                                                     |
| AC-013 | REQ-NS-013    | Integration-layer grep touches no LSEL surface                                                    |
| AC-014 | REQ-NS-014    | Write surface limited to `nav-graph.json` + `.tmp`                                                |
| AC-015 | REQ-NS-015    | 002 audit write surface untouched                                                                 |
| AC-016 | REQ-NS-016    | 003 enrich write surface untouched                                                                |
| AC-017 | REQ-NS-017    | Malformed token → diagnostic + skip, no abort                                                     |
| AC-018 | REQ-NS-018    | Template-first: tokens documented under `internal/template/templates/` + neutrality guard green   |

## §G. Cross-References

- **Design report (SSOT for this SPEC's rationale)** — `.moai/reports/navigator-redesign-bas-20260805.{md,html}` (session 3db3f943).
- **SPEC-PROJECT-NAVIGATOR-001** (`status: completed`) — regen chain producing `capability-map.md`. M0 consumes its output. Non-overlap boundary REQ-PN-016 carried forward as REQ-NS-013.
- **SPEC-PROJECT-NAVIGATOR-002** (`status: completed`) — audit chain producing `audit-report.{md,json}`. M0 reads its JSON output as a join input. Non-overlap REQ-NA-011 + REQ-NT-019 carried forward as REQ-NS-015.
- **SPEC-PROJECT-NAVIGATOR-003** (`status: completed`) — enrich chain producing `capability-symbols.{md,json}`. M0 reuses its `atomicWrite` pattern (`navigator_enrich.go:128`), its `Provenance` model (`enrich.go:21`), its header-driven parser (`enrich.go:122`), and its fail-open capability gate (REQ-NT-002). Non-overlap REQ-NT-018/019 carried forward as REQ-NS-013/016.
- **SPEC-LSEL-LOCAL-EVOLUTION-001** (`status: completed`) — non-overlap boundary (REQ-NS-013).
- **`internal/mx/spec_association.go`** — the existing `SpecAssociator` (3-source: path-based + body-based + `@MX:SPEC` sub-line) whose output M0 consumes via REQ-NS-005. NOT modified by this SPEC.
- **`internal/cli/navigator_enrich.go`** — sibling Hidden cobra subcommand pattern; M0's join step mirrors its structure.
- **`.claude/rules/moai/core/verification-claim-integrity.md`** — the Provenance block (REQ-NS-009) is the operational mechanism that makes every binding record attributable to a git baseline.
- **CLAUDE.local.md §2 Template-First Rule + §25 Template Internal-Content Isolation** — bind REQ-NS-018.
