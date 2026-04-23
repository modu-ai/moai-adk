---
id: SPEC-V3R2-CON-001
title: "FROZEN/EVOLVABLE zone codification for core constitution"
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: Wave 4 SPEC Writer
priority: P0 Critical
phase: "v3.0.0 — Phase 1 — Constitution & Foundation"
module: "internal/constitution/, .claude/rules/moai/core/"
dependencies: []
related_gap: []
related_problem:
  - P-R02
  - P-R03
related_pattern:
  - S-4
  - X-1
related_principle:
  - P1
  - P2
  - P12
related_theme: "Layer 1: Constitution"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "v3r2, constitution, frozen-zone, evolvable-zone, zone-codification"
---

# SPEC-V3R2-CON-001: FROZEN/EVOLVABLE zone codification for core constitution

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-23 | Wave 4 SPEC Writer | Initial draft from master v3 §4 Layer 1, Wave 2 synthesis |

---

## 1. Goal (목적)

Generalize the FROZEN / EVOLVABLE zone model — currently codified only in the design subsystem constitution (`.claude/rules/moai/design/constitution.md` v3.3.0 §2) — to the **entire** moai-adk rule tree so that every HARD clause across core, agent-common-protocol, TRUST 5, SPEC format, @MX TAG protocol, 16-language neutrality, and Template-First discipline carries an explicit Zone (Frozen | Evolvable) assignment.

The existing design constitution proves the pattern works inside a vertical subsystem (design pipeline). v3 lifts the primitive up one layer: core constitution becomes the constitutional parent, design constitution becomes a nested subsystem whose FROZEN zone inherits core FROZEN.

This SPEC **codifies zones**; it does not **amend** any individual clause. The amendment protocol (who can evolve what, under which 5-layer safety gate) is scoped to SPEC-V3R2-CON-002.

## 2. Scope (범위)

### 2.1 In Scope

- Publish a single canonical "zone registry" file at `.claude/rules/moai/core/zone-registry.md` enumerating every HARD clause in the moai rule tree with its Zone, Rule ID (CONST-V3R2-NNN), source file+anchor, and human-readable clause text.
- Introduce typed primitives in `internal/constitution/zone.go`: `Zone` enum (Frozen, Evolvable), `Rule` struct with ID/Zone/File/Clause/CanaryGate fields.
- Scan and annotate the 4 load-bearing constitution files with inline zone markers: `CLAUDE.md`, `.claude/rules/moai/core/moai-constitution.md`, `.claude/rules/moai/core/agent-common-protocol.md`, `.claude/rules/moai/design/constitution.md` (already zoned in v3.3.0).
- Preserve §1.3 FROZEN list of master-v3 verbatim: SPEC+EARS format, TRUST 5, @MX TAG protocol, 16-language neutrality, Template-First discipline, AskUserQuestion-Only-for-Orchestrator, Claude Code as execution substrate.
- Provide a `moai constitution list` CLI diagnostic that prints the zone registry.
- Emit one `CONST-V3R2-NNN` ID per HARD clause so SPEC cross-references have stable anchors.

### 2.2 Out of Scope

- Amendment protocol and 5-layer safety gate (→ SPEC-V3R2-CON-002).
- Rule-tree consolidation and file merges (→ SPEC-V3R2-CON-003).
- Runtime enforcement of zone violations (FrozenGuard implementation belongs to SPEC-V3R2-CON-002).
- Changing the content of any HARD clause. This SPEC is an annotation pass.
- Design subsystem rule authoring (already FROZEN per design/constitution.md v3.3.0).

## 3. Environment (환경)

Current moai state (v2.13.2):

