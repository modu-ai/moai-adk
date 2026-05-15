---
id: SPEC-V3R2-WF-001
title: Skill Consolidation — Stage 1 (48 to 38)
version: "1.2.0"
status: completed
created: 2026-04-23
updated: 2026-04-25
author: Wave 2 SPEC writer (Layer 6/7/Cleanup)
priority: P1 High
phase: "v3.0.0 — Phase 4 — Skill Consolidation (Stage 1)"
module: ".claude/skills/, internal/template/templates/.claude/skills/"
dependencies: []
related_gap:
  - r4-skill-audit
  - problem-catalog-workflow-bloat
related_theme: "Theme 6 — Workflow Consolidation"
breaking: true
bc_id: [BC-V3R2-006]
lifecycle: spec-anchored
tags: "skill-consolidation, skill-audit, merge, retire, v3, workflow, breaking, stage1"
follow_up_spec: SPEC-V3R3-WF-001  # Stage 2 (38 → 24), reserved; not yet drafted
---

# SPEC-V3R2-WF-001: Skill Consolidation — Stage 1 (48 to 38)

## HISTORY

| Version | Date       | Author | Description                                                           |
|---------|------------|--------|-----------------------------------------------------------------------|
| 0.1.0   | 2026-04-23 | Wave 2 | Initial SPEC draft for 48 → 24 skill reduction per R4 skill audit     |
| 1.1.0   | 2026-04-25 | manager-spec | Audit response (plan-audit 2026-04-25 FAIL → revised). Scope honesty: target revised 48 → 38 (Stage 1 of 48 → 24 trajectory; Stage 2 reserved for SPEC-V3R3-WF-001). §6.1 demoted to logical grouping (§6.2 becomes single SoT). Dual-verdict rows (foundation-context, jit-docs) resolved to RETIRE. moai-design-tools split committed to Pencil → pencil-integration, Figma → archive. REQ-WF001-009/012/016 ACs added (AC-16/17/18). REQ-WF001-002 verdict-uniqueness task added. MIG-001 dependency edge upgraded to contract. AC-WF001-08 extended with broken-fixture rejection check. Typo "단일 단일" fixed. MERGED-INTO roll-up count corrected 3 → 4. M1 unblock criteria made objective per OQ. |
| 1.2.0   | 2026-05-15 | manager-docs | Sync-phase status drift resolution. Implementation Notes (2026-04-25, line ~381) record full completion of Stage 1 consolidation: 48 → 38 skill directories achieved. plan-audit (iter 1, 2026-05-15) verdict PASS @ 0.92. Status transition: in-progress → completed per V3R4-SPECLINT-DEBT-001 lifecycle convention. No code/document changes — frontmatter sync only. |

---

## 1. Goal (목적)

MoAI v2.13.2의 48개 skill에 대해 **Stage 1 consolidation**을 집행하여 최종 38개 skill 디렉터리(5개 merge target 흡수 + 11건 retirement + 1개 신규 design-system = 48 − 11 + 1 = 38)를 v3.0.0에 존속시킨다. 이 SPEC은 R4 skill audit(2026-04-23)의 verdict 분포(KEEP 12 / REFACTOR 14 / MERGE 15 / RETIRE 5 / UNCLEAR 2)에서 집행 가능한 부분을 단일 실행 계약으로 고정한다. 5개 merge cluster(thinking, design, database, templates-to-project, design-tools split)와 11건 retirement(§6.2)를 명시적 KEEP/REFACTOR/MERGE/RETIRE 판정으로 고정하고, 각 skill의 트리거 키워드 union 보존 및 bundled resource 재배치를 규정한다.

### 1.0 Staging Note — why 38, not 24

R4 audit은 "~24 skills" 라는 지향 목표를 제시했으나 §Per-skill audit table의 개별 verdict는 KEEP 37 + 신규 1 = 38 만 달성 가능하다. 잔여 14개 추가 RETIRE는 monitor/UNCLEAR/REFACTOR 범주 재평가가 필요하며 이는 본 SPEC의 단일 단위 실행 범위를 초과한다. 따라서 본 SPEC은 **Stage 1 (48 → 38)** 을 집행하고 **Stage 2 (38 → 24)** 는 `SPEC-V3R3-WF-001` 로 예약한다. 이 결정은 plan.md §Decision Log 에 근거와 함께 기록된다.

### 1.1 배경

R4 audit §Executive summary: "Total skills: 48, Recommended v3 skill inventory: ~24 skills (50% reduction from 48). Biggest single bloat: `moai` skill itself (18KB + 20 workflow files ≈ 300KB bundled md)." Audit는 9개 채점 차원(총 27점) 스코어링을 48개 skill 모두에 적용했으며 template/local drift는 0이다(§Section D). 주요 문제: (a) thinking triplet(3-way duplication, ~33KB), (b) kitchen-sink domain skills(backend/frontend/database 22 trigger keywords 각각), (c) platform triplet anti-pattern(3 vendors in one skill), (d) `moai-lang-*` skills referenced but absent(`.claude/rules/moai/languages/` 아래에 rules로 존재 — WF-005가 전담).

