---
id: SPEC-V3R6-SKILL-GEARS-ALIGN-001
title: "moai-workflow-spec SKILL을 GEARS 우선 가이드로 정렬"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/skills/moai-workflow-spec, .claude/skills/moai-foundation-core/modules, .claude/agents/core, internal/template/templates"
lifecycle: spec-anchored
tags: "gears, ears, skill, guide, alignment, dogfooding, wave-6, v3.0.0"
tier: M
issue_number: null
depends_on: [SPEC-V3R6-GEARS-MIGRATION-001]
related_specs: [SPEC-V3R6-GEARS-MIGRATION-001]
---

# SPEC-V3R6-SKILL-GEARS-ALIGN-001 — moai-workflow-spec SKILL GEARS 우선 정렬

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-23 | manager-spec | Initial draft — Wave 6 follow-up SPEC to predecessor SPEC-V3R6-GEARS-MIGRATION-001 (status: implemented v0.2.0, PR #1046 merged `134a43fac` 2026-05-22). Predecessor closed lint engine (`internal/spec/lint.go` `LegacyEARSKeyword`) + 4-locale docs-site migration guide. Predecessor did NOT update authoring guides (`.claude/skills/moai-workflow-spec/SKILL.md` + references + `moai-foundation-core/modules/spec-ears-format.md` + `manager-spec.md`). Baseline 50 EARS refs across 5 files (SKILL.md 18 + reference.md 13 + examples.md 1 + spec-ears-format.md 4 + manager-spec.md 14). Authors writing new SPECs receive EARS-only guidance and continue producing IF/THEN REQs that trigger `LegacyEARSKeyword` warnings. Tier M (3 artifacts) — markdown-only edits across 5 files + 5 template mirrors. Self-dogfooding: this SPEC's REQs are written in GEARS notation as the canonical example. |

## 1. Goal

Align the 5 SPEC-authoring guide artifacts (`.claude/skills/moai-workflow-spec/SKILL.md`, `.claude/skills/moai-workflow-spec/references/reference.md`, `.claude/skills/moai-workflow-spec/references/examples.md`, `.claude/skills/moai-foundation-core/modules/spec-ears-format.md`, `.claude/agents/core/manager-spec.md`) with the GEARS notation canonicalized in predecessor SPEC-V3R6-GEARS-MIGRATION-001 v0.2.0, so that NEW SPEC authors receive GEARS-first guidance while the 6-month EARS backward-compatibility window remains intact.

### 1.1 GEARS Canonical Definition (inherited from SPEC-V3R6-GEARS-MIGRATION-001 §1)

- **Five patterns**: Ubiquitous (unchanged) / `WHEN <event>` (unchanged) / `WHILE <state>` (unchanged, promoted) / `WHERE <precondition>` (reframed — capability gate / feature flag / static config) / **`IF/THEN` DEPRECATED → `WHEN <event-detected>`**
- **Unified compound pattern**: `[Where ...][While ...][When ...] The <subject> shall <behavior>`
- **Generalized subject**: `<subject>` replaces "the system" — any noun (system, component, service, agent, function, artifact). 88 existing SPECs retain "The system" for backward-compat; NEW SPECs MAY use generalized.
- **Lint behavior**: `LegacyEARSKeyword` warning emits on `IF/THEN` (warning non-strict / error `--strict`). Backward-compat window: 6 months from v3.0.0 release.
- **Cross-reference**: `docs-site/content/en/workflow-commands/moai-plan.md#gears-notation`

### 1.2 Out of Scope Declaration

This SPEC is a **guide alignment** SPEC. It does NOT:

#### 1.2.1 Out of Scope

- Modify `internal/spec/lint.go` or any Go source code (done in SPEC-V3R6-GEARS-MIGRATION-001 M2).
- Add/modify `internal/spec/lint_test.go` (done in predecessor M2).
- Touch 4-locale `docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-plan.md` (done in predecessor M3).
- Rewrite the 88 existing SPECs to GEARS (deferred to provisional `SPEC-V3R6-GEARS-SWEEP-001`).
- Change the `LegacyEARSKeyword` warning text, severity, or 6-month window policy.

## 2. Why

### 2.1 Drift Risk (predecessor implemented but authoring guides un-updated)

SPEC-V3R6-GEARS-MIGRATION-001 v0.2.0 emits `LegacyEARSKeyword` warning on `IF/THEN` REQs since 2026-05-22. The 5 guide artifacts above teach authors to write `IF/THEN` REQs (4-14 EARS refs each, totaling 50 references). Any NEW SPEC written using SKILL.md as reference will trigger lint warnings, contradicting the implementation intent of the predecessor.

### 2.2 Dogfooding Pressure

`.claude/skills/moai-workflow-spec/SKILL.md` is the canonical guide for SPEC authoring. If MoAI-ADK's own SPEC authors cannot use GEARS notation, downstream MoAI-ADK consumers will not adopt GEARS either. This SPEC closes the loop between lint engine + docs-site (predecessor) and authoring skill (this SPEC).

### 2.3 Cost Justification

Markdown-only edits across 5 files × 2 locations (local + template mirror) = 10 file touches. Zero Go code changes, zero new tests beyond `moai spec lint` self-verification. Tier M because 3-artifact SPEC structure (spec.md + plan.md + acceptance.md) is REQUIRED to capture binary ACs for cross-ref parity + LegacyEARSKeyword self-check + frontmatter canonical compliance.

## 3. Requirements (GEARS notation — self-dogfooding)

### REQ-SGA-001 (Ubiquitous)

The skill `moai-workflow-spec` **shall** present GEARS notation as the canonical SPEC authoring format throughout SKILL.md, references/reference.md, and references/examples.md.

### REQ-SGA-002 (Event-driven)

**When** a SPEC author reads `.claude/skills/moai-workflow-spec/SKILL.md` § "EARS Five Patterns" / § "EARS Format Deep Dive", the section **shall** present GEARS as the primary notation and EARS as legacy (6-month backward-compat) with a comparison table covering all five patterns.

### REQ-SGA-003 (Event-driven)

**When** the spec-format reference document `.claude/skills/moai-foundation-core/modules/spec-ears-format.md` is loaded by `manager-spec` or `manager-develop`, the document **shall** contain a deprecation banner identifying `IF/THEN` as deprecated in favor of `WHEN <event-detected>` and **shall** cross-link to `docs-site/content/en/workflow-commands/moai-plan.md#gears-notation`.

### REQ-SGA-004 (State-driven, generalized subject)

**While** the v3.0.0 6-month backward-compatibility window is active, the guide **shall** preserve all EARS pattern descriptions as legacy reference so existing 88 SPEC authors retain context.

### REQ-SGA-005 (Optional, capability)

**Where** an author wishes to leverage GEARS compound pattern, the guide **shall** provide at least one worked example using the unified `[Where ...][While ...][When ...] The <subject> shall <behavior>` form with a non-"the system" subject.

### REQ-SGA-006 (Event-driven)

**When** the orchestrator or a sub-agent reads `.claude/agents/core/manager-spec.md`, the agent description **shall** contain the keyword "GEARS" in the description block and the multi-locale keywords section (EN/KO/JA/ZH) **shall** include "GEARS" alongside "EARS".

### REQ-SGA-007 (Ubiquitous, generalized subject)

The 5 target files **shall** be mirrored byte-identical (or semantic-identical where the local copy contains runtime-managed lines) at their `internal/template/templates/<path>` counterparts per the Template-First Rule.

### REQ-SGA-008 (Event-driven)

**When** `moai spec lint .moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/spec.md` is executed, the linter **shall** emit zero `LegacyEARSKeyword` findings on this SPEC's own REQ-SGA-001 through REQ-SGA-N (self-dogfood verification — this SPEC's REQs use GEARS notation only).