- `.claude/rules/moai/core/moai-constitution.md` (266 LOC) — contains HARD rules for parallel execution, response language, TRUST 5, agent common behaviors, worktree isolation. No Zone markers.
- `.claude/rules/moai/core/agent-common-protocol.md` (157 LOC) — contains HARD rules for User Interaction Boundary, Language Handling, Output Format, Background Agent Write Restriction. No Zone markers.
- `CLAUDE.md` (~860 LOC) — project execution directive with "HARD Rules" section. No explicit Zone field.
- `.claude/rules/moai/design/constitution.md` v3.3.0 — **already** uses `[FROZEN]` and `[EVOLVABLE]` inline markers (§2). This is the prototype.
- No Go-side type for zones exists today. `internal/config/types.go` loads configuration but does not model constitutional rules.

References to §1.3 master-v3 FROZEN list and existing design constitution §2 confirm the pattern is desired; SPEC P-R02 flags constitutional sprawl risk (~800 lines across four files).

## 4. Assumptions (가정)

- The existing 7 FROZEN invariants in master-v3 §1.3 are correct and require no renegotiation (renegotiation would be a CON-002 amendment event, not a CON-001 codification event).
- The design constitution v3.3.0 serves as a template for syntax: `[FROZEN]` and `[EVOLVABLE]` prefixes on list items, grouped in a Section 2 "Zones" subsection.
- Every current HARD rule either is FROZEN (constitutional invariant) or EVOLVABLE (heuristic refinable by graduation). There is no "tentative" tier at annotation time; tentative rules are captured via SPEC-V3R2-EXT-001 memory taxonomy (reference/project tier), not here.
- Rule IDs `CONST-V3R2-NNN` are unique across the registry; renumbering upon merge in CON-003 is a zero-downtime mapping.
- SPEC authors do not need to memorize rule IDs; the registry is the source of truth.

## 5. Requirements (EARS)

### 5.1 Ubiquitous

- REQ-CON-001-001: The system SHALL provide a single canonical zone registry at `.claude/rules/moai/core/zone-registry.md` enumerating every HARD clause in the moai rule tree.
- REQ-CON-001-002: Each entry in the zone registry SHALL contain fields: `id` (CONST-V3R2-NNN), `zone` (Frozen | Evolvable), `file` (rule file path), `anchor` (section or line reference), `clause` (verbatim HARD text), `canary_gate` (boolean — whether amendment requires shadow evaluation).
- REQ-CON-001-003: The system SHALL define a Go type `internal/constitution.Zone` with exactly two values: `ZoneFrozen` and `ZoneEvolvable`.
- REQ-CON-001-004: The system SHALL define a Go type `internal/constitution.Rule` with exported fields matching the registry entry schema (ID, Zone, File, Clause, CanaryGate).
- REQ-CON-001-005: The 7 master-v3 §1.3 FROZEN invariants (SPEC+EARS, TRUST 5, @MX TAG, 16-language neutrality, Template-First, AskUserQuestion monopoly, CC substrate) SHALL each appear in the zone registry with `zone: Frozen` and distinct `CONST-V3R2-NNN` IDs.
- REQ-CON-001-006: The zone registry SHALL be loaded by `moai constitution list` CLI subcommand which prints entries in a reviewable table.
- REQ-CON-001-007: Zone registry entries SHALL preserve the verbatim clause text from the source file; registry is a view, not a rewrite.

### 5.2 Event-driven

- REQ-CON-001-010: WHEN a new HARD rule is added to any file under `.claude/rules/moai/`, the system SHALL require a corresponding zone-registry entry before CI passes.
- REQ-CON-001-011: WHEN the zone registry is updated, the system SHALL recompute Rule IDs only for new entries and SHALL preserve existing IDs across subsequent edits.
- REQ-CON-001-012: WHEN `moai constitution list --zone frozen` is invoked, the system SHALL filter the registry to entries where `zone == ZoneFrozen`.

### 5.3 State-driven

- REQ-CON-001-020: WHILE the zone registry contains zero entries with `zone == ZoneFrozen`, the system SHALL treat the registry as invalid and emit a doctor-level warning (`moai doctor constitution`).
- REQ-CON-001-021: WHILE the design subsystem constitution file is present, the `moai-domain-*` agency-absorbed FROZEN clauses SHALL be mirrored into the core registry with file pointers to the design constitution.

