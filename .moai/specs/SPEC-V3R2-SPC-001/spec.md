---
id: SPEC-V3R2-SPC-001
title: "EARS + hierarchical acceptance criteria"
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: Wave 4 SPEC Writer
priority: P0 Critical
phase: "v3.0.0 — Phase 5 — Harness + Evaluator"
module: "internal/spec/, .moai/specs/, .claude/rules/moai/workflow/spec-workflow.md"
dependencies:
  - SPEC-V3R2-CON-001
  - SPEC-V3R2-CON-002
related_gap: []
related_problem: []
related_pattern:
  - E-1
  - X-1
related_principle:
  - P1
  - P4
related_theme: "Layer 2: SPEC & TAG"
breaking: true
bc_id: [BC-V3R2-011]
lifecycle: spec-anchored
tags: "v3r2, spec, ears, hierarchical-acceptance, agent-as-judge"
---

# SPEC-V3R2-SPC-001: EARS + hierarchical acceptance criteria

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-23 | Wave 4 SPEC Writer | Initial draft from Agent-as-a-Judge R1 §9 hierarchical shape |

---

## 1. Goal (목적)

Upgrade SPEC acceptance-criteria format from a flat list of Given/When/Then statements to a **hierarchical** tree where a parent criterion can contain nested sub-criteria. The motivating shape is the Agent-as-a-Judge paper's DevAI benchmark (R1 §9), which maps 55 tasks onto **365 sub-requirements** — each top-level requirement splits into testable sub-criteria that an Agent judge can score per-node.

Current moai acceptance criteria (`AC-XXX-01: Given ... When ... Then ...`) is flat. Flat criteria force either over-coarse acceptance (one criterion covers multiple independent checks) or over-fine noise (every micro-check becomes a top-level criterion). The Agent-as-a-Judge pattern resolves this by allowing a criterion to have children (`AC-XXX-01.a`, `AC-XXX-01.b`) that inherit the parent's Given context but carry distinct When/Then.

Hierarchical acceptance unlocks: (a) per-sub-criterion scoring by evaluator-active (SPEC-V3R2-HRN-003), (b) partial-pass status at the parent level (3 of 4 children passed → parent is `refined`), (c) finer Sprint Contract state (SPEC-V3R2-HRN-002 carries forward children independently).

Flat legacy acceptance remains parseable; migration is automatic (BC-V3R2-011): a flat `AC-XXX-01` becomes `AC-XXX-01` with a single child `AC-XXX-01.a` carrying the same When/Then.

## 2. Scope (범위)

### 2.1 In Scope

- Extend Go types in `internal/spec/ears.go`: `Acceptance` struct gains `Children []Acceptance` field; ID format supports lowercase-letter nesting (`01.a`, `01.a.i`, up to 3 levels deep).
- Update SPEC markdown parser to read hierarchical acceptance syntax: a child is a list item immediately below its parent at deeper indentation.
- EARS modality table published verbatim in `.claude/rules/moai/workflow/spec-workflow.md` — no modality content change, only acceptance format extension.
- Back-compat: flat `AC-XXX-NN` entries are treated as a parent with exactly one child `AC-XXX-NN.a` carrying the same When/Then. SPEC linter (SPEC-V3R2-SPC-003) does not flag flat-only SPECs.
- Per-child REQ mapping: each child acceptance has its own `(maps REQ-...)` tail. Children may map to different REQs than their parent.
- Migration tool (SPEC-V3R2-MIG-001) wraps every existing flat acceptance as a 1-child parent for zero-loss upgrade.
- Renderer updates for `moai spec view SPEC-XXX` to display the tree.
- This SPEC constitutes a FROZEN-zone amendment to the SPEC system's EARS acceptance-criteria format. The amendment protocol defined in SPEC-V3R2-CON-002 applies: 5-layer safety gate (FrozenGuard, Canary, ContradictionDetector, RateLimiter, HumanOversight), explicit before/after schema text (see §11 below), and explicit Human Oversight approval are required before this SPEC may land. The change is additive (flat remains parseable) so Canary regression risk is bounded.

### 2.2 Out of Scope

