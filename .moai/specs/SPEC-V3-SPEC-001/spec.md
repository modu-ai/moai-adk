---
id: SPEC-V3-SPEC-001
title: "SPEC-to-SPEC Chaining — inheritance, templates, cross-REQ traceability"
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: GOOS
priority: P2 Medium
phase: "v3.0.0 — Phase 6a Tier 2 Strategic Differentiators"
module: "internal/spec/"
dependencies:
  - SPEC-V3-SCH-001
related_gap:
  - none (moai-unique feature, no CC equivalent)
related_theme: "Theme 9 — Internal Cleanup (adjacent); moai differentiation (master-v3 §1.1)"
breaking: false
lifecycle: spec-anchored
tags: "spec, chaining, inheritance, templates, traceability, moai-unique, v3"
---

# SPEC-V3-SPEC-001: SPEC-to-SPEC Chaining

## HISTORY

| Version | Date       | Author | Description                                   |
|---------|------------|--------|-----------------------------------------------|
| 0.1.0   | 2026-04-22 | GOOS   | Initial v3 draft (Wave 4, moai-unique bundle) |

---

## 1. Goal (목적)

Formalize SPEC-to-SPEC chaining as a first-class moai primitive: SPEC inheritance via `inherits:` frontmatter field, SPEC templates stored under `.moai/spec-templates/`, and cross-SPEC REQ traceability via `traces_to:` back-pointers. This is a moai-unique differentiator (no Claude Code equivalent) flagged in master-v3 §1.1 design commitment #3: *"Additive moai differentiation: SPEC-to-SPEC chaining, SPEC templates, SPEC lifecycle transitions, and the harness-based quality routing are moai-unique — v3 formalizes them as first-class primitives, not afterthoughts."* With 104 active SPECs today (findings-wave1-moai-current.md §11), ad-hoc cross-referencing via prose "related: SPEC-XXX" lines is insufficient for automated traceability, TAG sync, and REQ impact analysis.

### 1.1 배경

findings-wave1-moai-current.md §11 inventory:
- 104 active SPECs in `.moai/specs/SPEC-XXX/` with triple-file pattern (`spec.md`, `plan.md`, `acceptance.md`).
- EARS requirements format standardized; YAML frontmatter with `spec_id`, `phase`, `lifecycle`.
- Existing SPECs already use prose `Dependencies.Related` sections but without machine-parseable linkage.

Gap-matrix: this feature is **not in the CC gap matrix** because CC has no SPEC system. It is a moai-unique capability that must be designed from first principles, with guidance from:
- `.claude/rules/moai/workflow/spec-workflow.md` (existing Plan → Run → Sync pipeline)
- `.claude/skills/moai-workflow-spec/SKILL.md` (SPEC lifecycle transitions — `spec-first` → `spec-anchored` → `spec-as-source`, SDD 2025 standard)
- master-v3 §3.9 (Theme 9) — reference for SPEC archival pattern (`.agency/` archival migration)

Three concrete capabilities this SPEC delivers:

1. **SPEC Inheritance** — A child SPEC declares `inherits: SPEC-V3-PARENT-001` in frontmatter. At load time the parent's REQ and AC lists are merged into the child's view; child may override individual REQs by re-declaring the same REQ-ID with `override: true` attribute.
2. **SPEC Templates** — Reusable scaffolding stored under `.moai/spec-templates/`. Invoked via `moai plan --template=<name>` (CLI wiring in sibling SPEC-V3-CLI-001; this SPEC defines the format and resolver). Templates use Go `text/template` syntax with a stable variable set (`.Domain`, `.Phase`, `.Author`, `.Today`, `.SpecID`, `.Title`, `.Tags`).
3. **Cross-SPEC REQ Traceability** — `traces_to:` frontmatter field on a REQ (allowed in `plan.md` and `acceptance.md`) links a child REQ to a parent REQ by ID. The traceability index (`internal/spec/trace_index.go`) validates that every `traces_to:` target exists and builds a reverse-lookup map for `moai doctor spec --trace`.

### 1.2 Non-Goals

- Community SPEC marketplace / SPEC library (deferred to v3.1+ per priority-roadmap T2-DIFF-01 note).
- Runtime code-level `@SPEC:` tag integration (addressed by existing @MX tag protocol; this SPEC is about SPEC-document-level chaining).
- Automatic REQ merging for conflicting REQs (manual override required).
- Visual SPEC graph rendering (deferred; text-mode only in v3.0).
- SPEC versioning with semantic-version chains (SPECs use `version:` field but no semver enforcement in v3.0).
- Migration of all 104 existing SPECs to use `inherits:` (opt-in; existing SPECs work unchanged).
- Breaking changes to the 3-file SPEC structure (spec.md / plan.md / acceptance.md remain).
- Plan-auditor re-evaluation of chained SPECs (sibling concern; SPEC-V3-AGT-001 manages audit agents).
- Cross-project SPEC imports (SPECs are project-scoped; remote SPEC inheritance deferred).

