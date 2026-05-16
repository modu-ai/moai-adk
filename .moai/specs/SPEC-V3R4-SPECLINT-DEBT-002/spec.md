---
id: SPEC-V3R4-SPECLINT-DEBT-002
version: "0.2.0"
status: planned
created: 2026-05-16
updated: 2026-05-16
author: GOOS행님
priority: P1
tags: "v3r4, spec-lint, frontmatter, schema-canonical, plan-workflow, ssot, dual-schema-drift, foundation"
issue_number: null
title: plan workflow Pre-Write Frontmatter Checklist를 lint.go canonical (12-field)에 정렬 — dual-schema drift 영구 해소
phase: "v3.0.0 R4 — Foundation Cleanup"
module: ".claude/skills/moai/workflows/plan.md + .claude/skills/moai/team/plan.md + internal/spec/lint.go (조건부)"
dependencies: []
related_specs:
  - SPEC-V3R4-SPECLINT-DEBT-001
  - SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001
breaking: false
bc_id: []
lifecycle: spec-anchored
related_theme: "Foundation Cleanup — Schema SSOT 정립"
target_release: v3.0.0-rc1
---

# SPEC-V3R4-SPECLINT-DEBT-002 — plan workflow Frontmatter Checklist를 lint.go canonical로 정렬

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.2.0   | 2026-05-16 | manager-spec | Iteration 2 revision (plan-auditor review-1 0.62 FAIL → revise). D1: 75 SPECs 가정을 197 SPECs 실측 (51 snake_case `created_at:` dual-field, 53 `labels:` dual-field)으로 정정. Direction A 선택 근거 재구성 — snake_case는 redundant duplication (lint regression 0건)이며 cleanup은 opportunistic. D2: REQ-004를 Case C (team/plan.md에 Pre-Write Frontmatter Checklist 신설)로 재구성 — team/plan.md 존재 확인 (8065 bytes, 2026-05-14)되나 frontmatter checklist 섹션 부재. D3: 인용 commit SHA들 (`b2b7f32c7`/`ac6123be2`)이 squash-merge로 `--grep` 검색에는 잡히지 않으나 `git rev-parse`로 검증됨을 footnote에 명시. D4: SDF-001 dependency 상태를 "completed"로 갱신. D5: T-Wave3-002 finding count 6→3 정정 (AC-002와 정합). D6: AC-003 binary 검증 command 추가. D7: line 인용 정확도 향상 (lint.go:518-533, plan.md:449-458/461-465). D8: REQ-003을 schema property에서 분리. |
| 0.1.0   | 2026-05-16 | manager-spec | 초기 draft. SDF-001 run-phase (PR #939) hotfix commit `b2b7f32c7 fix(spec): SDF-001 spec.md frontmatter canonical 7-field 보강` 발생 → 근본 원인 분석 결과 plan workflow body (9-field, snake_case `created_at`/`updated_at`/`labels`/`issue_number`)와 internal/spec/lint.go FrontmatterSchemaRule (12-field, `created`/`updated`/`tags`/`title`/`phase`/`module`/`lifecycle`) 사이 dual-schema drift 확인. |

---

## 1. Goal

`.claude/skills/moai/workflows/plan.md` § "Pre-Write Frontmatter Checklist" (lines 447-473)를 `internal/spec/lint.go` `FrontmatterSchemaRule.Check()` (lines 517-533)의 12-field canonical schema와 정확히 일치시켜, manager-spec이 plan workflow를 가이드로 작성한 frontmatter가 `moai spec lint --strict` 검증을 통과하도록 한다. 추가로 `.claude/rules/moai/development/spec-frontmatter-schema.md` 단일 SSOT 문서를 신설하여 plan.md + lint.go + plan-auditor 3계층이 동일 schema source를 참조하도록 정립한다.

### 1.1 배경

- **Trigger 사건**: SDF-001 (SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001) run-phase에서 commit `b2b7f32c7 fix(spec): SDF-001 spec.md frontmatter canonical 7-field 보강` hotfix 발생. plan workflow를 따라 작성한 9-field frontmatter가 lint.go 12-field 검증에서 fail → 7개 field (`title`, `phase`, `module`, `lifecycle`, `created`, `updated`, `tags`) 보강 필요. (인용 commit SHA 검증 방법: footnote¹ 참조)
- **lint.go canonical (12 required fields, internal/spec/lint.go:518-533)**:
  - `id`, `title`, `version`, `status`, `created`, `updated`, `author`, `priority`, `phase`, `module`, `lifecycle`, `tags`
- **plan.md workflow body (9 required fields, .claude/skills/moai/workflows/plan.md:449-458)**:
  - `id`, `version`, `status`, `created_at`, `updated_at`, `author`, `priority`, `labels`, `issue_number`
- **불일치 매트릭스**:
  - plan.md가 `created_at`/`updated_at` 강제 → lint.go는 `created`/`updated` 요구 → field name drift
  - plan.md가 `labels` (array) 강제 → lint.go는 `tags` (CSV string) 요구 → field name + type drift
  - plan.md가 7개 field (`title`/`phase`/`module`/`lifecycle`) 미언급 → lint.go는 required → 누락
  - plan.md가 `issue_number` 강제 → lint.go는 미검증 (optional) → harmless 추가
  - plan.md "Rejected legacy aliases" 섹션 (.claude/skills/moai/workflows/plan.md:461-465)이 `created:`/`updated:`/`tags:`를 reject로 명시 → lint.go canonical과 정반대 (lint.go가 정확히 이 필드들을 요구함)
- **실측 현황 (main HEAD `b14290946` 기준, 2026-05-16T03:11 sync-phase merge 직후)**:
  - 총 **197 SPECs** (`find .moai/specs -name spec.md | wc -l` → 197).
  - **51 SPECs**가 `created_at:` snake_case alias 사용 (`grep -l "^created_at:" .moai/specs/*/spec.md | wc -l` → 51). **모두 canonical `created:`와 dual-field 공존** (`grep -l "^created_at:" | xargs grep -L "^created:" | wc -l` → 0). lint regression **0건**.
  - **53 SPECs**가 `labels:` array 사용 (`grep -l "^labels:" .moai/specs/*/spec.md | wc -l` → 53). 모두 canonical `tags:`와 dual-field 공존. lint regression **0건**.
  - `moai spec lint --strict` (main 전체): 0 ERROR / 0 WARNING.
  - **근본 원인**: SDF-001 hotfix 직후 manager-spec이 plan workflow body의 snake_case 가이드를 따르면서도 SDF-001 hotfix 학습 결과 canonical (`created:`/`tags:`)을 함께 작성하는 **방어적 dual-field 패턴**이 정착. lint regression은 차단되었으나 frontmatter bloat (redundant metadata 발생).
- **재발 위험**: 본 drift를 해소하지 않을 경우 향후 모든 새 SPEC 작성 시 manager-spec이 plan workflow를 충실히 따를수록 redundant dual-field 누적 → frontmatter bloat 가속 → 향후 schema 마이그레이션 비용 증가. lint regression은 SDF-001 hotfix 학습으로 차단되었으나 SSOT 부재로 같은 dual-field 패턴이 영속화될 우려.

---

¹ **Trigger evidence commit SHA 검증**: 본 SPEC 인용 commit (`b2b7f32c73ecc20133478a7a92e5306ec05696b1` SDF-001 frontmatter hotfix, `ac6123be28ea0124a3fb28ec03b3b50966c00054` SDF-001 sync prefix) 둘 다 squash-merge 과정에서 squash commit (`ce779f9ee feat(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001`, `b14290946`)으로 통합되어 `git log --grep "frontmatter canonical"` 검색에는 잡히지 않으나 `git rev-parse b2b7f32c7` / `git rev-parse ac6123be2`로 reflog 검증 가능. PR #939/#940 squash 머지로 main에 직접 commit으로는 존재하지 않지만, working-branch 작업 history는 보존됨.

### 1.2 사용자 결정 사항

다음 결정은 plan-phase에서 잠정 채택된다. run-phase에서 plan-auditor 결과에 따라 조정될 수 있다:

1. **방향 선택 (Direction A — lint.go canonical 유지)**: plan workflow를 lint.go에 정렬한다. **197 기존 SPECs 모두 canonical (`created:`/`updated:`/`tags:`) 필드를 이미 보유**하므로 lint regression 0건. 51건의 `created_at:` / 53건의 `labels:` snake_case alias는 canonical과 dual-field로 공존 — **redundant duplication**이며 능동적 drift가 아님. cleanup은 opportunistic (별도 follow-up SPEC 후보, 본 SPEC scope 외). 대안 (Direction B — plan workflow canonical로 lint.go 변경 + 197 SPECs 양방향 migration)은 schema 자체 변경 + lint.go regression 위험 + plan-auditor schema check 동반 갱신 부담으로 reject.
2. **issue_number 처리**: lint.go required 추가하지 않음. plan.md에서는 optional로 명시.
3. **snake_case 표준화 미적용**: Python/Ruby 컨벤션을 Go/YAML에 강제하지 않음. lint.go canonical (`created`/`updated`/`tags`)이 v3.0 시점 SSOT.
4. **SSOT 문서 신설**: `.claude/rules/moai/development/spec-frontmatter-schema.md` 단일 진실 공급원. plan.md + lint.go + plan-auditor 모두 이 문서를 참조하도록 cross-reference.

---

## 2. Scope

### 2.1 In Scope

- **F-IN-001**: `.claude/skills/moai/workflows/plan.md` § "Pre-Write Frontmatter Checklist" (lines 447-473) body 수정 — 9 fields → 12 fields, snake_case → canonical, "Rejected legacy aliases" 섹션 반전.
- **F-IN-002**: `.claude/skills/moai/team/plan.md` (team variant)에 **Pre-Write Frontmatter Checklist 섹션 신설** (Case C). 실측 (plan-phase 검증): 파일은 존재 (8065 bytes, 2026-05-14) 하나 frontmatter checklist 섹션 부재 (`grep -i "frontmatter\|12 required\|9 required" team/plan.md` → 0 matches). team 모드에서도 manager-spec이 SPEC 작성을 담당 (team/plan.md:157 "Delegate SPEC creation to manager-spec subagent")하므로 solo plan.md의 checklist를 team variant에도 mirror하여 일관성 확보. ~75 LOC 삽입 (team/plan.md:155 "delegate to manager-spec" 직후, "SPEC output at:" 직전).
- **F-IN-003**: `.claude/rules/moai/development/spec-frontmatter-schema.md` 신설 — 단일 SSOT 문서, 12-field canonical + optional fields + rejected aliases 명시.
- **F-IN-004**: Regression test fixture — `internal/spec/testdata/frontmatter-schema/` 디렉토리 신설. 12-field 준수 fixture + 의도적 누락 fixture 각 1건 추가. lint.go FrontmatterSchemaRule 검증 + plan-auditor schema check 검증 양쪽 모두 expected behavior 확인.
- **F-IN-005**: `.claude/rules/moai/development/coding-standards.md` 또는 유사 cross-reference 위치에 신설 SSOT 문서 link 추가 (discoverability 확보).

### 2.2 Out of Scope (What NOT to Build)

- **F-OUT-001**: 197 기존 SPECs frontmatter dual-field cleanup — 51건 `created_at:` + 53건 `labels:` snake_case alias는 canonical과 dual-field 공존, lint regression 0건. opportunistic cleanup (별도 follow-up SPEC 후보). 본 SPEC은 미래 SPEC 작성 가이드만 정렬.
- **F-OUT-002**: lint.go `FrontmatterSchemaRule` 12-field 의미적 변경 — canonical 유지가 본 SPEC의 정의. field 추가/제거 모두 별도 SPEC 후보.
- **F-OUT-003**: `title`/`phase`/`module`/`lifecycle` 필드의 의미적 검증 강화 (예: `phase` 패턴 검증, `lifecycle` 값 enum 검증) — 별도 SPEC 후보 (V3R4-SPECLINT-DEBT-003 가능).
- **F-OUT-004**: `issue_number`를 lint.go required로 추가 — 50% 사용률 + tracking 미수립 SPEC도 정당, optional 유지가 합리적.
- **F-OUT-005**: snake_case 표준화 — Python/Ruby 관례를 Go/YAML에 강제하지 않음. lint.go canonical 보호.
- **F-OUT-006**: `dependencies`/`related_specs`/`bc_id`/`breaking`/`related_theme`/`target_release` 등 optional metadata 필드의 검증 강화 — DependencyExistsRule, DependencyCycleRule 등은 별도 rule로 이미 존재. 본 SPEC은 schema field 정렬만 다룸.
- **F-OUT-007**: plan workflow의 다른 section (Phase 1/2/3/4 구조) 변경 — Pre-Write Frontmatter Checklist (lines 447-473)에 한정.

---

## 3. EARS Requirements

### REQ-SDBT-002-001 (Ubiquitous)

The `.claude/skills/moai/workflows/plan.md` § "Pre-Write Frontmatter Checklist" body **shall** list exactly the 12 required fields enforced by `internal/spec/lint.go` `FrontmatterSchemaRule.Check()` — `id`, `title`, `version`, `status`, `created`, `updated`, `author`, `priority`, `phase`, `module`, `lifecycle`, `tags` — in canonical order.

### REQ-SDBT-002-002 (Ubiquitous)

The plan.md "Rejected legacy aliases" section **shall** list `created_at:`, `updated_at:`, `labels:`, `spec_id:` as rejected (lint.go canonical 보호 방향). The previous direction (rejecting `created:`/`updated:`/`tags:`) **shall not** appear.

### REQ-SDBT-002-003 (Event-driven)

**When** manager-spec composes a new SPEC frontmatter using plan workflow body as guide, the agent **shall** include all 12 canonical required fields (`id`, `title`, `version`, `status`, `created`, `updated`, `author`, `priority`, `phase`, `module`, `lifecycle`, `tags`) and **shall not** add snake_case aliases (`created_at:`, `updated_at:`, `labels:`) as redundant duplicates.

### REQ-SDBT-002-004 (Event-driven)

**When** `.claude/skills/moai/workflows/plan.md` Pre-Write Frontmatter Checklist section is modified, `.claude/skills/moai/team/plan.md` **shall** have an equivalent Pre-Write Frontmatter Checklist section created (Case C: new section addition, not sync of existing). The section content MUST match solo plan.md verbatim with team-mode adaptations only where applicable (e.g., teammate coordination references).

### REQ-SDBT-002-005 (State-driven)

**While** the SSOT reference document `.claude/rules/moai/development/spec-frontmatter-schema.md` exists, the three schema enforcement layers — plan workflow body, `internal/spec/lint.go` `FrontmatterSchemaRule`, and plan-auditor schema check — **shall** all reference this document as the single source of truth via explicit cross-reference links.

---

## 4. Acceptance Criteria (Summary)

자세한 Given-When-Then 시나리오는 `acceptance.md` 참조.

- **AC-SDBT-002-001**: manager-spec이 plan workflow body를 가이드로 작성한 새 SPEC이 `moai spec lint --strict`에서 FrontmatterInvalid 0건 통과.
- **AC-SDBT-002-002**: snake_case `created_at:` 또는 `labels:` field만 가진 (canonical 누락) frontmatter가 lint.go에서 정확히 **3 FrontmatterInvalid findings** (`created`, `updated`, `tags` 누락)를 보고. fixture는 `internal/spec/testdata/frontmatter-schema/invalid-snake-case-only/spec.md` (Wave 3).
- **AC-SDBT-002-003**: SDF-001과 동일한 hotfix 시나리오 (frontmatter 7-field 보강)가 본 SPEC 적용 후 재현 불가 — manager-spec 가이드만으로 lint pass. Binary 검증: `git log --since=2026-05-16 --grep "fix(spec).*frontmatter canonical" --oneline | wc -l` MUST = 0 (sync-phase 시점 baseline; 본 SPEC 머지 이후 추가 hotfix commit 없음 확인).
- **AC-SDBT-002-004**: SSOT 문서 `.claude/rules/moai/development/spec-frontmatter-schema.md`가 존재하며 plan.md / lint.go 주석 / plan-auditor가 모두 이 문서를 참조.
- **AC-SDBT-002-005**: `internal/spec/testdata/frontmatter-schema/` regression fixture가 lint.go 검증 + plan-auditor schema check 양쪽에서 expected behavior 입증.

---

## 5. Out of Scope (Exclusions)

본 SPEC에서 **명시적으로 제외**되는 항목:

1. **197 기존 SPECs frontmatter dual-field cleanup** — 51건 `created_at:` + 53건 `labels:` snake_case alias는 canonical과 dual-field 공존, lint regression 0건. opportunistic cleanup (별도 SPEC 후보), 본 SPEC scope 외.
2. **lint.go `FrontmatterSchemaRule` 12-field 변경** — canonical 유지가 본 SPEC의 전제. field 추가/제거는 별도 SPEC.
3. **`title`/`phase`/`module`/`lifecycle` 의미적 검증 강화** — field 존재만 검증, 값 검증은 별도 SPEC 후보.
4. **`issue_number` → lint.go required 추가** — 50% 사용률 + optional 유지 권장.
5. **snake_case 표준화** — Python/Ruby 관례 강제하지 않음.
6. **plan workflow body의 다른 section 변경** — Pre-Write Frontmatter Checklist (lines 447-473)에 한정.
7. **`dependencies`/`related_specs` 등 optional metadata 검증 강화** — 별도 rule로 이미 존재 (DependencyExistsRule 등).
8. **SPEC frontmatter YAML parser 변경** — 기존 internal/spec/parse.go 동작 유지.

---

## 6. Reference Implementation

- `internal/spec/lint.go:502-569` — `FrontmatterSchemaRule` 정의 + `Check()` 함수 전체.
  - `internal/spec/lint.go:518-533` — `required[]` slice (12 canonical field 정의).
- `.moai/specs/SPEC-V3R4-SPECLINT-DEBT-001/spec.md` — 12-field 패턴 + optional metadata 적용 예시 (본 SPEC도 동일 패턴).
- `.claude/skills/moai/workflows/plan.md:447-473` — 수정 대상 section "Pre-Write Frontmatter Checklist" 전체.
  - `.claude/skills/moai/workflows/plan.md:449-458` — Required 9 fields 목록 (수정 대상).
  - `.claude/skills/moai/workflows/plan.md:461-465` — "Rejected legacy aliases" 섹션 (방향 반전 대상).
- `.moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001/` — Trigger 사건 (hotfix commit `b2b7f32c7`, footnote¹ 참조).

---

## 7. Constraints

- **C-001 (Frontmatter schema)**: 본 SPEC 자체의 spec.md frontmatter는 lint.go 12-field canonical을 준수해야 한다 (id/title/version/status/created/updated/author/priority/phase/module/lifecycle/tags).
- **C-002 (Language)**: 모든 SPEC 문서 본문 (spec.md, plan.md, acceptance.md, spec-compact.md)은 한국어 (`conversation_language=ko`)로 작성.
- **C-003 (EARS scope)**: 최대 5개 EARS requirements 준수 (REQ-SDBT-002-001 ~ 005).
- **C-004 (No code implementation)**: plan-phase는 planning artifacts만 생성. 실제 plan.md / lint.go / SSOT 문서 수정은 run-phase에서 수행.
- **C-005 (Plan-in-main)**: BODP 평가 — Signal A=¬ (depends_on 없음, diff overlap 없음), B=¬ (작업 트리 co-location 없음), C=¬ (현재 branch open PR 없음) → main @ origin/main. plan-in-main 원칙 (PR #822 doctrine) 준수.
- **C-006 (Lifecycle)**: `spec-anchored` — implementation이 plan을 따르되 spec.md/plan.md/acceptance.md가 living document로 유지.

---

## 8. Risks & Mitigations

자세한 risk analysis는 `plan.md` § Risks 참조.

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| team/plan.md에 Pre-Write Frontmatter Checklist 신설이 team variant 본문 구조와 충돌 | Low | Medium | plan-phase 실측 완료 (team/plan.md:155-160 SPEC delegation 직후가 적절한 위치 확인); F-IN-002에 LOC + 삽입 위치 명시 |
| 51+53 dual-field SPECs가 cleanup 압박 (scope creep) | Medium | Low | spec.md §1.2 #1에 "opportunistic cleanup, 본 SPEC scope 외" 명시; F-OUT-001로 마이그레이션 명시 제외 |
| SSOT 문서 신설 후 cross-reference 누락 (plan.md/lint.go가 참조 안 함) | Medium | Low | run-phase Wave 3에서 plan.md, lint.go 주석, plan-auditor 모두 link 명시 |
| issue_number optional 처리가 향후 GitHub Issue tracking 정책과 충돌 | Low | Medium | F-OUT-004로 명시 — 정책 변경 시 별도 SPEC |
| Regression fixture가 lint.go 변경에 취약 (brittle) | Low | Low | fixture는 12-field 명시 + 의도적 누락만 — lint.go semantic 변경 시 fixture도 동반 업데이트 |

---

## 9. Dependencies

- **None (no blocking SPECs)**: 본 SPEC은 SPEC-V3R4-SPECLINT-DEBT-001 (completed) 및 SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 (completed, PR #940 merged 2026-05-16T03:11Z `b14290946`)의 사후 follow-up이지만 blocking 관계 없음. 둘 다 main에 merged 완료되어 본 SPEC 작업에 영향 없음.
- **Related (non-blocking)**:
  - SPEC-V3R4-SPECLINT-DEBT-001 (lint.go canonical 보호 baseline 제공)
  - SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 (trigger 사건 source)

---

## 10. Definition of Done

1. `.claude/skills/moai/workflows/plan.md` § Pre-Write Frontmatter Checklist가 12-field canonical로 갱신됨 (REQ-SDBT-002-001).
2. plan.md "Rejected legacy aliases" 섹션이 snake_case alias (`created_at:`, `updated_at:`, `labels:`)를 reject로 명시 (REQ-SDBT-002-002).
3. `.claude/skills/moai/team/plan.md` variant에 Pre-Write Frontmatter Checklist 섹션 신설 (Case C, REQ-SDBT-002-004). solo plan.md와 12-field schema verbatim 일치.
4. `.claude/rules/moai/development/spec-frontmatter-schema.md` SSOT 문서 신설 + plan.md / lint.go 주석 / plan-auditor 모두 cross-reference (REQ-SDBT-002-005).
5. `internal/spec/testdata/frontmatter-schema/` regression fixture 2건 (PASS fixture + FAIL fixture) 추가 + 검증 통과.
6. `moai spec lint --strict` 본 SPEC 디렉토리 (`.moai/specs/SPEC-V3R4-SPECLINT-DEBT-002/`) 0 ERROR / 0 WARNING.
7. AC-SDBT-002-001 ~ 005 모두 verified.
8. CHANGELOG `[Unreleased]` 항목 추가 (sync-phase).
9. frontmatter `status` `draft → in-progress → implemented → completed` lifecycle 정상 transition.
10. plan PR + run PR + sync PR 모두 main squash merge.