- Hierarchical REQ IDs (REQs stay flat: `REQ-V3R2-XXX-NNN`).
- EARS modality additions or changes (ubiquitous/event/state/optional/unwanted are preserved verbatim per FROZEN CON-001).
- Sprint Contract state shape (→ SPEC-V3R2-HRN-002).
- Per-sub-criterion scoring by evaluator (→ SPEC-V3R2-HRN-003).
- @MX TAG extensions (→ SPEC-V3R2-SPC-002).
- SPEC linter implementation (→ SPEC-V3R2-SPC-003).

## 3. Environment (환경)

Current moai state (v2.13.2):

- `.claude/rules/moai/workflow/spec-workflow.md` defines EARS modality table and acceptance format (flat).
- `.moai/specs/SPEC-XXX/spec.md` template (one of three mandatory files per SKILL moai-workflow-spec) has section "Acceptance Criteria" with `AC-ID-NN: Given ... When ... Then ... (maps REQ-...)` format.
- `internal/spec/` (to be created) — no existing Go parser for SPEC markdown. Today SPEC files are consumed by agents as raw text.
- Legacy v3 SPECs (e.g., `SPEC-V3-HOOKS-001/spec.md`) use flat format: `AC-HOOKS-001-01`, `AC-HOOKS-001-02`, … all sibling.
- Design constitution §11.4 Sprint Contract references "acceptance checklist" at criterion level — hierarchical shape not yet modeled.

Reference: R1 §9 Agent-as-a-Judge (Zhuge et al. 2024) demonstrates 55 → 365 requirement ratio on DevAI; hierarchical structure is essential for scoring per-node.

## 4. Assumptions (가정)

- Markdown list-item indentation is a reliable parse signal for child relationships. `AC-XXX-01` at column 0, `AC-XXX-01.a` at column 2+, etc.
- ID format is stable: top-level `AC-<DOMAIN>-<NNN>-<NN>` (e.g., `AC-SPC-001-05`), first-level children append `.a`, `.b`, `.c`; second-level children append `.a.i`, `.a.ii` Roman numerals; max depth 3.
- Flat SPECs from v2 and pre-CON-001 v3 are parseable without modification; the parser auto-wraps them.
- EARS modality vocabulary (Ubiquitous / Event-driven / State-driven / Optional / Unwanted / Complex) is FROZEN per CON-001 and not touched here.
- Every leaf acceptance node must have a `(maps REQ-...)` tail; intermediate nodes may omit.

## 5. Requirements (EARS)

### 5.1 Ubiquitous

- REQ-SPC-001-001: The system SHALL represent SPEC acceptance criteria as a tree where each node has an ID, a Given clause (optional if inherited), a When clause, a Then clause, a REQ-mapping list, and a list of children.
- REQ-SPC-001-002: The top-level acceptance ID format SHALL be `AC-<DOMAIN>-<NNN>-<NN>` where DOMAIN is 2-5 uppercase letters and NN is zero-padded 2 digits.
- REQ-SPC-001-003: First-level children SHALL append a lowercase letter `.a`, `.b`, `.c`; second-level descendants SHALL append lowercase roman `.a.i`, `.a.ii`; maximum tree depth SHALL be 3.
- REQ-SPC-001-004: Every leaf node SHALL carry a `(maps REQ-...)` REQ-mapping tail; intermediate nodes MAY omit the tail if all children provide it.
- REQ-SPC-001-005: The `Acceptance` Go struct SHALL expose fields: `ID string`, `Given string`, `When string`, `Then string`, `RequirementIDs []string`, `Children []Acceptance`.
- REQ-SPC-001-006: A child node SHALL inherit its parent's Given clause if the child's own Given is empty; explicit child Given overrides parent.
- REQ-SPC-001-007: The SPEC markdown format SHALL allow representing the tree via indented markdown list items with ID-prefix parsing.

### 5.2 Event-driven

- REQ-SPC-001-010: WHEN a SPEC parser encounters an `AC-XXX-NN` without any `.letter` suffix and no indented children below it, the parser SHALL treat it as a parent with exactly one synthesized child `.a` carrying the same Given/When/Then/REQ.
- REQ-SPC-001-011: WHEN the parser encounters `AC-XXX-NN.a` at deeper indentation below `AC-XXX-NN`, the parser SHALL attach the former as a child of the latter.
- REQ-SPC-001-012: WHEN two acceptance nodes share the same ID at the same level, the parser SHALL emit a `DuplicateAcceptanceID` error and halt.
- REQ-SPC-001-013: WHEN `moai spec view SPEC-XXX` is invoked, the renderer SHALL produce a tree-structured output with visible child indentation.