### 1.2 비목표 (Non-Goals)

- Skill 본문의 문체/톤 rewrite (프레임 보존이 merge 리스크 관리)
- Progressive Disclosure Level 2 token budget 재설계(`.claude/rules/moai/development/skill-authoring.md` 기존 규칙 유지)
- `moai-lang-*` skills 신규 생성 (SPEC-V3R2-WF-005가 rules 유지 결정을 codify)
- 신규 skill 카테고리 도입(v3.0에서 `moai-cmd-*`는 향후 v3.1에서 재논의)
- Evaluator-active 점수/Progressive Disclosure level2_tokens 수치 조정
- Agency-absorption 계약 수정 (`moai-domain-copywriting`, `moai-domain-brand-design` 프리즈; SPEC-AGENCY-ABSORB-001 계약 유지)

---

## 2. Scope (범위)

### 2.1 In Scope

- **Owns**: `.claude/skills/` 디렉터리 48→38 변경 (Stage 1), `internal/template/templates/.claude/skills/` 동일 변환, 각 KEEP/REFACTOR/MERGE/RETIRE 판정 집행.
- 신규 병합 skill 3종 신설(`moai-foundation-thinking` 확장, `moai-design-system` 신설, `moai-domain-database` 확장).
- 11건 RETIRE(archive) + 5건 MERGE target 재구성 + 6건 REFACTOR 라벨 주입 + KEEP 판정 유지.
- 48개 entry 각각에 단일 verdict 라벨 보장(REQ-WF001-002) 및 CI 검증 fixture 포함.
- 각 skill의 `related-skills` frontmatter 필드 재연결(merge 대상 skill 이름 alias 처리 포함).
- Trigger keyword union preservation: merge 대상 skill의 모든 trigger를 신규 skill의 frontmatter trigger에 union 병합.
- Bundled resource 재배치: 삭제되는 skill의 `modules/`, `references/` 중 재사용 가능한 asset을 신규 skill의 Level 3 payload로 이관.
- `moai` root skill의 `workflows/*.md` 20개 파일 범위 축소는 SPEC-V3R2-WF-002가 담당(본 SPEC은 `moai` skill 본체 KEEP but split 판정만 기록).

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- Skill 본문의 언어 번역 / style rewrite
- `moai-ref-*` 5개 skill의 trigger 키워드 재설계 (keyword-matching auto-activation 모델 유지)
- Stage 1 목표 38개를 벗어나는 추가 신규 skill 창설 (38 → 24 추가 감축은 SPEC-V3R3-WF-001에서 처리)
- Agency 흡수 skill 2종(`copywriting`, `brand-design`) 내용 수정 (FROZEN per .claude/rules/moai/design/constitution.md)
- `moai-workflow-testing` 22.5KB body split (본 SPEC은 REFACTOR 라벨만 부여; 실제 split은 별도 sub-SPEC)
- 16개 language rules → skills migration (SPEC-V3R2-WF-005가 codify)
- v3.1 이후의 `moai-cmd-*` promotion (본 SPEC은 v3.0 수치만 고정)

---

## 3. Environment (환경)

- 런타임: moai-adk-go (Go 1.26+), `internal/template/` embedded FS
- 영향 디렉터리:
  - 수정: `.claude/skills/<48 directories>/`, `internal/template/templates/.claude/skills/<48 directories>/`
  - 참조: `.claude/agents/`, `.claude/commands/`, `.claude/rules/moai/development/skill-authoring.md`
- 기준 상태: R4 audit 기준 48개 skill 디렉터리, template/local byte-identical
- 외부 레퍼런스: R4 audit §Per-skill audit table, §Merge clusters, §Recommended v3 skill inventory, synthesis pattern-library §M-1/M-4
- **Skill consolidation shares a behavioral contract with SPEC-V3R2-MIG-001** (dependency edge, not "references only"). WF-001 authors the artifact `.moai/decisions/skill-rename-map.yaml` per the schema in plan.md §2.5. MIG-001 in Phase 8 consumes that artifact and enacts `moai update` migrations on user local trees (REQ-WF001-009). The schema in plan.md §2.5 is a shared contract that BOTH SPECs MUST honor; any schema change requires a HUMAN GATE review before merge (see plan.md §Decision Log, `OQ-CONTRACT` resolution).

---

## 4. Assumptions (가정)