---

## 2. Scope (범위)

### 2.1 In Scope

- `internal/spec/` new package with:
  - `frontmatter.go` — extend existing SPEC frontmatter parser to accept `inherits: string` and `traces_to: []string` fields.
  - `inheritance.go` — resolver computing merged REQ/AC sets; cycle detection via DFS.
  - `trace_index.go` — forward and reverse index of `traces_to:` links; orphan detection.
  - `template.go` — SPEC template resolver and renderer (Go `text/template`).
  - `registry.go` — in-memory SPEC registry built from `.moai/specs/` scan; lazy-loaded.
- Frontmatter schema extensions (schema defined via SPEC-V3-SCH-001):
  - Top-level: `inherits: string` (optional; SPEC ID pattern `^SPEC-[A-Z0-9-]+-[0-9]+$`).
  - REQ-level (in plan.md/acceptance.md): `traces_to: []string` (optional; list of REQ IDs).
  - Top-level: `chain_depth_max: int` (internal diagnostic; default 5).
- `.moai/spec-templates/` directory with at least 2 seed templates shipped in the template tree:
  - `feature-basic/` — skeleton for a straightforward feature SPEC (3 files).
  - `refactor-safe/` — skeleton for a refactor SPEC with characterization-test emphasis.
- Validator: `moai doctor spec --trace` (wired in SPEC-V3-CLI-001) runs:
  - Cycle detection (no A inherits B inherits A).
  - Depth check (chain depth ≤ 5).
  - `traces_to:` target existence validation.
  - Orphan detection (REQ declared but never referenced).
  - Override correctness (override flag only when parent REQ exists).
- Lifecycle enrichment: existing `lifecycle: spec-first|spec-anchored|spec-as-source` field (SDD 2025) integrated — inherited SPECs inherit parent lifecycle by default; child may override explicitly.
- Documentation: `docs-site/content/{ko,en,ja,zh}/spec/chaining.md` (4-locale per CLAUDE.local.md §17.3; written during Phase 7).

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- SPEC marketplace / community library (v3.1+ integration with SPEC-V3-PLG-001 plugin system).
- Graphical SPEC dependency visualization (text reports only).
- Automatic SPEC migration (existing SPECs remain valid without `inherits:`).
- Remote SPEC inheritance across repositories.
- SPEC versioning semver enforcement.
- Plan-auditor changes (sibling SPEC scope).
- Python / other language SPEC format variants (Markdown + YAML frontmatter only).
- Real-time REQ change propagation (resolver is load-time; no reactive updates).
- SPEC locking / freezing mechanisms beyond lifecycle field.
- Multi-parent inheritance (single-parent chain only in v3.0; composition deferred to v3.2).
- `@SPEC:` code annotation rewriting (out of scope; @MX tag protocol handles this).
- Template variable extension API (Go `text/template` native set only; no custom funcs in v3.0 beyond `now`, `upper`, `lower`).

---

## 3. Environment (환경)

- 런타임: Go 1.23+, moai-adk-go v3.0.0+.
- Platforms: macOS / Linux / Windows (file I/O stdlib).
- Target directories:
  - `internal/spec/` (new package, ~6 files).
  - `.moai/spec-templates/` (new; runtime, deployed by template system).
  - `internal/template/templates/.moai/spec-templates/` (new; source).
  - `internal/config/schema/spec_schema.go` (new; extension to SPEC frontmatter schema).
- Dependencies: `gopkg.in/yaml.v3` (existing), `text/template` (stdlib).
- Integration points:
  - CLI: `moai plan --template=<name>` — wiring in SPEC-V3-CLI-001 (this SPEC defines resolver API).
  - CLI: `moai doctor spec --trace` — wiring in SPEC-V3-CLI-001 (this SPEC defines validator API).
  - Frontmatter schema: extended via SPEC-V3-SCH-001 validator/v10 tags.

---

## 4. Assumptions (가정)

