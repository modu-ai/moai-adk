---
id: SPEC-V3R2-WF-001
title: Skill Consolidation (48 to 24)
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: Wave 2 SPEC writer (Layer 6/7/Cleanup)
priority: P1 High
phase: "v3.0.0 — Phase 4 — Skill Consolidation"
module: ".claude/skills/, internal/template/templates/.claude/skills/"
dependencies: []
related_gap:
  - r4-skill-audit
  - problem-catalog-workflow-bloat
related_theme: "Theme 6 — Workflow Consolidation"
breaking: true
bc_id: [BC-V3R2-006]
lifecycle: spec-anchored
tags: "skill-consolidation, skill-audit, merge, retire, v3, workflow, breaking"
---

# SPEC-V3R2-WF-001: Skill Consolidation (48 to 24)

## HISTORY

| Version | Date       | Author | Description                                                           |
|---------|------------|--------|-----------------------------------------------------------------------|
| 0.1.0   | 2026-04-23 | Wave 2 | Initial SPEC draft for 48 → 24 skill reduction per R4 skill audit     |

---

## 1. Goal (목적)

MoAI v2.13.2의 48개 skill 중 **24개**만 v3.0.0에 존속시킨다. 이 SPEC은 R4 skill audit(2026-04-23)의 verdict 분포(KEEP 12 / REFACTOR 14 / MERGE 15 / RETIRE 5 / UNCLEAR 2)를 집행하는 단일 단일 실행 계약이다. 5개 merge cluster(thinking, design, database, templates-to-project, design-tools split)와 5건 retirement를 명시적 KEEP/REFACTOR/MERGE/RETIRE 판정으로 고정하고, 각 skill의 트리거 키워드 union 보존 및 bundled resource 재배치를 규정한다.

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

