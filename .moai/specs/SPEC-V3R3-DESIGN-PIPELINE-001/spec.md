---
id: SPEC-V3R3-DESIGN-PIPELINE-001
title: Hybrid Design Pipeline — DTCG 2025.10 + 3-Path Routing
version: "0.1.0"
status: draft
created_at: 2026-04-26
updated_at: 2026-04-26
author: manager-spec
priority: P0
phase: "v3.0.0 R3 — Phase C — Design Pipeline Hybridization"
module: ".claude/skills/moai-workflow-design-import/, .claude/skills/moai-design-system/, internal/design/dtcg/, .claude/skills/moai/workflows/design.md, .claude/rules/moai/design/constitution.md"
depends_on:
  - SPEC-V3R3-HARNESS-001
related_specs:
  - SPEC-DESIGN-CONST-AMEND-001
  - SPEC-AGENCY-ABSORB-001
  - SPEC-V3R3-PROJECT-HARNESS-001
breaking: false
bc_id: []
lifecycle: spec-anchored
labels: [design, dtcg, figma, pencil, claude-design, hybrid-pipeline, v3r3, phase-c]
related_theme: "Phase C — Design Pipeline Hybridization"
target_release: v2.17.0
issue_number: null
---

# SPEC-V3R3-DESIGN-PIPELINE-001: Hybrid Design Pipeline — DTCG 2025.10 + 3-Path Routing

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-04-26 | manager-spec | Initial draft. Phase C P0 — Hybrid design pipeline (Path A Claude Design / B1 Figma / B2 Pencil) with W3C DTCG 2025.10 token spec validator, depending on SPEC-V3R3-HARNESS-001 meta-harness skill. |

---

## 1. Goal (목적)

Hybrid design pipeline MUST route user input to one of three execution paths — Path A (Claude Design handoff bundle), Path B1 (Figma → meta-harness dynamic figma-extractor skill), or Path B2 (Pencil → meta-harness dynamic pencil-mcp skill) — and MUST validate every produced design token against the W3C Design Tokens Community Group (DTCG) 2025.10 specification through a Go-based validator. Path A behavior is preserved verbatim. The legacy `moai-workflow-pencil-integration` skill is removed; its capability is recreated dynamically through the meta-harness pattern established in SPEC-V3R3-HARNESS-001. The design constitution §4 (Phase Contracts) is amended additively to enumerate Path B1/B2 rows; FROZEN zones (§2, §3.1, §3.2, §3.3) remain inviolable.

### 1.1 Background

- Existing design pipeline supports only Path A (Claude Design handoff) via `moai-workflow-design-import` and a Pencil-only path via `moai-workflow-pencil-integration`.
- SPEC-V3R3-HARNESS-001 introduces the meta-harness mechanism that generates user-area skills (`my-harness-*`) on demand, replacing 16 statically-shipped skills including `moai-workflow-pencil-integration`.
- The W3C DTCG specification reached 2025.10 milestone with stable categories for color, dimension, font, fontFamily, fontWeight, duration, cubicBezier, number, strokeStyle, border, transition, shadow, gradient, and typography. Existing Go code has no DTCG validator; tokens flow into `expert-frontend` without schema enforcement.
- Figma integration was partially scoped under SPEC-AGENCY-ABSORB-001 but never landed because static skills could not capture per-project Figma quirks. The meta-harness model now enables dynamic generation of Figma extractors tailored to each project.
- Brand context (`.moai/project/brand/`) remains the constitutional parent per design constitution §3.3 — no path may override brand constraints.

### 1.2 Non-Goals

- This SPEC does NOT modify Path A (Claude Design handoff) behavior beyond adding the DTCG validator gate.
- This SPEC does NOT ship static Figma or Pencil skills — both are generated dynamically by meta-harness per project (depends on SPEC-V3R3-HARNESS-001).
- This SPEC does NOT modify FROZEN zones of the design constitution (§2, §3.1 Brand Context, §3.2 Design Brief, §3.3 Relationship). Only §4 Phase Contracts table receives an additive row update for Path B1/B2.
- This SPEC does NOT alter the GAN Loop contract (constitution §11) or evaluator leniency prevention mechanisms (§12).
- No network calls to Figma API or Pencil services from MoAI core — the dynamically generated extractor skills handle credentialed external access locally.

---

## 2. Scope

### 2.1 In Scope