### 5.3 State-driven

- REQ-SPC-001-020: WHILE a SPEC is being migrated from flat to hierarchical format by SPEC-V3R2-MIG-001, every flat `AC-XXX-NN` SHALL be wrapped as a single-child parent; original content preserved verbatim.
- REQ-SPC-001-021: WHILE an acceptance node's depth exceeds 3 levels, the parser SHALL emit a `MaxDepthExceeded` error naming the offending ID.

### 5.4 Optional

- REQ-SPC-001-030: WHERE the SPEC YAML frontmatter declares `acceptance_format: flat`, the parser SHALL refuse to attach children even if markdown indentation suggests them; used for SPECs explicitly opting out.
- REQ-SPC-001-031: WHERE `--shape trace` is passed to `moai spec view`, the output SHALL include node depth and parent link for each acceptance entry.

### 5.5 Complex

- REQ-SPC-001-040: WHILE parsing a SPEC AND the acceptance section contains a mix of top-level nodes with and without children, THEN nodes without children SHALL be auto-wrapped (synthesized `.a` child) AND the mixed-tree SHALL be accepted without warning.
- REQ-SPC-001-041: IF a child node's `(maps REQ-...)` tail references a REQ ID not declared in the Requirements section, THEN the parser SHALL emit a `DanglingRequirementReference` warning (not an error — full validation lives in SPEC-V3R2-SPC-003 linter).
- REQ-SPC-001-042: WHEN SPEC linter (SPEC-V3R2-SPC-003) computes REQ→AC coverage, parent and leaf acceptances SHALL both count toward coverage; a flat-only SPEC with 10 ACs and 10 REQs and 1:1 mapping achieves 100% coverage without needing children.

## 6. Acceptance Criteria

- AC-SPC-001-01: Given a SPEC with hierarchical acceptance `AC-X-01` containing two children `AC-X-01.a` and `AC-X-01.b`, When the parser loads it, Then `Acceptance{ID: "AC-X-01", Children: [{ID: "AC-X-01.a"}, {ID: "AC-X-01.b"}]}` is produced. (maps REQ-SPC-001-001, REQ-SPC-001-005, REQ-SPC-001-011)
- AC-SPC-001-02: Given a flat legacy SPEC with `AC-HOOKS-001-01: Given X When Y Then Z`, When the parser loads it, Then it produces `AC-HOOKS-001-01` with one synthesized child `AC-HOOKS-001-01.a` whose Given/When/Then match the parent verbatim. (maps REQ-SPC-001-010, REQ-SPC-001-020)
- AC-SPC-001-03: Given a child node without its own Given clause below a parent with Given "user authenticated", When the parser resolves the child, Then the child's effective Given is "user authenticated". (maps REQ-SPC-001-006)
- AC-SPC-001-04: Given two acceptance nodes with identical IDs at the same depth, When the parser runs, Then it emits `DuplicateAcceptanceID` and halts. (maps REQ-SPC-001-012)
- AC-SPC-001-05: Given acceptance `AC-X-01.a.i.x` (4-level depth), When the parser runs, Then it emits `MaxDepthExceeded` naming the offending ID. (maps REQ-SPC-001-021)
- AC-SPC-001-06: Given `moai spec view SPEC-XXX` is invoked on a hierarchical SPEC, When the renderer runs, Then the output visually indents children under parents with tree glyphs. (maps REQ-SPC-001-013)
- AC-SPC-001-07: Given a SPEC with YAML `acceptance_format: flat`, When indented child markdown is present, Then the parser ignores the nesting and produces flat siblings. (maps REQ-SPC-001-030)
- AC-SPC-001-08: Given `moai spec view SPEC-XXX --shape trace`, When invoked, Then the output for each node includes depth and parent ID fields. (maps REQ-SPC-001-031)
- AC-SPC-001-09: Given a SPEC with mixed top-level nodes (some with children, some without), When parsed, Then nodes without children are auto-wrapped and no warning is emitted. (maps REQ-SPC-001-040)
- AC-SPC-001-10: Given a child with `(maps REQ-X-042)` where REQ-X-042 is not in the Requirements section, When parsed, Then a `DanglingRequirementReference` warning is emitted but parsing continues. (maps REQ-SPC-001-041)
- AC-SPC-001-11: Given the migration tool SPEC-V3R2-MIG-001 runs on a legacy flat SPEC with 8 ACs and 8 REQs, When the post-migration file is re-parsed, Then the tree has 8 top-level nodes each with one synthesized `.a` child and no content is lost. (maps REQ-SPC-001-020)
- AC-SPC-001-12: Given a top-level acceptance with no children and no REQ-mapping tail, When parsed, Then the parser emits a `MissingRequirementMapping` warning (since leaf synthesized child inherits no mapping). (maps REQ-SPC-001-004)
- AC-SPC-001-13: Given a SPEC where parent `AC-X-01` omits its REQ tail but all three children `AC-X-01.a/b/c` carry distinct tails, When parsed, Then no warning is emitted because each leaf has its own mapping. (maps REQ-SPC-001-004)
- AC-SPC-001-14: Given a SPEC with 365 leaf acceptance nodes across 55 top-level parents (Agent-as-a-Judge DevAI shape), When parsed, Then the parser succeeds within a 500ms budget and the tree is well-formed. (maps REQ-SPC-001-001, REQ-SPC-001-003)