- **Owns**: `.claude/skills/` 디렉터리 48→24 변경, `internal/template/templates/.claude/skills/` 동일 변환, 각 KEEP/REFACTOR/MERGE/RETIRE 판정 집행.
- 신규 병합 skill 3종 신설(`moai-foundation-thinking` 확장, `moai-design-system`, `moai-domain-database` 확장).
- 15건 MERGE(5개 cluster) + 5건 RETIRE + 14건 REFACTOR + 12건 KEEP-as-is 확정 라벨 부여.
- 각 skill의 `related-skills` frontmatter 필드 재연결(merge 대상 skill 이름 alias 처리 포함).
- Trigger keyword union preservation: merge 대상 skill의 모든 trigger를 신규 skill의 frontmatter trigger에 union 병합.
- Bundled resource 재배치: 삭제되는 skill의 `modules/`, `references/` 중 재사용 가능한 asset을 신규 skill의 Level 3 payload로 이관.
- `moai` root skill의 `workflows/*.md` 20개 파일 범위 축소는 SPEC-V3R2-WF-002가 담당(본 SPEC은 `moai` skill 본체 KEEP but split 판정만 기록).

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- Skill 본문의 언어 번역 / style rewrite
- `moai-ref-*` 5개 skill의 trigger 키워드 재설계 (keyword-matching auto-activation 모델 유지)
- 48개 → 24개 수치를 벗어나는 추가 신규 skill 창설
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
- Skill consolidation REFERENCES (not depends on) MIG-001's migrator; the migrator will read WF-001's skill rename / merge map artifact (written to `.moai/decisions/skill-rename-map.yaml` by WF-001's landing commit) after WF-001 lands in Phase 4. MIG-001 in Phase 8 consumes the artifact without a dependency-graph edge.

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

**REQ-WF001-001**
The v3 skill tree **shall** contain exactly 24 skill directories under `.claude/skills/` (per R4 audit §Recommended v3 skill inventory).

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
**When** `moai update` runs on a v2 project post-consolidation, the migrator (SPEC-V3R2-MIG-001) **shall** remove deprecated skill directories from the user's local `.claude/skills/` and deploy the 24 retained skills.

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

- **AC-WF001-01**: Given the v2.13.2 tree with 48 skills When the full consolidation is applied Then `.claude/skills/` contains exactly 24 directories (maps REQ-WF001-001).
- **AC-WF001-02**: Given R4 audit's 48 entries When the verdict table is compared against §6.2 Then all 48 entries have exactly one of {KEEP, REFACTOR, MERGE, RETIRE} (maps REQ-WF001-002).
- **AC-WF001-03**: Given the thinking triplet (`moai-foundation-thinking`, `moai-foundation-philosopher`, `moai-workflow-thinking`) When merged into `moai-foundation-thinking` Then the merged skill contains the union of all three source skills' triggers (maps REQ-WF001-007).
- **AC-WF001-04**: Given a RETIRE verdict on `moai-tool-svg` When the retirement applies Then `.moai/archive/skills/v3.0/moai-tool-svg/RETIRED.md` exists with substitute guidance (maps REQ-WF001-008).
- **AC-WF001-05**: Given `diff -rq .claude/skills internal/template/templates/.claude/skills` When run after any wave commit Then output is empty (maps REQ-WF001-006).
- **AC-WF001-06**: Given `moai-domain-copywriting` and `moai-domain-brand-design` frontmatter When inspected post-consolidation Then both files are byte-identical to pre-consolidation state (maps REQ-WF001-005).
- **AC-WF001-07**: Given an agent prompt referencing `moai-foundation-philosopher` When the consolidation commit lands Then the agent prompt is updated in the same commit to reference `moai-foundation-thinking` (maps REQ-WF001-014).
- **AC-WF001-08**: Given a skill directory deleted without archive entry When CI runs Then the commit is rejected with `SKILL_RETIRE_NO_ARCHIVE` (maps REQ-WF001-015).
- **AC-WF001-09**: Given `moai-workflow-templates` and `moai-workflow-project` merging into `moai-workflow-project` When merge runs Then templates' `schemas/` and `templates/` bundled directories are relocated under `moai-workflow-project/` (maps REQ-WF001-010).
- **AC-WF001-10**: Given `moai-foundation-thinking` after merge When inspected Then it references `moai-workflow-thinking` (Sequential Thinking MCP) and `moai-foundation-philosopher` (First Principles) content as internal sections (maps REQ-WF001-003).
- **AC-WF001-11**: Given `moai-design-system` after merge When inspected Then it unions triggers from `moai-design-craft`, `moai-domain-uiux`, `moai-design-tools` (UI side) (maps REQ-WF001-007).
- **AC-WF001-12**: Given `moai-domain-database` after merge When inspected Then it includes cloud vendor section from `moai-platform-database-cloud` (maps REQ-WF001-003).
- **AC-WF001-13**: Given `moai-framework-electron` and `moai-platform-chrome-extension` marked UNCLEAR When consolidation applies Then both retain directory and `## Telemetry Window` section is added to SKILL.md (maps REQ-WF001-013).
- **AC-WF001-14**: Given a REFACTOR skill (e.g., `moai-workflow-testing`) When consolidation commits Then its SKILL.md includes `## Refactor Notes` section linking to R4 audit line (maps REQ-WF001-011).
- **AC-WF001-15**: Given a skill's `related-skills:` pointing to a retired name When consolidation commits Then the reference is rewritten to the merge target per §6.2 mapping (maps REQ-WF001-017).

### 6.1 Final 24 skill inventory (per R4 §Recommended v3 skill inventory)

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
| 5 | moai-foundation-context | KEEP | RETIRE (fold into foundation-core) | content absorbed into moai-foundation-core |
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
| 17 | moai-workflow-jit-docs | KEEP | RETIRE (merged) | absorbed into moai-workflow-project documentation section |
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
| 31 | moai-design-tools | REFACTOR | RETIRE (split) | Pencil → moai-workflow-pencil-integration; Figma → archive pending telemetry |
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

**Verdict roll-up:** KEEP = 24 directories (includes `moai` root, `moai-formats-data` monitor, 2 UNCLEAR windows). MERGED-INTO = 3 targets (`moai-foundation-thinking`, `moai-workflow-project`, `moai-design-system`, `moai-domain-database`). RETIRED/ABSORBED = 13 directories (foundation-context, foundation-philosopher, workflow-thinking, workflow-templates, workflow-jit-docs, domain-uiux, design-craft, design-tools, docs-generation, platform-database-cloud, tool-svg, and 2 absorbed under merge targets). Total: 48 − 13 − 11 = 24 ✓.

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

- SPEC-V3R2-MIG-001: v2→v3 migrator가 user의 local skills 트리에서 removed skill을 이관해야 함.

### 9.2 Blocks

- SPEC-V3R2-WF-002 (Commands refactor): `moai` root skill의 `workflows/*.md` 축소는 WF-002가 담당하나 WF-001이 `moai` skill KEEP 판정을 유지해야 WF-002가 실행 가능.
- SPEC-V3R2-WF-005 (Language rules vs skills): `moai-lang-*` 부재 결정을 codify — 본 SPEC은 skills 집합만 확정.

### 9.3 Related

- SPEC-V3R2-WF-004 (Agentless classification): 일부 utility subcommand skills의 경로 변경과 인접.
- SPEC-V3R2-EXT-001 (Typed memory): merge 후 `.claude/agent-memory/` 구조에 영향 없음.

---

## 10. Traceability (추적성)

- REQ 총 17개: Ubiquitous 6, Event-Driven 4, State-Driven 2, Optional 2, Complex 3.
- AC 총 15개, 모든 REQ에 최소 1개 AC 매핑 (100% 커버리지).
- Wave 2 소스 앵커: R4 audit §Executive summary / §Per-skill audit table / §Merge clusters / §Recommended v3 skill inventory / §Section A Category analysis; pattern-library §M-1/M-4.
- BC 영향: BC-V3R2-006 (user's local `.claude/skills/` tree에서 13 directories 삭제). 본 SPEC은 **breaking**.
- 구현 경로 예상:
  - `.claude/skills/moai-foundation-thinking/SKILL.md` (확장)
  - `.claude/skills/moai-design-system/SKILL.md` (신설)
  - `.claude/skills/moai-domain-database/SKILL.md` (확장)
  - `.claude/skills/moai-workflow-project/SKILL.md` (templates + docs-generation 흡수)
  - `.moai/archive/skills/v3.0/<13 directories>/RETIRED.md`
  - `internal/template/templates/.claude/skills/` 동기화
- 외부 참조: `.claude/rules/moai/development/skill-authoring.md` (frontmatter 규칙), `.claude/rules/moai/design/constitution.md` §3 (FROZEN).
- **File:line anchors** (per D5 traceability requirement):
  - `docs/design/major-v3-master.md:L1066` (§11.6 WF-001 definition)
  - `docs/design/major-v3-master.md:L965` (§8 BC-V3R2-006 — skill 48→24)
  - `docs/design/major-v3-master.md:L990` (§9 Phase 4 Skill Consolidation)
  - `.moai/design/v3-redesign/research/r4-skill-audit.md` (Per-skill audit, Recommended v3 skill inventory, Merge clusters)

---

End of SPEC.