- R4 audit §Per-skill audit table의 48개 verdict는 authoritative하며 본 SPEC 시점(2026-04-23)까지 유효하다.
- Agency 흡수 skill 2종은 FROZEN 상태이며 내용 변경 없이 그대로 존속한다.
- Template-First 규칙(CLAUDE.local.md §2)에 따라 변경은 `internal/template/templates/` 우선 적용 후 `make build`로 전파된다.
- Progressive Disclosure keyword-matching auto-activation 모델은 v3에서도 유지된다(모든 `moai-ref-*`는 0 static references로도 활성화 가능).
- `moai-domain-copywriting`, `moai-domain-brand-design`의 `related-skills` 필드는 merge 후 alias로 유지되어 agent prompt 호환성을 보장한다.
- R4 audit의 `Per-skill audit table` 48개 entry 모두 하단 §6.2 판정표에 1:1 매핑된다.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-WF001-001** (Stage 1 target)
The v3 skill tree **shall** contain exactly 38 skill directories under `.claude/skills/` after Stage 1 consolidation (baseline 48 − 11 RETIRE + 1 NEW (`moai-design-system`) = 38). The long-term R4 audit target of ~24 skills is deferred to Stage 2 (SPEC-V3R3-WF-001) and is NOT a pass/fail criterion for this SPEC.

**REQ-WF001-002**
Each of the 48 current skills **shall** receive exactly one verdict label from {KEEP, REFACTOR, MERGE, RETIRE} recorded in §6.2 判定표.

**REQ-WF001-003**
Every MERGE verdict **shall** cite the target skill name and the source skill's trigger keyword union carried forward into the target's frontmatter `triggers:` or `related-skills` field.

**REQ-WF001-004**
Every RETIRE verdict **shall** include a migration note pointing to the substitute skill(s) that consumers must switch to.

**REQ-WF001-005**
The Agency-absorbed skills (`moai-domain-copywriting`, `moai-domain-brand-design`) **shall not** be modified by this SPEC (FROZEN per `.claude/rules/moai/design/constitution.md`).

**REQ-WF001-006**
Template tree (`internal/template/templates/.claude/skills/`) and local tree (`.claude/skills/`) **shall** remain byte-identical after each wave commit (enforced by `diff -rq`).

### 5.2 Event-Driven Requirements

**REQ-WF001-007**
**When** a MERGE cluster is applied, the system **shall** preserve the union of `triggers:` and `related-skills:` from all source skills in the merged target's frontmatter.

**REQ-WF001-008**
**When** a RETIRE verdict is applied, the system **shall** archive the retired skill to `.moai/archive/skills/v3.0/<skill-name>/` with a `RETIRED.md` note recording the migration substitute.

**REQ-WF001-009**
**When** `moai update` runs on a v2 project post-consolidation, the migrator (SPEC-V3R2-MIG-001) **shall** remove the 11 Stage-1 deprecated skill directories from the user's local `.claude/skills/` and deploy the 38 retained skills per the schema defined in plan.md §2.5 (shared contract; see spec.md §9.1.1).

**REQ-WF001-010**
**When** bundled resources (`modules/`, `references/`) exist in a source skill being merged, the system **shall** relocate still-referenced resources into the target skill's Level 3 payload and delete unused resources.

### 5.3 State-Driven Requirements

**REQ-WF001-011**
**While** a skill is marked REFACTOR, the skill directory **shall** remain in the v3 tree but its SKILL.md **shall** include a `## Refactor Notes` section pointing at the R4 audit line item.

**REQ-WF001-012**
**While** the `moai` root skill retains its 20 bundled `workflows/*.md` files, SPEC-V3R2-WF-002 **shall** be a strict dependency for any reduction; this SPEC **shall not** modify `moai/workflows/`.

### 5.4 Optional Requirements

**REQ-WF001-013**
**Where** a skill is marked UNCLEAR in R4 audit (e.g., `moai-framework-electron`, `moai-platform-chrome-extension`), the v3 tree **shall** retain the skill with a telemetry-gated re-audit window of 60 days.

**REQ-WF001-014**
**Where** an agent's prompt explicitly references a retired or merged skill name by string literal, the maintainer **shall** update the agent prompt within the same commit that applies the skill verdict.

### 5.5 Complex Requirements (Unwanted Behavior / Composite)

**REQ-WF001-015 (Unwanted Behavior)**
**If** a skill directory is deleted without its retirement note being created in `.moai/archive/skills/v3.0/`, **then** CI **shall** reject the commit with `SKILL_RETIRE_NO_ARCHIVE`.

**REQ-WF001-016 (Unwanted Behavior)**
**If** a MERGE target loses any trigger keyword that existed in its merged sources, **then** the skill-audit CI check **shall** fail with `SKILL_TRIGGER_DROP`.

