---
id: SPEC-V3R6-WORKFLOW-PLAN-GEARS-ALIGN-001
title: "/moai plan workflow skill bundle GEARS notation alignment"
version: "0.1.3"
status: implemented
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/skills/moai/workflows/plan + internal/template/templates/.claude/skills/moai/workflows/plan"
lifecycle: spec-anchored
tags: "gears, notation-alignment, workflow-plan, skill-body, sprint-10, cohort-4-of-8"
depends_on: [SPEC-V3R6-GEARS-MIGRATION-001, SPEC-V3R6-SKILL-GEARS-ALIGN-001, SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001, SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001]
tier: M
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-25 | manager-spec | Initial plan-phase draft — Sprint 10 GEARS sweep cohort entry SPEC #4 of 8. Cohort progression: SKILL-GEARS-ALIGN-001 closed `ebe492670`, PLAN-AUDITOR-GEARS-ALIGN-001 closed `ebe492670`, FOUNDATION-CORE-GEARS-ALIGN-001 closed `0156c7003`, **WORKFLOW-PLAN-GEARS-ALIGN-001 (THIS)**, downstream: DOCS-SITE-FULL, WORKFLOW-SPEC-EXTRAS, MISC-DOCS, RULES-GO-DOCS. |
| 0.1.1 | 2026-05-25 | manager-spec | iter-2 focused fix per plan-auditor iter-1 PASS 0.867 (NOT skip-eligible). D1 RESOLVED: plan.md M4 staging count 11→10 (Option A — phantom "ownership backfill marker" removed, aligned with acceptance.md AC-WPG-010). D2 RESOLVED: new AC-WPG-011 added with direct grep verification for spec-assembly.md → spec-frontmatter-schema.md cross-link (Option A — closes REQ-WPG-009 trace orphan, traceability 0.85→0.93). D3 RESOLVED: AC-WPG-007 grep pattern anchored to `^Only in .*: \.gitkeep$`. Side-effect updates: spec-assembly.md edit zones 6→7 (cross-link addition counts as edit zone), AC count 10→11 across all 3 files, total edit zones 13→14. Predicted iter-2 plan-auditor: 0.90+ skip-eligible. |
| 0.1.2 | 2026-05-25 | manager-spec | iter-3 mechanical fix per plan-auditor iter-2 PASS-WITH-DEBT 0.873. D_new1 RESOLVED: 4 stale "13 edit zones" sites updated to "14" (spec.md §B.2 line 45 + spec.md §C.2 table spec-assembly.md row + spec.md §C.2 table Total row + plan.md §B.1 + plan.md §B.3 MP-1; iter-2 added 1 NEW edit zone for spec-assembly.md cross-link addition). D_new2 RESOLVED: plan.md §B.3 MP-3 "13 REQs × 10 ACs" → "13 REQs × 11 ACs" (residual stale AC count from iter-1). D_new3 RESOLVED: HISTORY tables added to plan.md AND acceptance.md per Option A (consistency across all 3 artifacts; previously HISTORY existed only in spec.md). No new REQs, no new ACs, no milestone re-decomposition. Predicted iter-3 plan-auditor: 0.92+ skip-eligible (Consistency 0.74→0.92 + Completeness 0.92→0.94). |
| 0.1.3 | 2026-05-25 | manager-docs | Sync-phase completion: status `in-progress → implemented` per Status Transition Ownership Matrix. Version bump, HISTORY entry added. 4-artifact sync_commit_sha backfill + CHANGELOG entry + progress.md §E.4 populated. |

## §A — Goals

The skill bundle owning `/moai plan` orchestration shall present GEARS as the primary notation for SPEC authoring guidance, while preserving EARS as the explicit 6-month backward-compatibility reference for the 88 pre-v3 SPECs that the lint engine continues to accept (legacy window expires 2026-11-22 per `SPEC-V3R6-GEARS-MIGRATION-001`).

When a SPEC author opens any of the 4 in-scope workflow files (`plan.md`, `plan/context-discovery.md`, `plan/clarity-interview.md`, `plan/spec-assembly.md`), the workflow body shall surface GEARS notation as the canonical first-class form. The 4 mirrored template files under `internal/template/templates/.claude/skills/moai/workflows/plan/` shall stay byte-for-byte parity-aligned with the local files so that `moai init` newly-bootstrapped projects receive the same GEARS-first guidance.

