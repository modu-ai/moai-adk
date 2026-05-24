---
id: SPEC-V3R6-SKILL-GEARS-ALIGN-001
title: "Implementation Plan — moai-workflow-spec SKILL GEARS 우선 정렬"
version: "0.2.0"
status: implemented
created: 2026-05-23
updated: 2026-05-25
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/skills/moai-workflow-spec, .claude/skills/moai-foundation-core/modules, .claude/agents/core, internal/template/templates"
lifecycle: spec-anchored
tags: "gears, ears, skill, guide, alignment, plan, wave-6, v3.0.0"
tier: M
issue_number: null
depends_on: [SPEC-V3R6-GEARS-MIGRATION-001]
related_specs: [SPEC-V3R6-GEARS-MIGRATION-001]
---

# SPEC-V3R6-SKILL-GEARS-ALIGN-001 — Implementation Plan

## 1. Implementation Overview

This SPEC executes guide-alignment markdown edits across 5 target files in 2 locations (local + template) = 10 file touches plus 1 chore commit. The run-phase is delegated to **`manager-develop`** with `cycle_type=ddd` and the Section A-E 5-section delegation prompt template (Tier M REQUIRED per `.claude/rules/moai/workflow/spec-workflow.md`).

### 1.1 Strategy

- **Markdown-only edits**: zero Go code, zero new tests beyond CLI self-verification via `moai spec lint`.
- **Template-First Rule**: every local edit must be mirrored at `internal/template/templates/<path>` (CLAUDE.local.md §2 [HARD]).
- **Atomic milestones**: 6 milestones (M0 pre-flight, M1 SKILL.md, M2 references/, M3 spec-ears-format.md, M4 manager-spec.md, M5 template mirrors, M6 chore + status bump). M5 mirrors are produced inline with M1-M4 (preferred) OR batched at M5 end (fallback).
- **Self-dogfooding**: this SPEC's REQs are written in GEARS notation. AC-SGA-001 self-verifies `LegacyEARSKeyword = 0` after run-phase.

### 1.2 Out of Scope (plan-level)

#### 1.2.1 Out of Scope

- Re-implementation of `internal/spec/lint.go` rules (covered in SPEC-V3R6-GEARS-MIGRATION-001 M2).
- 4-locale docs-site changes (covered in SPEC-V3R6-GEARS-MIGRATION-001 M3).
- Bulk rewrite of 88 existing SPECs (deferred to provisional SPEC-V3R6-GEARS-SWEEP-001).
- Cross-Wave conflict mediation with Wave 1 in-progress SPECs (M0 detects + blocker if found).

## 2. Section A-E Delegation Prompt Template (manager-develop entry)

The orchestrator delegates run-phase via the following 5-section prompt to `manager-develop` (cycle_type=ddd, Tier M REQUIRED):

### Section A — Context