**REQ-WF001-017 (Complex: State + Event)**
**While** the consolidation migration runs, **when** an unresolved `related-skills` reference points to a deleted skill name, the system **shall** automatically rewrite it to the merge target's name using the §6.2 mapping table.

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-WF001-01**: Given the v2.13.2 tree with 48 skills When the full Stage 1 consolidation is applied Then `.claude/skills/` contains exactly 38 directories (48 − 11 RETIRE + 1 NEW) (maps REQ-WF001-001).
- **AC-WF001-02**: Given R4 audit's 48 entries When the verdict table is compared against §6.2 Then all 48 entries have exactly one of {KEEP, REFACTOR, MERGE, RETIRE} (maps REQ-WF001-002).
- **AC-WF001-03**: Given the thinking triplet (`moai-foundation-thinking`, `moai-foundation-philosopher`, `moai-workflow-thinking`) When merged into `moai-foundation-thinking` Then the merged skill contains the union of all three source skills' triggers (maps REQ-WF001-007).
- **AC-WF001-04**: Given a RETIRE verdict on `moai-tool-svg` When the retirement applies Then `.moai/archive/skills/v3.0/moai-tool-svg/RETIRED.md` exists with substitute guidance (maps REQ-WF001-008).
- **AC-WF001-05**: Given `diff -rq .claude/skills internal/template/templates/.claude/skills` When run after any wave commit Then output is empty (maps REQ-WF001-006).
- **AC-WF001-06**: Given `moai-domain-copywriting` and `moai-domain-brand-design` frontmatter When inspected post-consolidation Then both files are byte-identical to pre-consolidation state (maps REQ-WF001-005).
- **AC-WF001-07**: Given an agent prompt referencing `moai-foundation-philosopher` When the consolidation commit lands Then the agent prompt is updated in the same commit to reference `moai-foundation-thinking` (maps REQ-WF001-014).
- **AC-WF001-08**: Given a skill directory deleted without archive entry When the dry-run audit script runs locally AND when a deliberately-broken CI fixture (missing `RETIRED.md`) is inserted in `.moai/specs/SPEC-V3R2-WF-001/fixtures/ci-reject/` Then (a) the local dry-run emits `SKILL_RETIRE_NO_ARCHIVE` diagnostics AND (b) the fixture-verifier task exits non-zero demonstrating CI rejection behavior (maps REQ-WF001-015).
- **AC-WF001-09**: Given `moai-workflow-templates` and `moai-workflow-project` merging into `moai-workflow-project` When merge runs Then templates' `schemas/` and `templates/` bundled directories are relocated under `moai-workflow-project/` (maps REQ-WF001-010).
- **AC-WF001-10**: Given `moai-foundation-thinking` after merge When inspected Then it references `moai-workflow-thinking` (Sequential Thinking MCP) and `moai-foundation-philosopher` (First Principles) content as internal sections (maps REQ-WF001-003).
- **AC-WF001-11**: Given `moai-design-system` after merge When inspected Then it unions triggers from `moai-design-craft`, `moai-domain-uiux`, `moai-design-tools` (UI side) (maps REQ-WF001-007).
- **AC-WF001-12**: Given `moai-domain-database` after merge When inspected Then it includes cloud vendor section from `moai-platform-database-cloud` (maps REQ-WF001-003).
- **AC-WF001-13**: Given `moai-framework-electron` and `moai-platform-chrome-extension` marked UNCLEAR When consolidation applies Then both retain directory and `## Telemetry Window` section is added to SKILL.md (maps REQ-WF001-013).
- **AC-WF001-14**: Given a REFACTOR skill (e.g., `moai-workflow-testing`) When consolidation commits Then its SKILL.md includes `## Refactor Notes` section linking to R4 audit line (maps REQ-WF001-011).
- **AC-WF001-15**: Given a skill's `related-skills:` pointing to a retired name When consolidation commits Then the reference is rewritten to the merge target per §6.2 mapping (maps REQ-WF001-017).
- **AC-WF001-16**: Given `moai update` running on a v2 project post-consolidation When the migrator reads `.moai/decisions/skill-rename-map.yaml` Then (a) the artifact exists with schema version 1, (b) every entry in the `retires` and `merges` sections has a corresponding migration plan in MIG-001's consumer code, verified by a static schema-match check performed in T1.7-9 (maps REQ-WF001-009).
- **AC-WF001-17**: Given this SPEC executes consolidation When any Wave (1.1 through 1.7) completes Then `git diff HEAD~N -- .claude/skills/moai/workflows/` shows zero changes AND `diff` between `moai/workflows/` contents before and after Stage 1 is empty (maps REQ-WF001-012 — this SPEC shall not modify moai/workflows/).
- **AC-WF001-18**: Given a MERGE target missing a trigger keyword that existed in at least one merged source When the audit script runs against the `.moai/specs/SPEC-V3R2-WF-001/fixtures/trigger-drop/` fixture (where one trigger has been intentionally dropped from a merge target) Then the fixture verifier exits non-zero with diagnostic `SKILL_TRIGGER_DROP: <trigger_name>` demonstrating the CI behavior the real guard rail will enforce (maps REQ-WF001-016).