## §B — Background

### B.1 Cohort context

Sprint 10 GEARS sweep cohort sequence (8 SPECs total):

1. SPEC-V3R6-SKILL-GEARS-ALIGN-001 (Tier M) — CLOSED `ebe492670` — `moai-workflow-spec` skill bundle (5 local + 5 mirror)
2. SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001 (Tier S) — CLOSED `ebe492670` — `plan-auditor.md` agent body
3. SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001 (Tier M) — CLOSED `0156c7003` — `moai-foundation-core` skill bundle (10 local + 10 mirror)
4. **SPEC-V3R6-WORKFLOW-PLAN-GEARS-ALIGN-001 (Tier M) — THIS SPEC** — `/moai plan` skill bundle (4 local + 4 mirror)
5-8. DOCS-SITE-FULL / WORKFLOW-SPEC-EXTRAS / MISC-DOCS / RULES-GO-DOCS (downstream)

### B.2 Why the `/moai plan` skill bundle matters

`/moai plan` is the entry point through which every SPEC author begins requirement authoring. Three of its four files contain explicit EARS notation references (plan.md ×4, clarity-interview.md ×3, spec-assembly.md ×6 = 13 total). When the workflow body describes EARS as the canonical SPEC notation form, downstream authors consult that text first, propagating EARS notation into new SPEC bodies even though `SPEC-V3R6-GEARS-MIGRATION-001` v0.2.0 has already made GEARS canonical. This SPEC closes that discrepancy at the orchestration-skill body layer.

### B.3 Template mirror parity rationale

The `internal/template/templates/` mirror is what `moai init` deploys into freshly-created user projects. If local body says GEARS-first but the template mirror still says EARS-first, every new project pulled via `moai init` after the 6-month window expires will receive stale guidance. Mirror parity is enforced byte-for-byte during sync-phase artifact emission.

## §C — Scope

### §C.1 In-scope file inventory (4 local + 4 mirror = 8 files)

**Local files**:
- `.claude/skills/moai/workflows/plan.md` (7,257 bytes, 4 EARS references confirmed)
- `.claude/skills/moai/workflows/plan/clarity-interview.md` (11,531 bytes, 3 EARS references)
- `.claude/skills/moai/workflows/plan/context-discovery.md` (5,721 bytes, 0 notation references — out of scope for content edits, mirror parity only)
- `.claude/skills/moai/workflows/plan/spec-assembly.md` (28,423 bytes, 6 EARS references)

**Template mirror files** (byte-for-byte parity targets):
- `internal/template/templates/.claude/skills/moai/workflows/plan.md`
- `internal/template/templates/.claude/skills/moai/workflows/plan/clarity-interview.md`
- `internal/template/templates/.claude/skills/moai/workflows/plan/context-discovery.md`
- `internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md`

**Discovery variance from estimate** (transparently documented per L46 attribution discipline): initial paste-ready estimate said "9 files"; actual discovered count is **8 .md files + 1 .gitkeep placeholder in template mirror plan/ directory**. The `.gitkeep` is a presentational empty marker (0 bytes) that exists only in the template mirror; it has no notation content and requires no edits but will not be deleted. Effective edit scope is **8 .md files**.

### §C.2 Notation reference distribution

| File | EARS refs | GEARS refs | Estimated edit zones |
|------|-----------|------------|----------------------|
| `plan.md` | 4 | 0 | 4 (description block + intro + 2 phase routing table cells) |
| `clarity-interview.md` | 3 | 0 | 3 (Phase 1B output description + 2 transition headers) |
| `spec-assembly.md` | 6 | 0 | 7 (frontmatter checklist + REQ format spec + AC traceability + quality gate refs + NEW cross-link to spec-frontmatter-schema.md SSOT per REQ-WPG-009 / AC-WPG-011 added in iter-2) |
| `context-discovery.md` | 0 | 0 | 0 (mirror parity only) |
| **Total** | **13** | **0** | **14** |