- Working location: `/Users/goos/MoAI/moai-adk-go` (main checkout).
- Current branch: `main` (or feat branch if Hybrid Trunk Tier M decided otherwise).
- SPEC ID: `SPEC-V3R6-SKILL-GEARS-ALIGN-001`.
- Predecessor: `SPEC-V3R6-GEARS-MIGRATION-001` (implemented v0.2.0, PR #1046 merged 2026-05-22).
- Tier: **M** (3 artifacts confirmed; plan-auditor PASS threshold 0.80).
- Cycle type: `ddd` (per `.moai/config/sections/quality.yaml`).

### Section B — Known Issues 자동 주입 (Tier M REQUIRED)

- **B1 Cross-platform**: N/A (markdown-only).
- **B2 Cross-SPEC**: Depends on `SPEC-V3R6-GEARS-MIGRATION-001` (implemented); related to provisional `SPEC-V3R6-GEARS-SWEEP-001` (not blocking).
- **B3 C-HRA-008**: N/A (no Go subagent code).
- **B4 Frontmatter Canonical**: 12 required fields per `.claude/rules/moai/development/spec-frontmatter-schema.md` + `tier: M` + `related_specs` + no snake_case alias.
- **B5 CI 3-tier**: spec-lint only (golangci-lint / Test N/A for markdown SPEC). Baseline drift inherited from GEARS-MIGRATION-001 (7 FrontmatterInvalid on SPEC-V3R5-INIT-WIZARD-EXPANSION-001) — 0 NEW regression obligation.
- **B6 spec-lint Heading**: `### X.Y Out of Scope` h3 form required (h2 `## Out of Scope` rejected as `MissingExclusions` ERROR).
- **B7 observer.go**: N/A.
- **B8 Working Tree Hygiene**: PRESERVE 14 pre-existing untracked items + 6 modified/deleted files baseline intact (per spec.md §5.1 verified snapshot). 15th untracked = this SPEC directory itself.

### Section C — Pre-flight Check List

```bash
git rev-parse HEAD                                                    # confirm baseline
git branch --show-current                                             # confirm working branch
git status --porcelain | wc -l                                        # baseline count (≥ 20 entering run)
ls .moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/                       # 3 artifacts present
ls .moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/spec.md                  # predecessor implemented
grep -c "GEARS" .claude/skills/moai-workflow-spec/SKILL.md            # baseline 0
grep -c "GEARS" .claude/agents/core/manager-spec.md                   # baseline 0
grep -c "GEARS" .claude/skills/moai-foundation-core/modules/spec-ears-format.md  # baseline 0
grep -c "{#gears-notation}" docs-site/content/en/workflow-commands/moai-plan.md  # baseline ≥ 1 (anchor from predecessor)
which moai && moai spec lint --help 2>&1 | head -5                    # CLI availability
```

If any pre-flight check fails, halt and report blocker.

### Section D — Constraints

**PRESERVE intact** (do not modify):

- 5 Wave 1 SPEC directories: `.moai/specs/SPEC-V3R6-{RULES-COMPRESS,SKILL-CONSOLIDATE,SKILL-COMPRESS,CODE-COMMENTS-EN}-001/`
- `.moai/specs/.moai/` (orphan stash artifact)
- `docs-site/content/{ko,en,ja,zh}/book/`, `docs-site/data/menu/extra.yaml`, `docs-site/layouts/_default/redirect.html`, `docs-site/scripts/gen_menu.py`, `docs-site/static/book/`
- `.moai/research/moai-adk-current-state-2026-05-22.md`, `.moai/research/v3.0-design-2026-05-22.md`
- `internal/hook/.moai/` working dir leak
- `internal/template/github_tmpl_parse_test.go` (untracked)
- 6 modified/deleted files (canonical, per `git status --porcelain` 2026-05-23): `M .moai/config/sections/git-convention.yaml` + `D .moai/config/sections/github-actions.yaml` + `M .moai/config/sections/language.yaml` + `M .moai/config/sections/quality.yaml` + `M .moai/harness/usage-log.jsonl` + `M .moai/specs/SPEC-CI-MULTI-LLM-001/spec.md`

**Out of Scope**: see §1.2.1.

**Prohibited**:

- `--no-verify`, `--amend`, force-push
- Snake_case frontmatter alias (`created_at`, `updated_at`, `spec_id`)
- `## Out of Scope` h2 standalone
- Direct user prompting (subagents: blocker report only)

**Mandatory**:

- Conventional Commits (`feat(SPEC-V3R6-SKILL-GEARS-ALIGN-001): ...` or `docs(SPEC-V3R6-SKILL-GEARS-ALIGN-001): ...`).
- Template-First Rule: every local edit mirrored at template path.

### Section E — Self-Verification Deliverables (manager-develop completion)

- All 13 ACs verified via documented commands in acceptance.md.
- `moai spec lint` on this spec.md emits 0 `LegacyEARSKeyword` + 0 `FrontmatterInvalid`.
- 5 file pairs byte-identical via `diff -q`.
- Working tree status: pre-existing 14 untracked + 6 modified/deleted intact (per spec.md §5.1); only new modifications during run-phase are the 5 local + 5 template-mirror file pairs plus chore touches.
- Completion report includes: file-by-file edit summary, 13/13 AC verification command outputs, baseline drift diagnosis (NEW regression 0 vs inherited).

## 3. Execution Milestones (Priority-Ordered)

Per CLAUDE.md §11 Time Estimation: priority-ordered, no time predictions.

### Milestone M0 — Pre-flight + Cross-Wave Conflict Scan (priority: P0)

**Goal**: Confirm baseline state and apply the conditional resolution strategy declared in spec.md §Risk R6 (Wave 1 SKILL-COMPRESS-001 KNOWN collision on `moai-workflow-spec/SKILL.md` + `references/examples.md`).

**Actions**:

1. Run Section C pre-flight check list.
2. Cross-Wave conflict scan (record the result, do NOT halt on positive match):

```bash
for w in SPEC-V3R6-RULES-COMPRESS-001 SPEC-V3R6-SKILL-CONSOLIDATE-001 SPEC-V3R6-SKILL-COMPRESS-001; do
  echo "--- $w ---"
  grep -nE "moai-workflow-spec/SKILL\.md|spec-ears-format\.md|manager-spec\.md|references/examples\.md|references/reference\.md" .moai/specs/$w/*.md || echo "no overlap"
done
```

3. SKILL-COMPRESS-001 status check (HARD gate from spec.md §Risk R6):

```bash
grep "^status:" .moai/specs/SPEC-V3R6-SKILL-COMPRESS-001/spec.md
gh pr list --state merged --search "SPEC-V3R6-SKILL-COMPRESS-001" --json number,title,mergedAt 2>/dev/null || echo "no merged PR"
```

4. Branch on SKILL-COMPRESS-001 status:
   - **Branch (a)** — status = `implemented` AND merged PR exists: proceed to M1 with body-aware patches. Re-baseline EARS reference counts (run `grep -c "EARS" .claude/skills/moai-workflow-spec/SKILL.md` and use the new value, not the 18 captured at plan-time).
   - **Branch (b)** — status ∈ {`draft`, `planned`, `in-progress`}: STOP. File a structured blocker report to the orchestrator listing the SKILL-COMPRESS-001 status, the colliding file paths, and the three resolution options (defer / scope-to-non-body / user manual coordination). Do NOT touch SKILL.md or examples.md body in this run.

5. RULES-COMPRESS-001 + SKILL-CONSOLIDATE-001 are verified non-overlapping (per plan-auditor S4 finding 2026-05-23) — record this in progress.md but proceed.

**Exit criterion**: pre-flight checks PASS, SKILL-COMPRESS-001 status recorded in progress.md, branch (a) OR (b) decided. Under branch (b), this SPEC's run-phase halts here.

### Milestone M1 — SKILL.md GEARS-First Restructure (priority: P0)

**Goal**: Convert `.claude/skills/moai-workflow-spec/SKILL.md` § "EARS Five Patterns" / § "EARS Format Deep Dive" sections to GEARS-first guidance.

**Target file**: `.claude/skills/moai-workflow-spec/SKILL.md` (18 EARS refs baseline).

**Edit operations**:

1. Add a new sub-section "GEARS Five Patterns (current notation)" before existing "EARS Five Patterns" — the new section becomes primary; existing section relabeled as "EARS Five Patterns (legacy — 6-month backward-compat)".
2. Insert a comparison table covering all five patterns with columns: Pattern / EARS (legacy) / GEARS (current) / Notes.
3. Reframe `WHERE` from "Optional feature" to "Capability gate / feature flag / static config" (the GEARS clarification).
4. Add an "IF/THEN deprecated" callout box explaining migration to `WHEN <event-detected>`.
5. Add a "Generalized subject substitution" sub-section with ≥ 2 non-"the system" examples ("The skill shall ...", "The agent shall ...", "The component shall ...").
6. Add at least one compound pattern example: `**Where** project is initialized **While** strict mode is active **When** a SPEC author runs lint, the lint engine **shall** ...`.
7. Add cross-link line: `See [GEARS notation](https://adk.mo.ai.kr/en/workflow-commands/moai-plan/#gears-notation) — 4-locale.`
8. Add reference to predecessor: `Lint behavior canonicalized in SPEC-V3R6-GEARS-MIGRATION-001 v0.2.0 (PR #1046).`
9. Update Quick Reference (line 50): EARS Five Patterns → GEARS Five Patterns (current; EARS legacy supported during 6-month window).

**Validation after edit**:

- `grep -c "GEARS" .claude/skills/moai-workflow-spec/SKILL.md` ≥ 5
- AC-SGA-002, AC-SGA-005, AC-SGA-006, AC-SGA-009 partial PASS.

**Exit criterion**: SKILL.md retains all legacy EARS content (no deletion) and adds GEARS-first sections + comparison table + compound example + cross-link.

### Milestone M2 — references/{reference,examples}.md GEARS Migration Notes (priority: P1)

**Goal**: Update SKILL.md companion references to align with M1 changes.

**Target files**:

- `.claude/skills/moai-workflow-spec/references/reference.md` (13 EARS refs baseline)
- `.claude/skills/moai-workflow-spec/references/examples.md` (1 EARS ref baseline)

**Edit operations**:

1. **reference.md**: Add a "GEARS Migration" sub-section before existing EARS sections; preserve EARS sections as legacy reference; cross-link to SKILL.md GEARS section. Add deprecation banner to any IF/THEN entries.
2. **examples.md**: Add at least one GEARS-form example using the unified compound pattern `[Where ...][While ...][When ...] The <subject> shall <behavior>` with a non-"the system" subject. Existing examples retained but annotated `[Legacy EARS form — equivalent GEARS form: ...]` where applicable.

**Validation after edit**:

- `grep -c "GEARS" .claude/skills/moai-workflow-spec/references/reference.md` ≥ 2
- AC-SGA-006 verification command shows compound clause example in examples.md.

**Exit criterion**: Both reference files cross-link to SKILL.md GEARS sections and contain at least one GEARS-form example.

### Milestone M3 — spec-ears-format.md Deprecation Banner + Cross-Link (priority: P1)

**Goal**: Add upfront deprecation banner and cross-link to `.claude/skills/moai-foundation-core/modules/spec-ears-format.md`.

**Target file**: `.claude/skills/moai-foundation-core/modules/spec-ears-format.md` (4 EARS refs baseline).

**Edit operations**:

1. Insert a deprecation banner within the first 30 lines (top of document, after title):

```markdown
> **DEPRECATED in v3.0.0**: This document describes legacy EARS notation. For NEW SPECs, use GEARS notation per [GEARS migration guide](https://adk.mo.ai.kr/en/workflow-commands/moai-plan/#gears-notation). Backward-compatibility for legacy EARS REQs is supported for 6 months from v3.0.0 release. See SPEC-V3R6-GEARS-MIGRATION-001 v0.2.0 for the lint engine canonicalization.
```

2. Annotate IF/THEN reference with deprecation marker (do not delete — legacy retention).
3. Optional file rename consideration: `spec-ears-format.md` → `spec-gears-format.md`. **Decision: do NOT rename in this SPEC** — Open Question Q1 default. Add redirect note at top: `> NOTE: This file describes EARS legacy notation. GEARS canonical guide is at SKILL.md "GEARS Five Patterns".` If rename is required, defer to a follow-up SPEC.

**Validation after edit**:

- `head -30 .claude/skills/moai-foundation-core/modules/spec-ears-format.md | grep -c "DEPRECATED\|deprecated\|GEARS"` ≥ 1
- AC-SGA-003, AC-SGA-008 PASS.

**Exit criterion**: Banner visible in first 30 lines; cross-link to docs-site present; existing EARS content unchanged except deprecation annotations.

### Milestone M4 — manager-spec.md Description + 4-Locale Keywords (priority: P1)

**Goal**: Update `.claude/agents/core/manager-spec.md` frontmatter so dispatcher matches "GEARS" keyword across all 4 locales.

**Target file**: `.claude/agents/core/manager-spec.md` (14 EARS refs baseline).

**Edit operations**:

1. Frontmatter `description:` block: add ", GEARS" alongside ", EARS" in the description sentence ("SPEC creation specialist. Use PROACTIVELY for EARS/GEARS-format requirements, acceptance criteria, ...").
2. Frontmatter keyword sections (EN/KO/JA/ZH): add "GEARS" alongside "EARS" in each locale line:
   - `EN: SPEC, requirement, specification, EARS, GEARS, acceptance criteria, user story, planning`
   - `KO: SPEC, 요구사항, 명세서, EARS, GEARS, 인수조건, 유저스토리, 기획`
   - `JA: SPEC, 要件, 仕様書, EARS, GEARS, 受入基準, ユーザーストーリー`
   - `ZH: SPEC, 需求, 规格书, EARS, GEARS, 验收标准, 用户故事`
3. Body: where EARS is described as the format, add "GEARS (current) / EARS (legacy, 6-month backward-compat window)" clarification.

**Validation after edit**:

- `grep -c "GEARS" .claude/agents/core/manager-spec.md` ≥ 5
- `grep -E "^\s*(EN|KO|JA|ZH):" .claude/agents/core/manager-spec.md | grep -c "GEARS"` = 4
- AC-SGA-007 PASS.

**Exit criterion**: Agent frontmatter dispatcher matches "GEARS" in all 4 locale keyword lines + description block.

### Milestone M5 — Template Mirror Parity (priority: P0 — interleaved with M1-M4)

**Goal**: Mirror M1-M4 edits to `internal/template/templates/<path>` counterparts.

**Target files (5 pairs)**:

| Local | Template Mirror |
|-------|-----------------|
| `.claude/skills/moai-workflow-spec/SKILL.md` | `internal/template/templates/.claude/skills/moai-workflow-spec/SKILL.md` |
| `.claude/skills/moai-workflow-spec/references/reference.md` | `internal/template/templates/.claude/skills/moai-workflow-spec/references/reference.md` |
| `.claude/skills/moai-workflow-spec/references/examples.md` | `internal/template/templates/.claude/skills/moai-workflow-spec/references/examples.md` |
| `.claude/skills/moai-foundation-core/modules/spec-ears-format.md` | `internal/template/templates/.claude/skills/moai-foundation-core/modules/spec-ears-format.md` |
| `.claude/agents/core/manager-spec.md` | `internal/template/templates/.claude/agents/core/manager-spec.md` |

**Preferred strategy**: Edit local file → immediately copy/mirror to template path in same milestone (M1-M4). M5 then becomes a final verification step.

**Fallback strategy**: Batch all template mirrors at M5 end if interleaving complicates `manager-develop` execution.

**Validation**:

```bash
for f in [5 file paths]; do
  diff -q "$f" "internal/template/templates/$f" || echo "DRIFT: $f"
done
```

**Exit criterion**: Zero `DRIFT:` lines. AC-SGA-004 PASS.

### Milestone M6 — Chore Commit: status:implemented + version 0.2.0 (priority: P2)

**Goal**: Update spec.md frontmatter `status: draft` → `implemented` and `version: 0.1.0` → `0.2.0`. Write `progress.md` capturing M0-M5 evidence.

**Edit operations**:

1. Update `spec.md` frontmatter: `status: implemented` + `version: "0.2.0"` + `updated: <run-phase date>`.
2. Update HISTORY table with v0.2.0 entry: M1-M5 commit hashes + 13/13 AC PASS confirmation.
3. Write `.moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/progress.md` with milestone-by-milestone evidence (paste-able AC verification outputs).
4. Mirror updated spec.md to template path (consistency with M5).

**Validation**:

- AC-SGA-010 (frontmatter canonical) re-verify PASS.
- AC-SGA-011 (h3 Out of Scope) re-verify PASS.
- 13/13 ACs PASS via documented commands.

**Exit criterion**: spec.md status=implemented v0.2.0; progress.md exists with AC evidence; commit message `chore(SPEC-V3R6-SKILL-GEARS-ALIGN-001): status implemented v0.2.0`.

## 4. Technical Approach

### 4.1 Edit Mechanics (markdown-only)

All edits use the Edit/MultiEdit tools on markdown files. No Go code, no test files. The `manager-develop` agent should NOT delegate to `expert-backend` or `expert-frontend` — Tier M markdown-only does not require domain experts.

### 4.2 Local-Template Mirror Strategy

Two equivalent strategies — prefer (A):

- **(A) Interleaved**: After each M1-M4 local edit, immediately Edit/Write the template mirror with identical content. M5 verifies byte-identity.
- **(B) Batched**: Complete M1-M4 on local files first, then M5 batch-mirrors via `cp` or sequential Edit. Higher risk of drift if local files are re-edited between mirror passes.

`manager-develop` should declare which strategy is in use in the M5 progress entry.

### 4.3 Self-Lint Verification (AC-SGA-001)

After M0-M5 completion, run:

```bash
moai spec lint .moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/spec.md 2>&1 | tee .moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/lint-output.txt
grep -c "LegacyEARSKeyword" .moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/lint-output.txt
grep -c "FrontmatterInvalid" .moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/lint-output.txt
```

Both counts must be `0`. The `lint-output.txt` is kept as progress.md evidence.

### 4.4 Plan-Auditor (Tier M REQUIRED)

After plan-phase artifact creation, orchestrator delegates to `Task(plan-auditor)` for PASS threshold `0.80`. plan-auditor evaluates:

- D1 (Completeness): 12-field frontmatter + 3 artifacts + Section A-E + HISTORY (target ≥ 0.80)
- D2 (EARS/GEARS Compliance): REQ-SGA-001 .. REQ-SGA-012 use GEARS notation only (target ≥ 0.85)
- D3 (Traceability): 100% REQ ↔ AC mapping (target ≥ 0.85)
- D4 (Bias / Skeptical Stance): External evidence (predecessor commits, lint engine doc) (target ≥ 0.80)

Average ≥ 0.80 = PASS. If REVISE, orchestrator applies BLOCKING fixes and re-runs plan-auditor (iter 2). STOP signal if iter(N+1) < iter(N).

## 5. Risks (run-phase + plan-phase intersection)

Inherited from spec.md §4 (R1-R6). Plan-phase additions:

### Risk R7 (plan) — Template-First Rule fatigue

Risk: `manager-develop` may forget local-template mirroring on later milestones (M4 forgets to mirror manager-spec.md).

Mitigation: M5 verification command (`diff -q` loop) catches drift. AC-SGA-004 enforces final byte-identity. Plan explicitly designates M5 as the final gate.

### Risk R8 (plan) — Plan-Auditor sub-0.80 on first pass

Risk: plan-auditor iter 1 returns REVISE (precedent: 50% of Wave 0/1 SPECs).

Mitigation: 13 binary ACs + 100% REQ traceability matrix + canonical frontmatter pre-applied in this plan + h3 Out of Scope present. Expected iter 1 PASS with margin. If REVISE, orchestrator applies fix + iter 2.

## 6. Dependencies + Open Questions

### 6.1 Hard Dependencies

- `SPEC-V3R6-GEARS-MIGRATION-001` status: implemented v0.2.0 — verified in M0 pre-flight.
- `moai spec lint` CLI subcommand (`internal/cli/spec_lint.go`) exists — verified in M0 pre-flight.
- Hugo Goldmark `{#gears-notation}` anchor exists in `docs-site/content/en/workflow-commands/moai-plan.md` — verified in M0 pre-flight.

### 6.2 Open Questions (default decisions documented)

- **Q1 — Rename `spec-ears-format.md` → `spec-gears-format.md`?**
  Default: **DO NOT rename in this SPEC** (banner + redirect note suffices). Rename defers to follow-up SPEC to avoid cross-ref breakage in 14 callers (`grep -rn "spec-ears-format" .claude/ internal/template/templates/.claude/`).

- **Q2 — Use `moai spec lint` CLI or fallback to `go test -run TestEARSModalityRule`?**
  Default: **Use `moai spec lint`** (verified available in M0). Fallback documented in AC-SGA-001 alternate.

- **Q3 — Should this SPEC's REQs emit `LegacyEARSKeyword` warning if the lint regex matches their EARS-style modality keywords?**
  Default: **No expected warning** — REQ-SGA-001..012 use GEARS notation (WHEN/WHILE/WHERE/Ubiquitous form) and avoid IF/THEN tokens. AC-SGA-001 self-verifies. If the lint regex unexpectedly flags this SPEC, the run-phase blocker triggers refinement (e.g., remove residual IF tokens from spec.md).

### 6.3 Re-Delegation Procedure

If `manager-develop` encounters a missing input or ambiguous decision not covered by Q1-Q3 defaults:

1. Return structured "Missing Inputs" blocker report (no `AskUserQuestion` from sub-agent).
2. Orchestrator runs `AskUserQuestion` round with user.
3. Orchestrator re-delegates to `manager-develop` with injected answers.

## 7. Verification Pipeline (Pre-Merge Gate)

| Stage | Verification | Pass Criterion |
|-------|--------------|----------------|
| Plan-Phase | `Task(plan-auditor)` | Score ≥ 0.80 |
| Run-Phase M0 | Section C pre-flight | All 9 checks PASS |
| Run-Phase M1 | AC-SGA-002, AC-SGA-005, AC-SGA-006, AC-SGA-009 partial | grep counts meet thresholds |
| Run-Phase M2 | AC-SGA-006 | Compound clause in examples.md |
| Run-Phase M3 | AC-SGA-003, AC-SGA-008 | Banner + 6-month note |
| Run-Phase M4 | AC-SGA-007 | 4-locale keywords + description |
| Run-Phase M5 | AC-SGA-004 | Zero `DRIFT:` lines |
| Run-Phase M6 | AC-SGA-001, AC-SGA-010, AC-SGA-011, AC-SGA-013, AC-SGA-012 | 13/13 PASS |
| Pre-Commit | `moai spec lint` on spec.md | 0 LegacyEARSKeyword + 0 FrontmatterInvalid |
| Pre-Merge | All 13 ACs PASS + plan-auditor PASS | Mergeable |

## 8. References

- Predecessor SPEC: `.moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/spec.md` v0.2.0
- spec-workflow doctrine: `.claude/rules/moai/workflow/spec-workflow.md`
- Frontmatter schema: `.claude/rules/moai/development/spec-frontmatter-schema.md`
- Template-First Rule: `CLAUDE.local.md` §2 [HARD]
- plan-auditor agent: `.claude/agents/core/plan-auditor.md`
- manager-develop agent: `.claude/agents/core/manager-develop.md`
- Lint engine: `internal/spec/lint.go` `EARSModalityRule` `LegacyEARSKeyword`
- Docs-site anchor: `docs-site/content/en/workflow-commands/moai-plan.md#gears-notation`