### 6.1 Surviving skill logical grouping (non-authoritative; see §6.2 for directory count SoT)

> **NOTE (v1.1.0)**: §6.1 is a **logical grouping** of surviving skills for narrative clarity. The authoritative single source of truth for the post-Stage-1 directory count is §6.2. Stage 1 preserves 38 directories (not 24); the "24" long-term target is reserved for Stage 2 (SPEC-V3R3-WF-001). §6.1 enumerates the groups that R4 audit called out as the "core inventory"; additional KEEP/REFACTOR/UNCLEAR rows in §6.2 (e.g., `moai-platform-auth`, `moai-platform-deployment`, `moai-framework-electron`, `moai-platform-chrome-extension`, `moai-workflow-research`, `moai-workflow-pencil-integration`, `moai-workflow-design-context`, `moai-workflow-design-import`, `moai-domain-brand-design`, `moai-domain-db-docs`, `moai-formats-data`, `moai` root) all survive Stage 1 but are not duplicated in the groupings below.

**Foundation (4)**
1. `moai-foundation-core` (KEEP)
2. `moai-foundation-cc` (KEEP)
3. `moai-foundation-quality` (KEEP)
4. `moai-foundation-thinking` (MERGED — absorbs `moai-foundation-philosopher` + `moai-workflow-thinking`)

**Workflow (8)**
5. `moai-workflow-spec` (KEEP)
6. `moai-workflow-tdd` (KEEP)
7. `moai-workflow-ddd` (KEEP)
8. `moai-workflow-testing` (REFACTOR — split bundled modules/ into Level-3)
9. `moai-workflow-project` (KEEP — absorbs `moai-workflow-templates` + `moai-docs-generation`)
10. `moai-workflow-worktree` (KEEP)
11. `moai-workflow-loop` (KEEP — Ralph)
12. `moai-workflow-gan-loop` (KEEP)

**Design pipeline (4)**
13. `moai-workflow-design-context` (KEEP)
14. `moai-workflow-design-import` (KEEP — Path A handoff)
15. `moai-design-system` (NEW, MERGED — absorbs `moai-design-craft` + `moai-domain-uiux` + Pencil portion of `moai-design-tools`; Figma portion either absorbed into `moai-workflow-pencil-integration` or retired pending telemetry)
16. `moai-domain-copywriting` (KEEP — FROZEN agency contract)

**Domain (3)**
17. `moai-domain-backend` (REFACTOR — narrower "API design decision matrix")
18. `moai-domain-frontend` (REFACTOR — router to ref-react + library-nextra)
19. `moai-domain-database` (MERGED — absorbs `moai-platform-database-cloud`; `moai-domain-db-docs` remains separate workflow skill)

**Tools + Libraries (4)**
20. `moai-tool-ast-grep` (KEEP)
21. `moai-library-mermaid` (KEEP)
22. `moai-library-shadcn` (KEEP)
23. `moai-library-nextra` (KEEP)

**Agent-extending reference (5 — counted as aggregate item 24)**
24. `moai-ref-*` aggregate: `moai-ref-api-patterns`, `moai-ref-git-workflow`, `moai-ref-owasp-checklist`, `moai-ref-react-patterns`, `moai-ref-testing-pyramid` (all KEEP)

**Special item** (counted inside 24 via `moai` root position in `moai-foundation-core` ecosystem but retained as its own directory; SPEC-V3R2-WF-002 handles `moai/workflows/` reduction)
- `moai` root skill is retained but not double-counted; its fate is WF-002-bound.

### 6.2 판정표 (Verdict table for all 48 R4-audit entries)