### §C.3 Out-of-scope

The following are explicitly NOT in scope for this SPEC and shall remain untouched:

- `.claude/skills/moai-foundation-core/*` (closed in `0156c7003` by SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001)
- `.claude/skills/moai-workflow-spec/*` (closed in `ebe492670` by SPEC-V3R6-SKILL-GEARS-ALIGN-001)
- `.claude/agents/meta/plan-auditor.md` (closed in `ebe492670` by SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001)
- `internal/spec/lint.go` and `internal/spec/lint_*_test.go` (lint code stable per cohort precedent — owned by SPEC-V3R6-GEARS-MIGRATION-001)
- `.claude/skills/moai/workflows/run.md`, `sync.md`, `mx.md`, `loop.md`, `fix.md`, `clean.md`, `feedback.md`, `project.md` (downstream cohort #5-#8 territory)
- `.claude/skills/moai/team/plan.md` (team-mode plan skill — separate scope, follows-on under WORKFLOW-SPEC-EXTRAS cohort #6 candidate)
- The 88 pre-v3 SPEC bodies (preserved verbatim through the 6-month legacy window per `SPEC-V3R6-GEARS-MIGRATION-001`)

## §D — Requirements

GEARS notation self-dogfood targets ≥80% of REQ entries below. Mix:

- Ubiquitous: `The <subject> shall <behavior>`
- Event-driven: `When <event> the <subject> shall <behavior>`
- State-driven: `While <state> the <subject> shall <behavior>`
- Capability gate: `Where <capability> the <subject> shall <behavior>`
- Event-detected (unwanted): `When <undesired-condition-detected> the <subject> shall <response>` (replaces deprecated IF/THEN)
- Compound: `[Where ...] [While ...] [When ...] the <subject> shall <behavior>`

### §D.1 Ubiquitous requirements

**REQ-WPG-001** — The `/moai plan` workflow skill body (`plan.md` entry file) shall present GEARS as the canonical SPEC authoring notation in its description frontmatter and in its intro paragraph, replacing the current EARS-only phrasing.

**REQ-WPG-002** — The `clarity-interview.md` sub-skill shall describe the Phase 1B output as "GEARS-format requirements (EARS retained as legacy reference)" wherever it currently describes the output as "EARS-format requirements".

**REQ-WPG-003** — The `spec-assembly.md` sub-skill shall use the phrase "GEARS structure with the 5 GEARS patterns (Ubiquitous, Event-driven `When`, State-driven `While`, Capability-gate `Where`, Event-detected unwanted)" wherever it currently uses the phrase "EARS structure with all 5 requirement types".

**REQ-WPG-004** — Every modified workflow file shall cross-link the canonical GEARS authoring reference at `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format so that a downstream SPEC author can navigate from `/moai plan` to the full GEARS specification in one hop.

### §D.2 Event-driven requirements (`When` form)

**REQ-WPG-005** — When the `manager-spec` subagent renders Phase 1B output per the `clarity-interview.md` flow, the orchestration skill body shall instruct it to produce GEARS-notation requirements by default and the body shall not instruct it to emit EARS notation as the default.

**REQ-WPG-006** — When a SPEC author opens any one of the 4 in-scope workflow files, the body shall display the GEARS notation reference table or a GEARS-first phrase before any EARS reference appears in reading order.

**REQ-WPG-007** — When the `manager-docs` subagent runs sync-phase artifact emission for this SPEC, the orchestrator shall verify that each modified local file has an identical (byte-for-byte) mirror at `internal/template/templates/.claude/skills/moai/workflows/plan/<path>`.

### §D.3 State-driven requirements (`While` form)

**REQ-WPG-008** — While the 6-month backward-compatibility window for `SPEC-V3R6-GEARS-MIGRATION-001` v0.2.0 remains active (until 2026-11-22), the workflow body shall retain the EARS notation reference as an explicit legacy footnote — it shall not delete EARS notation references outright.

**REQ-WPG-009** — While the `spec-assembly.md` Phase 2 SPEC document creation section describes the frontmatter checklist, the body shall reference the canonical schema at `.claude/rules/moai/development/spec-frontmatter-schema.md` rather than restate the 12 canonical fields inline.

### §D.4 Capability-gate requirements (`Where` form)

**REQ-WPG-010** — Where the `/moai plan` workflow operates against a project that contains 1+ SPEC bodies authored in legacy EARS notation (pre-v3 SPECs in `.moai/specs/`), the workflow body shall not warn the SPEC author about the EARS notation — legacy SPEC notation is valid until 2026-11-22 and out-of-band lint warnings are emitted by `internal/spec/lint.go` `LegacyEARSKeyword` rule, not by the workflow skill body.

### §D.5 Event-detected unwanted requirements (`When <undesired-condition>` form)

**REQ-WPG-011** — When a SPEC author attempts to add a new SPEC body using the deprecated `IF/THEN` modality, the workflow body shall not display authoring guidance suggesting `IF/THEN` is acceptable for new SPECs (the workflow body shall present `When <event-detected>` as the canonical event-detected pattern, per the deprecated form noted in PLAN-AUDITOR-GEARS-ALIGN-001 M3).

**REQ-WPG-012** — When the sync-phase mirror parity check detects byte-level divergence between any local file and its template mirror counterpart, the orchestrator shall halt the sync-phase commit and surface the divergence to the user (not silently overwrite either side).

### §D.6 Compound requirements

**REQ-WPG-013** — `[Where the project has been initialized with v3.0.0 or later] [When a SPEC author runs /moai plan]` the workflow shall describe GEARS as the primary SPEC notation in the first user-facing display turn — `[While the legacy EARS-format window remains active]` it shall additionally surface a 1-line footnote naming EARS notation as the legacy reference and citing the SSOT at `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format.

**GEARS notation self-dogfood tally**: 13 REQs total. GEARS-form REQs: 1 Ubiquitous (REQ-WPG-001/002/003/004 = 4 Ubiquitous) + 3 Event-driven `When` (REQ-WPG-005/006/007) + 2 State-driven `While` (REQ-WPG-008/009) + 1 Capability-gate `Where` (REQ-WPG-010) + 2 Event-detected unwanted (REQ-WPG-011/012) + 1 Compound (REQ-WPG-013) = **13/13 = 100% GEARS notation compliance** (exceeds ≥80% target).

## §E — Acceptance Criteria Reference

See `acceptance.md` for the full AC-WPG-001..AC-WPG-011 mandatory matrix (11 ACs; iter-2 added AC-WPG-011 closing REQ-WPG-009 trace orphan via direct grep verification), traceability mapping (REQ-WPG-XXX → AC-WPG-XXX), and predecessor pattern fidelity table.

## §F — Non-Goals

- Modifying `internal/spec/lint.go` or extending the `LegacyEARSKeyword` lint rule (owned by SPEC-V3R6-GEARS-MIGRATION-001)
- Migrating any of the 88 pre-v3 SPEC bodies from EARS to GEARS notation (preserved verbatim per the 6-month backward-compat window)
- Modifying the `moai-workflow-spec` skill bundle (closed by SPEC-V3R6-SKILL-GEARS-ALIGN-001, ebe492670)
- Modifying the `moai-foundation-core` skill bundle (closed by SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001, 0156c7003)
- Modifying the `plan-auditor` agent body (closed by SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001, ebe492670)
- Modifying the `.claude/skills/moai/team/plan.md` team-mode workflow (out of this cohort scope; downstream WORKFLOW-SPEC-EXTRAS cohort #6 candidate)
- Adding new `/moai plan` features, sub-phases, or Tooling — this SPEC is a pure notation-alignment edit cycle

## §G — Risks

### G.1 Mirror parity drift (Med)

The local `.claude/skills/moai/workflows/plan/` directory and its template mirror at `internal/template/templates/.claude/skills/moai/workflows/plan/` must stay byte-for-byte identical for the in-scope 4 .md files. Risk: an asymmetric edit (e.g., updating local file but forgetting the mirror) causes `moai init` of new projects to receive stale guidance after the 6-month window expires. **Mitigation**: Milestone M3 enforces explicit `diff -q` parity check before sync-phase staging; AC-WPG-007 codifies the verification.

### G.2 Cohort SSOT cross-attribution leakage (Low-Med)

L48 SSOT discipline — Sprint 10 cohort has SKILL-GEARS-ALIGN-001 + PLAN-AUDITOR-GEARS-ALIGN-001 + FOUNDATION-CORE-GEARS-ALIGN-001 already closed. If this SPEC's manager-develop spawn accidentally touches `.claude/skills/moai-foundation-core/*` or `.claude/skills/moai-workflow-spec/*` files (their domains), it violates the L48 SSOT path discipline and the L46 attribution discipline simultaneously. **Mitigation**: Milestone M1 uses path-specific `git add` (8 exact paths), M4 enforces `git diff --cached --name-only | sort -u | wc -l` ≥ 8 pre-commit assertion (L59 staging area scope verification from COORD-001 sync).

### G.3 IF/THEN deprecated modality re-introduction (Low)

Risk: a sloppy GEARS rewrite at any of the 7 spec-assembly.md edit zones may accidentally introduce or fail to remove `IF/THEN` patterns. **Mitigation**: AC-WPG-006 includes a `grep -E 'IF .* THEN'` sentinel scan post-edit; AC-WPG-009 confirms `LegacyEARSKeyword` lint rule self-dogfood does not regress (lint clean across the 88 pre-v3 SPECs + 0 NEW IF/THEN introductions in the modified files).

### G.4 Plan-auditor below skip threshold (Low)

Risk: Tier M plan-auditor score may fall below 0.85 PASS threshold or 0.90 skip-eligible threshold. **Mitigation**: 13 REQs (100% GEARS self-dogfood), 11 mandatory ACs (iter-2 added AC-WPG-011 closing REQ-WPG-009 trace orphan), explicit traceability matrix with 100% direct coverage, clear scope boundaries, explicit out-of-scope enumeration. Target plan-auditor 0.90+ (skip-eligible).

### G.5 Discovered count variance below estimate (Low — already mitigated)

Initial paste-ready estimate said "9 files" but actual discovered count is "8 .md files + 1 .gitkeep". Variance is -1 (-11%) below estimate. **Mitigation**: explicitly documented in §C.1 with full transparency per L46 attribution discipline. Effective edit scope is 8 .md files; the .gitkeep is preserved untouched.

## §H — Cross-References

- `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format (canonical GEARS authoring SSOT — cross-link target for REQ-WPG-004)
- `.claude/skills/moai-foundation-core/SKILL.md` § GEARS Format (concise GEARS reference, closed by FOUNDATION-CORE-GEARS-ALIGN-001 `0156c7003`)
- `.claude/agents/meta/plan-auditor.md` § MP-2 GEARS/EARS label (closed by PLAN-AUDITOR-GEARS-ALIGN-001 `ebe492670`)
- `SPEC-V3R6-GEARS-MIGRATION-001` v0.2.0 (PR #1046) — canonical lint engine + 6-month backward-compat window definition
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Canonical 12 Required Fields (referenced by REQ-WPG-009)
- `internal/spec/lint.go` `LegacyEARSKeyword` rule (out-of-band lint warning emitter referenced by REQ-WPG-010)
- `progress.md` § E.1 Plan-phase Audit-Ready Signal (plan_commit_sha backfill marker)

## Exclusions (What NOT to Build)

- No new `/moai plan` phases, sub-phases, decision points, or AskUserQuestion rounds shall be added. Pure notation-alignment edit only.
- No EARS notation references shall be entirely deleted from the workflow body — EARS retention is mandated by REQ-WPG-008 (6-month legacy window).
- No `internal/spec/lint.go` modifications. Lint code stable per cohort precedent.
- No team-mode `.claude/skills/moai/team/plan.md` edits. Out of cohort scope.
- No predecessor SPEC body modifications (spec.md/plan.md/acceptance.md/progress.md of GEARS-MIGRATION-001, SKILL-GEARS-ALIGN-001, PLAN-AUDITOR-GEARS-ALIGN-001, FOUNDATION-CORE-GEARS-ALIGN-001). L48 SSOT discipline.
- No CHANGELOG entry in plan-phase — CHANGELOG is owned by manager-docs sync-phase per Status Transition Ownership Matrix.