- `.claude/skills/moai-workflow-design-import/SKILL.md` body extended with three-path entry-point routing (Path A preserved verbatim, Path B1 and B2 added).
- `.claude/skills/moai-design-system/SKILL.md` body extended with DTCG 2025.10 token spec reference and validator invocation guidance.
- `internal/design/dtcg/` new Go package: token JSON validator covering color, dimension, font, fontFamily, fontWeight, duration, cubicBezier, number, strokeStyle, border, transition, shadow, gradient, typography categories per DTCG 2025.10.
- `internal/design/dtcg/dtcg_test.go` unit tests (TDD RED → GREEN) covering minimum 6 categories: color, dimension, font, typography, shadow, border.
- `.claude/skills/moai/workflows/design.md` workflow extended with Path A/B1/B2 routing logic and AskUserQuestion path-selection prompt (Path A is the recommended default and appears first).
- `.claude/rules/moai/design/constitution.md` §4 Phase Contracts table updated additively with Path B1 (figma-extractor row) and Path B2 (pencil-mcp row); HISTORY entry appended.
- Removal of `moai-workflow-pencil-integration` skill from static skill set (concurrent with SPEC-V3R3-HARNESS-001 BC-V3R3-007 16-skill removal).
- Template-First mirrors under `internal/template/templates/` for every modified skill, workflow, and rule file.
- Integration with `expert-frontend` agent so that token JSON consumption automatically invokes the DTCG validator.

### 2.2 Out of Scope

- Static Figma or Pencil skills — generated dynamically by meta-harness, not shipped statically.
- Visual diff tooling for design QA — deferred.
- Auto-translation between DTCG 2024.x and 2025.10 tokens — out of scope; users running 2024.x tokens MUST upgrade manually.
- Modification of `moai-domain-copywriting` or `moai-domain-brand-design` skills (FROZEN per design constitution §2 [EVOLVABLE] but not part of this SPEC scope).
- Network egress to Figma or Pencil from MoAI core packages.

---

## 3. Stakeholders

| Role | Interest |
|------|----------|
| Solo developer (design-first) | Single workflow command routes to the correct path without manual skill selection. |
| Brand owner | Brand context constraints (`.moai/project/brand/`) remain authoritative across all three paths. |
| Frontend engineer | DTCG-validated tokens guarantee schema correctness before code generation. |
| Plan-auditor | Constitution §4 amendments are additive; FROZEN zones untouched; verifiable via diff. |
| User upgrading from `moai-workflow-pencil-integration` | Migration path is automatic via meta-harness; no manual configuration loss. |

---

## 4. Exclusions (What NOT to Build)

[HARD] This SPEC explicitly EXCLUDES the following — building any of these is a scope violation:

1. Modification of FROZEN zones in `.claude/rules/moai/design/constitution.md` (§2, §3.1 Brand Context, §3.2 Design Brief, §3.3 Relationship, §5 Safety Architecture, §11 GAN Loop contract, §12 Evaluator Leniency Prevention).
2. Static Figma or Pencil skill files under `.claude/skills/` (excluded by reliance on SPEC-V3R3-HARNESS-001 meta-harness pattern).
3. Network egress to Figma API, Pencil cloud, or any third-party service from `internal/design/dtcg/` Go package.
4. Auto-conversion of legacy DTCG 2024.x tokens — manual upgrade only.
5. Replacement of `expert-frontend` agent's existing token-consumption flow — DTCG validator is added as a pre-consumption gate, not a replacement.
6. Modification of brand context files (`.moai/project/brand/brand-voice.md`, `visual-identity.md`, `target-audience.md`) — brand remains the constitutional parent.
7. UI/UX rendering or preview generation — out of scope; consumers handle rendering.

---

## 5. Requirements (EARS format)

### REQ-DPL-001 (Ubiquitous — Path A Preservation)

The system **shall** preserve the existing Claude Design handoff bundle workflow (Path A) verbatim in behavior and output paths (`.moai/design/` reserved file paths per constitution §3.2: `tokens.json`, `components.json`, `assets/`, `import-warnings.json`, `brief/BRIEF-*.md`), and **shall** invoke the new DTCG 2025.10 validator on `tokens.json` before downstream consumption by `expert-frontend`.

### REQ-DPL-002 (Event-Driven — Path B1 Figma Routing)

**When** the user selects Path B1 (Figma source) via the `/moai design` AskUserQuestion prompt, the system **shall** invoke the meta-harness skill (defined by SPEC-V3R3-HARNESS-001) to generate a project-scoped figma-extractor skill at `.claude/skills/my-harness-figma-extractor/SKILL.md` with project-specific Figma file ID, page selectors, and credential reference.

### REQ-DPL-003 (Event-Driven — Path B2 Pencil Routing)