| # | Skill | R4 verdict | v3 action | Notes |
|---|-------|------------|-----------|-------|
| 1 | moai | KEEP (split) | KEEP | `workflows/*.md` reduction deferred to SPEC-V3R2-WF-002 |
| 2 | moai-foundation-core | KEEP | KEEP | absorb `moai-foundation-context` content into §Token Budget section |
| 3 | moai-foundation-cc | KEEP | KEEP | unify `reference/` vs `references/` dir naming |
| 4 | moai-foundation-quality | KEEP | KEEP | — |
| 5 | moai-foundation-context | RETIRE | RETIRE (fold into foundation-core) | OQ-7 resolved v1.1.0: R4 column corrected from KEEP to RETIRE; single verdict = RETIRE. Content absorbed into moai-foundation-core. |
| 6 | moai-foundation-thinking | MERGE | MERGE target | unions 3-way thinking triplet |
| 7 | moai-foundation-philosopher | MERGE | RETIRE (merged) | absorbed into moai-foundation-thinking |
| 8 | moai-workflow-thinking | MERGE | RETIRE (merged) | absorbed into moai-foundation-thinking |
| 9 | moai-workflow-spec | KEEP | KEEP | — |
| 10 | moai-workflow-tdd | KEEP | KEEP | — |
| 11 | moai-workflow-ddd | KEEP | KEEP | — |
| 12 | moai-workflow-testing | REFACTOR | REFACTOR | split 43-file bundled tree into Level-3 |
| 13 | moai-workflow-templates | MERGE | RETIRE (merged) | absorbed into moai-workflow-project |
| 14 | moai-workflow-project | KEEP | KEEP | absorbs templates + docs-generation |
| 15 | moai-workflow-worktree | KEEP | KEEP | — |
| 16 | moai-workflow-loop | KEEP | KEEP | — |
| 17 | moai-workflow-jit-docs | RETIRE | RETIRE (merged) | OQ-1 resolved v1.1.0: R4 column corrected from KEEP to RETIRE; single verdict = RETIRE. Absorbed into moai-workflow-project documentation section. |
| 18 | moai-workflow-research | KEEP | KEEP (monitor) | retain experimental loop |
| 19 | moai-workflow-gan-loop | KEEP | KEEP | — |
| 20 | moai-workflow-design-import | KEEP | KEEP | — |
| 21 | moai-workflow-design-context | KEEP | KEEP | — |
| 22 | moai-workflow-pencil-integration | KEEP (monitor) | KEEP (absorbs Pencil portion of design-tools) | — |
| 23 | moai-domain-backend | REFACTOR | REFACTOR | narrow to "API design decision matrix" |
| 24 | moai-domain-frontend | REFACTOR | REFACTOR | router-only (ref-react, library-nextra) |
| 25 | moai-domain-database | REFACTOR | MERGE target | absorbs platform-database-cloud |
| 26 | moai-domain-uiux | MERGE | RETIRE (merged) | absorbed into moai-design-system |
| 27 | moai-domain-copywriting | KEEP | KEEP (FROZEN) | agency contract |
| 28 | moai-domain-brand-design | KEEP | KEEP (FROZEN) | agency contract |
| 29 | moai-domain-db-docs | KEEP | KEEP | separate workflow skill (migration parser) |
| 30 | moai-design-craft | MERGE | RETIRE (merged) | absorbed into moai-design-system |
| 31 | moai-design-tools | REFACTOR | RETIRE (split) | OQ-4 resolved v1.1.0: Pencil portion → `moai-workflow-pencil-integration` (authoritative target); Figma portion → archive under `.moai/archive/skills/v3.0/moai-design-tools/figma/` (no active substitute, revisit in Stage 2). The earlier tasks.md reference to `moai-design-system` as Pencil destination is SUPERSEDED by this resolution. |
| 32 | moai-docs-generation | REFACTOR | RETIRE (merged) | absorbed into moai-workflow-project |
| 33 | moai-platform-deployment | REFACTOR | REFACTOR | shrink triplet to Vercel-only; Railway/Convex doc-only |
| 34 | moai-platform-auth | REFACTOR | REFACTOR | retain triplet scope, narrower guidance per vendor |
| 35 | moai-platform-database-cloud | REFACTOR | RETIRE (merged) | absorbed into moai-domain-database |
| 36 | moai-platform-chrome-extension | KEEP (monitor) | KEEP (UNCLEAR window) | telemetry-gated 60-day window |
| 37 | moai-framework-electron | KEEP (monitor) | KEEP (UNCLEAR window) | telemetry-gated 60-day window |
| 38 | moai-library-nextra | KEEP (monitor) | KEEP | — |
| 39 | moai-library-mermaid | KEEP | KEEP | — |
| 40 | moai-library-shadcn | KEEP | KEEP | — |
| 41 | moai-tool-ast-grep | KEEP | KEEP | — |
| 42 | moai-tool-svg | REFACTOR | RETIRE | zero references, niche |
| 43 | moai-formats-data | KEEP (monitor) | KEEP (monitor) | TOON + JSON/YAML pattern library |
| 44 | moai-ref-api-patterns | KEEP | KEEP | — |
| 45 | moai-ref-git-workflow | KEEP | KEEP | — |
| 46 | moai-ref-react-patterns | KEEP | KEEP | — |
| 47 | moai-ref-testing-pyramid | KEEP | KEEP | — |
| 48 | moai-ref-owasp-checklist | KEEP | KEEP | — |

**Verdict roll-up (v1.1.0 corrected):**
- **RETIRE (archived)** = **11 directories**, each with a verdict cell of RETIRE in the v3 action column:
  `moai-foundation-context`, `moai-foundation-philosopher`, `moai-workflow-thinking`, `moai-workflow-templates`, `moai-workflow-jit-docs`, `moai-domain-uiux`, `moai-design-craft`, `moai-design-tools`, `moai-docs-generation`, `moai-platform-database-cloud`, `moai-tool-svg`.