## 7. Constraints (제약)

- Maximum tree depth 3 levels. Beyond that, SPEC authors should split into separate parent acceptances.
- ID format letters (`.a`, `.b`) are case-sensitive lowercase; Roman numerals (`.i`, `.ii`) lowercase.
- Flat-SPEC back-compat is FROZEN — no future v3 release may break it (breaking back-compat would require CON-002 amendment).
- Performance: parser must complete in <500ms for the largest realistic tree (365 leaves).
- Storage: acceptance trees add negligible disk footprint; no separate file, inline in spec.md.
- Migration for existing SPECs: BC-V3R2-011 covered by SPEC-V3R2-MIG-001 with automatic 1-child wrapping; no user action required.

## 8. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Markdown indentation ambiguity (tabs vs spaces) | MEDIUM | MEDIUM | Parser normalizes to spaces; minimum 2-space indent for child |
| Authors nest too deeply | LOW | LOW | Depth cap 3; MaxDepthExceeded error guides restructuring |
| REQ-coverage computation changes break legacy SPEC linter | LOW | MEDIUM | SPEC-V3R2-SPC-003 explicitly counts both leaves and parents; back-compat preserved |
| Hierarchical acceptance confuses human reviewers | LOW | LOW | `moai spec view` with tree glyphs; training examples in spec-workflow.md |
| ID collisions across depth levels | LOW | MEDIUM | Parser validates uniqueness at each depth level; REQ-SPC-001-012 |

## 9. Dependencies

### 9.1 Blocked by

- SPEC-V3R2-CON-001 (EARS format FROZEN status must be confirmed in zone registry; SPC-001 extends acceptance shape only, not modality).

### 9.2 Blocks

- SPEC-V3R2-SPC-003 (SPEC linter must understand the hierarchical AC format to compute REQ-AC coverage).
- SPEC-V3R2-HRN-002 (Sprint Contract carries per-leaf criterion state across iterations).
- SPEC-V3R2-HRN-003 (hierarchical scoring by evaluator-active depends on tree shape).
- SPEC-V3R2-MIG-001 (migration tool wraps flat ACs).

### 9.3 Related

- R1 §9 Agent-as-a-Judge DevAI benchmark (365 sub-requirements).
- SPEC-V3R2-SPC-002 (@MX TAG) — parallel extension of Layer 2, no direct dependency.
- design-constitution §11.4 Sprint Contract — criteria shape converges here.

## 10. Traceability

