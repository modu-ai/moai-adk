---
id: SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001
title: "moai-foundation-core SKILL bundle을 GEARS 우선 가이드로 정렬"
version: "0.1.1"
status: completed
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/skills/moai-foundation-core, internal/template/templates/.claude/skills/moai-foundation-core"
lifecycle: spec-anchored
tags: "gears, ears, skill, foundation, core, guide, alignment, dogfooding, sprint-10, v3.0.0"
tier: M
issue_number: null
depends_on: [SPEC-V3R6-GEARS-MIGRATION-001, SPEC-V3R6-SKILL-GEARS-ALIGN-001, SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001]
related_specs: [SPEC-V3R6-GEARS-MIGRATION-001, SPEC-V3R6-SKILL-GEARS-ALIGN-001, SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001]
sync_commit_sha: "a853f2954"
---

# SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001 — moai-foundation-core SKILL Bundle GEARS 우선 정렬

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-25 | manager-spec | Initial draft — Sprint 10 GEARS sweep cohort entry SPEC #3 of 8. Predecessor SPECs SPEC-V3R6-SKILL-GEARS-ALIGN-001 (Tier M, commits `1f3a734d8+353150294+7b8f939b7+263db0600` 4-phase CLOSED, 5 authoring guide files + 5 template mirrors with 12 GEARS-notation REQs self-dogfooded + 13/13 ACs PASS + plan-auditor 0.892 not skip-eligible) and SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001 (Tier S minimal, commits `906f9285e+6366d7428+7cd25e386+ebe492670` 4-phase CLOSED, plan-auditor.md MP-2 EARS→EARS/GEARS + M3 Unwanted IF/THEN→`shall not` + M2 failure modes IF/THEN deprecation entry + #gears-notation cross-link, 9 REQs + 8 ACs + plan-auditor 0.913 skip-eligible) established the migration pattern. This SPEC extends the cohort to the `moai-foundation-core` SKILL bundle: 1 SKILL.md entry point + 17 modules + 2 references = 20 files local + 20 template mirrors = **40 files total scope** (significantly more than initial "10 files" estimate; documented honestly per L46 attribution discipline). Baseline scan: 30 EARS/GEARS/shall/WHEN/WHILE/WHERE/IF…THEN references discovered across local files; `modules/spec-ears-format.md` already carries the v3.0.0 DEPRECATED banner pointing at the canonical GEARS guide (no further edit needed) but downstream loaders (`SKILL.md` §SPEC-First DDD + `references/examples.md` + `references/reference.md` + `modules/spec-first-ddd.md` + `modules/progressive-disclosure.md` + `modules/commands-reference.md`) still describe EARS as the primary notation, creating drift between the canonical GEARS authoritative source (`.claude/skills/moai-workflow-spec/SKILL.md` § "GEARS Format") and the foundation tier. Tier M markdown-only edits across ≤7 files (those with EARS references) + their template mirrors. Self-dogfooding: this SPEC's REQs are written in GEARS notation. |
| 0.1.1 | 2026-05-25 | orchestrator | L60 retroactive Mx-phase backfill — `status: implemented → completed`. progress.md §A Lifecycle Mx row (L31 "Mx-phase COMPLETE (SKIP) — `(this commit)` placeholder") + §E.5 Mx-phase Audit-Ready Signal (L207+ EVALUATE-SKIP per mx-tag-protocol.md §a, 0 .go files) were already authored but `mx_commit_sha` frontmatter field never backfilled (L60 chicken-and-egg) and spec.md status drift persisted (L67 manager-docs scope-creep). Cross-file consistency restored. Body 변경 없음. sync `a853f2954`; this commit은 atomic close terminator, mx 본문 progress.md self-reference. |

## 1. Goal

Align the `moai-foundation-core` SKILL bundle (SKILL.md + 17 modules + 2 references = 20 local files + 20 template mirrors = 40 files) with the GEARS notation canonicalized in predecessor SPECs SPEC-V3R6-GEARS-MIGRATION-001 v0.2.0 (lint engine + 4-locale docs-site), SPEC-V3R6-SKILL-GEARS-ALIGN-001 v0.2.0 (moai-workflow-spec authoring guide), and SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001 v0.2.0 (plan-auditor agent), so that the foundation-tier SKILL bundle presents GEARS as the primary notation while the 6-month EARS backward-compatibility window remains intact for the 88 legacy SPECs.