### REQ-SGA-009 (Unwanted)

The guide **shall not** introduce new `IF/THEN` REQ examples in any of the 5 target files; existing `IF/THEN` REQ examples used as legacy reference **shall** be annotated with deprecation markers (e.g., `[DEPRECATED — use WHEN]`).

### REQ-SGA-010 (Optional, capability)

**Where** an author wishes to use a generalized subject other than "The system", the guide in SKILL.md **shall** explicitly describe the substitution pattern with at least two non-"the system" examples ("The skill shall ...", "The agent shall ...").

### REQ-SGA-011 (Ubiquitous)

The frontmatter of all 3 artifacts (spec.md / plan.md / acceptance.md) **shall** contain the canonical 12 required fields per `.claude/rules/moai/development/spec-frontmatter-schema.md` with no snake_case alias and **shall** include `tier: M` plus `related_specs: [SPEC-V3R6-GEARS-MIGRATION-001]`.

### REQ-SGA-012 (Event-driven)

**When** any of the 3 artifacts contains content excluded from this SPEC's scope, the exclusion **shall** be declared under an `### X.Y Out of Scope` h3 sub-section (h2 `## Out of Scope` is prohibited by spec-lint `MissingExclusions` rule per CLAUDE.md §11 B6).

## 4. Risks + Mitigations