### 5.4 Optional

- REQ-CON-001-030: WHERE the environment variable `MOAI_CONSTITUTION_STRICT=1` is set, the system SHALL fail startup if the zone registry cannot be parsed or contains duplicate IDs.

### 5.5 Complex

- REQ-CON-001-040: WHILE the zone registry is being loaded AND a rule file referenced by an entry is missing from disk, the loader SHALL emit a structured error identifying the missing file and mark the corresponding rule as `orphan: true` without halting registry initialization.
- REQ-CON-001-041: WHEN a SPEC document references a Rule ID that does not appear in the zone registry, the SPEC linter (SPEC-V3R2-SPC-003) SHALL raise a dangling-reference warning; this reinforces the registry as the single source of truth for constitutional citations.

## 6. Acceptance Criteria

- AC-CON-001-01: Given a fresh v3 install, When the user runs `moai constitution list`, Then the output contains at least 7 entries with `zone: Frozen` corresponding verbatim to master-v3 §1.3 invariants. (maps REQ-CON-001-005, REQ-CON-001-006)
- AC-CON-001-02: Given the zone registry exists, When `moai constitution list --zone frozen` is invoked, Then only Frozen-zone entries are printed. (maps REQ-CON-001-012)
- AC-CON-001-03: Given a developer edits `.claude/rules/moai/core/moai-constitution.md` to add a new `[HARD]` clause without updating `zone-registry.md`, When CI runs, Then the build fails with a message naming the missing registry entry. (maps REQ-CON-001-010)
- AC-CON-001-04: Given the Go package `internal/constitution`, When `zone.go` is imported, Then it exposes exactly two Zone values (`ZoneFrozen`, `ZoneEvolvable`) and a `Rule` struct whose exported fields match the registry schema. (maps REQ-CON-001-003, REQ-CON-001-004)
- AC-CON-001-05: Given a zone registry entry for CONST-V3R2-001 "SPEC+EARS FROZEN", When compared against `.claude/rules/moai/workflow/spec-workflow.md` anchor text, Then the `clause` field matches verbatim. (maps REQ-CON-001-007)
- AC-CON-001-06: Given `MOAI_CONSTITUTION_STRICT=1`, When the zone registry contains duplicate IDs, Then `moai doctor` exits with non-zero status. (maps REQ-CON-001-030)
- AC-CON-001-07: Given the zone registry references a non-existent rule file, When the loader is invoked, Then it emits a structured error naming the missing file and marks the entry as orphaned but does not panic. (maps REQ-CON-001-040)
- AC-CON-001-08: Given the design constitution v3.3.0 FROZEN list, When the core registry is loaded, Then design-subsystem FROZEN clauses are mirrored with pointers to `.claude/rules/moai/design/constitution.md`. (maps REQ-CON-001-021)
- AC-CON-001-09: Given a rule ID `CONST-V3R2-042` assigned in one registry build, When a subsequent build adds new entries, Then `CONST-V3R2-042` SHALL still refer to the same clause (ID stability). (maps REQ-CON-001-011)
- AC-CON-001-10: Given the `moai constitution list` output, When `zone-registry.md` is grep-searched for each printed Rule ID, Then 100% of printed IDs appear in the file. (maps REQ-CON-001-001, REQ-CON-001-006)
- AC-CON-001-11: Given a SPEC document with YAML `related_rule: [CONST-V3R2-999]` (non-existent), When SPEC-V3R2-SPC-003 linter runs, Then a dangling-reference warning is raised citing the unknown ID. (maps REQ-CON-001-041)
- AC-CON-001-12: Given master-v3 §1.3 "AskUserQuestion monopoly", When the zone registry is queried for that invariant, Then exactly one Frozen-zone entry is returned with `file: .claude/rules/moai/core/agent-common-protocol.md` and the verbatim clause. (maps REQ-CON-001-005, REQ-CON-001-007)