### 1.1 GEARS Canonical Definition (inherited from SPEC-V3R6-GEARS-MIGRATION-001 §1)

- **Five patterns**: Ubiquitous (unchanged) / `When <event>` (unchanged trigger semantics) / `While <state>` (unchanged, promoted to first-class) / `Where <capability / feature flag / static config>` (reframed — capability gate, no longer "Optional") / **`IF/THEN` DEPRECATED → `When <event-detected>`**
- **Unified compound pattern**: `[Where ...][While ...][When ...] The <subject> shall <behavior>`
- **Generalized subject**: `<subject>` replaces hardcoded "the system" — any noun (system, component, service, agent, function, artifact). 88 existing SPECs retain "The system" for readability; NEW SPECs MAY use generalized subjects
- **Lint behavior**: `LegacyEARSKeyword` warning emits on `IF/THEN` (warning non-strict / error `moai spec lint --strict`). Backward-compat window: 6 months from v3.0.0 release (expires **2026-11-22**)
- **Cross-reference (canonical authoring guide)**: `.claude/skills/moai-workflow-spec/SKILL.md` § "GEARS Format" (current notation) and § "EARS Format" (legacy reference) — established by SPEC-V3R6-SKILL-GEARS-ALIGN-001
- **Cross-reference (docs-site, 4-locale)**: `https://adk.mo.ai.kr/{en,ko,ja,zh}/workflow-commands/moai-plan/#gears-notation`

### 1.2 Out of Scope Declaration

This SPEC is a **guide alignment** SPEC for the foundation-tier SKILL bundle. It does NOT:

#### 1.2.1 Out of Scope (Exclusions)

- **EXC-FCG-001**: Modify `internal/spec/lint.go` or any Go source code (canonical migration done in SPEC-V3R6-GEARS-MIGRATION-001 M2).
- **EXC-FCG-002**: Add/modify `internal/spec/lint_test.go` (done in predecessor M2).
- **EXC-FCG-003**: Touch 4-locale `docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-plan.md` (done in SPEC-V3R6-GEARS-MIGRATION-001 M3).
- **EXC-FCG-004**: Modify `.claude/skills/moai-workflow-spec/**` (handled by SPEC-V3R6-SKILL-GEARS-ALIGN-001 v0.2.0).
- **EXC-FCG-005**: Modify `.claude/agents/meta/plan-auditor.md` (handled by SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001 v0.2.0).
- **EXC-FCG-006**: Modify `.claude/agents/core/manager-spec.md` (handled by SPEC-V3R6-SKILL-GEARS-ALIGN-001 v0.2.0).
- **EXC-FCG-007**: Rewrite the 88 existing SPECs to GEARS (deferred to provisional `SPEC-V3R6-GEARS-SWEEP-001`).
- **EXC-FCG-008**: Modify `modules/spec-ears-format.md` body — the file already carries the v3.0.0 DEPRECATED banner (lines 9-15) pointing at the canonical GEARS guide; the file is intentionally retained as legacy reference for the 6-month backward-compatibility window and MUST remain as-is. Only the SKILL.md cross-link to it MAY be re-labeled "(legacy reference, deprecated)" without modifying the target file itself.
- **EXC-FCG-009**: Modify the predecessor SPECs' spec.md / plan.md / acceptance.md bodies (SKILL-GEARS-ALIGN-001, PLAN-AUDITOR-GEARS-ALIGN-001, GEARS-MIGRATION-001) — SSOT preservation per Status Transition Ownership Matrix.
- **EXC-FCG-010**: Adjust scope to downstream Sprint 10 cohort SPECs (WORKFLOW-PLAN, DOCS-SITE-FULL, WORKFLOW-SPEC-EXTRAS, MISC-DOCS, RULES-GO-DOCS) — each owned by its own SPEC.

## 2. Why

### 2.1 Drift Risk (predecessors implemented but foundation tier un-updated)