**When** the user selects Path B2 (Pencil source) via the `/moai design` AskUserQuestion prompt, the system **shall** invoke the meta-harness skill to generate a project-scoped pencil-mcp skill at `.claude/skills/my-harness-pencil-mcp/SKILL.md` configured for the project's `.pen` file locations and Pencil MCP server endpoint.

### REQ-DPL-004 (Ubiquitous — DTCG Validator)

The system **shall** provide a Go validator at `internal/design/dtcg/` that parses token JSON files and validates each token against the W3C DTCG 2025.10 specification, covering at minimum the following categories: color, dimension, font, fontFamily, fontWeight, duration, cubicBezier, number, strokeStyle, border, transition, shadow, gradient, typography. Validation failures **shall** produce structured error reports identifying the offending token path and the violated rule.

### REQ-DPL-005 (Event-Driven — Workflow Path Branching)

**When** `/moai design` is invoked, the workflow **shall** branch on user-selected path before delegating to any specialist agent: Path A → existing import flow; Path B1 → meta-harness figma-extractor generation then import; Path B2 → meta-harness pencil-mcp generation then import. The selection **shall** be persisted in `.moai/design/path-selection.json` for audit.

### REQ-DPL-006 (Ubiquitous — Pencil-Integration Removal)

The system **shall** remove `moai-workflow-pencil-integration` from the static skill set as part of SPEC-V3R3-HARNESS-001 BC-V3R3-007 16-skill removal, and `/moai design` **shall** continue to support Pencil workflows exclusively through Path B2 (dynamically generated `my-harness-pencil-mcp`).

### REQ-DPL-007 (Ubiquitous — Constitution §4 Amendment)

The system **shall** amend `.claude/rules/moai/design/constitution.md` §4 Phase Contracts table additively with two new rows for Path B1 (figma-extractor: BRIEF + Figma file ID → tokens.json + components.json) and Path B2 (pencil-mcp: BRIEF + .pen files → tokens.json + components.json), and **shall not** modify any other §4 row or any other constitution section. A HISTORY entry **shall** be appended documenting the additive change.

### REQ-DPL-008 (Event-Driven — AskUserQuestion Path Selection)

**When** `/moai design` requires path selection, the system **shall** present an AskUserQuestion with exactly three options ordered as: Path A (Claude Design) marked "(권장)" first, Path B1 (Figma) second, Path B2 (Pencil) third. Each option **shall** include a description explaining the prerequisite (Claude Design handoff bundle / Figma file access / `.pen` files).

### REQ-DPL-009 (State-Driven — Brand Context Priority)

**While** any of the three paths is executing, the system **shall** load brand context from `.moai/project/brand/` and **shall** treat brand constraints as authoritative — token values from any path that conflict with brand visual-identity.md MUST be flagged as warnings and presented for user resolution. Brand context **shall not** be overridden by any path.

### REQ-DPL-010 (Event-Driven — Validator Auto-Invocation)

**When** `expert-frontend` consumes a `tokens.json` produced by any path (A, B1, or B2), the agent **shall** invoke the DTCG validator (REQ-DPL-004) before generating any frontend code. Invalid tokens **shall** block code generation and **shall** surface a structured error report to the orchestrator.

### REQ-DPL-011 (Unwanted — FROZEN Zone Protection)

**If** any pipeline component attempts to write to a FROZEN zone defined in design constitution §2 (constitution file itself outside §4 Phase Contracts table, §3.1 Brand Context, §3.2 Design Brief content, §3.3 Relationship rules, §5 Safety Architecture, §11 GAN Loop contract, §12 Evaluator Leniency Prevention), **then** the build **shall** fail with a clear error message identifying the violated FROZEN zone, and the change **shall not** be persisted.

### REQ-DPL-012 (Ubiquitous — Template-First Mirror)

The system **shall** mirror every modified skill (`moai-workflow-design-import/SKILL.md`, `moai-design-system/SKILL.md`), workflow (`design.md`), and rule file (`constitution.md`) to `internal/template/templates/` so `moai init` deploys the updated assets to new projects.

---

## 6. Acceptance Coverage Map

| AC ID    | Covers REQ-IDs                                                  |
|----------|-----------------------------------------------------------------|
| AC-DPL-01 | REQ-DPL-001, REQ-DPL-002, REQ-DPL-003 (3-path workflow)         |
| AC-DPL-02 | REQ-DPL-004 (DTCG validator unit tests, 6+ categories)          |
| AC-DPL-03 | REQ-DPL-005, REQ-DPL-008 (workflow branching + AskUserQuestion) |
| AC-DPL-04 | REQ-DPL-006 (pencil-integration removal compatibility)          |
| AC-DPL-05 | REQ-DPL-009, REQ-DPL-011 (brand context FROZEN protection)      |
| AC-DPL-06 | REQ-DPL-010 (validator auto-invocation, GAN loop integration)   |
| AC-DPL-07 | REQ-DPL-007, REQ-DPL-012 (constitution §4 amendment + Template-First) |