### Risk R1 — Self-lint failure on this SPEC

Risk: REQ-SGA-001 through REQ-SGA-012 contain EARS-style keywords (`When`, `While`, `Where`, `Ubiquitous`). If the lint regex misclassifies these as legacy `IF/THEN`, AC-SGA-001 fails.

Mitigation: REQ statements use only `WHEN`, `WHILE`, `WHERE`, plain Ubiquitous form. Zero `IF/THEN` tokens. AC-SGA-001 verification command grep-checks `LegacyEARSKeyword` count = 0 on `moai spec lint` output of this spec.md.

### Risk R2 — Template-mirror drift after edits

Risk: Local `.claude/skills/...` and `internal/template/templates/.claude/skills/...` files drift after edits (per CLAUDE.local.md §2 [HARD] Template-First Rule).

Mitigation: AC-SGA-004 verifies byte-identical mirrors via `diff -r` on each of the 5 file pairs. Plan.md M5 milestone enforces local-then-template (or simultaneous Edit) ordering.

### Risk R3 — Excessive scope creep into 88-SPEC sweep

Risk: While editing guide artifacts, author may be tempted to rewrite example SPEC snippets that mirror legacy EARS, expanding scope.

Mitigation: 1.2.1 Out of Scope explicitly excludes 88-SPEC bulk rewrite. Plan.md M1-M4 each have file-list constraints. M5 only mirrors changes, never rewrites.

### Risk R4 — Hugo Goldmark cross-link breakage

Risk: REQ-SGA-003 requires cross-link to `docs-site/content/en/workflow-commands/moai-plan.md#gears-notation` (Hugo Goldmark `{#gears-notation}` anchor). If the predecessor's anchor format diverges from Hugo's runtime resolution, the link breaks.

Mitigation: AC-SGA-009 verifies anchor existence via grep on docs-site content. Anchor format `{#gears-notation}` already validated in predecessor PR #1046 (Hugo Goldmark attribute extension enabled in `docs-site/hugo.toml`).

### Risk R5 — Frontmatter alias drift

Risk: Author copies frontmatter from non-canonical SPEC examples and uses snake_case (`created_at` instead of `created`). spec-lint emits `FrontmatterInvalid`.

Mitigation: AC-SGA-010 grep-checks for snake_case alias absence (`created_at`, `updated_at`, `spec_id` MUST NOT appear). Plan.md M0 pre-flight verifies canonical 12 fields by manual cross-check against `.claude/rules/moai/development/spec-frontmatter-schema.md`.

### Risk R6 — Cross-Wave coordination (KNOWN CONFLICT with Wave 1 SKILL-COMPRESS-001)

Risk: Wave 1 SPEC-V3R6-SKILL-COMPRESS-001 explicitly declares scope (per its spec.md §3): *"`moai-workflow-spec` (2,394w → 1,500w): EARS 핵심만, example 분리"* — direct collision with this SPEC's M1 target `.claude/skills/moai-workflow-spec/SKILL.md` and M2 target `references/examples.md`. SKILL-CONSOLIDATE-001 and RULES-COMPRESS-001 do NOT touch the 5 target files (verified 2026-05-23). The collision is restricted to a single Wave 1 sibling.

Mitigation (plan-auditor S4 finding 2026-05-23, conditional resolution strategy):