The canonical GEARS migration (SPEC-V3R6-GEARS-MIGRATION-001 v0.2.0, PR #1046 merged `134a43fac` 2026-05-22) updated lint engine + 4-locale docs-site. Two follow-up SPECs SPEC-V3R6-SKILL-GEARS-ALIGN-001 v0.2.0 (5 authoring guide files in `moai-workflow-spec`) and SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001 v0.2.0 (`plan-auditor.md`) extended GEARS to the workflow-tier authoring layer. However, the foundation-tier `moai-foundation-core` SKILL bundle — which is consumed by **every** agent and skill that depends on the foundation principles — still presents EARS as the primary notation:

- `SKILL.md` § SPEC-First DDD describes "EARS Format" verbatim (line 116, 122) with five patterns Ubiquitous/Event-driven/State-driven/Unwanted/Optional, not GEARS
- `references/examples.md` § "EARS Format:" block (lines 65-69) gives EARS-labeled examples with "System shall …" (hardcoded subject)
- `references/reference.md` line 20 declares format = "EARS (Event-Action-Response-State)" (incorrect EARS expansion — actual EARS = "Easy Approach to Requirements Syntax"); lines 198, 377, 410, 447 perpetuate EARS-only guidance
- `modules/spec-first-ddd.md` lines 29-34 give EARS pattern table; lines 37, 153, 161 cross-reference EARS-only resources
- `modules/progressive-disclosure.md` lines 184, 201 anchor EARS as the SPEC format
- `modules/commands-reference.md` lines 78, 90, 94-102 describe EARS as the `/moai:1-plan` output format

A new agent or skill author consulting `moai-foundation-core` for SPEC authoring patterns receives EARS-first guidance, contradicting the canonical authoring guide (`moai-workflow-spec`) that now presents GEARS-first. This creates **author confusion** at the foundation layer — the single most-loaded skill in the agent context.

### 2.2 Dogfooding Pressure (cohort entry #3 of 8)

Sprint 10 GEARS sweep cohort (75+ file pairs, 8 SPECs, 6-month backward-compat window expiring 2026-11-22) requires the foundation tier to align early because downstream Sprint 10 SPECs (WORKFLOW-PLAN, DOCS-SITE-FULL, WORKFLOW-SPEC-EXTRAS, MISC-DOCS, RULES-GO-DOCS) reference `moai-foundation-core` for foundational patterns. If foundation tier remains EARS-first, downstream alignments will repeatedly reference an out-of-date foundation, creating circular drift.

### 2.3 Cost Justification

- Tier M scope (~40 file edits = 20 local + 20 template mirrors).
- Markdown-only changes; no Go source / no tests / no docs-site.
- 1-pass implementation expected (predecessor SKILL-GEARS-ALIGN-001 was also Tier M 1-pass at 0.892 plan-auditor; PLAN-AUDITOR-GEARS-ALIGN-001 was Tier S 1-pass at 0.913 skip-eligible).
- Plan-auditor target ≥0.85 (Tier M PASS threshold), skip-eligible 0.90+ achievable if scope discipline maintained.

## 3. Requirements (GEARS notation — self-dogfooding)

[HARD] All REQ-FCG-XXX in this SPEC use GEARS notation. ≥80% target met (12/12 = 100% GEARS, 0% legacy `IF/THEN`).

### REQ-FCG-001 (Ubiquitous)

The `moai-foundation-core` SKILL bundle shall present GEARS as the primary SPEC notation in any text describing requirement format, with the canonical authoring guide pointer to `.claude/skills/moai-workflow-spec/SKILL.md` § "GEARS Format".

### REQ-FCG-002 (Event-driven)

**When** a downstream agent or skill author loads `moai-foundation-core/SKILL.md` § "SPEC-First DDD - Development Workflow", the section shall describe the five GEARS patterns (Ubiquitous / `When <event>` / `While <state>` / `Where <capability>` / `When <event-detected>` replacing IF/THEN) before any EARS-legacy reference.

### REQ-FCG-003 (Event-driven)

**When** a downstream consumer reads `moai-foundation-core/references/reference.md` line 20 (current text "Format: EARS (Event-Action-Response-State)"), the line shall read "Format: GEARS (Generalized EARS) — primary notation; EARS retained as legacy reference for the 6-month backward-compatibility window" and the incorrect "Event-Action-Response-State" expansion shall be corrected to "Easy Approach to Requirements Syntax" wherever EARS is referenced as legacy.

### REQ-FCG-004 (State-driven, generalized subject)

**While** a Sprint 10 author writes a new SPEC referencing `moai-foundation-core` for notation guidance, the bundle shall surface GEARS examples that demonstrate generalized `<subject>` substitution (e.g., "The agent shall …", "The component shall …") alongside the legacy "The system shall …" form, so the author learns both forms.

### REQ-FCG-005 (Where, capability gate)

**Where** the `moai-foundation-core/modules/spec-ears-format.md` file already carries the v3.0.0 DEPRECATED banner (verified at lines 9-15 of current content) pointing at the canonical GEARS guide, the file body shall remain unmodified and cross-references from sibling files MAY add the inline label "(legacy reference, deprecated — see GEARS Format guide)" when linking to it.

### REQ-FCG-006 (Event-driven)

**When** `moai-foundation-core/references/examples.md` § "EARS Format:" block (lines 65-69 of current content) is updated, it shall be re-labeled "GEARS Format (current):" with the five GEARS patterns including at least one generalized-subject example and one `Where <capability>` example, with an "EARS Format (legacy reference, 6-month backward-compat):" sub-block retained underneath for the 88 legacy SPECs.

### REQ-FCG-007 (Ubiquitous, generalized subject)

The bundle shall not introduce any new `IF <condition> THEN <action>` modality in updated content; any pre-existing `IF/THEN` text in the touched files (lines other than those inside protected legacy reference blocks or unmodified bodies per REQ-FCG-005) shall be rewritten to `When <event-detected>` GEARS form.

### REQ-FCG-008 (Event-driven)

**When** `moai-foundation-core/modules/spec-first-ddd.md` lines 29-34 (EARS pattern table) and lines 37/153/161 (cross-references) are updated, the pattern table shall use GEARS naming (Ubiquitous / Event-driven `When` / State-driven `While` / Capability gate `Where` / `When <event-detected>` replacing the legacy `IF/THEN` row) and cross-references shall point primarily at `.claude/skills/moai-workflow-spec/SKILL.md` § "GEARS Format" with `modules/spec-ears-format.md` labeled "(legacy reference)".

### REQ-FCG-009 (Unwanted)

The bundle shall not modify `internal/spec/lint.go` or any Go source/test file; the GEARS lint canonicalization is owned by SPEC-V3R6-GEARS-MIGRATION-001 v0.2.0 and out of this SPEC's scope per EXC-FCG-001 / EXC-FCG-002.

### REQ-FCG-010 (Where, capability)

**Where** template mirror parity is required (every `.claude/skills/moai-foundation-core/<path>` edit MUST be mirrored to `internal/template/templates/.claude/skills/moai-foundation-core/<path>`), the run-phase shall apply identical content to both the local SKILL bundle and the template mirror, so that fresh `moai init` invocations receive GEARS-aligned content from `make build`-regenerated `embedded.go`.

### REQ-FCG-011 (Ubiquitous)

The bundle shall preserve the v3.0.0 DEPRECATED banner in `modules/spec-ears-format.md` (lines 9-15) verbatim and shall not add any text suggesting EARS removal or 6-month-window early termination.

### REQ-FCG-012 (Event-driven)

**When** the run-phase implementation completes, a self-lint pass shall confirm zero new `LegacyEARSKeyword` warnings introduced by this SPEC's edits (acceptable: pre-existing warnings on lines of `modules/spec-ears-format.md` legacy reference body that are intentionally preserved per REQ-FCG-005).

## 4. Risks + Mitigations

### Risk R1 — Self-lint failure on this SPEC

The 12 GEARS REQs above use generalized subjects and `When <event-detected>` for what could be read as conditional. Mitigation: each REQ explicitly tagged with its GEARS pattern in the heading (e.g., "REQ-FCG-002 (Event-driven)"). Plan-auditor expected to recognize the self-dogfood pattern (precedent in SKILL-GEARS-ALIGN-001 and PLAN-AUDITOR-GEARS-ALIGN-001).

### Risk R2 — Template-mirror drift after edits

40-file scope (20 local + 20 mirror) increases drift risk. Mitigation: REQ-FCG-010 explicit + plan.md M5 verification step (`diff -r .claude/skills/moai-foundation-core/ internal/template/templates/.claude/skills/moai-foundation-core/` MUST emit zero diffs after edits). `make build` regeneration of `embedded.go` is run-phase responsibility post-edits.

### Risk R3 — Excessive scope creep into Sprint 10 cohort SPECs

Tempting to fix downstream `moai-workflow-*` or `moai-domain-*` skills while editing foundation. Mitigation: EXC-FCG-010 explicit + L48 SSOT discipline + scope assertion in plan.md §C.1.

### Risk R4 — Discovered file count significantly exceeds initial estimate

User prompt said "10 files"; actual discovered scope = 40 files. Mitigation: L46 attribution discipline — discovered count documented honestly in plan.md §C.1, milestone decomposition adjusted to match actual scope (M2 expanded to handle ≤7 files needing edits vs initial estimate of 10).

### Risk R5 — Cross-cohort coordination with downstream SPECs

Downstream Sprint 10 SPECs may reference patterns currently in foundation. Mitigation: cohort sequencing per Sprint 10 memo (entry #3 of 8; downstream #4-8 land after this SPEC closes).

### Risk R6 — `modules/spec-ears-format.md` already partially aligned

Predecessor SPEC-V3R6-SKILL-GEARS-ALIGN-001 v0.2.0 already updated this file's banner (lines 9-15 carry the v3.0.0 DEPRECATED notice). Mitigation: EXC-FCG-008 + REQ-FCG-005 + REQ-FCG-011 — file body MUST NOT be modified; only sibling cross-references to it MAY be re-labeled "(legacy reference, deprecated)".

## 5. Dependencies + Cross-References

- **depends_on**: SPEC-V3R6-GEARS-MIGRATION-001 (canonical lint engine + 4-locale docs-site, v0.2.0 PR #1046 merged `134a43fac` 2026-05-22), SPEC-V3R6-SKILL-GEARS-ALIGN-001 (moai-workflow-spec authoring guide, v0.2.0 4-phase CLOSED 2026-05-25 `ebe492670`), SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001 (plan-auditor.md alignment, v0.2.0 4-phase CLOSED 2026-05-25 `ebe492670`).
- **related_specs**: same as depends_on (predecessor cohort).
- **Sprint 10 cohort downstream (not depends_on)**: SPEC-V3R6-WORKFLOW-PLAN-GEARS-ALIGN-001, SPEC-V3R6-DOCS-SITE-FULL-GEARS-ALIGN-001, SPEC-V3R6-WORKFLOW-SPEC-EXTRAS-GEARS-ALIGN-001, SPEC-V3R6-MISC-DOCS-GEARS-ALIGN-001, SPEC-V3R6-RULES-GO-DOCS-GEARS-ALIGN-001 (provisional names per Sprint 10 cohort memo; each landed after this SPEC closes).

### 5.1 Working-Tree Baseline Snapshot (2026-05-25 plan-phase entry)

- Local file count: 20 (1 SKILL.md + 17 modules + 2 references)
- Template mirror file count: 20 (1:1 parity verified via `find` 2026-05-25)
- Total scope: 40 files
- Discovered EARS/GEARS reference count (local only): 30 matches across 8 distinct files
- Files needing edit (preliminary): SKILL.md, references/examples.md, references/reference.md, modules/spec-first-ddd.md, modules/progressive-disclosure.md, modules/commands-reference.md, modules/INDEX.md (cross-ref labels only) = 7 files local + 7 template mirrors = **14 files actually modified** (subset of 40-file scope)
- Files intentionally NOT edited: SKILL.md (already aligned in part), `modules/spec-ears-format.md` (per EXC-FCG-008 — body preserved), modules/trust-*.md (no SPEC notation content), modules/delegation-*.md (no SPEC notation content), modules/token-optimization.md (no SPEC notation content), modules/modular-system.md (no SPEC notation content), modules/agents-reference.md (no SPEC notation content), modules/execution-rules.md (no SPEC notation content), modules/patterns.md (verify in M2 — likely no edit)

Final M2 edit set finalized at run-phase boundary after `grep -rn` re-confirmation.

## 6. References

- Canonical GEARS authoring guide: `.claude/skills/moai-workflow-spec/SKILL.md` § "GEARS Format" (established by SPEC-V3R6-SKILL-GEARS-ALIGN-001 v0.2.0)
- Canonical lint behavior: `internal/spec/lint.go` `LegacyEARSKeyword` rule + `internal/spec/lint_test.go` (established by SPEC-V3R6-GEARS-MIGRATION-001 v0.2.0)
- 4-locale docs-site GEARS migration guide: `https://adk.mo.ai.kr/{en,ko,ja,zh}/workflow-commands/moai-plan/#gears-notation`
- plan-auditor MP-2 label updated to EARS/GEARS by SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001 v0.2.0
- Sprint 10 GEARS sweep cohort memo: see MEMORY.md entry "Sprint 10 GEARS sweep cohort 2/8 close + 6 SPECs paste-ready" (2026-05-25)
- Backward-compatibility window expiry: 2026-11-22 (6 months from v3.0.0 release)
- Status Transition Ownership Matrix SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix
- SPEC ID Pre-Write Self-Check Protocol: `.claude/agents/core/manager-spec.md` § Step 4