- **MERGED-INTO (absorbing targets)** = **4 targets**: `moai-foundation-thinking`, `moai-workflow-project`, `moai-design-system` (NEW), `moai-domain-database`. (`moai-foundation-core` also absorbs `moai-foundation-context` content but is itself a pre-existing KEEP directory, not a new merge target.)
- **KEEP / KEEP (FROZEN) / KEEP (UNCLEAR window) / KEEP (monitor) / REFACTOR** = **37 pre-existing directories** survive Stage 1.
- **NEW** = 1 directory (`moai-design-system`).
- **Stage 1 arithmetic**: 48 baseline − 11 RETIRE + 1 NEW = **38 surviving directories** (REQ-WF001-001 target).
- **Stage 2 target (38 → 24)**: deferred to SPEC-V3R3-WF-001. The previous v0.1.0 roll-up line "48 − 13 − 11 = 24 ✓" was mathematically incoherent (the 13 and 11 were non-partitioning) and has been retired in v1.1.0.

---

## 7. Constraints (제약)

- FROZEN: `moai-domain-copywriting`, `moai-domain-brand-design` (agency contract per `.claude/rules/moai/design/constitution.md` §3).
- `moai-ref-*` 5개는 description-based auto-activation 모델을 유지한다 (R4 §Per-dimension scoring notes, "dead by grep, alive by design").
- 9-direct-dep 정책: 새 외부 의존성 도입 금지.
- Template-First(CLAUDE.local.md §2) + 언어 중립성(§15) 준수. 삭제되는 skill은 `.moai/archive/skills/v3.0/`에 아카이브된다.
- Progressive Disclosure Level 2 token budget(`skill-authoring.md`의 5000 token ceiling)을 merge 후 신규 target SKILL.md가 초과해서는 안 된다.

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 영향 | 완화 |
|---|---|---|
| Merge target이 Level 2 token budget을 초과 | Progressive Disclosure 위반 | merge 시 unused sections를 Level 3로 이관, max 5000 token 검증 CI |
| Trigger keyword union이 spam처럼 길어져 activation 정확도 저하 | skill auto-selection 오작동 | merge 후 keyword dedup + 테스트 |
| Agency FROZEN 계약 위반 실수 | GAN loop 계약 파괴 | pre-commit hook에서 agency skill 2종 byte-compare |
| Agent prompt가 retired skill 이름을 하드코딩 | 런타임 활성화 실패 | §6.2 mapping 기반 agent prompt 일괄 치환 + grep CI |
| `moai-formats-data`, `moai-framework-electron` 의 telemetry 부재 | UNCLEAR 판정 근거 불충분 | 60-day window 중 SessionStart 훅 activation count 로깅 추가 |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- (none — this SPEC is self-contained for repo-side enactment)

### 9.1.1 Shared Contract (co-signed with SPEC-V3R2-MIG-001)

- **SPEC-V3R2-MIG-001** consumes `.moai/decisions/skill-rename-map.yaml` (schema v1, defined in plan.md §2.5) to perform user-local migrations during `moai update`.
- REQ-WF001-009 mandates the behavioral contract: WHEN `moai update` runs THEN MIG-001 SHALL remove deprecated skill directories and deploy the Stage 1 38-skill set.
- Schema ownership: WF-001 authors the artifact; any schema change MUST be reviewed by the MIG-001 author via a HUMAN GATE (plan.md §Decision Log, `OQ-CONTRACT` entry).
- AC-WF001-16 verifies the contract existence and schema integrity at Wave 1.7.

### 9.2 Blocks

- SPEC-V3R2-WF-002 (Commands refactor): `moai` root skill의 `workflows/*.md` 축소는 WF-002가 담당하나 WF-001이 `moai` skill KEEP 판정을 유지해야 WF-002가 실행 가능.
- SPEC-V3R2-WF-005 (Language rules vs skills): `moai-lang-*` 부재 결정을 codify — 본 SPEC은 skills 집합만 확정.

### 9.3 Related

- SPEC-V3R2-WF-004 (Agentless classification): 일부 utility subcommand skills의 경로 변경과 인접.
- SPEC-V3R2-EXT-001 (Typed memory): merge 후 `.claude/agent-memory/` 구조에 영향 없음.

---

## 10. Traceability (추적성)