## 7. Constraints (제약)

- The 7 FROZEN invariants from master-v3 §1.3 MUST NOT be rephrased; registry entries preserve source clause text verbatim.
- Registry IDs follow `CONST-V3R2-NNN` (zero-padded 3 digits). First 50 IDs reserved for pre-existing clauses discovered during annotation; 51-99 reserved for design-subsystem mirrors; 100+ for new post-annotation additions.
- `zone-registry.md` is authored by hand, not auto-generated. Auto-generation is an explicit non-goal because the file is a human-readable artifact.
- The Go type `internal/constitution.Zone` uses `uint8` underlying representation to match the existing design-system `Zone` prototype (master-v3 §4 Layer 1 code listing).
- Registry loader performance budget: <10ms for 200 entries on cold start.
- Binary size impact: adding `internal/constitution/` MUST NOT grow `bin/moai` by more than 50KB.
- Template-First discipline: every new file under `.claude/rules/moai/core/` has a twin in `internal/template/templates/.claude/rules/moai/core/`. (aligns with existing CLAUDE.local.md §2)

## 8. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Annotation ambiguity — is a rule FROZEN or EVOLVABLE? | HIGH | MEDIUM | Default to FROZEN when unclear; CON-002 amendment protocol can demote to EVOLVABLE with human approval + canary |
| Registry drift against source files | MEDIUM | HIGH | CI check (REQ-CON-001-010) rejects edits without registry updates; `moai doctor constitution` re-verifies in field |
| ID renumbering breaks SPEC cross-references | LOW | HIGH | REQ-CON-001-011 pins ID stability; append-only IDs, never recycle |
| Design subsystem mirror creates redundancy with design constitution | LOW | LOW | Mirror entries reference the design file; no content duplication |
| Large registry (>200 entries) becomes unreadable | LOW | LOW | `moai constitution list --zone frozen --file X` filters; markdown table group-by-file |

## 9. Dependencies

### 9.1 Blocked by

- None. CON-001 is the foundation of Phase 1.

### 9.2 Blocks

- SPEC-V3R2-CON-002 (amendment protocol needs zone registry to know what it is amending).
- SPEC-V3R2-CON-003 (consolidation pass moves rule files and updates registry pointers).
- SPEC-V3R2-SPC-003 (SPEC linter checks `related_rule` against registry per REQ-CON-001-041).
- SPEC-V3R2-RT-005 (multi-layer settings provenance uses zone IDs for rule-source attribution).

### 9.3 Related

- `.claude/rules/moai/design/constitution.md` v3.3.0 §2 — zone model prototype.
- master-v3 §1.3 — canonical FROZEN invariant list.
- master-v3 §4 Layer 1 — Go primitive sketch.

## 10. Traceability

- Theme: Layer 1 Constitution (master-v3 §4 "Constitution").
- Principles: P1 SPEC as Constitutional Contract (master-v3 §3, design-principles.md §P1); P2 Constitutional Governance with FROZEN/EVOLVABLE zones (master-v3 §3, design-principles.md §P12); P12 File-First Primitives (design-principles.md §P11 — registry is a markdown file).
- Problems: P-R02 Constitutional sprawl (problem-catalog.md — consolidation pass); P-R03 CLAUDE.md/common-protocol duplication (problem-catalog.md — registry deduplicates).
- Patterns: S-4 FROZEN + Graduation (pattern-library.md §S-4); X-1 Markdown + YAML Frontmatter (pattern-library.md §X-1 — registry is markdown).
- Wave 1 sources: R1 §18 Constitutional AI (declarative governance); R3 §4 Adoption Candidate 7 (typed taxonomy formalization).
- Wave 2 sources: design-principles.md §P12 (zones as governance); problem-catalog.md Cluster 6 adjacency.