- **M0 pre-flight HARD gate**: Before any M1 edit, manager-develop MUST `gh pr list --state merged --search SPEC-V3R6-SKILL-COMPRESS-001` AND check `.moai/specs/SPEC-V3R6-SKILL-COMPRESS-001/spec.md` frontmatter `status:` field.
- **Resolution branch (a)** — SKILL-COMPRESS-001 status = `implemented` (Wave 1 sibling complete): proceed normally with body-aware patches. Re-baseline EARS reference count against the compressed body (expect ≤ 17 instead of 18 in SKILL.md, ≤ 1 unchanged in examples.md).
- **Resolution branch (b)** — SKILL-COMPRESS-001 status = `draft|planned|in-progress` (still pending): file blocker report to orchestrator. Orchestrator MUST trigger AskUserQuestion to user with three options: (i) defer this SPEC's M1+M2 until SKILL-COMPRESS-001 merge, OR (ii) scope this SPEC's M1 to non-overlapping surfaces ONLY (frontmatter description/keywords + new GEARS section header at end of file, leaving body compression to SKILL-COMPRESS-001), OR (iii) escalate to user manual coordination of merge order.
- **No silent overwrite**: under no circumstance may manager-develop edit the body of the 5 target files while SKILL-COMPRESS-001 is `in-progress` — that would silently invalidate the compressed body planned by the Wave 1 sibling.

Documented pattern reference: GEARS-MIGRATION-001 AC-GM-007 second clause (cross-Wave caveat — see `.moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/spec.md`).

## 5. Dependencies + Cross-References

- **Hard dependency**: SPEC-V3R6-GEARS-MIGRATION-001 (status: implemented v0.2.0). This SPEC operates inside the contract established by the predecessor (lint regex, warning message, 6-month window, anchor format).
- **Related — non-blocking**: SPEC-V3R6-GEARS-SWEEP-001 (provisional, 88-SPEC rewrite). This SPEC creates the guide that the sweep will consume.
- **Cross-Wave caveat** (acknowledged): Wave 1 in-progress SPECs may co-touch `.claude/skills/moai-workflow-spec/SKILL.md`. M0 pre-flight detects.
- **Doctrine references**: CLAUDE.md §1 HARD Rules, CLAUDE.local.md §2 Template-First Rule, `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier, `.claude/rules/moai/development/spec-frontmatter-schema.md`.

### 5.1 Working-Tree Baseline Snapshot (2026-05-23 plan-phase entry)

Baseline `git status --porcelain` count = **20** at plan-phase entry. After plan-phase artifact creation = **21** (+ 1 untracked `.moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/` directory). Modified/deleted files (6): `.moai/config/sections/{git-convention,github-actions(D),language,quality}.yaml` + `.moai/harness/usage-log.jsonl` + `.moai/specs/SPEC-CI-MULTI-LLM-001/spec.md`. Untracked items (15): 5 Wave 1 SPEC dirs + 4-locale `docs-site/content/{en,ko,ja,zh}/book/` + 2 `.moai/research/*-2026-05-22.md` files + `docs-site/data/menu/extra.yaml` + `docs-site/layouts/_default/redirect.html` + `docs-site/scripts/gen_menu.py` + `docs-site/static/book/` + `internal/hook/.moai/` + `internal/template/github_tmpl_parse_test.go` + `.moai/specs/.moai/` + this SPEC dir. M0 pre-flight re-verifies this snapshot remains stable between plan-phase and run-phase entry; any drift triggers blocker report.

## 6. References

- SPEC-V3R6-GEARS-MIGRATION-001 spec.md §1 (GEARS canonical definition)
- SPEC-V3R6-GEARS-MIGRATION-001 implementation commits: `431a999be` (M2 lint) + `0bdbae7c2` (M2 tests) + `b3d2a52da` (M3 docs) + `c8ad2c542` (M4 chore v0.2.0)
- GEARS paper: DEV Community 2026-01-23 "Generalized Expression for AI-Ready Specs"
- EARS official guide: Mavin, A. "EARS: The Easy Approach to Requirements Syntax"
- Lint rule source: `internal/spec/lint.go` `EARSModalityRule.Check()` `LegacyEARSKeyword` finding