Coverage: 12 REQs ↔ 7 ACs, 100% (every REQ appears in at least one AC).

---

## 7. Constraints

- [HARD] Path A behavior preservation — no behavioral regression in Claude Design handoff bundle import.
- [HARD] Brand context priority — `.moai/project/brand/` constraints win on conflict per design constitution §3.3.
- [HARD] FROZEN zone immutability — design constitution §2, §3.1, §3.2, §3.3, §5, §11, §12 are not modified.
- [HARD] Template-First — every modified skill / workflow / rule has a mirror under `internal/template/templates/`.
- [HARD] No network egress from `internal/design/dtcg/` package.
- [HARD] DTCG validator MUST cover at least 6 categories (color, dimension, font, typography, shadow, border) in unit tests.
- [HARD] AskUserQuestion path selection — Path A first, marked "(권장)".
- [HARD] Pencil-integration removal coordinated with SPEC-V3R3-HARNESS-001 BC-V3R3-007 — no orphaned references in routing tables.
- All instruction files in English (per `coding-standards.md` Language Policy).
- 16-language neutrality preserved — Path B1/B2 generated extractors produce tokens that work across all 16 supported languages.

---

## 8. Risks

| Risk | Mitigation |
|------|------------|
| DTCG 2025.10 spec drift before v2.17 release | Pin validator to a specific spec snapshot date; document the snapshot in `internal/design/dtcg/SPEC.md`. |
| Path B1/B2 generated extractors produce malformed tokens | DTCG validator gates downstream consumption (REQ-DPL-010); evaluator-active flags inconsistencies in GAN loop. |
| Constitution §4 amendment introduces conflicts with existing rows | Additive-only update; plan-auditor verifies no row modification beyond appended Path B1/B2 rows. |
| Pencil-integration removal breaks projects mid-flight | SPEC-V3R3-HARNESS-001 BC-V3R3-007 migration script archives the legacy skill and invokes meta-harness regeneration on first `/moai design` after upgrade. |
| AskUserQuestion path selection confuses non-design users | Default Path A "(권장)" is the safest fallback; description text explains prerequisites; Help text in `/moai design --help`. |
| Brand context conflict resolution UX | Conflicts surfaced as warnings via AskUserQuestion, not silent overrides; user explicitly chooses brand-vs-token precedence per conflict. |
| FROZEN zone false positive | FROZEN zone matcher uses path prefix + section heading hash; verified by IT against the current constitution version. |

---

## 9. Dependencies

| SPEC | Relationship | Notes |
|------|--------------|-------|
| SPEC-V3R3-HARNESS-001 | Hard prerequisite | Provides meta-harness skill that generates `my-harness-figma-extractor` and `my-harness-pencil-mcp` for Path B1/B2. Without HARNESS-001, Path B1/B2 cannot operate. |
| SPEC-DESIGN-CONST-AMEND-001 | Reference | Establishes the §3 tripartite structure and FROZEN zone rules this SPEC respects. |
| SPEC-AGENCY-ABSORB-001 | Reference | Established `.moai/design/` reserved paths and brand context conventions this SPEC preserves. |
| SPEC-V3R3-PROJECT-HARNESS-001 | Soft co-dependency | The Socratic interview in PROJECT-HARNESS-001 may pre-select a default design path; this SPEC honors that selection but does not require it. |

---

## 10. Glossary

- **DTCG 2025.10**: W3C Design Tokens Community Group token specification, October 2025 milestone snapshot.
- **Path A**: Claude Design handoff bundle import (existing flow, preserved verbatim).
- **Path B1**: Figma source via meta-harness-generated `my-harness-figma-extractor` skill.
- **Path B2**: Pencil source via meta-harness-generated `my-harness-pencil-mcp` skill.
- **Reserved file paths**: `tokens.json`, `components.json`, `assets/`, `import-warnings.json`, `brief/BRIEF-*.md` (constitution §3.2).
- **FROZEN zone**: Constitutional sections §2, §3.1, §3.2, §3.3, §5, §11, §12 — immutable to evolution per design constitution.
- **Brand context**: `.moai/project/brand/{brand-voice.md, visual-identity.md, target-audience.md}` — constitutional parent that wins on conflict.
- **Meta-harness**: SPEC-V3R3-HARNESS-001 skill that generates user-area skills dynamically.