- REQ 총 17개: Ubiquitous 6, Event-Driven 4, State-Driven 2, Optional 2, Complex 3.
- AC 총 18개 (v1.1.0 에서 AC-16/17/18 추가), 모든 REQ에 최소 1개 AC 매핑 (17/17 = 100% 커버리지; v0.1.0 에서 누락되었던 REQ-009/012/016 이 v1.1.0 에서 각각 AC-16/17/18 로 매핑됨).
- REQ → AC 매핑: REQ-001→AC-01, REQ-002→AC-02, REQ-003→AC-10/12, REQ-004→AC-04, REQ-005→AC-06, REQ-006→AC-05, REQ-007→AC-03/11, REQ-008→AC-04, REQ-009→AC-16, REQ-010→AC-09, REQ-011→AC-14, REQ-012→AC-17, REQ-013→AC-13, REQ-014→AC-07, REQ-015→AC-08, REQ-016→AC-18, REQ-017→AC-15.
- Wave 2 소스 앵커: R4 audit §Executive summary / §Per-skill audit table / §Merge clusters / §Recommended v3 skill inventory / §Section A Category analysis; pattern-library §M-1/M-4.
- BC 영향: BC-V3R2-006 (user's local `.claude/skills/` tree에서 **11 directories** 삭제 — Stage 1). 본 SPEC은 **breaking**.
- 구현 경로 예상:
  - `.claude/skills/moai-foundation-thinking/SKILL.md` (확장)
  - `.claude/skills/moai-design-system/SKILL.md` (신설)
  - `.claude/skills/moai-domain-database/SKILL.md` (확장)
  - `.claude/skills/moai-workflow-project/SKILL.md` (templates + docs-generation + jit-docs 흡수)
  - `.claude/skills/moai-foundation-core/SKILL.md` (context 흡수)
  - `.moai/archive/skills/v3.0/<11 directories>/RETIRED.md`
  - `.moai/specs/SPEC-V3R2-WF-001/fixtures/ci-reject/` (broken-fixture test for AC-08)
  - `.moai/specs/SPEC-V3R2-WF-001/fixtures/trigger-drop/` (broken-fixture test for AC-18)
  - `internal/template/templates/.claude/skills/` 동기화
- 외부 참조: `.claude/rules/moai/development/skill-authoring.md` (frontmatter 규칙), `.claude/rules/moai/design/constitution.md` §3 (FROZEN).
- **File:line anchors** (per D5 traceability requirement):
  - `docs/design/major-v3-master.md:L1066` (§11.6 WF-001 definition)
  - `docs/design/major-v3-master.md:L965` (§8 BC-V3R2-006 — skill 48→24)
  - `docs/design/major-v3-master.md:L990` (§9 Phase 4 Skill Consolidation)
  - `.moai/design/v3-redesign/research/r4-skill-audit.md` (Per-skill audit, Recommended v3 skill inventory, Merge clusters)

---


### Out of Scope

- N/A (legacy SPEC)

## Implementation Notes (완료 기록)

### Completion Summary

**Status**: Completed (2026-04-25)

**Commits**: 11 implementation commits (Waves 1.1–1.7 + progress.md + fixup)
- All 48 skills received exactly one verdict label (REQ-WF001-002)
- 38 skill directories retained post-Stage 1 (REQ-WF001-001)
- 5 merge clusters assembled with trigger keyword union preservation (REQ-WF001-003, REQ-WF001-007)
- 11 retirement archives created in `.moai/archive/skills/v3.0/` (REQ-WF001-008)
- Template/local byte-identity maintained via `diff -rq` (REQ-WF001-006)
- Agency FROZEN contract preserved (REQ-WF001-005, AC-WF001-06)
- `.moai/decisions/skill-rename-map.yaml` artifact authored per contract (REQ-WF001-009, AC-WF001-16)
- CI fixtures implemented for retirement archive check (AC-WF001-08) and trigger drop detection (AC-WF001-18)

### Key Implementation Decisions (v1.1.0 Amendments)

1. **OQ-7 Resolution** (moai-foundation-context verdict): Corrected R4 audit verdict from KEEP to RETIRE; content absorbed into `moai-foundation-core` as §Token Budget section.
2. **OQ-1 Resolution** (moai-workflow-jit-docs verdict): Corrected R4 verdict from KEEP to RETIRE; merged into `moai-workflow-project` docs section.
3. **OQ-4 Resolution** (moai-design-tools split): Pencil portion → `moai-workflow-pencil-integration` (authoritative); Figma portion → `.moai/archive/` (no substitute, Stage 2 revisit).
4. **Stage 1 scope honesty** (§1.0): Confirmed 48→38 is achievable within single SPEC; 38→24 (Stage 2) reserved for SPEC-V3R3-WF-001.
5. **MIG-001 contract enforcement**: Schema change requires HUMAN GATE review before merge (plan.md §Decision Log, `OQ-CONTRACT` resolution).

### Test Coverage

- AC-WF001-01 through AC-WF001-18 all implemented (18/18 acceptance criteria)
- REQ-WF001-001 through REQ-WF001-017 all satisfied (17/17 requirements)
- Broken-fixture CI tests validate retirement archive requirements and trigger-drop blocking
- Drift Guard triggered after implementation to confirm scope tracking

### Known Limitations

- Stage 2 (38→24 transition) requires re-evaluation of MONITOR and REFACTOR categories; deferred per staging decision
- Telemetry windows for 2 UNCLEAR skills (`moai-framework-electron`, `moai-platform-chrome-extension`) require SessionStart hook logging (not yet implemented; tracked for v3.0 post-release)

---

End of SPEC.