- A-SPEC-001-001: All 104 existing SPECs parse cleanly with the extended frontmatter schema (no breakage since new fields are optional).
- A-SPEC-001-002: Users opt into `inherits:` gradually; no migration bulk rewrite required.
- A-SPEC-001-003: SPEC templates in `.moai/spec-templates/` are under user control; users may author project-specific templates.
- A-SPEC-001-004: Single-parent inheritance is sufficient for v3.0; multi-parent composition is rare enough to defer.
- A-SPEC-001-005: Chain depth ≤ 5 covers all realistic use cases (typical depth is 1–2 levels).
- A-SPEC-001-006: REQ-ID uniqueness within a SPEC is already enforced by existing convention (EARS requirements always have unique `REQ-DOMAIN-NNN-NNN` IDs).
- A-SPEC-001-007: `traces_to:` is an authoring convenience, not a runtime enforcement — missing `traces_to:` does not block SPEC acceptance.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-SPEC-001-001 (Ubiquitous) — frontmatter fields**
The SPEC frontmatter parser **shall** accept the following new optional top-level fields: `inherits string`, `chain_depth_max int`. Absence of these fields **shall** result in the SPEC being treated as a standalone (non-chained) SPEC.

**REQ-SPEC-001-002 (Ubiquitous) — traces_to field**
The EARS requirements parser **shall** accept an optional `traces_to: []string` attribute on individual REQ entries in `plan.md` and `acceptance.md`. Absence **shall** be treated as no cross-SPEC link.

**REQ-SPEC-001-003 (Ubiquitous) — template directory**
The `moai init` and `moai update` workflows **shall** deploy a `.moai/spec-templates/` directory populated with at least two seed templates (`feature-basic`, `refactor-safe`).

**REQ-SPEC-001-004 (Ubiquitous) — registry**
The package **shall** expose `spec.Registry` with methods `Load(root string)`, `Get(id string) (*Spec, error)`, `List() []*Spec`, `ResolveChain(id string) (*ResolvedSpec, error)`.

**REQ-SPEC-001-005 (Ubiquitous) — inheritance format**
The value of `inherits:` **shall** match the regex `^SPEC-[A-Z0-9]+(-[A-Z0-9]+)*-[0-9]+$` (same pattern as existing SPEC IDs). Invalid formats **shall** produce a validation error.

### 5.2 Event-Driven Requirements

**REQ-SPEC-001-010 (Event-Driven) — cycle detection**
**When** `ResolveChain(id)` is invoked, the resolver **shall** perform depth-first traversal of `inherits:` edges and return `ErrSpecCycle` with the cycle path if any cycle is detected.

**REQ-SPEC-001-011 (Event-Driven) — depth enforcement**
**When** chain depth exceeds `chain_depth_max` (default 5), `ResolveChain` **shall** return `ErrSpecDepthExceeded` naming the terminal SPEC ID.

**REQ-SPEC-001-012 (Event-Driven) — REQ merge**
**When** a child SPEC inherits from a parent, the resolver **shall** compute the merged REQ set as: child REQs override parent REQs with the same `REQ-ID` when `override: true` is set; otherwise parent REQ is retained and a warning is logged for silent shadowing.

**REQ-SPEC-001-013 (Event-Driven) — AC merge**
**When** a child SPEC inherits from a parent, the resolver **shall** compute the merged AC set with the same override semantics as REQ merge.

**REQ-SPEC-001-014 (Event-Driven) — traces_to validation**
**When** the trace-index validator runs, it **shall** verify every `traces_to:` target REQ exists in either the same SPEC or a SPEC reachable via the registry. Unresolved targets **shall** produce `ErrTraceTargetNotFound` with the source REQ-ID and target.

**REQ-SPEC-001-015 (Event-Driven) — template render**
**When** `moai plan --template=<name> --id=SPEC-XXX --title="..."` is invoked (CLI in SPEC-V3-CLI-001), the template resolver **shall** render all three files (spec.md, plan.md, acceptance.md) using Go `text/template` syntax with variable set `{Domain, Phase, Author, Today, SpecID, Title, Tags}`. Missing required variables **shall** produce a clear error; unknown variables **shall** render as empty strings with a warning.

### 5.3 State-Driven Requirements

**REQ-SPEC-001-020 (State-Driven) — lifecycle inheritance**
**While** a child SPEC does not declare an explicit `lifecycle:` field, the resolver **shall** inherit the parent's `lifecycle:` value. Explicit child declaration overrides.

**REQ-SPEC-001-021 (State-Driven) — single-parent constraint**
**While** a SPEC has multiple `inherits:` values (not allowed in v3.0), the parser **shall** reject with `ErrMultiParentNotSupported`.

### 5.4 Optional Features