- Theme: Layer 2 SPEC & TAG (master-v3 §4).
- Principles: P1 SPEC as Constitutional Contract (design-principles.md §P1 — acceptance tree is the scoreable contract); P4 Evaluator Judgments Fresh, Contract State Durable (design-principles.md §P4 — per-leaf criterion state is durable).
- Patterns: E-1 Agent-as-a-Judge (pattern-library.md §E-1 — 365-sub-req shape); X-1 Markdown + YAML Frontmatter (pattern-library.md §X-1 — tree is encoded in markdown indentation).
- Wave 1 sources: R1 §9 Agent-as-a-Judge (Zhuge et al. 2024, *arXiv:2410.10934*).
- Wave 2 sources: design-principles.md §P4 (evaluator fresh / contract durable), pattern-library.md §E-1 priority 9.
- **File:line anchors** (per D5 traceability requirement):
  - `docs/design/major-v3-master.md:L1035` (§11.2 SPC-001 definition)
  - `docs/design/major-v3-master.md:L970` (§8 BC-V3R2-011 — hierarchical acceptance)
  - `docs/design/major-v3-master.md:L48-56` (§1.3 FROZEN invariants — SPEC EARS format)
  - `docs/design/major-v3-master.md:L992` (§9 Phase 5 Harness + Evaluator)
  - `.claude/rules/moai/workflow/spec-workflow.md` (EARS modality + acceptance format target)

## 11. FROZEN-zone Amendment — Before/After Schema

This SPEC amends a FROZEN invariant (SPEC system EARS acceptance format, per master-v3 §1.3). Per SPEC-V3R2-CON-002, the explicit before/after schema text and the Human Oversight approval record MUST accompany the landing commit.

### 11.1 Before (v2.x acceptance format)

SPEC acceptance criteria are a flat list of `AC-<DOMAIN>-NNN-NN: Given / When / Then` entries under `## 10. Acceptance Criteria`. Each entry is a sibling of all others. No hierarchical nesting is permitted. Each entry carries a single `(maps REQ-...)` tail.

Example:

```
- AC-SPEC-001-01: Given ..., When ..., Then ... (maps REQ-SPEC-001-001)
- AC-SPEC-001-02: Given ..., When ..., Then ... (maps REQ-SPEC-001-002)
- AC-SPEC-001-03: Given ..., When ..., Then ... (maps REQ-SPEC-001-003)
```

### 11.2 After (v3.0+ acceptance format)

SPEC acceptance criteria MAY be flat (backward-compat preserved) OR hierarchical with `AC-<DOMAIN>-NNN-NN.a / .b / .c` sub-scenarios. Maximum nesting depth is 3 levels (`NN`, `NN.a`, `NN.a.i`). Each child inherits its parent's Given context but carries distinct When/Then and its own `(maps REQ-...)` tail. Flat SPECs are automatically wrapped by the linter as 1-child parents (`AC-XXX-NN` ⇒ `AC-XXX-NN` with single child `AC-XXX-NN.a` carrying the identical When/Then) — no source edit required.

Example:

```
- AC-SPEC-001-01: Given ..., When ..., Then ... (maps REQ-SPEC-001-001)
  - AC-SPEC-001-01.a: Given inherited, When variant-A, Then ... (maps REQ-SPEC-001-001)
  - AC-SPEC-001-01.b: Given inherited, When variant-B, Then ... (maps REQ-SPEC-001-002)
- AC-SPEC-001-02: Given ..., When ..., Then ... (maps REQ-SPEC-001-003)   # flat still valid
```

### 11.3 Amendment Safety Gate (per CON-002)

- **FrozenGuard**: Change target is the SPEC acceptance format inside `.claude/rules/moai/workflow/spec-workflow.md`. Change is strictly additive (children optional; flat remains parseable). Lowered safety thresholds: none.
- **Canary**: Re-parse last 10 landed v2 SPECs with the new parser. All 10 MUST parse without warnings and yield identical REQ-coverage set as the flat parser. Implementation gate.
- **ContradictionDetector**: No existing rule prohibits nesting — the flat constraint was an under-specification, not an explicit rule. No contradiction.
- **RateLimiter**: This is 1 of ≤3 FROZEN amendments per v3.x cycle (alongside HRN-002 evaluator memory amendment). Within bound.
- **HumanOversight**: Maintainer approval required at plan-auditor iteration 2 sign-off. Approval record attached to landing commit.