**REQ-SPEC-001-030 (Optional) — override flag**
**Where** a child REQ is intended to replace a parent REQ, the child REQ **shall** be marked with `override: true` (in EARS front-block YAML). Absence of this flag results in shadowing warning per REQ-SPEC-001-012.

**REQ-SPEC-001-031 (Optional) — reverse traceability report**
**Where** `moai doctor spec --trace` is invoked (SPEC-V3-CLI-001), the validator **shall** emit a reverse-lookup table: for each REQ that appears as a `traces_to:` target, list all child REQs referencing it.

### 5.5 Unwanted Behavior

**REQ-SPEC-001-040 (Unwanted) — no silent multi-parent**
If a SPEC declares `inherits:` as a YAML array with ≥ 2 elements, then the parser **shall NOT** silently pick one. It **shall** return `ErrMultiParentNotSupported` with the declared list.

**REQ-SPEC-001-041 (Unwanted) — no cross-project imports**
If `inherits:` references a SPEC ID not present in the local `.moai/specs/` registry, then the resolver **shall NOT** attempt to fetch from any remote source. It **shall** return `ErrSpecNotFound`.

**REQ-SPEC-001-042 (Unwanted) — no implicit template variable**
If a template references an undeclared variable that is not in the documented variable set `{Domain, Phase, Author, Today, SpecID, Title, Tags}`, then the renderer **shall NOT** silently substitute. It **shall** leave the variable literal in place and log a warning (matching Go `text/template` `missingkey=zero` escalated to warn).

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-SPEC-001-01**: A SPEC with `inherits: SPEC-V3-EXAMPLE-001` loads successfully and `ResolveChain` returns a merged REQ set combining parent and child. (maps REQ-SPEC-001-001, -012)
- **AC-SPEC-001-02**: A SPEC with `inherits: SPEC-DOES-NOT-EXIST-999` returns `ErrSpecNotFound` on load. (maps REQ-SPEC-001-041)
- **AC-SPEC-001-03**: A cycle `SPEC-A inherits SPEC-B inherits SPEC-A` triggers `ErrSpecCycle` with the cycle path `[A→B→A]`. (maps REQ-SPEC-001-010)
- **AC-SPEC-001-04**: A chain of depth 7 with `chain_depth_max: 5` triggers `ErrSpecDepthExceeded`. (maps REQ-SPEC-001-011)
- **AC-SPEC-001-05**: Child REQ `REQ-X-001` with `override: true` replaces parent's `REQ-X-001`; without `override: true` parent wins and warning is logged. (maps REQ-SPEC-001-012, -030)
- **AC-SPEC-001-06**: A REQ with `traces_to: [REQ-V3-PARENT-001-001]` passes trace-index validation when parent exists; fails with `ErrTraceTargetNotFound` when parent REQ is absent. (maps REQ-SPEC-001-014)
- **AC-SPEC-001-07**: `moai plan --template=feature-basic --id=SPEC-TEST-001 --title="Test Feature"` renders three files with substituted variables; `{{.Today}}` becomes today's date; unknown `{{.Foo}}` leaves literal with warning. (maps REQ-SPEC-001-015, -042)
- **AC-SPEC-001-08**: A fresh `moai init` project has `.moai/spec-templates/feature-basic/` and `.moai/spec-templates/refactor-safe/` directories with template files. (maps REQ-SPEC-001-003)
- **AC-SPEC-001-09**: A SPEC with `inherits:` as an array of 2 elements fails validation with `ErrMultiParentNotSupported`. (maps REQ-SPEC-001-040, -021)
- **AC-SPEC-001-10**: `go test ./internal/spec/...` passes with ≥ 85% coverage; existing 104 SPECs in `.moai/specs/` parse without errors after SPEC lands.

---

## 7. Constraints (제약)

- **[HARD] Single-parent in v3.0**: multi-parent composition deferred to v3.2 per §1.2 non-goal.
- **[HARD] Additive only**: no existing SPEC requires modification; inheritance is opt-in.
- **[HARD] Chain depth ≤ 5**: enforced by resolver; config may tighten but not loosen beyond 5.
- **[HARD] In-repo only**: SPECs must exist in local `.moai/specs/`; no remote fetch.
- **[HARD] 3-file structure preserved**: spec.md / plan.md / acceptance.md remain (Template renderer enforces this).
- **[HARD] REQ-ID format preserved**: `REQ-{DOMAIN}-{NNN}-{NNN}` pattern unchanged.
- **[HARD] No new direct Go deps**: reuse `gopkg.in/yaml.v3` and `text/template` stdlib; 9-direct-dep budget preserved.
- **[HARD] Template-first**: `.moai/spec-templates/` content added to `internal/template/templates/.moai/spec-templates/` FIRST per CLAUDE.local.md §2.
- **[HARD] Language-neutral templates**: template content uses generic placeholders; no specific programming language bias (CLAUDE.local.md §15).
- **[HARD] Existing SPEC compatibility**: all 104 current SPECs in `.moai/specs/` **must** parse without change.

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 확률 | 영향 | 완화 |
|--------|------|------|------|
| Inheritance surprises users who expect child REQs to shadow parent silently | Medium | Medium | REQ-SPEC-001-012 mandates warning on silent shadow; require explicit `override: true` to replace; docs-site migration guide explains |
| `traces_to:` link rot as parent SPECs are refactored | Medium | Low | Trace-index validator (REQ-SPEC-001-014) catches orphans; `moai doctor spec --trace` in CI |
| Template variable set expansion needs exceed v3.0 scope | Low | Low | Variable set is a stable union; extension deferred to v3.2 with clear deprecation path |
| Chain depth 5 is too restrictive for complex domains | Low | Low | Configurable via `chain_depth_max`; default 5 covers 99% of realistic cases |
| Existing 104 SPECs accidentally match `inherits:` regex and break | Very Low | Medium | Current SPECs don't use this field; regex enforced on write, not read; migration script validates pre-release |
| Cycle detection performance on large SPEC graphs | Low | Low | DFS with visited set; 104 SPECs is trivial; complexity O(V+E) |
| Multi-parent confusion from users coming from other SPEC frameworks | Low | Low | Clear `ErrMultiParentNotSupported` error with link to docs; v3.2 roadmap documents future composition |
| Template rendering errors at `moai init` mask real issues | Low | Medium | Clear error messages with file/line; fallback to shipping templates from embedded tree |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3-SCH-001** — SPEC frontmatter schema extension uses validator/v10 tags for new fields (`inherits`, `chain_depth_max`, `override`, `traces_to`).

### 9.2 Blocks

- **SPEC-V3-CLI-001** — `moai plan --template=<name>` and `moai doctor spec --trace` CLI surfaces consume this SPEC's resolver API.
- Future v3.1 **SPEC-V3-SPEC-002** (SPEC marketplace) — depends on chaining primitives here.
- Future v3.2 **SPEC-V3-SPEC-003** (multi-parent composition) — extends single-parent model.

### 9.3 Related

- `.claude/rules/moai/workflow/spec-workflow.md` — Plan → Run → Sync pipeline.
- `.claude/skills/moai-workflow-spec/SKILL.md` — SDD 2025 lifecycle transitions.
- **SPEC-V3-OUT-001** — `--json` output format for `moai doctor spec --trace` reports.
- **SPEC-V3-MIG-001** — migration framework pattern reused for future SPEC-registry schema migrations.

---

## 10. Traceability (추적성)

- Theme: master-v3 §1.1 design commitment #3 (moai differentiation); priority-roadmap T2-DIFF-01.
- Gap rows: none (moai-unique; no CC equivalent in gap matrix).
- Wave 1 sources:
  - findings-wave1-moai-current.md §11 (current SPEC state: 104 SPECs, triple-file pattern, EARS format)
  - `.claude/rules/moai/workflow/spec-workflow.md` (workflow integration points)
  - `.claude/skills/moai-workflow-spec/SKILL.md` (lifecycle transitions, SDD 2025)
- BC-ID: none (additive; existing SPECs unaffected).
- REQ 총 개수: 15 (Ubiquitous 5, Event-Driven 6, State-Driven 2, Optional 2, Unwanted 3 discrete — sum 18, collapsed overlap to 15 unique IDs).
- 예상 AC 개수: 10.
- 코드 구현 예상 경로:
  - `internal/spec/frontmatter.go` (REQ-SPEC-001-001, -002, -005)
  - `internal/spec/registry.go` (REQ-SPEC-001-004)
  - `internal/spec/inheritance.go` (REQ-SPEC-001-010, -011, -012, -013, -020, -021, -030, -040, -041)
  - `internal/spec/trace_index.go` (REQ-SPEC-001-014, -031)
  - `internal/spec/template.go` (REQ-SPEC-001-003, -015, -042)
  - `internal/config/schema/spec_schema.go`
  - `internal/template/templates/.moai/spec-templates/feature-basic/` (3 files)
  - `internal/template/templates/.moai/spec-templates/refactor-safe/` (3 files)
  - Test files: `internal/spec/*_test.go` (≥ 85% coverage).

---

End of SPEC.
